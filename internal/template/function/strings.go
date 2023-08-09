package function

import (
	"errors"
	"regexp"
	"strings"

	"github.com/spf13/cast"
)

func Split(str interface{}, sep interface{}) []string {
	return strings.Split(cast.ToString(str), cast.ToString(sep))
}

func Replace(str interface{}, old interface{}, new interface{}, count interface{}) string {
	return strings.Replace(cast.ToString(str), cast.ToString(old), cast.ToString(new), cast.ToInt(count))
}

func ReReplace(str interface{}, old interface{}, new interface{}) (string, error) {
	re, err := regexp.Compile(cast.ToString(old))
	if err != nil {
		return "", errors.New("Invalid regular expression: " + cast.ToString(old))
	}
	return re.ReplaceAllLiteralString(cast.ToString(str), cast.ToString(new)), nil
}

func LCase(str interface{}) string {
	return strings.ToLower(cast.ToString(str))
}

func UCase(str interface{}) string {
	return strings.ToUpper(cast.ToString(str))
}

func TCase(str interface{}) string {
	return strings.Title(cast.ToString(str))
}

func HasPrefix(str interface{}, prefix interface{}) bool {
	return strings.HasPrefix(cast.ToString(str), cast.ToString(prefix))
}

func HasSuffix(str interface{}, suffix interface{}) bool {
	return strings.HasSuffix(cast.ToString(str), cast.ToString(suffix))
}

func Join(sep interface{}, strs ...interface{}) string {
	return strings.Join(cast.ToStringSlice(strs), cast.ToString(sep))
}

func Trim(str interface{}) string {
	return strings.TrimSpace(cast.ToString(str))
}
