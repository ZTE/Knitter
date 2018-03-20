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

package transdsl

type Fragment interface {
	Exec(transInfo *TransInfo) error
	RollBack(transInfo *TransInfo)
}

func forEachFragments(fragments []Fragment, transInfo *TransInfo) (int, error) {
	for i := 0; i < len(fragments); i++ {
		err := fragments[i].Exec(transInfo)
		if err != nil {
			if IsErrorEqual(err, ErrTransEnd) {
				return 0, err
			}
			return i, err
		}
	}
	return 0, nil
}

func backEachFragments(fragments []Fragment, transInfo *TransInfo, index int) {
	if index <= 0 {
		return
	}
	index--
	for ; index >= 0; index-- {
		fragments[index].RollBack(transInfo)
	}
}
