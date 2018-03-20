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
	"fmt"
	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/pkg/adapter"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-iaas"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/astaxie/beego"
	"io/ioutil"
)

// Operations about Master
type CniMasterPortController struct {
	beego.Controller
}

// @Title create port
// @Description create port and return it to knitter
// @Param	body		body 	models.CreatePortReq	true		"The master create port"
// @Success 200 {object} models.CreatePortResp
// @Failure 403 params is not enough
// @Failure 406 create port failed
// @router / [post]
func (c *CniMasterPortController) Post() {
	klog.Infof("@@@Create logical port START")
	defer klog.Infof("@@@Create logical port END")

	defer RecoverRsp500(&c.Controller)
	paasTenantID := c.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&c.Controller)
		return
	}
	body, _ := ioutil.ReadAll(c.Ctx.Input.Context.Request.Body)
	tranID := models.TranID(c.GetString("req_id"))
	if isBulk(body) {
		klog.Infof("TranID[createBulkPorts] ---> [%v]", tranID)
		req, err1 := buildReqObj(body, tranID, paasTenantID)
		if err1 != nil {
			klog.Errorf("Unmarshall from http request body failed, error: %v", err1)
			UnmarshalErr403(&c.Controller, err1)
			return
		}
		resp, err2 := models.CreateBulkPorts(req)
		if err2 != nil {
			klog.Errorf("CreateBulkPorts failed, error: %v", err2)
			UnmarshalErr403(&c.Controller, err2)
			return
		}
		klog.Infof("Create bulk ports success Ports info: %v", resp)
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}

	CreatePortRequestProcess(c, tranID, body)
}

func CreatePortProcess(pc *CniMasterPortController, reqObj models.CreatePortReq) {
	klog.Infof(" create port for sr_iov: networkName: %s", reqObj.NetworkName)

	portInfo, err := models.CreateLogicalPort(&reqObj)
	if err != nil {
		klog.Infof(" Create port for sr_iov in network[%s] vnic_type[%s] failed, error: %v, Response 406",
			reqObj.NetworkName, reqObj.VnicType, err)
		HandleErr406(&pc.Controller, err)
		return
	}
	pc.Data["json"] = portInfo
	pc.ServeJSON()
}

func isBulk(body []byte) bool {
	bodyJSON, err := adapter.NewObjectFromBytes(body)
	if err != nil {
		return false
	}
	fmt.Println(bodyJSON)
	_, err = bodyJSON.GetObjectArray("ports")
	if err == nil {
		return true
	}
	return false
}

func buildReqObj(body []byte, tranID models.TranID, tenantID string) (*mgriaas.MgrBulkPortsReq, error) {
	reqObj := agtmgr.AgtBulkPortsReq{}
	klog.Infof("createBulkPorts http request body is: %s", string(body))
	err := adapter.Unmarshal(body, &reqObj)
	if err != nil {
		klog.Errorf("Unmarshall from http request body failed, error: %v", err)
		return nil, err
	}
	mgrBulkPortsReq := mgriaas.MgrBulkPortsReq{Ports: make([]*mgriaas.MgrPortReq, 0)}
	mgrBulkPortsReq.TranId = string(tranID)
	for _, port := range reqObj.Ports {
		req := &mgriaas.MgrPortReq{}
		req.AgtPortReq = port
		req.TenantId = tenantID
		mgrBulkPortsReq.Ports = append(mgrBulkPortsReq.Ports, req)
	}
	return &mgrBulkPortsReq, nil
}

func CreatePortRequestProcess(pc *CniMasterPortController, tranID models.TranID, body []byte) {
	// get network_name and vnic_type for eio
	var reqObj models.CreatePortReq

	klog.Infof("raw http request body is: %v", string(body))
	err := json.Unmarshal(body, &reqObj)
	if err != nil {
		klog.Errorf("Unmarshall from http request body failed, error: %v", err)
		UnmarshalErr403(&pc.Controller, err)
		return
	}

	reqObj.TenantID = pc.GetString(":user")
	if reqObj.VnicType == "" {
		// logical port default vnic type: "normal"
		reqObj.VnicType = constvalue.LogicalPortDefaultVnicType
	}
	klog.Info("Create-PORT-REQ-from-Agent: ReqID[", tranID,
		"]Ten[", reqObj.TenantID, "]Network[", reqObj.NetworkName,
		"]PortName[", reqObj.PortName, "]GatewayType[", reqObj.VnicType)
	CreatePortProcess(pc, reqObj)
}

