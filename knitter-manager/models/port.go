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

package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
	"k8s.io/client-go/tools/cache"

	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-agt"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-iaas"
	"github.com/ZTE/Knitter/pkg/klog"
)

const (
	PortStatusActive = "ACTIVE"
	portStatusDown   = "DOWN"
	//fixedIPKey       = "fixedIP"
	noFixedIPKey = "noFixedIP"
	IPGroupKey   = "IPGroup"
)

const (
	OpenStackOpsRetryTime      = 3
	OpenStackOpsRetryIntval    = 3
	CheckPortStatusIntval      = 5
	CheckPortStatusNerrumber   = 60
	CheckPortStatusIntvalShort = 3
	CheckPortStatusNumberShort = 3
)

const (
	PortTypeAttatched    = "AttatchedPort"
	PortTypeNonattatched = "NonattatchedPort"

	PortDetachOp = "NovaDetachPort"
	PortDeleteOp = "NeutronDeletePort"
)

type ExceptPort struct {
	ID string `json:"id"`

	PortType string `json:"port_type"`
	HostID   string `json:"host_id"`

	ExceptTime   string `json:"except_time"`
	ExceptReason string `json:"except_reason"`
	ExceptStat   string `json:"except_stat"`

	ExpectProcOps []string `json:"expect_proc_ops"`
}

func SaveExceptPort(portID, portType, hostID, exceptReason string) error {
	klog.Infof("SaveExceptPort: start to save port[id: %s] to exceptional ports repo", portID)
	port := ExceptPort{ID: portID, PortType: portType, ExceptReason: exceptReason}
	if portType == PortTypeAttatched {
		port.HostID = hostID
		port.ExpectProcOps = []string{PortDetachOp, PortDeleteOp}
	} else {
		port.ExpectProcOps = []string{PortDeleteOp}
	}
	port.ExceptTime = time.Now().UTC().Format("2006-01-02T15:04:05Z")

	url := dbaccessor.GetKeyOfExceptionalPort(port.ID)
	value, err := json.Marshal(port)
	if err != nil {
		klog.Errorf("SaveExceptPort: marshal exceptional port[%v] failed, error:%v", port, err)
		return fmt.Errorf("%v:marshal exceptional port failed", err)
	}

	err = common.GetDataBase().SaveLeaf(url, string(value))
	if err != nil {
		klog.Errorf("SaveExceptPort: save exceptional port[%v] to DB failed, error:%v", port, err)
		return fmt.Errorf("%v:save exceptional port to DB failed", err)
	}

	klog.Infof("SaveExceptPort: finish to save port[id: %s] to exceptional ports repo", portID)
	return nil
}

type TranID string

/* interface between knitter_master and knitter type define start */
type CreatePortReq struct {
	agtmgr.AgtPortReq
}

type PortVMOpsReq struct {
	VMID   string
	PortID string
}

/* port operations abstract */
type PortOpsAPI interface {
	CreatePort(tranID TranID, iaasObj iaasaccessor.IaaS, reqObj *CreatePortReq) (*iaasaccessor.Interface, error)
	CreateBulkPorts(mgrBulkPortsReq *mgriaas.MgrBulkPortsReq) ([]*iaasaccessor.Interface, error)
	DeletePort(iaasObj iaasaccessor.IaaS, tranID TranID, portID string) error

	AttachPortToVM(tranID TranID, iaasObj iaasaccessor.IaaS, reqObj *PortVMOpsReq) (*iaasaccessor.Interface, error)
	DetachPortFromVM(tranID TranID, iaasObj iaasaccessor.IaaS, reqObj *PortVMOpsReq) error

	CheckNicStatus(tranID TranID, status string, checkNum, intval int, iaasObj iaasaccessor.IaaS, portID string) error
}

type PortOps struct{}

func (self *PortOps) CreateBulkPorts(mgrBulkPortsReq *mgriaas.MgrBulkPortsReq) ([]*iaasaccessor.Interface, error) {
	var err error
	var intersNotFixIP []*iaasaccessor.Interface = nil

	intersNotFixIP, err = createIaasBulkPorts(mgrBulkPortsReq)
	if err != nil {
		klog.Errorf("CreateBulkPorts createIaasBulkPorts failed. error: [%v]", err.Error())
		return nil, err
	}
	return intersNotFixIP, nil
}

var createBulkPortsFromIPGroup = func(req *mgriaas.MgrBulkPortsReq) ([]*iaasaccessor.Interface, error) {
	if len(req.Ports) == 0 {
		return nil, nil
	}
	var err error
	portsFromIPGroup := make([]*iaasaccessor.Interface, 0)
	portObjs := make([]*PortObj, 0)
	defer func() {
		if err != nil {
			for _, deletPort := range portObjs {
				err := destoryIPGroupPort(deletPort)
				if err != nil {
					klog.Warningf("createBulkPortsFromIPGroup:destoryIPGroupPort roll back err ,error is [%v]", err)
				}
			}
		}
	}()
	for _, port := range req.Ports {
		ig := IPGroup{TenantID: port.TenantId, NetworkID: port.NetworkId, ID: port.IPGroupId, Name: port.IPGroupName}
		newPort, err := ig.ObtainIP()
		if err != nil {
			klog.Errorf("CreateBulkPortsFromIPGroup ObtainIP failed. error: [%v]", err.Error())
			return portsFromIPGroup, err
		}

		newPort.Name = port.PortName
		newPort.IPGroupID = port.IPGroupId
		portsFromIPGroup = append(portsFromIPGroup, newPort)
		portObj := &PortObj{
			ID:        newPort.Id,
			TenantID:  port.TenantID,
			IPGroupID: port.IPGroupId,
			NetworkID: port.NetworkId,
		}
		portObjs = append(portObjs, portObj)
	}
	return portsFromIPGroup, nil
}

var createPortFromIPGroup = func(reqObj *CreatePortReq, networkID string) (*iaasaccessor.Interface, error) {
	ig := IPGroup{TenantID: reqObj.TenantID, NetworkID: networkID, Name: reqObj.IPGroupName}
	igInModel, err := ig.GetIPGroupByName()
	if err != nil {
		klog.Errorf("GetIPGroupByName of "+
			"group[name: %v] failed, error: [%v]", reqObj.IPGroupName, err)
		return nil, err
	}

	mgrBulkPortsReqFromIPGroup := &mgriaas.MgrBulkPortsReq{}
	mgrPortReq := &mgriaas.MgrPortReq{}
	mgrPortReq.IPGroupId = igInModel.ID
	mgrPortReq.TenantId = reqObj.TenantID
	mgrPortReq.NetworkId = networkID
	mgrPortReq.IPGroupName = igInModel.Name
	mgrBulkPortsReqFromIPGroup.Ports = append(mgrBulkPortsReqFromIPGroup.Ports, mgrPortReq)
	intersFromIPGroup, err := createBulkPortsFromIPGroup(mgrBulkPortsReqFromIPGroup)
	if err != nil {
		klog.Errorf("CreatePort createBulkPortsFromIPGroup failed. error: [%v]", err.Error())
		return nil, err
	}

	intersFromIPGroup[0].Name = reqObj.PortName
	klog.Infof(" createPort: CreatePort OK, port detail is :%v", *(intersFromIPGroup[0]))
	return intersFromIPGroup[0], nil
}

