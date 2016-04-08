/**
 * @file   optik.go
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
 * Implementation of OPTIK-integer lock.
**/

package optik

import (
    "math"
    "sync/atomic"
    "time"
    "tools/volatile"
)

func pause() {
    cpause(1) /// FIXME: Better workaround ?
}

/// FIXME: Maybe find a better approximation of the C version (which executes 'cycles' times 'nop' in a simple loop)
func cpause(cycles time.Duration) {
    time.Sleep(cycles / 10) // time.Sleep is a _huge_ function in comparison with the C version 'cpause'
}

func optik_get_type_name() string {
    return "OPTIK-integer"
}

// -----------------------------------------------------------------------------

const (
    optik_deleted = math.MaxUint64
)

type Mutex uint64

func (ol *Mutex) iaf() Mutex {
    return Mutex(atomic.AddUint64((*uint64)(ol), 1))
}

func (ol *Mutex) daf() Mutex {
    return Mutex(atomic.AddUint64((*uint64)(ol), ^uint64(0)))
}

func (ol *Mutex) cas(old Mutex, new Mutex) bool {
    return atomic.CompareAndSwapUint64((*uint64)(ol), uint64(old), uint64(new))
}

// -----------------------------------------------------------------------------

func Is_locked(mutex Mutex) bool {
    return mutex & 1 == 1
}

func (ol *Mutex) Get_version_wait() Mutex {
    for {
        olv := Mutex(volatile.ReadUint64((*uint64)(ol)))
        if !Is_locked(olv) {
            return olv
        }
        cpause(128)
    }
}

func Is_deleted(ol Mutex) bool {
    return ol == optik_deleted
}

func Get_version(ol Mutex) uint32 {
    return uint32(ol)
}

func Get_n_locked(ol Mutex) uint32 {
    return uint32(ol >> 1)
}

func (ol *Mutex) Init() {
    *ol = 0
}

func Is_same_version(v1 Mutex, v2 Mutex) bool {
    return v1 == v2
}

func (ol *Mutex) TryLock_version(ol_old Mutex) bool {
    if Is_locked(ol_old) || *ol != ol_old {
        return false
    }
    return ol.cas(ol_old, ol_old + 1)
}

func (ol *Mutex) TryLock_vdelete(ol_old Mutex) bool {
    if Is_locked(ol_old) || *ol != ol_old {
        return false
    }
    return ol.cas(ol_old, optik_deleted)
}

func (ol *Mutex) Lock() bool {
    var ol_old Mutex
    for {
        for {
            ol_old = *ol
            if !Is_locked(ol_old) {
                break
            }
            pause();
        }
        if ol.cas(ol_old, ol_old + 1) {
            break
        }
    }
    return true
}

func (ol *Mutex) Lock_backoff() bool {
    var ol_old Mutex
    for {
        for {
            ol_old = *ol
            if !Is_locked(ol_old) {
                break
            }
            cpause(128)
        }
        if ol.cas(ol_old, ol_old + 1) {
            break
        }
    }
    return true
}

func (ol *Mutex) Lock_version(ol_old Mutex) bool {
    var ol_cur Mutex
    for {
        for {
            ol_cur = *ol
            if !Is_locked(ol_cur) {
                break
            }
            pause()
        }
        if ol.cas(ol_cur, ol_cur + 1) {
            break
        }
    }
    return ol_cur == ol_old
}

func (ol *Mutex) Lock_version_backoff(ol_old Mutex) bool {
    var ol_cur Mutex
    for {
        for {
            ol_cur = *ol
            if !Is_locked(ol_cur) {
                break
            }
            cpause(128)
        }
        if ol.cas(ol_cur, ol_cur + 1) {
            break
        }
    }
    return ol_cur == ol_old
}

func (ol *Mutex) TryLock() bool {
    var ol_new Mutex = *ol
    if Is_locked(ol_new) {
        return false
    }
    return ol.cas(ol_new, ol_new + 1)
}

func (ol *Mutex) Unlock() {
    ol.iaf()
}

func (ol *Mutex) Unlockv() Mutex {
    return ol.iaf();
}

func (ol *Mutex) Revert() {
    ol.daf();
}
