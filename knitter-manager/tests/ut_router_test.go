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

package test

import (
	"errors"
	"testing"

	"github.com/coreos/etcd/client"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/public"
	_ "github.com/ZTE/Knitter/knitter-manager/routers"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/golang/gostub"
)

func TestCreateRouterErrNotCfgIaaS(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, nil)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	cfg := string(`{"router":{"name":"test-create-router"}}`)
	resp := CreateRouter(cfg)
	Convey("TestCreateRouterErrNotCfgIaaS\n", t, func() {
		So(resp.Code, ShouldEqual, 401)
	})
}

func TestCreateRouterErr403(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	cfg := string(`{"router":{"name":"test-create-router"}`)
	resp := CreateRouter(cfg)
	Convey("TestCreateNetworkErrNotCfgIaaS\n", t, func() {
		So(resp.Code, ShouldEqual, 403)
	})
}

func TestCreateRouterErr406A(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockIaaS.EXPECT().CreateRouter(gomock.Any(),
		gomock.Any()).Return("", errors.New("create-router-error"))

	cfg := string(`{"router":{"name":"test-create-router"}}`)
	resp := CreateRouter(cfg)
	Convey("TestCreateRouterErr406A\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestCreateRouterErr406B(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockIaaS.EXPECT().CreateRouter(gomock.Any(),
		gomock.Any()).Return("router-uuid", nil)
	mockIaaS.EXPECT().GetRouter(gomock.Any()).Return(
		nil, errors.New("get-router-error"))

	cfg := string(`{"router":{"name":"test-create-router"}}`)
	resp := CreateRouter(cfg)
	Convey("TestCreateRouterErr406B\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestCreateRouterErr406C(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockIaaS.EXPECT().CreateRouter(gomock.Any(),
		gomock.Any()).Return("router-uuid", nil)
	router := iaasaccessor.Router{Id: "router-uuid"}
	mockIaaS.EXPECT().GetRouter(gomock.Any()).Return(
		&router, nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),
		gomock.Any()).Return(errors.New("error"))
	mockIaaS.EXPECT().DeleteRouter(gomock.Any()).Return(nil)

	cfg := string(`{"router":{"name":"test-create-router"}}`)
	resp := CreateRouter(cfg)
	Convey("TestCreateRouterErr406C\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestCreateRouterOK(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockIaaS.EXPECT().CreateRouter(gomock.Any(),
		gomock.Any()).Return("router-uuid", nil)
	router := iaasaccessor.Router{Id: "router-uuid"}
	mockIaaS.EXPECT().GetRouter(gomock.Any()).Return(
		&router, nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),
		gomock.Any()).Return(nil)

	cfg := string(`{"router":{"name":"test-create-router"}}`)
	resp := CreateRouter(cfg)
	Convey("TestCreateRouterOK\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestDeleteRouterERR401(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, nil)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	resp := DeleteRouter("the-uuid-of-router")
	Convey("TestDeleteRouterERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 401)
	})
}

func TestDeleteRouterERR404(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		"", errors.New("read-router-return-error"))

	resp := DeleteRouter("the-uuid-of-router")
	Convey("TestDeleteRouterERR404\n", t, func() {
		So(resp.Code, ShouldEqual, 404)
	})
}

func TestDeleteRouter404A(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		"router-uuid", nil)
	mockIaaS.EXPECT().DeleteRouter(gomock.Any()).Return(
		errors.New("error"))

	resp := DeleteRouter("the-uuid-of-router")
	Convey("TestDeleteRouter404A\n", t, func() {
		So(resp.Code, ShouldEqual, 404)
	})
}

func TestDeleteRouter404B(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		"router-uuid", nil)
	mockIaaS.EXPECT().DeleteRouter(gomock.Any()).Return(nil)
	mockDB.EXPECT().DeleteLeaf(gomock.Any()).Return(
		errors.New("error"))

	resp := DeleteRouter("the-uuid-of-router")
	Convey("TestDeleteRouter404B\n", t, func() {
		So(resp.Code, ShouldEqual, 404)
	})
}

