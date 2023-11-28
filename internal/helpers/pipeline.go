package helpers

/* import "crossplane-provider-codefresh/apis/resource/v1alpha1"

func comparePipelines(k8sPipeline *v1alpha1.Pipeline, apiPipeline *CodeFreshPipeline) bool {
	// Compare simple fields
	if k8sPipeline.Spec.ForProvider.Metadata.Name != apiPipeline.Name {
		return false
	}

	// Compare complex fields
	if !compareTriggers(k8sPipeline.Spec.ForProvider.Spec.Triggers, apiPipeline.Triggers) ||
		!compareCronTriggers(k8sPipeline.Spec.ForProvider.Spec.CronTriggers, apiPipeline.CronTriggers) ||
		!compareSteps(k8sPipeline.Spec.ForProvider.Spec.Steps, apiPipeline.Steps) ||
		!compareStringSlices(k8sPipeline.Spec.ForProvider.Spec.Stages, apiPipeline.Stages) ||
		!compareVariables(k8sPipeline.Spec.ForProvider.Spec.Variables, apiPipeline.Variables) ||
		!compareOptions(k8sPipeline.Spec.ForProvider.Spec.Options, apiPipeline.Options) ||
		!compareContexts(k8sPipeline.Spec.ForProvider.Spec.Contexts, apiPipeline.Contexts) {
		return false
	}

	return true
}

func compareTriggers(k8sTriggers []v1alpha1.PipelineTrigger, apiTriggers []CodeFreshTrigger) bool {
	if len(k8sTriggers) != len(apiTriggers) {
		return false
	}

	for i, k8sTrigger := range k8sTriggers {
		apiTrigger := apiTriggers[i]
		if k8sTrigger.Name != apiTrigger.Name || // other simple fields comparison
			!compareStringSlices(k8sTrigger.Events, apiTrigger.Events) {
			return false
		}
	}

	return true
}

func compareOptions(k8sOptions v1alpha1.PipelineOptions, apiOptions CodeFreshOptions) bool {
	// Direct comparison of all options fields
	return k8sOptions.NoCache == apiOptions.NoCache &&
		k8sOptions.NoCfCache == apiOptions.NoCfCache &&
		k8sOptions.ResetVolume == apiOptions.ResetVolume &&
		k8sOptions.EnableNotifications == apiOptions.EnableNotifications
}

func compareStringSlices(slice1 []string, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}

	return true
}
*/