func (self *PortOps) CreatePort(tranID TranID, iaasObj iaasaccessor.IaaS, reqObj *CreatePortReq) (*iaasaccessor.Interface, error) {
	paasNetworkInfo, err := GetNetworkByName(reqObj.TenantID, reqObj.NetworkName)
	if err != nil {
		klog.Errorf(" GetID of network[id: %s] failed, error: %v", reqObj.NetworkName, err)
		return nil, fmt.Errorf("%v:Get Network ID error", err)
	}
	networkID := paasNetworkInfo.ID
	subnetID := paasNetworkInfo.SubnetID
	iaasTenantID, err := iaas.GetIaasTenantIDByPaasTenantID(reqObj.TenantID)
	if err != nil {
		return nil, err
	}
	klog.Infof(" assembleResponseWithTranId: get network id: %s subid: %s", networkID, subnetID)
	portName := iaas.GetExternalPortName(iaasTenantID, common.GetPaaSID(),
		reqObj.NodeID, reqObj.PodName, reqObj.PortName)

	if reqObj.IPGroupName != "" {
		return createPortFromIPGroup(reqObj, networkID)
	}

	port, err := iaasObj.CreatePort(networkID, subnetID, portName, "", "", reqObj.VnicType)
	if err != nil {
		klog.Errorf(" CreatePort failed, error: %v", err)
		return nil, fmt.Errorf("%v:CreatePort error", err)
	}
	klog.Infof(" createPort: CreatePort OK, port detail is :%v", port)

	return port, nil
}

func (self *PortOps) DeletePort(iaasObj iaasaccessor.IaaS, tranID TranID, portID string) error {
	err := iaasObj.DeletePort(portID)
	if err != nil {
		klog.Errorf(" DeletePortWithTranId: delete port[id: %s] failed, error: %v", portID, err)
		return fmt.Errorf("%v:Delete port error", err)
	}

	return nil
}

func (self *PortOps) AttachPortToVM(tranID TranID, iaasObj iaasaccessor.IaaS, reqObj *PortVMOpsReq) (*iaasaccessor.Interface, error) {
	var err error

	klog.Infof(" AttachPortToVM: start to attach port[id:%s] to vm[id:%s]", reqObj.PortID, reqObj.VMID)
	nic, err := iaasObj.AttachPortToVM(reqObj.VMID, reqObj.PortID)
	if err != nil {
		klog.Errorf(" AttachPortToVM: AttachPort(vmId[%s], portinfo.ID[%s]) error: %v", reqObj.VMID, reqObj.PortID, err)
		return nil, fmt.Errorf("%v:Attach port to vm error", err)
	}

	return nic, nil
}

func (self *PortOps) DetachPortFromVM(tranID TranID, iaasObj iaasaccessor.IaaS, reqObj *PortVMOpsReq) error {
	klog.Infof(" DetachPortFromVM: start to detach port[id:%s] from vm[id:%s]", reqObj.PortID, reqObj.VMID)
	err := iaasObj.DetachPortFromVM(reqObj.VMID, reqObj.PortID)
	if err != nil {
		klog.Errorf(" DetachPortFromVM: DetachPort port[id: %s] from vm[id: %s] failed, error: %v",
			reqObj.PortID, reqObj.VMID, err)
		return fmt.Errorf("%v:detach port from vm error", err)
	}
	klog.Infof(" DetachPortFromVM: detach port vmId: %s, portId: %s. detach port from vm succeed",
		reqObj.VMID, reqObj.PortID)

	return nil
}

func (self *PortOps) CheckNicStatus(tranID TranID, status string, checkNum, intval int, iaasObj iaasaccessor.IaaS, portID string) error {
	for i := 0; i < checkNum; i++ {
		portInfo, err := iaasObj.GetPort(portID)
		if err != nil {
			klog.Errorf(" CheckNicStatus: get port[id: %s] failed, error: %v", portID, err)
			return fmt.Errorf("%v:get port error", err)
		}

		if portInfo.Status == status {
			klog.Infof(" CheckNicStatus: wait port[id: %s] become %s status, just go forward", status, portID)
			return nil
		}
		klog.Infof(" CheckNicStatus: port[id: %s] wait for %s status %d times, about %d seconds",
			portID, status, i, i*intval)
		klog.Infof(" CheckNicStatus: port[id: %s] get port info: %v", portInfo.Id, portInfo)
		if i+1 < checkNum {
			time.Sleep(time.Duration(intval) * time.Second)
		}
	}

	klog.Errorf(" CheckNicStatus: check port[id: %s] status %s timeout error", status, portID)
	return nil
}

type PortServiceAPI interface {
	CreatePort(tranID TranID, reqObj *CreatePortReq) (*PortObj, *mgragt.CreatePortResp, error)
	CreateBulkPorts(mgrBulkPortsReq *mgriaas.MgrBulkPortsReq) ([]*iaasaccessor.Interface, error)
	DeletePort(tranID TranID, portID, tenantID string) error

	AttachPortToVM(tranID TranID, tenantID string, reqObj *PortVMOpsReq) error
	DetachPortFromVM(tranID TranID, tenantID string, reqObj *PortVMOpsReq) error
}

type PortService struct{}

var PortServiceObj PortServiceAPI = &PortService{}

var GetPortServiceObj = func() PortServiceAPI {
	return PortServiceObj
}

func (self *PortService) CreateBulkPorts(mgrBulkPortsReq *mgriaas.MgrBulkPortsReq) ([]*iaasaccessor.Interface, error) {
	var portOpsObj PortOpsAPI = &PortOps{}
	nics, err := portOpsObj.CreateBulkPorts(mgrBulkPortsReq)
	if err != nil {
		klog.Errorf("CreateBulkPorts: PortOps CreateBulkPorts FAILED, error: [%v]", err.Error())
		return nil, err
	}

	return nics, nil
}

