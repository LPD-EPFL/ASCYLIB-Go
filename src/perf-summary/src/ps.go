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
    dataset float64
    tools   float64
    gortime float64
    crtime  float64
    kernel  float64
    pertool map[string]float64
}

func (acc *account) init() {
    acc.pertool = make(map[string]float64)
}

func (acc *account) digest(line string) {
    line = strings.TrimLeft(line, " ")

    percIndex := strings.IndexRune(line, '%')
    if percIndex == -1 {
        panic("Percentage not found")
    }
    perc, err := strconv.ParseFloat(line[:percIndex], 32)
    if err != nil {
        panic("Invalid percentage: " + err.Error())
    }

    if strings.Index(line, "[k]") != -1 {
        acc.kernel += perc
    } else {
        line = strings.TrimLeft(line[percIndex + 1:], " ")

        nameOffset := strings.Index(line, "[.]")
        if nameOffset == -1 {
            panic("Malformed line")
        }

        symbol := strings.Trim(line[nameOffset + 3:], " ")

        if strings.Index(symbol, "go_dataset.") != -1 {
            acc.dataset += perc
        } else {
            goOffset := strings.Index(symbol, "go_")
            if goOffset == 0 {
                acc.tools += perc
                ptOffset := strings.IndexRune(symbol, '.')
                if ptOffset == -1 {
                    panic("Unexpected symbol: " + symbol)
                }
                tool := symbol[goOffset + 3:ptOffset]
                v, ok := acc.pertool[tool]
                if ok {
                    acc.pertool[tool] = v + perc
                } else {
                    acc.pertool[tool] = perc
                }
            } else {
                commandEndOffset := strings.IndexRune(line, ' ')
                command := line[:commandEndOffset]
                if strings.Index(line[commandEndOffset:], command) != -1 {
                    acc.gortime += perc
                } else {
                    acc.crtime += perc
                }
            }
        }
    }
}

func (acc *account) output() {
    fmt.Printf("In dataset\t%.2f%%\n", acc.dataset)
    fmt.Printf("In tools\t%.2f%%\n", acc.tools)
    for k, v := range acc.pertool {
        fmt.Printf("  - %s\t%.2f%%\n", k, v)
    }
    fmt.Printf("In runtime\t%.2f%%\n", acc.gortime + acc.crtime)
    fmt.Printf("  - Go runtime\t%.2f%%\n", acc.gortime)
    fmt.Printf("  - C  runtime\t%.2f%%\n", acc.crtime)
    fmt.Printf("In kernel\t%.2f%%\n", acc.kernel)
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
