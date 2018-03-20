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

package knittermgrrole

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
)

type NetworkIDRole struct {
}

func (this *NetworkIDRole) Get(tenantName, networkName string) (string, error) {
	agtCtx := cni.GetGlobalContext()
	url := agtCtx.Mc.GetNetworkURL(tenantName, networkName)
	statCode, networkInfoByte, err := agtCtx.Mc.Get(url)
	if err != nil || statCode != 200 {
		klog.Errorf("Get network id error! %v", err)
		return "", err
	}
	networkInfoJSON, err := jason.NewObjectFromBytes(networkInfoByte)
	if err != nil {
		klog.Errorf("NetworkIdRole:Get:jason.NewObjectFromBytes err: %v, networkInfoByte: %v",
			err, networkInfoByte)
		return "", errobj.ErrJasonNewObjectFailed
	}
	networkID, err := networkInfoJSON.GetString("network_id")
	if err != nil {
		klog.Errorf("Get network id error! %v", err)
		return "", errobj.ErrJasonGetStringFailed
	}
	return networkID, nil
}
