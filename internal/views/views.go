package views

import (
	"errors"
	"fmt"
	html_template "html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	text_template "text/template"

	"github.com/glaciers-in-archives/snowman/internal/template/function"
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

	fmt.Println("Rendered page at " + path)
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

func DiscoverViews(layouts []string) ([]View, error) {
	var views []View

	data, err := ioutil.ReadFile("views.yaml")
	if err != nil {
		return nil, errors.New("Failed to read views.yaml")
	}

	var vConfigs viewConfigList
	if err := vConfigs.Parse(data); err != nil {
		return nil, errors.New("Failed to parse views.yaml")
	}

	fmt.Println("Found " + strconv.Itoa(len(vConfigs.Views)) + " views.")

	for _, viewConf := range vConfigs.Views {
		var multipageVariableHook *string
		re := regexp.MustCompile(`{{([\w\d_]+)}}`)
		if re.Match([]byte(viewConf.Output)) {
			multipageVariableHook = &re.FindAllStringSubmatch(viewConf.Output, 1)[0][1]
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
			TextTemplateA, err = text_template.New("").Funcs(getViewFuncs(viewConf)).Funcs(function.GetFileFuncs()).Funcs(function.GetQueryFuncs()).Funcs(function.GetStringFuncs()).Funcs(function.GetMathFuncs()).Funcs(function.GetUtilsFuncs()).Funcs(function.GetIncludeFuncs()).Funcs(function.GetJSONFuncs()).Funcs(function.GetRemoteFuncs()).ParseFiles(templates...)
		} else {
			HTMLTemplateA, err = html_template.New("").Funcs(getViewFuncs(viewConf)).Funcs(function.GetFileFuncs()).Funcs(function.GetQueryFuncs()).Funcs(function.GetStringFuncs()).Funcs(function.GetMathFuncs()).Funcs(function.GetUtilsFuncs()).Funcs(function.GetIncludeFuncs()).Funcs(function.GetJSONFuncs()).Funcs(function.GetRemoteFuncs()).ParseFiles(templates...)
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
