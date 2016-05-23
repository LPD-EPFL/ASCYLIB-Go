/**
 * @file   ttas.go
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
 * Test and test-and-set.
**/

package ttas

import (
    "runtime"
    "sync/atomic"
    "tools/volatile"
)

const (
    cnt_gosched uint = 1024 // How many loops before a call to Gosched()
)

// -----------------------------------------------------------------------------

type Mutex struct {
    state uint32
}

// -----------------------------------------------------------------------------

func (m *Mutex) TryLock() bool {
    return atomic.CompareAndSwapUint32(&m.state, 0, 1)
}

func (m *Mutex) Lock() {
    var i uint = 0 // Counter for Gosched()
    for {
        for { // Wait unlocked state
            if volatile.ReadUint32(&m.state) == 0 {
                break
            }
            i = i + 1
            if i % cnt_gosched == 0 {
                runtime.Gosched()
            }
        }
        if atomic.CompareAndSwapUint32(&m.state, 0, 1) {
            break
        }
    }
}

func (m *Mutex) Unlock() {
    atomic.StoreUint32(&m.state, 0)
}
