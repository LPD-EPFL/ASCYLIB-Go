/**
 * @file   simple.go
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
**/

package main

import (
    "dataset"
    "flag"
    "fmt"
    "math/rand"
    "sync"
    "sync/atomic"
    "test/prototype"
    "time"
    "tools/assert"
    "tools/thread"
    "unsafe"
)

// -----------------------------------------------------------------------------

// True if the tests are running
var running int32

// Thread run statistics
type stats_t struct {
    putting_count uint64
    putting_count_succ uint64
    getting_count uint64
    getting_count_succ uint64
    removing_count uint64
    removing_count_succ uint64
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

// -----------------------------------------------------------------------------

func main() {
    rand.Seed(time.Now().UnixNano())

    var duration uint
    var initial uint
    var num_threads uint
    var rng uint
    var update uint
    var put uint

    var update_rate, put_rate, get_rate float64

    { // Parameters
        flag.UintVar(&duration, "d", 1000, "Test duration in milliseconds")
        flag.UintVar(&initial, "i", 1024, "Number of elements to insert before test")
        flag.UintVar(&num_threads, "n", 1, "Number of threads")
        flag.UintVar(&rng, "r", 2048, "Range of integer values inserted in set")
        flag.UintVar(&update, "u", 20, "Percentage of update transactions")
        flag.UintVar(&put, "p", 10, "Percentage of put update transactions (should be less than percentage of updates)")
        flag.Parse()

        if put > update {
            fmt.Printf("** limiting put rate to update rate: old: %v / new: %v\n", put, update)
            put = update
        }

        assert.Assert(num_threads > 0, "The amount of test threads should be a positive integer")
        assert.Assert(update <= 100, "The update rate should not be greater than 100 (it is a percentage)")

        if !isPow2(initial) {
            temp := toPow2(initial)
            fmt.Printf("** rounding up initial (to make it power of 2): old: %v / new: %v\n", initial, temp)
            initial = temp
        }
        if rng < initial {
            rng = 2 * initial
        }
        fmt.Printf("## Initial: %v / Range: %v\n", initial, rng)
        {
            var kb float64 = float64(initial) * float64(unsafe.Sizeof(uint)) / 1024
            var mb float64 = kb / 1024
            fmt.Printf("Sizeof initial: %.2f KB = %.2f MB\n", kb, mb)
        }
        if !isPow2(rng) {
            temp := toPow2(rng)
            fmt.Printf("** rounding up range (to make it power of 2): old: %v / new: %v\n", rng, temp)
            rng = temp
        }
        update_rate = float64(update) / 100
        if put >= 0 {
            put_rate = float64(put) / 100
        } else {
            put_rate = float64(update_rate) / 2
        }
        get_rate = 1 - update_rate
    }

    set := dataset.New()
    var size uint

    { // DataSet initialization (kept while not found in test_simple.c)
        fmt.Printf("Adding %v entries to set...", initial)
        for i := initial; i > 0; i-- {
            set.Insert(prototype.Key(i), 0)
        }
        size = set.Size()
        fmt.Printf(" done.\n")
        assert.Assert(size == initial, fmt.Sprintf("Single-threaded set initialization failed: set size = %v", size))
    }

    var barrier sync.WaitGroup
    test := func(id uint, stats *stats_t) {
        for running != 0 { /// FIXME: Check behavior
            op := uint(rand.Intn(100))
            key := prototype.Key(rand.Intn(int(rng)) + 1)
            if (op < put) {
                if set.Insert(key, 0) {
                    stats.putting_count_succ++
                }
                stats.putting_count++
            } else if (op < update) {
                _, ok := set.Delete(key)
                if ok {
                    stats.removing_count_succ++
                }
                stats.removing_count++
            } else {
                if set.Has(key) {
                    stats.getting_count_succ++
                }
                stats.getting_count++
            }
        }
    }

    var putting_count_total uint64 = 0
    var putting_count_total_succ uint64 = 0
    var getting_count_total uint64 = 0
    var getting_count_total_succ uint64 = 0
    var removing_count_total uint64 = 0
    var removing_count_total_succ uint64 = 0

    { // Creating threads
        barrier.Add(1)
        fmt.Print("Creating threads: ")
        for i := uint(0); i < num_threads; i++ {
            if i == 0 {
                fmt.Print(i)
            } else {
                fmt.Print(", ", i)
            }
            id := i
            thread.Spawn(func() {
                stats := new(stats_t)
                barrier.Wait()

                test(id, stats)

                // Global stats update
                atomic.AddUint64(&putting_count_total, stats.putting_count)
                atomic.AddUint64(&putting_count_total_succ, stats.putting_count_succ)
                atomic.AddUint64(&getting_count_total, stats.getting_count)
                atomic.AddUint64(&getting_count_total_succ, stats.getting_count_succ)
                atomic.AddUint64(&removing_count_total, stats.removing_count)
                atomic.AddUint64(&removing_count_total_succ, stats.removing_count_succ)
            })
        }
        fmt.Println()
    }

    var actual_duration float64 // Actual test duration (in ms)

    { // Running threads
        fmt.Println("*** RUNNING ***")
        atomic.StoreInt32(&running, 1)
        start_time := time.Now()
        barrier.Done() // Threads were waiting for it

        <-time.After(time.Duration(duration) * time.Millisecond) // Wait for duration

        atomic.StoreInt32(&running, 0)
        actual_duration = float64(time.Since(start_time).Nanoseconds()) * float64(time.Nanosecond) / float64(time.Millisecond)
        thread.WaitAll() // Wait for threads to update global statistics
        fmt.Println("*** STOPPED ***")
    }

    { // Print global statistics
        assert.Assert(initial + uint(int64(putting_count_total_succ) - int64(removing_count_total_succ)) == set.Size(), "WRONG set size")

        total := putting_count_total + getting_count_total + removing_count_total
        putting_perc := 100.0 * (1 - (float64(total - putting_count_total) / float64(total)))
        putting_perc_succ := (1 - float64(putting_count_total - putting_count_total_succ) / float64(putting_count_total)) * 100
        getting_perc := 100.0 * (1 - (float64(total - getting_count_total) / float64(total)))
        getting_perc_succ := (1 - float64(getting_count_total - getting_count_total_succ) / float64(getting_count_total)) * 100
        removing_perc := 100.0 * (1 - (float64(total - removing_count_total) / float64(total)))
        removing_perc_succ := (1 - float64(removing_count_total - removing_count_total_succ) / float64(removing_count_total)) * 100

        fmt.Printf("    : %-10s | %-10s | %-11s | %-11s | %s\n", "total", "success", "succ %", "total %", "effective %")
        fmt.Printf("srch: %-10v | %-10v | %10.1f%% | %10.1f%% | \n", getting_count_total, getting_count_total_succ, getting_perc_succ, getting_perc)
        fmt.Printf("insr: %-10v | %-10v | %10.1f%% | %10.1f%% | %10.1f%%\n", putting_count_total, putting_count_total_succ, putting_perc_succ, putting_perc, (putting_perc * putting_perc_succ) / 100)
        fmt.Printf("rems: %-10v | %-10v | %10.1f%% | %10.1f%% | %10.1f%%\n", removing_count_total, removing_count_total_succ, removing_perc_succ, removing_perc, (removing_perc * removing_perc_succ) / 100)

        throughput := float64(putting_count_total + getting_count_total + removing_count_total) * 1000.0 / actual_duration
        fmt.Printf("#txs %v\t(%-10.0f)\n", num_threads, throughput)
        fmt.Printf("#Mops %.3f\n", throughput / 1e6)
    }
}
