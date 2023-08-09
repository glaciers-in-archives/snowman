package function

import (
	"math/rand"

	"github.com/spf13/cast"
)

func Add1(i interface{}) int64 {
	return cast.ToInt64(i) + 1
}

func Add(i ...interface{}) int64 {
	var a int64 = 0
	for _, b := range i {
		a += cast.ToInt64(b)
	}
	return a
}

func Sub(a, b interface{}) int64 {
	return cast.ToInt64(a) - cast.ToInt64(b)
}

func Div(a, b interface{}) int64 {
	return cast.ToInt64(a) / cast.ToInt64(b)
}

func Mod(a, b interface{}) int64 {
	return cast.ToInt64(a) % cast.ToInt64(b)
}

func Mul(a interface{}, v ...interface{}) int64 {
	val := cast.ToInt64(a)
	for _, b := range v {
		val = val * cast.ToInt64(b)
	}
	return val
}

func Rand(min, max int) int {
	return rand.Intn(max-min) + min
}
