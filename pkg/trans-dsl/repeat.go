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

type Repeat struct {
	Fragments []Fragment
	FuncVar   func() Fragment
}

func (this *Repeat) Exec(transInfo *TransInfo) error {
	this.Fragments = make([]Fragment, transInfo.Times)
	for i := 0; i < transInfo.Times; i++ {
		transInfo.RepeatIdx = i
		this.Fragments[i] = this.FuncVar()
		err := this.Fragments[i].Exec(transInfo)
		if err != nil {
			if err.Error() == ErrContinue.Error() {
				continue
			}
			if i == 0 {
				return err
			}
			i--
			for j := i; j >= 0; j-- {
				transInfo.RepeatIdx = j
				this.Fragments[j].RollBack(transInfo)
			}
			return err
		}
	}
	return nil
}

func (this *Repeat) RollBack(transInfo *TransInfo) {
	for i := transInfo.Times - 1; i >= 0; i-- {
		transInfo.RepeatIdx = i
		this.Fragments[i].RollBack(transInfo)
	}
}
