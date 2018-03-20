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
	"errors"
	"github.com/coreos/etcd/client"
	"github.com/golang/gostub"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/knitter-manager/tests/mock/db-mock"

	//"fmt"
	"encoding/json"
	"github.com/bouk/monkey"
)

//func TestNetworkObjRepoOps(t *testing.T) {
//	net1ID := "net1-id"
//	net2ID := "net2-id"
//	net3ID := "net3-id"
//	net4ID := "net4-id"
//
//	net1Name := "net1-name"
//	net2Name := "net2-name"
//	net3Name := "net3-name"
//	net4Name := "net4-name"
//
//	tenantNwID := "tenant-nw-id"
//	tenantOpsID := "tenant-Ops-id"
//
//	netObj1 := NetworkObject{
//		ID:         net1ID,
//		Name:       net1Name,
//		TenantID:   tenantNwID,
//		IsPublic:   true,
//		IsExternal: false,
//	}
//	netObj2 := NetworkObject{
//		ID:         net2ID,
//		Name:       net2Name,
//		TenantID:   tenantNwID,
//		IsPublic:   false,
//		IsExternal: true,
//	}
//	netObj3 := NetworkObject{
//		ID:         net3ID,
//		Name:       net3Name,
//		TenantID:   tenantOpsID,
//		IsPublic:   true,
//		IsExternal: false,
//	}
//	netObj4 := NetworkObject{
//		ID:         net4ID,
//		Name:       net4Name,
//		TenantID:   tenantOpsID,
//		IsPublic:   false,
//		IsExternal: true,
//	}
//
//	Convey("TestNetworkObjRepoOps:", t, func() {
//		err := GetNetObjRepoSingleton().Add(&netObj1)
//		So(err, ShouldBeNil)
//		err = GetNetObjRepoSingleton().Add(&netObj2)
//		So(err, ShouldBeNil)
//		err = GetNetObjRepoSingleton().Add(&netObj3)
//		So(err, ShouldBeNil)
//		err = GetNetObjRepoSingleton().Add(&netObj4)
//		So(err, ShouldBeNil)
//
//		obj1, err := GetNetObjRepoSingleton().Get(netObj1.ID)
//		So(err, ShouldBeNil)
//		So(obj1, ShouldPointTo, &netObj1)
//
//		obj2, err := GetNetObjRepoSingleton().Get(netObj2.ID)
//		So(err, ShouldBeNil)
//		So(obj2, ShouldPointTo, &netObj2)
//
//		obj3, err := GetNetObjRepoSingleton().Get(netObj3.ID)
//		So(err, ShouldBeNil)
//		So(obj3, ShouldPointTo, &netObj3)
//
//		obj4, err := GetNetObjRepoSingleton().Get(netObj4.ID)
//		So(err, ShouldBeNil)
//		So(obj4, ShouldPointTo, &netObj4)
//
//		obj5, err := GetNetObjRepoSingleton().Get("invalid-network-id")
//		So(err, ShouldEqual, errobj.ErrRecordNotExist)
//		So(obj5, ShouldBeNil)
//
//		// test list by name
//		netObj, err := GetNetObjRepoSingleton().ListByNetworkName(net1Name)
//		So(err, ShouldBeNil)
//		So(netObj[0], ShouldEqual, &netObj1)
//		netObj, err = GetNetObjRepoSingleton().ListByNetworkName(net2Name)
//		So(err, ShouldBeNil)
//		So(netObj[0], ShouldEqual, &netObj2)
//		netObj, err = GetNetObjRepoSingleton().ListByNetworkName(net3Name)
//		So(err, ShouldBeNil)
//		So(netObj[0], ShouldEqual, &netObj3)
//		netObj, err = GetNetObjRepoSingleton().ListByNetworkName(net4Name)
//		So(err, ShouldBeNil)
//		So(netObj[0], ShouldEqual, &netObj4)
//
//		// test list by tenant id
//		netObjs, err := GetNetObjRepoSingleton().ListByTenantID(tenantNwID)
//		So(err, ShouldBeNil)
//		So(len(netObjs), ShouldEqual, 2)
//		So(netObjs[0] == &netObj1 && netObjs[1] == &netObj2 || netObjs[0] == &netObj2 && netObjs[1] == &netObj1, ShouldBeTrue)
//		netObjs, err = GetNetObjRepoSingleton().ListByTenantID(tenantOpsID)
//		So(err, ShouldBeNil)
//		So(len(netObjs), ShouldEqual, 2)
//		So(netObjs[0] == &netObj3 && netObjs[1] == &netObj4 || netObjs[0] == &netObj4 && netObjs[1] == &netObj3, ShouldBeTrue)
//
//
//		// test list by public label
//		tmpObjs, err := GetNetObjRepoSingleton().ListByIsPublic("true")
//		So(err, ShouldBeNil)
//		//So(len(tmpObjs), ShouldEqual, 2) // todo: must repair other pre-testcases(add teardown networks ops)
//		fmt.Printf("")
//		So(tmpObjs[0] == &netObj1 && tmpObjs[1] == &netObj3 || tmpObjs[0] == &netObj3 && tmpObjs[1] == &netObj1, ShouldBeTrue)
//		tmpObjs, err = GetNetObjRepoSingleton().ListByIsPublic("false")
//		So(err, ShouldBeNil)
//		So(len(tmpObjs), ShouldEqual, 2)
//		So(tmpObjs[0] == &netObj2 && tmpObjs[1] == &netObj4 || tmpObjs[0] == &netObj4 && tmpObjs[1] == &netObj2, ShouldBeTrue)
//
//		// test list by external label
//		tmpObjs, err = GetNetObjRepoSingleton().ListByIsExternal("true")
//		So(err, ShouldBeNil)
//		So(len(tmpObjs), ShouldEqual, 2)
//		So(tmpObjs[0] == &netObj2 && tmpObjs[1] == &netObj4 || tmpObjs[0] == &netObj4 && tmpObjs[1] == &netObj2, ShouldBeTrue)
//		tmpObjs, err = GetNetObjRepoSingleton().ListByIsExternal("false")
//		So(err, ShouldBeNil)
//		So(len(tmpObjs), ShouldEqual, 2)
//		So(tmpObjs[0] == &netObj1 && tmpObjs[1] == &netObj3 || tmpObjs[0] == &netObj3 && tmpObjs[1] == &netObj1, ShouldBeTrue)
//
//		// delete
//		err = GetNetObjRepoSingleton().Del(netObj1.ID)
//		So(err, ShouldBeNil)
//
//		modNetObj2 := *obj2
//		modNetObj2.Name = "port2-new-name"
//		err = GetNetObjRepoSingleton().Update(&modNetObj2)
//		So(err, ShouldBeNil)
//		newObj2, err := GetNetObjRepoSingleton().Get(obj2.ID)
//		So(err, ShouldBeNil)
//		So(newObj2, ShouldPointTo, &modNetObj2)
//	})
//
//}

