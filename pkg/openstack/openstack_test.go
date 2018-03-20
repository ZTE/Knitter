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

package openstack

import (
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/pkg/adapter"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-iaas"
	. "github.com/golang/gostub"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCreateBulkPorts(t *testing.T) {
	mgrBulkPortsReq := mgriaas.MgrBulkPortsReq{Ports: []*mgriaas.MgrPortReq{
		{}, {},
	}}
	mgrBulkPortsReq.Ports[0].NetworkName = "networkName0"
	mgrBulkPortsReq.Ports[0].PortName = "portName0"
	mgrBulkPortsReq.Ports[0].TenantId = "tenantUser"
	mgrBulkPortsReq.Ports[0].NetworkId = "networkId0"
	mgrBulkPortsReq.Ports[0].SubnetId = "subnetId0"
	mgrBulkPortsReq.Ports[1].NetworkName = "networkName"
	mgrBulkPortsReq.Ports[1].PortName = "portName1"
	mgrBulkPortsReq.Ports[1].TenantId = "tenantUser"
	mgrBulkPortsReq.Ports[1].NetworkId = "networkId1"
	mgrBulkPortsReq.Ports[1].SubnetId = "subnetId1"

	ports := []*ports.Port{
		{ID: "port_id1", NetworkID: "network_id1", Status: "up",
			Name: "port_name_1", MACAddress: "00-aa-bb-cc", DeviceID: "dev_id1",
			FixedIPs: []ports.IP{{SubnetID: "subnet_id_1", IPAddress: "172.15.15.16"}}},
		{ID: "port_id2", NetworkID: "network_id2", Status: "up",
			Name: "port_name_2", MACAddress: "11-aa-bb-cc", DeviceID: "dev_id2",
			FixedIPs: []ports.IP{{SubnetID: "subnet_id_2", IPAddress: "172.15.17.17"}}},
	}

	outputs := []Output{
		{StubVals: Values{ports, nil}},
		{StubVals: Values{nil, errobj.ErrOpenstackCreateBulkPortsFailed}},
	}
	stubs := StubFuncSeq(&adapter.CreateBulkPorts, outputs)
	defer stubs.Reset()
	//stubs.StubFunc(&adapter.ExtractBulk, []*ports.Port{}, nil)
	convey.Convey("TestCreateBulkPorts for nil", t, func() {
		openStack := OpenStack{}
		interfaces, err := openStack.CreateBulkPorts(&mgrBulkPortsReq)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(interfaces[0].Id, convey.ShouldEqual, ports[0].ID)
		convey.So(interfaces[0].Name, convey.ShouldEqual, ports[0].Name)
		convey.So(interfaces[0].Status, convey.ShouldEqual, ports[0].Status)
		convey.So(interfaces[0].MacAddress, convey.ShouldEqual, ports[0].MACAddress)
		convey.So(interfaces[0].NetworkId, convey.ShouldEqual, ports[0].NetworkID)
		convey.So(interfaces[0].DeviceId, convey.ShouldEqual, ports[0].DeviceID)
		convey.So(interfaces[0].SubnetId, convey.ShouldEqual, ports[0].FixedIPs[0].SubnetID)
		convey.So(interfaces[1].Id, convey.ShouldEqual, ports[1].ID)
		convey.So(interfaces[1].Name, convey.ShouldEqual, ports[1].Name)
		convey.So(interfaces[1].Status, convey.ShouldEqual, ports[1].Status)
		convey.So(interfaces[1].MacAddress, convey.ShouldEqual, ports[1].MACAddress)
		convey.So(interfaces[1].NetworkId, convey.ShouldEqual, ports[1].NetworkID)
		convey.So(interfaces[1].DeviceId, convey.ShouldEqual, ports[1].DeviceID)
		convey.So(interfaces[1].SubnetId, convey.ShouldEqual, ports[1].FixedIPs[0].SubnetID)
	})

	convey.Convey("TestCreateBulkPorts for err", t, func() {
		openStack := OpenStack{}
		_, err := openStack.CreateBulkPorts(&mgrBulkPortsReq)
		convey.So(err, convey.ShouldEqual, errobj.ErrOpenstackCreateBulkPortsFailed)
	})
}

func Test_GetTenantID(t *testing.T) {
	convey.Convey("Test GetTenantID", t, func() {
		openStack := OpenStack{
			provider: &gophercloud.ProviderClient{
				TenantID: "tenantid",
			},
		}
		id := openStack.GetTenantID()
		convey.So(id, convey.ShouldEqual, "tenantid")
	})
}
