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
	"flag"
	"testing"

	"encoding/json"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/knitter-manager/public"
	_ "github.com/ZTE/Knitter/knitter-manager/routers"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/openstack"
	"github.com/antonholmquist/jason"
	"github.com/coreos/etcd/client"
	"github.com/golang/gostub"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

type TestCase struct {
	Name string
	Desp string
	Cfg  string
	Code int
}

const TestCaseLogPath = "/home/m11/knitter/src/knitter-manager/tests/logs"

func init() {
	klog.ConfigLog(TestCaseLogPath)
	flag.Parse()
}

func TestConfigRegularCheckTimeOK(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)

	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil)
	cfg := string(`{"regular_check": "30"}`)
	resp := ConfigOpenStack(cfg)
	Convey("TestConfigRegularCheckTimeOK\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestConfigRegularCheckTimeErr(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)

	MockPaasAdminCheck(mockDB)
	mockDB.EXPECT().SaveLeaf(gomock.Any(),
		gomock.Any()).Return(errors.New("save data error"))
	cfg := string(`{"regular_check": "30"}`)
	resp := ConfigOpenStack(cfg)
	Convey("TestConfigRegularCheckTimeErr\n", t, func() {
		So(resp.Code, ShouldEqual, 500)
	})
}

func TestConfigErr403(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)

	MockPaasAdminCheck(mockDB)
	cfg := string(`{"error_json": {"ip":"192.168.1.100"}`)
	resp := ConfigOpenStack(cfg)
	Convey("TestConfigErr403\n", t, func() {
		So(resp.Code, ShouldEqual, 400)
	})
}

//func TestManagerTypeNil(t *testing.T) {
//	resp := GetType()
//	Convey("TestManagerType\n", t, func() {
//		So(resp.Code, ShouldEqual, 200)
//		So(resp.Body.String(), ShouldEqual, "")
//	})
//}

func TestManagerType(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	//iaas.SetIaaS(constvalue.DefaultIaasTenantID, mockIaaS)

	Convey("TestManagerType\n", t, func() {
		MockPaasAdminCheck(mockDB)
		stub := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
		defer stub.Reset()
		mockIaaS.EXPECT().GetType().Return("vNM")
		resp := GetType()
		So(resp.Code, ShouldEqual, 200)
		So(resp.Body.String(), ShouldEqual, `"vNM"`)
	})

	Convey("TestManagerType\n", t, func() {
		MockPaasAdminCheck(mockDB)
		stub := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
		defer stub.Reset()
		mockIaaS.EXPECT().GetType().Return("TECS")
		resp := GetType()
		So(resp.Code, ShouldEqual, 200)
		So(resp.Body.String(), ShouldEqual, `"TECS"`)
	})

	Convey("TestManagerType\n", t, func() {
		MockPaasAdminCheck(mockDB)
		stub := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
		defer stub.Reset()
		mockIaaS.EXPECT().GetType().Return("EMBEDDED")
		resp := GetType()
		So(resp.Code, ShouldEqual, 200)
		So(resp.Body.String(), ShouldEqual, `"EMBEDDED"`)
	})

	Convey("TestManagerType\n", t, func() {
		MockPaasAdminCheck(mockDB)
		stub := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
		defer stub.Reset()
		mockIaaS.EXPECT().GetType().Return("VNFM")
		resp := GetType()
		So(resp.Code, ShouldEqual, 200)
		So(resp.Body.String(), ShouldEqual, `"VNFM"`)
	})
}

func TestCheckHealth(t *testing.T) {
	Convey("TestCheckHealth bad\n", t, func() {
		cfgMock := gomock.NewController(t)
		defer cfgMock.Finish()
		mockDB := NewMockDbAccessor(cfgMock)
		//mockIaaS := NewMockIaaS(cfgMock)
		common.SetDataBase(mockDB)
		stub := gostub.StubFunc(&iaas.GetIaaS, nil)
		defer stub.Reset()
		resp := GetHealth()
		So(resp.Code, ShouldEqual, 200)
		println(resp.Body.String())
		health := models.HealthObj{}
		json.Unmarshal(resp.Body.Bytes(), &health)
		So(health.Level, ShouldEqual, 0)
		So(health.State, ShouldEqual, "bad")
	})

	Convey("TestCheckHealth good\n", t, func() {
		cfgMock := gomock.NewController(t)
		defer cfgMock.Finish()
		mockDB := NewMockDbAccessor(cfgMock)
		mockIaaS := NewMockIaaS(cfgMock)
		common.SetDataBase(mockDB)
		stub := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
		defer stub.Reset()
		resp := GetHealth()
		So(resp.Code, ShouldEqual, 200)
		println(resp.Body.String())
		health := models.HealthObj{}
		json.Unmarshal(resp.Body.Bytes(), &health)
		So(health.Level, ShouldEqual, 0)
		So(health.State, ShouldEqual, "good")
	})
}