var fillNetworkInfoToReq = func(mgrBulkPortsReq *mgriaas.MgrBulkPortsReq) error {
	for _, mgrBulkPortReq := range mgrBulkPortsReq.Ports {
		network := Net{TenantUUID: mgrBulkPortReq.TenantId,
			Network: iaasaccessor.Network{Name: mgrBulkPortReq.NetworkName}}
		paasNetworkInfo, err := GetNetwork(&network)
		if err != nil {
			klog.Errorf("GetID of network[id: %v] failed, error: [%v]", mgrBulkPortReq.NetworkName, err)
			return fmt.Errorf("%v:Get Network ID error", err)
		}
		mgrBulkPortReq.NetworkId = paasNetworkInfo.ID
		mgrBulkPortReq.SubnetId = paasNetworkInfo.SubnetID
		if mgrBulkPortReq.SubnetId == "" {
			subnetID, err := iaas.GetIaaS(mgrBulkPortReq.TenantID).GetSubnetID(mgrBulkPortReq.NetworkId)
			if err != nil {
				klog.Errorf("GetSubnetID of network[id: %v] failed, error: [%v]", mgrBulkPortReq.NetworkId, err)
				return fmt.Errorf("%v:GetSubnetID error", err)
			}
			klog.Infof("createPort: get network[id: %v] subnet id: %s", mgrBulkPortReq.NetworkId, subnetID)
			mgrBulkPortReq.SubnetId = subnetID
		}

		if mgrBulkPortReq.IPGroupName != "" {
			ig := IPGroup{TenantID: mgrBulkPortReq.TenantId, NetworkID: mgrBulkPortReq.NetworkId, Name: mgrBulkPortReq.IPGroupName}
			igInModel, err := ig.GetIPGroupByName()
			if err != nil {
				klog.Errorf("GetIPGroupByName of group[name: %v] failed, error: [%v]", mgrBulkPortReq.IPGroupName, err)
				return fmt.Errorf("%v:GetIPGroupByName error", err)
			}

			mgrBulkPortReq.IPGroupId = igInModel.ID
		}
	}

	return nil
}

var buildBulkPortsReq = func(mgrBulkPortsReq *mgriaas.MgrBulkPortsReq) (map[string]*mgriaas.MgrBulkPortsReq, error) {
	err := fillNetworkInfoToReq(mgrBulkPortsReq)
	if err != nil {
		klog.Errorf("fillNetworkInfoToReq failed, error: [%v]", err.Error())
		return nil, err
	}

	mgrBulkPortsReqNotFixIP := &mgriaas.MgrBulkPortsReq{
		TranId: "",
		Ports:  []*mgriaas.MgrPortReq{},
	}
	mgrBulkPortsReqFromIPGroup := &mgriaas.MgrBulkPortsReq{
		TranId: "",
		Ports:  []*mgriaas.MgrPortReq{},
	}

	mgrBulkPortsReqNotFixIP.TranId = mgrBulkPortsReq.TranId
	mgrBulkPortsReqFromIPGroup.TranId = mgrBulkPortsReq.TranId
	//iaasID := common.GetIaaSTenants()
	paasID := common.GetPaaSID()
	for _, mgrBulkPortReqPortsTmp := range mgrBulkPortsReq.Ports {
		if mgrBulkPortReqPortsTmp.IPGroupId != "" {
			mgrBulkPortsReqFromIPGroup.Ports = append(mgrBulkPortsReqFromIPGroup.Ports, mgrBulkPortReqPortsTmp)
			continue
		}
		iaasTenantID, err := iaas.GetIaasTenantIDByPaasTenantID(mgrBulkPortReqPortsTmp.TenantID)
		if err != nil {
			return nil, err
		}
		mgrBulkPortReqPortsTmp.PortName = iaas.GetExternalPortName(iaasTenantID, paasID,
			mgrBulkPortReqPortsTmp.NodeID, mgrBulkPortReqPortsTmp.PodName,
			mgrBulkPortReqPortsTmp.PortName)
		mgrBulkPortsReqNotFixIP.Ports = append(mgrBulkPortsReqNotFixIP.Ports,
			mgrBulkPortReqPortsTmp)
	}

	portsMap := make(map[string]*mgriaas.MgrBulkPortsReq, 3)
	portsMap[noFixedIPKey] = mgrBulkPortsReqNotFixIP
	portsMap[IPGroupKey] = mgrBulkPortsReqFromIPGroup
	return portsMap, nil
}

// create port entry function
func (self *PortService) CreatePort(tranID TranID, reqObj *CreatePortReq) (*PortObj, *mgragt.CreatePortResp, error) {
	var portOpsObj PortOpsAPI = &PortOps{}
	var portInfo *iaasaccessor.Interface
	var err error
	for idx := 0; idx < OpenStackOpsRetryTime; idx++ {
		portInfo, err = portOpsObj.CreatePort(tranID, iaas.GetIaaS(reqObj.TenantID), reqObj)
		if err == nil {
			break
		} else if idx+1 == OpenStackOpsRetryTime {
			klog.Errorf("CreatePort: createPort in networkName:%s vnicType:%s error: %v, retry %d times timeout",
				reqObj.NetworkName, reqObj.VnicType, err, OpenStackOpsRetryTime)
			return nil, nil, errors.New("createPort failed: " + err.Error())
		}
		klog.Errorf("CreatePort: createPort in networkName:%s vnicType:%s failed, error: %v, just retry %d time",
			reqObj.NetworkName, reqObj.VnicType, err, idx+1)
		time.Sleep(OpenStackOpsRetryIntval * time.Second)
	}
	klog.Infof("CreatePort: create port call success", tranID)

	defer func() {
		if err != nil {
			delErr := portOpsObj.DeletePort(iaas.GetIaaS(reqObj.TenantID), tranID, portInfo.Id)
			if delErr != nil {
				klog.Errorf("CreatePort: delete port[id: %s] failed, error: %v", portInfo.Id, err)
				SaveExceptPort(portInfo.Id, PortTypeNonattatched, "", "delete port failed")
			}
		}
	}()

	respInfo, err := AssembleResponse(iaas.GetIaaS(reqObj.TenantID), portInfo)
	if err != nil {
		klog.Errorf("CreatePort: assembleResponseWithTranId failed, error: %v", err)
		return nil, nil, fmt.Errorf("%v:assembleResponseWithTranId failed", err)
	}

	portObj := MakePortObj(portInfo, reqObj)
	klog.Tracef("CreatePort: Create port[%v] SUCC", portObj)
	return portObj, respInfo, nil
}

