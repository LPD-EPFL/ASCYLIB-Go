/**
 * @file   channels.go
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
 * Channel communication micro-benchmark.
**/

package main

import (
    "flag"
    "fmt"
    "runtime"
    "strconv"
    "sync"
    "sync/atomic"
    "time"
    "unsafe"
    "../src/tools/volatile"
    "../src/tools/xorshift"
)

// Methods to use
const (
    random      uint = iota // Clients speak to random servers
    round_robin uint = iota // Clients speak to one server, one after the other (likely ~ to random ??)
    shared      uint = iota // Clients speak to affected server (equi-distribution)
)

// -----------------------------------------------------------------------------

// Message type
type payload uint
type message struct {
    data payload
    ret  chan(message)
}

// Shared
var method    uint // Communication method used
var nbservers uint // Amount of servers
var servers   []chan(message) // Servers' channels
var running   uint32 = 1
var startwg   sync.WaitGroup
var barrier   sync.WaitGroup

// Statistics
var gbflow uint64 = 0 // Amount of sent/received messages
var gbltnc uint64 = 0 // Latency sum (in ns)

/** Client goroutine.
 * @param id Client ID
**/
func client(id uint) {
    var nbflow uint64 = 0 // Amount of sent/received messages
    var latency uint64 = 0 // Latency sum (in ns)
    ret := make(chan(message)) // Return channel

    switch (method) {
    case random:
        var xorshift xorshift.State
        xorshift.Init()
        startwg.Wait()
        for volatile.ReadUint32(&running) > 0 {
            servers[xorshift.Intn(uint32(nbservers))] <- message{0, ret}
            <-ret
            nbflow++
        }
    case round_robin:
        var target uint = 0
        startwg.Wait()
        for volatile.ReadUint32(&running) > 0 {
            servers[target] <- message{0, ret}
            <-ret
            nbflow++
            target = (target + 1) % nbservers
        }
    case shared:
        var target uint = id % nbservers
        startwg.Wait()
        for volatile.ReadUint32(&running) > 0 {
            servers[target] <- message{0, ret}
            <-ret
            nbflow++
        }
    }

    // Add statistics
    atomic.AddUint64(&gbflow, nbflow)
    atomic.AddUint64(&gbltnc, latency)

    barrier.Done()
}

/** Server goroutine.
 * @param id Server ID
**/
func server(id uint) {
    for {
        msg, ok := <-servers[id]
        if !ok {
            break
        }
        msg.ret <- msg
    }
}

// -----------------------------------------------------------------------------

/** Convert the method number to a string.
 * @param id Method id
 * @return String constant
**/
func methodToString(id uint) string {
    switch (id) {
    case random:
        return "random"
    case round_robin:
        return "round-robin"
    case shared:
        return "shared"
    default:
        panic("Unknow method id")
    }
}

/** Main function.
**/
func main() {
    var duration  uint64 // Test duration (ms)
    var nbclients uint   // Amount of clients
    var buffer    uint   // Per channel buffer count
    var maxprocs  int    // runtime.GOMAXPROCS parameter (0 for default)

    { // Command line parsing
        flag.Uint64Var(&duration, "d", 1000, "Test duration (ms)")
        flag.UintVar(&nbclients, "c", 1, "Amount of clients")
        flag.UintVar(&nbservers, "s", 1, "Amount of servers")
        flag.UintVar(&method, "m", random, "Communication method used (random: " + strconv.FormatUint(uint64(random), 10) + ", round-robin: " + strconv.FormatUint(uint64(round_robin), 10) + ", shared: " + strconv.FormatUint(uint64(shared), 10) + ")")
        flag.UintVar(&buffer, "b", 1, "Per server channel buffer count, must be greater than 0")
        flag.IntVar(&maxprocs, "x", 4, "runtime.GOMAXPROCS parameter, 0 for default")
        flag.Parse()

        if buffer == 0 { // Invalid buffer count
            panic("Per channel buffer count mcust be greater than 0")
        }
    }

    { // Initialize
        fmt.Println("Initialization...")
        fmt.Println("-", nbclients, "client(s) <->", nbservers, "server(s)")
        fmt.Println("- method used:", methodToString(method))
        fmt.Println("- payload size =", unsafe.Sizeof(payload), "bytes")
        fmt.Println("- message size =", unsafe.Sizeof(message), "bytes")

        runtime.GOMAXPROCS(maxprocs)
        startwg.Add(1)
        barrier.Add(int(nbclients))
        servers = make([]chan(message), nbservers)
        for i := uint(0); i < nbclients; i++ {
            go client(i)
        }
        for i := uint(0); i < nbservers; i++ {
            servers[i] = make(chan(message), buffer)
            go server(i)
        }
    }

    { // Test
        fmt.Println("Running...")
        start := time.Now() // A bit before Done() because client goroutines will stop a bit after running = 0
        startwg.Done()

        <-time.After(time.Duration(duration) * time.Millisecond) // Wait for duration

        running = 0
        duration = uint64(time.Since(start).Nanoseconds()) // Measuring approximative test duration
        barrier.Wait()
        for i := uint(0); i < nbservers; i++ { // Close servers' channels
            close(servers[i])
        }
    }

    { // Statistics (duration is here in ns)
        xchg := 2 * gbflow
        tput := float64(xchg * uint64(unsafe.Sizeof(message))) * 1000 / float64(duration)
        avgl := float64(duration) * float64(nbclients) / (1000 * float64(xchg)) // Supposing time(chan send/recv) >> time(local ops) -- not true actually! use 'shared' mode for lowest time(local ops)

        fmt.Println("Statistics...")
        fmt.Println("- message count   =", xchg, "messages")
        fmt.Println("- avg. throughput ~", tput, "MB/s")
        fmt.Println("- average latency ~", avgl, "µs") // Average latency for goroutine A -> goroutine B
    }
}
