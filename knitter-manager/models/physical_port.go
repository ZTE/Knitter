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
	"sync"

	"k8s.io/client-go/tools/cache"

	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/pkg/adapter"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-agt"
	"github.com/ZTE/Knitter/pkg/klog"
)

// physical port object persistence functions
type PhysicalPort struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`

	// layer 2-3 info
	VnicType   string `json:"vnic_type"`
	IP         string `json:"ip"`
	MacAddress string `json:"mac_address"`
	NetworkID  string `json:"network_id"`
	SubnetID   string `json:"subnet_id"`

	// owner info
	NodeID    string `json:"node_id"`
	ClusterID string `json:"cluster_id"`
	OwnerType string `json:"owner_type"`
	TenantID  string `json:"tenant_id"`
}

func getPhysicalPortsKey() string {
	return KnitterManagerKeyRoot + "/physical_ports"
}

func createPhysicalPortKey(portID string) string {
	return getPhysicalPortsKey() + "/" + portID
}

func SavePhysicalPort(port *PhysicalPort) error {
	portInBytes, err := json.Marshal(port)
	if err != nil {
		klog.Errorf("SavePhysicalPort: json.Marshal(port: %v) FAILED, error: %v", port, err)
		return errobj.ErrMarshalFailed
	}
	key := createPhysicalPortKey(port.ID)
	err = common.GetDataBase().SaveLeaf(key, string(portInBytes))
	if err != nil {
		klog.Errorf("SavePhysicalPort: SaveLeaf(key: %s, value: %s) FAILED, error: %v", key, string(portInBytes), err)
		return err
	}
	klog.Infof("SavePhysicalPort: save physical port[%v] SUCC", port)
	return nil
}

func DeletePhysicalPort(portID string) error {
	key := createPhysicalPortKey(portID)
	err := common.GetDataBase().DeleteLeaf(key)
	if err != nil && !IsKeyNotFoundError(err) {
		klog.Errorf("DeletePhysicalPort: DeleteLeaf(key: %s) FAILED, error: %v", key, err)
		return err
	}
	klog.Infof("DeletePhysicalPort: delete physical port[id: %s] SUCC", portID)
	return nil
}

func GetPhysicalPort(portID string) (*PhysicalPort, error) {
	key := createPhysicalPortKey(portID)
	value, err := common.GetDataBase().ReadLeaf(key)
	if err != nil {
		klog.Errorf("GetPhysicalPort: ReadLeaf(key: %s) FAILED, error: %v", key, err)
		return nil, err
	}
	port, err := UnmarshalPhysPort([]byte(value))
	if err != nil {
		klog.Errorf("GetPhysicalPort: UnmarshalPhysPort(%s) FAILED, error: %v", value, err)
		return nil, err
	}

	klog.Infof("GetPhysicalPort: get physical port[%v] SUCC", port)
	return port, nil
}

var UnmarshalPhysPort = func(value []byte) (*PhysicalPort, error) {
	var port PhysicalPort
	err := json.Unmarshal([]byte(value), &port)
	if err != nil {
		klog.Errorf("UnmarshalPhysPort: json.Unmarshal(%v) FAILED, error: %v", value, err)
		return nil, errobj.ErrUnmarshalFailed
	}

	klog.Infof("UnmarshalPhysPort: physical port[%v] SUCC", port)
	return &port, nil
}

func GetAllPhysicalPorts() ([]*PhysicalPort, error) {
	key := getPhysicalPortsKey()
	nodes, err := common.GetDataBase().ReadDir(key)
	if err != nil {
		klog.Errorf("GetAllPhysicalPorts: ReadDir(key: %s) FAILED, error: %v", key, err)
		return nil, err
	}

	ports := make([]*PhysicalPort, 0)
	for _, node := range nodes {
		port, err := UnmarshalPhysPort([]byte(node.Value))
		if err != nil {
			klog.Errorf("GetAllPhysicalPorts: UnmarshalPhysPort(port: %s) FAILED, error: %v", node.Value, err)
			return nil, err
		}
		ports = append(ports, port)
	}

	klog.Tracef("GetAllPhysicalPorts: get all ports: %v SUCC", ports)
	return ports, nil
}

func MakePhysicalPort(portObj *PortObj, vnicType string) *PhysicalPort {
	return &PhysicalPort{
		ID:   portObj.ID,
		Name: portObj.Name,
		//Status:     portObj.Status, // OwnerTypeNode
		VnicType:   vnicType,
		IP:         portObj.IP,
		MacAddress: portObj.MACAddress,
		NetworkID:  portObj.NetworkID,
		SubnetID:   portObj.SubnetID,
		ClusterID:  portObj.ClusterID,
		OwnerType:  constvalue.OwnerTypeNode,
		NodeID:     portObj.NodeID,
		TenantID:   portObj.TenantID,
	}
}

func init() {
	GetPhysPortObjRepoSingleton().Init()
}

// physical port table start
type PhysPortObj struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`

	// layer 2-3 info
	VnicType   string `json:"vnic_type"`
	IP         string `json:"ip"`
	MacAddress string `json:"mac_address"`
	NetworkID  string `json:"network_id"`
	SubnetID   string `json:"subnet_id"`

	// owner info
	NodeID    string `json:"node_id"`
	ClusterID string `json:"cluster_id"`
	OwnerType string `json:"owner_type"`
	TenantID  string `json:"tenant_id"`
}

