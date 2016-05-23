/**
 * @file   hashtable_optik1.go
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
 * Using one lazy list per bucket
**/

package dataset

import (
    "runtime"
    "tools/optik"
    "tools/share"
)

const (
    FindIsDef bool = true
    read_only_fail bool = true
    maxhtlength uint = 65536 // # of buckets
)

// -----------------------------------------------------------------------------

type node struct {
    key share.Key
    val share.Val
    next *node
}

type bucket struct {
    head *node
    lock optik.Mutex
}

type DataSet struct {
    hash uint
    buckets []bucket
}

// -----------------------------------------------------------------------------

func (bckt *bucket) init() {
    bckt.head = nil
    bckt.lock.Init()
}

func new_node(key share.Key, val share.Val, next *node) *node {
    nd := new(node)
    nd.key = key
    nd.val = val
    nd.next = next
    return nd
}

// -----------------------------------------------------------------------------

func New() *DataSet {
    set := new(DataSet)
    set.hash = maxhtlength - 1
    set.buckets = make([]bucket, maxhtlength)
    for i := uint(0); i < maxhtlength; i++ {
        set.buckets[i].init()
    }
    return set
}

func (set *DataSet) Destroy() {
}

func (set *DataSet) Size() uint {
    var size uint = 0
    for i := uint(0); i < maxhtlength; i++ {
        node := set.buckets[i].head
        for node != nil {
            size++
            node = node.next
        }
    }
    return size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    curr := set.buckets[uint(key) & set.hash].head
    for curr != nil && curr.key < key {
        curr = curr.next
    }
    if curr != nil && curr.key == key {
        return curr.val, true
    }
    return 0, false
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    bucket := &set.buckets[uint(key) & set.hash]

    var curr, pred *node
    for {
        pred_ver := bucket.lock.Load()
        curr = bucket.head
        for curr != nil && curr.key < key {
            pred = curr
            curr = curr.next
        }
        if curr != nil && curr.key == key {
            return false
        }
        if bucket.lock.TryLock_version(pred_ver) {
            break
        }
        runtime.Gosched() // In order not to fight against the GC
    }

    newnode := new_node(key, val, curr)
    if pred != nil {
        pred.next = newnode
    } else {
        bucket.head = newnode
    }
    bucket.lock.Unlock()

    return true
}

func (set *DataSet) Delete(key share.Key) (share.Val, bool) {
    bucket := &set.buckets[uint(key) & set.hash]

    var curr, pred *node
    for {
        pred_ver := bucket.lock.Load()
        curr = bucket.head
        for curr != nil && curr.key < key {
            pred = curr
            curr = curr.next
        }
        if curr == nil || curr.key != key {
            return 0, false
        }
        if bucket.lock.TryLock_version(pred_ver) {
            break
        }
        runtime.Gosched() // In order not to fight against the GC
    }

    result := curr.val
    if pred != nil {
        pred.next = curr.next
    } else {
        bucket.head = curr.next
    }
    bucket.lock.Unlock()

    return result, true
}
