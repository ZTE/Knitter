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

package models

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestIsValidIps(t *testing.T) {

	Convey("TestIsValidIps1:", t, func() {
		ig := &IPGroup{IPs: "", SizeStr: ""}
		b := ig.IsValidIPs()
		So(b, ShouldEqual, true)
	})

	Convey("TestIsValidIps2:", t, func() {
		ig := &IPGroup{IPs: "[1.1.1.1]", SizeStr: "1"}
		b := ig.IsValidIPs()
		So(b, ShouldEqual, false)
	})

	Convey("TestIsValidIps3:", t, func() {
		ig := &IPGroup{SubnetCidr: "1.1.1.0/24",
			SubnetPools: []AllocationPool{{Start: "1.1.1.1", End: "1.1.1.253"}},
			IPs:         "[1.1.1.1]",
			SizeStr:     ""}
		b := ig.IsValidIPs()
		So(b, ShouldEqual, true)
		So(ig.Mode, ShouldEqual, AddressMode)
	})

	Convey("TestIsValidIps4:", t, func() {
		ig := &IPGroup{IPs: "", SizeStr: "1"}
		b := ig.IsValidIPs()
		So(b, ShouldEqual, true)
		So(ig.Mode, ShouldEqual, CountMode)
	})
}

func TestAnalyzeUpdateIps(t *testing.T) {

	Convey("TestAnalyzeUpdateIps1:", t, func() {
		ig := &IPGroup{Mode: AddressMode, IPsSlice: []string{"1.1.1.1"}}
		err := ig.AnalyzeIPs(nil)
		So(err, ShouldBeNil)
		So(len(ig.AddIPs), ShouldEqual, 1)
		So(ig.AddIPs[0], ShouldEqual, "1.1.1.1")
	})

	Convey("TestAnalyzeUpdateIps2:", t, func() {
		ig := &IPGroup{Mode: CountMode, Size: 0}
		err := ig.AnalyzeIPs(&IPGroupInDB{IPs: []IPInDB{{IPAddr: "1.1.1.1", Used: false, PortID: "port0"}}})
		So(err, ShouldBeNil)
		So(len(ig.DelIPs), ShouldEqual, 1)
		So(ig.DelIPs["1.1.1.1"], ShouldEqual, "port0")
	})
}

func TestMakeBulkPortsReq(t *testing.T) {

	Convey("TestMakeBulkPortsReq1:", t, func() {
		ig := &IPGroup{Mode: AddressMode, AddIPs: []string{"1.1.1.1"}}
		ports := ig.makeBulkPortsReq()
		So(ports, ShouldNotBeNil)
		So(len(ports.Ports), ShouldEqual, 1)
		So(ports.Ports[0].FixIP, ShouldEqual, "1.1.1.1")
	})

	Convey("TestMakeBulkPortsReq2:", t, func() {
		ig := &IPGroup{Mode: CountMode, AddSize: 1}
		ports := ig.makeBulkPortsReq()
		So(ports, ShouldNotBeNil)
		So(len(ports.Ports), ShouldEqual, 1)
		So(ports.Ports[0].FixIP, ShouldBeEmpty)
	})
}
