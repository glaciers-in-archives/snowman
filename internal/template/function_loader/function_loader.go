package function_loader

import (
	"html/template"
	"os"
	"time"

	"github.com/glaciers-in-archives/snowman/internal/template/function"
)

// FunctionLoader is a function that returns a `template.FuncMap`
// with all default functions minus `include`, `include_text` and `current_view`.
// `include_text` and `include` are not here because the depend on all the other
// functions and therefore would cause a circular dependency.
// `current_view` is not here because it is only available in the context of a view.
func FunctionLoader() template.FuncMap {
	var functions = map[string]interface{}{
		"to_json":   function.ToJSON,
		"from_json": function.FromJSON,

		"read_file": function.ReadFile,

		"add1": function.Add1,
		"add":  function.Add,
		"sub":  function.Sub,
		"div":  function.Div,
		"mod":  function.Mod,
		"mul":  function.Mul,
		"rand": function.Rand,

		"query": function.Query,

		"get_remote":             function.GetRemote,
		"get_remote_with_config": function.GetRemoteWithConfig,

		"split":      function.Split,
		"replace":    function.Replace,
		"re_replace": function.ReReplace,
		"lcase":      function.LCase,
		"ucase":      function.UCase,
		"tcase":      function.TCase,
		"join":       function.Join,
		"has_prefix": function.HasPrefix,
		"has_suffix": function.HasSuffix,
		"trim":       function.Trim,
		"contains":   function.Contains,

		"md5":    function.MD5,
		"sha1":   function.SHA1,
		"sha256": function.SHA256,

		"dict_set":     function.Set,
		"dict_get":     function.Get,
		"dict_create":  function.Dict,
		"dict_unset":   function.Unset,
		"dict_has_key": function.HasKey,
		"dict_pluck":   function.Pluck,
		"dict_find":    function.Dig,
		"dict_keys":    function.Keys,
		"dict_pick":    function.Pick,
		"dict_omit":    function.Omit,
		"dict_values":  function.Values,

		"safe_html": function.SafeHTML,
		"uri":       function.URI,
		"config":    function.Config,
		"version":   function.Version,
		"type":      function.Type,
		"now":       time.Now,
		"env":       os.Getenv,
	}

	return template.FuncMap(functions)
}
