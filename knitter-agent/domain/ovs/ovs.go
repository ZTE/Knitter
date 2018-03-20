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

package ovs

import (
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra/os-encap"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/go-ini/ini"
	"github.com/vishvananda/netlink"
	"math/rand"
	"os/exec"
	"strings"
	"time"
)

const (
	ASCII0      = 48
	ASCIIA      = 65
	ASCIIa      = 97
	NumRange    = 10
	LetterRange = 26
)

type VethPair struct {
	VethNameOfPod    string `json:"vetha"`
	VethNameOfBridge string `json:"vethb"`
	VethMacOfPod     string `json:"vetha_mac"`
	VethMacOfBridge  string `json:"vethb_mac"`
}

//To get phyNet-ovs pair
func GetPhyNwOvsPair(phyNwStr string) (*map[string]string, error) {
	var phyNwOvsPair = make(map[string]string)
	parts := strings.Split(phyNwStr, ",")
	if len(parts) < 1 {
		klog.Infof("Unexpected physnet-ovs config format. string:[%v]", phyNwStr)
		return nil, errors.New("unexpected physnet-ovs config format")
	}
	for _, pair := range parts {
		part := strings.Split(pair, ":")
		if len(part) < 2 {
			klog.Infof("Unexpected physnet-ovs config format.string:[%v]", phyNwStr)
			return nil, errors.New("unexpected physnet-ovs config format")
		}
		phyNwOvsPair[part[0]] = part[1]
	}
	return &phyNwOvsPair, nil
}

//To get phyNet-ovs map string
func GetPhyNwOvsMapStr(confPath string) (string, error) {
	cfg, err := InitConfig(confPath)
	if err != nil {
		klog.Errorf("Init config from [%v] error! %v", confPath, err)
		return "", err
	}
	section, err := cfg.GetSection("ovs")
	if err != nil {
		klog.Errorf("cfg.GetSection[ovs] error! %v", err)
		return "", err
	}
	key, err := section.GetKey("bridge_mappings")
	if err != nil {
		klog.Errorf("Section.GetKey[bridge_mappings] error! %v", err)
		return "", err
	}
	phyNwStr := key.String()
	return phyNwStr, nil
}

//To init a config by config file
func InitConfig(confPath string) (*ini.File, error) {
	cfg, err := ini.Load(confPath)
	if err != nil {
		klog.Errorf("Init config from [%v] error! %v", confPath, err)
		return nil, err
	}
	//speed up read operations(if only read)
	cfg.BlockMode = false
	return cfg, nil
}

func GetSectionName(PortType string) string {
	switch PortType {
	case "direct":
		return "sriov_nic"
	case "physical":
		return "physical"
	}
	return ""
}

func GetDriverbyPortType(cfg *ini.File, phyNw, PortType string) (string, error) {
	if PortType == "normal" {
		return GetNormalDriver(cfg, phyNw)
	}
	return GetCommonDriver(cfg, phyNw, PortType)
}

func GetNormalDriver(cfg *ini.File, phyNw string) (string, error) {
	section, err := cfg.GetSection("ovs")
	if err != nil {
		klog.Errorf("GetNormalDriver:GetSection return err:%v", err)
		return GetSriovDriver(cfg, phyNw)
	}
	driver, err := GetDriverFromSection(section, phyNw)
	if err != nil {
		klog.Errorf("GetNormalDriver:GetDriverFromSection return err:%v", err)
		return GetSriovDriver(cfg, phyNw)
	}
	return driver, nil
}

func GetSriovDriver(cfg *ini.File, phyNw string) (string, error) {
	section, err := cfg.GetSection("sriov_nic")
	if err != nil {
		klog.Errorf("GetSriovDriver GetSection retrun err:%v", err)
		return "", fmt.Errorf("getSriovDriver GetSection retrun err:%v", err)
	}
	driver, err := GetDriverFromSection(section, phyNw)
	if err != nil {
		klog.Errorf("GetSriovDriver GetDriverFromSection retrun err:%v", err)
		return "", fmt.Errorf("getSriovDriver GetDriverFromSection retrun err:%v", err)
	}
	return driver, nil
}

