/**
 * @file   hashtable_copy.go
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
 * Similar to Java's CopyOnWriteArrayList. One array per bucket.
 * http://docs.oracle.com/javase/7/docs/api/java/util/concurrent/CopyOnWriteArrayList.html
**/

package dataset

import (
    "sync"
    "tools/share"
)

const (
    read_only_fail bool = true
)

// -----------------------------------------------------------------------------

type keyval struct {
    key share.Key
    val share.Val
}

type array struct {
    size uint
    table []keyval
}

type DataSet struct {
    num_buckets uint
    hash uint
    lock []sync.Mutex
    arrays []*array
}

// -----------------------------------------------------------------------------

func new_array(size uint) *array {
    array := new(array)
    array.size = size
    array.table = make([]keyval, size)
    return array
}

func (all_cur *array) cpy_array_search(key share.Key) bool {
    for i := uint(0); i < all_cur.size; i++ {
        if all_cur.table[i].key == key {
            return true
        }
    }
    return false
}

// -----------------------------------------------------------------------------

func New() *DataSet {
    set := new(DataSet)
    set.num_buckets = share.NumBuckets
    set.hash = set.num_buckets - 1
    set.lock = make([]sync.Mutex, share.NumBuckets)
    set.arrays = make([]*array, share.NumBuckets)
    for i := uint(0); i < set.num_buckets; i++ {
        set.arrays[i] = new_array(0)
    }
    return set
}

func (set *DataSet) Size() uint {
    var s uint = 0
    for i := uint(0); i < set.num_buckets; i++ {
        s += set.arrays[i].size
    }
    return s
}

func (set *DataSet) Find(key share.Key) (res share.Val, ok bool) {
    all_cur := set.arrays[uint(key) & set.hash]
    for i := uint(0); i < all_cur.size; i++ {
        if all_cur.table[i].key == key {
            return all_cur.table[i].val, true
        }
    }
    return 0, false
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    bucket := uint(key) & set.hash
    var all_old *array

    if read_only_fail {
        all_old = set.arrays[bucket]
        if all_old.cpy_array_search(key) {
            return false
        }
    }
    set.lock[bucket].Lock()
    defer set.lock[bucket].Unlock()

    all_old = set.arrays[bucket]
    all_new := new_array(all_old.size + 1)
    var i uint
    for i = 0; i < all_old.size; i++ {
        if all_old.table[i].key == key {
            return false
        }
        all_new.table[i].key = all_old.table[i].key
        all_new.table[i].val = all_old.table[i].val
    }
    all_new.table[i].key = key
    all_new.table[i].val = val
    set.arrays[bucket] = all_new

    return true
}

func (set *DataSet) Delete(key share.Key) (result share.Val, ok bool) {
    bucket := uint(key) & set.hash
    var all_old *array

    result = 0
    ok = false

    if read_only_fail {
        all_old = set.arrays[bucket]
        if !all_old.cpy_array_search(key) {
            return
        }
    }

    set.lock[bucket].Lock()
    defer set.lock[bucket].Unlock()
    all_old = set.arrays[bucket]
    all_new := new_array(all_old.size - 1)

    var i, n uint = 0, 0
    for ; i < all_old.size; i++ {
        if all_old.table[i].key == key {
            result = all_old.table[i].val
            ok = true
        } else {
            all_new.table[n].key = all_old.table[i].key
            all_new.table[n].val = all_old.table[i].val
            n++
        }
    }

    if ok {
        set.arrays[bucket] = all_new
    }

    return
}
