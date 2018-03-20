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
	"bufio"
	"bytes"
	"errors"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/infra/os-encap"
	"github.com/ZTE/Knitter/pkg/klog"
	"io/ioutil"
	"os/exec"
	//	"runtime/debug"
	"encoding/json"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/adapter"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/port-role"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/antonholmquist/jason"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func BuildNormalNic(paasPort *manager.Port, netFunc string) (*Dpdknic, error) {
	_, ipnet, err := net.ParseCIDR(paasPort.CIDR)
	if err != nil {
		klog.Errorf("buildNomalNic:net.ParseCIDR(subnetinfo.CIDR) error! -%v", err)
		return nil, err
	}
	maskstr := fmt.Sprintf("%v", ipnet.Mask[0]) + "." + fmt.Sprintf("%v", ipnet.Mask[1]) + "." +
		fmt.Sprintf("%v", ipnet.Mask[2]) + "." + fmt.Sprintf("%v", ipnet.Mask[3])
	normalNic := &Dpdknic{
		BusInfo:    "",
		Name:       paasPort.Name,
		IP:         paasPort.FixedIPs[0].Address,
		Mask:       maskstr,
		Gateway:    paasPort.GatewayIP,
		Mac:        paasPort.MACAddress,
		Function:   netFunc,
		Accelerate: "false",
	}
	return normalNic, nil
}

func DestroyResidualBrintIntfcs() {
	intfcs, err := getAllBrintIntfcs("br-int")
	if err != nil {
		klog.Errorf("destroyResidualBrintIntfcs: get all br-int interfaces error: %v", err)
		return
	}

	for _, intfc := range intfcs {
		if portrole.GetPortTableSingleton().IsExist(intfc) {
			klog.Infof("destroyResidualBrintIntfcs: found interface: %s in portMgr", intfc)
		} else {
			klog.Errorf("destroyResidualBrintIntfcs: interface: %s not found in portMgr, destroy it", intfc)
			err = DelVethFromOvs("br-int", intfc)
			if err != nil {
				klog.Errorf("destroyResidualBrintIntfcs: DelVethFromOvs interface: %s error: %v", intfc, err)
			}
		}
	}
}

var getAllBrintIntfcs = func(bridge string) ([]string, error) {
	ovsctl, _ := exec.LookPath("ovs-vsctl")
	ovsctlArgs := []string{"list-ports", bridge}
	ovsctlOutput, err := exec.Command(ovsctl, ovsctlArgs...).CombinedOutput()
	if err != nil {
		klog.Errorf("ovs-vsctl list-ports error ! -%v output: %s", err,
			string(ovsctlOutput))
		return nil, err
	}

	ovsPorts := strings.Fields(string(ovsctlOutput))
	var ports []string
	for _, port := range ovsPorts {
		if strings.HasPrefix(port, "veth") {
			ports = append(ports, port)
		}
	}
	return ports, nil
}

var getAllBrNames = func() ([]string, error) {
	ovsctlArgs := []string{"list-br"}
	ovsctlOutput, err := osencap.Exec(constvalue.OvsVsctl, ovsctlArgs...)
	if err != nil {
		klog.Errorf("ovs-vsctl list-br error ! -%v output: %s", err,
			string(ovsctlOutput))
		return nil, err
	}
	brNames := strings.Fields(string(ovsctlOutput))
	klog.Infof("getAllBrNames: get all bridges result: %v", brNames)
	return brNames, nil
}

func DelOvsBrsNousedInterfacesLoop() {
	klog.Info("DelOvsBrsNousedInterfacesLoop: starting loop")
	for {
		time.Sleep(time.Duration(constvalue.CheckNousedIfsIntval) * time.Second)
		brNames, err := getAllBrNames()
		if err != nil {
			klog.Errorf("DelOvsBrsNousedInterfacesLoop: getAllBrNames FAILED, error: %v", err)
			continue
		}
		klog.Info("DelOvsBrsNousedInterfacesLoop: START cycle")
		for _, brName := range brNames {
			DelNousedOvsBrInterfaces(brName)
		}
		klog.Info("DelOvsBrsNousedInterfacesLoop: END cycle")
	}
}

func DelNousedOvsBrInterfacesLoop(brName string) {
	klog.Infof("DelNousedOvsBrInterfacesLoop: starting loop")
	for {
		time.Sleep(time.Duration(constvalue.CheckNousedIfsIntval) * time.Second)
		klog.Info("DelNousedOvsBrInterfacesLoop: START cycle")
		DelNousedOvsBrInterfaces(brName)
		klog.Info("DelNousedOvsBrInterfacesLoop: END cycle")
	}
}

