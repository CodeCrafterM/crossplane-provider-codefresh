package constants

const (
	ErrNotProject      = "managed resource is not a Project custom resource"
	ErrCreatingProject = "error creating project in CodeFresh"
	ErrTrackPCUsage    = "cannot track ProviderConfig usage"
	ErrGetPC           = "cannot get ProviderConfig"
	ErrGetCreds        = "cannot get credentials"
	ErrNewClient       = "cannot create new Service"

	ErrAssertCodeFreshService  = "cannot assert service as CodeFreshAPI"
	ErrExpectedCodeFreshClient = "expected a CodeFreshAPIClient"
	ErrUpdatingProject         = "error updating project in CodeFresh"
	ErrUpdatingProjectStatus   = "error updating project status with project ID"
	ErrDeletingProject         = "error deleting project in CodeFresh"
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