var AssembleResponse = func(iaasObj iaasaccessor.IaaS, port *iaasaccessor.Interface) (*mgragt.CreatePortResp, error) {
	subnet, err := getSubnetInfo(iaasObj, port.SubnetId)
	if err != nil {
		klog.Errorf("AssembleResponse: getSubnetInfo[subnetID: %s] failed, error: %v", port.SubnetId, err)
		delPortErr := iaasObj.DeletePort(port.Id)
		if delPortErr != nil {
			klog.Errorf(" assembleResponse: recyle port after attach->getport->failed failed, error: %v", delPortErr)
			klog.Errorf(" assembleResponse: EXCEPT-MARK->PORT[id:%s]", port.Id)
			// save exceptional port
			SaveExceptPort(port.Id, PortTypeNonattatched, "", "create port failed, then delete it failed")
		}
		return nil, errors.New("AssembleResponse: GetSubnet error")
	}
	createPortRsp, err := makeResponse(port, subnet)
	if err != nil {
		klog.Errorf("AssembleResponse: makeResponse(port:[%v], subnet[%v]) failed, error: %v", port, subnet, err)
		delPortErr := iaasObj.DeletePort(port.Id)
		if delPortErr != nil {
			klog.Errorf(" assembleResponse: recyle port after attach->getport->failed failed, error: %v", delPortErr)
			klog.Errorf(" assembleResponse: EXCEPT-MARK->PORT[id:%s]", port.Id)
			// save exceptional port
			SaveExceptPort(port.Id, PortTypeNonattatched, "", "create port failed, then delete it failed")
		}
		return nil, errors.New("AssembleResponse: makeResponse error")
	}
	return createPortRsp, nil
}

func getSubnetInfo(iaasObj iaasaccessor.IaaS, subnetID string) (*iaasaccessor.Subnet, error) {
	klog.Infof(" assembleResponseWithTranId: "+
		"get subnet id: %s", subnetID)
	subnet, err := iaasObj.GetSubnet(subnetID)
	if err != nil {
		klog.Errorf(" assembleResponseWithTranId: GetSubnet[id: %s] failed, error: %v", subnetID, err)
		return nil, fmt.Errorf("%v:GetSubnet error", err)
	}

	return subnet, nil
}

func makeResponse(port *iaasaccessor.Interface, subnet *iaasaccessor.Subnet) (*mgragt.CreatePortResp, error) {
	//fill response info to knitter
	klog.Infof("makeResponse:port:%s\n", port)
	var portInfo mgragt.CreatePortInfo

	portInfo.Cidr = subnet.Cidr
	portInfo.GatewayIP = subnet.GatewayIp
	portInfo.MacAddress = port.MacAddress
	portInfo.Name = iaas.GetOriginalPortName(port.Name)
	portInfo.NetworkID = port.NetworkId
	portInfo.PortID = port.Id

	// let rspInfo.SecurityGroups be blank
	var fixIP = ports.IP{SubnetID: subnet.Id, IPAddress: port.Ip}
	portInfo.FixedIps = append(portInfo.FixedIps, fixIP)
	klog.Infof("portInfo:%s\n", portInfo)
	return &mgragt.CreatePortResp{Port: portInfo}, nil
}

func (self *PortService) DeletePort(tranID TranID, portID, tenantID string) error {
	klog.Infof(" DeletePort: delete port[id: %s] start", portID)
	var err error
	var portOpsObj PortOpsAPI = &PortOps{}
	for idx := 0; idx < OpenStackOpsRetryTime; idx++ {
		err = portOpsObj.DeletePort(iaas.GetIaaS(tenantID), tranID, portID)
		if err == nil {
			break
		} else if idx+1 == OpenStackOpsRetryTime {
			klog.Errorf(" DeletePort: delete port[id: %s] failed, error: %v, retry %d times timeout, exit",
				portID, err, OpenStackOpsRetryTime)
			// save exceptional port
			SaveExceptPort(portID, PortTypeNonattatched, "", "delete port failed")
			return err
		}
		time.Sleep(OpenStackOpsRetryIntval * time.Second)
		klog.Errorf(" DeletePort: delete port[id: %s] failed, error: %v. just retry %d time", portID, err, idx+1)
	}

	klog.Infof("DeletePort: DeleteLogicalPort(portID: %s) SUCC", portID)
	return nil
}

func (self *PortService) AttachPortToVM(tranID TranID, tenantID string, reqObj *PortVMOpsReq) error {
	klog.Infof(" AttachPortToVM: attach port[id: %s] to vm[id: %s] start", reqObj.PortID, reqObj.VMID)
	var portOpsObj PortOpsAPI = &PortOps{}
	_, err := portOpsObj.AttachPortToVM(tranID, iaas.GetIaaS(tenantID), reqObj)
	if err != nil {
		// attach to vm failed does not mean the port has not been attached to vm actually,
		// need later process the port according to the error code
		klog.Errorf(" AttachPortToVM: attach port[id: %s] to vm[id: %s] error: %v, retry %d times timeout, just exit",
			reqObj.PortID, reqObj.VMID, err, OpenStackOpsRetryTime)
		SaveExceptPort(reqObj.PortID, PortTypeAttatched, reqObj.VMID, "attach port to vm failed")
		return fmt.Errorf("%v:attach port to vm error", err)
	}
	klog.Infof(" AttachPortToVM: attach port[id: %s] "+
		"to vm[id: %s] success", reqObj.PortID, reqObj.VMID)

	if iaas.GetIaaS(tenantID).GetType() == "VNFM" {
		return nil
	}
	err = portOpsObj.CheckNicStatus(tranID, PortStatusActive, CheckPortStatusNumberShort,
		CheckPortStatusIntvalShort, iaas.GetIaaS(tenantID), reqObj.PortID)
	if err != nil {
		klog.Errorf(" AttachPortToVM: check port[id: %s] "+
			"to status: %s failed, error: %v",
			reqObj.PortID, PortStatusActive, err)
		detachErr := portOpsObj.DetachPortFromVM(tranID, iaas.GetIaaS(tenantID), reqObj)
		if detachErr != nil {
			klog.Errorf(" AttachPortToVM: Detach port[id: %s] failed, error: %v", reqObj.PortID, err)
			// save exceptional port
			SaveExceptPort(reqObj.PortID, PortTypeAttatched, reqObj.VMID,
				"attach port to vm but wait ACTIVE status timeout")
		}
		return fmt.Errorf("%v:checkNicStatus error", err)
	}
	return nil
}