func TestSaveNetwork(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	mockDB.EXPECT().SaveLeaf("/knitter/manager/networks/network-id", gomock.Any()).Return(nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestSaveNetwork", t, func() {
		err := SaveNetwork(&Network{ID: "network-id"})
		So(err, ShouldBeNil)
	})
}

func TestSaveNetwork_DBFailed(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	errStr := "etcd cluster misconfig"
	mockDB.EXPECT().SaveLeaf("/knitter/manager/networks/network-id", gomock.Any()).Return(errors.New(errStr))
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestSaveNetwork_DBFailed", t, func() {
		err := SaveNetwork(&Network{ID: "network-id"})
		So(err.Error(), ShouldEqual, errStr)
	})
}

func TestDeleteNetwork(t *testing.T) {
	netID := "network-id"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	mockDB.EXPECT().DeleteLeaf("/knitter/manager/networks/network-id").Return(nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestDeleteNetwork", t, func() {
		err := DelNetwork(netID)
		So(err, ShouldBeNil)
	})
}

func TestDeleteNetwork_DBFailed(t *testing.T) {
	netID := "network-id"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	errStr := "etcd cluster misconfig"
	mockDB.EXPECT().DeleteLeaf("/knitter/manager/networks/network-id").Return(errors.New(errStr))
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestDeleteNetwork_DBFailed", t, func() {
		err := DelNetwork(netID)
		So(err.Error(), ShouldEqual, errStr)
	})
}
func TestGetNetwork(t *testing.T) {
	netID := "network-id"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	netStr := `{"ID":"network-id"}`
	mockDB.EXPECT().ReadLeaf("/knitter/manager/networks/network-id").Return(netStr, nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestGetNetwork", t, func() {
		net, err := GetNetwork2(netID)
		So(err, ShouldBeNil)
		So(net.ID, ShouldEqual, netID)
	})
}

func TestGetNetwork_DBFailed(t *testing.T) {
	netID := "network-id"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	errStr := "etcd cluster misconfig"
	mockDB.EXPECT().ReadLeaf("/knitter/manager/networks/network-id").Return("", errors.New(errStr))
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestGetNetwork_DBFailed", t, func() {
		net, err := GetNetwork2(netID)
		So(err.Error(), ShouldEqual, errStr)
		So(net, ShouldBeNil)
	})
}

