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
	"github.com/ZTE/Knitter/knitter-agent/domain/bind"
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/pod-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/ovs"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/physical-resource-role"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-agt"

	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	"github.com/antonholmquist/jason"
	"github.com/coreos/etcd/client"
)

type KnitterInfo struct {
	ReqBody             []byte
	KnitterObj          *knitterobj.KnitterObj
	portObj             *portobj.PortObj
	podObj              *podobj.PodObj
	mgrPort             *manager.Port
	vethNameOk          bool
	vethPair            *ovs.VethPair
	Nics                []bind.Dpdknic
	Chan                chan int
	ChanFlag            bool
	ports               []*client.Node
	podJSON             *jason.Object
	DpdkLabel           string
	c0ImageName         string // knitter.go: getC0ImageName()
	ovsBr               string
	isBmAccelerateForC0 []bool
	South               bool   // used for baremetal south bound port name prefix
	c0PodLabel          string // used in detach procedure
	southIfs            map[int]*SouthInterface

	IsAttachOrDetachFlag   bool // when flag type is ture meaning Attach , is false meaning Detach
	isDetachPortErrorFlag  bool // used to avoid delete data if detach procedure meet error
	isDetachTransErrorFlag bool // used to avoid delete data if detach procedure meet error
}

func isDetachPortError(info *KnitterInfo) bool {
	return info.isDetachPortErrorFlag
}

func setDetachPortError(info *KnitterInfo) {
	info.isDetachPortErrorFlag = true
	setDetachTransError(info)
}

func isDetachTransError(info *KnitterInfo) bool {
	return info.isDetachTransErrorFlag
}

func setDetachTransError(info *KnitterInfo) {
	info.isDetachTransErrorFlag = true
}

type SouthInterface struct {
	repeatIdx int
	port      *mgragt.CreatePortInfo
	vnicRole  *physicalresourcerole.VnicRole
}