func (self *PortService) DetachPortFromVM(tranID TranID, tenantID string, reqObj *PortVMOpsReq) error {
	klog.Infof(" DetachPortFromVM: detach port[id: %s] from vm[id: %s] start", reqObj.PortID, reqObj.VMID)
	var portOpsObj PortOpsAPI = &PortOps{}
	for idx := 0; idx < OpenStackOpsRetryTime; idx++ {
		err := portOpsObj.DetachPortFromVM(tranID, iaas.GetIaaS(tenantID), reqObj)
		if err == nil {
			break
		} else if idx+1 == OpenStackOpsRetryTime {
			klog.Errorf(" DetachPortFromVM: detach port[%s] from vm[%s] error: %v, retry %d times timeout, just exit",
				reqObj.PortID, reqObj.VMID, err, OpenStackOpsRetryTime)
			// save exceptional port
			SaveExceptPort(reqObj.PortID, PortTypeAttatched, reqObj.VMID, "detach port from vm failed")
			return errors.New("detach port from vm error")
		}

		time.Sleep(OpenStackOpsRetryIntval * time.Second)
		klog.Errorf(" DetachPortFromVM: detach port[id: %s] from vm[id: %s] failed, errro: %v, just retry %d time",
			reqObj.PortID, reqObj.VMID, err, idx+1)
	}

	klog.Infof(" DetachPortFromVM: detach port[id: %s] to vm[id: %s] success", reqObj.PortID, reqObj.VMID)

	if iaas.GetIaaS(tenantID).GetType() == "VNFM" {
		return nil
	}
	/*after nova detach port, check port state. If port state is "DOWN", neutron will delay delete port.*/
	err := portOpsObj.CheckNicStatus(tranID, portStatusDown,
		CheckPortStatusNumberShort, CheckPortStatusIntvalShort,
		iaas.GetIaaS(tenantID), reqObj.PortID)
	if err != nil {
		klog.Errorf("cmdDel: DeletePort port[id: %s] failed, error: %v", reqObj.PortID, err)
		// save exceptional port
		SaveExceptPort(reqObj.PortID, PortTypeNonattatched, reqObj.VMID,
			"detach port from vm but wait DOWN status timeout")
		return fmt.Errorf("%v:checkNicStatus error", err)
	}

	return nil
}

func savePortsToDBAndCache(intersAll []*iaasaccessor.Interface, createReq *CreatePortReq) error {
	ports := TransCreatePortsToLogicalPorts(intersAll, createReq)
	// todo: add defer rollback code in future
	for _, port := range ports {
		err := SaveLogicalPort(port)
		if err != nil {
			klog.Errorf("CreateBulkPorts: SaveLogicalPort[%+v] FAIL, error: %v", port, err)
			return err
		}

		portObj := TransLogicalPortToPortObj(port)
		err = GetPortObjRepoSingleton().Add(portObj)
		if err != nil {
			klog.Errorf("CreateBulkPorts: GetPortObjRepoSingleton().Add[%+v] FAIL, error: %v", portObj, err)
			return err
		}
	}

	return nil
}

func deletePortsFromDBAndCache(intersAll []*iaasaccessor.Interface, createReq *CreatePortReq) error {
	ports := TransCreatePortsToLogicalPorts(intersAll, createReq)
	for _, port := range ports {
		err := DeleteLogicalPort(port.ID)
		if err != nil {
			klog.Errorf("deletePortsFromDBAndCache: DeleteLogicalPort[%v] FAIL, error: %v", port, err)
			return err
		}

		portObj := TransLogicalPortToPortObj(port)
		err = GetPortObjRepoSingleton().Del(portObj.ID)
		if err != nil {
			klog.Errorf("deletePortsFromDBAndCache: GetPortObjRepoSingleton().Del[%+v] FAIL, error: %v", portObj, err)
			return err
		}
	}

	return nil
}

func rollbackDeleteIaasPorts(intersFromIPGroup []*iaasaccessor.Interface, intersNotFixIP []*iaasaccessor.Interface, createReq *CreatePortReq) {
	for _, intface := range intersFromIPGroup {
		portObj := &PortObj{
			TenantID:  createReq.TenantID,
			NetworkID: createReq.NetworkName,
			ID:        intface.Id,
			IPGroupID: intface.IPGroupID,
		}
		err := destoryIPGroupPort(portObj)
		if err != nil {
			klog.Warningf("rollbackDeleteIaasPorts : destoryIPGroupPort(portObj:[%v]) err, error is [%v]", portObj, err)
		}
	}
	for _, intface := range intersNotFixIP {
		iaasObj := iaas.GetIaaS(createReq.TenantID)
		err := iaasObj.DeletePort(intface.Id)
		if err != nil {
			klog.Warningf("rollbackDeleteIaasPorts : iaasObj.DeletePort(intface.Id:[%v]) err, error is [%v]", intface.Id, err)
		}

	}

}

func rollbackDeletePortsFromDBAndCache(intersAll []*iaasaccessor.Interface, createReq *CreatePortReq) {
	err := deletePortsFromDBAndCache(intersAll, createReq)
	if err != nil {
		klog.Warningf("rollbackDeletePortsFromDBAndCache: deletePortsFromDBAndCache(intersAll:[%v], createReq[%v]) err, error is [%v]")
	}
}

func CreateBulkPorts(req *mgriaas.MgrBulkPortsReq) (*mgragt.CreatePortsResp, error) {
	//todo rename this function
	var intersFromIPGroup []*iaasaccessor.Interface
	var intersNotFixIP []*iaasaccessor.Interface
	createReq := &CreatePortReq{
		AgtPortReq: agtmgr.AgtPortReq{
			ClusterID: req.Ports[0].ClusterID,
			PodNs:     req.Ports[0].PodNs,
			PodName:   req.Ports[0].PodName,
			TenantID:  req.Ports[0].TenantId,
		},
	}
	intersAll := []*iaasaccessor.Interface{}
	var err error

	defer func() {
		if err != nil {
			rollbackDeleteIaasPorts(intersFromIPGroup, intersNotFixIP, createReq)
			rollbackDeletePortsFromDBAndCache(intersAll, createReq)
		}
	}()

	portsMap, err := buildBulkPortsReq(req)
	if err != nil {
		klog.Errorf("CreateBulkPorts buildBulkPortsReq failed. error: [%v]", err.Error())
		return nil, err
	}

	// get ports from ipgroups
	mgrBulkPortsReqNotFixIP := portsMap[noFixedIPKey]
	mgrBulkPortsReqFromIPGroup := portsMap[IPGroupKey]

	intersFromIPGroup, err = createBulkPortsFromIPGroup(mgrBulkPortsReqFromIPGroup)
	if err != nil {
		klog.Errorf("CreateBulkPorts createBulkPortsFromIPGroup failed. error: [%v]", err.Error())
		return nil, err
	}

	// create iaas ports
	intersNotFixIP, err = createNormalBulkPorts(mgrBulkPortsReqNotFixIP)
	if err != nil {
		klog.Errorf("CreateBulkPorts: createNormalBulkPorts FAIL, error: %v", err)
		return nil, err
	}

	intersAll = append(intersAll, intersFromIPGroup...)
	intersAll = append(intersAll, intersNotFixIP...)
	err = savePortsToDBAndCache(intersAll, createReq)
	if err != nil {
		klog.Errorf("CreateBulkPorts: savePortsToDBAndCache FAIL, error: %v", err)
		return nil, err
	}

	resp := &mgragt.CreatePortsResp{}
	for _, inter := range intersAll {
		var port *mgragt.CreatePortResp
		port, err = AssembleResponse(iaas.GetIaaS(req.Ports[0].TenantId), inter)
		if err != nil {
			return nil, err
		}
		resp.Ports = append(resp.Ports, port.Port)
	}

	return resp, nil
}

