package views

import (
	"errors"
	"fmt"
	html_template "html/template"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	text_template "text/template"
	"time"

	"github.com/glaciers-in-archives/snowman/internal/config"
	"github.com/glaciers-in-archives/snowman/internal/sparql"
	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/spf13/cast"
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

func DiscoverViews(templates []string, repo sparql.Repository, siteConfig config.SiteConfig) ([]View, error) {
	var views []View

	var functionMap = map[string]interface{}{
		"now":     time.Now,
		"split":   strings.Split,
		"replace": strings.Replace,
		"lcase":   strings.ToLower,
		"ucase":   strings.ToUpper,
		"tcase":   strings.Title,
		"env":     os.Getenv,
		"join":    utils.Join,

		"safe_html": utils.StringToHTML,
		"query":     repo.InlineQuery,
		"config":    siteConfig.Get,
		"uri":       utils.ToURI,

		"add1": func(i interface{}) int64 { return cast.ToInt64(i) + 1 },
		"add": func(i ...interface{}) int64 {
			var a int64 = 0
			for _, b := range i {
				a += cast.ToInt64(b)
			}
			return a
		},
		"sub": func(a, b interface{}) int64 { return cast.ToInt64(a) - cast.ToInt64(b) },
		"div": func(a, b interface{}) int64 { return cast.ToInt64(a) / cast.ToInt64(b) },
		"mod": func(a, b interface{}) int64 { return cast.ToInt64(a) % cast.ToInt64(b) },
		"mul": func(a interface{}, v ...interface{}) int64 {
			val := cast.ToInt64(a)
			for _, b := range v {
				val = val * cast.ToInt64(b)
			}
			return val
		},
		"rand": func(min, max int) int { return rand.Intn(max-min) + min },
	}

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

		// ParseFiles requries the base template as the last item therfore we add it again
		templates = append(templates, templatePath)

		var TextTemplateA *text_template.Template
		var HTMLTemplateA *html_template.Template
		if viewConf.Unsafe {
			funcMap := text_template.FuncMap(functionMap)
			TextTemplateA, err = text_template.New("").Funcs(funcMap).ParseFiles(templates...)
		} else {
			funcMap := html_template.FuncMap(functionMap)
			HTMLTemplateA, err = html_template.New("").Funcs(funcMap).ParseFiles(templates...)
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
