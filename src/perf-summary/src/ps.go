/**
 * @file   ps.go
 * @author Sébastien Rouault <sebastien.rouault@epfl.ch>
 *
 * @section LICENSE
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * any later version. Please see https://gnu.org/licenses/gpl.html
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * @section DESCRIPTION
 *
 * Quick'n'dirty conversion from 'perf report' dump files to simple summary.
**/

package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)

// ―――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

// Account time spend on each "subsection"
type account struct {
    total   uint64
    main    uint64
    dataset uint64
    tools   uint64
    gortime uint64
    crtime  uint64
    pertool map[string]uint64
}

func (acc *account) init() {
    acc.pertool = make(map[string]uint64)
}

func (acc *account) digest(line string) {
    // Fields: Overhead, Samples, Command, Shared Object, Symbol
    fields := strings.Split(line, "\t")
    fields[4] = fields[4][4:] // Ignoring "[.] " part of the symbol name
    // Sample count
    count, err := strconv.ParseUint(fields[1][1:], 10, 64) // Trim always needed
    if err != nil {
        panic("Invalid sample count: " + err.Error())
    }
    acc.total += count

    // From test module (main)
    if fields[4][0:5] == "main." {
        acc.main += count
        return
    }
    // From dataset module
    if fields[4][0:11] == "go_dataset." {
        acc.dataset += count
        return
    }
    // From some tool module
    if fields[4][0:3] == "go_" {
        acc.tools += count
        ptOffset := strings.IndexRune(fields[4], '.')
        if ptOffset == -1 {
            panic("Unexpected symbol: " + fields[4])
        }
        tool := fields[4][3:ptOffset]
        v, ok := acc.pertool[tool]
        if ok {
            acc.pertool[tool] = v + count
        } else {
            acc.pertool[tool] = count
        }
        return
    }
    // From Go runtime (= from inside the main binary)
    if fields[2] == fields[3][:len(fields[2])] {
        acc.gortime += count
        return
    }
    // Else from "C runtime"
    acc.crtime += count
}

func (acc *account) output() {
    total := float64(acc.total) / 100

    fmt.Printf("In dataset\t%.2f%%\n", float64(acc.dataset) / total)
    fmt.Printf("In main/test\t%.2f%%\n", float64(acc.main) / total)
    fmt.Printf("In tools\t%.2f%%\n", float64(acc.tools) / total)
    for k, v := range acc.pertool {
        fmt.Printf("  - %s\t%.2f%%\n", k, float64(v) / total)
    }
    fmt.Printf("In runtime\t%.2f%%\n", float64(acc.gortime + acc.crtime) / total)
    fmt.Printf("  - Go runtime\t%.2f%%\n", float64(acc.gortime) / total)
    fmt.Printf("  - C  runtime\t%.2f%%\n", float64(acc.crtime) / total)
}

// ―――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

func main() {
    var acc account
    acc.init()

    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        line := scanner.Text()
        if len(line) > 0 && line[0] != '#' {
            acc.digest(line)
        }
    }

    acc.output()
}
