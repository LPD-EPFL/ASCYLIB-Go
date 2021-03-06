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
 *
 * Simple test module.
**/

package main

import (
    "dataset"
    "flag"
    "fmt"
    "strconv"
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

    { // Parameters
        flag.UintVar(&duration, "d", 1000, "Test duration in milliseconds")
        flag.UintVar(&initial, "i", 1024, "Number of elements to insert before test")
        flag.UintVar(&num_threads, "n", 1, "Number of threads")
        flag.UintVar(&rng, "r", 2048, "Range of integer values inserted in set")
        flag.UintVar(&update, "u", 20, "Percentage of update transactions")
        flag.UintVar(&put, "p", 10, "Percentage of put update transactions (should be less than percentage of updates)")
        flag.UintVar(&load_factor, "c", 1, "Load factor for the hash table")
        flag.UintVar(&share.Concurrency, "l", 512, "Concurrency level for the hash table")
        flag.UintVar(&share.NumBuckets, "b", 64, "Amount of buckets for the hash table")
        flag.Parse()

        assert.Assert(num_threads > 0, "The amount of test threads should be a positive integer")

        if dataset.FindIsDef {
            assert.Assert(update <= 100, "The update rate should not be greater than 100 (it is a percentage)")
            if put > update {
                fmt.Printf("** limiting put rate to update rate: old: %v / new: %v\n", put, update)
                put = update
            }
        } else {
            assert.Assert(update != 0, "The update rate should not be null for a non-searchable dataset")
            if put > 100 {
                fmt.Printf("** limiting put rate to update rate: old: %v / new: 100\n", put)
                put = 100
            } else {
                put = put * 100 / update // Scale put too
            }
            update = 100
        }

        if !isPow2(initial) {
            temp := toPow2(initial)
            fmt.Printf("** rounding up initial (to make it power of 2): old: %v / new: %v\n", initial, temp)
            initial = temp
        }
        share.Capacity = initial / load_factor
        share.LevelMax = log2(initial)
        if !isPow2(share.Concurrency) {
            temp := toPow2(share.Concurrency)
            fmt.Printf("** rounding up concurrency (to make it power of 2): old: %v / new: %v\n", share.Concurrency, temp)
            share.Concurrency = temp
        }
        if rng < initial {
            rng = 2 * initial
        }
        fmt.Printf("## Initial: %v / Range: %v\n", initial, rng)
        {
            var kb float64 = float64(initial) * float64(unsafe.Sizeof(uint(0))) / 1024
            var mb float64 = kb / 1024
            fmt.Printf("Sizeof initial: %.2f KB = %.2f MB\n", kb, mb)
        }
        if !isPow2(rng) {
            temp := toPow2(rng)
            fmt.Printf("** rounding up range (to make it power of 2): old: %v / new: %v\n", rng, temp)
            rng = temp
        }
    }

    set := dataset.New()
    var size uint

    { // DataSet initialization (kept while not found in test_simple.c)
        fmt.Printf("Adding %v entries to set...", initial)
        for i := initial; i > 0; i-- {
            set.Insert(share.Key(i), 0)
        }
        size = set.Size()
        fmt.Printf(" done.\n")
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
                _, ok := set.Find(key)
                if ok {
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
            thread.Spawn(func() {
                stats := new(stats_t)
                barrier.Wait()

                test(stats)

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
        { // Assert set size
            ssize := set.Size()
            wsize := uint(int64(initial) + int64(putting_count_total_succ) - int64(removing_count_total_succ))
            assert.Assert(wsize == ssize, "WRONG set size: " + strconv.Itoa(int(ssize)) + " instead of " + strconv.Itoa(int(wsize)))
        }

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
        fmt.Printf("#txs %v\t(%-10.0f\n", num_threads, throughput)
        fmt.Printf("#Mops %.3f\n", throughput / 1e6)
    }

    set.Destroy()
}
