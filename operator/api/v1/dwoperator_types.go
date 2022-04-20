/*
Copyright 2022.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DwOperatorSpec defines the desired state of DwOperator
type DwOperatorSpec struct {
	// DeploymentName is the identity of the dwserver deployment
	DeploymentName string `json:"name"`
	// Namespace of the dw deployment, if omitted, the default value is the default namespace
	Namespace string `json:"namespace"`
	//Replicas defines number of replicas for the dwserver
	Replicas int32 `json:"replicas"`
}

// DwOperatorStatus defines the observed state of DwOperator
type DwOperatorStatus struct {
	TotalScheduled int32            `json:"total_scheduled"`
	TotalRunning   int32            `json:"total_running"`
	ZoneRunning    map[string]int32 `json:"availability_zone"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DwOperator is the Schema for the dwoperators API
type DwOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DwOperatorSpec   `json:"spec,omitempty"`
	Status DwOperatorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DwOperatorList contains a list of DwOperator
type DwOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DwOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DwOperator{}, &DwOperatorList{})
}
