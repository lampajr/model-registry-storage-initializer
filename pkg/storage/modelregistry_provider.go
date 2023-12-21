package storage

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	kserve "github.com/kserve/kserve/pkg/agent/storage"
	"github.com/opendatahub-io/model-registry/pkg/api"
	"github.com/opendatahub-io/model-registry/pkg/core"
	"github.com/opendatahub-io/model-registry/pkg/openapi"
	"google.golang.org/grpc"
)

const MR kserve.Protocol = "model-registry://"

type ModelRegistryProvider struct {
	Client    api.ModelRegistryApi
	Providers map[kserve.Protocol]kserve.Provider
}

func NewModelRegistryProvider(conn *grpc.ClientConn) (*ModelRegistryProvider, error) {
	client, err := core.NewModelRegistryService(conn)
	if err != nil {
		return nil, err
	}

	return &ModelRegistryProvider{
		Client:    client,
		Providers: map[kserve.Protocol]kserve.Provider{},
	}, nil
}

var _ kserve.Provider = (*ModelRegistryProvider)(nil)

// storageUri formatted like model-registry://{registeredModelName}/{versionName}
func (p *ModelRegistryProvider) DownloadModel(modelDir string, modelName string, storageUri string) error {
	log.Printf("Download model indexed in model registry: modelName=%s, storageUri=%s, modelDir=%s", modelName, storageUri, modelDir)

	// Parse the URI to retrieve the needed information to query model registry (modelArtifact)
	mrUri := strings.TrimPrefix(storageUri, string(MR))
	tokens := strings.SplitN(mrUri, "/", 2)

	if len(tokens) == 0 || len(tokens) > 2 {
		return fmt.Errorf("invalid model registry URI, use like model-registry://{registeredModelName}/{versionName}")
	}

	registeredModelName := tokens[0]
	var versionName *string
	if len(tokens) == 2 {
		versionName = &tokens[1]
	}

	// Fetch the registered model
	model, err := p.Client.GetRegisteredModelByParams(&registeredModelName, nil)
	if err != nil {
		return err
	}

	// Fetch model version by name or latest if not specified
	var version *openapi.ModelVersion
	if versionName != nil {
		version, err = p.Client.GetModelVersionByParams(versionName, model.Id, nil)
		if err != nil {
			return err
		}
	} else {
		versions, err := p.Client.GetModelVersions(api.ListOptions{
			OrderBy:   (*string)(openapi.ORDERBYFIELD_CREATE_TIME.Ptr()),
			SortOrder: (*string)(openapi.SORTORDER_DESC.Ptr()),
		}, model.Id)
		if err != nil {
			return err
		}

		if versions.Size == 0 {
			return fmt.Errorf("no versions associated to registered model %s", registeredModelName)
		}
		version = &versions.Items[0]
	}

	artifacts, err := p.Client.GetModelArtifacts(api.ListOptions{
		OrderBy:   (*string)(openapi.ORDERBYFIELD_CREATE_TIME.Ptr()),
		SortOrder: (*string)(openapi.SORTORDER_DESC.Ptr()),
	}, version.Id)
	if err != nil {
		return err
	}

	if artifacts.Size == 0 {
		return fmt.Errorf("no versions associated to registered model %s", registeredModelName)
	}
	modelArtifact := &artifacts.Items[0]

	// Call appropriate kserve provider based on the indexed model artifact URI
	if modelArtifact.Uri == nil {
		return fmt.Errorf("model artifact %s has empty URI", *modelArtifact.Id)
	}

	protocol, err := extractProtocol(*modelArtifact.Uri)
	if err != nil {
		return err
	}

	provider, err := kserve.GetProvider(p.Providers, protocol)
	if err != nil {
		return err
	}

	modelName = registeredModelName
	if version.Name != nil {
		modelName = fmt.Sprintf("%s-%s", modelName, *version.Name)
	}

	return provider.DownloadModel(modelDir, modelName, *modelArtifact.Uri)
}

func extractProtocol(storageURI string) (kserve.Protocol, error) {
	if storageURI == "" {
		return "", fmt.Errorf("there is no storageUri supplied")
	}

	if !regexp.MustCompile("\\w+?://").MatchString(storageURI) {
		return "", fmt.Errorf("there is no protocol specified for the storageUri")
	}

	for _, prefix := range kserve.SupportedProtocols {
		if strings.HasPrefix(storageURI, string(prefix)) {
			return prefix, nil
		}
	}
	return "", fmt.Errorf("protocol not supported for storageUri")
}
