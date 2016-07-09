/**
 * @file   ldi.go
 * @author Sébastien Rouault <sebastien.rouault@epfl.ch>
 *
 * @section LICENSE
 *
 * Copyright (c) 2014 Vasileios Trigonakis <vasileios.trigonakis@epfl.ch>,
 *                    Tudor David <tudor.david@epfl.ch>
 *                    Distributed Programming Lab (LPD), EPFL
 *
 * ASCYLIB is free software: you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation, version 2
 * of the License.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * @section DESCRIPTION
 *
 * Latency distribution test.
**/

package main

import (
    "dataset"
    "flag"
    "fmt"
    "math"
    "sync"
    "sync/atomic"
    "tools/share"
    "time"
    "tools/assert"
    "tools/thread"
    "tools/volatile"
    "tools/xorshift"
    "unsafe"
)

const (
    calibrate_avgs  = 200 // How many measures for each average
    calibrate_reps  = 5   // How many repetitions (with pause time)
    calibrate_pause = 50  // Pause time (in ms)
)

// -----------------------------------------------------------------------------

// True if the tests are running
var running int32

// Thread run statistics
type stats_t struct {
    put_count    uint64
    put_time     uint64
    get_count    uint64
    get_time     uint64
    remove_count uint64
    remove_time  uint64
}

// -----------------------------------------------------------------------------

func isPow2(x uint) bool {
    return (x != 0) && (x & (x - 1)) == 0
}

func toPow2(x uint) uint {
    var y uint = 1
    for {
        x >>= 1
        if x == 0 {
            return y
        }
        y <<= 1
    }
}

func log2(x uint) uint {
    var y uint = 0
    for x > 1 {
        x >>= 1
        y++
    }
    return y
}

// -----------------------------------------------------------------------------

