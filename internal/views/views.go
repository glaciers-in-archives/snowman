package views

import (
	"errors"
	"fmt"
	html_template "html/template"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	text_template "text/template"

	function "github.com/glaciers-in-archives/snowman/internal/template/child_template_function"
	"github.com/glaciers-in-archives/snowman/internal/template/function_loader"
	"gopkg.in/yaml.v2"
)

type viewConfigList struct {
	Views []viewConfig `yaml:"views"`
}

func (c *viewConfigList) Parse(data []byte) error {
	return yaml.Unmarshal(data, &c)
}

type viewConfig struct {
	Output       string `yaml:"output"`
	QueryFile    string `yaml:"query"`
	TemplateFile string `yaml:"template"`
	Unsafe       bool   `yaml:"unsafe"`
}

type View struct {
	ViewConfig            viewConfig
	TextTemplate          *text_template.Template
	HTMLTemplate          *html_template.Template
	TemplateName          string
	MultipageVariableHook *string
}

func (v *View) RenderPage(path string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if v.ViewConfig.Unsafe {
		if err := v.TextTemplate.ExecuteTemplate(f, v.TemplateName, data); err != nil {
			return err
		}
	} else {
		if err := v.HTMLTemplate.ExecuteTemplate(f, v.TemplateName, data); err != nil {
			return err
		}
	}

	f.Close()

	return nil
}

func getViewFuncs(currentViewConfig viewConfig) html_template.FuncMap {
	var viewFuncs = map[string]interface{}{
		"current_view": func() viewConfig {
			return currentViewConfig
		},
	}
	return html_template.FuncMap(viewFuncs)
}

var multipageVariablePattern = regexp.MustCompile(`{{ *([\w\d_]+) *}}`)

// normalizeOutput removes the optional whitespace around a multipage variable
// so that "entities/{{ id }}.html" and "entities/{{id}}.html" compare as equal.
func normalizeOutput(output string) string {
	return multipageVariablePattern.ReplaceAllString(output, "{{$1}}")
}

// selectViewConfigs returns the views whose output matches the given selector.
// The selector is the output of a view as it's written in views.yaml.
func selectViewConfigs(viewConfigs []viewConfig, selector string) ([]viewConfig, error) {
	var selected []viewConfig
	for _, viewConf := range viewConfigs {
		if normalizeOutput(viewConf.Output) == normalizeOutput(selector) {
			selected = append(selected, viewConf)
		}
	}

	if len(selected) == 0 {
		message := "Unable to find a view with the output " + selector + ". The available views are:"
		for _, viewConf := range viewConfigs {
			message += "\n  " + viewConf.Output
		}
		return nil, errors.New(message)
	}

	return selected, nil
}

// DiscoverViews reads views.yaml and prepares the views found in it. If
// viewSelector isn't empty, only the views with that output are prepared.
func DiscoverViews(layouts []string, viewSelector string) ([]View, error) {
	var views []View

	data, err := os.ReadFile("views.yaml")
	if err != nil {
		return nil, errors.New("Failed to read views.yaml")
	}

	var vConfigs viewConfigList
	if err := vConfigs.Parse(data); err != nil {
		return nil, errors.New("Failed to parse views.yaml")
	}

	viewConfigs := vConfigs.Views
	if viewSelector != "" {
		viewConfigs, err = selectViewConfigs(vConfigs.Views, viewSelector)
		if err != nil {
			return nil, err
		}
		fmt.Println("Building " + strconv.Itoa(len(viewConfigs)) + " of the project's " + strconv.Itoa(len(vConfigs.Views)) + " views.")
	} else {
		fmt.Println("Building project with " + strconv.Itoa(len(viewConfigs)) + " views.")
	}

	for _, viewConf := range viewConfigs {
		var multipageVariableHook *string
		if match := multipageVariablePattern.FindStringSubmatch(viewConf.Output); match != nil {
			multipageVariableHook = &match[1]
		}

		templatePath := "templates/" + viewConf.TemplateFile
		if _, err := os.Stat(templatePath); err != nil {
			return nil, errors.New("Unable to find the template file " + viewConf.TemplateFile)
		}

		_, file := filepath.Split(templatePath)

		templates := append(layouts, templatePath)

		var TextTemplateA *text_template.Template
		var HTMLTemplateA *html_template.Template
		if viewConf.Unsafe {
			TextTemplateA, err = text_template.New("").Funcs(getViewFuncs(viewConf)).Funcs(function_loader.FunctionLoader()).Funcs(function.GetIncludeFuncs()).ParseFiles(templates...)
		} else {
			HTMLTemplateA, err = html_template.New("").Funcs(getViewFuncs(viewConf)).Funcs(function_loader.FunctionLoader()).Funcs(function.GetIncludeFuncs()).ParseFiles(templates...)
		}

		if err != nil {
			return nil, err
		}

		view := View{
			ViewConfig:            viewConf,
			HTMLTemplate:          HTMLTemplateA,
			TextTemplate:          TextTemplateA,
			TemplateName:          file,
			MultipageVariableHook: multipageVariableHook,
		}
		views = append(views, view)
	}
	return views, nil
}
