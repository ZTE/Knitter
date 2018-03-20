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
	"github.com/ZTE/Knitter/knitter-agent/infra/alg"
	"github.com/ZTE/Knitter/pkg/klog"
	"strconv"
)

var idAllocator *alg.IDAllocator

type VlanIDAllocatorRole struct {
}

func InitVlanIDSlice(buff []int) {
	idAllocator = alg.NewIDAllocator()
	idAllocator.InitByHistory(buff)
}

func (this *VlanIDAllocatorRole) Alloc() string {
	return strconv.Itoa(idAllocator.Alloc())
}

func (this *VlanIDAllocatorRole) Free(vlanID string) error {
	id, err := strconv.Atoi(vlanID)
	if err != nil {
		klog.Errorf("VlanIdAllocator:Free strconv vlanid: %v to int err: %v", vlanID, err)
		return err
	}
	idAllocator.Free(id)
	return nil
}
