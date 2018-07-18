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

type IPGroupObject struct {
	ID        string
	Name      string
	NetworkID string
	IPs       []IPInDB
	TenantID  string
}

type IPGroupObjectRepo struct {
	indexer cache.Indexer
}

var igObjRepo IPGroupObjectRepo

func GetIPGroupObjRepoSingleton() *IPGroupObjectRepo {
	return &igObjRepo
}

func IPGroupObjKeyFunc(obj interface{}) (string, error) {
	if obj == nil {
		klog.Error("IPGroupObjKeyFunc: obj arg is nil")
		return "", errobj.ErrObjectPointerIsNil
	}

	if igObj, ok := obj.(string); ok {
		return igObj, nil
	}

	igObj, ok := obj.(*IPGroupObject)
	if !ok {
		klog.Error("IPGroupObjKeyFunc: obj arg is not type: *IPGroupObject")
		return "", errobj.ErrArgTypeMismatch
	}

	return igObj.ID, nil
}

func IPGroupTenantIDIndexFunc(obj interface{}) ([]string, error) {
	if obj == nil {
		klog.Error("IPGroupTenantIDIndexFunc: obj arg is nil")
		return nil, errobj.ErrObjectPointerIsNil
	}

	igObj, ok := obj.(*IPGroupObject)
	if !ok {
		klog.Error("IPGroupTenantIDIndexFunc: obj arg is not type: *IPGroupObject")
		return nil, errobj.ErrArgTypeMismatch
	}

	return []string{igObj.TenantID}, nil
}

func IPGroupNetworkIDIndexFunc(obj interface{}) ([]string, error) {
	if obj == nil {
		klog.Error("IPGroupNetworkIDIndexFunc: obj arg is nil")
		return nil, errobj.ErrObjectPointerIsNil
	}

	igObj, ok := obj.(*IPGroupObject)
	if !ok {
		klog.Error("IPGroupNetworkIDIndexFunc: obj arg is not type: *IPGroupObject")
		return nil, errobj.ErrArgTypeMismatch
	}

	return []string{igObj.NetworkID}, nil
}

const (
	IPGroupTenantIDIndex  = "tenant_id"
	IPGroupNetworkIDIndex = "network_id"
)

func (p *IPGroupObjectRepo) Init() {
	indexers := cache.Indexers{
		IPGroupTenantIDIndex:  IPGroupTenantIDIndexFunc,
		IPGroupNetworkIDIndex: IPGroupNetworkIDIndexFunc}
	p.indexer = cache.NewIndexer(IPGroupObjKeyFunc, indexers)
}

func init() {
	GetIPGroupObjRepoSingleton().Init()
}

func (p *IPGroupObjectRepo) Add(igObj *IPGroupObject) error {
	err := p.indexer.Add(igObj)
	if err != nil {
		klog.Errorf("IPGroupObjectRepo.Add: add obj[%v] to repo FAILED, error: %+v", igObj, err)
		return err
	}
	klog.Infof("IPGroupObjectRepo.Add: add obj[%+v] to repo SUCC", igObj)
	return nil
}

func (p *IPGroupObjectRepo) Del(ID string) error {
	err := p.indexer.Delete(ID)
	if err != nil {
		klog.Errorf("IPGroupObjectRepo.Del: Delete(netID: %s) FAILED, error: %v", ID, err)
		return err
	}

	klog.Infof("IPGroupObjectRepo.Del: Delete(netID: %s) SUCC", ID)
	return nil
}

func (p *IPGroupObjectRepo) Get(ID string) (*IPGroupObject, error) {
	item, exists, err := p.indexer.GetByKey(ID)
	if err != nil {
		klog.Errorf("IPGroupObjectRepo.Get: netID[%s]'s object FAILED, error: %v", ID, err)
		return nil, err
	}
	if !exists {
		klog.Errorf("IPGroupObjectRepo.Get: netID[%s]'s object not found", ID)
		return nil, errobj.ErrRecordNotExist
	}

	igObj, ok := item.(*IPGroupObject)
	if !ok {
		klog.Errorf("IPGroupObjectRepo.Get: netID[%s]'s object[%v] type not match IPGroupObject", ID, item)
		return nil, errobj.ErrObjectTypeMismatch
	}
	klog.Infof("IPGroupObjectRepo.Get: netID[%s]'s object[%v] SUCC", ID, igObj)
	return igObj, nil
}

