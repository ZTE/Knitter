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

package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/astaxie/beego"
	"io/ioutil"
	"net/http"
	"time"
)

// Operations about tenant
type TenantController struct {
	beego.Controller
}

type EncapPaasTenant struct {
	Tenant *models.PaasTenant `json:"tenant"`
}

type EncapPaasTenants struct {
	Tenants []*models.PaasTenant `json:"tenants"`
}

// @Title get
// @Description show tenant
// @Param       tenant_uuid             path    string  true            "The tenant_uuid you want to show"
// @Success 200 {string} get success!
// @Failure 404 tenant not Exist
// @router /:user [get]
func (self *TenantController) Get() {
	defer RecoverRsp500(&self.Controller)
	id := self.Ctx.Input.Param(":user")
	klog.Info("Request tenant_id: ", id)

	tenant := models.Tenant{TenantUUID: id}
	tenantInfo, err := tenant.GetTenantInfo()
	if err != nil {
		HandleErr(&self.Controller, err)
		return
	}
	var status string = "ACTIVE"
	if tenantInfo.IsCancelling == true {
		status = "DELETING"
	}
	self.Data["json"] = PaasTenantRsp{
		Tenant: TenantRsp{
			CreatedAt: tenantInfo.CreateTime,
			ID:        tenantInfo.TenantUUID,
			Name:      tenantInfo.TenantName,
			NetNumber: tenantInfo.NetNum,
			IaasTenant: IaasTenantRsp{
				TenantName: tenantInfo.IaasTenantName,
				ID:         tenantInfo.IaasTenantID,
			},
			Quotas: QuotasRsp{
				Network: tenantInfo.Quota,
			},
			Status: status,
		},
	}
	self.ServeJSON()
}

// @Title delete
// @Description delete tenant all networks
// @Param	tenant_uuid		path 	string	true		"The tenant_uuid you want to delete"
// @Success 200 {string} delete success!
// @Failure 404 tenant not Exist
// @router /:user [delete]
func (self *TenantController) Delete() {
	var err error

	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	tenant := &models.Tenant{TenantName: paasTenantID, TenantUUID: paasTenantID}
	klog.Infof("start Delete Tenant [ %v ]!", tenant.TenantUUID)
	tenantKey := dbaccessor.GetKeyOfTenantSelf(paasTenantID)
	tenantValue, err1 := common.GetDataBase().ReadLeaf(tenantKey)
	if err1 != nil {
		klog.Errorf("Tenant [%v] not exist error! -%v", paasTenantID, err1)
		NotfoundErr404(&self.Controller, fmt.Errorf("%v:tenant not exist error", err1))
		err = err1
		return
	}
	json.Unmarshal([]byte(tenantValue), tenant)
	if tenant.IsCancelling == true {
		klog.Errorf("Tenant [%v] is now being deleted !", paasTenantID)
		HandleErr(&self.Controller, models.BuildErrWithCode(http.StatusConflict, fmt.Errorf("Tenant [%v] is already being deleted", paasTenantID)))
		err = errors.New("Tenant is already being delete")
		return
	}
	klog.Infof("set Tenant[uuid: %v] cancelling status START", tenant.TenantUUID)
	err = tenant.Cancel()
	if err != nil {
		klog.Errorf("tenant.Cancel() [id: %s] FAILED, error%v", paasTenantID, err)
		HandleErr(&self.Controller, err)
		return
	}
	klog.Infof("set Tenant[UUID: %s] cancelling status SUCC", tenant.TenantUUID)

	self.Data["json"] = map[string]string{"msg": fmt.Sprintf("Start to delete tenant [%v]'s networks.", paasTenantID)}
	self.ServeJSON()
}

