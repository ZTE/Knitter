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

package errobj

import (
	"errors"
)

func IsEqual(errLeft, errRight error) bool {
	return errLeft.Error() == errRight.Error()
}

var (
	ErrAny                              = errors.New("any")
	Err403                              = errors.New("bad request json body")
	Err406                              = errors.New("create port err")
	ErrOpenstackCreateBulkPortsFailed   = errors.New("openStack createbulkports err")
	ErrMarshalFailed                    = errors.New("json.Marshal Err")
	ErrUnmarshalFailed                  = errors.New("json.Unmarshal Err")
	ErrRestfulPostFailed                = errors.New("restful post Err")
	ErrHTTPPostStatusCode               = errors.New("restful post status code err")
	ErrRequestNeedAdminPermission       = errors.New("request need admin permission")
	ErrCheckAllocationPools             = errors.New("allocation pools check error")
	ErrTenantHasPodsInUse               = errors.New("tenant has pods in use")
	ErrNetworkHasPortsInUse             = errors.New("network has ports in use")
	ErrNetworkHasIGsInUse               = errors.New("network has ip groups in use")
	ErrEtcdRestoreFromEtcdAdmToolFailed = errors.New("etcd restore from EtcdAdmTool failed err")
	ErrArgTypeMismatch                  = errors.New("argument type mismatch")
)

var (
	ErrJSON  = errors.New("bad request json body")
	ErrAuth  = errors.New("auth err")
	ErrDoing = errors.New("request is doning")
	ErrTmOut = errors.New("request time out")
)

var (
	ErrObjectPointerIsNil = errors.New("object pointer is nil")

	ErrRecordAlreadyExist = errors.New("record already exsit")
	ErrRecordNotExist     = errors.New("record already exsit")
	ErrObjectTypeMismatch = errors.New("object type mismatch")
)

var (
	ErrNetworkNotExist             = errors.New("network not exist")
	ErrNetworkTypeNotSupported     = errors.New("network type not supported")
	ErrVlanTransparentConflictArgs = errors.New("vlan_transparent confict args")
	ErrInvalidVlanID               = errors.New("invalid vlan ID")
)

var (
	ErrRequiredFieldIsNil = errors.New("required field is nil")
	ErrGatewayTypeError   = errors.New("Gateway type error")
)
