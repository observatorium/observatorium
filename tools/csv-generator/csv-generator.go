package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/blang/semver"

	csvv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/api/pkg/lib/version"
)

var (
	csvVersion          = flag.String("csv-version", "", "the unified CSV version")
	replacesCsvVersion  = flag.String("replaces-csv-version", "", "the unified CSV version this new CSV will replace")
	skipRange           = flag.String("skip-range", "", "the CSV version skip range")
	operatorCSVTemplate = flag.String("operator-csv-template-file", "", "path to csv template example")

	operatorImage = flag.String("operator-image", "", "operator container image")

	inputManifestsDir = flag.String("manifests-directory", "", "The directory containing the extra manifests to be included in the registry bundle")

	outputDir = flag.String("olm-bundle-directory", "", "The directory to output the unified CSV and CRDs to")

	annotationsFile = flag.String("annotations-from", "", "add metadata annotations from given file")
	maintainersFile = flag.String("maintainers-from", "", "add maintainers list from given file")
	descriptionFile = flag.String("description-from", "", "replace the description with the content of the given file")

	semverVersion *semver.Version
)

func finalizedCsvFilename() string {
	return "observatorium-operator.v" + *csvVersion + ".clusterserviceversion.yaml"
}

func copyFile(src string, dst string) {
	srcFile, err := os.Open(src)
	if err != nil {
		panic(err)
	}
	defer srcFile.Close()

	outFile, err := os.Create(dst)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, srcFile)
	if err != nil {
		panic(err)
	}
}

type csvUserData struct {
	Description      string
	ExtraAnnotations map[string]string
	Maintainers      map[string]string
}

func generateUnifiedCSV(userData csvUserData) {

	operatorCSV := UnmarshalCSV(*operatorCSVTemplate)
	strategySpec := operatorCSV.Spec.InstallStrategy.StrategySpec

	// this forces us to update this logic if another deployment is introduced.
	if len(strategySpec.DeploymentSpecs) != 1 {
		panic(fmt.Errorf("expected 1 deployment, found %d", len(strategySpec.DeploymentSpecs)))
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
			URL:  "https://github.com/observatorium/configuration",
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
	operatorCSV.Spec.Description = `
Observatorium Operator provides the ability to install the components that comprise the Observatorium project.`
	if userData.Description != "" {
		operatorCSV.Spec.Description = userData.Description
	}

	operatorCSV.Spec.DisplayName = "Observatorium Operator"

	if userData.Maintainers != nil {
		for name, email := range userData.Maintainers {
			operatorCSV.Spec.Maintainers = append(operatorCSV.Spec.Maintainers, csvv1.Maintainer{
				Name:  name,
				Email: email,
			})
		}
	}
	operatorCSV.Spec.Maintainers = nil
	operatorCSV.Spec.Icon = nil

	// Set Annotations
	if *skipRange != "" {
		operatorCSV.Annotations["olm.skipRange"] = *skipRange
	}

	// write CSV to out dir
	writer := strings.Builder{}
	MarshallObject(operatorCSV, &writer)
	outputFilename := filepath.Join(*outputDir, finalizedCsvFilename())
	err := ioutil.WriteFile(outputFilename, []byte(writer.String()), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("CSV written to %s\n", outputFilename)
}

func readFileOrPanic(filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return data
}

func readKeyValueMapFromFileOrPanic(filename string) map[string]string {
	kvMap := make(map[string]string)
	if err := json.Unmarshal(readFileOrPanic(filename), &kvMap); err != nil {
		panic(err)
	}
	return kvMap
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
		panic(err)
	}

	userData := csvUserData{
		Description: `
Observatorium Operator provides the ability to install the components that comprise the Observatorium project.`,
		ExtraAnnotations: make(map[string]string),
		Maintainers:      make(map[string]string),
	}

	if *annotationsFile != "" {
		userData.ExtraAnnotations = readKeyValueMapFromFileOrPanic(*annotationsFile)
	}
	if *maintainersFile != "" {
		userData.Maintainers = readKeyValueMapFromFileOrPanic(*maintainersFile)
	}
	if *descriptionFile != "" {
		userData.Description = string(readFileOrPanic(*descriptionFile))
	}

	// start with a fresh output directory if it already exists
	os.RemoveAll(*outputDir)

	// create output directory
	os.MkdirAll(*outputDir, os.FileMode(0775))

	generateUnifiedCSV(userData)
}