type PhysPortObjRepo struct {
	Lock    sync.RWMutex
	indexer cache.Indexer
}

var physPortObjRepo PhysPortObjRepo

func GetPhysPortObjRepoSingleton() *PhysPortObjRepo {
	return &physPortObjRepo
}

func PhysPortObjKeyFunc(obj interface{}) (string, error) {
	if obj == nil {
		klog.Error("PhysPortObjKeyFunc: obj arg is nil")
		return "", errobj.ErrObjectPointerIsNil
	}

	physPortObj, ok := obj.(*PhysPortObj)
	if !ok {
		klog.Error("PhysPortObjKeyFunc: obj arg is not type: *PhysPortObj")
		return "", errobj.ErrArgTypeMismatch
	}

	return physPortObj.ID, nil
}

func PhysPortNetworkIDIndexFunc(obj interface{}) ([]string, error) {
	if obj == nil {
		klog.Error("PhysPortNetworkIDIndexFunc: obj arg is nil")
		return nil, errobj.ErrObjectPointerIsNil
	}

	physPortObj, ok := obj.(*PhysPortObj)
	if !ok {
		klog.Error("PhysPortNetworkIDIndexFunc: obj arg is not type: *PhysPortObj")
		return nil, errobj.ErrArgTypeMismatch
	}

	return []string{physPortObj.NetworkID}, nil
}

func PhysPortTenantIDIndexFunc(obj interface{}) ([]string, error) {
	if obj == nil {
		klog.Error("PhysPortTenantIDIndexFunc: obj arg is nil")
		return nil, errobj.ErrObjectPointerIsNil
	}

	physPortObj, ok := obj.(*PhysPortObj)
	if !ok {
		klog.Error("PhysPortTenantIDIndexFunc: obj arg is not type: *PhysPortObj")
		return nil, errobj.ErrArgTypeMismatch
	}

	return []string{physPortObj.TenantID}, nil
}

func (p *PhysPortObjRepo) Init() {
	indexers := cache.Indexers{
		NetworkIDIndex: PhysPortNetworkIDIndexFunc,
		TenantIDIndex:  PhysPortTenantIDIndexFunc}
	p.indexer = cache.NewIndexer(PhysPortObjKeyFunc, indexers)
}

func (p *PhysPortObjRepo) Add(port *PhysPortObj) error {
	err := p.indexer.Add(port)
	if err != nil {
		klog.Errorf("PhysPortObjRepo.Add: add obj[%v] to repo FAILED, error: %v", port, err)
		return err
	}
	klog.Infof("PhysPortObjRepo.Add: add obj[%v] to repo SUCC", port)
	return nil
}

