/**
 * @file   skiplist_seq.go
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
 * A sequential skiplist (= doesn't support concurrency).
**/

package dataset

import (
    "tools/assert"
    "tools/share"
    "tools/xorshift"
    "unsafe"
)

const (
    FindIsDef bool = true
    maxlevel = uint(32)
)

// -----------------------------------------------------------------------------

type node struct {
    key share.Key
    val share.Val
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

func new_simple_node(key share.Key, val share.Val, toplevel uint32) *node {
    elem := new(node)
    elem.key = key
    elem.val = val
    elem.toplevel = toplevel
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

func is_marked(i *node) bool {
    return (uintptr(unsafe.Pointer(i)) & 1) != 0
}

func unset_mark(i *node) *node {
    return (*node)(unsafe.Pointer(uintptr(unsafe.Pointer(i)) &^ 0x01))
}

func set_mark(i *node) *node {
    return (*node)(unsafe.Pointer(uintptr(unsafe.Pointer(i)) | 0x01))
}

// -----------------------------------------------------------------------------

func New() *DataSet {
    assert.Assert(share.LevelMax <= maxlevel, "'LevelMax' is above maximum level")
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
        if (!is_marked(node.next[0])) {
            size++
        }
        node = unset_mark(node.next[0])
    }
    return size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    var node, next *node = set.head, nil
    for i := int(node.toplevel - 1); i >= 0; i-- {
        next = node.next[i]
        for next.key < key {
            node = next
            next = node.next[i]
        }
    }
    node = node.next[0]
    if node.key == key {
        return node.val, true
    }
    return 0, false
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    var preds, succs [maxlevel]*node
    var node, next *node = set.head, nil
    for i := int(node.toplevel - 1); i >= 0; i-- {
        next = node.next[i]
        for next.key < key {
            node = next
            next = node.next[i]
        }
        preds[i] = node
        succs[i] = node.next[i]
    }
    node = node.next[0]
    if node.key != key {
        l := get_rand_level()
        node = new_simple_node(key, val, uint32(l))
        for i := uint(0); i < l; i++ {
            node.next[i] = succs[i]
            preds[i].next[i] = node
        }
        return true
    }
    return false
}

func (set *DataSet) Delete(key share.Key) (share.Val, bool) {
    var preds, succs [maxlevel]*node
    var node, next *node = set.head, nil
    for i := int(node.toplevel - 1); i >= 0; i-- {
        next = node.next[i]
        for next.key < key {
            node = next
            next = node.next[i]
        }
        preds[i] = node
        succs[i] = node.next[i]
    }
    if next.key == key {
        result := next.val
        for i := uint32(0); i < set.head.toplevel; i++ {
            if succs[i].key == key {
                preds[i].next[i] = succs[i].next[i]
            }
        }
        return result, true
    }
    return 0, false
}
