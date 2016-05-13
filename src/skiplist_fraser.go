/**
 * @file   skiplist_fraser.go
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
 * Lock-based skip list implementation of the Fraser algorithm
 * "Practical Lock Freedom", K. Fraser, PhD dissertation, September 2003.
**/

package dataset

import (
    "sync/atomic"
    "tools/share"
    "tools/xorshift"
    "unsafe"
)

const (
    FindIsDef bool = true
    fraser_max_level = uint(64)
)

// -----------------------------------------------------------------------------

type node struct {
    key share.Key
    val share.Val
    deleted uint32
    toplevel uint32
    next []*node
}

type DataSet struct {
    head *node
}

// -----------------------------------------------------------------------------

var state xorshift.State

func init_rand_level() {
    state.Init()
}

func get_rand_level() uint {
    level := uint(1)
    for i := uint(0); i < share.LevelMax - 1; i++ {
        if state.Intn(100) < 50 {
            level++
        } else {
            break
        }
    }
    return level
}

// -----------------------------------------------------------------------------

func is_marked(i *node) bool {
    return (uintptr(unsafe.Pointer(i)) & 1) != 0
}

func unset_mark(i *node) *node {
    return (*node)(unsafe.Pointer(uintptr(unsafe.Pointer(i)) &^ 1))
}

func set_mark(i *node) *node {
    return (*node)(unsafe.Pointer(uintptr(unsafe.Pointer(i)) | 1))
}

// -----------------------------------------------------------------------------

func new_simple_node(key share.Key, val share.Val, toplevel uint32) *node {
    elem := new(node)
    elem.key = key
    elem.val = val
    elem.toplevel = toplevel
    elem.deleted = 0
    elem.next = make([]*node, share.LevelMax)
    return elem
}

func new_node(key share.Key, val share.Val, next *node, toplevel uint32) *node {
    node := new_simple_node(key, val, toplevel)
    for i := uint(0); i < share.LevelMax; i++ {
        node.next[i] = next
    }
    return node
}

// -----------------------------------------------------------------------------

func (set *DataSet) fraser_search(key share.Key, left_list []*node, right_list []*node) {
retry:
    left := set.head
    for i := int(share.LevelMax - 1); i >= 0; i-- {
        left_next := left.next[i]
        if is_marked(left_next) {
            goto retry
        }

        /* Find unmarked node pair at this level */
        var right, right_next *node
        for right = left_next;; right = right_next {
            /* Skip a sequence of marked nodes */
            right_next = right.next[i]
            for is_marked(right_next) {
                right = unset_mark(right_next)
                right_next = right.next[i]
            }
            if right.key >= key {
                break
            }
            left = right
            left_next = right_next
        }

        /* Ensure left and right nodes are adjacent */
        if left_next != right {
            if !atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&left.next[i])), unsafe.Pointer(left_next), unsafe.Pointer(right)) {
                goto retry
            }
        }

        if left_list != nil {
            left_list[i] = left
        }
        if right_list != nil {
            right_list[i] = right
        }
    }
}

func mark_node_ptrs(n *node) {
    for i := int(n.toplevel - 1); i >= 0; i-- {
        for {
            n_next := n.next[i]
            if is_marked(n_next) {
                break
            }
            if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&n.next[i])), unsafe.Pointer(n_next), unsafe.Pointer(set_mark(n_next))) {
                break
            }
        }
    }
}

// -----------------------------------------------------------------------------

func New() *DataSet {
    init_rand_level()
    set := new(DataSet)
    max := new_node(share.KEY_MAX, 0, nil, uint32(share.LevelMax))
    min := new_node(share.KEY_MIN, 0, max, uint32(share.LevelMax))
    set.head = min
    return set
}

func (set *DataSet) Destroy() {
}

func (set *DataSet) Size() uint {
    var size uint = 0
    node := unset_mark(set.head.next[0])
    for node.next[0] != nil {
        if !is_marked(node.next[0]) {
            size++
        }
        node = unset_mark(node.next[0])
    }
    return size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    var succs [fraser_max_level]*node
    set.fraser_search(key, nil, succs[:])
    if succs[0].key == key && succs[0].deleted == 0 {
        return succs[0].val, true
    }
    return 0, false
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    var succs, preds [fraser_max_level]*node
    new_node := new_simple_node(key, val, uint32(get_rand_level()))

retry:
    set.fraser_search(key, preds[:], succs[:])

    /* Update the value field of an existing node */
    if succs[0].key == key { // Value already in list
        if succs[0].deleted != 0 { // Value is deleted: remove it and retry
            mark_node_ptrs(succs[0])
            goto retry
        }
        return false
    }

    for i := uint32(0); i < new_node.toplevel; i++ {
        new_node.next[i] = succs[i]
    }

    /* Node is visible once inserted at lowest level */
    if !atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&preds[0].next[0])), unsafe.Pointer(succs[0]), unsafe.Pointer(new_node)) {
        goto retry
    }

    for i := uint32(1); i < new_node.toplevel; i++ {
        for {
            pred := preds[i]
            succ := succs[i]
            /* Update the forward pointer if it is stale */
            new_next := new_node.next[i]
            if is_marked(new_next) {
                return true
            }
            if new_next != succ && !atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&new_node.next[i])), unsafe.Pointer(unset_mark(new_next)), unsafe.Pointer(succ)) {
                break; // Give up if pointer is marked
            }
            /* Check for old reference to a k node */
            if succ.key == key {
                succ = unset_mark(succ.next[0])
            }
            /* We retry the search if the CAS fails */
            if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&pred.next[i])), unsafe.Pointer(succ), unsafe.Pointer(new_node)) {
                break
            }

            set.fraser_search(key, preds[:], succs[:])
        }
    }
    return true
}

func (set *DataSet) Delete(key share.Key) (share.Val, bool) {
    var succs [fraser_max_level]*node

    set.fraser_search(key, nil, succs[:])

    if succs[0].key != key {
        return 0, false
    }

    /* 1. Node is logically deleted when the deleted field is not 0 */
    if succs[0].deleted != 0 {
        return 0, false
    }

    if atomic.AddUint32(&succs[0].deleted, 1) == 1 {
        /* 2. Mark forward pointers, then search will remove the node */
        mark_node_ptrs(succs[0])
        result := succs[0].val
        set.fraser_search(key, nil, nil)
        return result, true
    }
    return 0, false
}
