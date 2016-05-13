/**
 * @file   hashtable_java.go
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
 * Similar to Java's ConcurrentHashMap.
 * Doug Lea. 1.3.4. http://gee.cs.oswego.edu/dl/classes/EDU/oswego/
 * cs/dl/util/concurrent/intro.html, 2003.
**/

package dataset

import (
    "runtime"
    "sync/atomic"
    "tools/share"
    "tools/volatile"
    "unsafe"
)

const (
    FindIsDef bool = true
    base_load_factor float32 = 1
    read_only_fail bool = true
)

// -----------------------------------------------------------------------------

// Like sync.Mutex because it lacks TryLock...
type mutex struct {
    state uint32
}

func (m *mutex) tryLock() bool {
    return atomic.CompareAndSwapUint32(&m.state, 0, 1)
}

func (m *mutex) lock() {
    for {
        for { // Wait unlocked state
            if volatile.ReadUint32(&m.state) == 0 {
                break
            }
        }
        if atomic.CompareAndSwapUint32(&m.state, 0, 1) {
            break
        }
        runtime.Gosched()
    }
}

func (m *mutex) unlock() {
    atomic.StoreUint32(&m.state, 0)
}

// -----------------------------------------------------------------------------

type node struct {
    key share.Key
    val share.Val
    next *node
}

type segment struct {
    num_buckets uint
    hash uint
    lock mutex
    modifications uint32
    size uint32
    load_factor float32
    size_limit uint32
    table []*node
}

type DataSet struct {
    num_segments uint
    hash uint
    hash_seed uint
    segments []*segment
}

// -----------------------------------------------------------------------------

func uintLog2(nb uint) uint {
    var r uint = 0
    for nb >= 2 {
        r++
        nb = nb >> 1
    }
    return r
}

func hash(key share.Key, hash_seed uint) uint {
    return uint(key) >> hash_seed
}

// -----------------------------------------------------------------------------

func new_node(key share.Key, val share.Val, next *node) *node {
    node := new(node)
    node.key = key
    node.val = val
    node.next = next
    return node
}

func new_segment(capacity uint, load_factor float32) *segment {
    seg := new(segment)
    seg.table = make([]*node, capacity)
    seg.num_buckets = capacity
    seg.hash = capacity - 1
    seg.modifications = 0
    seg.size = 0
    seg.load_factor = load_factor
    seg.size_limit = uint32(seg.load_factor * float32(capacity))
    if seg.size_limit == 0 {
        seg.size_limit = 1
    }
    return seg
}

func (set *DataSet) segment_rehash(seg_num uint, newn *node) {
    seg_old := set.segments[seg_num]
    seg_new := new_segment(seg_old.num_buckets << 1, seg_old.load_factor)
    mask_new := seg_new.hash
    for b := uint(0); b < seg_old.num_buckets; b++ {
        curr := seg_old.table[b]
        if curr != nil {
            next := curr.next
            idx := hash(curr.key, set.hash_seed) & mask_new
            if next == nil { /* single node on list */
                seg_new.table[idx] = curr
            } else { /* reuse consecutive sequence at same slot */
                last_run := curr
                last_idx := idx
                for last := next; last != nil; last = last.next {
                    k := hash(last.key, set.hash_seed) & mask_new
                    if k != last_idx {
                        last_idx = k
                        last_run = last
                    }
                }
                seg_new.table[last_idx] = last_run
                /* clone remaining */
                for p := curr; p != last_run; p = p.next {
                    k := hash(p.key, set.hash_seed) & mask_new
                    seg_new.table[k] = new_node(p.key, p.val, seg_new.table[k])
                }
            }
        }
    }
    new_idx := hash(newn.key, set.hash_seed) & mask_new; /* add the new node */
    newn.next = seg_new.table[new_idx]
    seg_new.table[new_idx] = newn
    seg_new.size = seg_old.size + 1
    set.segments[seg_num] = seg_new
}

func (set *DataSet) contains(seg *segment, key share.Key) bool {
    curr := seg.table[hash(key, set.hash_seed) & seg.hash]
    for curr != nil {
        if curr.key == key {
            return true
        }
        curr = curr.next
    }
    return false
}

// -----------------------------------------------------------------------------

func New() *DataSet {
    set := new(DataSet)
    if share.Capacity < share.Concurrency {
        share.Capacity = share.Concurrency
    }
    set.num_segments = share.Concurrency
    set.segments = make([]*segment, set.num_segments)
    set.hash = set.num_segments - 1
    set.hash_seed = uintLog2(set.num_segments)
    capacity_seg := share.Capacity / set.num_segments
    for s := uint(0); s < set.num_segments; s++ {
        set.segments[s] = new_segment(capacity_seg, base_load_factor)
    }
    return set
}

func (set *DataSet) Destroy() {
}

func (set *DataSet) Size() uint {
    var size uint = 0
    for s := uint(0); s < set.num_segments; s++ {
        seg := set.segments[s]
        for i := uint(0); i < seg.num_buckets; i++ {
            curr := seg.table[i]
            for curr != nil {
                size++
                curr = curr.next
            }
        }
    }
    return size
}

func (set *DataSet) Find(key share.Key) (res share.Val, ok bool) {
    seg := set.segments[uint(key) & set.hash]
    curr := seg.table[hash(key, set.hash_seed) & seg.hash]
    for curr != nil {
        if curr.key == key {
            return curr.val, true
        }
        curr = curr.next
    }
    return 0, false
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    var seg *segment
    var seg_lock *mutex
    seg_num := uint(key) & set.hash

    if read_only_fail {
        seg = set.segments[seg_num]
        if set.contains(seg, key) {
            return false
        }
    }

    for {
        seg = (*segment)(volatile.ReadPointer((*unsafe.Pointer)(unsafe.Pointer(&set.segments[seg_num]))))
        seg_lock = &seg.lock
        if seg_lock.tryLock() {
            break
        }
        runtime.Gosched()
    }

    bucket := &seg.table[hash(key, set.hash_seed) & seg.hash]
    curr := *bucket
    var pred *node
    for curr != nil {
        if curr.key == key {
            seg_lock.unlock()
            return false
        }
        pred = curr
        curr = curr.next
    }
    n := new_node(key, val, nil)
    sizepp := seg.size + 1
    if sizepp >= seg.size_limit {
        set.segment_rehash(seg_num, n)
    } else {
        if pred != nil {
            pred.next = n
        } else {
            *bucket = n
        }
        seg.size = sizepp
        seg_lock.unlock()
    }
    return true
}

func (set *DataSet) Delete(key share.Key) (result share.Val, ok bool) {
    var seg *segment
    var seg_lock *mutex
    seg_num := uint(key) & set.hash

    if read_only_fail {
        seg = set.segments[seg_num]
        if !set.contains(seg, key) {
            return 0, false
        }
    }

    for {
        seg = (*segment)(volatile.ReadPointer((*unsafe.Pointer)(unsafe.Pointer(&set.segments[seg_num]))))
        seg_lock = &seg.lock
        if seg_lock.tryLock() {
            break
        }
        runtime.Gosched()
    }

    bucket := &seg.table[hash(key, set.hash_seed) & seg.hash]
    curr := *bucket
    var pred *node
    for curr != nil {
        if curr.key == key {
            /* do the remove */
            if pred != nil {
                pred.next = curr.next
            } else {
                *bucket = curr.next
            }
            seg.size--
            seg_lock.unlock()
            return curr.val, true
        }
        pred = curr
        curr = curr.next
    }
    seg_lock.unlock()
    return 0, false
}
