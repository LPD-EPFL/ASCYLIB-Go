/**
 * @file   priorityqueue_lotanshavit_lf.go
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
 * I. Lotan and N. Shavit. Skiplist-based concurrent priority queues.
 * In Parallel and Distributed Processing Symposium, 2000. IPDPS 2000. Proceedings.
 * 14th International, pages 263268. IEEE, 2000.
**/

package dataset

import (
    "sync/atomic"
    "tools/share"
    "tools/xorshift"
    "unsafe"
)

const (
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

func (set *DataSet) fraser_search(key share.Key, left_list []*node, right_list []*node) bool {
retry:
    left := set.head
    var right *node
    for i := int(share.LevelMax - 1); i >= 0; i-- {
        left_next := left.next[i]
        if is_marked(left_next) {
            goto retry
        }

        /* Find unmarked node pair at this level */
        var right_next *node
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
    return right.key == key
}

func (set *DataSet) fraser_search_no_cleanup(key share.Key, left_list []*node, right_list []*node) bool {
    left := set.head
    var right *node
    for i := int(share.LevelMax - 1); i >= 0; i-- {
        left_next := unset_mark(left.next[i])
        right = left_next
        for {
            if !is_marked(right.next[i]) {
                if right.key >= key {
                    break
                }
                left = right
            }
            right = unset_mark(right.next[i])
        }
        left_list[i] = left
        right_list[i] = right
    }
    return right.key == key
}

func (set *DataSet) fraser_search_no_cleanup_succs(key share.Key, right_list []*node) bool {
    left := set.head
    var right *node
    for i := int(share.LevelMax - 1); i >= 0; i-- {
        left_next := unset_mark(left.next[i])
        right = left_next
        for {
            if !is_marked(right.next[i]) {
                if right.key >= key {
                    break
                }
                left = right
            }
            right = unset_mark(right.next[i])
        }
        right_list[i] = right
    }
    return right.key == key
}

func (set *DataSet) fraser_left_search(key share.Key) *node {
    left_prev := set.head
    var left *node
    for lvl := int(share.LevelMax - 1); lvl >= 0; lvl-- {
        left = unset_mark(left_prev.next[lvl])
        for left.key < key || is_marked(left.next[lvl]) {
            if !is_marked(left.next[lvl]) {
                left_prev = left
            }
            left = unset_mark(left.next[lvl])
        }
        if left.key == key {
            break
        }
    }
    return left
}

func mark_node_ptrs(n *node) bool {
    var cas bool = false
    for i := int(n.toplevel - 1); i >= 0; i-- {
        for {
            n_next := n.next[i]
            if is_marked(n_next) {
                cas = false
                break
            }
            cas = atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&n.next[i])), unsafe.Pointer(unset_mark(n_next)), unsafe.Pointer(set_mark(n_next)))
            if cas {
                break
            }
        }
    }
    return cas
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
    size := uint(0)
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
    left := set.fraser_left_search(key)
    if (left.key == key) {
        return left.val, true
    }
    return 0, false
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    var succs, preds [fraser_max_level]*node
retry:
    found := set.fraser_search_no_cleanup(key, preds[:], succs[:])
    if found {
        return false
    }
    elem := new_simple_node(key, val, uint32(get_rand_level()))
    for i := uint32(0); i < elem.toplevel; i++ {
        elem.next[i] = succs[i]
    }
    // Node is visible once inserted at lowest level
    if !atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&preds[0].next[0])), unsafe.Pointer(unset_mark(succs[0])), unsafe.Pointer(elem)) {
        goto retry
    }
    for i := uint32(1); i < elem.toplevel; i++ {
        for {
            pred := preds[i]
            succ := succs[i]
            if is_marked(elem.next[i]) {
                return true
            }
            if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&pred.next[i])), unsafe.Pointer(succ), unsafe.Pointer(elem)) {
                break
            }
            set.fraser_search(key, preds[:], succs[:])
        }
    }
    return true
}

func (set *DataSet) Delete(key share.Key) (share.Val, bool) {
    elem := unset_mark(set.head.next[0])
    for elem.next[0] != nil {
        if !is_marked(elem.next[elem.toplevel - 1]) {
            if mark_node_ptrs(elem) {
                result := elem.val
                set.fraser_search(elem.key, nil, nil)
                return result, true
            }
        }
        elem = unset_mark(elem.next[0])
    }
    return 0, false
}
