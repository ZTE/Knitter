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

package bind

import (
	"github.com/ZTE/Knitter/knitter-agent/infra/util/base"
	"github.com/ZTE/Knitter/pkg/klog"
	"sync"
)

type Pci struct {
	ID      string
	Path    string
	BusInfo string

	Writer base.Writer
}

var pciMutex sync.Mutex

func lockPci() {
	pciMutex.Lock()
}
func unlockPci() {
	pciMutex.Unlock()
}

func NewPci(ID, path, busInfo string) *Pci {
	return &Pci{
		ID:      ID,
		Path:    path,
		BusInfo: busInfo,
	}
}

func (self *Pci) Unbind() error {
	operationPath := self.UnbindPath() + "unbind"
	err := self.Writer.WriteFile(operationPath, []byte(self.BusInfo), 0200)
	if err != nil {
		klog.Errorf("bind-Pci-unbind:Failed to write bus-info %s to file %q: %v, unbind failed", self.BusInfo, operationPath, err)
		return err
	}
	return nil
}

func (self *Pci) UnbindPath() string {
	return self.Path +
		"/devices/" +
		self.BusInfo +
		"/driver/"
}

func (self *Pci) SetWriter(writer base.Writer) {
	self.Writer = writer
}
