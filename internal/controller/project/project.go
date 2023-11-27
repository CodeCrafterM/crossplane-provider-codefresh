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

	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"crossplane-provider-codefresh/apis/resource/v1alpha1"
	apisv1alpha1 "crossplane-provider-codefresh/apis/v1alpha1"
	"crossplane-provider-codefresh/internal/features"

	codefreshclient "crossplane-provider-codefresh/internal/client"
	"crossplane-provider-codefresh/internal/constants"
	"crossplane-provider-codefresh/internal/helpers"
)

// Setup adds a controller that reconciles Project managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ProjectGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ProjectGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:  mgr.GetClient(),
			usage: resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: func(creds []byte, logger logging.Logger) (interface{}, error) {
				return codefreshclient.NewCodeFreshService(creds, o.Logger)
			},
			logger: o.Logger.WithValues("controller", name),
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.Project{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(creds []byte, logger logging.Logger) (interface{}, error)
	logger       logging.Logger
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return nil, errors.New(constants.ErrNotProject)
	}

	c.logger.Info("Connecting to CodeFresh", "project", cr.GetName())

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, constants.ErrTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, constants.ErrGetPC)
	}

	cd := pc.Spec.Credentials
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, constants.ErrGetCreds)
	}

	svc, err := c.newServiceFn(data, c.logger)
	if err != nil {
		return nil, errors.Wrap(err, constants.ErrNewClient)
	}

	// Assert the type of svc to CodeFreshAPI
	service, ok := svc.(codefreshclient.CodeFreshAPI)
	if !ok {
		return nil, errors.New(constants.ErrAssertCodeFreshService)
	}

	return newExternal(c.kube, c.logger, service), nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client client.Client
	/*service interface{}*/
	service codefreshclient.CodeFreshAPI
	logger  logging.Logger
}

func newExternal(client client.Client, logger logging.Logger, service codefreshclient.CodeFreshAPI) *external {
	return &external{
		client:  client,
		logger:  logger,
		service: service,
	}
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalObservation{}, errors.New(constants.ErrNotProject)
	}

	if !ok {
		return managed.ExternalObservation{}, errors.New(constants.ErrExpectedCodeFreshClient)
	}

	projectID := cr.Status.AtProvider.ProjectID
	if projectID == "" {
		// No project ID means the project hasn't been created yet.
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	var projectDetails constants.ProjectDetails
	err := c.service.GetResource(ctx, "projects", projectID, &projectDetails)
	if err != nil {
		// Check if the error is due to the project not being found
		if errors.Is(err, codefreshclient.ErrResourceNotFound) || helpers.IsResourceNotFoundErr(err, "Project") {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, err
	}

	var variables []constants.ProjectVariable //nolint:prealloc
	for _, v := range cr.Spec.ForProvider.ProjectVariables {
		variables = append(variables, constants.ProjectVariable{Key: v.Key, Value: v.Value})
	}

	// Check if the project name, tags, and variables are up to date
	nameUpToDate := projectDetails.ProjectName == cr.Spec.ForProvider.ProjectName
	tagsUpToDate := helpers.AreTagsEqual(projectDetails.ProjectTags, cr.Spec.ForProvider.ProjectTags)
	varsUpToDate := helpers.AreSlicesEqual(projectDetails.ProjectVariables, variables)

	resourceUpToDate := nameUpToDate && tagsUpToDate && varsUpToDate

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  resourceUpToDate,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalCreation{}, errors.New(constants.ErrNotProject)
	}

	if !ok {
		return managed.ExternalCreation{}, errors.New(constants.ErrExpectedCodeFreshClient)
	}

	var variables []constants.ProjectVariable //nolint:prealloc
	for _, v := range cr.Spec.ForProvider.ProjectVariables {
		variables = append(variables, constants.ProjectVariable{
			Key:   v.Key,
			Value: v.Value,
		})
	}

	// Set up the parameters for project creation
	params := constants.ProjectCreateParams{
		ProjectName:      cr.Spec.ForProvider.ProjectName,
		ProjectTags:      cr.Spec.ForProvider.ProjectTags,
		ProjectVariables: variables,
	}

	// Response struct to hold the created project's ID
	var respData constants.CreateProjectResponse

	// Use the generic CreateResource method
	if err := c.service.CreateResource(ctx, "projects", params, &respData); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, constants.ErrCreatingProject)
	}

	// Store the project ID in the status
	cr.Status.AtProvider.ProjectID = respData.ProjectID

	// Update the status of the resource with the new project ID
	if err := c.client.Status().Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, constants.ErrUpdatingProjectStatus)
	}

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(constants.ErrNotProject)
	}

	if !ok {
		return managed.ExternalUpdate{}, errors.New(constants.ErrExpectedCodeFreshClient)
	}

	// Convert ProjectVariables to the expected format
	var variables []constants.ProjectVariable //nolint:prealloc
	for _, v := range cr.Spec.ForProvider.ProjectVariables {
		variables = append(variables, constants.ProjectVariable{Key: v.Key, Value: v.Value})
	}

	// Define the update parameters including tags and variables
	updateParams := map[string]interface{}{
		"projectName": cr.Spec.ForProvider.ProjectName,
		"tags":        cr.Spec.ForProvider.ProjectTags,
		"variables":   variables,
	}

	// Update the resource
	err := c.service.UpdateResource(ctx, "projects", cr.Status.AtProvider.ProjectID, updateParams, nil)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, constants.ErrUpdatingProject)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return errors.New(constants.ErrNotProject)
	}

	if !ok {
		return errors.New(constants.ErrExpectedCodeFreshClient)
	}

	// Delete the resource
	err := c.service.DeleteResource(ctx, "projects", cr.Status.AtProvider.ProjectID)
	if err != nil {
		return errors.Wrap(err, constants.ErrDeletingProject)
	}

	return nil
}
