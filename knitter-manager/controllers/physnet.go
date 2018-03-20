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
	"errors"
	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"github.com/astaxie/beego"
	"io/ioutil"
	"net/http"
)

type Physnet struct {
	DefaultPhysnet string `json:"default_physnet"`
}

// Operations about tenant
type PhysnetController struct {
	beego.Controller
}

// @Title update
// @Description update default physnet
// @Success 200 {string} models.Physnet.DefaultPhysnet
// @Failure 403 invalid request body
// @Failure 406 update defaultpyhsnet error
// @router / [post]
func (self *PhysnetController) Update() {
	defer RecoverRsp500(&self.Controller)
	if iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	iaasType := iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID).GetType()
	if iaasType != "vNM" {
		HandleErr(&self.Controller, models.BuildErrWithCode(http.StatusNotImplemented, errors.New("iaas type is not vNM")))
		return
	}
	body, _ := ioutil.ReadAll(self.Ctx.Input.Context.Request.Body)
	physnetReqJSON, err := jason.NewObjectFromBytes(body)
	if err != nil {
		UnmarshalErr403(&self.Controller, err)
		return
	}
	defaultPhysnet, errDef := physnetReqJSON.GetString("default_physnet")
	if errDef != nil {
		klog.Errorf("controller Update Err: GetObject[default_physnet] Err: %v", errDef)
		HandleErr(&self.Controller, models.BuildErrWithCode(http.StatusBadRequest, errDef))
		return
	}
	errModels := models.UpdateDefaultPhysnet(defaultPhysnet)
	if errModels != nil {
		klog.Errorf("controller Update Err: UpdateDefaultPhysnet Err: %v", errModels)
		HandleErr(&self.Controller, errModels)
		return
	}
	self.Data["json"] = Physnet{DefaultPhysnet: defaultPhysnet}
	self.ServeJSON()
}

// @Title get
// @Description update default physnet
// @Failure 403 invalid request
// @Failure 406 get defaultpyhsnet error
// @router / [get]
func (self *PhysnetController) Get() {
	defer RecoverRsp500(&self.Controller)
	if iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	iaasType := iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID).GetType()
	if iaasType != "vNM" {
		UnmarshalErr403(&self.Controller, errors.New("iaas type is not vNM"))
		return
	}

	defaultPhysnet, errModels := iaas.GetDefaultPhysnet()
	if errModels != nil {
		klog.Errorf("controller Update Err: GetDefaultPhysnet Err: %v", errModels)
		HandleErr406(&self.Controller, errModels)
		return
	}

	self.Data["json"] = Physnet{DefaultPhysnet: defaultPhysnet}
	self.ServeJSON()
}
