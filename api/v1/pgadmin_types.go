/*
Copyright © contributors to CloudNativePG, established as
CloudNativePG a Series of LF Projects, LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PgAdminSpec defines the desired state of PgAdmin
type PgAdminSpec struct {
	// DefaultEmail is the email used for the default pgAdmin account.
	// +kubebuilder:default:=admin@example.com
	DefaultEmail string `json:"defaultEmail,omitempty"`

	// DefaultPassword is the password for the default pgAdmin account.
	// +kubebuilder:default:=admin
	DefaultPassword string `json:"defaultPassword,omitempty"`

	// Replicas is the number of pgAdmin instances.
	// +kubebuilder:default:=1
	Replicas *int32 `json:"replicas,omitempty"`

	// Image is the container image for pgAdmin.
	// +kubebuilder:default:="image: ghcr.io/haneeshpld/pgadmin4-nonroot:latest"
	Image string `json:"image,omitempty"`
}

// PgAdminStatus defines the observed state of PgAdmin
type PgAdminStatus struct {
	// You can add status fields here.
	// For example: Conditions, Deployment status, etc.
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PgAdmin is the Schema for the pgadmins API
type PgAdmin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PgAdminSpec   `json:"spec,omitempty"`
	Status PgAdminStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PgAdminList contains a list of PgAdmin
type PgAdminList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PgAdmin `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PgAdmin{}, &PgAdminList{})
}
