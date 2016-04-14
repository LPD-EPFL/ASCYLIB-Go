/**
 * @file   volatile.go
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
 * Simple workaround to make sure a variable is re-read in Go.
 * Rely on the fact that the linker doesn't perform further optimization steps (inlining + optimizing out 'mov (addr), val').
**/

package volatile

import (
    "unsafe"
)

// -----------------------------------------------------------------------------

/** Read an integer.
 * @param p Pointer on the integer
 * @return Value of the integer
**/
func ReadInt8(p *int8) int8 {
    return *p
}
func ReadInt16(p *int16) int16 {
    return *p
}
func ReadInt32(p *int32) int32 {
    return *p
}
func ReadInt64(p *int64) int64 {
    return *p
}
func ReadUint8(p *uint8) uint8 {
    return *p
}
func ReadUint16(p *uint16) uint16 {
    return *p
}
func ReadUint32(p *uint32) uint32 {
    return *p
}
func ReadUint64(p *uint64) uint64 {
    return *p
}

/** Read a pointer.
 * @param p Pointer on the pointer
 * @return Value of the pointer
**/
func ReadPointer(p *unsafe.Pointer) unsafe.Pointer {
    return *p
}
