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

package context

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/object/bridge-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra/concurrency_ctrl"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type IsTenantNetworkNeedDel struct {
}

func (this IsTenantNetworkNeedDel) Ok(transInfo *transdsl.TransInfo) bool {
	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	networkID := knitterInfo.portObj.LazyAttr.NetAttr.ID
	concurrencyctrl.ChanMapLock.Lock()
	value, ok := concurrencyctrl.ChanMap[networkID]
	if ok {
		knitterInfo.Chan = value
	} else {
		knitterInfo.Chan = make(chan int, 1)
		knitterInfo.Chan <- 1
		concurrencyctrl.ChanMap[networkID] = knitterInfo.Chan
	}
	concurrencyctrl.ChanMapLock.Unlock()

	<-knitterInfo.Chan
	knitterInfo.ChanFlag = true

	bridgeObj := bridgeobj.GetBridgeObjSingleton()
	flag := bridgeObj.BrintRole.NeedDelTenantNetworkTable(networkID)
	if flag {
		klog.Infof("***IsTenantNetworkNeedDel: true***")
	} else {
		knitterInfo.Chan <- 1
		knitterInfo.ChanFlag = false
		klog.Infof("***IsTenantNetworkNeedDel: false***")
	}
	return flag
}
