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

func TestNetworkOK(t *testing.T) {
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

func TestCreateNetworkErr(t *testing.T) {
	stubsSaveData := StubFunc(&SaveData, errors.New("SAVE-DATA-ERR"))
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	stubsDeleteData := StubFunc(&DeleteData, nil)
	defer stubsSaveData.Reset()
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	defer stubsDeleteData.Reset()
	var newNetworkName string = "Create-Network-For-TEST"
	m := GetEmbeddedNetwrokManager()
	convey.Convey("TestCreateNetwork---ERR\n", t, func() {
		newNet, err := m.CreateNetwork(newNetworkName)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(newNet, convey.ShouldBeNil)
	})
}

func TestDeleteNetworkErrNotFound(t *testing.T) {
	stubsSaveData := StubFunc(&SaveData, nil)
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	stubsDeleteData := StubFunc(&DeleteData, nil)
	defer stubsSaveData.Reset()
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	defer stubsDeleteData.Reset()
	var newNetworkName string = "Create-Network-For-TEST"
	m := GetEmbeddedNetwrokManager()
	convey.Convey("TestDeleteNetwork---ERR\n", t, func() {
		err := m.DeleteNetwork(newNetworkName)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestDeleteNetworkErrDelNet(t *testing.T) {
	stubsSaveData := StubFunc(&SaveData, nil)
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	stubsDeleteData := StubFunc(&DeleteData, errors.New("DEL-DATA-ERR"))
	defer stubsSaveData.Reset()
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	defer stubsDeleteData.Reset()
	var newNetworkName string = "TestDeleteNetworkErrDelNet"
	var newNetworkID string
	m := GetEmbeddedNetwrokManager()
	convey.Convey("TestCreateNetwork---OK\n", t, func() {
		netNet, err := m.CreateNetwork(newNetworkName)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(netNet.Id, convey.ShouldNotEqual, "")
		newNetworkID = netNet.Id
	})

	convey.Convey("TestDeleteNetwork---ERR\n", t, func() {
		err := m.DeleteNetwork(newNetworkID)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	/*
	*This is a bug, when delete vxlan id of network ok, but delete
	*netwrok error.
	 */
	convey.Convey("TestCreateNetwork---ERR\n", t, func() {
		id, err := m.CreateNetwork(newNetworkName)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(id, convey.ShouldNotEqual, "")
	})
}

func TestDeleteNetworkWithSubnetErrFreeVxlanID(t *testing.T) {
	stubsSaveData := StubFunc(&SaveData, nil)
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	stubsDeleteData := StubFunc(&DeleteData, nil)
	defer stubsSaveData.Reset()
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	defer stubsDeleteData.Reset()
	var newNetworkName string = "TestDeleteNetworkWithSubnetErrFreeVxlanID"
	var newNetworkID string
	m := GetEmbeddedNetwrokManager()
	convey.Convey("TestCreateNetwork---OK\n", t, func() {
		newNet, err := m.CreateNetwork(newNetworkName)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(newNet.Id, convey.ShouldNotEqual, "")
		newNetworkID = newNet.Id
	})

	convey.Convey("TestCreateSubnet---OK\n", t, func() {
		id, err := m.CreateSubnet(newNetworkID,
			"192.168.1.1/24", "192.168.1.1", []subnets.AllocationPool{})
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(id, convey.ShouldNotEqual, "")
	})

	stubsSaveData = StubFunc(&SaveData, errors.New("SAVE-DATA-ERR"))
	convey.Convey("TestDeleteNetwork---ERR\n", t, func() {
		err := m.DeleteNetwork(newNetworkID)
		convey.So(err, convey.ShouldNotEqual, nil)
	})

}

func TestDeleteNetworkErrDelSub(t *testing.T) {
	stubsSaveData := StubFunc(&SaveData, nil)
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	stubsDeleteData := StubFunc(&DeleteData, nil)
	defer stubsSaveData.Reset()
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	defer stubsDeleteData.Reset()
	var newNetworkName string = "TestDeleteNetworkErrDelSub"
	var newNetworkID string
	m := GetEmbeddedNetwrokManager()
	convey.Convey("TestCreateNetwork---OK\n", t, func() {
		newNet, err := m.CreateNetwork(newNetworkName)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(newNet.Id, convey.ShouldNotEqual, "")
		newNetworkID = newNet.Id
	})

	convey.Convey("TestCreateSubnet---OK\n", t, func() {
		id, err := m.CreateSubnet(newNetworkID,
			"192.168.1.1/24", "192.168.1.1", []subnets.AllocationPool{})
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(id, convey.ShouldNotEqual, "")
	})

	stubsDeleteData = StubFunc(&DeleteData, errors.New("DELETE-DATA-ERR"))
	convey.Convey("TestDeleteNetwork---ERR\n", t, func() {
		err := m.DeleteNetwork(newNetworkID)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestGetNetworkIDErrNotFound(t *testing.T) {
	stubsSaveData := StubFunc(&SaveData, nil)
	stubsReadDir := StubFunc(&ReadDataDir, nil, errors.New("NO-DATA"))
	stubsReadData := StubFunc(&ReadData, "", errors.New("NO-DATA"))
	stubsDeleteData := StubFunc(&DeleteData, nil)
	defer stubsSaveData.Reset()
	defer stubsReadDir.Reset()
	defer stubsReadData.Reset()
	defer stubsDeleteData.Reset()
	var newNetworkName string = "Create-Network-For-TEST"
	m := GetEmbeddedNetwrokManager()
	convey.Convey("TestCreateNetwork---OK\n", t, func() {
		id, err := m.CreateNetwork(newNetworkName)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(id, convey.ShouldNotEqual, "")
	})

	convey.Convey("TestGetNetworkID---ERR\n", t, func() {
		id, err := m.GetNetworkID(newNetworkName + "ERR")
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(id, convey.ShouldEqual, "")
	})

	convey.Convey("TestGetNetwork---ERR\n", t, func() {
		id, err := m.GetNetwork(newNetworkName)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(id, convey.ShouldEqual, nil)
	})

	convey.Convey("TestGetNetwork---ERR\n", t, func() {
		id, err := m.GetNetworkExtenAttrs(newNetworkName)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(id, convey.ShouldEqual, nil)
	})
}
