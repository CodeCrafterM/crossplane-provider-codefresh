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

package project

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"crossplane-provider-codefresh/internal/constants"

	"crossplane-provider-codefresh/apis/resource/v1alpha1"
	"crossplane-provider-codefresh/internal/client"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
)

// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

func TestObserve(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		setup  func(*client.MockCodeFreshAPIClient)
		want   want
	}{
		"ProjectDoesNotExist": {
			reason: "Should return ResourceDoesNotExist when project does not exist.",
			args: args{
				ctx: context.TODO(),
				mg: &v1alpha1.Project{
					Status: v1alpha1.ProjectStatus{
						AtProvider: v1alpha1.ProjectObservation{
							ProjectID: "non-existent-project",
						},
					},
				},
			},
			setup: func(m *client.MockCodeFreshAPIClient) {
				m.MockGetResourceResponse = nil
				m.MockGetResourceErr = client.ErrResourceNotFound
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
		"ProjectExists": {
			reason: "Should return ResourceExists when project exists.",
			args: args{
				ctx: context.TODO(),
				mg: &v1alpha1.Project{
					Spec: v1alpha1.ProjectSpec{
						ForProvider: v1alpha1.ProjectParameters{
							ProjectName: "TestProject",
							ProjectTags: []string{"tag1", "tag2"},
							ProjectVariables: []v1alpha1.ProjectVariable{
								{Key: "var1", Value: "value1"},
								{Key: "var2", Value: "value2"},
							},
						},
					},
					Status: v1alpha1.ProjectStatus{
						AtProvider: v1alpha1.ProjectObservation{
							ProjectID: "existing-project",
						},
					},
				},
			},
			setup: func(m *client.MockCodeFreshAPIClient) {
				m.MockGetResourceResponse = &constants.ProjectDetails{
					ProjectID:   "existing-project",
					ProjectName: "TestProject",
					ProjectTags: []string{"tag1", "tag2"},
					ProjectVariables: []constants.ProjectVariable{
						{Key: "var1", Value: "value1"},
						{Key: "var2", Value: "value2"},
					},
				}
				m.MockGetResourceErr = nil
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: managed.ConnectionDetails{}, // this line is updated
				},
				err: nil,
			},
		},
		// Add more test cases as needed.
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &client.MockCodeFreshAPIClient{}
			tc.setup(mockClient)
			e := external{service: mockClient}
			got, err := e.Observe(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}
