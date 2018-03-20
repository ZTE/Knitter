package services

import (
	"encoding/json"
	"errors"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/antonholmquist/jason"
	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"

	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/daos"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/knitter-monitor/infra"
	"github.com/ZTE/Knitter/pkg/klog"
)

const ipReg = "^(1\\d{2}|2[0-4]\\d|25[0-5]|[1-9]\\d|[1-9])\\." +
	"(1\\d{2}|2[0-4]\\d|25[0-5]|[1-9]\\d|\\d)\\." +
	"(1\\d{2}|2[0-4]\\d|25[0-5]|[1-9]\\d|\\d)\\." +
	"(1\\d{2}|2[0-4]\\d|25[0-5]|[1-9]\\d|\\d)$"

func GetPortService() PortServiceInterface {
	return &portService{}

}

type PortServiceInterface interface {
	//todo
	NewPortsWithEagerAttrFromK8s(pod *PodForCreatPort, nwJSON *jason.Object) ([]*Port, error)
	NewPortsWithEagerAttrAndLazyAttr(pod *PodForCreatPort, nwJSON *jason.Object) ([]*Port, error)
	DeleteBulkPorts(tenantID string, portIDs []string) error
}

type portService struct {
}

func newPortFromPortForDB(db *daos.PortForDB) *Port {
	portEagerAttr := PortEagerAttr{
		NetworkName:  db.NetworkName,
		NetworkPlane: db.NetworkPlane,
		PortName:     db.PortName,
		VnicType:     db.VnicType,
		Accelerate:   db.Accelerate,
		PodName:      db.PodName,
		PodNs:        db.PodNs,
		FixIP:        db.FixIP,
		IPGroupName:  db.IPGroupName,
		Metadata:     db.Metadata,
		Combinable:   db.Combinable,
		Roles:        db.Roles,
	}
	portLazyAttr := PortLazyAttr{
		ID:         db.ID,
		Name:       db.LazyName,
		TenantID:   db.TenantID,
		MacAddress: db.MacAddress,
		FixedIps:   db.FixedIps,
		GatewayIP:  db.GatewayIP,
		Cidr:       db.Cidr,
	}
	return &Port{
		EagerAttr: portEagerAttr,
		LazyAttr:  portLazyAttr,
	}
}

func (ps *portService) NewPortsWithEagerAttrAndLazyAttr(pod *PodForCreatPort, nwJSON *jason.Object) ([]*Port, error) {
	klog.Infof("NewPortsWithEagerAttrAndLazyAttr start, podname is [%v]", pod.PodName)
	klog.Debugf("NewPortsWithEagerAttrAndLazyAttr: pod is [%+v]", pod)
	klog.Debugf("NewPortsWithEagerAttrAndLazyAttr: nwJSON is [%v]", nwJSON)

	mports, err := ps.NewPortsWithEagerAttrFromK8s(pod, nwJSON)
	if err != nil {
		klog.Errorf("ps.NewPortsWithEagerAttrFromK8s(pod :[%v] , nwJSON:[%v]) err, error is [%v]", pod, nwJSON, err)
		return nil, err
	}
	createBulkPortsResp, err := ps.CreateBulkPorts(pod, mports)
	if err != nil {
		klog.Errorf("ps.NewPortsWithEagerAttrFromK8s() err, error is [%v]", err)
		return nil, err
	}
	//todo refactor
	err = fillPortLazyAttr(createBulkPortsResp, mports, pod.TenantID)
	if err != nil {
		klog.Errorf("fillPortLazyAttr(createBulkPortsResp, mports, pod.TenantID) err, error is [%v]", err)
		return nil, err
	}
	return mports, nil

}