func GetCommonDriver(cfg *ini.File, phyNw, PortType string) (string, error) {
	SecName := GetSectionName(PortType)
	section, err := cfg.GetSection(SecName)
	if err != nil {
		klog.Infof("cfg.GetSection:%v return err:%v", SecName, err)
		return "", err
	}
	return GetDriverFromSection(section, phyNw)
}

func GetDriverFromSection(section *ini.Section, phyNw string) (string, error) {
	keyName := GetMappingKey(section)
	key, err := section.GetKey(keyName)
	if err != nil {
		klog.Info("GetKey keyname:%v return err:%v", keyName, err)
		return "", fmt.Errorf("getKey keyname:%v return err:%v", keyName, err)
	}
	phyNwOvsPair, err := GetPhyNwOvsPair(key.String())
	if err != nil {
		klog.Infof("Cannot transform config to [physnet:ovs/vf/pf] from section-[%v] key-[%v].", section.Name(), key.Name())
		return "", err
	}
	_, ok := (*phyNwOvsPair)[phyNw]
	if !ok {
		klog.Infof("Cannot find physnet [%v] in section-[%v] key-[%v]", phyNw, section.Name(), key.Name())
		return "", fmt.Errorf("cannot find physnet [%v] in section-[%v] key-[%v]", phyNw, section.Name(), key.Name())
	}
	driver := section.Name()
	klog.Infof("GetDriverbyPortType success driver is %v", driver)
	return driver, nil
}

func GetMappingKey(section *ini.Section) string {
	switch section.Name() {
	case "ovs":
		return "bridge_mappings"
	case "sriov_nic":
		return "physical_device_mappings"
	case "physical":
		return "physical_pf_mappings"
	}
	return ""
}

//To search a specific driver name by physical network

//To get driver name by physical network
var GetNwMechDriver = func(phyNw, PortType, confPath string) (string, error) {
	cfg, err := InitConfig(confPath)
	if err != nil {
		klog.Errorf("Init config from [%v] error! %v", confPath, err)
		return "", err
	}
	driver, err := GetDriverbyPortType(cfg, phyNw, PortType)
	if err != nil {
		klog.Errorf("GetDriverbyPortType VnicType:%v, PhyNw:%v, return err:%v", PortType, phyNw, err)
		return "", err
	}
	if strings.Contains(driver, "sriov_nic") {
		driver = strings.TrimRight(driver, "_nic")
	}
	return driver, nil
}

//To get ovs bridge by phyNet
var GetOvsBrg = func(phyNw string, confPath string) (string, error) {
	phyNwStr, err := GetPhyNwOvsMapStr(confPath)
	if err != nil {
		klog.Errorf("GetOvsBrg: GetBrgMaps() error! %v", err)
		return "", err
	}
	phyNwOvsPair, err := GetPhyNwOvsPair(phyNwStr)
	if err != nil {
		klog.Errorf("GetOvsBrg: GetPhyNwOvsPair(phyNwStr) error %v, ", err)
		return "", err
	}
	if len(*phyNwOvsPair) == 1 && phyNw == "" {
		klog.Info("phyNw is nil and ovs has configure one bridge!")
		for _, ovsBrg := range *phyNwOvsPair {
			return ovsBrg, nil
		}
	}
	ovsBrg, err1 := (*phyNwOvsPair)[phyNw]
	if err1 == false {
		klog.Errorf("Never configured [%v] in config file!", phyNw)
		return "", errors.New("phyNet does not exist error")
	}
	return ovsBrg, nil
}

//To get veth mac by veth name
func GetVethMac(vethName string) (string, error) {
	link, err := GetLinkByName(vethName)
	if err != nil {
		klog.Errorf("netlink.LinkByName(vethName) error: %v", err)
		return "", err
	}
	return link.Attrs().HardwareAddr.String(), nil
}

func GetrandSuffix() string {
	length := 7
	kinds := [][2]int{{NumRange, ASCII0}, {LetterRange, ASCIIA}, {LetterRange, ASCIIa}}
	randStr := make([]string, length)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		ikind := r.Intn(3)
		scope, base := kinds[ikind][0], kinds[ikind][1]
		randStr = append(randStr, string(base+r.Intn(scope)))
	}
	randSuffix := strings.Join(randStr, "")
	return randSuffix
}

