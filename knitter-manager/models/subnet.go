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

	"k8s.io/client-go/tools/cache"

	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/pkg/klog"
)

type AllocationPool struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Subnet struct {
	ID         string           `json:"id"`
	NetworkID  string           `json:"network_id"`
	Name       string           `json:"name"`
	CIDR       string           `json:"cidr"`
	GatewayIP  string           `json:"gateway_ip"`
	TenantID   string           `json:"tenant_id"`
	AllocPools []AllocationPool `json:"allocation_pools"`
}

func getSubnetsKey() string {
	return KnitterManagerKeyRoot + "/subnets"
}

func createSubnetKey(subnetID string) string {
	return getSubnetsKey() + "/" + subnetID
}

func SaveSubnet(subnet *Subnet) error {
	subnetInBytes, err := json.Marshal(subnet)
	if err != nil {
		klog.Errorf("SaveSubnet: json.Marshal(network: %v) FAILED, error: %v", subnet, err)
		return errobj.ErrMarshalFailed
	}
	key := createSubnetKey(subnet.ID)
	err = common.GetDataBase().SaveLeaf(key, string(subnetInBytes))
	if err != nil {
		klog.Errorf("SaveSubnet: SaveLeaf(key: %s, value: %s) FAILED, error: %v", key, string(subnetInBytes), err)
		return err
	}
	klog.Infof("SaveSubnet: save network[%v] SUCC", subnet)
	return nil
}

func DelSubnet(subnetID string) error {
	key := createSubnetKey(subnetID)
	err := common.GetDataBase().DeleteLeaf(key)
	if err != nil && !IsKeyNotFoundError(err) {
		klog.Errorf("DelSubnet: DeleteLeaf(key: %s) FAILED, error: %v", key, err)
		return err
	}
	klog.Infof("DelSubnet: delete subnet[id: %s] SUCC", subnetID)
	return nil
}

var UnmarshalSubnet = func(value []byte) (*Subnet, error) {
	var subnet Subnet
	err := json.Unmarshal([]byte(value), &subnet)
	if err != nil {
		klog.Errorf("UnmarshalSubnet: json.Unmarshal(%s) FAILED, error: %v", string(value), err)
		return nil, errobj.ErrUnmarshalFailed
	}

	klog.Infof("UnmarshalSubnet: subnet[%v] SUCC", subnet)
	return &subnet, nil
}

func GetSubnet(subnetID string) (*Subnet, error) {
	key := createSubnetKey(subnetID)
	value, err := common.GetDataBase().ReadLeaf(key)
	if err != nil {
		klog.Errorf("GetSubnet: ReadLeaf(key: %s) FAILED, error: %v", key, err)
		return nil, err
	}

	subnet, err := UnmarshalSubnet([]byte(value))
	if err != nil {
		klog.Errorf("GetSubnet: UnmarshalSubnet(%v) FAILED, error: %v", value, err)
		return nil, errobj.ErrUnmarshalFailed
	}

	klog.Infof("GetSubnet: get subnet[%v] SUCC", subnet)
	return subnet, nil
}

func GetAllSubnets() ([]*Subnet, error) {
	key := getSubnetsKey()
	nodes, err := common.GetDataBase().ReadDir(key)
	if err != nil {
		klog.Errorf("GetAllSubnets: ReadDir(key: %s) FAILED, error: %v", key, err)
		return nil, err
	}

	subnets := make([]*Subnet, 0)
	for _, node := range nodes {
		subnet, err := UnmarshalSubnet([]byte(node.Value))
		if err != nil {
			klog.Errorf("GetAllSubnets: UnmarshalSubnet(net: %s) FAILED, error: %v", node.Value, err)
			return nil, err
		}
		subnets = append(subnets, subnet)
	}

	klog.Tracef("GetAllSubnets: get all logical networks: %v SUCC", subnets)
	return subnets, nil
}

type SubnetObject struct {
	ID         string           `json:"id"`
	NetworkID  string           `json:"network_id"`
	Name       string           `json:"name"`
	CIDR       string           `json:"cidr"`
	GatewayIP  string           `json:"gateway_ip"`
	TenantID   string           `json:"tenant_id"`
	AllocPools []AllocationPool `json:"allocation_pools"`
}

type SubnetObjectRepo struct {
	indexer cache.Indexer
}

var subnetObjRepo SubnetObjectRepo

func GetSubnetObjRepoSingleton() *SubnetObjectRepo {
	return &subnetObjRepo
}

func SubnetObjKeyFunc(obj interface{}) (string, error) {
	if obj == nil {
		klog.Error("SunbetObjKeyFunc: obj arg is nil")
		return "", errobj.ErrObjectPointerIsNil
	}

	subnetObj, ok := obj.(*SubnetObject)
	if !ok {
		klog.Error("SunbetObjKeyFunc: obj arg is not type: *SubnetObject")
		return "", errobj.ErrArgTypeMismatch
	}

	return subnetObj.ID, nil
}

func SubnetNetworkIDIndexFunc(obj interface{}) ([]string, error) {
	if obj == nil {
		klog.Error("SubnetNetworkIDIndexFunc: obj arg is nil")
		return nil, errobj.ErrObjectPointerIsNil
	}

	subnetObj, ok := obj.(*SubnetObject)
	if !ok {
		klog.Error("SubnetNetworkIDIndexFunc: obj arg is not type: *SubnetObject")
		return nil, errobj.ErrArgTypeMismatch
	}

	return []string{subnetObj.NetworkID}, nil
}

