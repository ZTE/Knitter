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

package common

import (
	"errors"
	"github.com/ZTE/Knitter/knitter-manager/tests/mock/db-mock"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/uuid"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCase1GetPaasUUID(t *testing.T) {
	var wantUUID string = "UUID-PAAS-KNITTER"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	SetDataBase(mockDB)

	gomock.InOrder(
		mockDB.EXPECT().ReadLeaf(dbaccessor.GetKeyOfPaasUUID()).Return(wantUUID, nil),
	)

	convey.Convey("TestCase1GetPaasUUID\n", t, func() {
		id := GetPaasUUID()
		convey.So(id, convey.ShouldEqual, wantUUID)
	})
}

func TestCase2GetPaasUUID(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	SetDataBase(mockDB)

	gomock.InOrder(
		mockDB.EXPECT().ReadLeaf(dbaccessor.GetKeyOfPaasUUID()).Return("",
			errors.New("unknow errors")),
	)

	convey.Convey("TestCase2GetPaasUUID\n", t, func() {
		id := GetPaasUUID()
		convey.So(id, convey.ShouldEqual, uuid.NIL.String())
	})
}

func TestCase3GetPaasUUID(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	SetDataBase(mockDB)

	gomock.InOrder(
		mockDB.EXPECT().ReadLeaf(dbaccessor.GetKeyOfPaasUUID()).Return("",
			errors.New(ErrorKeyNotFound)),
		mockDB.EXPECT().SaveLeaf(dbaccessor.GetKeyOfPaasUUID(),
			gomock.Any()).Return(nil),
	)

	convey.Convey("TestCase3GetPaasUUID\n", t, func() {
		id := GetPaasUUID()
		convey.So(id, convey.ShouldNotEqual, uuid.NIL.String())
	})
}

func TestCase4GetPaasUUID(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	SetDataBase(mockDB)

	gomock.InOrder(
		mockDB.EXPECT().ReadLeaf(dbaccessor.GetKeyOfPaasUUID()).Return("",
			errors.New(ErrorKeyNotFound)),
		mockDB.EXPECT().SaveLeaf(dbaccessor.GetKeyOfPaasUUID(),
			gomock.Any()).Return(errors.New("unknow error happen")),
	)

	convey.Convey("TestCase4GetPaasUUID\n", t, func() {
		id := GetPaasUUID()
		convey.So(id, convey.ShouldEqual, uuid.NIL.String())
	})
}

func TestCase1GetOpenstackCfg(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	SetDataBase(mockDB)

	gomock.InOrder(
		mockDB.EXPECT().ReadLeaf(dbaccessor.GetKeyOfOpenstack()).Return("",
			errors.New(ErrorKeyNotFound)),
	)

	convey.Convey("TestCase1GetOpenstackCfg\n", t, func() {
		cfg := GetOpenstackCfg()
		convey.So(cfg, convey.ShouldEqual, "")
	})
}

func TestCase2GetOpenstackCfg(t *testing.T) {
	const OpenstackConfig string = "right openstack configuration"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	SetDataBase(mockDB)

	gomock.InOrder(
		mockDB.EXPECT().ReadLeaf(dbaccessor.GetKeyOfOpenstack()).Return(
			OpenstackConfig, nil),
	)

	convey.Convey("TestCase2GetOpenstackCfg\n", t, func() {
		cfg := GetOpenstackCfg()
		convey.So(cfg, convey.ShouldEqual, OpenstackConfig)
	})
}

func TestCase1GetVnfmCfg(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	SetDataBase(mockDB)

	gomock.InOrder(
		mockDB.EXPECT().ReadLeaf(dbaccessor.GetKeyOfVnfm()).Return(
			"", errors.New(ErrorKeyNotFound)),
	)

	convey.Convey("TestCase1GetVnfmCfg\n", t, func() {
		cfg := GetVnfmCfg()
		convey.So(cfg, convey.ShouldEqual, "")
	})
}

