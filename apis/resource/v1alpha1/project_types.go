/*
Copyright 2022 The Crossplane Authors.

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

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type ProjectVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ProjectMetadata struct {
	CreatedAt string `json:"createdAt"`
}

type ProjectDetails struct {
	AccountID                   string            `json:"accountId"`
	ProjectName                 string            `json:"projectName"`
	UpdatedAt                   string            `json:"updatedAt"`
	ProjectMetadata             ProjectMetadata   `json:"metadata"`
	ProjectImage                string            `json:"image"`
	ProjectTags                 []string          `json:"tags"`
	ProjectVariables            []ProjectVariable `json:"variables"`
	ProjectTotalPipelinesNumber int               `json:"pipelinesNumber"`
	ProjectID                   string            `json:"id"`
	IsFavorite                  bool              `json:"favorite"`
}

type ProjectCreateParams struct {
	ProjectName      string            `json:"projectName,omitempty"`
	ProjectTags      []string          `json:"tags,omitempty"`
	ProjectVariables []ProjectVariable `json:"variables,omitempty"`
}

type CreateProjectResponse struct {
	ProjectID string `json:"id"`
}

// ProjectParameters are the configurable fields of a Project.
type ProjectParameters struct {
	ConfigurableField string `json:"configurableField"`
	ProjectName       string `json:"projectName,omitempty"`
	// *optional
	ProjectTags []string `json:"projectTags,omitempty"`
	// *optional
	ProjectVariables []ProjectVariable `json:"projectVariables,omitempty"`
}

// ProjectObservation are the observable fields of a Project.
type ProjectObservation struct {
	ObservableField string `json:"observableField,omitempty"`
	ProjectID       string `json:"projectId,omitempty"`
}

// A ProjectSpec defines the desired state of a Project.
type ProjectSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ProjectParameters `json:"forProvider"`
}

// A ProjectStatus represents the observed state of a Project.
type ProjectStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ProjectObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Project is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,codefresh}
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectSpec   `json:"spec"`
	Status ProjectStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProjectList contains a list of Project
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}

// Project type metadata.
var (
	ProjectKind             = reflect.TypeOf(Project{}).Name()
	ProjectGroupKind        = schema.GroupKind{Group: Group, Kind: ProjectKind}.String()
	ProjectKindAPIVersion   = ProjectKind + "." + SchemeGroupVersion.String()
	ProjectGroupVersionKind = SchemeGroupVersion.WithKind(ProjectKind)
)

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
}
