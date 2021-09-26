package function

import (
	"bytes"
	"errors"
	html_template "html/template"
	"os"
	"path/filepath"
	text_template "text/template"
)

func include(templatePath string, arguments ...interface{}) (html_template.HTML, error) {
	templatePath = "templates/" + templatePath
	if _, err := os.Stat(templatePath); err != nil {
		return "", errors.New("Unable to find the template file " + templatePath)
	}

	tpl, err := html_template.New("").Funcs((GetHTMLIncludeFuncs())).Funcs(GetHTMLQueryFuncs()).Funcs(GetHTMLStringFuncs()).Funcs(GetHTMLMathFuncs()).Funcs(GetHTMLUtilsFuncs()).ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var renderedTpl bytes.Buffer
	switch len(arguments) {
	case 0:
		if err := tpl.ExecuteTemplate(&renderedTpl, filepath.Base(templatePath), nil); err != nil {
			return "", err
		}
	case 1:
		if err := tpl.ExecuteTemplate(&renderedTpl, filepath.Base(templatePath), arguments[0]); err != nil {
			return "", err
		}
	}

	return html_template.HTML(renderedTpl.String()), nil
}

func include_text(templatePath string, arguments ...interface{}) (string, error) {
	templatePath = "templates/" + templatePath
	if _, err := os.Stat(templatePath); err != nil {
		return "", errors.New("Unable to find the template file " + templatePath)
	}

	tpl, err := text_template.New("").Funcs((GetTextIncludeFuncs())).Funcs(GetTextQueryFuncs()).Funcs(GetTextStringFuncs()).Funcs(GetTextMathFuncs()).Funcs(GetTextUtilsFuncs()).ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var renderedTpl bytes.Buffer
	switch len(arguments) {
	case 0:
		if err := tpl.ExecuteTemplate(&renderedTpl, filepath.Base(templatePath), nil); err != nil {
			return "", err
		}
	case 1:
		if err := tpl.ExecuteTemplate(&renderedTpl, filepath.Base(templatePath), arguments[0]); err != nil {
			return "", err
		}
	}

	return renderedTpl.String(), nil
}

var includeFuncs = map[string]interface{}{}

func init() {
	includeFuncs["include"] = include
	includeFuncs["include_text"] = include_text
}

func GetHTMLIncludeFuncs() html_template.FuncMap {
	return html_template.FuncMap(includeFuncs)
}

func GetTextIncludeFuncs() text_template.FuncMap {
	return text_template.FuncMap(includeFuncs)
}
