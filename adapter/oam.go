package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	meshmodel "github.com/layer5io/meshkit/models/meshmodel/core/v1alpha1"
	"github.com/layer5io/meshkit/models/oam/core/v1alpha1"
	"github.com/layer5io/meshkit/utils"
	"github.com/layer5io/meshkit/utils/manifests"
)

var (
	Manifests  = "MANIFESTS"
	HelmCharts = "HELM_CHARTS"
)

type MeshModelConfig struct {
	Category         string
	CategoryMetadata map[string]interface{}
	Metadata         map[string]interface{}
}

// StaticCompConfig is used to configure CreateComponents
type StaticCompConfig struct {
	MeshModelName   string //Used in Adding ModelName onto Core Meshmodel components. Pass it the same as meshName in OAM components
	URL             string //URL
	Method          string //Use the constants exported by package. Manifests or Helm
	MeshModelPath   string
	MeshModelConfig MeshModelConfig
	DirName         string           //The directory's name. By convention, it should be the version name
	Config          manifests.Config //Filters required to create definition and schema
	Force           bool             //When set to true, if the file with same name already exists, they will be overridden
}

// CreateComponents generates components for a given configuration and stores them.
func CreateComponents(scfg StaticCompConfig) error {
	meshmodeldirName, _ := getLatestDirectory(scfg.MeshModelPath)
	meshmodelDir := filepath.Join(scfg.MeshModelPath, scfg.DirName)
	_, err := os.Stat(meshmodelDir)
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(meshmodelDir, 0777)
		if err != nil {
			return ErrCreatingComponents(err)
		}
	}
	var comp *manifests.Component
	switch scfg.Method {
	case Manifests:
		comp, err = manifests.GetFromManifest(context.Background(), scfg.URL, manifests.SERVICE_MESH, scfg.Config)
	case HelmCharts:
		comp, err = manifests.GetFromHelm(context.Background(), scfg.URL, manifests.SERVICE_MESH, scfg.Config)
	default:
		return ErrCreatingComponents(errors.New("invalid generation method. Must be either Manifests or HelmCharts"))
	}
	if err != nil {
		return ErrCreatingComponents(err)
	}
	if comp == nil {
		return ErrCreatingComponents(errors.New("no components found"))
	}

	for i, def := range comp.Definitions {
		schema := comp.Schemas[i]
		name := getNameFromWorkloadDefinition([]byte(def))
		meshmodelFileName := name + "_meshmodel.json"
		err = createMeshModelComponentsFromLegacyOAMComponents([]byte(def), schema, filepath.Join(meshmodelDir, meshmodelFileName), scfg.MeshModelName, scfg.MeshModelConfig)
		if err != nil {
			return ErrCreatingComponents(err)
		}
	}
	//For Meshmodel components
	if meshmodeldirName != "" {
		err = copyCoreComponentsToNewVersion(filepath.Join(scfg.MeshModelPath, meshmodeldirName), filepath.Join(scfg.MeshModelPath, scfg.DirName), scfg.DirName, true)
		if err != nil {
			return ErrCreatingComponents(err)
		}
	}
	return nil
}
func convertOAMtoMeshmodel(def []byte, schema string, isCore bool, meshmodelname string, mcfg MeshModelConfig) ([]byte, error) {
	var oamdef v1alpha1.WorkloadDefinition
	err := json.Unmarshal(def, &oamdef)
	if err != nil {
		return nil, err
	}
	var c meshmodel.ComponentDefinition
	c.Metadata = make(map[string]interface{})
	metaname := strings.Split(manifests.FormatToReadableString(oamdef.ObjectMeta.Name), ".")
	var displayname string
	if len(metaname) > 0 {
		displayname = metaname[0]
	}
	c.DisplayName = displayname
	c.Model.Category = meshmodel.Category{
		Name: mcfg.Category,
	}
	if mcfg.CategoryMetadata != nil {
		c.Model.Category.Metadata = mcfg.CategoryMetadata
	}
	c.Metadata = mcfg.Metadata
	if isCore {
		c.APIVersion = oamdef.APIVersion
		c.Kind = oamdef.ObjectMeta.Name
		c.Model.Version = oamdef.Spec.Metadata["version"]
		c.Model.Name = meshmodelname
	} else {
		c.APIVersion = oamdef.Spec.Metadata["k8sAPIVersion"]
		c.Kind = oamdef.Spec.Metadata["k8sKind"]
		c.Model.Version = oamdef.Spec.Metadata["meshVersion"]
		c.Model.Name = oamdef.Spec.Metadata["meshName"]
	}
	c.Model.DisplayName = manifests.FormatToReadableString(c.Model.Name)
	c.Model.Name = strings.ToLower(c.Model.Name)
	c.Model.Metadata = c.Metadata
	c.Format = meshmodel.JSON
	c.Schema = schema
	byt, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return byt, nil
}