func DelNousedOvsBrInterfaces(brName string) error {
	klog.Trace("DelNousedOvsBrInterfaces: START")
	ports, err := getAllBrintIntfcs(brName)
	if err != nil {
		klog.Errorf("DelNousedOvsBrInterfacesLoop: getAllBrintIntfcs(%s) FAILED, error: %v", brName, err)
		return err
	}
	klog.Tracef("DelNousedOvsBrInterfacesLoop: getAllBrintIntfcs(%s) get all interfaces: %v", brName, ports)

	ifs := ListNousedOvsBrInterfaces(ports)
	if len(ifs) == 0 {
		klog.Infof("DelNousedOvsBrInterfacesLoop: ListNonusedOvsBrInterfaces(%v) returns empty list", ports)
		return nil
	}
	klog.Infof("DelNousedOvsBrInterfacesLoop: ListNonusedOvsBrInterfaces() returns interfaces: %v", ifs)

	DelOvsBrInterfaces(brName, ifs)
	klog.Trace("DelNousedOvsBrInterfacesLoop: END cycle")
	return nil
}

var ListNousedOvsBrInterfaces = func(ports []string) []string {
	nonusedIfs := make([]string, 0)
	for _, portName := range ports {
		// by default, ovs bridge interface name is same as port name
		if isInterfaceDetached(portName) {
			nonusedIfs = append(nonusedIfs, portName)
		}
	}
	return nonusedIfs
}

func isInterfaceDetached(ifName string) bool {
	args := []string{"list", "interface", ifName}
	output, err := osencap.Exec(constvalue.OvsVsctl, args...)
	if err != nil {
		klog.Errorf("isInterfaceDetached: %s %s FAILED, error: %v, output: %s",
			constvalue.OvsVsctl, args, err, string(output))
		return false
	}

	isNewLineFunc := func(r rune) bool { return r == '\n' }
	fields := strings.FieldsFunc(output, isNewLineFunc)

	for _, field := range fields {
		if strings.HasPrefix(field, constvalue.OvsIfErrorField) &&
			strings.Contains(field, constvalue.OvsIfErrorNoDev) {
			klog.Infof("isInterfaceDetached: found %s detached", ifName)
			return true
		}
	}

	klog.Tracef("isInterfaceDetached: found %s not detached", ifName)
	return false
}

var DelOvsBrInterfaces = func(brName string, ifs []string) {
	klog.Infof("DelOvsBrInterfaces: delete %v from %s START", ifs, brName)
	for _, ifName := range ifs {
		err := DelVethFromOvs(brName, ifName)
		if err != nil {
			klog.Errorf("DelOvsBrInterfaces: DelVethFromOvs %s from %s FAILED, error: %v", brName, ifName, err)
		}
		klog.Infof("DelOvsBrInterfaces: DelVethFromOvs %s from %s SUCC", brName, ifName)
	}

	klog.Infof("DelOvsBrInterfaces: delete %v from %s FINI", ifs, brName)
}

var GetEthtoolOutputFunc func(ethName string) ([]byte, error)

func AddVethToOvsWithVlanID(ovsBridge, vethPort, vlanID string) error {
	_, err := osencap.Exec("ovs-vsctl", "add-port", ovsBridge, vethPort, "tag="+vlanID)
	return err
}

func ActiveVethPort(vethPort string) error {
	ovsctl, _ := exec.LookPath("ip")
	ovsctlArgs := []string{"link", "set", "dev", vethPort, "up"}
	output, err := exec.Command(ovsctl, ovsctlArgs...).CombinedOutput()
	if err != nil {
		param := " link set dev  " + vethPort + " up"
		klog.Error("ERROR:" + ovsctl + param +
			"\n OUTPUT:" + string(output) +
			"\n CAUSE:" + err.Error())
		return fmt.Errorf("%v:ActiveVethPort:exec.command() ERROR", err)
	}
	return nil
}

func DeactivateVethPort(vethPort string) error {
	ovsctl, err := exec.LookPath("ip")
	if err != nil {
		klog.Errorf("DeactivateVethPort: ip command not found in system")
	}
	ovsctlArgs := []string{"link", "set", "dev", vethPort, "down"}
	output, err := exec.Command(ovsctl, ovsctlArgs...).CombinedOutput()
	if err != nil {
		param := " link set dev  " + vethPort + " down"
		klog.Error("ERROR:" + ovsctl + param +
			"\n OUTPUT:" + string(output) +
			"\n CAUSE:" + err.Error())
		return fmt.Errorf("%v:DeactivateVethPort:exec.command() ERROR", err)
	}
	return nil
}

func DelVethFromOvs(ovsBridge, vethPort string) error {
	_, err := osencap.Exec("ovs-vsctl", "del-port", ovsBridge, vethPort)
	return err
}