// @Title create
// @Description create tenant with lan network
// @Param	body		body 	models.EncapNetwork	true		"The tenant_uuid you want to create"
// @Success 200 {string} models.Network.Id
// @Failure 403 invalid request body
// @Failure 406 create tenant error
// @router /:user [post]
func (self *TenantController) Post() {
	var err error
	defer RecoverRsp500(&self.Controller)
	if iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	tenantID := self.GetString(":user")
	tenant := getDefaultTenant(tenantID)
	tenantKey := dbaccessor.GetKeyOfTenantSelf(tenantID)

	tenantValue, err := common.GetDataBase().ReadLeaf(tenantKey)
	if err == nil {
		json.Unmarshal([]byte(tenantValue), tenant)
		klog.Errorf("Tenant [%v] is already in etcd, status [is being deleted? = %v]!",
			tenant.TenantUUID, tenant.IsCancelling)

		HandleErr(&self.Controller,
			models.BuildErrWithCode(http.StatusConflict, errors.New("tenant already exists")))
		err = errors.New("tenant already exists")
		err = errors.New("tenant " + tenant.TenantUUID + " already exists")
		return
	}
	tenant.CreateTime = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	if iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID).GetType() == "VNFM" {
		tenant.IaasTenantID = constvalue.DefaultIaasTenantID
		tenant.TenantName = constvalue.DefaultIaasTenantName
		tenant.SaveTenantToEtcd()
		self.Data["json"] = tenant
		self.ServeJSON()
		return
	}
	adminTenantInfo := &models.Tenant{}
	adminTenantInfo, err = models.GetPaasTenantInfoFromDB(constvalue.PaaSTenantAdminDefaultUUID)
	if err != nil {
		HandleErr(&self.Controller,
			models.BuildErrWithCode(http.StatusInternalServerError,
				errors.New("get iaas infomation failed")))
		return
	}
	tenant.IaasTenantID = adminTenantInfo.IaasTenantID
	tenant.IaasTenantName = adminTenantInfo.IaasTenantName
	tenant.NetNum = 1
	tenant.SaveTenantToEtcd()
	net := getDefaultNet(tenantID)
	klog.Infof("Create Tenant uuid [ %v ]!", net.TenantUUID)
	err = net.Create()
	if err != nil {
		klog.Errorf("Create tenant [%v]'s lan network error!", tenant.TenantUUID)
		key := dbaccessor.GetKeyOfTenant(tenantID)
		common.GetDataBase().DeleteDir(key)
		HandleErr(&self.Controller, fmt.Errorf("%v:create tenant lan network error", err))
		return
	}
	self.Data["json"] = PaasTenantRsp{
		Tenant: TenantRsp{
			CreatedAt: tenant.CreateTime,
			ID:        tenant.TenantUUID,
			Name:      tenant.TenantName,
			NetNumber: tenant.NetNum,
			IaasTenant: IaasTenantRsp{
				TenantName: tenant.IaasTenantName,
				ID:         tenant.IaasTenantID,
			},
			Quotas: QuotasRsp{
				Network: tenant.Quota,
			},
			Status: "ACTIVE",
		},
	}
	self.ServeJSON()
	return
}

func getDefaultNet(tenantID string) *models.Net {
	net := models.Net{}
	net.Network.Name = "lan"
	net.Subnet.Cidr = "100.100.0.0/16"
	net.Subnet.GatewayIp = "100.100.0.1"
	net.Public = false
	net.TenantUUID = tenantID
	net.Description = "Default tenant network"
	return &net
}

func getDefaultTenant(tenantID string) *models.Tenant {
	t := models.Tenant{
		TenantName:   tenantID,
		TenantUUID:   tenantID,
		IsCancelling: false,
		Networks:     dbaccessor.GetKeyOfNetworkGroup(tenantID),
		Interfaces:   dbaccessor.GetKeyOfInterfaceGroup(tenantID),
		Quota:        models.QuotaNoAdmin,
		NetNum:       0}
	return &t
}

// @Title update
// @Description update tenant quota
// @Success 200 {string} models.Tenant
// @Failure 406 update quota error
// @router /:user/quota/ [put]
func (self *TenantController) Update() {
	defer RecoverRsp500(&self.Controller)

	tenant := models.Tenant{}
	tenant.TenantUUID = self.GetString(":user")
	quota, isValid := models.ConvertQuota(self.GetString("value"), models.DefaultQuotaNoAdmin)
	if isValid == false {
		Err400(&self.Controller, errors.New("invalid input quota"))
		return
	}
	netNum := models.GetNetNumOfTenant(tenant.TenantUUID)
	if quota < netNum {
		Err400(&self.Controller, fmt.Errorf("quota:%v is less than net number:%v, input error", quota, netNum))
		return
	}
	klog.Infof("start update Tenant [ %v ], Quota [ %v ]!", tenant.TenantUUID, tenant.Quota)
	err := tenant.UpdateQuota(quota)
	if err != nil {
		klog.Errorf("tenant:%v, quota:%v, error:%v, UpdateQuota Failed",
			tenant.TenantUUID, quota, err)
		HandleErr(&self.Controller, err)
		return
	}
	self.Data["json"] = tenant
	self.Data["json"] = EncapPaasTenant{Tenant: &models.PaasTenant{
		TenantName: tenant.TenantName,
		TenantUUID: tenant.TenantUUID,
		NetNum:     netNum,
		CreateTime: tenant.CreateTime,
		Quotas:     models.PaasQuotas{Network: tenant.Quota},
		Status:     "ACTIVE",
	}}
	self.ServeJSON()
	return
}

// @Title get
// @Description list tenant
// @Success 200 {string} get success!
// @Failure 404 tenant not Exist
// @router / [get]
func (self *TenantController) GetAll() {
	defer RecoverRsp500(&self.Controller)
	tenants := models.GetAllTenants()
	if tenants == nil {
		HandleErr(&self.Controller, models.BuildErrWithCode(404, errors.New("tenants not Exist")))
		return
	}
	tenantsRsp := make([]*TenantRsp, 0)
	for _, tenantInfo := range tenants {
		var status string = "ACTIVE"
		if tenantInfo.IsCancelling == true {
			status = "DELETING"
		}
		tenantRsp := &TenantRsp{
			CreatedAt: tenantInfo.CreateTime,
			ID:        tenantInfo.TenantUUID,
			Name:      tenantInfo.TenantName,
			NetNumber: tenantInfo.NetNum,
			IaasTenant: IaasTenantRsp{
				TenantName: tenantInfo.IaasTenantName,
				ID:         tenantInfo.IaasTenantID,
			},
			Quotas: QuotasRsp{
				Network: tenantInfo.Quota,
			},
			Status: status,
		}
		tenantsRsp = append(tenantsRsp, tenantRsp)
	}
	self.Data["json"] = PaasTenantsRsp{Tenants: tenantsRsp}
	self.ServeJSON()
}

