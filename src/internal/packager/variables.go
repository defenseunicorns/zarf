package packager

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/types"
)

// fillActiveTemplate handles setting the active variables and reloading the base template.
func (p *Packager) fillActiveTemplate() error {
	packageVariables, err := utils.FindYamlTemplates(&p.cfg.Pkg, "###ZARF_PKG_VAR_", "###")
	if err != nil {
		return err
	}

	for key := range p.cfg.CreateOpts.SetVariables {
		value := p.cfg.CreateOpts.SetVariables[key]
		// Ensure uppercase for VIPER
		packageVariables[strings.ToUpper(key)] = &value
	}

	for key, value := range packageVariables {
		if value == nil && !config.CommonOptions.Confirm {
			setVal, err := p.promptVariable(types.ZarfPackageVariable{
				Name: key,
			})

			if err == nil {
				packageVariables[key] = &setVal
			} else {
				return err
			}
		} else if value == nil {
			return fmt.Errorf("variable '%s' must be '--set' when using the '--confirm' flag", key)
		}
	}

	templateMap := map[string]string{}
	for key, value := range packageVariables {
		// Variable keys are always uppercase in the format ###ZARF_PKG_VAR_KEY###
		templateMap[strings.ToUpper(fmt.Sprintf("###ZARF_PKG_VAR_%s###", key))] = *value
	}

	return utils.ReloadYamlTemplate(&p.cfg.Pkg, templateMap)
}

// setActiveVariables handles setting the active variables used to template component files.
func (p *Packager) setActiveVariables() error {
	for key := range p.cfg.DeployOpts.SetVariables {
		value := p.cfg.DeployOpts.SetVariables[key]
		// Ensure uppercase for VIPER
		p.cfg.SetVariableMap[strings.ToUpper(key)] = value
	}

	for _, variable := range p.cfg.Pkg.Variables {
		_, present := p.cfg.SetVariableMap[variable.Name]

		// Variable is present, no need to continue checking
		if present {
			continue
		}

		// First set default (may be overridden by prompt)
		p.cfg.SetVariableMap[variable.Name] = variable.Default

		// Variable is set to prompt the user
		if variable.Prompt && !config.CommonOptions.Confirm {
			// Prompt the user for the variable
			val, err := p.promptVariable(variable)

			if err != nil {
				return err
			}

			p.cfg.SetVariableMap[variable.Name] = val
		}
	}

	return nil
}

// InjectImportedVariable determines if an imported package variable exists in the active config and adds it if not.
func (p *Packager) InjectImportedVariable(importedVariable types.ZarfPackageVariable) {
	presentInActive := false
	for _, configVariable := range p.cfg.Pkg.Variables {
		if configVariable.Name == importedVariable.Name {
			presentInActive = true
		}
	}

	if !presentInActive {
		p.cfg.Pkg.Variables = append(p.cfg.Pkg.Variables, importedVariable)
	}
}

// InjectImportedConstant determines if an imported package constant exists in the active config and adds it if not.
func (p *Packager) InjectImportedConstant(importedConstant types.ZarfPackageConstant) {
	presentInActive := false
	for _, configVariable := range p.cfg.Pkg.Constants {
		if configVariable.Name == importedConstant.Name {
			presentInActive = true
		}
	}

	if !presentInActive {
		p.cfg.Pkg.Constants = append(p.cfg.Pkg.Constants, importedConstant)
	}
}

func (p *Packager) promptVariable(variable types.ZarfPackageVariable) (value string, err error) {

	if variable.Description != "" {
		message.Question(variable.Description)
	}

	prompt := &survey.Input{
		Message: fmt.Sprintf("Please provide a value for \"%s\"", variable.Name),
		Default: variable.Default,
	}

	if err = survey.AskOne(prompt, &value); err != nil {
		return "", err
	}

	return value, nil
}
