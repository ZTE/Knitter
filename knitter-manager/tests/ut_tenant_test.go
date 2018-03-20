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
	"encoding/json"
	"errors"
	"github.com/ZTE/Knitter/knitter-manager/controllers"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/bouk/monkey"
	"github.com/golang/gostub"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"reflect"
	"testing"
)

//func TestUpdateTenant_OK(t *testing.T) {
//	cfgMock := gomock.NewController(t)
//	defer cfgMock.Finish()
//	mockDB := NewMockDbAccessor(cfgMock)
//	common.SetDataBase(mockDB)
//	//common.SetIaaS(nil)
//
//	MockPaasAdminCheck(mockDB)
//	tenantInfo := string(`{
//        		"tenant_uuid": "paas-admin",
//        		"net_quota": 2,
//        		"net_number": 1
//    	}`)
//	var list1 []*client.Node
//	node := client.Node{Key: "network-uuid", Value: "network-info"}
//	list1 = append(list1, &node)
//	gomock.InOrder(
//		mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(tenantInfo, nil),
//		mockDB.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil),
//	)
//
//	resp := UpdateTenant("5")
//	Convey("TestUpdateTenant_OK\n", t, func() {
//		So(resp.Code, ShouldEqual, 200)
//	})
//}

func TestUpdateTenant_OK(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)

	MockPaasAdminCheck(mockDB)

	monkey.Patch(models.ConvertQuota, func(quota string, defaultValue int) (int, bool) {
		return 5, true
	})
	defer monkey.UnpatchAll()
	stubs := gostub.StubFunc(&models.GetNetNumOfTenant, 1)
	defer stubs.Reset()
	var tenant *models.Tenant
	monkey.PatchInstanceMethod(reflect.TypeOf(tenant), "UpdateQuota",
		func(_ *models.Tenant, quota int) error {
			return nil
		})
	resp := UpdateTenant("5")
	Convey("TestUpdateTenant_OK\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestGetTenantAdmin_OK(t *testing.T) {
	var te *models.Tenant
	mtenant := &models.Tenant{
		TenantName:     "admin",
		TenantUUID:     "admin",
		Networks:       "networks",
		Interfaces:     "interfaces",
		Quota:          100,
		NetNum:         4,
		IsCancelling:   false,
		CreateTime:     "createTime",
		IaasTenantID:   "iaasTenantID1",
		IaasTenantName: "iaasTenanName1",
	}
	monkey.PatchInstanceMethod(reflect.TypeOf(te), "GetTenantInfo",
		func(*models.Tenant) (_ *models.Tenant, _ error) {
			return mtenant, nil
		})
	defer monkey.UnpatchAll()
	resp := GetTenant("admin")
	Convey("TestGetTenantAdmin_OK\n", t, func() {
		tenant := controllers.PaasTenantRsp{}
		json.Unmarshal(resp.Body.Bytes(), &tenant)
		So(tenant.Tenant.ID, ShouldEqual, "admin")
		So(tenant.Tenant.Name, ShouldEqual, "admin")
		So(tenant.Tenant.NetNumber, ShouldEqual, 4)
		So(tenant.Tenant.Quotas.Network, ShouldEqual, 100)
	})
}

func TestGetTenants_OK(t *testing.T) {
	mtenants := []*models.Tenant{
		{
			TenantName:     "admin",
			TenantUUID:     "admin",
			Networks:       "networks",
			Interfaces:     "interfaces",
			Quota:          100,
			NetNum:         4,
			IsCancelling:   false,
			CreateTime:     "createTime",
			IaasTenantID:   "iaasTenantID",
			IaasTenantName: "iaasTenanName",
		},
		{
			TenantName:     "network",
			TenantUUID:     "network",
			Networks:       "networks1",
			Interfaces:     "interfaces1",
			Quota:          100,
			NetNum:         4,
			IsCancelling:   false,
			CreateTime:     "createTime",
			IaasTenantID:   "iaasTenantID1",
			IaasTenantName: "iaasTenanName1",
		},
	}
	monkey.Patch(models.GetAllTenants, func() []*models.Tenant {
		return mtenants
	})
	defer monkey.UnpatchAll()
	resp := GetTenants()
	Convey("TestGetTenantAdmin_OK\n", t, func() {
		tenants := controllers.PaasTenantsRsp{}
		json.Unmarshal(resp.Body.Bytes(), &tenants)
		So(tenants.Tenants[0].ID, ShouldEqual, "admin")
		So(tenants.Tenants[0].Name, ShouldEqual, "admin")
		So(tenants.Tenants[0].NetNumber, ShouldEqual, 4)
		So(tenants.Tenants[0].Quotas.Network, ShouldEqual, 100)
		So(tenants.Tenants[1].ID, ShouldEqual, "network")
		So(tenants.Tenants[1].Name, ShouldEqual, "network")
		So(tenants.Tenants[1].NetNumber, ShouldEqual, 4)
		So(tenants.Tenants[1].Quotas.Network, ShouldEqual, 100)
	})
}