func GetOVSList() (string, error) {
	OVSList, err := osencap.Exec("ovs-vsctl", "list-br")
	return OVSList, err
}

func getEthtoolOutput(ethName string) ([]byte, error) {
	ethtool, _ := exec.LookPath("ethtool")
	ethtoolArgs := []string{"-i", ethName}
	var ethtoolOutput []byte
	var err error
	for i := 0; i < 10; i++ {
		ethtoolOutput, err = exec.Command(ethtool, ethtoolArgs...).CombinedOutput()
		if err != nil {
			klog.Errorf("bind-getEthtoolOutput:retryTimes: [%v], Unable to query interface %v, err: %v, output: %v", i, ethName, err, string(ethtoolOutput))
			time.Sleep(time.Second)
			continue
		}
		return ethtoolOutput, nil
	}
	return nil, fmt.Errorf("%v:bind-getEthtoolOutput:exec.command() error", err)
}

func init() {
	GetEthtoolOutputFunc = getEthtoolOutput
}

func GetBusInfo(ethName string) (string, error) {
	ethtoolOutput, err := GetEthtoolOutputFunc(ethName)
	if err != nil {
		klog.Errorf("bind-GetBusInfo:getEthtoolOutputFunc failed, err: %v", err)
		return "", err
	}

	busInfo := ""
	ethtoolScanner := bufio.NewScanner(bytes.NewBuffer(ethtoolOutput))
	for ethtoolScanner.Scan() {
		line := strings.TrimSpace(ethtoolScanner.Text())
		if len(line) == 0 {
			continue
		}

		if strings.Contains(line, "bus-info:") {
			parts := strings.Split(line, " ")
			if len(parts) != 2 {
				klog.Errorf("bind-GetBusInfo:Unexpected bus-info output: %v", parts)
			}

			busInfo = parts[1]
			break
		}
	}

	klog.Info("bind-GetBusInfo:Query success, bus-info:%s", busInfo)

	return busInfo, nil
}

var DockerExec = func(containerId, program, para string) error {
	dockerCtl, _ := exec.LookPath("docker")
	dockerPsArgs := []string{"exec", containerId, program, string(para)}
	output, err := exec.Command(dockerCtl, dockerPsArgs...).CombinedOutput()
	if err != nil || strings.Contains(strings.ToUpper(string(output)), "ERROR") {
		klog.Infof("DockerExec-err:output :%v", string(output))
		param := " docker exec " + containerId + " " + program + " " + para
		klog.Errorf("ERROR:" + dockerCtl + param +
			"\n OUTPUT:" + string(output) +
			"\n CAUSE:" + err.Error())
		return fmt.Errorf("%v:dockerExec:exec.command() ERROR", err)
	}
	klog.Infof("DockerExec-suc:output :%v", string(output))
	return nil
}

var Nsenter = func(pid, program, para string) error {
	nsenterCtl, _ := exec.LookPath("nsenter")
	nsenterArgs := []string{"-t", pid, "--mount", "--uts", "--ipc", "--net", "--pid", program, para}
	output, err := exec.Command(nsenterCtl, nsenterArgs...).CombinedOutput()
	if err != nil {
		param := " nsenter -t  " + pid + " --mount --uts --ipc --net --pid " + program + " " + para
		klog.Errorf("ERROR:" + nsenterCtl + param +
			"\n OUTPUT:" + string(output) +
			"\n CAUSE:" + err.Error())
		return fmt.Errorf("%v:Nsenter:exec.command() ERROR", err)
	}
	return nil
}

func GetContainerDir() (string, error) {
	dockerInfo, err := GetDockerInfoByKey([]string{"Docker", "Root", "Dir:"})
	if err != nil {
		klog.Errorf("GetDockerInfoByKey error ! -%v", err)
		return "", fmt.Errorf("%v:GetDockerInfoByKey(Docker Root Dir:) ERROR", err)
	}
	klog.Infof("GetDockerInfoByKey:info is %v", dockerInfo[0])
	containersDir := dockerInfo[0] + "/containers"
	return containersDir, nil
}

func GetDockerInfoByKey(key []string) ([]string, error) {
	dockerCtl, _ := exec.LookPath("docker")
	dockerCtlArgs := []string{"info"}
	dockerCtlOutput, err := exec.Command(dockerCtl, dockerCtlArgs...).CombinedOutput()
	if err != nil {
		klog.Errorf("docker info error ! -%v output: %s", err,
			string(dockerCtlOutput))
		return nil, err
	}
	dockerInfo := strings.Fields(string(dockerCtlOutput))
	klog.Infof("no place dockerInfo:%v", dockerInfo)

	var infos []string
	var count = 0

	for _, item := range dockerInfo {
		if count == len(key) {
			klog.Infof("Docker Root Dir:%v", item)
			infos = append(infos, item)
			break
		}
		if strings.Compare(string(item), key[count]) == 0 {
			klog.Infof("item:%v || key[%v]:%v", string(item), count, key[count])
			count = count + 1
			continue
		}
		count = 0
	}
	if count == 0 {
		klog.Infof("Do not find the Docker Root Dir, use the docker default conf!")
		infos = append(infos, "/var/lib/docker")
	}
	return infos, nil
}

