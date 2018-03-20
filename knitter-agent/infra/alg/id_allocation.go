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

package alg

import (
	"github.com/ZTE/Knitter/pkg/klog"
	"sync"
)

const offset int = 100

var vlanIDSlice []bool = make([]bool, offset)
var progress int = 0

type IDAllocator struct {
	idSlice  []bool
	progress int
	lock     sync.Mutex
}

func NewIDAllocator() *IDAllocator {
	return &IDAllocator{idSlice: make([]bool, offset)}
}

func (this *IDAllocator) InitByHistory(historySlice []int) {
	if len(historySlice) == 0 {
		return
	}

	buff := make([]int, len(historySlice))
	index := 0
	max := 0
	for _, id := range historySlice {
		buff[index] = id
		if id > max {
			max = id
		}
		index++
	}

	this.progress = max / offset
	for index := offset; index < (this.progress+1)*offset; index++ {
		this.idSlice = append(this.idSlice, false)
	}

	for _, id := range buff {
		this.idSlice[id] = true
	}
}

func (this *IDAllocator) Alloc() int {
	this.lock.Lock()
	defer this.lock.Unlock()
	for index, flag := range this.idSlice {
		if index == 0 {
			continue
		}
		if flag == false {
			this.idSlice[index] = true
			klog.Infof("alloc id:%v", index)
			return index
		}
	}

	progress++
	for index := progress * offset; index < (progress+1)*offset; index++ {
		vlanIDSlice = append(vlanIDSlice, false)
	}
	vlanIDSlice[progress*offset] = true
	return progress * offset
}

func (this *IDAllocator) Free(id int) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if id >= len(this.idSlice) {
		return
	}
	this.idSlice[id] = false
}
