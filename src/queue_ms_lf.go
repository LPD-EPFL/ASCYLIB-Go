/**
 * @file   queue_ms_lf.go
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
 * A simple lock-free queue.
**/

package dataset

import (
    "runtime"
    "sync/atomic"
    "tools/share"
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
    head *node
    tail *node
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
    set := new(DataSet)
    node := new_node(0, 0, nil)
    set.head = node
    set.tail = node
    return set
}

func (set *DataSet) Destroy() {
}

func (set *DataSet) Size() uint {
    size := uint(0)
    node := set.head
    for node.next != nil {
        size++
        node = node.next
    }
    return size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    return 0, true
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    elem := new_node(key, val, nil)
    var tail *node
    for {
        tail = set.tail
        next := tail.next
        if tail == set.tail {
            if next == nil {
                if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&tail.next)), unsafe.Pointer(next), unsafe.Pointer(elem)) {
                    break
                }
            } else {
                atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&set.tail)), unsafe.Pointer(tail), unsafe.Pointer(next))
            }
        }
        runtime.Gosched()
    }
    atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&set.tail)), unsafe.Pointer(tail), unsafe.Pointer(elem))
    return true
}

func (set *DataSet) Delete(key share.Key) (share.Val, bool) {
    var next *node
    for {
        head := set.head
        tail := set.tail
        next = head.next
        if head == set.head {
            if head == tail {
                if next == nil {
                    return 0, false
                }
                atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&set.tail)), unsafe.Pointer(tail), unsafe.Pointer(next))
            } else {
                if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&set.head)), unsafe.Pointer(head), unsafe.Pointer(next)) {
                    break
                }
            }
        }
        runtime.Gosched()
    }
    return next.val, true
}
