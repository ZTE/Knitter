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
	"sync"
)

type PodTableRole struct {
	isDetachingPodMap map[string]bool
	rwLock            sync.RWMutex
}

var podTableSingleton *PodTableRole
var podTableSingletonLock sync.Mutex

func GetPodTableSingleton() *PodTableRole {
	if podTableSingleton != nil {
		return podTableSingleton
	}

	podTableSingletonLock.Lock()
	defer podTableSingletonLock.Unlock()
	if podTableSingleton == nil {
		podTableSingleton = &PodTableRole{
			make(map[string]bool),
			sync.RWMutex{}}
	}
	return podTableSingleton
}

func (this *PodTableRole) IsExisted(podNs, podName, containerID string) bool {
	this.rwLock.RLock()
	defer this.rwLock.RUnlock()
	key := podNs + "." + podName + "." + containerID
	_, ok := this.isDetachingPodMap[key]
	return ok
}

func (this *PodTableRole) TryInsert(podNs, podName, containerID string) bool {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	key := podNs + "." + podName + "." + containerID
	if _, ok := this.isDetachingPodMap[key]; ok {
		return false
	}
	this.isDetachingPodMap[key] = true

	return true
}

func (this *PodTableRole) Delete(podNs, podName, containerID string) {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	key := podNs + "." + podName + "." + containerID
	delete(this.isDetachingPodMap, key)
}
