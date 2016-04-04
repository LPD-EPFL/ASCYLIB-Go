/**
 * @file   thread.go
 * @author Sébastien Rouault <sebastien.rouault@epfl.ch>
 *
 * @section LICENSE
 *
 * Copyright (c) 2014 Sébastien Rouault <sebastien.rouault@epfl.ch>
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
 * Simple thread manipulation in Go.
**/

package thread

import (
    "math"
    "runtime"
    "sync"
    "sync/atomic"
    "syscall"
    "unsafe"
)

// -----------------------------------------------------------------------------

// Entry point of a goroutine "locked" on a thread
type EntryPoint func()

// Mutex to change max procs
var lock sync.Mutex

// Wait group for spawned threads
var wait sync.WaitGroup

// -----------------------------------------------------------------------------

// Core affinity
var core_id uintptr
var core_cnt uintptr = uintptr(runtime.NumCPU())

// Data structure to describe CPU mask
const cpu_set_size uintptr = 1024 // Size of the set, in bytes
const cpu_set_div uintptr = 8 // = sizeof(uint64)
const cpu_set_length uintptr = cpu_set_size / cpu_set_div
type cpu_set_t struct {
    bits [cpu_set_length]uint64
}

/** Set thread affinity to the next CPU.
**/
func set_next_cpu() {
    var cpu_set cpu_set_t
    var my_core_id = (atomic.AddUintptr(&core_id, 1) - 1) % core_cnt // Atomic add
    var pos_big uintptr = my_core_id / (cpu_set_div * 8) // 8 * sizeof(uint64) CPU id per division
    var pos_small uintptr = my_core_id % (cpu_set_div * 8)
    cpu_set.bits[pos_big] = 1 << pos_small
    res, _, _ := syscall.Syscall(syscall.SYS_SCHED_SETAFFINITY, 0, cpu_set_size, uintptr(unsafe.Pointer(&(cpu_set.bits))))
    if res != 0 {
        panic("System call failed")
    }
}

/** Set thread affinity to all CPU.
**/
func set_all_cpu() {
    var cpu_set cpu_set_t
    var i uintptr
    for i = 0; i < cpu_set_length; i++ {
        cpu_set.bits[i] = math.MaxUint64
    }
    res, _, _ := syscall.Syscall(syscall.SYS_SCHED_SETAFFINITY, 0, cpu_set_size, uintptr(unsafe.Pointer(&(cpu_set.bits))))
    if res != 0 {
        panic("System call failed")
    }
}

// -----------------------------------------------------------------------------

/** Start a new goroutine, and (try to) "lock" it on a (new) thread.
 * @param ep Function to call
**/
func Spawn(ep EntryPoint) {
    lock.Lock()
    runtime.GOMAXPROCS(runtime.GOMAXPROCS(0) + 1) // To try and have a new thread for the goroutine
    lock.Unlock()
    wait.Add(1)
    go (func() {
        runtime.LockOSThread()
        set_next_cpu()
        ep()
        set_all_cpu()
        wait.Done()
        lock.Lock()
        runtime.GOMAXPROCS(runtime.GOMAXPROCS(0) - 1)
        lock.Unlock()
    })()
}

/** Wait for all the spawned threads to terminate.
**/
func WaitAll() {
    wait.Wait()
}
