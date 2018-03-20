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

package osrole

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/ovs"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/vishvananda/netlink"
)

type VethPairRole struct {
}

func (this VethPairRole) Create(args ...string) (ovs.VethPair, error) {
	return ovs.CreateVethPair(args...)
}

func (this VethPairRole) Destroy(vethName string) {
	link, err := ovs.GetLinkByName(vethName)
	if err == nil {
		err = netlink.LinkDel(link)
		if err != nil {
			klog.Errorf("VethPairRole:Destroy:netlink.LinkDel err: %v", err)
		}
	} else {
		klog.Errorf("VethPairRole:Destroy:ovs.GetLinkByName err: %v", err)
	}

}
