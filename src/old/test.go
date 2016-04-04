/**
 * @file   test.go
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
 * A Lazy Concurrent List-Based Set Algorithm,
 * S. Heller, M. Herlihy, V. Luchangco, M. Moir, W.N. Scherer III, N. Shavit
 * p.3-16, OPODIS 2005
 * lazy.c is part of ASCYLIB
**/

package main

import (
    "assert"
    "fmt"
    "linkedlist_lazy"
    "math/rand"
    "sync"
    "sync/atomic"
    "thread"
    "time"
    "unsafe"
)

// -----------------------------------------------------------------------------

// True if the tests are running
var running int32

// -----------------------------------------------------------------------------

// Thread run statistics
type stats_t struct {
    nb_add uint64
    nb_added uint64
    nb_remove uint64
    nb_removed uint64
    nb_contains uint64
    nb_found uint64
    nb_aborts uint64
    nb_aborts_locked_read uint64
    nb_aborts_locked_write uint64
    nb_aborts_validate_read uint64
    nb_aborts_validate_write uint64
    nb_aborts_validate_commit uint64
    nb_aborts_invalid_memory uint64
    max_retries uint64
}

// -----------------------------------------------------------------------------

func main() {
    duration := 1000
    initial := 1024
    nb_threads := 2
    rng := 2 * 1024
    seed := 0
    update := 20
    unit_tx := 2
    alternate := false
    effective := true
    verbose := true
    set := linkedlist_lazy.New()
    var barrier sync.WaitGroup

    var reads, effreads, updates, effupds uint64

    test := func(id int, stats *stats_t) {
        var val linkedlist_lazy.Key
        var last linkedlist_lazy.Key = -1
        unext := rand.Intn(100) < update

        barrier.Wait()

        for atomic.LoadInt32(&running) == 1 {
            if unext {
                if (last < 0) {
                    val = linkedlist_lazy.Key(rand.Intn(rng))
                    if set.Insert(val, linkedlist_lazy.Val(val)) {
                        stats.nb_added++
                        last = val
                    }
                    stats.nb_add++
                } else {
                    if alternate {
                        _, ok := set.Delete(last)
                        if ok {
                            stats.nb_removed++
                        }
                        last = -1
                    } else {
                        val = linkedlist_lazy.Key(rand.Intn(rng) + 1)
                        _, ok := set.Delete(val)
                        if ok {
                            stats.nb_removed++
                            last = -1
                        }
                    }
                    stats.nb_remove++
                }
            } else {
                if alternate {
                    if update == 0 {
                        if (last < 0) {
                            val = 0 // first always equals 0 (see test.c)
                            last = val
                        } else {
                            val = linkedlist_lazy.Key(rand.Intn(rng) + 1)
                            last = -1
                        }
                    } else {
                        if last < 0 {
                            val = linkedlist_lazy.Key(rand.Intn(rng) + 1)
                        } else {
                            val = last
                        }
                    }
                } else {
                    val = linkedlist_lazy.Key(rand.Intn(rng) + 1)
                }
                if set.Has(val) {
                    stats.nb_found++
                }
                stats.nb_contains++
            }
            if effective {
                unext = ((100 * (stats.nb_added + stats.nb_removed)) < (uint64(update) * (stats.nb_add + stats.nb_remove + stats.nb_contains)))
            } else {
                unext = rand.Intn(100) < update
            }
        }
    }

    assert.Assert(duration >= 0, "Test duration should be a non-negative integer")
    assert.Assert(initial >= 0, "Initial set size should be a non-negative integer")
    assert.Assert(nb_threads > 0, "The amount of test threads should be a positive integer")
    assert.Assert(rng > 0 && rng >= initial, "The value range should be both a positive integer and greater than or equal to the initial set size")
    assert.Assert(update >= 0 && update <= 100, "The update rate should be both a positive integer and less than or equal to 100")

    fmt.Println("Set type     : linked list")
    fmt.Println("Duration     :", duration, "ms")
    fmt.Println("Initial size :", initial)
    fmt.Println("Nb threads   :", nb_threads)
    fmt.Println("Value range  :", rng)
    fmt.Println("Seed         :", seed)
    fmt.Println("Update rate  :", update)
    fmt.Println("Lock alg     :", unit_tx)
    fmt.Println("Alternate    :", alternate)
    fmt.Println("Effective    :", effective)
    fmt.Println("Type sizes   : int =", unsafe.Sizeof(int), "/ ptr =", unsafe.Sizeof(*interface{}), "/ word =", unsafe.Sizeof(uintptr))

    var size int64
    rand.Seed(time.Now().UnixNano())

    { // IntSet initialization
        tens := 1
        ten_perc := initial / 10
        ten_perc_nxt := ten_perc
        fmt.Printf("Adding %d entries to set\n", initial)
        if initial < 10000 {
            i := 0
            for i < initial {
                val := linkedlist_lazy.Key(rand.Intn(rng) + 1)
                if set.Insert(val, 0) {
                    if i == ten_perc_nxt {
                        fmt.Printf("\r%02d%%  ", tens * 10)
                        tens++
                        ten_perc_nxt = tens * ten_perc
                    }
                    i++
                }
            }
        } else {
            for i := initial; i > 0; i-- {
                set.Insert(linkedlist_lazy.Key(i), 0)
            }
        }
        fmt.Printf("\n")
        size = int64(set.Size())
        fmt.Printf("Set size     : %d\n", size)
    }

    { // Creating threads
        barrier.Add(1)
        fmt.Print("Creating threads: ")
        for i := 0; i < nb_threads; i++ {
            if i == 0 {
                fmt.Print(i)
            } else {
                fmt.Print(", ", i)
            }
            id := i
            thread.Spawn(func() {
                barrier.Wait()
                stats := new(stats_t)
                test(id, stats)
                { // Global stats update
                    atomic.AddUint64(&reads, stats.nb_contains)
                    atomic.AddUint64(&effreads, stats.nb_contains + (stats.nb_add - stats.nb_added) + (stats.nb_remove - stats.nb_removed))
                    atomic.AddUint64(&updates, (stats.nb_add + stats.nb_remove))
                    atomic.AddUint64(&effupds, stats.nb_removed + stats.nb_added)
                    atomic.AddInt64(&size, int64(stats.nb_added) - int64(stats.nb_removed))
                }
                if verbose {
                    fmt.Printf("Thread %v:\n  #add        : %v\n    #added    : %v\n  #remove     : %v\n    #removed  : %v\n  #contains   : %v\n  #found      : %v\n", id, stats.nb_add, stats.nb_added, stats.nb_remove, stats.nb_removed, stats.nb_contains, stats.nb_found)
                }
            })
        }
        fmt.Println()
    }

    var actual_duration float64 // Actual test duration (in ms)

    { // Running threads
        fmt.Println("*** RUNNING... ***")
        atomic.StoreInt32(&running, 1)
        start_time := time.Now()
        barrier.Done() // Threads were waiting for it
        <-time.After(time.Duration(duration) * time.Millisecond) // Wait for duration
        atomic.StoreInt32(&running, 0)
        actual_duration = float64(time.Since(start_time).Nanoseconds()) * float64(time.Nanosecond) / float64(time.Millisecond)
        thread.WaitAll() // Wait for threads to print and update statistics
        fmt.Println("*** STOPPED ***")
    }

    { // Print global statistics
        size_after := int64(set.Size())
        fmt.Printf("Set size      : %d (expected: %d)\n", size_after, size)
        assert.Assert(size_after == size, "The set size has changed !!")
        fmt.Printf("Duration      : %f ms\n", actual_duration)
        fmt.Printf("#txs          : %v (%f / s)\n", reads + updates, float64(reads + updates) * 1000.0 / actual_duration)

        fmt.Print("#read txs     : ")
        if effective {
            fmt.Printf("%v (%f / s)\n", effreads, float64(effreads) * 1000.0 / actual_duration)
            fmt.Printf("  #contains   : %v (%f / s)\n", reads, float64(reads) * 1000.0 / actual_duration)
        } else {
            fmt.Printf("%v (%f / s)\n", reads, float64(reads) * 1000.0 / actual_duration)
        }

        fmt.Printf("#eff. upd rate: %f \n", 100.0 * float64(effupds) / float64(effupds + effreads))

        fmt.Printf("#update txs   : ")
        if effective {
            fmt.Printf("%v (%f / s)\n", effupds, float64(effupds) * 1000.0 / actual_duration)
            fmt.Printf("  #upd trials : %v (%f / s)\n", updates, float64(updates) * 1000.0 / actual_duration)
        } else {
            fmt.Printf("%v (%f / s)\n", updates, float64(updates) * 1000.0 / actual_duration)
        }
    }
}
