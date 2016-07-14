/**
 * @file   trace.go
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
 * Test module which traces execution.
**/

package main

import (
    "dataset"
    "flag"
    "fmt"
    "os"
    "runtime/trace"
    "sync"
    "sync/atomic"
    "tools/share"
    "time"
    "tools/assert"
    "tools/thread"
    "tools/volatile"
    "tools/xorshift"
)

// -----------------------------------------------------------------------------

// True if the tests are running
var running int32

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
                put = update
            }
        } else {
            assert.Assert(update != 0, "The update rate should not be null for a non-searchable dataset")
            if put > 100 {
                put = 100
            } else {
                put = put * 100 / update // Scale put too
            }
            update = 100
        }

        if !isPow2(initial) {
            temp := toPow2(initial)
            initial = temp
        }
        share.Capacity = initial / load_factor
        share.LevelMax = log2(initial)
        if !isPow2(share.Concurrency) {
            temp := toPow2(share.Concurrency)
            share.Concurrency = temp
        }
        if rng < initial {
            rng = 2 * initial
        }
        if !isPow2(rng) {
            temp := toPow2(rng)
            rng = temp
        }
    }

    set := dataset.New()
    var size uint

    { // DataSet initialization (kept while not found in test_simple.c)
        for i := initial; i > 0; i-- {
            set.Insert(share.Key(i), 0)
        }
        size = set.Size()
        assert.Assert(size == initial, fmt.Sprintf("Single-threaded set initialization failed: set size = %v", size))
    }

    var barrier sync.WaitGroup
    test := func(id uint) {
        var xorshf xorshift.State
        xorshf.Init()
        for volatile.ReadInt32(&running) != 0 {
            op := uint(xorshf.Intn(100))
            key := share.Key(xorshf.Intn(uint32(rng)) + 1)
            if (op < put) {
                set.Insert(key, 0)
            } else if (op < update) {
                set.Delete(key)
            } else {
                set.Find(key)
            }
        }
    }

    { // Creating threads
        barrier.Add(1)
        for i := uint(0); i < num_threads; i++ {
            id := i
            thread.Spawn(func() {
                barrier.Wait()
                test(id)
            })
        }
    }

    if err := trace.Start(os.Stdout); err != nil { // Begin tracing
        fmt.Println(err.Error())
        set.Destroy()
        return
    }

    { // Running threads
        atomic.StoreInt32(&running, 1)
        barrier.Done() // Threads were waiting for it

        <-time.After(time.Duration(duration) * time.Millisecond) // Wait for duration

        atomic.StoreInt32(&running, 0)
        thread.WaitAll() // Wait for threads to update global statistics
    }

    trace.Stop() // End tracing

    set.Destroy()
}
