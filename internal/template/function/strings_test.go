package function

import (
	"testing"

	"github.com/spf13/cast"
)

type containsTest struct {
	arg1 interface{}
	arg2 interface{}
	want bool
}

var tests = []containsTest{
	{"hello", "ell", true},
	{"hello", "world", false},
}

func TestContains(t *testing.T) {
	for _, test := range tests {
		if got := Contains(test.arg1, test.arg2); got != test.want {
			t.Errorf("Output %q not equal to expected %q", cast.ToString(got), cast.ToString(test.want))
		}
	}
}
