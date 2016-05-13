/**
 * @file   skiplist_optik1.go
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
 * A skip-list algorithm design with OPTIK.
 * High-level description of the algorithm:
 * - Search: Simply traverse the levels of the skip list
 * - Parse (i.e., traverse to the point you want to modify): Traverse
 *     and keep track of the predecessor node for the target key at each level
 *     as well as the OPTIK version of each predecessor. Unlike other skip lists
 *     this one does not need to keep track of successor nodes for validation.
 *     optik_trylock_version takes care of validation.
 * - Insert: do the parse and the start from level 0, lock with trylock_version and
 *     insert the new node. If the trylock fails, reparse and continue from the previous
 *     level. The state flag of a node indicates whether a node is fully linked.
 * - Delete: parse and then try to do optik_trylock_vdelete on the node. If successful,
 *     try to grab the lock with optik_trylock_version on all levels and then unlink
 *     the node. If one of the trylock calls fail, release all locks and retry.
**/

package dataset

import (
    "tools/optik"
    "tools/share"
    "tools/xorshift"
)

const (
    FindIsDef bool = true
    optik_max_level = uint(64)
)

// -----------------------------------------------------------------------------

type node struct {
    key share.Key
    val share.Val
    toplevel uint32
    state uint32
    lock optik.Mutex
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
    elem.state = 0
    elem.lock.Init()
    elem.next = make([]*node, share.LevelMax)
    return elem
}

func new_node(key share.Key, val share.Val, next *node, toplevel uint32) *node {
    node := new_simple_node(key, val, toplevel)
    for i := uint32(0); i < toplevel; i++ {
        node.next[i] = next
    }
    return node
}

// -----------------------------------------------------------------------------

func (set *DataSet) optik_search(key share.Key, preds []*node, predsv []optik.Mutex, node_foundv *optik.Mutex) *node {
restart:
    var node_found *node = nil
    pred := set.head
    predv := set.head.lock
    for i := int(pred.toplevel - 1); i >= 0; i-- {
        curr := pred.next[i]
        currv := curr.lock
        for key > curr.key {
            predv = currv
            pred = curr
            curr = pred.next[i]
            currv = curr.lock
        }
        if optik.Is_deleted(predv) {
            goto restart
        }
        preds[i] = pred
        predsv[i] = predv
        if key == curr.key {
            node_found = curr
            *node_foundv = currv
        }
    }
    return node_found
}

func (set *DataSet) optik_left_search(key share.Key) *node {
    pred := set.head
    for i := int(pred.toplevel - 1); i >= 0; i-- {
        curr := pred.next[i]
        for key > curr.key {
            pred = curr
            curr = pred.next[i]
        }
        if key == curr.key {
            return curr
        }
    }
    return nil
}

func unlock_levels_down(nodes []*node, low int, high int) {
    var old *node = nil
    for i := high; i >= low; i-- {
        if old != nodes[i] {
            nodes[i].lock.Unlock()
        }
        old = nodes[i]
    }
}

func unlock_levels_up(nodes []*node, low int, high int) {
    var old *node = nil
    for i := low; i < high; i++ {
        if old != nodes[i] {
            nodes[i].lock.Unlock()
        }
        old = nodes[i]
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
    node := set.head.next[0] // We have at least 2 elements
    for node.next[0] != nil {
        if !optik.Is_deleted(node.lock) {
            size++
        }
        node = node.next[0]
    }
    return size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    nd := set.optik_left_search(key)
    if nd != nil && !optik.Is_deleted(nd.lock) {
        return nd.val, true
    }
    return 0, false
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    var preds  [optik_max_level]*node
    var predsv [optik_max_level]optik.Mutex
    var unused optik.Mutex
    var node_new *node = nil

    toplevel := int(get_rand_level())
    inserted_upto := int(0)

restart:
    node_found := set.optik_search(key, preds[:], predsv[:], &unused)
    if node_found != nil {
        if inserted_upto == 0 {
            if !optik.Is_deleted(node_found.lock) {
                return false
            } else { // There is a logically deleted node -- wait for it to be physically removed
                goto restart
            }
        }
    }
    if node_new == nil {
        node_new = new_simple_node(key, val, uint32(toplevel))
    }
    var pred_prev *node = nil
    for i := inserted_upto; i < toplevel; i++ {
        pred := preds[i]
        if pred_prev != pred && !pred.lock.TryLock_version(predsv[i]) {
            unlock_levels_down(preds[:], inserted_upto, i - 1)
            inserted_upto = i
            goto restart
        }
        node_new.next[i] = pred.next[i]
        pred.next[i] = node_new
        pred_prev = pred
    }
    node_new.state = 1
    unlock_levels_down(preds[:], inserted_upto, toplevel - 1)
    return true
}

func (set *DataSet) Delete(key share.Key) (share.Val, bool) {
    var preds  [optik_max_level]*node
    var predsv [optik_max_level]optik.Mutex
    var node_foundv optik.Mutex

    my_delete := false

restart:
    node_found := set.optik_search(key, preds[:], predsv[:], &node_foundv)
    if node_found == nil {
        return 0, false
    }

    if !my_delete {
        if optik.Is_deleted(node_found.lock) || node_found.state == 0 {
            return 0, false
        }
        if !node_found.lock.TryLock_vdelete(node_foundv) {
            if (optik.Is_deleted(node_found.lock)) {
                return 0, false
            } else {
                goto restart
            }
        }
    }

    my_delete = true

    toplevel_nf := node_found.toplevel
    var pred_prev *node = nil
    for i := int(0); i < int(toplevel_nf); i++ {
        pred := preds[i]
        if pred_prev != pred && !pred.lock.TryLock_version(predsv[i]) {
            unlock_levels_down(preds[:], 0, i - 1)
            goto restart
        }
        pred_prev = pred
    }

    for i := uint32(0); i < toplevel_nf; i++ {
        preds[i].next[i] = node_found.next[i]
    }
    unlock_levels_down(preds[:], 0, int(toplevel_nf - 1))
    return node_found.val, true
}