func CreateVeth(vethPair *VethPair) error {
	ipLi, _ := exec.LookPath("ip")
	ipLiArgs := []string{"link", "add", vethPair.VethNameOfPod, "type", "veth", "peer", "name", vethPair.VethNameOfBridge}
	ipLinkOutput, err := exec.Command(ipLi, ipLiArgs...).CombinedOutput()
	if err != nil {
		klog.Errorf("CreateVethPair: ip link add veth error: %v, output: %s",
			err, string(ipLinkOutput))
		if strings.Contains(err.Error(), "exit") &&
			strings.Contains(err.Error(), "status") &&
			strings.Contains(err.Error(), "2") {
			klog.Errorf("CreateVethPair: Veth pair already exists! Retrying creation!")
		}
		return err
	}
	return nil
}

const MaxTimesRetryCreateVeth int = 10
const MaxTimesRetryFindVeth int = 10

func GetVethPairMacs(vethPair *VethPair) error {
	for i := 0; i < MaxTimesRetryFindVeth; i++ {
		var err error
		vethPair.VethMacOfPod, err = GetVethMac(vethPair.VethNameOfPod)
		if err != nil {
			klog.Error("CreateVethPair: get Mac by ",
				vethPair.VethNameOfPod, " error:", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}
		vethPair.VethMacOfBridge, err = GetVethMac(vethPair.VethNameOfBridge)
		if err != nil {
			klog.Error("CreateVethPair: get Mac by ",
				vethPair.VethNameOfBridge, " error:", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}
		return nil
	}
	return errors.New("get veth pair macs failed")
}

//To create a veth pair for pod & ovs
func CreateVethPair(args ...string) (VethPair, error) {
	for i := 0; i < MaxTimesRetryCreateVeth; i++ {
		var vethAName, vethBName string
		if len(args) == 0 {
			randSuffix := GetrandSuffix()
			vethAName = "vethP" + randSuffix
			vethBName = "vethO" + randSuffix
		} else {
			vethAName = args[0]
			vethBName = args[1]
		}
		vethPair := VethPair{VethNameOfPod: vethAName,
			VethNameOfBridge: vethBName}
		err := CreateVeth(&vethPair)
		if err != nil {
			klog.Errorf("CreateVethPair: ip link add veth error: %v", err)
			continue
		}
		time.Sleep(50 * time.Millisecond)
		err = GetVethPairMacs(&vethPair)
		if err != nil {
			err = DeleteVethPair(vethPair)
			continue
		}
		return vethPair, nil
	}
	vethPair := VethPair{}
	return vethPair, errors.New("create-Veth-Pair-ERROR")
}

func GetLinkByName(name string) (netlink.Link, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	for _, link := range links {
		if link.Attrs().Name == name {
			return link, nil
		}
	}

	for _, link := range links {
		klog.Tracef("expect to Find[%s], now find[%s]", name, link.Attrs().Name)
	}
	return nil, errors.New("Link " + name + " not found")
}

//To delete a veth pair
var DeleteVethPair = func(vethPair VethPair) error {
	link, err := GetLinkByName(vethPair.VethNameOfPod)
	if err == nil {
		err := netlink.LinkDel(link)
		if err != nil {
			klog.Errorf("Delete veth [%v] error: %v", link, err)
			return err
		}
		return nil
	}
	klog.Errorf("Cannot find veth [%v] error: %v", vethPair.VethNameOfPod, err)
	link, err = GetLinkByName(vethPair.VethNameOfBridge)
	if err == nil {
		err := netlink.LinkDel(link)
		if err != nil {
			klog.Errorf("Delete veth [%v] error: %v", link, err)
			return err
		}
		return nil
	}
	klog.Errorf("Cannot find veth [%v] error: %v", vethPair.VethNameOfBridge, err)
	return errors.New("cannot find veth pair error")
}

var TimeSleepFunc = func(d time.Duration) {
	time.Sleep(d)
}

// check ovs
func addOvsBr(brName string) error {
	maxRetryTimes := constvalue.MaxRetryTimesOfOvsOp
	retryInterval := constvalue.MaxRetryIntervalOfOvsOp
	ovsctlArgs := []string{"--if-exists", "del-br", brName, "--", "add-br", brName}
	for i := 0; i < maxRetryTimes; i++ {
		_, err := osencap.Exec(constvalue.OvsVsctl, ovsctlArgs...)
		if err != nil {
			klog.Errorf("ovsctrl add-br %s [%d] times failed, total retry %d times!", brName, i, maxRetryTimes)
			TimeSleepFunc(time.Duration(retryInterval) * time.Second)
			continue
		}
		return nil
	}
	return errobj.ErrWaitOvsUsableFailed
}

func delOvsBr(brName string) error {
	maxRetryTimes := constvalue.MaxRetryTimesOfOvsOp
	retryInterval := constvalue.MaxRetryIntervalOfOvsOp
	ovsctlArgs := []string{"del-br", brName}
	for i := 0; i < maxRetryTimes; i++ {
		_, err := osencap.Exec(constvalue.OvsVsctl, ovsctlArgs...)
		if err != nil {
			klog.Errorf("ovsctrl del-br %s [%d] times failed, total retry %d times!", brName, i, maxRetryTimes)
			TimeSleepFunc(time.Duration(retryInterval) * time.Second)
			continue
		}
		return nil
	}
	return errobj.ErrWaitOvsUsableFailed
}

func checkOvsPortOp(brName string) error {
	maxRetryTimes := constvalue.MaxRetryTimesOfOvsOp
	retryInterval := constvalue.MaxRetryIntervalOfOvsOp
	var err error
	var outPut string
	port := "checkport" + GetrandSuffix()
	for idx := 1; idx <= maxRetryTimes; idx++ {
		klog.Infof("IsOvsUsable: checkOvsPort time: %d start", idx)
		outPut, err = checkOvsPort(brName, port)
		if err == nil {
			klog.Infof("IsOvsUsable: checkOvsPort time: %d SUCC", idx)
			return nil
		}
		if strings.Contains(outPut, "database connection failed") {
			klog.Errorf("IsOvsUsable: checkOvsPort time: %d FAIL, database connection failed", idx)
		} else {
			klog.Errorf("IsOvsUsable: checkOvsPort time: %d FAIL, error: %v", idx, err)
		}
		TimeSleepFunc(time.Duration(retryInterval) * time.Second)
		continue
	}
	klog.Errorf("IsOvsUsable: checkOvsPort all times FAIL, error: %v", err)
	return errobj.ErrWaitOvsUsableFailed
}

func WaitOvsUsable() error {
	var err error
	brName := constvalue.OvsBrcheck
	err = addOvsBr(brName)
	if err != nil {
		klog.Infof("WaitOvsUsable: addOvsBr: %s FAILED, error: %v", brName, err)
		return err
	}

	err = checkOvsPortOp(brName)
	if err != nil {
		klog.Infof("WaitOvsUsable: checkOvsPortOp FAILED, error: %v", err)
		return err
	}

	err = delOvsBr(brName)
	if err != nil {
		klog.Infof("WaitOvsUsable: delOvsBr: %s FAILED, error: %v", brName, err)
		return err
	}

	return nil
}

func checkOvsPort(br, port string) (string, error) {
	argsAdd := []string{"--if-exists", "del-port", port,
		"--", "add-port", br, port,
		"--", "set", "Interface", port, "options:df_default=false"}
	outPut, err := osencap.Exec("ovs-vsctl", argsAdd...)
	if err != nil {
		klog.Errorf("checkOvsPort: ovs-vsctl add-port: %s FAIL, error :%v, outPut: %s", port, err, outPut)
		return outPut, err
	}

	argsDel := []string{"--if-exists", "del-port", port}
	outPut, err = osencap.Exec("ovs-vsctl", argsDel...)
	if err != nil {
		klog.Errorf("checkOvsPort: ovs-vsctl del-port: %s FAIL, error :%v, outPut: %s", port, err, outPut)
		//return outPut, err
	}
	return outPut, err
}
