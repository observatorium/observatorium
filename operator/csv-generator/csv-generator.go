package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"github.com/operator-framework/api/pkg/lib/version"
	csvv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

var (
	csvVersion          = flag.String("csv-version", "", "The unified CSV version")
	replacesCsvVersion  = flag.String("replaces-csv-version", "", "The unified CSV version this new CSV will replace")
	skipRange           = flag.String("skip-range", "", "The CSV version skip range")
	operatorCSVTemplate = flag.String("operator-csv-template-file", "", "Path to csv template example")

	operatorImage = flag.String("operator-image", "", "Operator container image")

	inputManifestsDir = flag.String("manifests-directory", "", "The directory containing the extra manifests to be included in the registry bundle")

	outputDir = flag.String("olm-bundle-directory", "", "The directory to output the unified CSV and CRDs to")

	annotationsFile = flag.String("annotations-from", "", "Add metadata annotations from the given file")
	maintainersFile = flag.String("maintainers-from", "", "Add maintainers list from the given file")
	descriptionFile = flag.String("description-from", "", "Replace the description with the content of the given file")

	semverVersion *semver.Version
)

func finalizedCSVFilename() string {
	return "observatorium-operator.v" + *csvVersion + ".clusterserviceversion.yaml"
}

type csvUserData struct {
	Description      string
	ExtraAnnotations map[string]string
	Maintainers      map[string]string
}

func generateUnifiedCSV(userData csvUserData) error {
	operatorCSV, err := UnmarshalCSV(*operatorCSVTemplate)
	if err != nil {
		return err
	}
	strategySpec := operatorCSV.Spec.InstallStrategy.StrategySpec

	// this forces us to update this logic if another deployment is introduced.
	if len(strategySpec.DeploymentSpecs) != 1 {
		log.Fatal("expected 1 deployment %v", operatorCSV)
		log.Fatal("expected 1 deployment, found: ", len(strategySpec.DeploymentSpecs))
	}

	strategySpec.DeploymentSpecs[0].Spec.Template.Spec.Containers[0].Image = *operatorImage
	strategySpec.DeploymentSpecs[0].Spec.Template.Spec.ServiceAccountName = "observatorium-operator"

	// Inject display names and descriptions for our crds
	for i, definition := range operatorCSV.Spec.CustomResourceDefinitions.Owned {
		switch definition.Name {
		case "observatoria.core.observatorium.io":
			operatorCSV.Spec.CustomResourceDefinitions.Owned[i].Description = "Observatorium"
			operatorCSV.Spec.CustomResourceDefinitions.Owned[i].DisplayName = "Observatorium"
		}
	}

	// Re-serialize deployments and permissions into csv strategy.
	operatorCSV.Spec.InstallStrategy.StrategySpec = strategySpec

	operatorCSV.Annotations["containerImage"] = *operatorImage
	for key, value := range userData.ExtraAnnotations {
		operatorCSV.Annotations[key] = value
	}

	// Set correct csv versions and name
	v := version.OperatorVersion{Version: *semverVersion}
	operatorCSV.Spec.Version = v
	operatorCSV.Name = "observatorium-operator.v" + *csvVersion
	if *replacesCsvVersion != "" {
		operatorCSV.Spec.Replaces = "observatorium-operator.v" + *replacesCsvVersion
	} else {
		operatorCSV.Spec.Replaces = ""
	}

	// Set api maturity
	operatorCSV.Spec.Maturity = "alpha"

	// Set links
	operatorCSV.Spec.Links = []csvv1.AppLink{
		{
			Name: "Source Code",
			URL:  "https://github.com/observatorium/deployments",
		},
	}

	// Set Keywords
	operatorCSV.Spec.Keywords = []string{
		"observatorium",
		"prometheus",
		"thanos",
	}

	// Set Provider
	operatorCSV.Spec.Provider = csvv1.AppLink{
		Name: "Red Hat",
	}

	// Set Description
	operatorCSV.Spec.Description = userData.Description
	operatorCSV.Spec.DisplayName = "Observatorium Operator"

	if userData.Maintainers != nil {
		operatorCSV.Spec.Maintainers = []csvv1.Maintainer{}
		for name, email := range userData.Maintainers {
			operatorCSV.Spec.Maintainers = append(operatorCSV.Spec.Maintainers, csvv1.Maintainer{
				Name:  name,
				Email: email,
			})
		}
	}

	// Set Annotations
	if *skipRange != "" {
		operatorCSV.Annotations["olm.skipRange"] = *skipRange
	}

	// write CSV to out dir
	writer := strings.Builder{}
	MarshallObject(operatorCSV, &writer)
	outputFilename := filepath.Join(*outputDir, finalizedCSVFilename())
	if err := ioutil.WriteFile(outputFilename, []byte(writer.String()), 0644); err != nil {
		return err
	}

	fmt.Printf("CSV written to %s\n", outputFilename)

	return nil
}

func readFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func readKeyValueMapFromFile(filename string) (map[string]string, error) {
	kvMap := make(map[string]string)
	data, err := readFile(filename)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &kvMap); err != nil {
		return nil, err
	}
	return kvMap, nil
}

func main() {
	flag.Parse()

	if *csvVersion == "" {
		log.Fatal("--csv-version is required")
	} else if *operatorCSVTemplate == "" {
		log.Fatal("--operator-csv-template-file is required")
	} else if *operatorImage == "" {
		log.Fatal("--operator-image is required")
	} else if *outputDir == "" {
		log.Fatal("--olm-bundle-directory is required")
	}

	var err error
	// Set correct csv versions and name
	semverVersion, err = semver.New(*csvVersion)
	if err != nil {
		log.Fatal("Couldn't parse CSV version:\n%v", err)
	}

	userData := csvUserData{
		Description: `
Observatorium Operator provides the ability to install the components that comprise the Observatorium project.`,
		ExtraAnnotations: make(map[string]string),
		Maintainers:      make(map[string]string),
	}

	if *annotationsFile != "" {
		value, err := readKeyValueMapFromFile(*annotationsFile)
		if err != nil {
			log.Fatal("Couldn't read annotations file:\n%v", err)
		}
		userData.ExtraAnnotations = value
	}
	if *maintainersFile != "" {
		value, err := readKeyValueMapFromFile(*maintainersFile)
		if err != nil {
			log.Fatal("Couldn't read maintainers file:\n%v", err)
		}
		userData.Maintainers = value
	}
	if *descriptionFile != "" {
		data, err := readFile(*descriptionFile)
		if err != nil {
			log.Fatal("Couldn't read description file:\n%v", err)
		}
		userData.Description = string(data)
	}

	// start with a fresh output directory if it already exists
	if err := os.RemoveAll(*outputDir); err != nil {
		log.Fatal("Couldn't remove existing output directory:\n%v", err)
	}

	// create output directory
	if err := os.MkdirAll(*outputDir, os.FileMode(0775)); err != nil {
		log.Fatal("Couldn't create output directory:\n%v", err)
	}

	if err := generateUnifiedCSV(userData); err != nil {
		log.Fatal("Couldn't generate CSV:\n%v", err)
	}
}
