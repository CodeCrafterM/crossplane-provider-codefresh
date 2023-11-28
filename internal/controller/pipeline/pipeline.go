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

package pipeline

import (
	"context"
	"fmt"

	"github.com/crossplane/crossplane-runtime/pkg/logging"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"crossplane-provider-codefresh/apis/resource/v1alpha1"
	apisv1alpha1 "crossplane-provider-codefresh/apis/v1alpha1"
	"crossplane-provider-codefresh/internal/features"
	"crossplane-provider-codefresh/internal/helpers"

	codefreshclient "crossplane-provider-codefresh/internal/client"
	"crossplane-provider-codefresh/internal/constants"
)

const (
	errNotPipeline        = "managed resource is not a Pipeline custom resource"
	errorFetchingPipeline = "Error occurred while fetching pipeline details"
	errCreatingPipeline   = "error creating pipeline"
	errUpdatingPipeline   = "error updating pipeline"
	errDeletingPipeline   = "something went wrong while deleting the pipeline"

	debugObservingPipelineResource = "Observing Pipeline resource"
	debugPipelineIDNotFound        = "Pipeline ID not found in status; pipeline resource not created yet"
)

// Setup adds a controller that reconciles Pipeline managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.PipelineGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.PipelineGroupVersionKind),
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
		For(&v1alpha1.Pipeline{}).
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
	cr, ok := mg.(*v1alpha1.Pipeline)
	if !ok {
		return nil, errors.New(errNotPipeline)
	}

	c.logger.Info("Connecting to CodeFresh", "pipeline", cr.GetName())

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
	cr, ok := mg.(*v1alpha1.Pipeline)
	if !ok {
		c.logger.Debug(errNotPipeline, "actualType", fmt.Sprintf("%T", mg))
		return managed.ExternalObservation{}, errors.New(errNotPipeline)
	}

	c.logger.Debug(debugObservingPipelineResource, "name", cr.GetName())

	pipelineID := cr.Status.AtProvider.ID
	if pipelineID == "" {
		c.logger.Debug(debugPipelineIDNotFound)
		// No pipeline ID means the pipeline hasn't been created yet.
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	var pipelineDetails v1alpha1.PipelineDetails
	err := c.service.GetResource(ctx, "pipelines", pipelineID, &pipelineDetails)
	if err != nil {
		c.logger.Debug(errorFetchingPipeline, "error", err, "pipelineID", pipelineID)
		// Check if the error is due to the pipeline not being found
		if errors.Is(err, codefreshclient.ErrResourceNotFound) || helpers.IsResourceNotFoundErr(err, "Pipeline") {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, err
	}

	nameUpToDate := false
	if len(pipelineDetails.Docs) > 0 {
		nameUpToDate = pipelineDetails.Docs[0].Metadata.Name == cr.Spec.ForProvider.Metadata.Name
		c.logger.Debug("Comparing pipeline names", "observedName", pipelineDetails.Docs[0].Metadata.Name, "expectedName", cr.Spec.ForProvider.Metadata.Name)
	} else {
		c.logger.Debug("No documents found in pipeline details")
	}

	resourceUpToDate := nameUpToDate

	c.logger.Debug("Observed pipeline resource", "resourceUpToDate", resourceUpToDate)
	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  resourceUpToDate,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Pipeline)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPipeline)
	}

	// Set up the parameters for pipeline creation
	params := v1alpha1.PipelineCreateParams{
		Metadata: v1alpha1.PipelineMetadata{
			Name: cr.Spec.ForProvider.Metadata.Name,
		},
		Spec: v1alpha1.PipelineSpecStruct{
			Triggers:  cr.Spec.ForProvider.Spec.Triggers,
			Steps:     cr.Spec.ForProvider.Spec.Steps,
			Stages:    cr.Spec.ForProvider.Spec.Stages,
			Variables: cr.Spec.ForProvider.Spec.Variables,
			Options:   cr.Spec.ForProvider.Spec.Options,
			// Contexts:  cr.Spec.ForProvider.Spec.Contexts,
		},
	}

	// Response struct to hold the created pipeline's ID
	var respData v1alpha1.CreatePipelineResponse

	// Use the generic CreateResource method
	if err := c.service.CreateResource(ctx, "pipelines", params, &respData); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreatingPipeline)
	}

	// Store the pipeline ID in the status
	cr.Status.AtProvider.ID = respData.Metadata.ID
	/*	cr.Status.AtProvider.Name = respData.Metadata.Name */

	// Update the status of the resource with the new pipeline ID
	if err := c.client.Status().Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errUpdatingPipeline)
	}

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Pipeline)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPipeline)
	}

	fmt.Printf("Updating: %+v", cr)

	if !ok {
		return managed.ExternalUpdate{}, errors.New(constants.ErrExpectedCodeFreshClient)
	}

	/*	updateParams := v1alpha1.PipelineCreateParams{
			Metadata: v1alpha1.PipelineMetadata{
				Name: cr.Spec.ForProvider.Metadata.Name,
			},
			Spec: v1alpha1.PipelineSpecStruct{
				Triggers:  cr.Spec.ForProvider.Spec.Triggers,
				Steps:     cr.Spec.ForProvider.Spec.Steps,
				Stages:    cr.Spec.ForProvider.Spec.Stages,
				Variables: cr.Spec.ForProvider.Spec.Variables,
				Options:   cr.Spec.ForProvider.Spec.Options,
				// Contexts:  cr.Spec.ForProvider.Spec.Contexts,
			},
		}

		// Update the resource
		err := c.service.UpdateResource(ctx, "pipelines", cr.Status.AtProvider.ID, updateParams, nil)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdatingPipeline)
		}*/

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Pipeline)
	if !ok {
		return errors.New(constants.ErrExpectedCodeFreshClient)
	}

	// Delete the resource
	err := c.service.DeleteResource(ctx, "pipelines", cr.Status.AtProvider.ID)
	if err != nil {
		return errors.Wrap(err, errDeletingPipeline)
	}

	return nil
}

/*func transformTriggers(triggers []v1alpha1.PipelineTrigger) []v1alpha1.PipelineTrigger {
	var transformed []v1alpha1.PipelineTrigger
	for _, t := range triggers {
		// Copy the trigger and modify Contexts
		newTrigger := t
		newTrigger.Contexts = [][]v1alpha1.NullType{{v1alpha1.NullType{}}}

		transformed = append(transformed, newTrigger)
	}
	return transformed
}
*/