//func TestCheckHealthBad(t *testing.T) {
//	stub := gostub.StubFunc(&common.InitIaaS, nil)
//	defer stub.Reset()
//	common.SetDataBase(nil)
//	common.SetIaaS(nil)
//
//	resp := GetHealth()
//	Convey("TestCheckHealth\n", t, func() {
//		So(resp.Code, ShouldEqual, 200)
//		println(resp.Body.String())
//		health := models.HealthObj{}
//		json.Unmarshal(resp.Body.Bytes(), &health)
//		So(health.Level, ShouldEqual, 0)
//		So(health.State, ShouldEqual, "bad")
//	})
//}

func TestConfigERR505(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)

	MockPaasAdminCheck(mockDB)
	cfg := string(`{"error_config_req": "error"}`)
	resp := ConfigOpenStack(cfg)
	Convey("TestConfigERR505\n", t, func() {
		So(resp.Code, ShouldEqual, 400)
	})
}

func TestInitNoAuth(t *testing.T) {

	testCaseList := []TestCase{
		{Name: "Unit-Test", Desp: "TestInitNoAuth" + "---want---" + "Error(From cfg.GetObject)", Code: 404,
			Cfg: string(`{"iaas": {"noauth_openstack": {}  } }`),
		},
		{Name: "Unit-Test", Desp: "TestInitNoAuth" + "---want---" + "Error(From noauthCfg.GetString(ip))", Code: 404,
			Cfg: string(`{"iaas": {"noauth_openstack": {"config": {"url": "noauth_openstack_config_url","tenant_id": ""} } } }`),
		},
		{Name: "Unit-Test", Desp: "TestInitNoAuth" + "---want---" + "Error(From noauthCfg.GetString(ip) -> '')", Code: 404,
			Cfg: string(`{"iaas": {"noauth_openstack": {"config": {"ip": "","url": "noauth_openstack_config_url","tenant_id": ""} } } }`),
		},
		{Name: "Unit-Test", Desp: "TestInitNoAuth" + "---want---" + "Error(From noauthCfg.GetString)", Code: 404,
			Cfg: string(`{"iaas": {"noauth_openstack": {"config": {"ip": "man_network_ip","url": "noauth_openstack_config_url"} } } }`),
		},
		{Name: "Unit-Test", Desp: "TestInitNoAuth" + "---want---" + "All OK", Code: 200,
			Cfg: string(`{"iaas": {"noauth_openstack": {"config": {"ip": "man_network_ip","url": "noauth_openstack_config_url","tenant_id": ""} } } }`),
		},
	}

	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetDefaultPhysnet, "", errors.New("default physnet record is not exist"))
	defer stubs.Reset()
	mockDB.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil)

	stubs.StubFunc(&iaas.CheckNetService, nil)
	defer stubs.Reset()
	Convey("TestInitNoAuth \n", t, func() {
		for _, testCase := range testCaseList {
			Convey(testCase.Desp, func() {
				cfgJason, _ := jason.NewObjectFromBytes([]byte(testCase.Cfg))
				resp := iaas.InitNoAuth(cfgJason)
				if testCase.Code == 404 {
					So(resp, ShouldNotBeNil)
				} else {
					So(resp, ShouldBeNil)
				}
			})
		}
	})
}

func TestSetNetQuota(t *testing.T) {
	cfg := string(`{"net_quota": {"admin":"100", "no_admin":"10"}}`)
	obj, _ := jason.NewObjectFromBytes([]byte(cfg))
	models.SetNetQuota(obj)
}

func TestConvertAttachReqMax(t *testing.T) {
	expNum := ""
	atuNum := models.ConvertAttachReqMax(expNum)
	Convey("TestConvertAttachReqMax \n", t, func() {
		So(atuNum, ShouldEqual, openstack.MaxReqForAttach)
	})
	expNum = "error-num"
	atuNum = models.ConvertAttachReqMax(expNum)
	Convey("TestConvertAttachReqMax \n", t, func() {
		So(atuNum, ShouldEqual, openstack.MaxReqForAttach)
	})
	expNum = "31"
	atuNum = models.ConvertAttachReqMax(expNum)
	Convey("TestConvertAttachReqMax \n", t, func() {
		So(atuNum, ShouldEqual, openstack.MaxReqForAttach)
	})
	expNum = "6"
	atuNum = models.ConvertAttachReqMax(expNum)
	Convey("TestConvertAttachReqMax \n", t, func() {
		So(atuNum, ShouldEqual, 6)
	})
}

func TestUpdateEtcd4NetQuota(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)

	var list1 []*client.Node
	node := client.Node{Key: "network-uuid", Value: "network-info"}
	list1 = append(list1, &node)
	tenantInfo := string(`{
        		"tenant_uuid": "paas-admin",
        		"net_quota": 0,
        		"net_number": 1
    	}`)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list1, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(tenantInfo, nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil)

	err := models.UpdateEtcd4NetQuota()
	Convey("TestUpdateEtcd4NetQuota\n", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestUpdateEtcd4NetQuota_Err1(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)

	mockDB.EXPECT().ReadDir(gomock.Any()).Return(nil, errors.New("read dir error"))

	err := models.UpdateEtcd4NetQuota()
	Convey("TestUpdateEtcd4NetQuota\n", t, func() {
		So(err.Error(), ShouldEqual, "read dir error")
	})
}

