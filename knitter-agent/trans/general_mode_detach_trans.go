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

func NewGeneralModeDetachTrans() *transdsl.Transaction {
	trans := &transdsl.Transaction{
		Fragments: []transdsl.Fragment{
			&transdsl.Optional{Spec: new(context.IsDeployPod),
				Fragment: new(context.EndAction)},
			new(context.GetPodAction),
			&transdsl.Optional{Spec: new(context.IsK8SRecycleResource),
				Fragment: newK8SRecycleResourceProcedure()},
			newGeneralModePhysicalResourceCleanupProcedure(),
			new(context.DeleteLogicPortsForPodAction),
			//new(context.ReportDeletePodToManagerAction),
		},
	}
	return trans
}

func newRemovePortFromPodProcedure() transdsl.Fragment {
	procedure := &transdsl.Procedure{
		Fragments: []transdsl.Fragment{
			newDetachPodFromBridgeProcedure(),
			&transdsl.Optional{Spec: new(context.IsTenantNetworkNeedDel),
				Fragment: new(context.RemoveNetFromFlowMgrAction)}}}
	return procedure
}

func newDetachPodFromBridgeProcedure() transdsl.Fragment {
	procedure := &transdsl.Procedure{
		Fragments: []transdsl.Fragment{
			new(context.DetachPortFromBrintAction),
			new(context.DestroyVethPairAction),
			new(context.DestroyNeutronPortAction)}}
	return procedure
}

func newGeneralModePhysicalResourceCleanupProcedure() transdsl.Fragment {
	procedure := &transdsl.Procedure{
		Fragments: []transdsl.Fragment{
			new(context.GetAllPortsOfPodAction),
			&transdsl.Repeat{FuncVar: newRemovePortFromPodProcedure},
			new(context.DelPodAction),
			new(context.CleanPhysicalResourceRecordAction),
		},
	}
	return procedure
}

func newK8SRecycleResourceProcedure() transdsl.Fragment {
	procedure := &transdsl.Procedure{
		Fragments: []transdsl.Fragment{
			new(context.K8SRecycleResourceAction),
			new(context.EndAction),
		},
	}
	return procedure
}