func (p *PhysPortObjRepo) Del(portID string) error {
	err := p.indexer.Delete(portID)
	if err != nil {
		klog.Errorf("PhysPortObjRepo.Del: Delete(portID: %s) FAILED, error: %v", portID, err)
		return err
	}

	klog.Infof("PhysPortObjRepo.Del: Delete(portID: %s) SUCC", portID)
	return nil
}

func (p *PhysPortObjRepo) Get(portID string) (*PhysPortObj, error) {
	item, exists, err := p.indexer.GetByKey(portID)
	if err != nil {
		klog.Errorf("PhysPortObjRepo.Get: portID[%s]'s object FAILED, error: %v", portID, err)
		return nil, err
	}
	if !exists {
		klog.Errorf("PhysPortObjRepo.Get: portID[%s]'s object not found", portID)
		return nil, errobj.ErrRecordNotExist
	}

	port, ok := item.(*PhysPortObj)
	if !ok {
		klog.Errorf("PhysPortObjRepo.Get: portID[%s]'s object[%v] type not match *PhysPortObj", portID, item)
		return nil, errobj.ErrObjectTypeMismatch
	}
	klog.Infof("PhysPortObjRepo.Get: portID[%s]'s object[%v] SUCC", portID, port)
	return port, nil
}

func (p *PhysPortObjRepo) Update(port *PhysPortObj) error {
	err := p.indexer.Update(port)
	if err != nil {
		klog.Errorf("PhysPortObjRepo.Update: Update[%v] FAILED, error: %v", port, err)
	}

	klog.Infof("PhysPortObjRepo.Update: Update[%v] SUCC", port)
	return nil
}

func (p *PhysPortObjRepo) ListByNetworkID(networkID string) ([]*PhysPortObj, error) {
	objs, err := p.indexer.ByIndex(NetworkIDIndex, networkID)
	if err != nil {
		klog.Errorf("PhysPortObjRepo.ListByNetworkID: ByIndex(network_id: %s) FAILED, error: %v", networkID, err)
		return nil, err
	}

	klog.Infof("PhysPortObjRepo.ListByNetworkID: ByIndex(network_id: %s) SUCC, array: %v", networkID, objs)
	portObjs := make([]*PhysPortObj, 0)
	for _, obj := range objs {
		port, ok := obj.(*PhysPortObj)
		if !ok {
			klog.Errorf("PhysPortObjRepo.ListByNetworkID: index result object: %v is not type *PhysPortObj, skip", obj)
			continue
		}
		portObjs = append(portObjs, port)
	}
	return portObjs, nil
}

func (p *PhysPortObjRepo) ListByTenantID(tenantID string) ([]*PhysPortObj, error) {
	objs, err := p.indexer.ByIndex(TenantIDIndex, tenantID)
	if err != nil {
		klog.Errorf("PhysPortObjRepo.ListByTenantID: ByIndex(tenant_id: %s) FAILED, error: %v", tenantID, err)
		return nil, err
	}

	klog.Infof("PhysPortObjRepo.ListByTenantID: ByIndex(tenant_id: %s) SUCC, array: %v", tenantID, objs)
	portObjs := make([]*PhysPortObj, 0)
	for _, obj := range objs {
		port, ok := obj.(*PhysPortObj)
		if !ok {
			klog.Errorf("PhysPortObjRepo.ListByTenantID: index result object: %v is not type *PhysPortObj, skip", obj)
			continue
		}
		portObjs = append(portObjs, port)
	}
	return portObjs, nil
}

// physical port table end