const (
	SubnetNetworkIDIndex = "network_id"
)

func (p *SubnetObjectRepo) Init() {
	indexers := cache.Indexers{
		SubnetNetworkIDIndex: SubnetNetworkIDIndexFunc}
	p.indexer = cache.NewIndexer(SubnetObjKeyFunc, indexers)
}

func (p *SubnetObjectRepo) Add(subnetObj *SubnetObject) error {
	err := p.indexer.Add(subnetObj)
	if err != nil {
		klog.Errorf("SubnetObjectRepo.Add: add obj[%v] to repo FAILED, error: %v", subnetObj, err)
		return err
	}
	klog.Infof("SubnetObjectRepo.Add: add obj[%v] to repo SUCC", subnetObj)
	return nil
}

func (p *SubnetObjectRepo) Del(ID string) error {
	err := p.indexer.Delete(ID)
	if err != nil {
		klog.Errorf("SubnetObjectRepo.Del: Delete(subnetID: %s) FAILED, error: %v", ID, err)
		return err
	}

	klog.Infof("SubnetObjectRepo.Del: Delete(subnetID: %s) SUCC", ID)
	return nil
}

func (p *SubnetObjectRepo) Get(ID string) (*SubnetObject, error) {
	item, exists, err := p.indexer.GetByKey(ID)
	if err != nil {
		klog.Errorf("SubnetObjectRepo.Get: subnetID[%s]'s object FAILED, error: %v", ID, err)
		return nil, err
	}
	if !exists {
		klog.Errorf("SubnetObjectRepo.Get: subnetID[%s]'s object not found", ID)
		return nil, errobj.ErrRecordNotExist
	}

	subnetObj, ok := item.(*SubnetObject)
	if !ok {
		klog.Errorf("SubnetObjectRepo.Get: subnetID[%s]'s object[%v] type not match SubnetObject", ID, item)
		return nil, errobj.ErrObjectTypeMismatch
	}
	klog.Infof("SubnetObjectRepo.Get: subnetID[%s]'s object[%v] SUCC", ID, subnetObj)
	return subnetObj, nil
}

func (p *SubnetObjectRepo) Update(subnetObj *SubnetObject) error {
	err := p.indexer.Update(subnetObj)
	if err != nil {
		klog.Errorf("SubnetObjectRepo.Update: Update[%v] FAILED, error: %v", subnetObj, err)
	}

	klog.Infof("SubnetObjectRepo.Update: Update[%v] SUCC", subnetObj)
	return nil
}

func (p *SubnetObjectRepo) listByIndex(indexName, indexKey string) ([]*SubnetObject, error) {
	objs, err := p.indexer.ByIndex(indexName, indexKey)
	if err != nil {
		klog.Errorf("SubnetObjectRepo.listByIndex: ByIndex(indexName: %s, indexKey: %s) FAILED, error: %v",
			indexName, indexKey, err)
		return nil, err
	}

	klog.Infof("SubnetObjectRepo.listByIndex: ByIndex(indexName: %s, indexKey: %s) SUCC, array: %v",
		indexName, indexKey, objs)
	subnetObjs := make([]*SubnetObject, 0)
	for _, obj := range objs {
		subnetObj, ok := obj.(*SubnetObject)
		if !ok {
			klog.Errorf("NetworkObjectRepo.listByIndex: index result object: %v is not type *SubnetObject, skip", obj)
			continue
		}
		subnetObjs = append(subnetObjs, subnetObj)
	}
	return subnetObjs, nil
}

func (p *SubnetObjectRepo) ListByNetworID(networkID string) ([]*SubnetObject, error) {
	return p.listByIndex(SubnetNetworkIDIndex, networkID)
}

func init() {
	GetSubnetObjRepoSingleton().Init()
}

func TransSubnetToSubnetObject(subnet *Subnet) *SubnetObject {
	return &SubnetObject{
		ID:         subnet.ID,
		NetworkID:  subnet.NetworkID,
		Name:       subnet.Name,
		CIDR:       subnet.CIDR,
		GatewayIP:  subnet.GatewayIP,
		TenantID:   subnet.TenantID,
		AllocPools: subnet.AllocPools,
	}
}

func LoadAllSubnetObjects() error {
	subnets, err := GetAllSubnets()
	if err != nil {
		klog.Errorf("LoadAllSubnetObjects: GetAllSubnets FAILED, error: %v", err)
		return err
	}

	for _, subnet := range subnets {
		subnetObj := TransSubnetToSubnetObject(subnet)
		err = GetSubnetObjRepoSingleton().Add(subnetObj)
		if err != nil {
			klog.Errorf("LoadAllSubnetObjects: GetSubnetObjRepoSingleton().Add(subnetObj: %v) FAILED, error: %v",
				subnetObj, err)
			return err
		}
		klog.Tracef("LoadAllSubnetObjects: GetSubnetObjRepoSingleton().Add(subnetObj: %v) SUCC",
			subnetObj)
	}
	return nil
}
