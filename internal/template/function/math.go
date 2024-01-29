package function

import (
	"math/rand"

	"github.com/spf13/cast"
)

func Add1(i interface{}) int64 {
	return cast.ToInt64(i) + 1
}

func Add(a, b interface{}) int64 {
	return cast.ToInt64(a) + cast.ToInt64(b)
}

func Sub(a, b interface{}) int64 {
	return cast.ToInt64(a) - cast.ToInt64(b)
}

func Div(a, b interface{}) int64 {
	return cast.ToInt64(a) / cast.ToInt64(b)
}

func Mul(a, b interface{}) int64 {
	return cast.ToInt64(a) * cast.ToInt64(b)
}

func Mod(a, b interface{}) int64 {
	return cast.ToInt64(a) % cast.ToInt64(b)
}

func Rand(min, max int) int {
	return rand.Intn(max-min) + min
}
