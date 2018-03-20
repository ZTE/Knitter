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

package trans

import (
	"github.com/ZTE/Knitter/knitter-agent/context"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

func NewGeneralModeAttachTrans() *transdsl.Transaction {
	trans := &transdsl.Transaction{
		Fragments: []transdsl.Fragment{
			&transdsl.Optional{Spec: new(context.IsDeployPod),
				Fragment: newConnectToBrdepProcedure()},
			new(context.GetPodAction),
			new(context.GetNetworkAttrsAction),
			&transdsl.Repeat{FuncVar: newAddPortToPodProcedure},
			new(context.SavePodAction),
		},
	}
	return trans
}

func newConnectToBrdepProcedure() transdsl.Fragment {
	procedure := &transdsl.Procedure{
		Fragments: []transdsl.Fragment{
			new(context.ConnectToBrdepAction),
			new(context.EndAction)}}
	return procedure
}

func newAddPortToPodProcedure() transdsl.Fragment {
	procedure := &transdsl.Procedure{
		Fragments: []transdsl.Fragment{
			&transdsl.Optional{Spec: new(context.IsTenantNetworkNotExist),
				Fragment: newAttachTenantNetworkToBridgeProcedure()},
			newAttachPodToBridgeProcedure()}}
	return procedure
}

func newAttachTenantNetworkToBridgeProcedure() transdsl.Fragment {
	procedure := &transdsl.Procedure{
		Fragments: []transdsl.Fragment{
			new(context.GetVniAction),
			new(context.AddNetToFlowMgrAction)}}
	return procedure
}

func newAttachPodToBridgeProcedure() transdsl.Fragment {
	procedure := &transdsl.Procedure{
		Fragments: []transdsl.Fragment{
			new(context.GeneralModeGetMgrPortAction),
			new(context.CreateVethPairAction),
			new(context.AttachPortToBrIntAction),
			new(context.AttachPortToPodAction),
			new(context.SaveVethToLocalDBAction)}}
	return procedure
}