func TestGetTenants_ERR1(t *testing.T) {
	monkey.Patch(models.GetAllTenants, func() []*models.Tenant {
		return nil
	})
	resp := GetTenants()
	Convey("TestGetTenants_ERR1\n", t, func() {
		tenants := controllers.EncapPaasTenants{}
		json.Unmarshal(resp.Body.Bytes(), &tenants)
		So(resp.Code, ShouldEqual, 404)

	})
}

func TestUpdateTenant_ConvertQuota_false(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)
	MockPaasAdminCheck(mockDB)
	monkey.Patch(models.ConvertQuota, func(quota string, defaultValue int) (int, bool) {
		return 10, false
	})
	defer monkey.UnpatchAll()
	resp := UpdateTenant("")
	Convey("TestUpdateTenant_ConvertQuota_false\n", t, func() {
		So(resp.Code, ShouldEqual, 400)
	})
}

func TestUpdateTenant_GetNetNumOfTenant_greaterthan(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)

	MockPaasAdminCheck(mockDB)
	monkey.Patch(models.ConvertQuota, func(quota string, defaultValue int) (int, bool) {
		return 100, true
	})
	defer monkey.UnpatchAll()
	stubs := gostub.StubFunc(&models.GetNetNumOfTenant, 101)
	defer stubs.Reset()
	resp := UpdateTenant("1001")
	Convey("TestUpdateTenant_GetNetNumOfTenant_greaterthan\n", t, func() {
		So(resp.Code, ShouldEqual, 400)
	})
}

func TestUpdateTenant_UpdateQuota_err(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)

	MockPaasAdminCheck(mockDB)
	monkey.Patch(models.ConvertQuota, func(quota string, defaultValue int) (int, bool) {
		return 100, true
	})
	defer monkey.UnpatchAll()
	stubs := gostub.StubFunc(&models.GetNetNumOfTenant, 77)
	defer stubs.Reset()
	var te *models.Tenant
	monkey.PatchInstanceMethod(reflect.TypeOf(te), "UpdateQuota",
		func(_ *models.Tenant, quota int) error {
			return models.BuildErrWithCode(http.StatusNotFound, errors.New("read leaf err"))
		})
	resp := UpdateTenant("Error Value")
	Convey("TestUpdateTenant_UpdateQuota_err\n", t, func() {
		So(resp.Code, ShouldEqual, 404)
	})
}

func TestDeleteTenant_ERR1(t *testing.T) {

	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, nil)
	defer stubs.Reset()

	resp := DeleteTenant()
	Convey("TestDeleteTenant_ERR1\n", t, func() {
		So(resp.Code, ShouldEqual, 401)
	})

}

func TestDeleteTenant_ERR2(t *testing.T) {

	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return("", errors.New("read leaf error"))
	resp := DeleteTenant()
	Convey("TestDeleteTenant_ERR2\n", t, func() {
		So(resp.Code, ShouldEqual, 404)
	})
}

func TestDeleteTenant_ERR3(t *testing.T) {

	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()
	tenantInfo := string(`{
        		"tenant_uuid": "paas-admin",
  				"is_cancelling": true
    	}`)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(tenantInfo, nil)
	resp := DeleteTenant()
	Convey("TestDeleteTenant_ERR3\n", t, func() {
		So(resp.Code, ShouldEqual, 409)
	})
}

func TestDeleteTenant_SetTenantCancellingStatusErr(t *testing.T) {

	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()
	tenantInfo := models.Tenant{
		TenantName:   "paas-admin",
		TenantUUID:   "paas-admin",
		IsCancelling: false,
		Networks:     "networks1",
		Interfaces:   "interfaces1",
		Quota:        1,
		NetNum:       1,
		CreateTime:   "createtime1",
		IaasTenantID: "tenantid1",
	}
	tenantInfoByte, _ := json.Marshal(&tenantInfo)
	mockDB.EXPECT().ReadLeaf("/paasnet/tenants/paas-admin/self").Return(string(tenantInfoByte), nil)
	saveTenantInfo := models.Tenant{
		TenantName:   "paas-admin",
		TenantUUID:   "paas-admin",
		IsCancelling: true,
		Networks:     "networks1",
		Interfaces:   "interfaces1",
		Quota:        1,
		NetNum:       1,
		CreateTime:   "createtime1",
		IaasTenantID: "tenantid1",
	}
	saveTenantInfoByte, _ := json.Marshal(&saveTenantInfo)
	mockDB.EXPECT().SaveLeaf("/paasnet/tenants/paas-admin/self", string(saveTenantInfoByte)).Return(errors.New("etcd misconfig"))
	resp := DeleteTenant()
	Convey("TestDeleteTenant_SetTenantCancellingStatusErr\n", t, func() {
		So(resp.Code, ShouldEqual, 500)
	})
}