func (ps *portService) NewPortsWithEagerAttrFromK8s(pod *PodForCreatPort, nwJSON *jason.Object) ([]*Port,
	error) {

	portArray, err := nwJSON.GetObjectArray("ports")
	if err != nil || len(portArray) == 0 {
		if err == nil {
			err = errobj.ErrGetPortConfigError
		}
		klog.Errorf("Get port config error! %v", err)
		return nil, err
	}
	var mports []*Port
	for _, portObj := range portArray {
		port := &Port{}
		err := port.fillPortEagerAttr(pod.PodNs, pod.PodName, portObj)
		if err != nil {
			return nil, err
		}
		mports = append(mports, port)
	}
	combinedPortObjs, err := combinePortObjs(mports)
	if err != nil {
		return nil, err
	}
	for i, combinedPortObj := range combinedPortObjs {
		klog.Infof("Post combination combinedPortObjs[%v]'s roles: %v", i, combinedPortObj.EagerAttr.Roles)
	}
	return combinedPortObjs, nil
}

func combinePortObjs(ports []*Port) ([]*Port, error) {
	for i := 0; i < len(ports); i++ {
		ports[i].EagerAttr.Roles = append(ports[i].EagerAttr.Roles, ports[i].EagerAttr.NetworkPlane)
		for j := i + 1; j < len(ports); j++ {
			if IsSameRoles(ports[i], ports[j]) {
				err := IsSameRolesIllegal(ports[i], ports[j])
				if err != nil {
					return nil, err
				}
				if ports[j].EagerAttr.PortName == constvalue.DefaultPortName {
					ports[i].EagerAttr.PortName = constvalue.DefaultPortName
				}
				ports[i].EagerAttr.Roles = append(ports[i].EagerAttr.Roles, ports[j].EagerAttr.NetworkPlane)
				ports = delEleOfSliceByIndex(ports, j)
				//back one step because of delete j'st element in slice
				j--
			} else if isSameC1PortWithDiffCombineAttr(ports[i], ports[j]) {
				err := errors.New("netplanes: " +
					ports[i].EagerAttr.NetworkPlane + " and " + ports[j].EagerAttr.NetworkPlane +
					" are all using the same C0's port with the same NetworkName: " +
					ports[i].EagerAttr.NetworkName + ", they must be combined")
				return nil, err
			}
		}
	}
	return ports, nil
}

func isSameC1PortWithDiffCombineAttr(portObj1, portObj2 *Port) bool {
	isSameC1RolesWithDiffCombineAttr := portObj1.EagerAttr.Combinable != portObj2.EagerAttr.Combinable &&
		portObj1.EagerAttr.NetworkName == portObj2.EagerAttr.NetworkName &&
		IsCTNetPlane(portObj1.EagerAttr.NetworkPlane) &&
		IsCTNetPlane(portObj2.EagerAttr.NetworkPlane) &&
		portObj1.EagerAttr.Accelerate == "true" &&
		portObj2.EagerAttr.Accelerate == "true"
	if isSameC1RolesWithDiffCombineAttr {
		return true
	}
	return false
}

func IsSameRoles(portObj1, portObj2 *Port) bool {
	return portObj1.EagerAttr.Combinable == "true" &&
		portObj2.EagerAttr.Combinable == "true" &&
		portObj1.EagerAttr.NetworkName == portObj2.EagerAttr.NetworkName &&
		portObj1.EagerAttr.VnicType == portObj2.EagerAttr.VnicType &&
		portObj1.EagerAttr.Accelerate == portObj2.EagerAttr.Accelerate &&
		portObj1.EagerAttr.PodName == portObj2.EagerAttr.PodName &&
		portObj1.EagerAttr.PodNs == portObj2.EagerAttr.PodNs &&
		portObj1.EagerAttr.IPGroupName == portObj2.EagerAttr.IPGroupName
	//portObj1.EagerAttr.Metadata == portObj2.EagerAttr.Metadata
}