func TestUpdateEtcd4NetQuota_Err2(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)

	var list1 []*client.Node
	node := client.Node{Key: "/paasnet/tenants/admin", Value: "network-info"}
	list1 = append(list1, &node)

	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list1, nil)

	err := models.UpdateEtcd4NetQuota()
	Convey("TestUpdateEtcd4NetQuota\n", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestUpdateEtcd4NetQuota_Err3(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)

	var list1 []*client.Node
	node := client.Node{Key: "network-uuid", Value: "network-info"}
	list1 = append(list1, &node)
	tenantInfo := string(`{
        		"tenant_uuid": "paas-admin",
        		"net_quota": 0,
        		"net_number": 1
    	}`)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list1, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(tenantInfo, errors.New("read leaf error"))

	err := models.UpdateEtcd4NetQuota()
	Convey("TestUpdateEtcd4NetQuota\n", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestUpdateEtcd4NetQuota_Err4(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)

	var list1 []*client.Node
	node := client.Node{Key: "network-uuid", Value: "network-info"}
	list1 = append(list1, &node)
	tenantInfo := string(`{
        		"tenant_uuid": "paas-admin",
        		"net_quota": 0,
        		"net_number": 1
    	}`)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list1, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(tenantInfo, nil)
	mockDB.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(errors.New("save leaf error"))

	err := models.UpdateEtcd4NetQuota()
	Convey("TestUpdateEtcd4NetQuota\n", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestSaveVnfmConfg(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)
	Convey("TestSaveVnfmConfg---OK\n", t, func() {
		mockDB.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil)
		value := []byte{45, 46, 47}
		err := models.SaveVnfmConfg(value)
		So(err, ShouldBeNil)
	})
	Convey("TestSaveVnfmConfg---ERR\n", t, func() {
		mockDB.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(errors.New("save leaf error"))
		value := []byte{45, 46, 47}
		err := models.SaveVnfmConfg(value)
		So(err, ShouldNotBeNil)
	})
}

//func TestCheckVnfmConfig(t *testing.T) {
//	conf := vnfm.VnfmConf{
//		URL: "http://10.62.100.202:80/api/nf_m_i/v1/",
//		NfInstanceID: "1",
//	}
//	cfgMock := gomock.NewController(t)
//	defer cfgMock.Finish()
//	mockHTTPMethods := mock_http.NewMockHTTPMethods(cfgMock)
//	stubs := gostub.StubFunc(&http.GetHTTPClientObj, mockHTTPMethods)
//	defer stubs.Reset()
//	Convey("TestCheckVnfmConfig---OK\n", t, func() {
//		mockHTTPMethods.EXPECT().Delete(gomock.Any()).Return(404, errors.New("404 err"))
//		err := models.CheckVnfmConfig(conf)
//		So(err, ShouldEqual, nil)
//	})
//	Convey("TestCheckVnfmConfig---ERR\n", t, func() {
//		mockHTTPMethods.EXPECT().Delete(gomock.Any()).Return(500, errors.New("500 err"))
//		err := models.CheckVnfmConfig(conf)
//		So(err, ShouldNotEqual, nil)
//	})
//}

func TestInitConfigurationOK(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	cfg := string(`{
		"init_configuration": "init_configuration_obj",
		"networks": "networks_obj"
	}`)
	stubs := gostub.StubFunc(&models.CfgInit, nil)
	defer stubs.Reset()
	resp := APIPostInitConfiguration(cfg)
	Convey("Test_InitConfiguration_OK\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestInitConfigurationErr(t *testing.T) {

	Convey("Test_InitConfiguration_ERR_CfgInit\n", t, func() {
		cfgMock := gomock.NewController(t)
		defer cfgMock.Finish()
		cfg := string(`{
			"init_configuration": "init_configuration_obj",
			"networks": "networks_obj"
		}`)
		stubs := gostub.StubFunc(&models.CfgInit, errors.New("404::GetSceneByKnitterJSON err"))
		defer stubs.Reset()
		resp := APIPostInitConfiguration(cfg)
		So(resp.Code, ShouldEqual, 404)
	})
	Convey("Test_InitConfiguration_ERR_NewObjectFromBytes\n", t, func() {
		cfgMock := gomock.NewController(t)
		defer cfgMock.Finish()
		cfg := string(`{
			"init_configuration": "init_configuration_obj"
			"networks": "networks_obj"
		}`)
		resp := APIPostInitConfiguration(cfg)
		So(resp.Code, ShouldEqual, 415)
	})
}
