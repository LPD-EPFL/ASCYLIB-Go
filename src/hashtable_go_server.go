/**
 * @file   hashtable_go_server.go
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
 * Hashtable with native map structure.
 * One "server" goroutine, using channels.
 *
 * This implementation should be used like:
 *   ch := set.<Op>Async()
 *   <do something else>
 *   res := <-ch // Only when first needed
 * Here it is not the case, so useful only to estimate overhead.
**/

package dataset

import (
    "tools/share"
)

const (
    FindIsDef bool = true
    query_buffer_size uint = 16
)

// -----------------------------------------------------------------------------

type bucket struct {
    queries chan interface{}
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

// Call types
type SizeAsyncCall struct {
    res chan<- SizeAsyncRes
}
type FindAsyncCall struct {
    key share.Key
    res chan<- FindAsyncRes
}
type InsertAsyncCall struct {
    key share.Key
    val share.Val
    res chan<- InsertAsyncRes
}
type DeleteAsyncCall struct {
    key share.Key
    res chan<- DeleteAsyncRes
}

// -----------------------------------------------------------------------------

func (set *DataSet) getBucket(key share.Key) *bucket {
    return &set.buckets[uint(key) % share.NumBuckets]
}

// -----------------------------------------------------------------------------

func (set *DataSet) SizeAsync() <-chan SizeAsyncRes {
    res := make(chan SizeAsyncRes, 1)
    go func() {
        queries := make([](chan SizeAsyncRes), share.NumBuckets)
        var sum uint = 0
        for i := uint(0); i < share.NumBuckets; i++ { // Queries
            queries[i] = make(chan SizeAsyncRes, 1)
            set.buckets[i].queries <- &SizeAsyncCall{queries[i]}
        }
        for i := uint(0); i < share.NumBuckets; i++ { // Collect
            sum += (<-queries[i]).size
        }
        res <- SizeAsyncRes{sum}
    }()
    return res
}

func (set *DataSet) FindAsync(key share.Key) <-chan FindAsyncRes {
    res := make(chan FindAsyncRes, 1)
    set.getBucket(key).queries <- &FindAsyncCall{key, res}
    return res
}

func (set *DataSet) InsertAsync(key share.Key, val share.Val) <-chan InsertAsyncRes {
    res := make(chan InsertAsyncRes, 1)
    set.getBucket(key).queries <- &InsertAsyncCall{key, val, res}
    return res
}

func (set *DataSet) DeleteAsync(key share.Key) <-chan DeleteAsyncRes {
    res := make(chan DeleteAsyncRes, 1)
    set.getBucket(key).queries <- &DeleteAsyncCall{key, res}
    return res
}

// -----------------------------------------------------------------------------

func New() *DataSet {
    set := new(DataSet)
    set.buckets = make([]bucket, share.NumBuckets)
    for i := uint(0); i < share.NumBuckets; i++ {
        set.buckets[i].set = make(map[share.Key]share.Val)
        set.buckets[i].queries = make(chan interface{}, query_buffer_size)
        go func(bucket *bucket) { // Server goroutine
            for {
                anyquery := <-bucket.queries
                if anyquery == nil {
                    return
                }
                switch query := anyquery.(type) {
                case *SizeAsyncCall:
                    query.res <- SizeAsyncRes{uint(len(bucket.set))}
                case *FindAsyncCall:
                    val, ok := bucket.set[query.key]
                    query.res <- FindAsyncRes{val, ok}
                case *InsertAsyncCall:
                    _, has := bucket.set[query.key]
                    if has {
                        query.res <- InsertAsyncRes{false}
                        continue
                    }
                    bucket.set[query.key] = query.val
                    query.res <- InsertAsyncRes{true}
                case *DeleteAsyncCall:
                    val, ok := bucket.set[query.key]
                    if !ok {
                        query.res <- DeleteAsyncRes{0, false}
                        continue
                    }
                    delete(bucket.set, query.key)
                    query.res <- DeleteAsyncRes{val, true}
                default:
                    panic("Unknow query")
                }
            }
        }(&set.buckets[i])
    }
    return set
}

func (set *DataSet) Destroy() {
    for i := uint(0); i < share.NumBuckets; i++ {
        close(set.buckets[i].queries)
    }
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
