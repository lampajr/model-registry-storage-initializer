package main

import (
	"log"
	"os"

	"github.com/kubeflow/model-registry/pkg/openapi"
	"github.com/lampajr/model-registry-storage-initializer/pkg/storage"
)

const (
	modelRegistryBaseUrlEnv     = "MODEL_REGISTRY_BASE_URL"
	modelRegistrySchemeEnv      = "MODEL_REGISTRY_SCHEME"
	modelRegistryBaseUrlDefault = "localhost:8080"
	modelRegistrySchemeDefault  = "http"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: ./mr-storage-initializer <src-uri> <dest-path>")
	}

	sourceUri := os.Args[1]
	destPath := os.Args[2]

	log.Printf("Initializing, args: src_uri [%s] dest_path[ [%s]\n", sourceUri, destPath)

	baseUrl, ok := os.LookupEnv(modelRegistryBaseUrlEnv)
	if !ok || baseUrl == "" {
		baseUrl = modelRegistryBaseUrlDefault
	}

	scheme, ok := os.LookupEnv(modelRegistrySchemeEnv)
	if !ok || scheme == "" {
		scheme = modelRegistrySchemeDefault
	}

	cfg := openapi.NewConfiguration()
	cfg.Host = baseUrl
	cfg.Scheme = scheme
	provider, err := storage.NewModelRegistryProvider(cfg)
	if err != nil {
		log.Fatalf("Error initiliazing model registry provider: %v", err)
	}

	if err := provider.DownloadModel(destPath, "", sourceUri); err != nil {
		log.Fatalf(err.Error())
	}
}
