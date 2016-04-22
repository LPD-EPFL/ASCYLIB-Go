/**
 * @file   linkedlist_harris_opt.go
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
 * A Pragmatic Implementation of Non-blocking Linked Lists,
 * Timothy L Harris,
 * DISC 2001.
 *
 * Optimized version
**/

package dataset

import (
    "sync/atomic"
    "tools/share"
    "unsafe"
)

// -----------------------------------------------------------------------------

type node struct {
    key share.Key
    val share.Val
    next *node
}

type DataSet struct {
    head *node
}

// -----------------------------------------------------------------------------

func is_marked_ref(i *node) bool {
    return (uintptr(unsafe.Pointer(i)) & 1) != 0
}

func get_unmarked_ref(w *node) *node {
    return (*node)(unsafe.Pointer(uintptr(unsafe.Pointer(w)) &^ 1))
}

func get_marked_ref(w *node) *node {
    return (*node)(unsafe.Pointer(uintptr(unsafe.Pointer(w)) | 1))
}

func physical_delete_right(left_node *node, right_node *node) bool {
    return atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&left_node.next)), unsafe.Pointer(right_node), unsafe.Pointer(get_unmarked_ref(right_node.next)))
}

// -----------------------------------------------------------------------------

func new_node(key share.Key, val share.Val, next *node) *node {
    node := new(node)
    node.key = key
    node.val = val
    node.next = next
    return node
}

func list_search(set *DataSet, key share.Key) (left_node *node, right_node *node) {
    left_node = set.head
    right_node = left_node.next
    for {
        if !is_marked_ref(right_node.next) {
            if right_node.key >= key {
                return
            }
            left_node = right_node
        } else {
            physical_delete_right(left_node, right_node)
        }
        right_node = get_unmarked_ref(right_node.next)
    }
}

// -----------------------------------------------------------------------------

func New() *DataSet {
    set := new(DataSet)
    max := new_node(share.KEY_MAX, 0, nil)
    min := new_node(share.KEY_MIN, 0, max)
    set.head = min
    return set
}

func (set *DataSet) Destroy() {
}

func (set *DataSet) Size() uint {
    var size uint = 0
    node := get_unmarked_ref(set.head.next) // We have at least 2 elements
    for get_unmarked_ref(node.next) != nil {
        if !is_marked_ref(node.next) {
            size++
        }
        node = get_unmarked_ref(node.next)
    }
    return size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    node := set.head.next
    for node.key < key {
        node = get_unmarked_ref(node.next)
    }
    if node.key == key && !is_marked_ref(node.next) {
        return node.val, true
    }
    return 0, false
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    for {
        left_node, right_node := list_search(set, key)
        if right_node.key == key {
            return false
        }
        node_add := new_node(key, val, right_node)
        if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&left_node.next)), unsafe.Pointer(right_node), unsafe.Pointer(node_add)) { // Try to swing left_node's unmarked next pointer to a new node
            return true
        }
    }
}

func (set *DataSet) Delete(key share.Key) (result share.Val, ok bool) {
    var left_node *node
    var right_node *node
    for {
        left_node, right_node = list_search(set, key)
        if right_node.key != key {
            return 0, false
        }
        unmarked_ref := get_unmarked_ref(right_node.next) // Try to mark right_node as logically deleted
        marked_ref := get_marked_ref(unmarked_ref)
        if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&right_node.next)), unsafe.Pointer(unmarked_ref), unsafe.Pointer(marked_ref)) {
            break
        }
    }
    result, ok = right_node.val, true
    physical_delete_right(left_node, right_node)
    return
}
