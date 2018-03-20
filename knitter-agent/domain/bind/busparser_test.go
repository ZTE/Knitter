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

package bind

import (
	"errors"
	"github.com/ZTE/Knitter/knitter-agent/domain/adapter"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/infra/os-encap"
	. "github.com/golang/gostub"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetC0ImageNameSucc(t *testing.T) {
	outputs := []Output{
		{StubVals: Values{"", nil}, Times: 2},
		{StubVals: Values{"c0ImageName", nil}},
	}
	stubs := StubFuncSeq(&adapter.ReadLeafFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	convey.Convey("TestGetC0ImageNameSucc\n", t, func() {
		c0ImageName := GetC0ImageName()

		convey.So(c0ImageName, convey.ShouldEqual, "c0ImageName")
	})
}

func TestGetC0ImageNameErr1(t *testing.T) {
	outputs := []Output{
		{StubVals: Values{"", nil}, Times: 3},
		{StubVals: Values{"c0ImageName", nil}},
	}
	stubs := StubFuncSeq(&adapter.ReadLeafFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	convey.Convey("TestGetC0ImageNameErr1\n", t, func() {
		c0ImageName := GetC0ImageName()

		convey.So(c0ImageName, convey.ShouldEqual, "")
	})
}

func TestGetC0ImageNameErr2(t *testing.T) {
	stubs := StubFunc(&adapter.ReadLeafFromDb, "", nil)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	convey.Convey("TestGetC0ImageNameErr2\n", t, func() {
		c0ImageName := GetC0ImageName()

		convey.So(c0ImageName, convey.ShouldEqual, "")
	})
}

func TestDelNousedOvsBrInterfaces(t *testing.T) {
	convey.Convey("TestDelNousedOvsBrInterfaces", t, func() {
		convey.Convey("SUCC test", func() {
			ports := []string{"vethOtest1", "vethOtest2", "vethOtest3", "vethOtest4"}
			stubs := StubFunc(&getAllBrintIntfcs, ports, nil)
			ifs := []string{"vethOtest3", "vethOtest4"}
			stubs.StubFunc(&ListNousedOvsBrInterfaces, ifs)
			stubs.StubFunc(&DelOvsBrInterfaces)
			defer stubs.Reset()

			brName := "br-int"
			err := DelNousedOvsBrInterfaces(brName)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("FAIL test: getAllBrintIntfcs error", func() {
			errStr := "ovs exec failed"
			stubs := StubFunc(&getAllBrintIntfcs, nil, errors.New(errStr))
			defer stubs.Reset()

			brName := "br-int"
			err := DelNousedOvsBrInterfaces(brName)
			convey.So(err.Error(), convey.ShouldEqual, errStr)
		})

		convey.Convey("SUCC test: ListNousedOvsBrInterfaces error", func() {
			ports := []string{"vethOtest1", "vethOtest2", "vethOtest3", "vethOtest4"}
			stubs := StubFunc(&getAllBrintIntfcs, ports, nil)
			stubs.StubFunc(&ListNousedOvsBrInterfaces, nil)
			defer stubs.Reset()

			brName := "br-int"
			err := DelNousedOvsBrInterfaces(brName)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

var output1 = `
_uuid               : 26adcf30-c022-47a6-b89c-cb1086e6d818
admin_state         : up
bfd                 : {}
bfd_status          : {}
cfm_fault           : []
cfm_fault_status    : []
cfm_flap_count      : []
cfm_health          : []
cfm_mpid            : []
cfm_remote_mpids    : []
cfm_remote_opstate  : []
duplex              : full
error               : []
external_ids        : {}
ifindex             : 23
ingress_policing_burst: 0
ingress_policing_rate: 0
lacp_current        : []
link_resets         : 1
link_speed          : 10000000000
link_state          : up
lldp                : {}
mac                 : []
mac_in_use          : "4a:21:fa:1e:ac:d1"
mtu                 : 1500
name                : "vethO1HTc3B3"
ofport              : 7
ofport_request      : []
options             : {}
other_config        : {}
statistics          : {collisions=0, rx_bytes=648, rx_crc_err=0, rx_dropped=0, rx_errors=0, rx_frame_err=0, rx_over_err=0, rx_packets=8, tx_bytes=648, tx_dropped=0, tx_errors=0, tx_packets=8}
status              : {driver_name=veth, driver_version="1.0", firmware_version=""}
type                : ""
`

var output2 = `_uuid               : 96f836f1-9432-4d2b-bb3e-eb8bda7666f0
admin_state         : up
bfd                 : {}
bfd_status          : {}
cfm_fault           : []
cfm_fault_status    : []
cfm_flap_count      : []
cfm_health          : []
cfm_mpid            : []
cfm_remote_mpids    : []
cfm_remote_opstate  : []
duplex              : []
error               : []
external_ids        : {}
ifindex             : 19
ingress_policing_burst: 0
ingress_policing_rate: 0
lacp_current        : []
link_resets         : 1
link_speed          : []
link_state          : up
lldp                : {}
mac                 : []
mac_in_use          : "fa:16:3e:49:28:c8"
mtu                 : 1500
name                : "eth35972"
ofport              : 4
ofport_request      : []
options             : {}
other_config        : {}
statistics          : {collisions=0, rx_bytes=356, rx_crc_err=0, rx_dropped=0, rx_errors=0, rx_frame_err=0, rx_over_err=0, rx_packets=4, tx_bytes=1226, tx_dropped=0, tx_errors=0, tx_packets=15}
status              : {driver_name=virtio_net, driver_version="1.0.0", firmware_version=""}
type                : ""
`

var output3 = `_uuid               : cc6ec592-c58e-434d-9209-38f79cd8ce0b
admin_state         : []
bfd                 : {}
bfd_status          : {}
cfm_fault           : []
cfm_fault_status    : []
cfm_flap_count      : []
cfm_health          : []
cfm_mpid            : []
cfm_remote_mpids    : []
cfm_remote_opstate  : []
duplex              : []
error               : "could not open network device vethOtest3 (No such device)"
external_ids        : {}
ifindex             : []
ingress_policing_burst: 0
ingress_policing_rate: 0
lacp_current        : []
link_resets         : 0
link_speed          : []
link_state          : []
lldp                : {}
mac                 : []
mac_in_use          : []
mtu                 : []
name                : vethOtest3
ofport              : -1
ofport_request      : []
options             : {}
other_config        : {}
statistics          : {}
status              : {}
type                : ""
`

var output4 = `_uuid               : cc6ec592-c58e-434d-9209-38f79cd8c123
admin_state         : []
bfd                 : {}
bfd_status          : {}
cfm_fault           : []
cfm_fault_status    : []
cfm_flap_count      : []
cfm_health          : []
cfm_mpid            : []
cfm_remote_mpids    : []
cfm_remote_opstate  : []
duplex              : []
error               : "could not open network device vethOtest4 (No such device)"
external_ids        : {}
ifindex             : []
ingress_policing_burst: 0
ingress_policing_rate: 0
lacp_current        : []
link_resets         : 0
link_speed          : []
link_state          : []
lldp                : {}
mac                 : []
mac_in_use          : []
mtu                 : []
name                : vethOtest4
ofport              : -1
ofport_request      : []
options             : {}
other_config        : {}
statistics          : {}
status              : {}
type                : ""
`

func TestListNousedOvsBrInterfaces(t *testing.T) {
	ports := []string{"vethOtest1", "vethOtest2", "vethOtest3", "vethOtest4"}

	convey.Convey("TestListNousedOvsBrInterfaces", t, func() {
		convey.Convey("SUCC test", func() {
			convey.Convey("empty input test", func() {
				ifs := ListNousedOvsBrInterfaces(nil)
				convey.So(ifs, convey.ShouldBeEmpty)
			})

			convey.Convey("normal test", func() {
				expIfs := []string{"vethOtest3", "vethOtest4"}
				outputs := []Output{
					{StubVals: Values{output1, nil}},
					{StubVals: Values{output2, nil}},
					{StubVals: Values{output3, nil}},
					{StubVals: Values{output4, nil}},
				}
				stubs := StubFuncSeq(&osencap.Exec, outputs)
				defer stubs.Reset()
				ifs := ListNousedOvsBrInterfaces(ports)
				convey.So(ifs, convey.ShouldResemble, expIfs)
			})
		})

		convey.Convey("EXCEPTION test", func() {
			convey.Convey("normal interface EXCEPTION test", func() {
				expIfs := []string{"vethOtest3", "vethOtest4"}
				outputs := []Output{
					{StubVals: Values{output1, nil}},
					{StubVals: Values{"ovs-vsctl excute failed", errors.New("ovsdb error")}},
					{StubVals: Values{output3, nil}},
					{StubVals: Values{output4, nil}},
				}
				stubs := StubFuncSeq(&osencap.Exec, outputs)
				defer stubs.Reset()
				ifs := ListNousedOvsBrInterfaces(ports)
				convey.So(ifs, convey.ShouldResemble, expIfs)
			})

			convey.Convey("detached interface EXCEPTION test", func() {
				expIfs := []string{"vethOtest4"}
				outputs := []Output{
					{StubVals: Values{output1, nil}},
					{StubVals: Values{output2, nil}},
					{StubVals: Values{"ovs-vsctl excute failed", errors.New("ovsdb error")}},
					{StubVals: Values{output4, nil}},
				}
				stubs := StubFuncSeq(&osencap.Exec, outputs)
				defer stubs.Reset()
				ifs := ListNousedOvsBrInterfaces(ports)
				convey.So(ifs, convey.ShouldResemble, expIfs)
			})

			convey.Convey("empty result test", func() {
				expIfs := []string{}
				outputs := []Output{
					{StubVals: Values{"ovs-vsctl excute failed", errors.New("ovsdb error")}, Times: 4},
				}
				stubs := StubFuncSeq(&osencap.Exec, outputs)
				defer stubs.Reset()
				ifs := ListNousedOvsBrInterfaces(ports)
				convey.So(ifs, convey.ShouldResemble, expIfs)
			})
		})
	})
}

func TestDelOvsBrInterfaces(t *testing.T) {
	outputs := []Output{
		{StubVals: Values{"delete port succ", nil}, Times: 2},
		{StubVals: Values{"delete port fail", errors.New("ovs internal error")}, Times: 2},
	}
	stubs := StubFuncSeq(&osencap.Exec, outputs)
	defer stubs.Reset()

	brName := "br-int"
	ifs := []string{"vethOtest1", "vethOtest2", "vethOtest3", "vethOtest4"}
	DelOvsBrInterfaces(brName, ifs)
}

func Test_getAllBrNames(t *testing.T) {
	output := `
	br_api
	br_mgt
	br_eio
	`
	stubs := StubFunc(&osencap.Exec, output, nil)
	defer stubs.Reset()
	expBrNames := []string{"br_api", "br_mgt", "br_eio"}

	convey.Convey("TestgetAllBrNames", t, func() {
		brNames, err := getAllBrNames()
		convey.So(err, convey.ShouldBeNil)
		convey.So(brNames, convey.ShouldResemble, expBrNames)
	})
}

func Test_getAllBrNames_Failed(t *testing.T) {
	errStr := "ovsdb not ready"
	stubs := StubFunc(&osencap.Exec, "", errors.New(errStr))
	defer stubs.Reset()

	convey.Convey("Test_getAllBrNames_Failed", t, func() {
		brNames, err := getAllBrNames()
		convey.So(err.Error(), convey.ShouldEqual, errStr)
		convey.So(brNames, convey.ShouldBeNil)
	})
}
