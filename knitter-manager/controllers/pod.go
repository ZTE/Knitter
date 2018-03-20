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
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/astaxie/beego"
)

// Operations about Pod
type PodController struct {
	beego.Controller
}

// @Title Get
// @Description find pod by pod_name
// @Param	pod_name	path 	string	true		"the pod_name you want to get"
// @Success 200 {object} models.EncapPod
// @Failure 404 :pod_name is empty
// @router /:pod_name [get]
func (self *PodController) Get() {
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	name := self.Ctx.Input.Param(":pod_name")
	if name == "" {
		Err400(&self.Controller,
			errors.New("input-pod-name-error"))
		return
	}
	pod := models.Pod{Name: name}
	pod.SetTenantID(paasTenantID)
	p, err := pod.GetFromEtcd(paasTenantID, name)
	if err != nil {
		NotfoundErr404(&self.Controller, err)
		return
	}
	self.Data["json"] = models.EncapPod{Pod: p}
	self.ServeJSON()
}

// @Title GetAll
// @Description get all pods
// @Success 200 {object} models.EncapPods
// @router / [get]
func (self *PodController) GetAll() {
	klog.Warning("=========== url:", self.Ctx.Request.URL.Path, "  method:", self.Ctx.Request.Method, "   ===== begin ====")
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	pod := &models.Pod{}
	pod.SetTenantID(paasTenantID)
	pods := pod.ListAll()
	self.Data["json"] = models.EncapPods{Pods: pods}

	klog.Warning("=========== url:", self.Ctx.Request.URL.Path, "  method:", self.Ctx.Request.Method, "   ===== end ====")
	self.ServeJSON()
}
