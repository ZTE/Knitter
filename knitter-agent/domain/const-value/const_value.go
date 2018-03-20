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

package constvalue

const (
	OvsVsctl             string = "ovs-vsctl"
	OvsOfctl             string = "ovs-ofctl"
	OvsBrint             string = "br-int"
	OvsBrtun             string = "br-tun"
	OvsBrcheck           string = "br-nwnode-check"
	OvsBr0               string = "obr0"
	OvsIfErrorField      string = "error"
	OvsIfErrorNoDev      string = "No such device"
	CheckNousedIfsIntval int    = 5 * 60
	InvalidNetID         int    = 4096
	DefaultTunIntPort    uint   = 1
	DefaultMtu           string = "1500"

	MaxRetryTimesOfOvsOp    int = 40
	MaxRetryIntervalOfOvsOp int = 5

	MaxPciDeviceNum = 25

	HostTypePhysicalMachine = "bare_metal"
	HostTypeVirtualMachine  = "virtual_machine"
	MaxNwSendDelPodTimes    = 6
)

const (
	NetPlaneStd     = "std"
	NetPlaneEio     = "eio"
	NetPlaneControl = "control"
	NetPlaneMedia   = "media"
	NetPlaneOam     = "oam"
)

const (
	MechDriverOvs      = "normal"
	MechDriverSriov    = "direct"
	MechDriverPhysical = "physical"
)

const (
	BondBackup  = "active-backup"
	BondBalance = "balance-xor"
	BondPairNum = 2
)

const (
	EventFailForCreatepod   = "CreatePodNetworkFailed."
	EventFailForGetpod      = "GetPodFailed."
	EventFailForParseconfig = "ParseBluePrintOrConfigFailed."
	EventFailForNorthport   = "NorthPort."
	EventFailForSouthport   = "SouthPort."
)

const HTTPDefaultTimeoutInSec = 60 // default http GET/POST request timeout in second

const PaaSTenantAdminDefaultUUID = "admin"

const SKIP string = "REPLICATE-SKIP"

const LocalDBDataDir = "/root/nwnode/data-dir"

const LogicalPortDefaultVnicType = "normal"

const (
	VnicType = "vnic"
	VfType   = "vf"
	VethType = "veth"
	PfType   = "pf"
	Br0Vnic  = "br0_vnic"
	C0Vf     = "c0_vf"
	C0Vnic   = "c0_vnic"
)

const (
	AgentGetPodRetryTimes = 20
	AgentGetPodWaitSecond = 5
)
