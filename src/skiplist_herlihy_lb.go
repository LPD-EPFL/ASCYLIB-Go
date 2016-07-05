/**
 * @file   skiplist_herlihy_lb.go
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
 * Fine-grained locking skip list.
 * C implementation of the Herlihy et al. algorithm originally designed for managed programming language.
 * "A Simple Optimistic Skiplist Algorithm", M. Herlihy, Y. Lev, V. Luchangco, N. Shavit, p.124-138, SIROCCO 2007.
**/

package dataset

import (
    "runtime"
    "tools/share"
    "tools/ttas"
    "tools/volatile"
    "tools/xorshift"
)

const (
    FindIsDef bool = true
    herlihy_max_level = uint(64)
)

// -----------------------------------------------------------------------------

type node struct {
    key share.Key
    val share.Val
    toplevel uint32
    marked bool
    fullylinked bool
    lock ttas.Mutex
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
    elem.marked = false
    elem.fullylinked = false
    elem.next = make([]*node, toplevel)
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

func ok_to_delete(elem *node, found int) bool {
    return elem.fullylinked && (int(elem.toplevel - 1) == found) && !elem.marked
}

func (set *DataSet) optimistic_search(key share.Key, preds []*node, succs []*node) int {
restart:
    found := -1
    pred := set.head
    for i := int(pred.toplevel - 1); i >= 0; i-- {
        curr := pred.next[i]
        for key > curr.key {
            pred = curr
            curr = pred.next[i]
        }
        if preds != nil {
            preds[i] = pred
            if pred.marked {
                runtime.Gosched() // In order not to fight with the GC
                goto restart
            }
        }
        succs[i] = curr
        if found == -1 && key == curr.key {
            found = i
        }
    }
    return found
}

func (set *DataSet) optimistic_left_search(key share.Key) *node {
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

func (set *DataSet) unlock_levels(nodes []*node, highestlevel uint) {
    var old *node = nil
    for i := uint(0); i <= highestlevel; i++ {
        if (old != nodes[i]) {
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
    max.fullylinked = true
    min.fullylinked = true
    set.head = min
    return set
}

func (set *DataSet) Destroy() {
}

func (set *DataSet) Size() uint {
    var size uint = 0
    node := set.head.next[0] // We have at least 2 elements
    for node.next[0] != nil {
        if (node.fullylinked && !node.marked) {
            size++
        }
        node = node.next[0]
    }
    return size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    nd := set.optimistic_left_search(key)
    if nd != nil && !nd.marked && nd.fullylinked {
        return nd.val, true
    }
    return 0, false
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    var succs, preds [herlihy_max_level]*node

    toplevel := get_rand_level()
    backoff := uint(1)

    for {
        found := set.optimistic_search(key, preds[:], succs[:])
        if (found != -1) {
            node_found := succs[found]
            if (!node_found.marked) {
                for (!volatile.ReadBool(&node_found.fullylinked)) {
                    runtime.Gosched()
                }
                return false
            }
            continue
        }

        highest_locked := -1
        var prev_pred *node = nil
        valid := true
        for i := uint(0); valid && (i < toplevel); i++ {
            pred := preds[i]
            succ := succs[i]
            if pred != prev_pred {
                pred.lock.Lock()
                highest_locked = int(i)
                prev_pred = pred
            }
            valid = !pred.marked && !succ.marked && pred.next[i] == succ
        }

        if (!valid) { // Unlock the predecessors before leaving
            set.unlock_levels(preds[:], uint(highest_locked)) // Unlocks the global-lock in the GL case
            if (backoff > 5000) {
                runtime.Gosched() // Rough approximation of: nop_rep(backoff & MAX_BACKOFF)
            }
            backoff <<= 1
            continue
        }

        new_node := new_simple_node(key, val, uint32(toplevel))

        for i := uint(0); i < toplevel; i++ {
            new_node.next[i] = succs[i]
        }

        for i := uint(0); i < toplevel; i++ {
            preds[i].next[i] = new_node
        }

        new_node.fullylinked = true

        set.unlock_levels(preds[:], uint(highest_locked))
        return true
    }
}

func (set *DataSet) Delete(key share.Key) (share.Val, bool) {
    var succs, preds [herlihy_max_level]*node
    var node_todel *node

    node_todel = nil
    is_marked := false
    toplevel := -1
    backoff := uint(1)

    for {
        found := set.optimistic_search(key, preds[:], succs[:])

        /* If not marked and ok to delete, then mark it */
        if !(is_marked || (found != -1 && ok_to_delete(succs[found], found))) {
            return 0, false
        }

        if (!is_marked) {
            node_todel = succs[found]

            node_todel.lock.Lock()
            toplevel = int(node_todel.toplevel)

            /* Unless it has been marked meanfor */
            if (node_todel.marked) {
                node_todel.lock.Unlock()
                return 0, false
            }

            node_todel.marked = true
            is_marked = true
        }

        /* Physical deletion */
        highest_locked := -1
        var prev_pred *node = nil
        valid := true
        for i := int(0); valid && (i < toplevel); i++ {
            pred := preds[i]
            succ := succs[i]
            if pred != prev_pred {
                pred.lock.Lock()
                highest_locked = int(i)
                prev_pred = pred
            }
            valid = !pred.marked && pred.next[i] == succ
        }

        if !valid {
            set.unlock_levels(preds[:], uint(highest_locked))
            if (backoff > 5000) {
                runtime.Gosched() // Rough approximation of: nop_rep(backoff & MAX_BACKOFF)
            }
            backoff <<= 1
            continue
        }

        for i := int(toplevel - 1); i >= 0; i-- {
            preds[i].next[i] = node_todel.next[i]
        }

        val := node_todel.val

        node_todel.lock.Unlock()
        set.unlock_levels(preds[:], uint(highest_locked))

        return val, true
    }
}
