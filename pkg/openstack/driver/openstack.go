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

package driver

import (
	"encoding/json"
	"github.com/ZTE/Knitter/pkg/klog"
	"runtime"
	"sync"
)

var opInstance *OpenStack
var opOnce sync.Once

func NewOpenstack() *OpenStack {
	opOnce.Do(func() {
		opInstance = &OpenStack{}
	})

	return opInstance
}

type OpenStack struct {
	NeutronClient
	NovaClient
}

type OpenStackConf struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Url        string `json:"url"`
	Tenantid   string `json:"tenantid"`
	TenantName string `json:"tenantname"`
}

func (self *OpenStack) SetOpenstackConfig(cfgStr string) error {
	defer func() {
		if err := recover(); err != nil {
			stackInfo := make([]byte, 2048)
			runtime.Stack(stackInfo, false)
			klog.Info("IaaS-SetOpenstackConfig-panic:", string(stackInfo))
		}
	}()

	klog.Info("IaaS-SetOpenstackConfig-begin")
	auth := getAuthSingleton()
	klog.Info("IaaS-SetOpenstackConfig-GetAuthSingleton-end")
	config := OpenStackConf{}
	err := json.Unmarshal([]byte(cfgStr), &config)
	if err != nil {
		klog.Info("IaaS-SetOpenstackConfig-Unmarshal-error:", err)
		return err
	}

	klog.Info("IaaS-SetOpenstackConfig-Unmarshal-end:", config)
	auth.setConf(config)

	klog.Info("IaaS-SetOpenstackConfig-end:", auth)
	return nil
}

func (self *OpenStack) Auth() error {
	return getAuthSingleton().auth()
}

func (self *OpenStack) GetTenantUUID(cfg string) (string, error) {
	return getAuthSingleton().TenantID, nil
}

func (self *OpenStack) GetType() string {
	return "TECS"
}
