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
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/astaxie/beego"
	"io/ioutil"
)

type PhyPortController struct {
	beego.Controller
}

// @router / [post]
func (o *PhyPortController) Post() {
	klog.Infof("@@PhyPortController: post start ,reqbody is [%v]", string(o.Ctx.Input.RequestBody))
	defer RecoverRsp500(&o.Controller)
	paasTenantID := o.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&o.Controller)
		return
	}
	body, _ := ioutil.ReadAll(o.Ctx.Input.Context.Request.Body)
	portInfo, err := models.CreateAndAttach(body, paasTenantID)
	if err != nil {
		klog.Errorf("@@PhyPortController:CreateAndAttach port to vm failed, error:%v", err)
		HandleErr(&o.Controller, err)
		return
	}
	klog.Infof("@@PhyPortController: CreateAndAttach req[%s] resp:port[%+v] to /vm SUCC", body, portInfo)
	o.Data["json"] = portInfo
	o.ServeJSON()
}

// @router /:vm_id/:port_id [delete]
func (o *PhyPortController) Delete() {
	defer RecoverRsp500(&o.Controller)
	paasTenantID := o.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&o.Controller)
		return
	}
	tranID := models.TranID(o.GetString("req_id"))
	vmID := o.Ctx.Input.Param(":vm_id")
	portID := o.Ctx.Input.Param(":port_id")
	klog.Infof("@@PhyPortController: DetachAndDelete: ReqID[", tranID,
		"]User[", paasTenantID, "]PortID[", portID, "]VmID[", vmID, "]")
	err := models.DetachAndDelete(tranID, vmID, portID, paasTenantID)
	if err != nil {
		return
	}
	o.Data["json"] = "detach success!"
	klog.Infof("@@PhyPortController: DetachAndDelete port[id: %s] to vm[id: %s] success, Response Code:200", portID, vmID)
	o.ServeJSON()
}
