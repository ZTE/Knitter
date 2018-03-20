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

	"encoding/json"
	_ "fmt"
	"github.com/bouk/monkey"
)

func TestSaveSubnet(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	mockDB.EXPECT().SaveLeaf("/knitter/manager/subnets/subnet-id", gomock.Any()).Return(nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestSaveSubnet", t, func() {
		err := SaveSubnet(&Subnet{ID: "subnet-id"})
		So(err, ShouldBeNil)
	})
}

func TestSaveSubnet_DBFailed(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	errStr := "etcd cluster misconfig"
	mockDB.EXPECT().SaveLeaf("/knitter/manager/subnets/subnet-id", gomock.Any()).Return(errors.New(errStr))
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestSaveSubnet_DBFailed", t, func() {
		err := SaveSubnet(&Subnet{ID: "subnet-id"})
		So(err.Error(), ShouldEqual, errStr)
	})
}

func TestDeleteSubnet(t *testing.T) {
	subnetID := "subnet-id"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	mockDB.EXPECT().DeleteLeaf("/knitter/manager/subnets/subnet-id").Return(nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestDeleteSubnet", t, func() {
		err := DelSubnet(subnetID)
		So(err, ShouldBeNil)
	})
}

func TestDeleteSubnet_DBFailed(t *testing.T) {
	subnetID := "subnet-id"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	errStr := "etcd cluster misconfig"
	mockDB.EXPECT().DeleteLeaf("/knitter/manager/subnets/subnet-id").Return(errors.New(errStr))
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestDeleteSubnet_DBFailed", t, func() {
		err := DelSubnet(subnetID)
		So(err.Error(), ShouldEqual, errStr)
	})
}

func TestGetSubnet(t *testing.T) {
	subnetID := "subnet-id"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	subnetStr := `{"ID":"subnet-id"}`
	mockDB.EXPECT().ReadLeaf("/knitter/manager/subnets/subnet-id").Return(subnetStr, nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestGetSubnet", t, func() {
		net, err := GetSubnet(subnetID)
		So(err, ShouldBeNil)
		So(net.ID, ShouldEqual, subnetID)
	})
}

func TestGetSubnet_DBFailed(t *testing.T) {
	subnetID := "subnet-id"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	errStr := "etcd cluster misconfig"
	mockDB.EXPECT().ReadLeaf("/knitter/manager/subnets/subnet-id").Return("", errors.New(errStr))
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestGetSubnet_DBFailed", t, func() {
		net, err := GetSubnet(subnetID)
		So(err.Error(), ShouldEqual, errStr)
		So(net, ShouldBeNil)
	})
}

func TestGetSubnet_UnmarshalFailed(t *testing.T) {
	subnetID := "subnet-id"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	subnetStr := `{"ID":"subnet-id"}`
	mockDB.EXPECT().ReadLeaf("/knitter/manager/subnets/subnet-id").Return(subnetStr, nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	guard := monkey.Patch(json.Unmarshal, func(_ []byte, _ interface{}) error {
		return errors.New("unexpected content")
	})
	defer guard.Unpatch()

	Convey("TestGetSubnet_UnmarshalFailed", t, func() {
		net, err := GetSubnet(subnetID)
		So(err, ShouldEqual, errobj.ErrUnmarshalFailed)
		So(net, ShouldBeNil)
	})
}

func TestUnmarshalSubnet(t *testing.T) {
	Convey("TestUnmarshalSubnet", t, func() {
		net, err := UnmarshalSubnet([]byte(`{"ID":"subnet-id1"}`))
		So(err, ShouldBeNil)
		So(net.ID, ShouldEqual, "subnet-id1")
	})
}

func TestUnmarshalSubnet_Fail(t *testing.T) {
	errObj := errors.New("invalid content")
	guard := monkey.Patch(json.Unmarshal, func(_ []byte, _ interface{}) error {
		return errObj
	})
	defer guard.Unpatch()

	Convey("TestUnmarshalSubnet_Fail", t, func() {
		physPort, err := UnmarshalSubnet([]byte(`{"ID":"subnet-id1"}`))
		So(err, ShouldEqual, errobj.ErrUnmarshalFailed)
		So(physPort, ShouldBeNil)
	})
}

func TestGetAllSubnets(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)

	nodes := []*client.Node{
		{
			Key:   "/knitter/manager/subnets/subnet-id1",
			Value: `{"ID":"subnet-id1"}`,
		},
		{
			Key:   "/knitter/manager/subnets/subnet-id2",
			Value: `{"ID":"subnet-id2"}`,
		},
		{
			Key:   "/knitter/manager/subnets/subnet-id3",
			Value: `{"ID":"subnet-id3"}`,
		},
	}
	mockDB.EXPECT().ReadDir("/knitter/manager/subnets").Return(nodes, nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)

	outputs := []gostub.Output{
		{StubVals: gostub.Values{&Subnet{ID: "subnet-id1"}, nil}},
		{StubVals: gostub.Values{&Subnet{ID: "subnet-id2"}, nil}},
		{StubVals: gostub.Values{&Subnet{ID: "subnet-id3"}, nil}},
	}
	stub.StubFuncSeq(&UnmarshalSubnet, outputs)
	defer stub.Reset()

	Convey("TestGetAllSubnets", t, func() {
		subnets, err := GetAllSubnets()
		So(err, ShouldBeNil)
		So(len(subnets), ShouldEqual, 3)
		So(subnets[0].ID, ShouldEqual, "subnet-id1")
		So(subnets[1].ID, ShouldEqual, "subnet-id2")
		So(subnets[2].ID, ShouldEqual, "subnet-id3")
	})
}

func TestGetAllSubnets_DBFailed(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)

	errStr := "etcd cluster misconfig"
	mockDB.EXPECT().ReadDir("/knitter/manager/subnets").Return(nil, errors.New(errStr))

	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestGetAllSubnets_DBFailed", t, func() {
		networks, err := GetAllSubnets()
		So(err.Error(), ShouldEqual, errStr)
		So(networks, ShouldBeNil)
	})
}

func TestGetAllSubnets_UnmarshalFail(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)

	nodes := []*client.Node{
		{
			Key:   "/knitter/manager/subnets/subnet-id1",
			Value: `{"ID":"subnet-id1"}`,
		},
		{
			Key:   "/knitter/manager/subnets/subnet-id2",
			Value: `{"ID":"subnet-id2"}`,
		},
		{
			Key:   "/knitter/manager/subnets/subnet-id3",
			Value: `{"ID":"subnet-id3"}`,
		},
	}
	mockDB.EXPECT().ReadDir("/knitter/manager/subnets").Return(nodes, nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)

	errStr := "etcd cluster misconfig"
	outputs := []gostub.Output{
		{StubVals: gostub.Values{&Subnet{ID: "subnet-id1"}, nil}},
		{StubVals: gostub.Values{&Subnet{ID: "subnet-id2"}, nil}},
		{StubVals: gostub.Values{nil, errors.New(errStr)}},
	}
	stub.StubFuncSeq(&UnmarshalSubnet, outputs)
	defer stub.Reset()

	Convey("TestGetAllSubnets_UnmarshalFail", t, func() {
		subnets, err := GetAllSubnets()
		So(err.Error(), ShouldEqual, errStr)
		So(subnets, ShouldBeNil)
	})
}