// TODO: After OAM is completely removed from meshkit, replace this with fetching native meshmodel components. For now, reuse OAM functions
func createMeshModelComponentsFromLegacyOAMComponents(def []byte, schema string, path string, meshmodel string, mcfg MeshModelConfig) (err error) {
	byt, err := convertOAMtoMeshmodel(def, schema, false, meshmodel, mcfg)
	if err != nil {
		return err
	}
	err = writeToFile(path, byt, true)
	return
}

// Meshery core components are versioned alongside their corresponding Adapter components,
// which, in turn, are versioned with respect to the infrastructure under management; e.g. "Istio Mesh".
// Every time that managed components are generated for a new infrastructure version (e.g.  service mesh version),
// the latest core components are to be replicated (copied) and assigned the latest infrastructure version.
// The schema of the replicated core components can be augmented or left as-is depending upon the need to do so.
func copyCoreComponentsToNewVersion(fromDir string, toDir string, newVersion string, isMeshmodel bool) error {
	files, err := os.ReadDir(fromDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		//core definition file or core schema file
		if !strings.Contains(strings.TrimSuffix(f.Name(), ".json"), ".") || !strings.Contains(strings.TrimSuffix(f.Name(), ".meshery.layer5io.schema.json"), ".") {
			fsource, err := os.Open(filepath.Join(fromDir, f.Name()))
			if err != nil {
				return err
			}
			defer fsource.Close()
			content, err := io.ReadAll(fsource)
			if err != nil {
				return err
			}
			//only for definition files
			if !strings.Contains(strings.TrimSuffix(f.Name(), ".json"), ".") {
				if isMeshmodel {
					content, err = modifyMeshmodelVersionInDefinition(content, newVersion)
				} else {
					content, err = modifyVersionInDefinition(content, newVersion)
				}
				if err != nil {
					return err
				}
			}
			err = writeToFile(filepath.Join(toDir, f.Name()), content, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func modifyMeshmodelVersionInDefinition(old []byte, newversion string) (new []byte, err error) {
	var def meshmodel.ComponentDefinition
	err = json.Unmarshal(old, &def)
	if err != nil {
		return
	}
	def.Model.Version = newversion
	new, err = json.Marshal(def)
	return
}
func modifyVersionInDefinition(old []byte, newversion string) (new []byte, err error) {
	var def v1alpha1.WorkloadDefinition
	err = json.Unmarshal(old, &def)
	if err != nil {
		return
	}
	if def.Spec.Metadata == nil { //to avoid panic
		def.Spec.Metadata = make(map[string]string)
	}
	def.Spec.Metadata["version"] = newversion
	new, err = json.Marshal(def)
	return
}
func getLatestDirectory(path string) (string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	filenames := []string{}
	for _, f := range files {
		filenames = append(filenames, f.Name())
	}
	filenames = utils.SortDottedStringsByDigits(filenames)
	if len(filenames) != 0 {
		return filenames[len(filenames)-1], nil
	}
	return "", fmt.Errorf("no directory found")
}

// create a file with this filename and stuff the string
func writeToFile(path string, data []byte, force bool) error {
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) { //There some other error than non existence of file
		return err
	}

	if err == nil { //file already exists
		if !force { // Dont override existing file, skip it
			fmt.Println("File already exists,skipping...")
			return nil
		}
		err := os.Remove(path) //Remove the existing file, before overriding it
		if err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0666)
}

// getNameFromWorkloadDefinition takes out name from workload definition
func getNameFromWorkloadDefinition(definition []byte) string {
	var wd v1alpha1.WorkloadDefinition
	err := json.Unmarshal(definition, &wd)
	if err != nil {
		return ""
	}
	return wd.Spec.DefinitionRef.Name
}

// This will be depracated once all adapters migrate to new method of component creation( using static config) and registeration
type DynamicComponentsConfig struct {
	TimeoutInMinutes time.Duration
	URL              string
	GenerationMethod string
	Config           manifests.Config
	Operation        string
}