func IsSameRolesIllegal(portObj1, portObj2 *Port) error {
	const eioNetPlane string = "eio"
	var err error
	var isEioAndC0NetPlaneConflict, isNetPlanReused bool
	isEioAndC0NetPlaneConflict = (IsCTNetPlane(portObj1.EagerAttr.NetworkPlane) && portObj2.EagerAttr.NetworkPlane == eioNetPlane ||
		IsCTNetPlane(portObj2.EagerAttr.NetworkPlane) && portObj1.EagerAttr.NetworkPlane == eioNetPlane) &&
		portObj1.EagerAttr.Accelerate == "true" && portObj2.EagerAttr.Accelerate == "true"
	if isEioAndC0NetPlaneConflict {
		err = errors.New("try to combine netplanes: " + portObj1.EagerAttr.NetworkPlane + " and " +
			portObj2.EagerAttr.NetworkPlane + " with same direct and accelerate attr error")
		klog.Errorf("IsSameRolesIllegal: %v", err)
		return err
	}

	isNetPlanReused = portObj1.EagerAttr.NetworkPlane == portObj2.EagerAttr.NetworkPlane
	if isNetPlanReused {
		err = errors.New("netplanes: " + portObj1.EagerAttr.NetworkPlane + " is reused in the combined port")
		klog.Errorf("IsSameRolesIllegal: %v", err)
		return err
	}

	return nil
}

func delEleOfSliceByIndex(slice []*Port, index int) []*Port {
	return append(slice[:index], slice[index+1:]...)
}

func IsCTNetPlane(netPlane string) bool {
	return netPlane == constvalue.NetPlaneControl ||
		netPlane == constvalue.NetPlaneMedia ||
		netPlane == constvalue.NetPlaneOam
}

var destoryBulkPorts = func(mc *infra.ManagerClient, tenantIDs, portIDs []string) {
	managerClient := infra.GetManagerClient()
	for i, tenantID := range tenantIDs {
		err := managerClient.DeleteNeutronPort(tenantID, portIDs[i])
		if err != nil {
			klog.Errorf("CreateNeutronBulkPortsAction.RollBack:knitterMgrObj.NeutronPortRole.Destroy error: %v", err)
		}
	}
}

func (ps *portService) DeleteBulkPorts(tenantID string, portIDs []string) error {
	managerClient := infra.GetManagerClient()
	for _, portID := range portIDs {
		err := managerClient.DeleteNeutronPort(tenantID, portID)
		if err != nil && !errobj.IsKeyNotFoundError(err) {
			klog.Errorf("DeleteBulkPorts:managerClient.DeleteNeutronPort(tenantID :[%v], portID:[%v]) error: %v", tenantID, portIDs, err)
			return err
		}
	}
	return nil
}

func (ps *portService) CreateBulkPorts(pod *PodForCreatPort, ports []*Port) (resp *infra.CreatePortsResp, err error) {
	resp = &infra.CreatePortsResp{}
	var crtBukPorts = false

	DestroyBulkPorts := func() {
		tenantIds := make([]string, 0)
		portIds := make([]string, 0)
		for _, port := range resp.Ports {
			tenantIds = append(tenantIds, pod.TenantID)
			portIds = append(portIds, port.PortID)
		}
		clientManager := infra.GetManagerClient()
		destoryBulkPorts(clientManager, tenantIds, portIds)
	}
	defer func() {
		if p := recover(); p != nil {
			debug.PrintStack()
		}
		if crtBukPorts && err != nil {
			DestroyBulkPorts()

		}
	}()

	managerClient := infra.GetManagerClient()
	reqs := ps.buildBulkPortsReq(pod, ports)
	if len(reqs.Ports) == 0 {
		klog.Infof("No need to creat ports!")
		klog.Infof("***CreateNeutronBulkPortsAction:Exec end***")
		return nil, err
	}
	portsByte, err := managerClient.CreateNeutronBulkPorts(pod.PodID, reqs, pod.TenantID)
	if err != nil {

		klog.Errorf("CreateNeutronBulkPorts: agtCtx.Mc.CreateNeutronBulkPorts failed, error! -%v", err)
		return nil, err
	}
	err = json.Unmarshal(portsByte, &resp)
	if err != nil {
		klog.Errorf("CreateNeutronBulkPorts: json.Unmarshal failed, err: %v", err)
		return nil, err
	}

	klog.Infof("CreateNeutronBulkPorts: create and Unmarshal result: [%v]", resp)
	return resp, err

}