func main() {
    var duration uint
    var initial uint
    var num_threads uint
    var rng uint
    var update uint
    var put uint
    var load_factor uint
    var only_results bool

    { // Parameters
        flag.UintVar(&duration, "d", 1000, "Test duration in milliseconds")
        flag.UintVar(&initial, "i", 1024, "Number of elements to insert before test")
        flag.UintVar(&num_threads, "n", 4, "Number of threads")
        flag.UintVar(&rng, "r", 2048, "Range of integer values inserted in set")
        flag.UintVar(&update, "u", 20, "Percentage of update transactions")
        flag.UintVar(&put, "p", 10, "Percentage of put update transactions (should be less than percentage of updates)")
        flag.UintVar(&load_factor, "c", 1, "Load factor for the hash table")
        flag.UintVar(&share.Concurrency, "l", 512, "Concurrency level for the hash table")
        flag.UintVar(&share.NumBuckets, "b", 64, "Amount of buckets for the hash table")
        flag.BoolVar(&only_results, "o", false, "Only print operation latencies")
        flag.Parse()

        assert.Assert(num_threads > 0, "The amount of test threads should be a positive integer")

        if dataset.FindIsDef {
            assert.Assert(update <= 100, "The update rate should not be greater than 100 (it is a percentage)")
            if put > update {
                if !only_results {
                    fmt.Printf("** limiting put rate to update rate: old: %v / new: %v\n", put, update)
                }
                put = update
            }
        } else {
            assert.Assert(update != 0, "The update rate should not be null for a non-searchable dataset")
            if put > 100 {
                if !only_results {
                    fmt.Printf("** limiting put rate to update rate: old: %v / new: 100\n", put)
                }
                put = 100
            } else {
                put = put * 100 / update // Scale put too
            }
            update = 100
        }

        if !isPow2(initial) {
            temp := toPow2(initial)
            if !only_results {
                fmt.Printf("** rounding up initial (to make it power of 2): old: %v / new: %v\n", initial, temp)
            }
            initial = temp
        }
        share.Capacity = initial / load_factor
        share.LevelMax = log2(initial)
        if !isPow2(share.Concurrency) {
            temp := toPow2(share.Concurrency)
            if !only_results {
                fmt.Printf("** rounding up concurrency (to make it power of 2): old: %v / new: %v\n", share.Concurrency, temp)
            }
            share.Concurrency = temp
        }
        if rng < initial {
            rng = 2 * initial
        }
        if !only_results {
            fmt.Printf("## Initial: %v / Range: %v\n", initial, rng)
        }
        {
            var kb float64 = float64(initial) * float64(unsafe.Sizeof(uint(0))) / 1024
            var mb float64 = kb / 1024
            if !only_results {
                fmt.Printf("Sizeof initial: %.2f KB = %.2f MB\n", kb, mb)
            }
        }
        if !isPow2(rng) {
            temp := toPow2(rng)
            if !only_results {
                fmt.Printf("** rounding up range (to make it power of 2): old: %v / new: %v\n", rng, temp)
            }
            rng = temp
        }
    }

    var calibration float64 // "Net weight" latency

    { // Calibration to remove call overhead
        if !only_results {
            fmt.Printf("Net latency: ")
        }
        calibration = func() float64 {
            var avgs [calibrate_reps]float64 // Averages (from 'inner' loop)
            var gavg float64 // Average of avgs
            for rep := 0;; { // Measurements
                var avg float64 = 0
                start := time.Now()
                time.Since(start)
                for cnt := 0; cnt < calibrate_avgs; cnt++ {
                    start := time.Now()
                    delta := time.Since(start)
                    avg = float64(cnt) / float64(cnt + 1) * avg + float64(delta) / float64(cnt + 1)
                }
                avgs[rep] = avg
                gavg = float64(rep) / float64(rep + 1) * gavg + avg / float64(rep + 1)
                rep++
                if rep == calibrate_reps {
                    break
                }
                <-time.After(time.Duration(calibrate_pause) * time.Millisecond) // Just a quick pause
            }
            { // Return closest from average (arbitrary method: allow calibration to return roughly the same "net latency" across runs)
                closest := avgs[0]
                delta := math.Abs(gavg - closest)
                for i := 1; i < calibrate_reps; i++ {
                    closest_c := avgs[i]
                    delta_c := math.Abs(gavg - closest_c)
                    if delta_c < delta {
                        closest = closest_c
                        delta = delta_c
                    }
                }
                return closest
            }
        }()
        if !only_results {
            fmt.Printf("%.2f ns\n", calibration)
        }
    }

    set := dataset.New()
    var size uint

    { // DataSet initialization (kept while not found in test_simple.c)
        if !only_results {
            fmt.Printf("Adding %v entries to set...", initial)
        }
        for i := initial; i > 0; i-- {
            set.Insert(share.Key(i), 0)
        }
        size = set.Size()
        if !only_results {
            fmt.Printf(" done.\n")
        }
        assert.Assert(size == initial, fmt.Sprintf("Single-threaded set initialization failed: set size = %v", size))
    }

    var barrier sync.WaitGroup
    test := func(stats *stats_t) {
        var xorshf xorshift.State
        xorshf.Init()
        for volatile.ReadInt32(&running) != 0 {
            op := uint(xorshf.Intn(100))
            key := share.Key(xorshf.Intn(uint32(rng)) + 1)
            if (op < put) {
                start := time.Now()
                set.Insert(key, 0)
                stats.put_time += uint64(time.Since(start))
                stats.put_count++
            } else if (op < update) {
                start := time.Now()
                set.Delete(key)
                stats.remove_time += uint64(time.Since(start))
                stats.remove_count++
            } else {
                start := time.Now()
                set.Find(key)
                stats.get_time += uint64(time.Since(start))
                stats.get_count++
            }
        }
    }

    var put_count_total    uint64 = 0
    var put_time_total     uint64 = 0
    var get_count_total    uint64 = 0
    var get_time_total     uint64 = 0
    var remove_count_total uint64 = 0
    var remove_time_total  uint64 = 0

    { // Creating threads
        barrier.Add(1)
        if !only_results {
            fmt.Print("Creating threads: ")
        }
        for i := uint(0); i < num_threads; i++ {
            if !only_results {
                if i == 0 {
                    fmt.Print(i)
                } else {
                    fmt.Print(", ", i)
                }
            }
            thread.Spawn(func() {
                stats := new(stats_t)
                barrier.Wait()

                test(stats)

                // Global stats update
                atomic.AddUint64(&put_count_total, stats.put_count)
                atomic.AddUint64(&put_time_total, stats.put_time)
                atomic.AddUint64(&get_count_total, stats.get_count)
                atomic.AddUint64(&get_time_total, stats.get_time)
                atomic.AddUint64(&remove_count_total, stats.remove_count)
                atomic.AddUint64(&remove_time_total, stats.remove_time)
            })
        }
        if !only_results {
            fmt.Println()
        }
    }

    { // Running threads
        if !only_results {
            fmt.Println("*** RUNNING ***")
        }
        atomic.StoreInt32(&running, 1)
        barrier.Done() // Threads were waiting for it

        <-time.After(time.Duration(duration) * time.Millisecond) // Wait for duration

        atomic.StoreInt32(&running, 0)
        thread.WaitAll() // Wait for threads to update global statistics
        if !only_results {
            fmt.Println("*** STOPPED ***")
        }
    }

    { // Statistics correction
        get_time_corr := uint64(calibration * float64(get_count_total))
        put_time_corr := uint64(calibration * float64(put_count_total))
        remove_time_corr := uint64(calibration * float64(remove_count_total))
        if get_time_total > get_time_corr {
            get_time_total -= get_time_corr
        } else {
            get_time_total = 0
        }
        if put_time_total > put_time_corr {
            put_time_total -= put_time_corr
        } else {
            put_time_total = 0
        }
        if remove_time_total > remove_time_corr {
            remove_time_total -= remove_time_corr
        } else {
            remove_time_total = 0
        }
    }

    { // Print global statistics
        total := put_count_total + get_count_total + remove_count_total
        put_perc := 100.0 * (1 - (float64(total - put_count_total) / float64(total)))
        get_perc := 100.0 * (1 - (float64(total - get_count_total) / float64(total)))
        remove_perc := 100.0 * (1 - (float64(total - remove_count_total) / float64(total)))

        get_time_ms := float64(get_time_total) / 1000000
        put_time_ms := float64(put_time_total) / 1000000
        remove_time_ms := float64(remove_time_total) / 1000000

        get_lat := 1000 * get_time_ms / float64(get_count_total)
        put_lat := 1000 * put_time_ms / float64(put_count_total)
        remove_lat := 1000 * remove_time_ms / float64(remove_count_total)

        if only_results {
            fmt.Printf("%v\n%v\n%v\n", get_lat, put_lat, remove_lat)
        } else {
            fmt.Printf("    : %-10s | %-11s | %-10s | %s\n", "count", "% total", "time (ms)", "latency (µs/ops)")
            fmt.Printf("srch: %-10v | %10.1f%% | %10.1f | %16.3f\n", get_count_total, get_perc, get_time_ms, get_lat)
            fmt.Printf("insr: %-10v | %10.1f%% | %10.1f | %16.3f\n", put_count_total, put_perc, put_time_ms, put_lat)
            fmt.Printf("rems: %-10v | %10.1f%% | %10.1f | %16.3f\n", remove_count_total, remove_perc, remove_time_ms, remove_lat)
        }
    }

    set.Destroy()
}
