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
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/pkg/klog"
)

type VxLanNetWork struct {
	NetworkType string `json:"network_type"`
	NetworkID   string `json:"network_id"`
	Vni         string `json:"vni"`
}

// todo
func GetNetworkInfo(user, networkName string, needProvider bool) (*PaasNetwork, error) {
	//paasNet, err := network.GetNetworkByName()
	paasNet, err := GetNetworkByName(user, networkName)
	if err != nil {
		klog.Errorf("Get network info of network[name: %s]"+
			"failed, error: %v", networkName, err)
		return nil, fmt.Errorf("%v:Get network info error", err)
	}
	if !needProvider {
		klog.Infof("Get network info of tenantID: %s network[name: %s] successfully.", user, networkName)
		return paasNet, nil
	}

	if paasNet.Provider.NetworkType != "vxlan" && paasNet.Provider.NetworkType != "vlan" &&
		paasNet.Provider.NetworkType != "flat" {
		klog.Error("GetNetworkInfo: ", err.Error(), ":", paasNet.Provider.NetworkType)
		klog.Errorf("GetNetworkInfo: provider network type is: %s, not supported", paasNet.Provider.NetworkType)
		return nil, errobj.ErrNetworkTypeNotSupported
	}

	klog.Infof("Get network info of network[name: %s] successfully.", networkName)
	return paasNet, nil
}

func GetNetworkVni(paasTenantID, networkID string) (*VxLanNetWork, error) {
	netExt, err := iaas.GetIaaS(paasTenantID).GetNetworkExtenAttrs(networkID)
	if err != nil {
		klog.Errorf("GetNetworkVni: GetIaaS().GetNetworkExtenAttrs error: %v",
			err)
		return nil, fmt.Errorf("%v:Get VNI from IaaS error", err)
	}

	if netExt.NetworkType == "vxlan" || netExt.NetworkType == "vlan" {
		klog.Info("NetworkType:", netExt.NetworkType)
		vxLanInfo := &VxLanNetWork{NetworkType: "vxlan",
			NetworkID: networkID, Vni: netExt.SegmentationID}
		klog.Info("Respond-Body:", vxLanInfo)
		return vxLanInfo, nil
	}
	err = errors.New("networktype type unsupported")
	klog.Error("GetNetworkVni: ", err.Error(),
		":", netExt.NetworkType)
	return nil, err
}

var DefaultPaaSNetwork string = constvalue.DefaultPaaSNetwork

func GetDefaultNetworkName() string {
	return DefaultPaaSNetwork
}