func createNormalBulkPorts(req *mgriaas.MgrBulkPortsReq) ([]*iaasaccessor.Interface, error) {
	nics, err := GetPortServiceObj().CreateBulkPorts(req)
	if err != nil {
		klog.Errorf("createNormalBulkPorts: GetPortServiceObj().CreateBulkPorts() req[%v] FAIL, error: %v", req, err)
		return nil, err
	}

	klog.Errorf("createNormalBulkPorts: create ports[%v] SUCC", nics)
	return nics, nil
}

func CreateLogicalPort(reqObj *CreatePortReq) (*mgragt.CreatePortResp, error) {
	portObj, rsp, err := GetPortServiceObj().CreatePort(TranID(1), reqObj)
	if err != nil {
		klog.Errorf("CreateLogicalPort: GetPortServiceObj().CreatePort(reqObj: %v) FAILED, error: %v", reqObj, err)
		return nil, err
	}

	logiPort := MakeLogicalPort(portObj)
	err = SaveLogicalPort(logiPort)
	if err != nil {
		klog.Errorf("CreateLogicalPort: SaveLogicalPort[%v] FAILED, error: %v", logiPort, err)
		return nil, err
	}

	err = GetPortObjRepoSingleton().Add(portObj)
	if err != nil {
		klog.Errorf("CreateLogicalPort: GetPortObjRepoSingleton().Add[%v] FAILED, error: %v", portObj, err)
		return nil, err
	}
	return rsp, nil
}

func DestroyLogicalPort(portID string) error {
	portObj, err := GetPortObjRepoSingleton().Get(portID)
	if err != nil {
		klog.Errorf("DestroyLogicalPort: GetPortObjRepoSingleton().Get() error: [%v], portID: [%v]", err.Error(), portID)
		return err
	}

	if portObj.IPGroupID != "" {
		err = destoryIPGroupPort(portObj)
		if err != nil {
			klog.Errorf("DestroyLogicalPort: destoryIPGroupPort(portID: %s) error: [%v]", portID, err)
			return err
		}
	} else {
		err = destroyNormalLogicalPort(portObj.ID, portObj.TenantID)
		if err != nil {
			klog.Errorf("DestroyLogicalPort: destroyNormalLogicalPort(portID: %s) error: [%v]", portID, err)
			return err
		}
	}

	err = DeleteLogicalPort(portID)
	if err != nil {
		klog.Errorf("DeletePort: DeleteLogicalPort(portID: %s) FAILED, error: %v", portID, err)
		return err
	}

	err = GetPortObjRepoSingleton().Del(portID)
	if err != nil {
		klog.Errorf("DeletePort: GetPortObjRepoSingleton().Del(portID: %s) FAILED, error: %v", portID, err)
		return err
	}

	return err
}

func destoryIPGroupPort(portObj *PortObj) error {
	klog.Infof(" DeletePort: delete port[id: %s] from ip group start", portObj.ID)
	ig := IPGroup{
		TenantID:  portObj.TenantID,
		NetworkID: portObj.NetworkID,
		ID:        portObj.IPGroupID}
	err := ig.ReleaseIP(portObj.ID)
	if err != nil {
		klog.Warningf("destoryIPGroupPort: ig.ReleaseIP(portID: %s) FAIL, error: %v", portObj.ID, err)

	}
	return err
}

func destroyNormalLogicalPort(portID, tenantID string) error {
	err := GetPortServiceObj().DeletePort(TranID(1), portID, tenantID)
	if err != nil {
		klog.Errorf("DestroyLogicalPort: GetPortServiceObj().DeletePort(portID: %s) FAILED, error: %v", portID, err)
		return err
	}

	return nil
}

func init() {
	GetPortObjRepoSingleton().Init()
}

// port table start
type PortObj struct {
	// key definition info
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status int    `json:"status"` // life cycle status, not IaaS layer status

	// layer 2-3 info
	IP         string `json:"ip"`
	MACAddress string `json:"mac_address"`

	NetworkID string `json:"network_id"`
	SubnetID  string `json:"subnet_id"`
	IPGroupID string `json:"ipgroup_id"`

	// owner info
	NodeID    string `json:"node_id"`
	ClusterID string `json:"cluster_id"`
	OwnerType string `json:"owner_type"` // node or pod(ms), agent...
	TenantID  string `json:"tenant_id"`
	PodName   string `json:"pod_name"`
	PodNs     string `json:"pod_ns"`
}

type PortArray struct {
	Lock  sync.RWMutex
	ports map[string]*PortObj
}

type PortObjRepo struct {
	Lock    sync.RWMutex
	indexer cache.Indexer
}

var portObjRepo PortObjRepo

func GetPortObjRepoSingleton() *PortObjRepo {
	return &portObjRepo
}

func PortObjKeyFunc(obj interface{}) (string, error) {
	if obj == nil {
		klog.Error("PortObjKeyFunc: obj arg is nil")
		return "", errobj.ErrObjectPointerIsNil
	}

	portObj, ok := obj.(*PortObj)
	if !ok {

		klog.Error("PortObjKeyFunc: obj arg is not type: *PortObj")
		return "", errobj.ErrArgTypeMismatch
	}

	return portObj.ID, nil
}

func NetworkIDIndexFunc(obj interface{}) ([]string, error) {
	if obj == nil {
		klog.Error("NetworkIDIndexFunc: obj arg is nil")
		return nil, errobj.ErrObjectPointerIsNil
	}

	portObj, ok := obj.(*PortObj)
	if !ok {
		klog.Error("NetworkIDIndexFunc: obj arg is not type: *PortObj")
		return nil, errobj.ErrArgTypeMismatch
	}

	return []string{portObj.NetworkID}, nil
}

func TenantIDIndexFunc(obj interface{}) ([]string, error) {
	if obj == nil {
		klog.Error("TenantIDIndexFunc: obj arg is nil")
		return nil, errobj.ErrObjectPointerIsNil
	}

	portObj, ok := obj.(*PortObj)
	if !ok {
		klog.Error("TenantIDIndexFunc: obj arg is not type: *PortObj")
		return nil, errobj.ErrArgTypeMismatch
	}

	return []string{portObj.TenantID}, nil
}

const (
	NetworkIDIndex = "network_id"
	TenantIDIndex  = "tenant_id"
)

func (p *PortObjRepo) Init() {
	indexers := cache.Indexers{
		NetworkIDIndex: NetworkIDIndexFunc,
		TenantIDIndex:  TenantIDIndexFunc}
	p.indexer = cache.NewIndexer(PortObjKeyFunc, indexers)
}