func TestDeleteRouter404C(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		"router-uuid", nil)
	mockIaaS.EXPECT().DeleteRouter(gomock.Any()).Return(nil)
	mockDB.EXPECT().DeleteLeaf(gomock.Any()).Return(nil)
	mockDB.EXPECT().DeleteDir(gomock.Any()).Return(
		errors.New("error"))

	resp := DeleteRouter("the-uuid-of-router")
	Convey("TestDeleteRouter404C\n", t, func() {
		So(resp.Code, ShouldEqual, 404)
	})
}

func TestDeleteRouterOK(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		"router-uuid", nil)
	mockIaaS.EXPECT().DeleteRouter(gomock.Any()).Return(nil)
	mockDB.EXPECT().DeleteLeaf(gomock.Any()).Return(nil)
	mockDB.EXPECT().DeleteDir(gomock.Any()).Return(nil)

	resp := DeleteRouter("the-uuid-of-router")
	Convey("TestDeleteRouterOK\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestGetRouterERR401(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, nil)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	resp := GetRouter("the-uuid-of-router")
	Convey("TestGetRouterERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 401)
	})
}

func TestGetRouterERR404B(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		"", errors.New("read-router-from-etcd-error"))

	resp := GetRouter("the-uuid-of-router")
	Convey("TestGetRouterERR404B\n", t, func() {
		So(resp.Code, ShouldEqual, 404)
	})
}

