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
	PaaSTenantAdminDefaultUUID = "admin"

	DefaultProviderPhysicalNetwork = "physnet0"
	DefaultProviderNetworkType     = "vlan"
	ProviderNetworkTypeVlan        = "vlan"
	VlanTransparentSegmentationID  = "0"
	MinVlanID                      = 1
	MaxVlanID                      = 4095

	NetworkNotFound = "NetworkNotFound"
	NetworkInUse    = "NetworkInUse"
	// todo noauth delete network should return body in future to remove this const string
	NoauthNetworkInUse = "Iaas return code 409"
	// embedded-server error string
	OverlayNetworkInUse = "Exist-port-on-subnet"

	GetRestoreStateRetryTimes        = 30
	GetLoadReourceRetryIntervalInSec = 5

	NetworkStatActive = "ACTIVE"
	NetworkStatDown   = "DOWN"

	DefaultIaasTenantID     = "defaultt_iaas_tenentid"
	DefaultIaasTenantName   = "default_iaas_tenent_name"
	ErrOfIaasInterfaceNil   = "Iaas Interface is nil"
	ErrOfIaasTenantNotExist = "Iaas Tenent is not exist"
	TestPaasTenantID        = "paas-admin"
	DefaultPaaSNetwork      = "net_api"
	DefaultPaaSCidr         = "192.168.0.0/16"
)

// const defined for restore state
const (
	RestoreProgressStatProcessFail = -1
	RestoreProgressStatProcessing  = 0
	RestoreProgressStatProcessSucc = 1
)

const (
	OwnerTypePod           = "Pod"
	OwnerTypeNode          = "Node"
	OwnerTypePaaSComponent = "PaaSComponent"
)

const (
	LogicalPortDefaultVnicType = "normal"
)

const (
	StrConstTrue  = "true"
	StrConstFalse = "false"
)

const (
	IpsTypeIpgroup = "ipgroup"
)

const (
	KnitterJSONPath = "/opt/cni/manager/knitter.json"
	TECS            = "TECS"
	EMBEDDED        = "EMBEDDED"
	VNM             = "vNM"
	VNFM            = "VNFM"
)

//operation log report
const (
	//paas.json
	ResultSuccess = "00000001"
	ResultFail    = "00000002"

	RiskLevelHigh   = "00000011"
	RiskLevelMedium = "00000012"
	RiskLevellow    = "00000013"

	OperationTypeApplication = "00000021"
	OperationTypeNode        = "00000022"
	OperationTypeProject     = "00000023"

	ActionTypeCreate = "00000101"
	ActionTypeDelete = "00000102"
	ActionTypeUpdate = "00000103"
	ActionTypeQuery  = "00000104"

	OperationTypeNetworkData    = "07000001"
	OperationTypeConfigurations = "07000002"
	OperationTypeGateway        = "07000003"
	OperationTypeIpgroup        = "07000004"
	OperationTypeNetwork        = "07000005"
	OperationTypePhysnet        = "07000006"
	OperationTypeTenant         = "07000007"

	//todo
	OperationDescriptionBackup                   = "07001001"
	OperationDescriptionRestore                  = "07001002"
	OperationDescriptionInjectionConfiguration   = "07001003"
	OperationDescriptionInitConfiguration        = "07001004"
	OperationDescriptionUpdateGatewayInformation = "07001005"
	OperationDescriptionCreateIpgroup            = "07001006"
	OperationDescriptionUpdateIpgroup            = "07001007"
	OperationDescriptionDeleteIpgroup            = "07001008"
	OperationDescriptionCreateNetwork            = "07001009"
	OperationDescriptionDeleteNetwork            = "07001010"
	OperationDescriptionUpdatePhysicalNetwork    = "07001011"
	OperationDescriptionCreateTenant             = "07001012"
	OperationDescriptionDeleteTenant             = "07001013"
	OperationDescriptionUpdateTenantQuota        = "07001014"
	//todo auto generation
)
