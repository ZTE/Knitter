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
	"strconv"

	"k8s.io/client-go/tools/cache"

	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/pkg/klog"
)

type ExtenAttrs struct {
	// Specifies the nature of the physical network mapped to this network
	// resource. Examples are flat, vlan, vxlan, or gre.
	NetworkType string `json:"provider:network_type"`

	// Identifies the physical network on top of which this network object is
	// being implemented. The OpenStack Networking API does not expose any facility
	// for retrieving the list of available physical networks. As an example, in
	// the Open vSwitch plug-in this is a symbolic name which is then mapped to
	// specific bridges on each compute host through the Open vSwitch plug-in
	// configuration file.
	PhysicalNetwork string `json:"provider:physical_network"`

	// Identifies an isolated segment on the physical network; the nature of the
	// segment depends on the segmentation model defined by network_type. For
	// instance, if network_type is vlan, then this is a vlan identifier;
	// otherwise, if network_type is gre, then this will be a gre key.
	SegmentationID string `json:"provider:segmentation_id"`

	VlanTransparent bool `json:"vlan_transparent"`
}

type Network struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	//Status      string `json:"state"`

	//SubnetIDs []string
	SubnetID string

	ExtAttrs ExtenAttrs

	TenantID    string `json:"tenant_id"`
	IsPublic    bool   `json:"is_public"`
	IsExternal  bool   `json:"is_external"`
	CreateTime  string `json:"create_time"`
	Description string `json:"description"`
}

func getNetworksKey() string {
	return KnitterManagerKeyRoot + "/networks"
}

func createNetworkKey(networkID string) string {
	return getNetworksKey() + "/" + networkID
}

func SaveNetwork(net *Network) error {
	netInBytes, err := json.Marshal(net)
	if err != nil {
		klog.Errorf("SaveNetwork: json.Marshal(network: %v) FAILED, error: %v", net, err)
		return errobj.ErrMarshalFailed
	}
	key := createNetworkKey(net.ID)
	err = common.GetDataBase().SaveLeaf(key, string(netInBytes))
	if err != nil {
		klog.Errorf("SaveNetwork: SaveLeaf(key: %s, value: %s) FAILED, error: %v", key, string(netInBytes), err)
		return err
	}
	klog.Infof("SaveNetwork: save network[%v] SUCC", net)
	return nil
}

func DelNetwork(netID string) error {
	key := createNetworkKey(netID)
	err := common.GetDataBase().DeleteLeaf(key)
	if err != nil && !IsKeyNotFoundError(err) {
		klog.Errorf("DelNetwork: DeleteLeaf(key: %s) FAILED, error: %v", key, err)
		return err
	}
	klog.Infof("DelNetwork: delete network[id: %s] SUCC", netID)
	return nil
}

var UnmarshalNetwork = func(value []byte) (*Network, error) {
	var net Network
	err := json.Unmarshal([]byte(value), &net)
	if err != nil {
		klog.Errorf("UnmarshalNetwork: json.Unmarshal(%s) FAILED, error: %v", string(value), err)
		return nil, errobj.ErrUnmarshalFailed
	}

	klog.Infof("UnmarshalNetwork: net[%v] SUCC", net)
	return &net, nil
}

func GetNetwork2(netID string) (*Network, error) {
	key := createNetworkKey(netID)
	value, err := common.GetDataBase().ReadLeaf(key)
	if err != nil {
		klog.Errorf("GetNetwork: ReadLeaf(key: %s) FAILED, error: %v", key, err)
		return nil, err
	}

	net, err := UnmarshalNetwork([]byte(value))
	if err != nil {
		klog.Errorf("GetNetwork: UnmarshalNetwork(%v) FAILED, error: %v", value, err)
		return nil, errobj.ErrUnmarshalFailed
	}

	klog.Infof("GetNetwork: get network[%v] SUCC", net)
	return net, nil
}

func GetAllNetworks() ([]*Network, error) {
	key := getNetworksKey()
	nodes, err := common.GetDataBase().ReadDir(key)
	if err != nil {
		klog.Errorf("GetAllNetworks: ReadDir(key: %s) FAILED, error: %v", key, err)
		return nil, err
	}

	networks := make([]*Network, 0)
	for _, node := range nodes {
		klog.Infof("GetAllNetworks: db raw network string: [%s]", node.Value)
		net, err := UnmarshalNetwork([]byte(node.Value))
		if err != nil {
			klog.Errorf("GetAllNetworks: UnmarshalNetwork(net: %s) FAILED, error: %v", node.Value, err)
			return nil, err
		}
		klog.Infof("GetAllNetworks: get new netObj: %+v", net)
		networks = append(networks, net)
	}

	klog.Tracef("GetAllNetworks: get all logical networks: %v SUCC", networks)
	return networks, nil
}

