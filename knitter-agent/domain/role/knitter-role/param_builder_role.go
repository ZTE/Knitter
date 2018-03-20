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

package knitterrole

import (
	"encoding/json"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/containernetworking/cni/pkg/skel"
)

type ParamBuilderRole struct {
}

func (this *ParamBuilderRole) Transform(reqBody []byte) (*cni.CniParam, *skel.CmdArgs, error) {
	klog.Infof("ParamBuilderRole.Transform: start  reqbody is %v", string(reqBody))
	var args *skel.CmdArgs
	err := json.Unmarshal(reqBody, &args)
	if err != nil {
		klog.Error("Unmarshal reqBody to skel.CmdArgs ERROR: ", err)
		return nil, nil, err
	}

	cniParam := &cni.CniParam{}
	err = cniParam.AnalyzeCniParam(args)
	if err != nil {
		klog.Errorf("Analyze CNI param ERROR!")
		return nil, nil, err
	}
	klog.Infof("Analyze CNI param succ: %v\n", *cniParam)
	klog.Infof("ParamBuilderRole.Transform: SUCC ", string(reqBody))
	return cniParam, args, nil

}
