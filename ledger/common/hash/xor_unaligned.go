/// The functions below were copied and modified from golang.org/x/crypto/sha3.
//
// Copyright (c) 2009 The Go Authors. All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:

//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

//go:build (amd64 || 386 || ppc64le) && !purego
// +build amd64 386 ppc64le
// +build !purego

package hash

import "unsafe"

var (
//copyIn512 = copyIn512Unaligned
)

// copyOutUnaligned copies 32 bytes to a hash array.
func copyOut(d *state) Hash {
	var out Hash
	ab := (*[HashLen]uint8)(unsafe.Pointer(&d.a[0]))
	copy(out[:], ab[:])
	return out
}

// copyIn512 copies two 32 bytes slices into the state
/*func copyIn512Unaligned(d *state, buf1, buf2 Hash) {
	ab := (*[stateSize << 3]uint8)(unsafe.Pointer(&d.a[0]))
	copy(ab[:], buf1[:])
	copy(ab[HashLen:], buf2[:])

	// dsbyte also contains the first one bit for the padding.
	d.a[8] = 0x6
	// copy the last padding bit
	d.a[16] = paddingEnd
}
*/
