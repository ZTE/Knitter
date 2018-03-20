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
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/uuid"
	"github.com/astaxie/beego"
	"io/ioutil"
	"net/http"
)

type IPGroupController struct {
	beego.Controller
}

type IPGroup struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	IPs   []IP   `json:"ips"`
	NetID string `json:"network_id"`
	Size  int    `json:"size"`
}

type IP struct {
	IPAddr string `json:"ip_addr"`
	Used   bool   `json:"used"`
}

type EncapIPGroup struct {
	IPGrp *IPGroup `json:"ipgroup"`
}

type EncapIPGroups struct {
	IPGrps []*IPGroup `json:"ipgroups"`
}

type EncapIPGroupReq struct {
	IPGrp *IPGroupReq `json:"ipgroup"`
}

type IPGroupReq struct {
	Name      string `json:"name"`
	IPs       string `json:"ips"`
	NetworkID string `json:"network_id"`
	Size      string `json:"size"`
}

// @Title create
// @Description create ip group
// @Param	body		body 	models.IpGroup	true		"configration for ip group"
// @Success 200 EncapIpGroup
// @Failure 404 invalid request body
// @Failure 500 create ip group error
// @router / [post]
func (self *IPGroupController) Post() {
	defer RecoverRsp500(&self.Controller)
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	body, _ := ioutil.ReadAll(self.Ctx.Input.Context.Request.Body)
	klog.Infof("enter in POSTIG body: [%v]", string(body))

	igReq, err := UnmarshalIPGroupReq(body)
	if err != nil {
		Err400(&self.Controller, err)
		return
	}

	if igReq.IPGrp != nil {
		ig := models.IPGroup{
			TenantID:  paasTenantID,
			NetworkID: igReq.IPGrp.NetworkID,
			Name:      igReq.IPGrp.Name,
			IPs:       igReq.IPGrp.IPs,
			ID:        uuid.NewUUID(),
			SizeStr:   igReq.IPGrp.Size,
		}
		igResp, err := ig.Create()
		if err != nil {
			HandleErr(&self.Controller, err)
			return
		}

		self.Data["json"] = EncapIPGroup{IPGrp: makeIPGrp(igResp)}
		self.ServeJSON()
		return
	}

	klog.Warning("IPGROUP-CREATE-REQ-UNKNOW:" + string(body))
	ErrorRequstRsp400(&self.Controller, string(body))
	return
}

// @Title GetAllIG
// @Description get all ip group
// @Success 200 {object} EncapIpGroups
// @router / [get]
func (self *IPGroupController) GetAll() {
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	network := self.GetString("network_id") // todo network_id? or network?
	klog.Infof("enter in GetAllIG network: [%v], user: [%v]", network, paasTenantID)

	ig := models.IPGroup{TenantID: paasTenantID, NetworkID: network}
	igs, err := ig.GetIGs()
	if err != nil {
		HandleErr(&self.Controller, err)
		return
	}

	self.Data["json"] = EncapIPGroups{IPGrps: makeIPGrps(igs)}
	self.ServeJSON()
	return
}

// @Title GetIG
// @Description find ip group by id
// @Param	id		path 	string	true		"the ip group id you want to get"
// @Success 200 {object} EncapIpGroup
// @Failure 404 : ip group Not Exist
// @router /:group [get]
func (self *IPGroupController) Get() {
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	group := self.GetString(":group")
	klog.Infof("enter in GetIG group: [%v], user: [%v]", group, paasTenantID)

	ig := models.IPGroup{TenantID: paasTenantID, ID: group}
	igResp, err := ig.Get()
	if err != nil {
		HandleErr(&self.Controller, err)
		return
	}

	self.Data["json"] = EncapIPGroup{IPGrp: makeIPGrp(igResp)}
	self.ServeJSON()
	return
}

// @Title PutIG
// @Description modify ip group by id
// @Param	id		path 	string	true		"the ip group id you want to modify"
// @Success 200 {object} EncapIpGroup
// @Failure 404 : ip group Not Exist
// @router /:group [put]
func (self *IPGroupController) Put() {
	defer RecoverRsp500(&self.Controller)
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}

	body, _ := ioutil.ReadAll(self.Ctx.Input.Context.Request.Body)
	klog.Infof("enter in PutIG body: [%v]", string(body))
	igReq, err := UnmarshalIPGroupReq(body)
	if err != nil {
		Err400(&self.Controller, err)
		return
	}

	if igReq.IPGrp != nil {
		ig := models.IPGroup{
			TenantID: paasTenantID,
			Name:     igReq.IPGrp.Name,
			IPs:      igReq.IPGrp.IPs,
			ID:       self.GetString(":group"),
			SizeStr:  igReq.IPGrp.Size,
		}
		igResp, err := ig.Update()
		if err != nil {
			HandleErr(&self.Controller, err)
			return
		}

		self.Data["json"] = EncapIPGroup{IPGrp: makeIPGrp(igResp)}
		self.ServeJSON()
		return
	}

	klog.Warning("IPGROUP-UPDATE-REQ-UNKNOW:" + string(body))
	ErrorRequstRsp400(&self.Controller, string(body))
	return
}

// @Title DeleteIG
// @Description delete ip group by id
// @Param	id		path 	string	true		"The ip group id you want to delete"
// @Success 200 {string} delete success!
// @Failure 404 ip group not Exist
// @router /:group [delete]
func (self *IPGroupController) Delete() {
	paasTenantID := self.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&self.Controller)
		return
	}
	group := self.GetString(":group")
	klog.Infof("enter in DeleteIG group: [%v], user: [%v]", group, paasTenantID)

	ig := models.IPGroup{TenantID: paasTenantID, ID: group}
	err := ig.Delete()
	if err != nil {
		HandleErr(&self.Controller, err)
		return
	}

	self.Redirect(self.Ctx.Request.URL.RequestURI(), http.StatusNoContent)
	self.ServeJSON()
	return
}

func makeIPGrps(igInModels []*models.IPGroupObject) []*IPGroup {
	igs := make([]*IPGroup, 0)
	for _, igInModel := range igInModels {
		ips := make([]IP, 0)
		for _, ip := range igInModel.IPs {
			ips = append(ips, IP{IPAddr: ip.IPAddr, Used: ip.Used})
		}

		igs = append(igs, &IPGroup{
			ID:    igInModel.ID,
			Name:  igInModel.Name,
			IPs:   ips,
			Size:  len(ips),
			NetID: igInModel.NetworkID,
		})
	}

	return igs
}

func makeIPGrp(igInModel *models.IPGroupObject) *IPGroup {
	ips := make([]IP, 0)
	for _, ip := range igInModel.IPs {
		ips = append(ips, IP{IPAddr: ip.IPAddr, Used: ip.Used})
	}

	return &IPGroup{
		ID:    igInModel.ID,
		Name:  igInModel.Name,
		IPs:   ips,
		Size:  len(ips),
		NetID: igInModel.NetworkID,
	}
}

var UnmarshalIPGroupReq = func(value []byte) (*EncapIPGroupReq, error) {
	var ig EncapIPGroupReq
	err := json.Unmarshal([]byte(value), &ig)
	if err != nil {
		klog.Errorf("UnmarshalIPGroupReq: json.Unmarshal(%s) FAILED, error: %v", string(value), err)
		return nil, errobj.ErrUnmarshalFailed
	}

	klog.Debugf("UnmarshalIPGroupReq: ig[%v] SUCC", ig)
	return &ig, nil
}
