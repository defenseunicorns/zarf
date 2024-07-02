// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package schema verifies that Zarf packages follow the Zarf schema
package schema

import (
	"fmt"
	"io/fs"
	"regexp"

	"github.com/defenseunicorns/zarf/src/pkg/layout"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/types"
	"github.com/xeipuuv/gojsonschema"
)

// ZarfSchema is exported so main.go can embed the schema file
var ZarfSchema fs.ReadFileFS

// Validate checks the Zarf package in the current directory against the Zarf schema
func Validate() ([]types.PackageFinding, error) {

	var untypedZarfPackage interface{}
	if err := utils.ReadYaml(layout.ZarfYAML, &untypedZarfPackage); err != nil {
		return nil, err
	}

	jsonSchema, err := ZarfSchema.ReadFile("zarf.schema.json")
	if err != nil {
		return nil, err
	}

	return validateSchema(jsonSchema, untypedZarfPackage)
}

func makeFieldPathYqCompat(field string) string {
	if field == "(root)" {
		return field
	}
	// \b is a metacharacter that will stop at the next non-word character (including .)
	// https://regex101.com/r/pIRPk0/1
	re := regexp.MustCompile(`(\b\d+\b)`)

	wrappedField := re.ReplaceAllString(field, "[$1]")

	return fmt.Sprintf(".%s", wrappedField)
}

func validateSchema(jsonSchema []byte, untypedZarfPackage interface{}) ([]types.PackageFinding, error) {
	var findings []types.PackageFinding

	schemaErrors, err := runSchema(jsonSchema, untypedZarfPackage)
	if err != nil {
		return nil, err
	}

	if len(schemaErrors) != 0 {
		for _, schemaErr := range schemaErrors {
			findings = append(findings, types.PackageFinding{
				YqPath:      makeFieldPathYqCompat(schemaErr.Field()),
				Description: schemaErr.Description(),
				Severity:    types.SevErr,
			})
		}
	}

	return findings, err
}

func runSchema(jsonSchema []byte, pkg interface{}) ([]gojsonschema.ResultError, error) {
	schemaLoader := gojsonschema.NewBytesLoader(jsonSchema)
	documentLoader := gojsonschema.NewGoLoader(pkg)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, err
	}

	if !result.Valid() {
		return result.Errors(), nil
	}
	return nil, nil
}
