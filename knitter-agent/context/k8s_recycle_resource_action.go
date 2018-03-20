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
)

type K8SRecycleResourceAction struct {
}

func (this *K8SRecycleResourceAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Info("***K8SRecycleResourceAction:Exec begin***")
	klog.Info("***K8SRecycleResourceAction:Exec end***")
	return nil
}

func (this *K8SRecycleResourceAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***K8SRecycleResourceAction:RollBack begin***")
	klog.Infof("***K8SRecycleResourceAction:RollBack end***")
}
