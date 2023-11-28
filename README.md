# CodeFresh CrossPlane Provider

## Description
This CrossPlane provider for CodeFresh CI/CD Platform encapsulates a comprehensive approach to manage CodeFresh resources through Kubernetes. It leverages a robust generic CodeFresh client for CRUD operations, enhancing Kubernetes' capabilities to interact seamlessly with CodeFresh's CI/CD resources. Currently, the provider includes advanced controllers for Project and Pipeline resources. The Project controller is fully functional with extensive testing for the observe method, while the Pipeline controller supports creation and deletion, with plans to integrate update functionality and detailed observe logic in future updates.

## Setup

Setting up the CodeFresh CrossPlane Provider involves a straightforward process to integrate it seamlessly with your Kubernetes environment. Here's how to get started:
1. <b>Initialize Development Environment:</b> Run make dev to provision a kind cluster and apply the CRDs necessary for the plugin, which are located in package/crds.
2. <b>Install Crossplane Controller:</b> Deploy the Crossplane controller on your Kubernetes cluster to manage the lifecycle of external resources.
3. <b>Run the Provider:</b> Execute go run cmd/provider/main.go --debug to start the CodeFresh provider with debug logging, facilitating real-time monitoring and troubleshooting.
4. <b>Configure API Access:</b>
   - Generate an API token from CodeFresh and encode it to base64.
   - Update the provider configuration in examples/provider/config.yaml with the encoded token. You can use tools like Sealed Secrets or Vault for enhanced secrets management.
5. <b>Customize Pipeline Manifest:</b> Edit the examples/pipeline/pipeline.yaml to align with your specific requirements, particularly focusing on the repository owner and name.
6. <b>Apply Provider and Pipeline Manifests:</b> Deploy the provider and pipeline configurations to your cluster using:
    - ```kubectl apply -f examples/provider/config.yaml```
    - ```kubectl apply -f examples/pipeline/pipeline.yaml```

## How to Use

The CodeFresh CrossPlane Provider bridges Kubernetes with CodeFresh, a robust CI/CD platform that streamlines the process of building, testing, and deploying applications. CodeFresh's unique approach to continuous integration and delivery allows for highly customizable pipelines, ensuring a flexible and scalable deployment process.

To make the most of this integration:
- Familiarize yourself with [CodeFresh's documentation](https://codefresh.io/docs/), which provides comprehensive guidance on creating and managing pipelines within their platform.
- Leverage the CrossPlane provider to manage CodeFresh resources as custom resources within Kubernetes. This enables a Kubernetes-native approach to CI/CD, aligning with cloud-native best practices.
- Explore the examples provided in the setup process to understand how to define and manage CodeFresh pipelines as custom resources in Kubernetes.

By integrating CodeFresh with CrossPlane, you gain the ability to manage CI/CD processes more efficiently, bringing the power of Kubernetes and the flexibility of CodeFresh together in a cohesive workflow.

## Development

We encourage community contributions to enrich and expand the capabilities of this provider. Developers interested in contributing can follow the provided guidelines to set up their development environment, run tests, and contribute code. (TODO: Provide detailed instructions for setting up a development environment, running tests, and contributing code.)
## Code of Conduct

This project adheres to a strict code of conduct. Participants are expected to uphold the values and standards outlined in the code. (TODO: Provide a link to the full code of conduct.)
## Current Status
-  Developed a generic CRUD client for CodeFresh, allowing for streamlined interactions with the platform.
-  Implemented Project and Pipeline resource controllers. The Project controller includes complete functionalities with unit tests covering the observe method.
-  The Pipeline controller is structured to enable resource creation and deletion. Future updates will introduce update functionality and a comprehensive observe logic to enhance its integration.
-  Future work includes expanding unit tests for all methods, and detailed documentation.

## TODO
- Enhance the CodeFresh client with additional features and capabilities.
- Expand unit testing coverage to include all methods and scenarios.
- Integrate support for more CodeFresh resources, broadening the scope of the provider.
- Develop comprehensive documentation and examples to guide users in leveraging the provider effectively.


# Resources

Refer to Crossplane's [CONTRIBUTING.md] file for more information on how the
Crossplane community prefers to work. The [Provider Development][provider-dev]
guide may also be of use.

This README is a dynamic document and will be regularly updated to reflect the latest advancements and additions to the project.

[CONTRIBUTING.md]: https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md
[provider-dev]: https://github.com/crossplane/crossplane/blob/master/contributing/guide-provider-development.md