type NetworkObject struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Status string `json:"state"`

	//SubnetIDs []string
	SubnetID string
	ExtAttrs ExtenAttrs

	TenantID    string `json:"tenant_id"`
	IsPublic    bool   `json:"is_public"`
	IsExternal  bool   `json:"is_external"`
	CreateTime  string `json:"create_time"`
	Description string `json:"description"`
}

type NetworkObjectRepo struct {
	indexer cache.Indexer
}

var netObjRepo NetworkObjectRepo

func GetNetObjRepoSingleton() *NetworkObjectRepo {
	return &netObjRepo
}

func NetworkObjKeyFunc(obj interface{}) (string, error) {
	if obj == nil {
		klog.Error("NetworkObjKeyFunc: obj arg is nil")
		return "", errobj.ErrObjectPointerIsNil
	}

	NetObj, ok := obj.(*NetworkObject)
	if !ok {
		klog.Error("NetworkObjKeyFunc: obj arg is not type: *NetworkObject")
		return "", errobj.ErrArgTypeMismatch
	}

	return NetObj.ID, nil
}

func NetworkNameIndexFunc(obj interface{}) ([]string, error) {
	if obj == nil {
		klog.Error("NetworkNameIndexFunc: obj arg is nil")
		return nil, errobj.ErrObjectPointerIsNil
	}

	netObj, ok := obj.(*NetworkObject)
	if !ok {
		klog.Error("NetworkNameIndexFunc: obj arg is not type: *NetworkObject")
		return nil, errobj.ErrArgTypeMismatch
	}

	return []string{netObj.Name}, nil
}

func NetworkTenantIDIndexFunc(obj interface{}) ([]string, error) {
	if obj == nil {
		klog.Error("NetworkTenantIDIndexFunc: obj arg is nil")
		return nil, errobj.ErrObjectPointerIsNil
	}

	netObj, ok := obj.(*NetworkObject)
	if !ok {
		klog.Error("NetworkTenantIDIndexFunc: obj arg is not type: *NetworkObject")
		return nil, errobj.ErrArgTypeMismatch
	}

	return []string{netObj.TenantID}, nil
}

func NetworkIsPlublicIndexFunc(obj interface{}) ([]string, error) {
	if obj == nil {
		klog.Error("NetworkIsPlublicIndexFunc: obj arg is nil")
		return nil, errobj.ErrObjectPointerIsNil
	}

	netObj, ok := obj.(*NetworkObject)
	if !ok {
		klog.Error("NetworkIsPlublicIndexFunc: obj arg is not type: *NetworkObject")
		return nil, errobj.ErrArgTypeMismatch
	}

	return []string{strconv.FormatBool(netObj.IsPublic)}, nil
}

func NetworkIsExternalIndexFunc(obj interface{}) ([]string, error) {
	if obj == nil {
		klog.Error("NetworkIsPlublicIndexFunc: obj arg is nil")
		return nil, errobj.ErrObjectPointerIsNil
	}

	netObj, ok := obj.(*NetworkObject)
	if !ok {
		klog.Error("NetworkIsPlublicIndexFunc: obj arg is not type: *NetworkObject")
		return nil, errobj.ErrArgTypeMismatch
	}

	return []string{strconv.FormatBool(netObj.IsExternal)}, nil
}

const (
	NetworkNameIndex       = "name"
	NetworkTenantIDIndex   = "tenant_id"
	NetworkIsPublicIndex   = "is_public"
	NetworkIsExternalIndex = "is_external"
)

func (p *NetworkObjectRepo) Init() {
	indexers := cache.Indexers{
		NetworkNameIndex:       NetworkNameIndexFunc,
		NetworkTenantIDIndex:   NetworkTenantIDIndexFunc,
		NetworkIsPublicIndex:   NetworkIsPlublicIndexFunc,
		NetworkIsExternalIndex: NetworkIsExternalIndexFunc}
	p.indexer = cache.NewIndexer(NetworkObjKeyFunc, indexers)
}

func (p *NetworkObjectRepo) Add(netObj *NetworkObject) error {
	err := p.indexer.Add(netObj)
	if err != nil {
		klog.Errorf("NetworkObjectRepo.Add: add obj[%v] to repo FAILED, error: %+v", netObj, err)
		return err
	}
	klog.Infof("NetworkObjectRepo.Add: add obj[%+v] to repo SUCC", netObj)
	return nil
}

func (p *NetworkObjectRepo) Del(ID string) error {
	err := p.indexer.Delete(ID)
	if err != nil {
		klog.Errorf("NetworkObjectRepo.Del: Delete(netID: %s) FAILED, error: %v", ID, err)
		return err
	}

	klog.Infof("NetworkObjectRepo.Del: Delete(netID: %s) SUCC", ID)
	return nil
}