func (p *IPGroupObjectRepo) Update(igObj *IPGroupObject) error {
	err := p.indexer.Update(igObj)
	if err != nil {
		klog.Errorf("IPGroupObjectRepo.Update: Update[%v] FAILED, error: %v", igObj, err)
	}

	klog.Infof("IPGroupObjectRepo.Update: Update[%v] SUCC", igObj)
	return nil
}

func (p *IPGroupObjectRepo) listByIndex(indexName, indexKey string) ([]*IPGroupObject, error) {
	objs, err := p.indexer.ByIndex(indexName, indexKey)
	if err != nil {
		klog.Errorf("IPGroupObjectRepo.listByIndex: ByIndex(indexName: %s, indexKey: %s) FAILED, error: %v",
			indexName, indexKey, err)
		return nil, err
	}

	klog.Infof("IPGroupObjectRepo.listByIndex: ByIndex(indexName: %s, indexKey: %s) SUCC, array: %v",
		indexName, indexKey, objs)
	igObjs := make([]*IPGroupObject, 0)
	for _, obj := range objs {
		igObj, ok := obj.(*IPGroupObject)
		if !ok {
			klog.Errorf("IPGroupObjectRepo.listByIndex: index result object: %v is not type *IPGroupObject, skip", obj)
			continue
		}
		igObjs = append(igObjs, igObj)
	}
	return igObjs, nil
}

func (p *IPGroupObjectRepo) List() []*IPGroupObject {
	objs := p.indexer.List()
	igObjs := []*IPGroupObject{}

	for _, obj := range objs {
		igObj, ok := obj.(*IPGroupObject)
		if !ok {
			klog.Errorf("IPGroupObjectRepo.List: List result object: %v is not type *IPGroupObject, skip", obj)
			continue
		}
		igObjs = append(igObjs, igObj)
	}
	return igObjs
}

func (p *IPGroupObjectRepo) ListByNetworkID(networkName string) ([]*IPGroupObject, error) {
	return p.listByIndex(IPGroupNetworkIDIndex, networkName)
}

func (p *IPGroupObjectRepo) ListByTenantID(tenantID string) ([]*IPGroupObject, error) {
	return p.listByIndex(IPGroupTenantIDIndex, tenantID)
}

func GetAllIPGroups() ([]*IPGroupInDB, error) {
	key := getIPGroupsKey()
	nodes, err := common.GetDataBase().ReadDir(key)
	if err != nil {
		klog.Errorf("GetAllIPGroups: ReadDir(key: %s) FAILED, error: %v", key, err)
		return nil, err
	}

	igs := make([]*IPGroupInDB, 0)
	for _, node := range nodes {
		klog.Infof("GetAllIPGroups: db raw ip group string: [%s]", node.Value)
		net, err := UnmarshalIPGroup([]byte(node.Value))
		if err != nil {
			klog.Errorf("GetAllIPGroups: UnmarshalIPGroup(net: %s) FAILED, error: %v", node.Value, err)
			return nil, err
		}
		klog.Infof("GetAllIPGroups: get new netObj: %+v", net)
		igs = append(igs, net)
	}

	klog.Tracef("GetAllIPGroups: get all logical ip groups: %v SUCC", igs)
	return igs, nil
}

var UnmarshalIPGroup = func(value []byte) (*IPGroupInDB, error) {
	var ig IPGroupInDB
	err := json.Unmarshal([]byte(value), &ig)
	if err != nil {
		klog.Errorf("UnmarshalIPGroup: json.Unmarshal(%s) FAILED, error: %v", string(value), err)
		return nil, errobj.ErrUnmarshalFailed
	}

	klog.Infof("UnmarshalIPGroup: ip group[%v] SUCC", ig)
	return &ig, nil
}

func LoadAllIPGroupObjects() error {
	igs, err := GetAllIPGroups()
	if err != nil {
		klog.Errorf("LoadAllIPGroupObjects: GetAllNetworks FAILED, error: %v", err)
		return err
	}

	for _, ig := range igs {
		igObj := TransIGInDBToIGObject(ig)
		err = GetIPGroupObjRepoSingleton().Add(igObj)
		if err != nil {
			klog.Errorf("LoadAllIPGroupObjects: GetIPGroupObjRepoSingleton().Add(igObj: %+v) FAILED, error: %v", igObj, err)
			return err
		}
		klog.Infof("LoadAllIPGroupObjects: GetIPGroupObjRepoSingleton().Add(igObj: %+v) SUCC", igObj)
	}
	return nil
}
