/**
 * @file   xorshift.go
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
 * Implementation of the Marsaglia's xorshf96 generator.
**/

package xorshift

import (
    "math/rand"
    "time"
)

// -----------------------------------------------------------------------------

type State struct {
    x uint32
    y uint32
    z uint32
}

// -----------------------------------------------------------------------------

func (state *State) Init() {
    r := rand.New(rand.NewSource(time.Now().UnixNano())) // Undefined behavior, but we only need a small source of entropy here...
    state.x = r.Uint32()
    state.y = r.Uint32()
    state.z = r.Uint32()
}

func (state *State) Intn(n uint32) uint32 {
    state.x ^= state.x << 16
    state.x ^= state.x >> 5
    state.x ^= state.x << 1
    t := state.x
    state.x = state.y
    state.y = state.z
    state.z = t ^ state.x ^ state.y
    return state.z % n
}
