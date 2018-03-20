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
	"github.com/ZTE/Knitter/knitter-agent/domain/adapter"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type IsK8SRecycleResource struct {
}

func (this *IsK8SRecycleResource) Ok(transInfo *transdsl.TransInfo) bool {
	//IsPodRestarted=true: sent delpod request, nw need recycle resource
	//request include netns, the netns is real exist;
	//k8s del pod(podns+podname), the deletionTimestamp is not nil
	var err error
	var delLabel string
	if transInfo.AppInfo.(*KnitterInfo).podJSON != nil {
		delLabel, err = adapter.JasonObjectGetString(transInfo.AppInfo.(*KnitterInfo).podJSON,
			"metadata", "deletionTimestamp")
		if err == nil {
			klog.Infof("***IsK8SRecycleResource: false***:[delLable:%v]", delLabel)
			return false
		}
		klog.Infof("***IsK8SRecycleResource: true***:[delLable:%v]", delLabel)
		return true
	}
	klog.Infof("***IsK8SRecycleResource: false***:pod json in k8s not exist")
	return false

}
