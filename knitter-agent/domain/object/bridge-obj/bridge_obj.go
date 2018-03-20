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

package bridgeobj

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/role/bridge-role"
	"github.com/ZTE/Knitter/pkg/klog"
)

type BridgeObj struct {
	BrintRole bridgerole.BrintRole
	BrtunRole bridgerole.BrtunRole
}

var bridgeObjSingleton *BridgeObj

func GetBridgeObjSingleton() *BridgeObj {
	if bridgeObjSingleton != nil {
		return bridgeObjSingleton
	}
	bridgeObjSingleton = &BridgeObj{}
	err := bridgeObjSingleton.BrintRole.LoadTenantNetworkTable()
	if err != nil {
		klog.Errorf("loadMemFileToCache err: %v", err)
	}

	return bridgeObjSingleton
}
