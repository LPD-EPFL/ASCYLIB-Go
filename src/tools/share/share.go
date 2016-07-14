/**
 * @file   share.go
 * @author Sébastien Rouault <sebastien.rouault@epfl.ch>
 *
 * @section LICENSE
 *
 * Copyright (c) 2016 Sébastien Rouault <sebastien.rouault@epfl.ch>
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
 * Shared pieces of code between test module and data structures.
**/

package share

import (
    "math"
)

// -----------------------------------------------------------------------------

// Common types
type Key int64
type Val int64

const (
    KEY_MIN = math.MinInt64
    KEY_MAX = math.MaxInt64
)

// Global variables
var Capacity uint
var Concurrency uint
var NumBuckets uint
var LevelMax uint