func TestGetRouterOK(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	router := string(`{
    "router": {
        "name": "router_show",
        "id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		router, nil)

	resp := GetRouter("the-uuid-of-router")
	Convey("TestGetRouterOK\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

var Config string = `{
		"router":{
			"name":"auto-test-create-router-should-be-delete"
			}
		}`
var errCfg string = `{
		"router":{
			"name":"auto-test-create-router-should-be-delete"
		}`

func TestUpdateRouterERR401(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, nil)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	resp := UpdateRouter("the-uuid-of-router", Config)
	Convey("TestGetRouterERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 401)
	})
}

func TestUpdateRouterERR403(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	resp := UpdateRouter("the-uuid-of-router", errCfg)
	Convey("TestUpdateRouterERR403\n", t, func() {
		So(resp.Code, ShouldEqual, 403)
	})
}

func TestUpdateRouterERR406A(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockIaaS.EXPECT().UpdateRouter(gomock.Any(),
		gomock.Any(), gomock.Any()).Return(errors.New("error"))

	resp := UpdateRouter("the-uuid-of-router", Config)
	Convey("TestUpdateRouterERR406A\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestUpdateRouterERR406B(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockIaaS.EXPECT().UpdateRouter(gomock.Any(),
		gomock.Any(), gomock.Any()).Return(nil)
	mockIaaS.EXPECT().GetRouter(gomock.Any()).Return(
		nil, errors.New("error"))

	resp := UpdateRouter("the-uuid-of-router", Config)
	Convey("TestUpdateRouterERR406B\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestUpdateRouterERR406C(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	createRouter := iaasaccessor.Router{Id: "router-uuid", Name: "name-router"}
	mockIaaS.EXPECT().UpdateRouter(gomock.Any(),
		gomock.Any(), gomock.Any()).Return(nil)
	mockIaaS.EXPECT().GetRouter(gomock.Any()).Return(
		&createRouter, nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),
		gomock.Any()).Return(errors.New("error"))

	resp := UpdateRouter("the-uuid-of-router", Config)
	Convey("TestUpdateRouterERR406B\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestUpdateRouterOK(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	createRouter := iaasaccessor.Router{Id: "router-uuid", Name: "name-router"}
	mockIaaS.EXPECT().UpdateRouter(gomock.Any(),
		gomock.Any(), gomock.Any()).Return(nil)
	mockIaaS.EXPECT().GetRouter(gomock.Any()).Return(
		&createRouter, nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),
		gomock.Any()).Return(nil)

	resp := UpdateRouter("the-uuid-of-router", Config)
	Convey("TestUpdateRouterOK\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestGetAllRouterERR401(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, nil)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	resp := GetAllRouter()
	Convey("TestGetAllRouterERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 401)
	})
}

func TestGetAllRouterOK1(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(nil,
		errors.New("read-dir-error"))

	resp := GetAllRouter()
	Convey("TestGetAllRouterOK1\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestGetAllRouterOK2(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(nil, nil)

	resp := GetAllRouter()
	Convey("TestGetAllRouterOK2\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestGetAllRouterOK3(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	router := string(`{
    "router": {
        "name": "network_show",
        "id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	var list1 []*client.Node
	node := client.Node{Key: "network-uuid", Value: "network-info"}
	list1 = append(list1, &node)
	list1 = append(list1, &node)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list1, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(router, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return("",
		errors.New("read-leaf-error"))

	resp := GetAllRouter()
	Convey("TestGetAllRouterOK3\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestAttachERR401(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, nil)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	resp := AttachRouter("uuid-for-routet", "uuid-for-network")
	Convey("TestAttachERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 401)
	})
}

func TestAttachERR403(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	resp := AttachRouter403("uuid-for-routet", "uuid-for-network")
	Convey("TestAttachERR403\n", t, func() {
		So(resp.Code, ShouldEqual, 403)
	})
}

func TestAttachERR406A(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		"", errors.New("error"))

	resp := AttachRouter("uuid-for-routet", "uuid-for-network")
	Convey("TestAttachERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestAttachERR406B(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	netInfo := string(`{
    "network": {
        "cidr": "123.124.125.0/24",
        "gateway": "123.124.125.1",
        "name": "network_show",
        "network_id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		netInfo, nil)
	mockIaaS.EXPECT().AttachNetToRouter(
		gomock.Any(), gomock.Any()).Return("", errors.New("error"))

	resp := AttachRouter("uuid-for-routet", "uuid-for-network")
	Convey("TestAttachERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestAttachERR406C(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	netInfo := string(`{
    "network": {
        "cidr": "123.124.125.0/24",
        "gateway": "123.124.125.1",
        "name": "network_show",
        "network_id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		netInfo, nil)
	mockIaaS.EXPECT().AttachNetToRouter(gomock.Any(),
		gomock.Any()).Return("uuid-for-port", nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),
		gomock.Any()).Return(nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),
		gomock.Any()).Return(errors.New("error"))

	resp := AttachRouter("uuid-for-routet", "uuid-for-network")
	Convey("TestAttachERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

/*
func TestAttachERR406D(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	common.SetIaaS(mockIaaS)

	netInfo := string(`{
    "network": {
        "cidr": "123.124.125.0/24",
        "gateway": "123.124.125.1",
        "name": "network_show",
        "network_id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		netInfo, nil)
	mockIaaS.EXPECT().AttachNetToRouter(gomock.Any(),
		gomock.Any()).Return("uuid-for-port",nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),gomock.Any()).Return(nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),gomock.Any()).Return(nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),gomock.Any()).Return(
		errors.New("error"))

	resp := AttachRouter("uuid-for-routet","uuid-for-network")
	Convey("TestAttachERR406C\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestAttachERR406E(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	common.SetIaaS(mockIaaS)

	netInfo := string(`{
    "network": {
        "cidr": "123.124.125.0/24",
        "gateway": "123.124.125.1",
        "name": "network_show",
        "network_id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		netInfo, nil)
	mockIaaS.EXPECT().AttachNetToRouter(gomock.Any(),
		gomock.Any()).Return("uuid-for-port",nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),gomock.Any()).Return(nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),gomock.Any()).Return(nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),gomock.Any()).Return(nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),gomock.Any()).Return(nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),
		gomock.Any()).Return(errors.New("error"))

	resp := AttachRouter("uuid-for-routet","uuid-for-network")
	Convey("TestAttachERR406C\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}
*/
func TestAttachOK(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	netInfo := string(`{
    "network": {
        "cidr": "123.124.125.0/24",
        "gateway": "123.124.125.1",
        "name": "network_show",
        "network_id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		netInfo, nil)
	mockIaaS.EXPECT().AttachNetToRouter(gomock.Any(),
		gomock.Any()).Return("uuid-for-port", nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),
		gomock.Any()).Return(nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),
		gomock.Any()).Return(nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),
		gomock.Any()).Return(nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),
		gomock.Any()).Return(nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),
		gomock.Any()).Return(nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),
		gomock.Any()).Return(nil)

	resp := AttachRouter("uuid-for-routet", "uuid-for-network")
	Convey("TestAttachERR406D\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestDetachERR401(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, nil)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	resp := DetachRouter("uuid-for-routet", "uuid-for-network")
	Convey("TestDetachERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 401)
	})
}

func TestDetachERR403(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	resp := DetachRouter403("uuid-for-routet", "uuid-for-network")
	Convey("TestDetachERR403\n", t, func() {
		So(resp.Code, ShouldEqual, 403)
	})
}

func TestDetachERR406A(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		"", errors.New("error"))

	resp := DetachRouter("uuid-for-routet", "uuid-for-network")
	Convey("TestAttachERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestDetachERR406B(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	MockPaasAdminCheck(mockDB)
	netInfo := string(`{
    "network": {
        "cidr": "123.124.125.0/24",
        "gateway": "123.124.125.1",
        "name": "network_show",
        "network_id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		netInfo, nil)
	mockIaaS.EXPECT().DetachNetFromRouter(gomock.Any(),
		gomock.Any()).Return("", errors.New("error"))

	resp := DetachRouter("uuid-for-routet", "uuid-for-network")
	Convey("TestDetachERR406B\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestDetachERR406C(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	netInfo := string(`{
    "network": {
        "cidr": "123.124.125.0/24",
        "gateway": "123.124.125.1",
        "name": "network_show",
        "network_id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		netInfo, nil)
	mockIaaS.EXPECT().DetachNetFromRouter(gomock.Any(),
		gomock.Any()).Return("uuid-for-port", nil)
	mockDB.EXPECT().DeleteLeaf(
		gomock.Any()).Return(errors.New("error"))

	resp := DetachRouter("uuid-for-routet", "uuid-for-network")
	Convey("TestDetachERR406C\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestDetachERR406D(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	netInfo := string(`{
    "network": {
        "cidr": "123.124.125.0/24",
        "gateway": "123.124.125.1",
        "name": "network_show",
        "network_id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		netInfo, nil)
	mockIaaS.EXPECT().DetachNetFromRouter(gomock.Any(),
		gomock.Any()).Return("uuid-for-port", nil)
	mockDB.EXPECT().DeleteLeaf(gomock.Any()).Return(nil)
	mockDB.EXPECT().DeleteLeaf(
		gomock.Any()).Return(errors.New("error"))

	resp := DetachRouter("uuid-for-routet", "uuid-for-network")
	Convey("TestDetachERR406D\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestDetachERR406E(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	netInfo := string(`{
    "network": {
        "cidr": "123.124.125.0/24",
        "gateway": "123.124.125.1",
        "name": "network_show",
        "network_id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		netInfo, nil)
	mockIaaS.EXPECT().DetachNetFromRouter(gomock.Any(),
		gomock.Any()).Return("uuid-for-port", nil)
	mockDB.EXPECT().DeleteLeaf(gomock.Any()).Return(nil)
	mockDB.EXPECT().DeleteLeaf(gomock.Any()).Return(nil)
	mockDB.EXPECT().DeleteDir(gomock.Any()).Return(errors.New("error"))

	resp := DetachRouter("uuid-for-routet", "uuid-for-network")
	Convey("TestDetachERR406D\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestDetachOK(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	netInfo := string(`{
    "network": {
        "cidr": "123.124.125.0/24",
        "gateway": "123.124.125.1",
        "name": "network_show",
        "network_id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	MockPaasAdminCheck(mockDB)
	gomock.InOrder(
		mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
			netInfo, nil),
		mockIaaS.EXPECT().DetachNetFromRouter(gomock.Any(),
			gomock.Any()).Return("uuid-for-port", nil),
		mockDB.EXPECT().DeleteLeaf(gomock.Any()).Return(nil),
		mockDB.EXPECT().DeleteLeaf(gomock.Any()).Return(nil),
		mockDB.EXPECT().DeleteDir(gomock.Any()).Return(nil),
	)

	resp := DetachRouter("uuid-for-routet", "uuid-for-network")
	Convey("TestDetachERR406E\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}