func MakePortObj(port *iaasaccessor.Interface, req *CreatePortReq) *PortObj {
	return &PortObj{
		ID:         port.Id,
		Name:       port.Name,
		Status:     CreatedOK,
		IP:         port.Ip,
		MACAddress: port.MacAddress,
		NetworkID:  port.NetworkId,
		SubnetID:   port.SubnetId,
		ClusterID:  req.ClusterID,
		OwnerType:  constvalue.OwnerTypePod,
		TenantID:   req.TenantID,
		PodNs:      req.PodNs,
		PodName:    req.PodName,
		IPGroupID:  port.IPGroupID,
	}
}

func (p *PortObjRepo) Add(portObj *PortObj) error {
	err := p.indexer.Add(portObj)
	if err != nil {
		klog.Errorf("PortObjRepo.Add: add obj[%v] to repo FAILED, error: %v", portObj, err)
		return err
	}
	klog.Infof("PortObjRepo.Add: add obj[%v] to repo SUCC", portObj)
	return nil
}

func (p *PortObjRepo) Del(portID string) error {
	err := p.indexer.Delete(portID)
	if err != nil {
		klog.Errorf("PortObjRepo.Del: Delete(portID: %s) FAILED, error: %v", portID, err)
		return err
	}

	klog.Infof("PortObjRepo.Del: Delete(portID: %s) SUCC", portID)
	return nil
}

func (p *PortObjRepo) Get(portID string) (*PortObj, error) {
	item, exists, err := p.indexer.GetByKey(portID)
	if err != nil {
		klog.Errorf("PortObjRepo.Get: portID[%s]'s object FAILED, error: %v", portID, err)
		return nil, err
	}
	if !exists {
		klog.Errorf("PortObjRepo.Get: portID[%s]'s object not found", portID)
		return nil, errobj.ErrRecordNotExist
	}

	portObj, ok := item.(*PortObj)
	if !ok {
		klog.Errorf("PortObjRepo.Get: portID[%s]'s object[%v] type not match PortObj", portID, item)
		return nil, errobj.ErrObjectTypeMismatch
	}
	klog.Infof("PortObjRepo.Get: portID[%s]'s object[%v] SUCC", portID, portObj)
	return portObj, nil
}

func (p *PortObjRepo) Update(portObj *PortObj) error {
	err := p.indexer.Update(portObj)
	if err != nil {
		klog.Errorf("PortObjRepo.Update: Update[%v] FAILED, error: %v", portObj, err)
	}

	klog.Infof("PortObjRepo.Update: Update[%v] SUCC", portObj)
	return nil
}

func (p *PortObjRepo) ListByNetworkID(networkID string) ([]*PortObj, error) {
	objs, err := p.indexer.ByIndex(NetworkIDIndex, networkID)
	if err != nil {
		klog.Errorf("PortObjRepo.ListByNetworkID: ByIndex(network_id, networkID: %s) FAILED, error: %v", networkID, err)
		return nil, err
	}

	klog.Infof("PortObjRepo.ListByNetworkID: ByIndex(network_id, networkID: %s) SUCC, array: %v", networkID, objs)
	portObjs := make([]*PortObj, 0)
	for _, obj := range objs {
		port, ok := obj.(*PortObj)
		if !ok {
			klog.Errorf("PortObjRepo.ListByNetworkID: index result object: %v is not type *PortObj, skip", obj)
			continue
		}
		portObjs = append(portObjs, port)
	}
	return portObjs, nil
}

func (p *PortObjRepo) ListByTenantID(tenantID string) ([]*PortObj, error) {
	objs, err := p.indexer.ByIndex(TenantIDIndex, tenantID)
	if err != nil {
		klog.Errorf("PortObjRepo.ListByTenantID: ByIndex(tenant_id, tenantID: %s) FAILED, error: %v", tenantID, err)
		return nil, err
	}

	klog.Infof("PortObjRepo.ListByTenantID: ByIndex(tenant_id, tenantID: %s) SUCC, array: %v", tenantID, objs)
	portObjs := make([]*PortObj, 0)
	for _, obj := range objs {
		port, ok := obj.(*PortObj)
		if !ok {
			klog.Errorf("PortObjRepo.ListByTenantID: index result object: %v is not type *PortObj, skip", obj)
			continue
		}
		portObjs = append(portObjs, port)
	}
	return portObjs, nil
}

// port table end

// port life cycle status definition
const (
	WaitCreating = iota
	Creating
	CreatedFailed
	CreatedOK
	WaitDestroy
	Destroyed
)

// logical port lifecycle management structure, should only contains least members
type LogicalPort struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status int    `json:"status"` // life cycle status, not IaaS layer status

	IP         string `json:"ip"`
	MACAddress string `json:"mac_address"`

	NetworkID string `json:"network_id"`
	SubnetID  string `json:"subnet_id"`
	IPGroupID string `json:"ipgroup_id"`

	NodeID    string `json:"node_id"`
	ClusterID string `json:"cluster_id"`

	OwnerType string `json:"owner_type"` // node or pod(ms), agent...
	TenantID  string `json:"tenant_id"`
	PodName   string `json:"pod_name"`
	PodNs     string `json:"pod_ns"`
}

func TransCreatePortsToLogicalPorts(ports []*iaasaccessor.Interface, req *CreatePortReq) []*LogicalPort {
	var logicPorts = []*LogicalPort{}
	for _, port := range ports {
		tmpPort := TransCreatePortToLogicalPort(port, req)
		logicPorts = append(logicPorts, tmpPort)
	}
	return logicPorts
}

func TransCreatePortToLogicalPort(port *iaasaccessor.Interface, req *CreatePortReq) *LogicalPort {
	return &LogicalPort{
		ID:         port.Id,
		Name:       port.Name,
		Status:     CreatedOK,
		IP:         port.Ip,
		MACAddress: port.MacAddress,
		NetworkID:  port.NetworkId,
		SubnetID:   port.SubnetId,
		IPGroupID:  port.IPGroupID,
		ClusterID:  req.ClusterID,
		NodeID:     port.VmId,
		OwnerType:  constvalue.OwnerTypePod,
		TenantID:   req.TenantID,
		PodNs:      req.PodNs,
		PodName:    req.PodName,
	}
}

func MakeLogicalPort(portObj *PortObj) *LogicalPort {
	return &LogicalPort{
		ID:         portObj.ID,
		Name:       portObj.Name,
		Status:     portObj.Status,
		IP:         portObj.IP,
		MACAddress: portObj.MACAddress,
		NetworkID:  portObj.NetworkID,
		SubnetID:   portObj.SubnetID,
		IPGroupID:  portObj.IPGroupID,
		ClusterID:  portObj.ClusterID,
		NodeID:     portObj.NodeID,
		OwnerType:  portObj.OwnerType,
		TenantID:   portObj.TenantID,
		PodNs:      portObj.PodNs,
		PodName:    portObj.PodName,
	}
}

