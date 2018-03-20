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
	. "github.com/golang/gostub"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPortOK(t *testing.T) {
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
		newNet, err := m.CreateNetwork(newNetworkName)
		convey.So(err, convey.ShouldEqual, nil)
		fmt.Print("\nNew-Network-ID:" + newNet.Id)
		convey.So(newNet.Id, convey.ShouldNotEqual, "")
		newNetworkID = newNet.Id
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
}

func TestCreatePortERR(t *testing.T) {
	stubsSaveData := StubFunc(&SaveData, nil)
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	stubsDeleteData := StubFunc(&DeleteData, nil)
	defer stubsSaveData.Reset()
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	defer stubsDeleteData.Reset()
	var newNetworkName string = "Create-Network-For-TestCreatePortERR"
	var newNetworkID, newSubnetID string
	m := GetEmbeddedNetwrokManager()
	convey.Convey("TestCreateNetwork---OK\n", t, func() {
		newNet, err := m.CreateNetwork(newNetworkName)
		convey.So(err, convey.ShouldEqual, nil)
		fmt.Print("\nNew-Network-ID:" + newNet.Id)
		convey.So(newNet.Id, convey.ShouldNotEqual, "")
		newNetworkID = newNet.Id
	})

	convey.Convey("TestCreateSubnet---OK\n", t, func() {
		subNet, err := m.CreateSubnet(newNetworkID,
			"192.168.1.1/24", "192.168.1.1", []subnets.AllocationPool{})
		convey.So(err, convey.ShouldEqual, nil)
		fmt.Print("\nNew-Subnet-ID:" + subNet.Id)
		convey.So(subNet.Id, convey.ShouldNotEqual, "")
		newSubnetID = subNet.Id
	})

	convey.Convey("TestCreatePort---ErrNetworkID\n", t, func() {
		port, err := m.CreatePort(newNetworkName, newSubnetID,
			"std", "", "", "")
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(port, convey.ShouldEqual, nil)
	})

	convey.Convey("TestCreatePort---ErrSubnetID\n", t, func() {
		port, err := m.CreatePort(newNetworkID, newNetworkID,
			"std", "", "", "")
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(port, convey.ShouldEqual, nil)
	})

	stubsSaveData = StubFunc(&SaveData, errors.New("DEL-ERROR"))
	convey.Convey("TestCreatePort---ErrDataSave\n", t, func() {
		port, err := m.CreatePort(newNetworkID, newSubnetID,
			"std", "", "", "")
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(port, convey.ShouldEqual, nil)
	})
}

func TestGetPortERR(t *testing.T) {
	stubsSaveData := StubFunc(&SaveData, nil)
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	stubsDeleteData := StubFunc(&DeleteData, nil)
	defer stubsSaveData.Reset()
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	defer stubsDeleteData.Reset()
	var newNetworkName string = "Create-Network-For-TestCreatePortERR"
	m := GetEmbeddedNetwrokManager()
	convey.Convey("TestCreatePort---ErrNetworkID\n", t, func() {
		port, err := m.GetPort(newNetworkName)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(port, convey.ShouldEqual, nil)
	})
}

func TestDeletePortERR(t *testing.T) {
	stubsSaveData := StubFunc(&SaveData, nil)
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	stubsDeleteData := StubFunc(&DeleteData, nil)
	defer stubsSaveData.Reset()
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	defer stubsDeleteData.Reset()
	var newNetworkName string = "Create-Network-For-TestDeletePortERR"
	var newNetworkID, newSubnetID, newPortID string
	m := GetEmbeddedNetwrokManager()
	convey.Convey("TestCreatePort---ErrNetworkID\n", t, func() {
		err := m.DeletePort(newNetworkName)
		convey.So(err, convey.ShouldNotEqual, nil)
	})

	convey.Convey("TestCreateNetwork---OK\n", t, func() {
		newNet, err := m.CreateNetwork(newNetworkName)
		convey.So(err, convey.ShouldEqual, nil)
		fmt.Print("\nNew-Network-ID:" + newNet.Id)
		convey.So(newNet.Id, convey.ShouldNotEqual, "")
		newNetworkID = newNet.Id
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

	stubsSaveData = StubFunc(&SaveData, errors.New("error"))
	convey.Convey("TestCreatePort---ErrNetworkID\n", t, func() {
		err := m.DeletePort(newPortID)
		convey.So(err, convey.ShouldNotEqual, nil)
	})

	stubsSaveData = StubFunc(&SaveData, nil)
	stubsDeleteData = StubFunc(&DeleteData, errors.New("error"))
	convey.Convey("TestCreatePort---ErrNetworkID\n", t, func() {
		err := m.DeletePort(newPortID)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

//func TestListPorts(t *testing.T) {
//	networkID := "net_api_id"
//	port1 := Interface{
//		Name       : "eth0",
//		ID         : "port1-id",
//		IP         : "192.168.1.101",
//		MacAddress : "",
//		NetworkID  : "net_api_id",
//		SubnetID   : "",
//	}
//	port2 := Interface{
//		Name       : "eth1",
//		ID         : "port2-id",
//		IP         : "192.168.1.102",
//		MacAddress : "",
//		NetworkID  : "net_api_id",
//		SubnetID   : "",
//	}
//	port3 := Interface{
//		Name       : "eth0",
//		ID         : "port3-id",
//		IP         : "192.168.2.103",
//		MacAddress : "",
//		NetworkID  : "net_control_id",
//		SubnetID   : "",
//	}
//	port4 := Interface{
//		Name       : "eth3",
//		ID         : "port4-id",
//		IP         : "192.168.1.103",
//		MacAddress : "",
//		NetworkID  : "net_api_id",
//		SubnetID   : "",
//	}
//
//	portManager = &Interfaces{list: make(map[string]*Interface)}
//	pm := GetPortManager()
//	pm.list[port1.ID] = &port1
//	pm.list[port2.ID] = &port2
//	pm.list[port3.ID] = &port3
//	pm.list[port4.ID] = &port4
//
//	expPorts := []*iaasaccessor.Interface{
//		{
//			Ip: port1.IP,
//			Id: port1.ID,
//			Name: port1.Name,
//			MacAddress: port1.MacAddress,
//			NetworkId: port1.NetworkID,
//			SubnetId: port1.SubnetID,
//		},
//		{
//			Ip: port2.IP,
//			Id: port2.ID,
//			Name: port2.Name,
//			MacAddress: port2.MacAddress,
//			NetworkId: port2.NetworkID,
//			SubnetId: port2.SubnetID,
//		},
//		{
//			Ip: port4.IP,
//			Id: port4.ID,
//			Name: port4.Name,
//			MacAddress: port4.MacAddress,
//			NetworkId: port4.NetworkID,
//			SubnetId: port4.SubnetID,
//		},
//	}
//
//	manager = &NetworkManager{port: portManager}
//	nm := GetEmbeddedNetwrokManager()
//	convey.Convey("TesListPorts\n", t, func() {
//		ports, err := nm.ListPorts(networkID)
//		convey.So(err, convey.ShouldBeNil)
//		convey.So(ports, convey.ShouldResemble, expPorts)
//	})
//}
