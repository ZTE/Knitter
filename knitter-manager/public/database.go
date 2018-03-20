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

package common

import (
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/uuid"
	"strings"
	"time"
)

var etcdDataBase dbaccessor.DbAccessor = nil
var managerUUID = ""

func SetDataBase(i dbaccessor.DbAccessor) error {
	etcdDataBase = i
	return nil
}

var GetDataBase = func() dbaccessor.DbAccessor {
	return etcdDataBase
}

func SetManagerUUID(id string) error {
	managerUUID = id
	return nil
}

func GetManagerUUID() string {
	return managerUUID
}

const ErrorKeyNotFound string = "Key not found"

var uuidPaaS string

func SetPaaSID() {
	uuidPaaS = GetPaasUUID()
}
func GetPaaSID() string {
	return uuidPaaS
}
func GetPaasUUID() string {
	key := dbaccessor.GetKeyOfPaasUUID()
	id, err := etcdDataBase.ReadLeaf(key)
	if err == nil {
		return id
	}
	desnotExist := strings.Contains(err.Error(), ErrorKeyNotFound)
	if !desnotExist {
		return uuid.NIL.String()
	}
	id = uuid.NewUUID()
	err = etcdDataBase.SaveLeaf(key, id)
	if err == nil {
		return id
	}
	return uuid.NIL.String()
}

func CheckDB() error {
	for {
		/*wait for etcd database OK*/
		e := dbaccessor.CheckDataBase(etcdDataBase)
		if e != nil {
			klog.Error("DataBase Config is ERROR", e)
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	for {
		SetPaaSID()
		if GetPaaSID() != uuid.NIL.String() {
			break
		}
	}

	return nil
}

func GetOpenstackCfg() string {
	key := dbaccessor.GetKeyOfOpenstack()
	cfg, err := etcdDataBase.ReadLeaf(key)
	if err != nil {
		return ""
	}
	return cfg
}

func GetVnfmCfg() string {
	key := dbaccessor.GetKeyOfVnfm()
	cfg, err := etcdDataBase.ReadLeaf(key)
	if err != nil {
		return ""
	}
	return cfg
}

func RegisterSelfToDb(serviceURL string) error {
	selfURL := dbaccessor.GetKeyOfKnitterManagerUrl()
	// local listen port should be READ from beego
	err := etcdDataBase.SaveLeaf(selfURL, serviceURL)
	if err != nil {
		klog.Error("paas-cni-manager register self to database",
			" accessor failed, error: ", err)
		return err
	}

	tmpStr := strings.TrimPrefix(serviceURL, "http://")
	tmpStr1 := strings.SplitAfter(tmpStr, ":")
	tmpStr2 := strings.TrimSuffix(tmpStr1[0], ":")
	SetManagerUUID(tmpStr2)
	return nil
}
