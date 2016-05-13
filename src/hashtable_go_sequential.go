/**
 * @file   hashtable_go_sequential.go
 * @author Sébastien Rouault <sebastien.rouault@epfl.ch>
 *
 * @section LICENSE
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
 * Hashtable with buckets of native map structure.
 * Sequential implementation.
**/

package dataset

import (
    "sync"
    "tools/share"
)

const (
    FindIsDef bool = true
)

// -----------------------------------------------------------------------------

type bucket struct {
    lock sync.Mutex
    set map[share.Key]share.Val
}

type DataSet struct {
    buckets []bucket
}

// -----------------------------------------------------------------------------

func (set *DataSet) getBucket(key share.Key) *bucket {
    return &set.buckets[uint(key) % share.NumBuckets]
}

// -----------------------------------------------------------------------------

func New() *DataSet {
    set := new(DataSet)
    set.buckets = make([]bucket, share.NumBuckets)
    for i := uint(0); i < share.NumBuckets; i++ {
        set.buckets[i].set = make(map[share.Key]share.Val)
    }
    return set
}

func (set *DataSet) Destroy() {
}

func (set *DataSet) Size() uint {
    var size uint = 0
    for i := uint(0); i < share.NumBuckets; i++ {
        bucket := &set.buckets[i]
        bucket.lock.Lock()
        size += uint(len(bucket.set))
        bucket.lock.Unlock()
    }
    return size
}

func (set *DataSet) Find(key share.Key) (res share.Val, ok bool) {
    bucket := set.getBucket(key)
    bucket.lock.Lock()
    defer bucket.lock.Unlock()
    res, ok = bucket.set[key]
    return
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    bucket := set.getBucket(key)
    bucket.lock.Lock()
    defer bucket.lock.Unlock()
    _, has := bucket.set[key]
    if has {
        return false
    }
    bucket.set[key] = val
    return true
}

func (set *DataSet) Delete(key share.Key) (res share.Val, ok bool) {
    bucket := set.getBucket(key)
    bucket.lock.Lock()
    defer bucket.lock.Unlock()
    res, ok = bucket.set[key]
    if !ok {
        return
    }
    delete(bucket.set, key)
    return
}
