package adapter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/layer5io/meshkit/models/meshmodel/core/types"
)

var (
	basePath, _         = os.Getwd()
	MeshmodelComponents = filepath.Join(basePath, "templates", "meshmodel", "components")
)

// AvailableVersions denote the component versions available statically
var AvailableVersions = map[string]bool{}

type meshmodelDefinitionPathSet struct {
	meshmodelDefinitionPath string
}

func RegisterMeshModelComponents(uuid, runtime, host, port string) error {
	meshmodelRDP := []MeshModelRegistrantDefinitionPath{}
	pathSets, err := loadMeshmodelComponents(MeshmodelComponents)
	if err != nil {
		return ErrRegisterComponents(err)
	}
	portint, _ := strconv.Atoi(port)
	for _, pathSet := range pathSets {
		meshmodelRDP = append(meshmodelRDP, MeshModelRegistrantDefinitionPath{
			EntityDefintionPath: pathSet.meshmodelDefinitionPath,
			Host:                host,
			Port:                portint,
			Type:                types.ComponentDefinition,
		})
	}

	return NewMeshModelRegistrant(meshmodelRDP, fmt.Sprintf("%s/api/meshmodel/components/register", runtime)).
		Register(uuid)
}

var versionLock sync.Mutex

func loadMeshmodelComponents(basepath string) ([]meshmodelDefinitionPathSet, error) {
	res := []meshmodelDefinitionPathSet{}
	if err := filepath.Walk(basepath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		res = append(res, meshmodelDefinitionPathSet{
			meshmodelDefinitionPath: path,
		})
		versionLock.Lock()
		AvailableVersions[filepath.Base(filepath.Dir(path))] = true // Getting available versions already existing on file system
		versionLock.Unlock()
		return nil
	}); err != nil {
		return nil, err
	}

	return res, nil
}

// RegisterModernMeshModelComponents registers generated component
// definitions using the modern registry API.
func RegisterModernMeshModelComponents(
	ctx context.Context,
	uuid, runtime, host, port string,
) error {
	pathSets, err := loadMeshmodelComponents(MeshmodelComponents)
	if err != nil {
		return ErrRegisterComponents(err)
	}

	paths := make([]string, 0, len(pathSets))
	for _, pathSet := range pathSets {
		paths = append(paths, pathSet.meshmodelDefinitionPath)
	}

	return RegisterModernMeshModelComponentFiles(
		ctx, uuid, runtime, host, port, paths,
	)
}

// RegisterModernMeshModelComponentFiles registers explicitly
// selected component definitions using the modern registry API.
func RegisterModernMeshModelComponentFiles(
	ctx context.Context,
	uuid, runtime, host, port string,
	paths []string,
) error {
	if strings.TrimSpace(runtime) == "" {
		return fmt.Errorf("meshery server address is required")
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid adapter port: %w", err)
	}

	connection, err := newAdapterConnection(host, portNumber, uuid)
	if err != nil {
		return err
	}

	registrant := ModernComponentRegistrant{
		Endpoint: strings.TrimRight(runtime, "/") +
			"/api/registry/components",
		Connection: connection,
	}

	return registrant.RegisterFiles(ctx, paths)
}
