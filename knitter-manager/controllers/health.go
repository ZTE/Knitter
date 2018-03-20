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
	"github.com/astaxie/beego"
	//"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/models"
)

// Operations about network
type HealthController struct {
	beego.Controller
}

// @Title Get
// @Description get health level
// @Success 200 {object} models.HealthObj
// @router / [get]
func (self *HealthController) Get() {
	//klog.Info("Request health level!")
	var obj = models.HealthObj{}
	obj.Level = obj.GetHealthLevel()
	obj.State = obj.GetHealthState()
	self.Data["json"] = obj
	self.ServeJSON()

	return
}

// Operations about network
type NetworkManagerController struct {
	beego.Controller
}

// @Title Get
// @Description get health level
// @Success 200 {object} models.HealthObj
// @router / [get]
func (self *NetworkManagerController) Get() {
	i := iaas.GetIaaS(constvalue.PaaSTenantAdminDefaultUUID)
	if i != nil {
		self.Data["json"] = i.GetType()
	} else {
		self.Data["json"] = ""
	}
	self.ServeJSON()
	return
}