// @Title delete
// @Description detach port from the given VM and delete it
// @Param	port_id		path 	string	true		"The objectId you want to delete"
// @Success 200 {string} delete success!
// @Failure 404 delete failed
// @router /:port_id [delete]
func (o *CniMasterPortController) Delete() {
	klog.Infof("@@@Delete logical port START")
	defer klog.Infof("@@@Delete logical port END")
	var err error

	defer RecoverRsp500(&o.Controller)
	paasTenantID := o.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&o.Controller)
		return
	}

	tranID := models.TranID(o.GetString("req_id"))
	portID := o.Ctx.Input.Param(":port_id")

	klog.Info("Delete-PORT-REQ-from-Agent: ReqID[", tranID,
		"]User[", paasTenantID, "]PortID[", portID)

	err = models.DestroyLogicalPort(portID)
	if err != nil {
		klog.Errorf("Delete port[id: %s] failed, error: %v", portID, err)
		Err500(&o.Controller, err)
		return
	}

	o.Data["json"] = "delete success!"
	klog.Infof("Delete port[id: %s] for sr_iov succeed, Response Code:200", portID)
	o.ServeJSON()
}

// @Title Attach port to virtual machine for virtualized scene
// @Description attach port and return it to knitter
// @Param	port_id		string 	models.AttachPortReq	true		"The port_id attaching to vm"
// @Success 200 {string} attach success!
// @Failure 406 create port failed
// @router /:vm_id/:port_id [post]
func (o *CniMasterPortController) Attach() {
	klog.Infof("@@@Attach logical port START")
	defer klog.Infof("@@@Attach logical port END")
	defer RecoverRsp500(&o.Controller)
	paasTenantID := o.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&o.Controller)
		return
	}

	tranID := models.TranID(o.GetString("req_id"))
	vmID := o.Ctx.Input.Param(":vm_id")
	portID := o.Ctx.Input.Param(":port_id")
	klog.Info("Attach-PORT-REQ-from-Agent: ReqID[", tranID,
		"]User[", paasTenantID, "]PortID[", portID, "]VmID[", vmID, "]")

	//vmID equals vmName for VNFM
	err := models.GetPortServiceObj().AttachPortToVM(tranID, paasTenantID, &models.PortVMOpsReq{VMID: vmID, PortID: portID})
	if err != nil {
		klog.Infof(" Attach port[id: %s] to vm[id: %s] failed, error: %v, Response Code:406", portID, vmID, err)
		HandleErr406(&o.Controller, err)
		return
	}
	o.Data["json"] = "attach success!"
	klog.Infof(" Attach port[id: %s] to vm[id: %s] success, Response Code:200", portID)
	o.ServeJSON()
}

// @Title delete
// @Description detach port from the given VM
// @Param	port_id		path 	string	true		"The port_id you want to detach"
// @Success 200 {string} detach success!
// @Failure 406 operation failed
// @router /:vm_id/:port_id [delete]
func (o *CniMasterPortController) Detach() {
	klog.Infof("@@@Detach logical port START")
	defer klog.Infof("@@@Detach logical port END")
	defer RecoverRsp500(&o.Controller)
	paasTenantID := o.GetString(":user")
	if iaas.GetIaaS(paasTenantID) == nil {
		RecoverRsp401(&o.Controller)
		return
	}

	tranID := models.TranID(o.GetString("req_id"))
	vmID := o.Ctx.Input.Param(":vm_id")
	portID := o.Ctx.Input.Param(":port_id")
	klog.Info("Detach-PORT-REQ-from-Agent: ReqID[", tranID,
		"]User[", paasTenantID, "]PortID[", portID, "]VmID[", vmID, "]")

	err := models.GetPortServiceObj().DetachPortFromVM(tranID, paasTenantID, &models.PortVMOpsReq{VMID: vmID, PortID: portID})
	if err != nil {
		klog.Infof(" Detach port[id: %s] from vm[id: %s] failed, error: %v, Response Code:406", portID, vmID, err)
		HandleErr406(&o.Controller, err)
		return
	}
	o.Data["json"] = "detach success!"
	klog.Infof(" Detach port[id: %s] to vm[id: %s] success, Response Code:200", portID)
	o.ServeJSON()
}
