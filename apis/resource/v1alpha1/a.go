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

/*import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
)

// KeyValue defines a key-value pair.
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PipelineStep defines a step in a Pipeline.
type PipelineStep struct {
	Name   string     `json:"name"`
	Values []KeyValue `json:"values"`
}

// PipelineContext defines a context in a Pipeline.
type PipelineContext struct {
	Values []KeyValue `json:"values"`
}

// PipelineTrigger as per CodeFresh API spec.
type PipelineTrigger struct {
	Name                       string             `json:"name"`
	Type                       string             `json:"type"`
	Repo                       string             `json:"repo"`
	Events                     []string           `json:"events"`
	PullRequestAllowForkEvents bool               `json:"pullRequestAllowForkEvents"`
	CommentRegex               string             `json:"commentRegex"`
	BranchRegex                string             `json:"branchRegex"`
	BranchRegexInput           string             `json:"branchRegexInput"`
	Provider                   string             `json:"provider"`
	Disabled                   bool               `json:"disabled"`
	Options                    PipelineOptions    `json:"options"`
	Context                    string             `json:"context"`
	Contexts                   []PipelineContext  `json:"contexts"`
	Variables                  []PipelineVariable `json:"variables"`
}

// PipelineCronTrigger as per CodeFresh API spec.
type PipelineCronTrigger struct {
	Event        string             `json:"event"`
	Name         string             `json:"name"`
	Type         string             `json:"type"`
	Message      string             `json:"message"`
	Expression   string             `json:"expression"`
	Verified     bool               `json:"verified"`
	Status       string             `json:"status"`
	Disabled     bool               `json:"disabled"`
	GitTriggerId string             `json:"gitTriggerId"`
	Branch       string             `json:"branch"`
	Options      PipelineOptions    `json:"options"`
	Variables    []PipelineVariable `json:"variables"`
}

// PipelineOptions as per CodeFresh API spec.
type PipelineOptions struct {
	NoCache             bool `json:"noCache"`
	NoCfCache           bool `json:"noCfCache"`
	ResetVolume         bool `json:"resetVolume"`
	EnableNotifications bool `json:"enableNotifications"`
}

// PipelineVariable as per CodeFresh API spec.
type PipelineVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PipelineParameters are the configurable fields of a Pipeline.
type PipelineParameters struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Triggers     []PipelineTrigger       `json:"triggers,omitempty"`
		CronTriggers []PipelineCronTrigger   `json:"cronTriggers,omitempty"`
		Steps        map[string]PipelineStep `json:"steps,omitempty"`
		Stages       []string                `json:"stages,omitempty"`
		Variables    []PipelineVariable      `json:"variables,omitempty"`
		Options      PipelineOptions         `json:"options,omitempty"`
		Contexts     []PipelineContext       `json:"contexts,omitempty"`
	} `json:"spec"`
}

// PipelineStepResponse defines a step in the Pipeline response.
type PipelineStepResponse struct {
	Title            string              `json:"title"`
	Type             string              `json:"type"`
	WorkingDirectory string              `json:"workingDirectory"`
	Arguments        map[string][]string `json:"arguments"`
}

// PipelineSpecResponse defines the spec part of the Pipeline response.
type PipelineSpecResponse struct {
	Triggers          map[string][]string             `json:"triggers"`
	Stages            []string                        `json:"stages"`
	Variables         map[string][]string             `json:"variables"`
	Contexts          map[string][]string             `json:"contexts"`
	TerminationPolicy map[string][]string             `json:"terminationPolicy"`
	ExternalResources map[string][]string             `json:"externalResources"`
	Steps             map[string]PipelineStepResponse `json:"steps"`
}

// PipelineMetadata holds metadata information of a pipeline.
type PipelineMetadata struct {
	Name               string                 `json:"name"`
	Project            string                 `json:"project"`
	ProjectId          string                 `json:"projectId"`
	Revision           int                    `json:"revision"`
	AccountId          string                 `json:"accountId"`
	CreatedAt          string                 `json:"created_at"`
	UpdatedAt          string                 `json:"updated_at"`
	Deprecate          map[string]interface{} `json:"deprecate"`
	Labels             map[string][]string    `json:"labels"`
	OriginalYamlString string                 `json:"originalYamlString"`
	ID                 string                 `json:"id"`
}

// PipelineDocument represents a single pipeline document in the response.
type PipelineDocument struct {
	Metadata     PipelineMetadata `json:"metadata"`
	Version      string           `json:"version"`
	Kind         string           `json:"kind"`
	Spec         PipelineSpec     `json:"spec"`
	LastExecuted string           `json:"last_executed"`
}

type PipelineDetails struct {
	Docs  []PipelineDocument `json:"docs"`
	Count int                `json:"count"`
}

type PipelineCreateParams struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec PipelineSpec `json:"spec"`
}

// CreatePipelineResponse represents the response from creating a pipeline.
type CreatePipelineResponse struct {
	Metadata struct {
		Name               string              `json:"name"`
		Project            string              `json:"project"`
		ProjectId          string              `json:"projectId"`
		Revision           int                 `json:"revision"`
		AccountId          string              `json:"accountId"`
		Labels             map[string][]string `json:"labels"`
		OriginalYamlString string              `json:"originalYamlString"`
		CreatedAt          string              `json:"created_at"`
		UpdatedAt          string              `json:"updated_at"`
		ID                 string              `json:"id"`
	} `json:"metadata"`
	Version string       `json:"version"`
	Kind    string       `json:"kind"`
	Spec    PipelineSpec `json:"spec"`
}

// PipelineObservation are the observable fields of a Pipeline.
type PipelineObservation struct {
	Metadata struct {
		Name string `json:"name"`
		ID   string `json:"id""`
	} `json:"metadata"`
	Version string               `json:"version"`
	Kind    string               `json:"kind"`
	Spec    PipelineSpecResponse `json:"spec"`
}

// A PipelineSpec defines the desired state of a Pipeline.
type PipelineSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       PipelineParameters `json:"forProvider"`
}

// A PipelineStatus represents the observed state of a Pipeline.
type PipelineStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          PipelineObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Pipeline is a managed resource that represents a CodeFresh Pipeline.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,codefresh}
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec"`
	Status PipelineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PipelineList contains a list of Pipeline
type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

// Pipeline type metadata.
var (
	PipelineKind             = reflect.TypeOf(Pipeline{}).Name()
	PipelineGroupKind        = schema.GroupKind{Group: Group, Kind: PipelineKind}.String()
	PipelineKindAPIVersion   = PipelineKind + "." + SchemeGroupVersion.String()
	PipelineGroupVersionKind = SchemeGroupVersion.WithKind(PipelineKind)
)

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}
*/
