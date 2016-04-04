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
    "test/prototype"
)

// -----------------------------------------------------------------------------

type DataSet struct {
}

// -----------------------------------------------------------------------------

func New() *DataSet {
    fmt.Println("Please implement me!")
    return new(DataSet)
}

func (set *DataSet) Size() uint {
    return 0
}

func (set *DataSet) Has(res prototype.Key) bool {
    _, ok := set.Find(res)
    return ok
}

func (set *DataSet) Find(key prototype.Key) (res prototype.Val, ok bool) {
    res, ok = 0, false
    return
}

func (set *DataSet) Insert(key prototype.Key, val prototype.Val) bool {
    return false
}

func (set *DataSet) Delete(key prototype.Key) (result prototype.Val, ok bool) {
    result, ok = 0, false
    return
}
