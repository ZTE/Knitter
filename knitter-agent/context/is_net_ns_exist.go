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
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"os"
	"strings"
)

type IsNetNsExist struct {
}

func (this *IsNetNsExist) Ok(transInfo *transdsl.TransInfo) bool {
	knitterObj := transInfo.AppInfo.(*KnitterInfo).KnitterObj
	netNsPath := strings.TrimRight(knitterObj.Args.Netns, "net")
	if netNsPath == "" {
		klog.Infof("***IsNetNsExist: false***")
		return false
	}
	_, err := os.Stat(netNsPath)
	if err == nil || os.IsExist(err) {
		klog.Infof("***IsNetNsExist: true***")
		return true
	}
	klog.Infof("***IsNetNsExist: false***")
	return false
}
