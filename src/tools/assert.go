/**
 * @file   assert.go
 * @author Sébastien Rouault <sebastien.rouault@epfl.ch>
 *
 * @section LICENSE
 *
 * Copyright (c) 2014 Sébastien Rouault <sebastien.rouault@epfl.ch>
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
 * Implementation of a simple assert function in Go.
**/

package assert

import (
    "fmt"
    "runtime"
    "path"
)

// -----------------------------------------------------------------------------

/** Print a message and exit the program if the condition is not true.
 * @param cond Condition tested
 * @param text Text to print before quitting if cond == false
**/
func Assert(cond bool, text string) {
    if !cond {
        _, file, line, ok := runtime.Caller(1)
        if ok {
            panic(fmt.Sprintf("Assertion failed at %s:%d (%s)\n", path.Base(file), line, text))
        } else {
            panic(fmt.Sprintf("Assertion failed (%s)\n", text))
        }
    }
}