// physical port ops
func CreateAndAttach(body []byte, tenantID string) (*mgragt.CreatePortResp, error) {
	reqObj4Create, err := buildIaasCreatePortReqObj(body)
	if err != nil {
		klog.Errorf("@@PhyPortController: Build Iaas Req Obj failed, error: %v", err)
		return nil, err
	}
	klog.Info("@@PhyPortController: CreateAndAttach:]User[", tenantID, "] AgentReqBody[", reqObj4Create, "]")
	portObj, portInfo, err := GetPortServiceObj().CreatePort("", reqObj4Create)
	if err != nil {
		klog.Errorf("@@PhyPortController: Create port network[%s] failed, error: %v", reqObj4Create.NetworkName, err)
		return nil, err
	}
	//attach port
	reqObj4Attach, err := buildIaasAttachPortReqObj(reqObj4Create.NodeID, portInfo.Port.PortID)
	if err != nil {
		klog.Errorf("@@PhyPortController: Build Iaas Attach Port Req Obj failed, error:%v", err)
	}
	errAttach := GetPortServiceObj().AttachPortToVM("", reqObj4Create.TenantID, reqObj4Attach)
	if errAttach != nil {
		GetPortServiceObj().DeletePort("", portInfo.Port.PortID, reqObj4Create.TenantID)
		klog.Errorf("@@PhyPortController:Attach Port To VM failed, then delete port[%v]. error:%v",
			portInfo.Port.PortID, errAttach)
		return nil, errAttach
	}

	physPort := MakePhysicalPort(portObj, reqObj4Create.VnicType)
	err = SavePhysicalPort(physPort)
	if err != nil {
		klog.Errorf("CreateLogicalPort: SaveLogicalPort[%v] FAILED, error: %v", physPort, err)
		return nil, err
	}

	physPortObj := TransPhysicalPortToPhysPortObj(physPort)
	err = GetPhysPortObjRepoSingleton().Add(physPortObj)
	if err != nil {
		klog.Errorf("CreateLogicalPort: GetPortObjRepoSingleton().Add[%v] FAILED, error: %v", portObj, err)
		return nil, err
	}

	return portInfo, nil
}

func buildIaasCreatePortReqObj(body []byte) (*CreatePortReq, error) {
	attachReq := agtmgr.AttachPortReq{}
	klog.Infof("@@PhyPortController: Agent request body is: %s", string(body))
	err := adapter.Unmarshal(body, &attachReq)
	if err != nil {
		klog.Errorf("@@PhyPortController:Unmarshal http request body failed, error: %v", err)
		return nil, err
	}
	createReq := CreatePortReq{
		AgtPortReq: agtmgr.AgtPortReq{
			TenantID:    attachReq.TenantID,
			NetworkName: attachReq.NetworkName,
			PortName:    attachReq.PortName,
			VnicType:    attachReq.VnicType,
			NodeID:      attachReq.NodeID,
			PodNs:       "",
			PodName:     "",
			FixIP:       attachReq.FixIP,
			ClusterID:   attachReq.ClusterID,
		},
	}
	return &createReq, nil
}

func buildIaasAttachPortReqObj(vmID, portID string) (*PortVMOpsReq, error) {
	if vmID == "" || portID == "" {
		return nil, errobj.Err403
	}
	reqObj := PortVMOpsReq{
		VMID:   vmID,
		PortID: portID,
	}
	return &reqObj, nil
}

func DetachAndDelete(tranID TranID, vmID, portID, paasTenantID string) error {
	//detach
	errDetach := GetPortServiceObj().DetachPortFromVM(tranID, paasTenantID, &PortVMOpsReq{VMID: vmID, PortID: portID})
	if errDetach != nil {
		klog.Infof("@@PhyPortController:Detach port[id: %s] from vm[id: %s] failed, error: %v", portID, vmID, errDetach)
		return errDetach
	}
	//delete
	errDelete := GetPortServiceObj().DeletePort("", portID, paasTenantID)
	if errDelete != nil {
		klog.Infof("@@PhyPortController: Delete port[id: %s] failed, error: %v", portID, errDelete)
		return errDelete
	}

	errDeleteDB := DeletePhysicalPort(portID)
	if errDeleteDB != nil {
		klog.Infof("@@PhyPortController: Delete Port[id: %s] from db failed, error: %v", portID, errDeleteDB)
		return errDeleteDB
	}

	errDeleteCache := GetPhysPortObjRepoSingleton().Del(portID)
	if errDeleteCache != nil {
		klog.Infof("@@PhyPortController: Delete Port[id: %s] from cache failed, error: %v", portID, errDeleteCache)
		return errDeleteCache
	}

	return nil
}
