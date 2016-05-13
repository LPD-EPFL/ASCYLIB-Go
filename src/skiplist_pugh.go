/**
 * @file   skiplist_pugh.go
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
 * William Pugh.
 * Concurrent Maintenance of Skip Lists. Technical report, 1990.
**/

package dataset

import (
    "sync"
    "tools/assert"
    "tools/share"
    "tools/xorshift"
)

const (
    FindIsDef bool = true
    maxlevel = uint(32)
    herlihy_maxlevel = uint(64) // Covers up to 2^64 elements
)

// -----------------------------------------------------------------------------

type node struct {
    key share.Key
    val share.Val
    toplevel uint32
    lock sync.Mutex
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

func get_lock(pred *node, key share.Key, lvl uint32) *node {
    succ := pred.next[lvl]
    for succ.key < key {
        pred = succ
        succ = succ.next[lvl]
    }

    pred.lock.Lock()
    succ = pred.next[lvl]
    for succ.key < key {
        pred.lock.Unlock()
        pred = succ
        pred.lock.Lock()
        succ = pred.next[lvl]
    }

    return pred
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
    node := set.head.next[0] // We have at least 2 elements
    for (node.next[0] != nil) {
        size++
        node = node.next[0]
    }
    return size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    pred := set.head
    for lvl := int(share.LevelMax - 1); lvl >= 0; lvl-- {
        succ := pred.next[lvl]
        for (succ.key < key) {
            pred = succ
            succ = succ.next[lvl]
        }
        if (succ.key == key) {
            return succ.val, true
        }
    }
    return 0, false
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    var update [herlihy_maxlevel]*node
    pred := set.head
    for lvl := int(share.LevelMax - 1); lvl >= 0; lvl-- {
        succ := pred.next[lvl]
        for succ.key < key {
            pred = succ
            succ = succ.next[lvl]
        }
        if succ.key == key {
            return false
        }
        update[lvl] = pred
    }

    rand_lvl := get_rand_level()

    pred = get_lock(pred, key, 0)
    if pred.next[0].key == key {
        pred.lock.Unlock()
        return false
    }

    n := new_simple_node(key, val, uint32(rand_lvl))
    n.lock.Lock()
    n.next[0] = pred.next[0] // We already hold the lock for lvl 0
    /// TODO: Ensure no reordoring here
    pred.next[0] = n
    pred.lock.Unlock()
    for lvl := uint32(1); lvl < n.toplevel; lvl++ {
        pred = get_lock(update[lvl], key, lvl)
        n.next[lvl] = pred.next[lvl]
        /// TODO: Ensure no reordoring here
        pred.next[lvl] = n
        pred.lock.Unlock()
    }
    n.lock.Unlock()
    return true
}

func (set *DataSet) Delete(key share.Key) (share.Val, bool) {
    var update [herlihy_maxlevel]*node
    var succ *node
    pred := set.head
    for lvl := int(share.LevelMax - 1); lvl >= 0; lvl-- {
        succ = pred.next[lvl]
        for succ.key < key {
            pred = succ
            succ = succ.next[lvl]
        }
        update[lvl] = pred
    }

    succ = pred
    for {
        succ = succ.next[0]
        if succ.key > key {
            return 0, false
        }
        succ.lock.Lock()
        if succ.key <= succ.next[0].key && succ.key == key {
            break
        }
        succ.lock.Unlock()
    }

    for lvl := int(succ.toplevel - 1); lvl >= 0; lvl-- {
        pred = get_lock(update[lvl], key, uint32(lvl))
        pred.next[lvl] = succ.next[lvl]
        succ.next[lvl] = pred
        pred.lock.Unlock()
    }
    succ.lock.Unlock()
    return succ.val, true
}
