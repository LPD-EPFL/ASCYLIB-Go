/**
 * @file   base.go
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
 * ...
**/

package dataset

import (
    "fmt"
    "tools/share"
)

const (
    FindIsDef bool = true
)

// -----------------------------------------------------------------------------

type DataSet struct {
}

// -----------------------------------------------------------------------------

func New() *DataSet {
    fmt.Println("Please implement me!")
    return new(DataSet)
}

func (set *DataSet) Destroy() {
}

func (set *DataSet) Size() uint {
    return 0
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    return 0, false
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    return false
}

func (set *DataSet) Delete(key share.Key) (share.Val, bool) {
    return 0, false
}
