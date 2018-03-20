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
	"github.com/vishvananda/netlink"
	//"github.com/ZTE/Knitter/pkg/klog"
	"bufio"
	"bytes"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/infra/os-encap"
	"github.com/ZTE/Knitter/pkg/klog"
	"strings"
)

type LinkRepository struct {
	Links []*Link
}

func NewLinkRepository() *LinkRepository {
	return &LinkRepository{Links: make([]*Link, 0)}
}

func (self *LinkRepository) Update(links []netlink.Link) {
	self.Links = make([]*Link, 0)
	self.AddRange(links)
}

func (self *LinkRepository) AddRange(links []netlink.Link) {
	for _, link := range links {
		l := NewLink(link)
		self.Links = append(self.Links, l)
	}
}

func (self *LinkRepository) FindLinkByMac(mac string) *Link {
	for _, link := range self.Links {
		if link.Type() == "device" &&
			link.Attrs().HardwareAddr.String() == mac {
			return link
		}
	}
	return nil
}

func (self *LinkRepository) FindLinkByName(name string) *Link {
	for _, link := range self.Links {
		if link.Type() == "veth" &&
			link.Attrs().Name == name {
			return link
		}
	}
	return nil
}

func (self *LinkRepository) FindVfLinkByName(name string) *Link {
	for _, link := range self.Links {
		if link.Type() == "device" &&
			link.Attrs().Name == name {
			return link
		}
	}
	return nil
}

const SysNetPath = "/sys/class/net"

//To find link by pci
func (self *LinkRepository) FindLinkByPci(pci string) *Link {
	defer func() {
		if err := recover(); err != nil {
			klog.Infof("FindEthNameByPci panic recover!")
		}
	}()
	var devPath string = SysNetPath

	lsOutput, err := osencap.Exec("ls", "-l", devPath)
	if err != nil {
		klog.Errorf("Unable to exec ls , err: %v, output: %s", err, string(lsOutput))
		return nil
	}
	lsScanner := bufio.NewScanner(bytes.NewBuffer([]byte(lsOutput)))
	for lsScanner.Scan() {
		line := strings.TrimSpace(lsScanner.Text())
		if len(line) == 0 {
			continue
		}
		if strings.Contains(line, "devices/pci") {
			tmpPci, ethName, err := self.GetPciEthMap(line)
			if err != nil {
				continue
			}
			if pci == tmpPci {
				tmpLink, err := netlink.LinkByName(ethName)
				if err != nil {
					klog.Errorf("Cannot find link by name [%v]. error: %v", ethName, err)
					return nil
				}
				link := NewLink(tmpLink)
				return link
			}
		}
	}
	klog.Infof("can not find link by pci:[%v]", pci)
	return nil
}

func (self *LinkRepository) GetPciEthMap(line string) (string, string, error) {
	parts := strings.Fields(line)
	for _, part := range parts {
		if strings.Contains(part, "devices/pci") {
			tmpSplit := strings.Split(part, "/")
			if len(tmpSplit) < 3 {
				continue
			}
			EthName := tmpSplit[len(tmpSplit)-1]
			BusInfo := tmpSplit[len(tmpSplit)-3]
			tmpSplit = strings.SplitAfter(BusInfo, "0000:")
			if len(tmpSplit) < 2 {
				continue
			}
			pci := tmpSplit[1]
			return pci, EthName, nil
		}
	}
	return "", "", fmt.Errorf("can not get pci-eth map")
}