func TestGetNetwork_UnmarshalFailed(t *testing.T) {
	netID := "network-id"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	netStr := `{"ID":"network-id"}`
	mockDB.EXPECT().ReadLeaf("/knitter/manager/networks/network-id").Return(netStr, nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	guard := monkey.Patch(json.Unmarshal, func(_ []byte, _ interface{}) error {
		return errors.New("unexpected content")
	})
	defer guard.Unpatch()

	Convey("TestGetNetwork_UnmarshalFailed", t, func() {
		net, err := GetNetwork2(netID)
		So(err, ShouldEqual, errobj.ErrUnmarshalFailed)
		So(net, ShouldBeNil)
	})
}

func TestUnmarshalNetwork(t *testing.T) {
	Convey("TestUnmarshalNetwork", t, func() {
		net, err := UnmarshalNetwork([]byte(`{"ID":"network-id1"}`))
		So(err, ShouldBeNil)
		So(net.ID, ShouldEqual, "network-id1")
	})
}

func TestUnmarshalNetwork_Fail(t *testing.T) {
	errObj := errors.New("invalid content")
	guard := monkey.Patch(json.Unmarshal, func(_ []byte, _ interface{}) error {
		return errObj
	})
	defer guard.Unpatch()

	Convey("TestUnmarshalNetwork_Fail", t, func() {
		physPort, err := UnmarshalNetwork([]byte(`{"ID":"network-id1"}`))
		So(err, ShouldEqual, errobj.ErrUnmarshalFailed)
		So(physPort, ShouldBeNil)
	})
}

func TestGetAllNetworks(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)

	nodes := []*client.Node{
		{
			Key:   "/knitter/manager/networks/network-id1",
			Value: `{"ID":"network-id1"}`,
		},
		{
			Key:   "/knitter/manager/networks/network-id2",
			Value: `{"ID":"network-id2"}`,
		},
		{
			Key:   "/knitter/manager/networks/network-id3",
			Value: `{"ID":"network-id3"}`,
		},
	}
	mockDB.EXPECT().ReadDir("/knitter/manager/networks").Return(nodes, nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)

	outputs := []gostub.Output{
		{StubVals: gostub.Values{&Network{ID: "network-id1"}, nil}},
		{StubVals: gostub.Values{&Network{ID: "network-id2"}, nil}},
		{StubVals: gostub.Values{&Network{ID: "network-id3"}, nil}},
	}
	stub.StubFuncSeq(&UnmarshalNetwork, outputs)
	defer stub.Reset()

	Convey("TestGetAllNetworks", t, func() {
		networks, err := GetAllNetworks()
		So(err, ShouldBeNil)
		So(len(networks), ShouldEqual, 3)
		So(networks[0].ID, ShouldEqual, "network-id1")
		So(networks[1].ID, ShouldEqual, "network-id2")
		So(networks[2].ID, ShouldEqual, "network-id3")
	})
}

func TestGetAllNetworks_DBFailed(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)

	errStr := "etcd cluster misconfig"
	mockDB.EXPECT().ReadDir("/knitter/manager/networks").Return(nil, errors.New(errStr))

	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestGetAllNetworks_DBFailed", t, func() {
		networks, err := GetAllNetworks()
		So(err.Error(), ShouldEqual, errStr)
		So(networks, ShouldBeNil)
	})
}

func TestGetAllNetworks_UnmarshalFail(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)

	nodes := []*client.Node{
		{
			Key:   "/knitter/manager/networks/network-id1",
			Value: `{"ID":"network-id1"}`,
		},
		{
			Key:   "/knitter/manager/networks/network-id2",
			Value: `{"ID":"network-id2"}`,
		},
		{
			Key:   "/knitter/manager/networks/network-id3",
			Value: `{"ID":"network-id3"}`,
		},
	}
	mockDB.EXPECT().ReadDir("/knitter/manager/networks").Return(nodes, nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)

	errStr := "etcd cluster misconfig"
	outputs := []gostub.Output{
		{StubVals: gostub.Values{&Network{ID: "network-id1"}, nil}},
		{StubVals: gostub.Values{&Network{ID: "network-id2"}, nil}},
		{StubVals: gostub.Values{nil, errors.New(errStr)}},
	}
	stub.StubFuncSeq(&UnmarshalNetwork, outputs)
	defer stub.Reset()

	Convey("TestGetAllNetworks_UnmarshalFail", t, func() {
		networks, err := GetAllNetworks()
		So(err.Error(), ShouldEqual, errStr)
		So(networks, ShouldBeNil)
	})
}