func (ps *portService) buildBulkPortsReq(pod *PodForCreatPort, ports []*Port) *infra.ManagerCreateBulkPortsReq {
	reqs := infra.ManagerCreateBulkPortsReq{Ports: make([]infra.ManagerCreatePortReq, 0)}
	for _, port := range ports {
		//if port.EagerAttr.Accelerate == "true" &&
		//	IsCTNetPlane(port.EagerAttr.NetworkPane) {
		//	continue
		//}

		managerPortReq := port.TransformToMangerCreatePortReq(pod)
		reqs.Ports = append(reqs.Ports, *managerPortReq)
	}
	return &reqs
}

type Port struct {
	EagerAttr PortEagerAttr
	LazyAttr  PortLazyAttr
}

func (p *Port) transferToPortForDB() *daos.PortForDB {
	return &daos.PortForDB{
		NetworkName:  p.EagerAttr.NetworkName,
		NetworkPlane: p.EagerAttr.NetworkPlane,
		PortName:     p.EagerAttr.PortName,
		VnicType:     p.EagerAttr.VnicType,
		Accelerate:   p.EagerAttr.Accelerate,
		PodName:      p.EagerAttr.PodName,
		PodNs:        p.EagerAttr.PodNs,
		FixIP:        p.EagerAttr.FixIP,
		IPGroupName:  p.EagerAttr.IPGroupName,
		Metadata:     p.EagerAttr.Metadata,
		Combinable:   p.EagerAttr.Combinable,
		Roles:        p.EagerAttr.Roles,

		ID:         p.LazyAttr.ID,
		LazyName:   p.LazyAttr.Name,
		TenantID:   p.LazyAttr.TenantID,
		MacAddress: p.LazyAttr.MacAddress,
		FixedIps:   p.LazyAttr.FixedIps,
		GatewayIP:  p.LazyAttr.GatewayIP,
		Cidr:       p.LazyAttr.Cidr,
	}
}

//todo refactor
func fillPortLazyAttr(resp *infra.CreatePortsResp, ports []*Port, tenantID string) error {
	for i, port := range ports {
		var createPortInfo infra.CreatePortInfo
		var flag = false
		for _, createPortInfo = range resp.Ports {
			//if port.EagerAttr.PortName == createPortInfo.Name {
			// todo: createPortInfo real iaas name should be generated in knitter-agent and stored in PortObj.LazyAttr
			if strings.HasSuffix(createPortInfo.Name, port.EagerAttr.PortName) {
				flag = true
				break
			}
		}
		if flag {
			ports[i].LazyAttr.ID = createPortInfo.PortID
			ports[i].LazyAttr.Name = createPortInfo.Name
			ports[i].LazyAttr.MacAddress = createPortInfo.MacAddress
			ports[i].LazyAttr.FixedIps = createPortInfo.FixedIps
			ports[i].LazyAttr.TenantID = tenantID
			ports[i].LazyAttr.Cidr = createPortInfo.Cidr
			ports[i].LazyAttr.GatewayIP = createPortInfo.GatewayIP
		} else if port.EagerAttr.Accelerate == "false" || port.EagerAttr.NetworkPlane == "eio" {
			klog.Errorf("createPortInfo:%v is not in bulkports, we must destory bulkports", port.EagerAttr.PortName)
			return errobj.ErrPortNtFound
		}

	}
	return nil
}

func (p *Port) fillPortEagerAttr(podNs, podName string, portJSON *jason.Object) error {
	eagerPort := &EagerPort{}
	err := eagerPort.Transform(portJSON)
	if err != nil {
		return err
	}
	//metadata
	metadataObj, err := portJSON.GetObject("attributes", "metadata")
	if err == nil && metadataObj != nil {
		p.EagerAttr.Metadata = metadataObj.Interface()
	} else {
		p.EagerAttr.Metadata = make(map[string]string)
	}
	p.EagerAttr.NetworkName = eagerPort.NetworkName
	p.EagerAttr.NetworkPlane = eagerPort.NetworkPlane
	p.EagerAttr.PortName = eagerPort.PortName
	p.EagerAttr.VnicType = eagerPort.VnicType
	p.EagerAttr.Accelerate = eagerPort.Accelerate
	p.EagerAttr.FixIP = eagerPort.FixIP
	p.EagerAttr.IPGroupName = eagerPort.IPGroupName
	p.EagerAttr.PodNs = podNs
	p.EagerAttr.PodName = podName
	p.EagerAttr.Combinable = eagerPort.Combinable
	return nil
}

