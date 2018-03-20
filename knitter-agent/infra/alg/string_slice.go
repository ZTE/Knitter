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

import "github.com/ZTE/Knitter/pkg/klog"

type StringSlice []string

func NewStringSlice() StringSlice {
	return make(StringSlice, 0)
}

func (this *StringSlice) Add(elem string) error {
	for _, v := range *this {
		if v == elem {
			klog.Errorf("StringSlice:Insert elem already exist")
			return ErrElemExist
		}
	}
	*this = append(*this, elem)
	return nil
}

func (this *StringSlice) Remove(elem string) error {
	flag := true
	for i, v := range *this {
		if v == elem {
			if i+2 > len(*this) {
				*this = (*this)[:i]

			} else {
				*this = append((*this)[:i], (*this)[i+1:]...)
			}
			flag = false
			break
		}
	}
	if flag {
		klog.Errorf("StringSlice:Delete elem not exist")
		return ErrElemNtExist
	}
	return nil
}
