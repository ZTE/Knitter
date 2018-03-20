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
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/astaxie/beego"
	"io/ioutil"
)

// Operations about network
type RouterController struct {
	beego.Controller
}

// @Title create
// @Description create router
// @Param	body		body 	models.EncapRouter	true		"configration for router"
// @Success 200 {string} models.Router.Id
// @Failure 403 invalid request body
// @Failure 406 create router error
// @router / [post]
func (self *RouterController) Post() {
	defer RecoverRsp500(&self.Controller)
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	var encapRouter models.EncapRouter
	body, _ := ioutil.ReadAll(self.Ctx.Input.Context.Request.Body)
	klog.Info(string(body))
	err := json.Unmarshal(body, &encapRouter)
	if err != nil {
		UnmarshalErr403(&self.Controller, err)
		return
	}
	klog.Info(encapRouter.Router.Name)
	r := iaasaccessor.Router{}
	rt := &models.Rt{Router: &r}
	rt.TenantUUID = paasTenantID
	err = rt.Create(*encapRouter.Router)
	if err != nil {
		HandleErr406(&self.Controller, err)
		return
	}
	self.Data["json"] = models.EncapRouter{Router: rt.Router}
	self.ServeJSON()
}

// @Title update
// @Description update router
// @Param	router_id		path 	string	true		"the router_id you want to update"
// @Param	body		body 	models.EncapRouter	true		"new configration for router"
// @Success 200 {string} models.Router.Id
// @Failure 403 invalid request body
// @Failure 406 update router error
// @router /:router_id/ [put]
func (self *RouterController) Update() {
	defer RecoverRsp500(&self.Controller)
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	var encapRouter models.EncapRouter
	body, _ := ioutil.ReadAll(self.Ctx.Input.Context.Request.Body)
	id := self.Ctx.Input.Param(":router_id")
	err := json.Unmarshal(body, &encapRouter)
	if err != nil {
		UnmarshalErr403(&self.Controller, err)
		return
	}
	klog.Info(encapRouter.Router.Name)
	r := iaasaccessor.Router{Id: id}
	rt := &models.Rt{Router: &r}
	rt.TenantUUID = paasTenantID
	err = rt.Update(*encapRouter.Router)
	if err != nil {
		HandleErr406(&self.Controller, err)
		return
	}
	self.Data["json"] = models.EncapRouter{Router: rt.Router}
	self.ServeJSON()
}

// @Title Get
// @Description find router by router_id
// @Param	router_id		path 	string	true		"the router_id you want to get"
// @Success 200 {object} models.EncapRouter
// @Failure 404 : Router Not Exist
// @router /:router_id [get]
func (self *RouterController) Get() {
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	id := self.Ctx.Input.Param(":router_id")
	klog.Info("Request router_id: ", id)
	r := iaasaccessor.Router{Id: id}
	rt := &models.Rt{Router: &r}
	rt.TenantUUID = paasTenantID
	ob, err := rt.GetFromEtcd()
	if err != nil {
		NotfoundErr404(&self.Controller, err)
		return
	}
	self.Data["json"] = models.EncapRouter{Router: ob}
	self.ServeJSON()
}

// @Title GetAll
// @Description get all routers
// @Success 200 {object} models.EncapRouters
// @router / [get]
func (self *RouterController) GetAll() {
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	r := iaasaccessor.Router{}
	rt := &models.Rt{Router: &r}
	rt.TenantUUID = paasTenantID
	obs := rt.ListAll()
	self.Data["json"] = models.EncapRouters{Routers: obs}
	self.ServeJSON()
}

// @Title delete
// @Description delete router by router_id
// @Param	router_id		path 	string	true		"The router_id you want to delete"
// @Success 200 {string} delete success!
// @Failure 404 Router not Exist
// @router /:router_id [delete]
func (self *RouterController) Delete() {
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	id := self.Ctx.Input.Param(":router_id")
	klog.Info("Request router_id: ", id)
	r := iaasaccessor.Router{Id: id}
	rt := &models.Rt{Router: &r}
	rt.TenantUUID = paasTenantID
	err := rt.DelByID()
	if err != nil {
		NotfoundErr404(&self.Controller, err)
		return
	}
	self.Data["json"] = ""
	self.ServeJSON()
}

// @Title attach network
// @Description attach network to router
// @Param	router_id		path 	string	true		"the router_id you want to get"
// @Param	body		body 	models.EncapNetwork	true		"network want to attach"
// @Success 200 {string} OK
// @Failure 403 invalid request body
// @Failure 406 attach router error
// @router /:router_id/attach [put]
func (self *RouterController) Attach() {
	defer RecoverRsp500(&self.Controller)
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	body, _ := ioutil.ReadAll(self.Ctx.Input.Context.Request.Body)
	klog.Info(string(body))
	id := self.Ctx.Input.Param(":router_id")
	var encapNetwork EncapPaasNetwork
	err := json.Unmarshal(body, &encapNetwork)
	if err != nil {
		UnmarshalErr403(&self.Controller, err)
		return
	}
	klog.Info(encapNetwork.Network.Name)
	r := iaasaccessor.Router{Id: id}
	rt := &models.Rt{Router: &r}
	rt.TenantUUID = paasTenantID
	err = rt.Attach(encapNetwork.Network.ID)
	if err != nil {
		HandleErr406(&self.Controller, err)
		return
	}
	self.Data["json"] = map[string]string{
		"OK": id + "Attach " + encapNetwork.Network.ID}
	self.ServeJSON()
}

// @Title detach network
// @Description detach network to router
// @Param	router_id		path 	string	true		"the router_id you want to get"
// @Param	body		body 	models.EncapNetwork	true		"network want to detach"
// @Success 200 {string} OK
// @Failure 403 invalid request body
// @Failure 406 detach router error
// @router /:router_id/detach [put]
func (self *RouterController) Detach() {
	defer RecoverRsp500(&self.Controller)
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	body, _ := ioutil.ReadAll(self.Ctx.Input.Context.Request.Body)
	klog.Info(string(body))
	id := self.Ctx.Input.Param(":router_id")
	var encapNetwork EncapPaasNetwork
	err := json.Unmarshal(body, &encapNetwork)
	if err != nil {
		UnmarshalErr403(&self.Controller, err)
		return
	}
	klog.Info(encapNetwork.Network.Name)
	r := iaasaccessor.Router{Id: id}
	rt := &models.Rt{Router: &r}
	rt.TenantUUID = paasTenantID
	err = rt.Detach(encapNetwork.Network.ID)
	if err != nil {
		HandleErr406(&self.Controller, err)
		return
	}
	self.Data["json"] = map[string]string{
		"OK": id + " Detach " + encapNetwork.Network.ID}
	self.ServeJSON()
}
