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
	"github.com/antonholmquist/jason"
	"github.com/astaxie/beego"
	"io/ioutil"
)

// Operations about Master
type VniController struct {
	beego.Controller
}

// @Title get VLan ID of network for sr_iov in ITRAN2.0 site side
// @Description create port and return it to knitter
// @Param	body		body 	models.Master	true		"The master get VLan ID"
// @Success 200 {object} models.GetVLanInfoResp
// @Failure 403 body is empty
// @router /:network_id [get]
func (o *VniController) Get() {
	defer RecoverRsp500(&o.Controller)
	paasTenantID := o.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&o.Controller)
		return
	}

	networkID := o.Ctx.Input.Param(":network_id")
	tranID := models.TranID(o.GetString("req_id"))
	klog.Info("Get-VNI-REQ-from-Agent: ReqID[", tranID,
		"]User[", paasTenantID, "]NetworkID[", networkID, "]")

	vxlanInfo, err := models.GetNetworkVni(paasTenantID, networkID)
	if err != nil {
		klog.Warning("get Network info for network:",
			" %s failed, error: %v", networkID, err)
		HandleErr406(&o.Controller, err)
		return
	}

	klog.Info(" get Network info for network:",
		" %s vlanInfo:%s success", networkID, vxlanInfo)
	o.Data["json"] = vxlanInfo
	o.ServeJSON()
}

const DefaultNetworkName string = "net_api"

// Operations about Master
type PaasNetController struct {
	beego.Controller
}

// @Title get network attrs
// @Description create port and return it to knitter
// @Param       body            body    models.Master   true            "The master get VLan ID"
// @Success 200 {object} models.GetVLanInfoResp
// @Failure 403 body is empty
// @router / [post]
func (self *PaasNetController) Post() {
	defer RecoverRsp500(&self.Controller)
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	body, _ := ioutil.ReadAll(self.Ctx.Input.Context.Request.Body)
	klog.Info("Req Body:", string(body))
	networkNamesJSON, err := jason.NewObjectFromBytes(body)
	if err != nil {
		UnmarshalErr403(&self.Controller, err)
		return
	}

	userName := self.GetString(":user")
	tranID := models.TranID(self.GetString("req_id"))
	needProvider, err := self.GetBool("provider")
	if err != nil || needProvider != true {
		needProvider = false
	}

	var paasNetworks []*models.PaasNetwork
	networkNames, err := networkNamesJSON.GetStringArray("network_names")
	if err != nil || networkNames == nil {
		klog.Warningf("get Network name from request body error: %v", err)
		HandleErr406(&self.Controller, err)
		return
	}

	klog.Info("Get-Network-REQ-from-Agent: ReqID[", tranID,
		"]User[", userName, "]NetworkName[", networkNames, "]")
	for _, networkName := range networkNames {
		paasNetwork, err := models.GetNetworkInfo(userName, networkName, needProvider)
		if err != nil {
			klog.Warningf(" get Network info for network:",
				" %s failed, error: %v", networkName, err)
			HandleErr406(&self.Controller, err)
			return
		}
		paasNetworks = append(paasNetworks, paasNetwork)
	}

	self.Data["json"] = paasNetworks
	self.ServeJSON()
	klog.Info(" Get networks [", networkNames,
		"]'s attributes by network names success")
}

// @Title get network info by name
// @Description create port and return it to knitter
// @Success 200 {object} models.GetVLanInfoResp
// @Failure 406 process is error
// @router /:network_name [get]
func (o *PaasNetController) Get() {
	defer RecoverRsp500(&o.Controller)
	GetNetworkByName(&o.Controller)
}

func GetNetworkByName(o *beego.Controller) {
	isDefault, err := o.GetBool("default", true)
	paasTenantID := o.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(o)
		return
	}
	networkName := o.Ctx.Input.Param(":network_name")
	if isDefault {
		networkName = models.GetDefaultNetworkName()
		klog.Infof("GetNetworkByName: models.GetDefaultNetworkName() return: %s", networkName)
	}
	userName := o.GetString(":user")
	tranID := models.TranID(o.GetString("req_id"))
	klog.Info("Get-Network-REQ-from-Agent: ReqID[", tranID,
		"]User[", userName, "]NetworkName[", networkName, "]")

	network, err := models.GetNetworkInfo(userName, networkName, false)
	if err != nil {
		klog.Warning(" get Network info for network:",
			" %s failed, error: %v", networkName, err)
		HandleErr406(o, err)
		return
	}

	o.Data["json"] = network
	o.ServeJSON()
	klog.Info(" get default network:",
		" %s vlanInfo:%s success", network.ID, network.Name)
}
