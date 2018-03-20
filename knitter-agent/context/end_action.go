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
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type EndAction struct {
	ErrCode error
}

func (this *EndAction) Exec(transInfo *transdsl.TransInfo) error {
	klog.Infof("***EndAction:Exec begin***")
	klog.Infof("***EndAction:Exec end***")
	if this.ErrCode != nil {
		return this.ErrCode
	}
	return errobj.ErrTransEnd
}

func (this *EndAction) RollBack(transInfo *transdsl.TransInfo) {
	// done
	klog.Infof("***EndAction:RollBack begin***")
	klog.Infof("***EndAction:RollBack end***")
}