func TransLogicalPortToPortObj(port *LogicalPort) *PortObj {
	return &PortObj{
		ID:         port.ID,
		Name:       port.Name,
		Status:     port.Status,
		IP:         port.IP,
		MACAddress: port.MACAddress,
		NetworkID:  port.NetworkID,
		SubnetID:   port.SubnetID,
		IPGroupID:  port.IPGroupID,
		ClusterID:  port.ClusterID,
		NodeID:     port.NodeID,
		OwnerType:  port.OwnerType,
		TenantID:   port.TenantID,
		PodNs:      port.PodNs,
		PodName:    port.PodName,
	}
}

const KnitterManagerKeyRoot = "/knitter/manager"

func getLogicalPortsKey() string {
	return KnitterManagerKeyRoot + "/ports"
}

func createLogicalPortKey(portID string) string {
	return getLogicalPortsKey() + "/" + portID
}

func SaveLogicalPort(port *LogicalPort) error {
	portInBytes, err := json.Marshal(port)
	if err != nil {
		klog.Errorf("SaveLogicalPort: json.Marshal(port: %v) FAILED, error: %v", port, err)
		return errobj.ErrMarshalFailed
	}
	key := createLogicalPortKey(port.ID)
	err = common.GetDataBase().SaveLeaf(key, string(portInBytes))
	if err != nil {
		klog.Errorf("SaveLogicalPort: SaveLeaf(key: %s, value: %s) FAILED, error: %v", key, string(portInBytes), err)
		return err
	}
	klog.Infof("SaveLogicalPort: save port[%v] SUCC", port)
	return nil
}

func DeleteLogicalPort(portID string) error {
	key := createLogicalPortKey(portID)
	err := common.GetDataBase().DeleteLeaf(key)
	if err != nil && !IsKeyNotFoundError(err) {
		klog.Errorf("SaveLogicalPort: DeleteLeaf(key: %s) FAILED, error: %v", key, err)
		return err
	}
	klog.Infof("DeleteLogicalPort: delete port[id: %s] SUCC", portID)
	return nil
}

func GetLogicalPort(portID string) (*LogicalPort, error) {
	key := createLogicalPortKey(portID)
	value, err := common.GetDataBase().ReadLeaf(key)
	if err != nil {
		klog.Errorf("GetLogicalPort: ReadLeaf(key: %s) FAILED, error: %v", key, err)
		return nil, err
	}

	port, err := UnmarshalLogicPort([]byte(value))
	if err != nil {
		klog.Errorf("GetLogicalPort: UnmarshalLogicPort(%v) FAILED, error: %v", value, err)
		return nil, errobj.ErrUnmarshalFailed
	}

	klog.Infof("GetLogicalPort: get port[%v] SUCC", port)
	return port, nil
}

var UnmarshalLogicPort = func(value []byte) (*LogicalPort, error) {
	var port LogicalPort
	err := json.Unmarshal([]byte(value), &port)
	if err != nil {
		klog.Errorf("UnmarshalLogicPort: json.Unmarshal(%s) FAILED, error: %v", string(value), err)
		return nil, errobj.ErrUnmarshalFailed
	}

	klog.Infof("UnmarshalLogicPort: logical port[%+v] SUCC", port)
	return &port, nil
}

func GetAllLogicalPorts() ([]*LogicalPort, error) {
	key := getLogicalPortsKey()
	nodes, err := common.GetDataBase().ReadDir(key)
	if err != nil {
		klog.Errorf("GetAllLogicalPorts: ReadDir(key: %s) FAILED, error: %v", key, err)
		return nil, err
	}

	ports := make([]*LogicalPort, 0)
	for _, node := range nodes {
		port, err := UnmarshalLogicPort([]byte(node.Value))
		if err != nil {
			klog.Errorf("GetAllLogicalPorts: UnmarshalLogicPort(port: %s) FAILED, error: %v", node.Value, err)
			return nil, err
		}
		ports = append(ports, port)
	}

	klog.Infof("GetAllLogicalPorts: get all logical ports: %v SUCC", ports)
	return ports, nil
}

var createIaasBulkPorts = func(req *mgriaas.MgrBulkPortsReq) ([]*iaasaccessor.Interface, error) {
	if len(req.Ports) == 0 {
		return nil, nil
	}

	var err error = nil
	var ports []*iaasaccessor.Interface
	for idx := 0; idx < OpenStackOpsRetryTime; idx++ {
		ports, err = iaas.GetIaaS(req.Ports[0].TenantID).CreateBulkPorts(req)
		klog.Errorf("%d times, error: %v", idx, err)
		if err == nil {
			return ports, nil
		}
	}

	klog.Errorf("createIaasBulkPorts: iaas create bulk ports FAIL, error: %v", err)
	return nil, err
}

func LoadAllPortObjs() error {
	logiPorts, err := GetAllLogicalPorts()
	if err != nil {
		klog.Errorf("LoadAllPortObjs: GetAllLogicalPorts FAILED, error: %v", err)
		return err
	}

	for _, logiPort := range logiPorts {
		portObj := TransLogicalPortToPortObj(logiPort)
		err = GetPortObjRepoSingleton().Add(portObj)
		if err != nil {
			klog.Errorf("LoadAllPortObjs: GetPortObjRepoSingleton().Add(portObj: %v) FAILED, error: %v", portObj, err)
			return err
		}
		klog.Tracef("LoadAllPortObjs: GetPortObjRepoSingleton().Add(portObj: %v) SUCC", portObj)
	}
	return nil
}

func TransPhysicalPortToPhysPortObj(physPort *PhysicalPort) *PhysPortObj {
	return &PhysPortObj{
		ID:         physPort.ID,
		Name:       physPort.Name,
		Status:     physPort.Status,
		VnicType:   physPort.VnicType,
		IP:         physPort.IP,
		MacAddress: physPort.MacAddress,
		NetworkID:  physPort.NetworkID,
		SubnetID:   physPort.SubnetID,
		NodeID:     physPort.NodeID,
		ClusterID:  physPort.ClusterID,
		OwnerType:  physPort.OwnerType,
		TenantID:   physPort.TenantID,
	}
}

func LoadPhysicalPortObjects() error {
	physPorts, err := GetAllPhysicalPorts()
	if err != nil {
		klog.Errorf("LoadPhysicalPortObjects: GetAllPhysicalPorts FAILED, error: %v", err)
		return err
	}

	for _, physPort := range physPorts {
		physPortObj := TransPhysicalPortToPhysPortObj(physPort)
		err = GetPhysPortObjRepoSingleton().Add(physPortObj)
		if err != nil {
			klog.Errorf("LoadPhysicalPortObjects: GetSubnetObjRepoSingleton().Add(subnetObj: %v) FAILED, error: %v",
				physPortObj, err)
			return err
		}
		klog.Tracef("LoadPhysicalPortObjects: GetSubnetObjRepoSingleton().Add(subnetObj: %v) SUCC",
			physPortObj)
	}
	return nil
}

type LoadResouceObjectFunc func() error