//c0
var GetPidFromDockerConf = func(containerId string) (string, error) {
	containerDir, err := GetContainerDir()
	if err != nil {
		klog.Errorf("GetPidFromDockerConf:GetContainerDir error! %v", err)
		return "", err
	}
	dockerConfPath := containerDir + "/" + containerId
	fileList, err := ioutil.ReadDir(dockerConfPath)
	if err != nil {
		klog.Errorf("GetPidFromDockerConf:ReadDir error! %v", err)
		return "", err
	}
	dockerConfFile := dockerConfPath + "/" + "config.json"
	for _, f := range fileList {
		rxConf := regexp.MustCompile(`^conf.*\.json$`)
		if rxConf.MatchString(f.Name()) {
			dockerConfFile = dockerConfPath + "/" + f.Name()
		}
	}
	dockerConf, err := ioutil.ReadFile(dockerConfFile)
	if err != nil {
		klog.Errorf("GetPidFromDockerConf:ReadFile error! %v", err)
		return "", err
	}
	dockerConfJSON, errJSON := jason.NewObjectFromBytes(dockerConf)
	if errJSON != nil {
		klog.Errorf("GetPidFromDockerConf:NewObjectFromBytes error! %v", errJSON)
		return "", errJSON
	}
	pidInt, errInt := dockerConfJSON.GetInt64("State", "Pid")
	klog.Infof("GetPidFromDockerConf:pid is %v", pidInt)
	if errInt != nil {
		klog.Errorf("GetPidFromDockerConf:GetInt64 error! %v", errJSON)
		return "", errInt
	}
	pidStr := strconv.FormatInt(pidInt, 10)
	return pidStr, nil
}

var GetC0ImageName = func() string {
	agtCtx := cni.GetGlobalContext()
	var c0ImageName = ""
	for i := 0; i < 3; i++ {
		key := dbaccessor.GetKeyOfImageNameForC0(agtCtx.ClusterID, agtCtx.HostIP)
		c0ImageName, _ = adapter.ReadLeafFromDb(key)
		if c0ImageName != "" {
			klog.Infof("C0 image is %v", c0ImageName)
			break
		}
		time.Sleep(10 * time.Second)
	}
	if c0ImageName == "" {
		klog.Errorf("GetC0ContainerIdByImageName: GetC0ContainerIdByImageName error : c0 image empty!")
	}
	return c0ImageName
}

func GetC0PodInfo() (*jason.Object, error) {
	agtCtx := cni.GetGlobalContext()
	pod := Pod{}
	key := dbaccessor.GetKeyOfPodForC0(agtCtx.ClusterID, agtCtx.HostIP)
	c0Pod, _ := adapter.ReadLeafFromDb(key)
	err := json.Unmarshal([]byte(c0Pod), &pod)
	if err != nil {
		klog.Errorf("json.Unmarshal port of pod error! -%v", err)
		return nil, err
	}

	_, podjson, err := agtCtx.K8s.GetPod(pod.K8sns, pod.Name)
	return podjson, err
}

var GetContainerIDByName = func() (s string, e error) {
	var c0ContainerID string
	c0ImageName := GetC0ImageName()
	if c0ImageName == "" {
		return "", errors.New("getContainerIdByName:Get c0 image error")
	}
	for i := 0; i <= 2; i++ {
		c0PodJSON, err := GetC0PodInfo()
		if err != nil {
			return "", err
		}
		containersList, errList := c0PodJSON.GetObjectArray("status", "containerStatuses")
		if errList != nil {
			return "", errList
		}
		for _, container := range containersList {
			imageName, _ := container.GetString("image")
			if imageName != c0ImageName {
				continue
			}
			_, errRunning := container.GetObject("state", "running")
			if errRunning != nil {
				klog.Warningf("GetContainerIdByName: C0 container is not running")
				break
			}
			c0ContainerID, _ = container.GetString("containerID")
			c0ContainerID = strings.TrimPrefix(c0ContainerID, "docker://")
			return c0ContainerID, nil
		}
		time.Sleep(5 * time.Second)
	}
	return "", errors.New("getContainerIdByName:Get c0 id error")
}
