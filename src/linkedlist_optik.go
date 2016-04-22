/**
 * @file   linkedlist_optik.go
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
 * A Lazy Concurrent List-Based Set Algorithm,
 * S. Heller, M. Herlihy, V. Luchangco, M. Moir, W.N. Scherer III, N. Shavit
 * p.3-16, OPODIS 2005
**/

package dataset

import (
    "tools/share"
    "tools/optik"
)

// -----------------------------------------------------------------------------

type node struct {
    key share.Key
    val share.Val
    next *node
    mutex optik.Mutex
}

type DataSet struct {
    head *node
}

// -----------------------------------------------------------------------------

func new_node(key share.Key, val share.Val, next *node) *node {
    node := new(node)
    node.key = key
    node.val = val
    node.next = next
    node.mutex.Init()
    return node
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
    node := set.head.next
    for node.next != nil {
        size++
        node = node.next
    }
    return size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    curr := set.head
    for curr.key < key {
        curr = curr.next
    }
    if curr.key == key {
        return curr.val, true
    }
    return 0, false
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    var pred_ver optik.Mutex
    for {
        var pred *node
        curr := set.head
        for {
            curr_ver := curr.mutex.Load()
            pred = curr
            pred_ver = curr_ver
            curr = curr.next
            if !(curr.key < key) {
                break
            }
        }
        if curr.key == key {
            return false
        }
        newnode := new_node(key, val, curr)
        if !pred.mutex.TryLock_version(pred_ver) {
            continue
        }
        pred.next = newnode
        pred.mutex.Unlock()
        return true
    }
}

func (set *DataSet) Delete(key share.Key) (share.Val, bool) {
    for {
        var pred *node
        var pred_ver optik.Mutex
        curr := set.head
        curr_ver := curr.mutex
        for {
            pred = curr
            pred_ver = curr_ver
            curr = curr.next;
            curr_ver = curr.mutex.Load()
            if !(curr.key < key) {
                break
            }
        }
        if curr.key != key {
            return 0, false
        }
        cnxt := curr.next
        if !pred.mutex.TryLock_version(pred_ver) {
            continue
        }
        if !curr.mutex.TryLock_version(curr_ver) {
            pred.mutex.Revert()
            continue
        }
        pred.next = cnxt
        pred.mutex.Unlock()
        return curr.val, true
    }
}
