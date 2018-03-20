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

func TestSubnetOK(t *testing.T) {
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

func TestCreateSubnetERR(t *testing.T) {
	stubsSaveData := StubFunc(&SaveData, nil)
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	stubsDeleteData := StubFunc(&DeleteData, nil)
	defer stubsSaveData.Reset()
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	defer stubsDeleteData.Reset()
	var newNetworkName string = "Create-Network-For-TestCreateSubnetERR"
	var newNetworkID string
	m := GetEmbeddedNetwrokManager()

	convey.Convey("TestCreateNetwork---OK\n", t, func() {
		newNet, err := m.CreateNetwork(newNetworkName)
		convey.So(err, convey.ShouldEqual, nil)
		fmt.Print("\nNew-Network-ID:" + newNet.Id)
		convey.So(newNet.Id, convey.ShouldNotEqual, "")
		newNetworkID = newNet.Id
	})

	convey.Convey("TestCreateSubnet---ErrNetworkID\n", t, func() {
		newNet, err := m.CreateSubnet(newNetworkName,
			"192.168.1.1/24", "192.168.1.1", []subnets.AllocationPool{})
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(newNet, convey.ShouldBeNil)
	})

	convey.Convey("TestCreateSubnet---ErrNetworkID\n", t, func() {
		newNet, err := m.CreateSubnet(newNetworkName,
			"192..1/24", "192.168.1.1", []subnets.AllocationPool{})
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(newNet, convey.ShouldBeNil)
	})

	stubsSaveData = StubFunc(&SaveData, errors.New("SAVE-DATA-ERROR"))
	convey.Convey("TestCreateSubnet---ErrNetworkID\n", t, func() {
		newNet, err := m.CreateSubnet(newNetworkID,
			"192.168.1.1/24", "192.168.1.1", []subnets.AllocationPool{})
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(newNet, convey.ShouldBeNil)
	})
}

func TestDeleteSubnetERR(t *testing.T) {
	stubsSaveData := StubFunc(&SaveData, nil)
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	stubsDeleteData := StubFunc(&DeleteData, nil)
	defer stubsSaveData.Reset()
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	defer stubsDeleteData.Reset()
	var newNetworkName string = "Create-Network-For-TestDeleteSubnetERR"
	var newNetworkID, newSubnetID string
	m := GetEmbeddedNetwrokManager()

	convey.Convey("TestDeleteSubnet---ErrNetworkID\n", t, func() {
		err := m.DeleteSubnet(newNetworkName)
		convey.So(err, convey.ShouldNotEqual, nil)
	})

	convey.Convey("TestCreateNetwork---OK\n", t, func() {
		newNet, err := m.CreateNetwork(newNetworkName)
		convey.So(err, convey.ShouldEqual, nil)
		fmt.Print("\nNew-Network-ID:" + newNet.Id)
		convey.So(newNet.Id, convey.ShouldNotEqual, "")
		newNetworkID = newNet.Id
	})

	convey.Convey("TestCreateSubnet---ErrNetworkID\n", t, func() {
		subNet, err := m.CreateSubnet(newNetworkID,
			"192.168.1.1/24", "192.168.1.1", []subnets.AllocationPool{})
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(subNet.Id, convey.ShouldNotEqual, "")
		newSubnetID = subNet.Id
	})

	stubsDeleteData = StubFunc(&DeleteData, errors.New("DEL-ERROR"))
	convey.Convey("TestDeleteSubnet---ErrNetworkID\n", t, func() {
		err := m.DeleteSubnet(newSubnetID)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestGetSubnetERR(t *testing.T) {
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	var newNetworkName string = "Create-Network-For-TestDeleteSubnetERR"
	m := GetEmbeddedNetwrokManager()

	convey.Convey("TestGetSubnetID---ErrNetworkID\n", t, func() {
		id, err := m.GetSubnetID(newNetworkName)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(id, convey.ShouldEqual, "")
	})

	convey.Convey("TestGetSubnet---ErrNetworkID\n", t, func() {
		id, err := m.GetSubnet(newNetworkName)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(id, convey.ShouldEqual, nil)
	})
}
