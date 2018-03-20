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

package knitteragentrole

import (
	"bytes"
	"encoding/json"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/containernetworking/cni/pkg/skel"
	"net/http"
)

type DetachTriggerRole struct {
}

const KnitterAgnent string = "http://127.0.0.1:6006/v1/pod"

func (this DetachTriggerRole) SendPostReqToSelf(args *skel.CmdArgs) error {
	reqURL := KnitterAgnent + "?operation=detach"
	klog.Infof("request url=%v", reqURL)
	bodyType := "application/json"
	reqJSON, err := json.Marshal(args)
	if err != nil {
		klog.Error("DetachTriggerRole:SendPostReqToSelf Marshal cni skel.CmdArgs Error:", err)
		return err
	}
	postReader := bytes.NewReader([]byte(reqJSON))
	resp, err := http.Post(reqURL, bodyType, postReader)
	if err != nil {
		klog.Errorf("DetachTriggerRole:SendPostReqToSelf KnitterAgent post err: %v", err)
		return errobj.ErrHTTPPostFailed
	}
	defer resp.Body.Close()

	klog.Info("DetachTriggerRole:SendPostReqToSelf success!")
	return nil
}
