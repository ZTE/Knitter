/*
Copyright 2018 ZTE Corporation. All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package uuid

import (
	"crypto/sha1"
	//	"crypto/md5"
	//	"crypto/rand"
	//  "encoding/binary"
	"fmt"
	"hash"
	"time"
)

type UUID [16]byte

var (
	// NIL is defined in RFC 4122 section 4.1.7.
	// The nil UUID is special form of UUID that is specified to have all 128 bits set to zero.
	NIL = &UUID{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	// NameSpaceDNS assume name to be a fully-qualified domain name.
	// Declared in RFC 4122 Appendix C.
//NameSpaceDNS = &UUID{
//	0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1,
//	0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8,
//}
// NameSpaceURL assume name to be a URL.
// Declared in RFC 4122 Appendix C.
//NameSpaceURL = &UUID{
//	0x6b, 0xa7, 0xb8, 0x11, 0x9d, 0xad, 0x11, 0xd1,
//	0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8,
//}
// NameSpaceOID assume name to be an ISO OID.
// Declared in RFC 4122 Appendix C.
//NameSpaceOID = &UUID{
//	0x6b, 0xa7, 0xb8, 0x12, 0x9d, 0xad, 0x11, 0xd1,
//	0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8,
//}
// NameSpaceX500 assume name to be a X.500 DN (in DER or a text output format).
// Declared in RFC 4122 Appendix C.
//NameSpaceX500 = &UUID{
//	0x6b, 0xa7, 0xb8, 0x14, 0x9d, 0xad, 0x11, 0xd1,
//	0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8,
//}
)

// Version of the UUID represents a kind of subtype specifier.
//func (u *UUID) Version() int {
//	return int(binary.BigEndian.Uint16(u[6:8]) >> 12)
//}

// String returns the human readable form of the UUID.
func (u *UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

func (u *UUID) variantRFC4122() {
	u[8] = (u[8] & 0x3f) | 0x80
}

func newByHash(hash hash.Hash, namespace *UUID, name []byte) *UUID {
	hash.Write(namespace[:])
	hash.Write(name[:])

	var uuid UUID
	copy(uuid[:], hash.Sum(nil)[:16])
	uuid.variantRFC4122()
	return &uuid
}

// NewV5 creates a new UUID with variant 5 as described in RFC 4122.
// Variant 5 based namespace-uuid and name and SHA-1 hash calculation.
func NewV5(namespaceUUID *UUID, name []byte) *UUID {
	uuid := newByHash(sha1.New(), namespaceUUID, name)
	uuid[6] = (uuid[6] & 0x0f) | 0x50
	return uuid
}

/*
func TestNewV5(t *testing.T) {
	namespace := NewNamespaceUUID("test")
	uuid := NewV5(namespace, []byte("test name"))
	if uuid.Version() != 5 {
		t.Errorf("invalid version %d - expected 5", uuid.Version())
	}
	t.Logf("UUID V5: %s", uuid)
}
*/
/*
// NewV4 creates a new UUID with variant 4 as described in RFC 4122.
// Variant 4 based on pure random bytes.
func NewV4() *UUID {
	buf := make([]byte, 16)
	rand.Read(buf)
	buf[6] = (buf[6] & 0x0f) | 0x40
	var uuid UUID
	copy(uuid[:], buf[:])
	uuid.variantRFC4122()
	return &uuid
}
/*
func TestNewV4(t *testing.T) {
	uuid := NewV4()
	if uuid.Version() != 4 {
		t.Errorf("invalid version %d - expected 4", uuid.Version())
	}
	t.Logf("UUID V4: %s", uuid)
}
*/
/*
// NewV3 creates a new UUID with variant 3 as described in RFC 4122.
// Variant 3 based namespace-uuid and name and MD-5 hash calculation.
func NewV3(namespace *UUID, name []byte) *UUID {
	uuid := newByHash(md5.New(), namespace, name)
	uuid[6] = (uuid[6] & 0x0f) | 0x30
	return uuid
}
/*
func TestNewV3(t *testing.T) {
	namespace := NewNamespaceUUID("test")
	uuid := NewV3(namespace, []byte("test name"))
	if uuid.Version() != 3 {
		t.Errorf("invalid version %d - expected 3", uuid.Version())
	}
	t.Logf("UUID V3: %s", uuid)
}
*/
// NewNamespaceUUID creates a namespace UUID by using the namespace name in the NIL name space.
// This is a different approach as the 4 "standard" namespace UUIDs which are timebased UUIDs (V1).
//func NewNamespaceUUID(namespace string) string {
//	namespace = time.Now().String()
//	return NewV5(NIL, []byte(namespace)).String()
//}
func NewUUID() string {
	namespace := time.Now().String()
	return NewV5(NIL, []byte(namespace)).String()
}

func GetUUID8Byte(s string) string {
	if len(s) < 8 {
		return s
	}
	return string([]byte(s)[:8])
}

/*
func ExampleNewNamespaceUUID() {
	fmt.Printf("UUID(test):        %s\n", NewNamespaceUUID("test"))
	fmt.Printf("UUID(myNameSpace): %s\n", NewNamespaceUUID("myNameSpace"))
	// Output:
	// UUID(test):        e8b764da-5fe5-51ed-8af8-c5c6eca28d7a
	// UUID(myNameSpace): 40e41e4d-01d6-5e36-8c6b-93edcdf1442d
	//
}
*/