type ExclusivePaasTenantReq struct {
	Name       string        `json:"name"`
	IaasTenant IaasTenantReq `json:"iaas_tenant"`
}

type IaasTenantReq struct {
	Endpoint   string  `json:"endpoint"`
	TenantName string  `json:"tenant_name"`
	Auth       AuthReq `json:"auth"`
}

type AuthReq struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type PaasTenantRsp struct {
	Tenant TenantRsp `json:"tenant"`
}

type TenantRsp struct {
	CreatedAt  string        `json:"created_at"`
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	NetNumber  int           `json:"net_number"`
	IaasTenant IaasTenantRsp `json:"iaas_tenant"`
	Quotas     QuotasRsp     `json:"quotas"`
	Status     string        `json:"status"`
}

type IaasTenantRsp struct {
	TenantName string `json:"tenant_name"`
	ID         string `json:"id"`
}

type QuotasRsp struct {
	Network int `json:"network"`
}

type PaasTenantsRsp struct {
	Tenants []*TenantRsp `json:"tenant"`
}

// @Title create exclusive tenant
// @Description create exclusive tenant with lan network
// @Success 200 {tenant} models.Network.Id
// @Failure 400 invalid request body
// @Failure 406 auth error
// @Failure 500 panic
// @Failure 409 already exist
// @router / [post]
func (self *TenantController) PostExclusive() {
	var err error
	defer RecoverRsp500(&self.Controller)
	klog.Infof("Start to Create exclusive tenant")
	body, _ := ioutil.ReadAll(self.Ctx.Input.Context.Request.Body)
	klog.Infof("Create exclusive tenant body is %v", string(body))
	tenantReq := &ExclusivePaasTenantReq{}
	err = json.Unmarshal([]byte(body), &tenantReq)
	if err != nil {
		klog.Errorf("Unmarshal err :%v, body is [%v]", err, string(body))
		HandleErr(&self.Controller,
			models.BuildErrWithCode(http.StatusBadRequest, errors.New("reqest body error")))
		return
	}
	err = tenantReq.Check()
	if err != nil {
		klog.Errorf("Reqest check err: %v", err)
		HandleErr(&self.Controller,
			models.BuildErrWithCode(http.StatusBadRequest, err))
		return
	}
	mExclusiveTenant := tenantReq.MakeModelsExclusiveTenant()
	err = mExclusiveTenant.Create()
	if err != nil {
		klog.Errorf("ExclusiveTenant.Create err: %v", err)
		HandleErr(&self.Controller, err)
		return
	}
	var status string = "ACTIVE"
	if mExclusiveTenant.IsCancelling == true {
		status = "DELETING"
	}
	self.Data["json"] = PaasTenantRsp{
		Tenant: TenantRsp{
			CreatedAt: mExclusiveTenant.CreateTime,
			ID:        mExclusiveTenant.PassID,
			Name:      mExclusiveTenant.PaasName,
			NetNumber: mExclusiveTenant.NetNum,
			IaasTenant: IaasTenantRsp{
				TenantName: mExclusiveTenant.IaasTenantName,
				ID:         mExclusiveTenant.IaasTenantID,
			},
			Quotas: QuotasRsp{
				Network: mExclusiveTenant.Quota,
			},
			Status: status,
		},
	}
	self.ServeJSON()
	return
}

func (self *ExclusivePaasTenantReq) Check() error {
	if self.Name == "" {
		return errors.New("name is null")
	}
	if self.IaasTenant.Endpoint == "" {
		return errors.New("endpoint is null")
	}
	if self.IaasTenant.TenantName == "" {
		return errors.New("tenant_name is null")
	}
	if self.IaasTenant.Auth.UserName == "" {
		return errors.New("user_name is null")
	}
	if self.IaasTenant.Auth.Password == "" {
		return errors.New("paasword is null")
	}
	return nil
}

func (self *ExclusivePaasTenantReq) MakeModelsExclusiveTenant() *models.ExclusiveTenant {
	return &models.ExclusiveTenant{
		PaasName:       self.Name,
		IaasEndpoint:   self.IaasTenant.Endpoint,
		IaasTenantName: self.IaasTenant.TenantName,
		IaasUserName:   self.IaasTenant.Auth.UserName,
		IaasPaasword:   self.IaasTenant.Auth.Password,
		PassID:         self.Name,
	}
}