func TestCase2GetVnfmCfg(t *testing.T) {
	const OpenstackConfig string = "right openstack configuration"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	SetDataBase(mockDB)

	gomock.InOrder(
		mockDB.EXPECT().ReadLeaf(dbaccessor.GetKeyOfVnfm()).Return(
			OpenstackConfig, nil),
	)

	convey.Convey("TestCase2GetVnfmCfg\n", t, func() {
		cfg := GetVnfmCfg()
		convey.So(cfg, convey.ShouldEqual, OpenstackConfig)
	})
}

func TestCase1RegisterSelfToDb(t *testing.T) {
	const ManagerURL string = "https://paas.zte.com.cn:9527"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	SetDataBase(mockDB)

	gomock.InOrder(
		mockDB.EXPECT().SaveLeaf(dbaccessor.GetKeyOfKnitterManagerUrl(),
			ManagerURL).Return(nil),
	)

	convey.Convey("TestCase1RegisterSelfToDb\n", t, func() {
		e := RegisterSelfToDb(ManagerURL)
		convey.So(e, convey.ShouldEqual, nil)
	})
}

func TestCase2RegisterSelfToDb(t *testing.T) {
	const ManagerURL string = "https://paas.zte.com.cn:9527"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	SetDataBase(mockDB)

	gomock.InOrder(
		mockDB.EXPECT().SaveLeaf(dbaccessor.GetKeyOfKnitterManagerUrl(),
			ManagerURL).Return(errors.New(ErrorKeyNotFound)),
	)

	convey.Convey("TestCase1RegisterSelfToDb\n", t, func() {
		e := RegisterSelfToDb(ManagerURL)
		convey.So(e, convey.ShouldNotEqual, nil)
	})
}

func TestCase1SetDataBase(t *testing.T) {
	convey.Convey("TestCase1SetDataBase\n", t, func() {
		e := SetDataBase(nil)
		convey.So(e, convey.ShouldEqual, nil)
	})
}

func TestCase1GetDataBase(t *testing.T) {
	convey.Convey("TestCase1GetDataBase\n", t, func() {
		e := SetDataBase(nil)
		convey.So(e, convey.ShouldEqual, nil)
		p := GetDataBase()
		convey.So(p, convey.ShouldEqual, nil)
	})
}

//todo fix ut error
/*func TestCase1CheckDB(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	SetDataBase(mockDB)

	gomock.InOrder(
		//case 1
		mockDB.EXPECT().SaveLeaf(gomock.Any(),
			gomock.Any()).Return(errors.New(ErrorKeyNotFound)),
		//case 2
		mockDB.EXPECT().SaveLeaf(gomock.Any(),
			gomock.Any()).Return(nil),
		mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
			"", errors.New(ErrorKeyNotFound)),
		//case 3
		mockDB.EXPECT().SaveLeaf(gomock.Any(),
			gomock.Any()).Return(nil),
		mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
			"Error-DATA", nil),
		//case 4
		mockDB.EXPECT().SaveLeaf(gomock.Any(),
			gomock.Any()).Return(nil),
		mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
			"DataBase-Config-is-OK", nil),
		mockDB.EXPECT().DeleteLeaf(gomock.Any()).Return(
			errors.New(ErrorKeyNotFound)),
		//case 5
		mockDB.EXPECT().SaveLeaf(gomock.Any(),
			gomock.Any()).Return(nil),
		mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
			"DataBase-Config-is-OK", nil),
		mockDB.EXPECT().DeleteLeaf(gomock.Any()).Return(nil),
		mockDB.EXPECT().DeleteDir(gomock.Any()).Return(
			errors.New(ErrorKeyNotFound)),
		//case 6
		mockDB.EXPECT().SaveLeaf(gomock.Any(),
			gomock.Any()).Return(nil),
		mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
			"DataBase-Config-is-OK", nil),
		mockDB.EXPECT().DeleteLeaf(gomock.Any()).Return(nil),
		mockDB.EXPECT().DeleteDir(gomock.Any()).Return(nil),
	)

	convey.Convey("TestCase1CheckDB\n", t, func() {
		e := CheckDB()
		convey.So(e, convey.ShouldEqual, nil)
	})
}*/
