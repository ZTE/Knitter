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

package networkserver

import (
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-iaas"
	. "github.com/golang/gostub"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	"github.com/smartystreets/goconvey/convey"
	"net"
	"testing"
)

/*
NOTE:
1. Create network success, but create subnet error, will lost network resource.

2. Free vxlan ID success, but delete network error, this network can not delete.
*/

func TestNetwork(t *testing.T) {
	stubsSaveData := StubFunc(&SaveData, nil)
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	stubsDeleteData := StubFunc(&DeleteData, nil)
	defer stubsSaveData.Reset()
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	defer stubsDeleteData.Reset()
	var newNetworkName string = "Create-Network-For-TEST"
	var newNetworkID string
	m := GetEmbeddedNetwrokManager()

	convey.Convey("TestCreateNetwork---OK\n", t, func() {
		netNet, err := m.CreateNetwork(newNetworkName)
		convey.So(err, convey.ShouldEqual, nil)
		fmt.Print("\nNew-Network-ID:" + netNet.Id)
		convey.So(netNet.Id, convey.ShouldNotEqual, "")
		newNetworkID = netNet.Id
	})

	convey.Convey("TestGetNetworkID---OK\n", t, func() {
		id, err := m.GetNetworkID(newNetworkName)
		convey.So(err, convey.ShouldEqual, nil)
		fmt.Print("\nNew-Network-ID:" + id)
		fmt.Print("\nWant-Network-ID:" + newNetworkID)
		convey.So(id, convey.ShouldEqual, newNetworkID)
	})

	convey.Convey("TestGetNetwork---OK\n", t, func() {
		net, err := m.GetNetwork(newNetworkID)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(net, convey.ShouldNotEqual, nil)
		fmt.Print("\nID:" + net.Id)
		fmt.Print("\nNAME:" + net.Name)
		convey.So(net.Id, convey.ShouldEqual, newNetworkID)
		convey.So(net.Name, convey.ShouldEqual, newNetworkName)
	})

	convey.Convey("TestGetNetworkExtenAttrs---OK\n", t, func() {
		net, err := m.GetNetworkExtenAttrs(newNetworkID)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(net, convey.ShouldNotEqual, nil)
		fmt.Print("\nID:" + net.Id)
		fmt.Print("\nNAME:" + net.Name)
		fmt.Print("\nNetworkType:" + net.NetworkType)
		fmt.Print("\nPhysicalNetwork:" + net.PhysicalNetwork)
		fmt.Print("\nSegmentationID:" + net.SegmentationID)
		convey.So(net.Id, convey.ShouldEqual, newNetworkID)
		convey.So(net.Name, convey.ShouldEqual, newNetworkName)
	})

	convey.Convey("TestDeleteNetwork---OK\n", t, func() {
		err := m.DeleteNetwork(newNetworkID)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestSubnet(t *testing.T) {
	stubsSaveData := StubFunc(&SaveData, nil)
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	stubsDeleteData := StubFunc(&DeleteData, nil)
	defer stubsSaveData.Reset()
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	defer stubsDeleteData.Reset()
	var newNetworkName string = "Create-Network-For-TEST"
	var newNetworkID, newSubnetID string
	m := GetEmbeddedNetwrokManager()
	convey.Convey("TestCreateNetwork---OK\n", t, func() {
		netNet, err := m.CreateNetwork(newNetworkName)
		convey.So(err, convey.ShouldEqual, nil)
		fmt.Print("\nNew-Network-ID:" + netNet.Id)
		convey.So(netNet.Id, convey.ShouldNotEqual, "")
		newNetworkID = netNet.Id
	})

	convey.Convey("TestCreateSubnet---OK\n", t, func() {
		subNet, err := m.CreateSubnet(newNetworkID,
			"192.168.1.1/24", "192.168.1.1", []subnets.AllocationPool{})
		convey.So(err, convey.ShouldEqual, nil)
		fmt.Print("\nNew-Subnet-ID:" + subNet.Id)
		convey.So(subNet.Id, convey.ShouldNotEqual, "")
		newSubnetID = subNet.Id
	})

	convey.Convey("TestGetSubnetID---OK\n", t, func() {
		id, err := m.GetSubnetID(newNetworkID)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(id, convey.ShouldEqual, newSubnetID)
	})

	convey.Convey("TestGetSubnet---OK\n", t, func() {
		sub, err := m.GetSubnet(newSubnetID)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(sub.Id, convey.ShouldEqual, newSubnetID)
		fmt.Print("\n-------Subnet-----------")
		fmt.Print("\nID:" + sub.Id)
		fmt.Print("\nNAME:" + sub.Name)
		fmt.Print("\nCidr:" + sub.Cidr)
		fmt.Print("\nGatewayIp:" + sub.GatewayIp)
		fmt.Print("\nNetworkId:" + sub.NetworkId)
		fmt.Print("\nTenantId:" + sub.TenantId)
	})

	convey.Convey("TestDeleteSubnet---OK\n", t, func() {
		err := m.DeleteSubnet(newSubnetID)
		convey.So(err, convey.ShouldEqual, nil)
	})

	convey.Convey("TestDeleteNetwork---OK\n", t, func() {
		err := m.DeleteNetwork(newNetworkID)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestPort(t *testing.T) {
	stubsSaveData := StubFunc(&SaveData, nil)
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	stubsDeleteData := StubFunc(&DeleteData, nil)
	defer stubsSaveData.Reset()
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	defer stubsDeleteData.Reset()
	var newNetworkName string = "Create-Network-For-TEST"
	var newNetworkID, newSubnetID, newPortID string
	m := GetEmbeddedNetwrokManager()
	convey.Convey("TestCreateNetwork---OK\n", t, func() {
		netNet, err := m.CreateNetwork(newNetworkName)
		convey.So(err, convey.ShouldEqual, nil)
		fmt.Print("\nNew-Network-ID:" + netNet.Id)
		convey.So(netNet.Id, convey.ShouldNotEqual, "")
		newNetworkID = netNet.Id
	})

	convey.Convey("TestCreateSubnet---OK\n", t, func() {
		subNet, err := m.CreateSubnet(newNetworkID,
			"192.168.1.1/24", "192.168.1.1", []subnets.AllocationPool{})
		convey.So(err, convey.ShouldEqual, nil)
		fmt.Print("\nNew-Subnet-ID:" + subNet.Id)
		convey.So(subNet.Id, convey.ShouldNotEqual, "")
		newSubnetID = subNet.Id
	})

	convey.Convey("TestCreatePort---OK\n", t, func() {
		port, err := m.CreatePort(newNetworkID, newSubnetID,
			"std", "", "", "")
		convey.So(err, convey.ShouldEqual, nil)
		fmt.Print("\nNew-Port-ID:" + port.Id)
		convey.So(port, convey.ShouldNotEqual, nil)
		newPortID = port.Id
	})

	convey.Convey("TestGetPort---OK\n", t, func() {
		port, err := m.GetPort(newPortID)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(port.Id, convey.ShouldEqual, newPortID)
		fmt.Print("\n-------Port-----------")
		fmt.Print("\nID:" + port.Id)
		fmt.Print("\nNAME:" + port.Name)
		fmt.Print("\nIP:" + port.Ip)
		fmt.Print("\nMAC:" + port.MacAddress)
		fmt.Print("\nNetworkId:" + port.NetworkId)
		fmt.Print("\nSubnetId:" + port.SubnetId)
	})

	convey.Convey("TestDeletePort---OK\n", t, func() {
		err := m.DeletePort(newPortID)
		convey.So(err, convey.ShouldEqual, nil)
	})

	convey.Convey("TestDeleteSubnet---OK\n", t, func() {
		err := m.DeleteSubnet(newSubnetID)
		convey.So(err, convey.ShouldEqual, nil)
	})

	convey.Convey("TestDeleteNetwork---OK\n", t, func() {
		err := m.DeleteNetwork(newNetworkID)
		convey.So(err, convey.ShouldEqual, nil)
	})

	convey.Convey("CreateBulkPorts---Err0\n", t, func() {
		ports := []*mgriaas.MgrPortReq{{AgtPortReq: agtmgr.AgtPortReq{PortName: "port"}, TenantId: "tenantId", NetworkId: "net", SubnetId: "subnet"}}
		req := &mgriaas.MgrBulkPortsReq{Ports: ports}
		_, err := m.CreateBulkPorts(req)
		convey.So(err, convey.ShouldNotEqual, nil)
	})

	convey.Convey("CreateBulkPorts---IpOffsetInvalidErr\n", t, func() {
		ports := []*mgriaas.MgrPortReq{{AgtPortReq: agtmgr.AgtPortReq{PortName: "port", FixIP: "100.100.0.1"}, TenantId: "tenant", NetworkId: "net", SubnetId: "subnet"}}
		req := &mgriaas.MgrBulkPortsReq{Ports: ports}
		m.net.list["net"] = &NetworkExtenAttrs{ID: "net"}
		m.sub.list["subnet"] = &PaasSubnet{Sub: &iaasaccessor.Subnet{Id: "subnet", NetworkId: "net", TenantId: "tenant", Cidr: "100.100.0.0/16"}}
		_, err := m.CreateBulkPorts(req)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "ipaddr-offset-invalid")
	})

	convey.Convey("CreateBulkPorts---IpUsedErr\n", t, func() {
		ports := []*mgriaas.MgrPortReq{{AgtPortReq: agtmgr.AgtPortReq{PortName: "port", FixIP: "100.100.0.2"}, TenantId: "tenant", NetworkId: "net", SubnetId: "subnet"}}
		req := &mgriaas.MgrBulkPortsReq{Ports: ports}
		m.net.list["net"] = &NetworkExtenAttrs{ID: "net"}
		m.sub.list["subnet"] = &PaasSubnet{Sub: &iaasaccessor.Subnet{Id: "subnet", NetworkId: "net", TenantId: "tenant", Cidr: "100.100.0.0/16"}}
		m.sub.list["subnet"].Pool = make(map[string]*net.IP, 1)
		ip := net.ParseIP("100.100.0.2")
		m.sub.list["subnet"].Pool["2"] = &ip
		_, err := m.CreateBulkPorts(req)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "ip-is-invalid-or-in-use")
	})

	convey.Convey("CreateBulkPorts---IpNotInPoolErr\n", t, func() {
		ports := []*mgriaas.MgrPortReq{{AgtPortReq: agtmgr.AgtPortReq{PortName: "port", FixIP: "100.10.0.2"}, TenantId: "tenant", NetworkId: "net", SubnetId: "subnet"}}
		req := &mgriaas.MgrBulkPortsReq{Ports: ports}
		m.net.list["net"] = &NetworkExtenAttrs{ID: "net"}
		m.sub.list["subnet"] = &PaasSubnet{Sub: &iaasaccessor.Subnet{Id: "subnet", NetworkId: "net", TenantId: "tenant", Cidr: "100.100.0.0/16"}}
		_, err := m.CreateBulkPorts(req)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "ipaddr-not-in-CIDR")
	})

	convey.Convey("CreateBulkPorts---OK\n", t, func() {
		ports := []*mgriaas.MgrPortReq{{AgtPortReq: agtmgr.AgtPortReq{PortName: "port", FixIP: "100.100.0.2"}, TenantId: "tenant", NetworkId: "net", SubnetId: "subnet"}}
		req := &mgriaas.MgrBulkPortsReq{Ports: ports}
		m.net.list["net"] = &NetworkExtenAttrs{ID: "net"}
		m.sub.list["subnet"] = &PaasSubnet{Sub: &iaasaccessor.Subnet{Id: "subnet", NetworkId: "net", TenantId: "tenant", Cidr: "100.100.0.0/16"}}
		_, err := m.CreateBulkPorts(req)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestOthers(t *testing.T) {
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	m := GetEmbeddedNetwrokManager()
	convey.Convey("TestGetAttachReq---OK\n", t, func() {
		number := m.GetAttachReq()
		convey.So(number, convey.ShouldEqual, 0)
	})

	convey.Convey("TestGetAttachReq---OK\n", t, func() {
		m.SetAttachReq(3)
	})

	convey.Convey("TestCreateBulkPorts---OK\n", t, func() {
		prot, err := m.CreateBulkPorts(nil)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(prot, convey.ShouldEqual, nil)
	})

	convey.Convey("TestGetTenantUUID---OK\n", t, func() {
		id, err := m.GetTenantUUID("")
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(id, convey.ShouldNotEqual, "")
	})

	convey.Convey("TestAuth---OK\n", t, func() {
		err := m.Auth()
		convey.So(err, convey.ShouldEqual, nil)
	})

	convey.Convey("TestGetType---OK\n", t, func() {
		t := m.GetType()
		convey.So(t, convey.ShouldEqual, "EMBEDDED")
	})

	convey.Convey("TestCreateRouter---OK\n", t, func() {
		r, e := m.CreateRouter("", "")
		convey.So(r, convey.ShouldEqual, "")
		convey.So(e, convey.ShouldNotEqual, nil)
	})

	convey.Convey("TestUpdateRouter---OK\n", t, func() {
		e := m.UpdateRouter("", "", "")
		convey.So(e, convey.ShouldNotEqual, nil)
	})

	convey.Convey("TestGetRouter---OK\n", t, func() {
		r, e := m.GetRouter("")
		convey.So(e, convey.ShouldNotEqual, nil)
		convey.So(r, convey.ShouldEqual, nil)
	})

	convey.Convey("TestDeleteRouter---OK\n", t, func() {
		e := m.DeleteRouter("")
		convey.So(e, convey.ShouldNotEqual, nil)
	})

	convey.Convey("TestAttachPortToVM---OK\n", t, func() {
		i, e := m.AttachPortToVM("", "")
		convey.So(e, convey.ShouldNotEqual, nil)
		convey.So(i, convey.ShouldEqual, nil)
	})

	convey.Convey("TestDetachPortFromVM---OK\n", t, func() {
		e := m.DetachPortFromVM("", "")
		convey.So(e, convey.ShouldNotEqual, nil)
	})

	convey.Convey("TestAttachNetToRouter---OK\n", t, func() {
		i, e := m.AttachNetToRouter("", "")
		convey.So(e, convey.ShouldNotEqual, nil)
		convey.So(i, convey.ShouldEqual, "")
	})

	convey.Convey("TestDetachNetFromRouter---OK\n", t, func() {
		i, e := m.DetachNetFromRouter("", "")
		convey.So(e, convey.ShouldNotEqual, nil)
		convey.So(i, convey.ShouldEqual, "")
	})

	convey.Convey("TestCreateProviderNetwork---OK\n", t, func() {
		s, e := m.CreateProviderNetwork("", "", "", "", false)
		convey.So(e, convey.ShouldNotEqual, nil)
		convey.So(s, convey.ShouldBeNil)
	})
}