func (p *NetworkObjectRepo) Get(ID string) (*NetworkObject, error) {
	item, exists, err := p.indexer.GetByKey(ID)
	if err != nil {
		klog.Errorf("NetworkObjectRepo.Get: netID[%s]'s object FAILED, error: %v", ID, err)
		return nil, err
	}
	if !exists {
		klog.Errorf("NetworkObjectRepo.Get: netID[%s]'s object not found", ID)
		return nil, errobj.ErrRecordNotExist
	}

	netObj, ok := item.(*NetworkObject)
	if !ok {
		klog.Errorf("NetworkObjectRepo.Get: netID[%s]'s object[%v] type not match NetworkObject", ID, item)
		return nil, errobj.ErrObjectTypeMismatch
	}
	klog.Infof("NetworkObjectRepo.Get: netID[%s]'s object[%v] SUCC", ID, netObj)
	return netObj, nil
}

func (p *NetworkObjectRepo) Update(netObj *NetworkObject) error {
	err := p.indexer.Update(netObj)
	if err != nil {
		klog.Errorf("NetworkObjectRepo.Update: Update[%v] FAILED, error: %v", netObj, err)
	}

	klog.Infof("NetworkObjectRepo.Update: Update[%v] SUCC", netObj)
	return nil
}

func (p *NetworkObjectRepo) listByIndex(indexName, indexKey string) ([]*NetworkObject, error) {
	objs, err := p.indexer.ByIndex(indexName, indexKey)
	if err != nil {
		klog.Errorf("NetworkObjectRepo.listByIndex: ByIndex(indexName: %s, indexKey: %s) FAILED, error: %v",
			indexName, indexKey, err)
		return nil, err
	}

	klog.Infof("NetworkObjectRepo.listByIndex: ByIndex(indexName: %s, indexKey: %s) SUCC, array: %v",
		indexName, indexKey, objs)
	netObjs := make([]*NetworkObject, 0)
	for _, obj := range objs {
		netObj, ok := obj.(*NetworkObject)
		if !ok {
			klog.Errorf("NetworkObjectRepo.listByIndex: index result object: %v is not type *NetworkObject, skip", obj)
			continue
		}
		netObjs = append(netObjs, netObj)
	}
	return netObjs, nil
}

func (p *NetworkObjectRepo) List() []*NetworkObject {
	objs := p.indexer.List()
	netObjs := []*NetworkObject{}

	for _, obj := range objs {
		netObj, ok := obj.(*NetworkObject)
		if !ok {
			klog.Errorf("NetworkObjectRepo.List: List result object: %v is not type *NetworkObject, skip", obj)
			continue
		}
		netObjs = append(netObjs, netObj)
	}
	return netObjs
}

func (p *NetworkObjectRepo) ListByNetworkName(networkName string) ([]*NetworkObject, error) {
	return p.listByIndex(NetworkNameIndex, networkName)
}

func (p *NetworkObjectRepo) ListByTenantID(tenantID string) ([]*NetworkObject, error) {
	return p.listByIndex(NetworkTenantIDIndex, tenantID)
}

func (p *NetworkObjectRepo) ListByIsPublic(isPublic string) ([]*NetworkObject, error) {
	return p.listByIndex(NetworkIsPublicIndex, isPublic)
}

func (p *NetworkObjectRepo) ListByIsExternal(isExternal string) ([]*NetworkObject, error) {
	return p.listByIndex(NetworkIsExternalIndex, isExternal)
}

func init() {
	GetNetObjRepoSingleton().Init()
}

func TransNetworkToNetworkObject(net *Network) *NetworkObject {
	return &NetworkObject{
		Name:        net.Name,
		ID:          net.ID,
		SubnetID:    net.SubnetID,
		ExtAttrs:    net.ExtAttrs,
		TenantID:    net.TenantID,
		IsPublic:    net.IsPublic,
		IsExternal:  net.IsExternal,
		CreateTime:  net.CreateTime,
		Description: net.Description,
	}
}

func LoadAllNetworkObjects() error {
	networks, err := GetAllNetworks()
	if err != nil {
		klog.Errorf("LoadAllNetworkObjects: GetAllNetworks FAILED, error: %v", err)
		return err
	}

	for _, net := range networks {
		netObj := TransNetworkToNetworkObject(net)
		err = GetNetObjRepoSingleton().Add(netObj)
		if err != nil {
			klog.Errorf("LoadAllNetworkObjects: GetNetObjRepoSingleton().Add(netObj: %+v) FAILED, error: %v", netObj, err)
			return err
		}
		klog.Infof("LoadAllNetworkObjects: GetPortObjRepoSingleton().Add(netObj: %+v) SUCC", netObj)
	}
	return nil
}
