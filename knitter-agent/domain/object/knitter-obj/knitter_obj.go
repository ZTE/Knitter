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

package knitterobj

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/knitter-role"
	"github.com/containernetworking/cni/pkg/skel"
)

type KnitterObj struct {
	ParamBuilderRole  knitterrole.ParamBuilderRole
	PodProtectionRole knitterrole.PodProtectionRole
	CniParam          *cni.CniParam
	Args              *skel.CmdArgs
	ReqBody           []byte
}

var CreateKnitterObj = func(reqBody []byte) (*KnitterObj, error) {
	knitterObj := new(KnitterObj)
	var err error
	knitterObj.CniParam, knitterObj.Args, err = knitterObj.ParamBuilderRole.Transform(reqBody)
	knitterObj.PodProtectionRole.PodNs = knitterObj.CniParam.PodNs
	knitterObj.PodProtectionRole.PodName = knitterObj.CniParam.PodName
	knitterObj.PodProtectionRole.ContainerID = knitterObj.CniParam.ContainerID
	knitterObj.ReqBody = reqBody
	return knitterObj, err
}
