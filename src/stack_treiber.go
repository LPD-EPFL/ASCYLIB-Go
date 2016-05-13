/**
 * @file   stack_treiber.go
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
 * Treiber's concurrent stack.
**/

package dataset

import (
    "runtime"
    "sync/atomic"
    "tools/share"
    "tools/volatile"
    "unsafe"
)

const (
    FindIsDef bool = false
)

// -----------------------------------------------------------------------------

type node struct {
    key share.Key
    val share.Val
    next *node
}

type DataSet struct {
    top *node
}

// -----------------------------------------------------------------------------

func new_node(key share.Key, val share.Val, next *node) *node {
    elem := new(node)
    elem.key = key
    elem.val = val
    elem.next = next
    return elem
}

// -----------------------------------------------------------------------------

func New() *DataSet {
    return new(DataSet)
}

func (set *DataSet) Destroy() {
}

func (set *DataSet) Size() uint {
    size := uint(0)
    node := set.top
    for node != nil {
        size++
        node = node.next
    }
    return size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    return 0, true // Not supposed to use Find with a stack...
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    elem := new_node(key, val, nil)
    for {
        top := (*node)(volatile.ReadPointer((*unsafe.Pointer)(unsafe.Pointer(&set.top))))
        elem.next = top
        if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&set.top)), unsafe.Pointer(top), unsafe.Pointer(elem)) {
            return true
        }
        runtime.Gosched()
    }
}

func (set *DataSet) Delete(key share.Key) (share.Val, bool) {
    for {
        top := (*node)(volatile.ReadPointer((*unsafe.Pointer)(unsafe.Pointer(&set.top))))
        if top == nil {
            return 0, false
        }
        if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&set.top)), unsafe.Pointer(top), unsafe.Pointer(top.next)) {
            return top.val, true
        }
        runtime.Gosched()
    }
}