func (p *Port) TransformToMangerCreatePortReq(pod *PodForCreatPort) *infra.ManagerCreatePortReq {
	return &infra.ManagerCreatePortReq{
		TenantID:    pod.TenantID,
		NetworkName: p.EagerAttr.NetworkName,
		PortName:    p.EagerAttr.PortName,
		//logic p has no mode direct and physical
		VnicType: constvalue.LogicalPortDefaultVnicType,
		//todo  how to get nodeID
		NodeID:      infra.GetClusterID(),
		PodNs:       p.EagerAttr.PodNs,
		PodName:     p.EagerAttr.PodName,
		FixIP:       p.EagerAttr.FixIP,
		IPGroupName: p.EagerAttr.IPGroupName,
	}
}

type PortEagerAttr struct {
	NetworkName  string //"attach_to_network"
	NetworkPlane string
	PortName     string //nic_name
	VnicType     string
	Accelerate   string
	PodName      string
	PodNs        string
	FixIP        string
	IPGroupName  string
	Combinable   string
	Metadata     interface{}
	Roles        []string
}

type PortLazyAttr struct {
	//NetworkID      string
	ID         string
	Name       string
	TenantID   string
	MacAddress string
	FixedIps   []ports.IP
	GatewayIP  string
	Cidr       string
}

type EagerPort struct {
	NetworkName  string
	NetworkPlane string
	PortName     string
	VnicType     string
	Accelerate   string
	FixIP        string
	IPGroupName  string
	PortFunc     string
	Combinable   string
}

func (ep *EagerPort) Transform(portJSON *jason.Object) error {
	networkName, err := portJSON.GetString("attach_to_network")
	if err != nil {
		klog.Error("Get port attach Network ERROR:", err)
		return err
	} else if networkName == "" {
		err = errors.New("port attach network is blank")
		klog.Error("port attach network is blank")
		return err
	}
	portName, err := portJSON.GetString("attributes", "nic_name")
	if err != nil || portName == "" {
		portName = "eth_" + networkName
		if len(portName) > 12 {
			rs := []rune(portName)
			portName = string(rs[0:12])
		}
	}
	if len(portName) > 12 {
		klog.Errorf("Lenth of port name is greater than 12")
		return errors.New("lenth of port name is illegal")
	}
	portFunc, err := portJSON.GetString("attributes", "function")
	if err != nil || portFunc == "" {
		portFunc = "std"
	}

	portType, err := portJSON.GetString("attributes", "nic_type")
	if err != nil || portType == "" {
		portType = "normal"
	}

	isUseDpdk, _ := portJSON.GetString("attributes", "accelerate")
	if isUseDpdk != "true" {
		isUseDpdk = "false"
	}
	ipAddress, err := portJSON.GetString("attributes", "ip_addr")
	if err != nil {
		klog.Infof("Get port ip_address failure")
	}
	if ipAddress == "" {
		klog.Infof("Get port ip_address is blank string")
	} else {
		isLegitimate, errIPAddress := isIPLegitimate(ipAddress)
		if errIPAddress != nil {
			klog.Errorf("check ip_address failure")
			return errIPAddress
		}
		if isLegitimate {
			klog.Infof("ip_address is legitimate")
		} else {
			klog.Errorf("ip_address is illegal")
			return errors.New("ip_address is illegal")
		}
	}

	ipGroupName, err := portJSON.GetString("attributes", "ip_group_name")
	if err != nil {
		klog.Infof("Get port ip_group_name failure")
	}

	combinable, _ := portJSON.GetString("attributes", "combinable")
	if combinable != "true" {
		combinable = "false"
	}

	ep.NetworkName = networkName
	ep.NetworkPlane = portFunc
	ep.PortName = portName
	ep.VnicType = portType
	ep.Accelerate = isUseDpdk
	ep.FixIP = ipAddress
	ep.IPGroupName = ipGroupName
	ep.Combinable = combinable
	return nil
}

func isIPLegitimate(ipAddress string) (bool, error) {
	return regexp.MatchString(ipReg, ipAddress)
}
