/**
 * @file   hashtable_go_postpone.go
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
 * Sequential implementation "postponed" in goroutines.
 *
 * This implementation should be used like:
 *   ch := set.<Op>Async()
 *   <do something else>
 *   res := <-ch // Only when first needed
 * Here it is not the case, so useful only to estimate overhead.
**/

package dataset

import (
    "sync"
    "tools/share"
)

// -----------------------------------------------------------------------------

type bucket struct {
    lock sync.Mutex
    set map[share.Key]share.Val
}

type DataSet struct {
    buckets []bucket
}

// Result types
type SizeAsyncRes struct {
    size uint
}
type FindAsyncRes struct {
    res share.Val
    ok bool
}
type InsertAsyncRes struct {
    ok bool
}
type DeleteAsyncRes struct {
    res share.Val
    ok bool
}

// -----------------------------------------------------------------------------

func (set *DataSet) getBucket(key share.Key) *bucket {
    return &set.buckets[uint(key) % share.NumBuckets]
}

// -----------------------------------------------------------------------------

func (set *DataSet) SizeAsync() <-chan SizeAsyncRes {
    res := make(chan SizeAsyncRes, 1)
    go func() {
        var size uint = 0
        for i := uint(0); i < share.NumBuckets; i++ {
            bucket := &set.buckets[i]
            bucket.lock.Lock()
            size += uint(len(bucket.set))
            bucket.lock.Unlock()
        }
        res <- SizeAsyncRes{size}
    }()
    return res
}

func (set *DataSet) FindAsync(key share.Key) <-chan FindAsyncRes {
    res := make(chan FindAsyncRes, 1)
    go func() {
        bucket := set.getBucket(key)
        bucket.lock.Lock()
        defer bucket.lock.Unlock()
        val, ok := bucket.set[key]
        res <- FindAsyncRes{val, ok}
    }()
    return res
}

func (set *DataSet) InsertAsync(key share.Key, val share.Val) <-chan InsertAsyncRes {
    res := make(chan InsertAsyncRes, 1)
    go func() {
        bucket := set.getBucket(key)
        bucket.lock.Lock()
        defer bucket.lock.Unlock()
        _, has := bucket.set[key]
        if has {
            res <- InsertAsyncRes{false}
            return
        }
        bucket.set[key] = val
        res <- InsertAsyncRes{true}
    }()
    return res
}

func (set *DataSet) DeleteAsync(key share.Key) <-chan DeleteAsyncRes {
    res := make(chan DeleteAsyncRes, 1)
    go func() {
        bucket := set.getBucket(key)
        bucket.lock.Lock()
        defer bucket.lock.Unlock()
        val, ok := bucket.set[key]
        if !ok {
            res <- DeleteAsyncRes{0, false}
            return
        }
        delete(bucket.set, key)
        res <- DeleteAsyncRes{val, true}
    }()
    return res
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
    res := <-set.SizeAsync()
    return res.size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    res := <-set.FindAsync(key)
    return res.res, res.ok
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    res := <-set.InsertAsync(key, val)
    return res.ok
}

func (set *DataSet) Delete(key share.Key) (share.Val, bool) {
    res := <-set.DeleteAsync(key)
    return res.res, res.ok
}
