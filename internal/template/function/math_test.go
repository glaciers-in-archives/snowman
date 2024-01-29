package function

import (
	"testing"

	"github.com/spf13/cast"
)

type add1Test struct {
	arg  interface{}
	want int64
}

var add1Tests = []add1Test{
	{1, 2},
	{"2", 3},
	{2.0, 3},
	{2.9, 3},
	{2.1, 3},
	{-1, 0},
	{"-2", -1},
}

var add1TestsWithError = []add1Test{
	{"a", 0},
	{"", 0},
	{nil, 0},
	{"", 1},
	{nil, 1},
}

func TestAdd1(t *testing.T) {
	for _, test := range add1Tests {
		if got := Add1(test.arg); got != test.want {
			if Type(got) != "int64" {
				t.Errorf("Output %q not equal to expected %q", Type(got), "int64")
			}

			t.Errorf("Output %q not equal to expected %q", cast.ToString(got), cast.ToString(test.want))
		}
	}
}

type addTest struct {
	arg  []interface{}
	want int64
}

var addTests = []addTest{
	{[]interface{}{1, 2}, 3},
	{[]interface{}{"2", 3}, 5},
	{[]interface{}{2.0, 3}, 5},
	{[]interface{}{2.9, 3}, 5},
	{[]interface{}{2.1, 3}, 5},
	{[]interface{}{-1, 0}, -1},
	{[]interface{}{"-2", -1}, -3},
	{[]interface{}{1, 2, 3, 4, 5}, 15},
	{[]interface{}{1, 2, "3", 4, 5, 6}, 21},
}

var addTestsWithError = []addTest{
	{[]interface{}{1, 2}, 4},
	{[]interface{}{"yuyu", 3}, 6},
	//{[]interface{}{3, 3, ""}, 6},
	//{[]interface{}{3, 3, nil}, 6},
	{[]interface{}{"~", 3}, 6},
}

func TestAdd(t *testing.T) {
	for _, test := range addTests {
		if got := Add(test.arg...); got != test.want {
			if Type(got) != "int64" {
				t.Errorf("Output %q not equal to expected %q", Type(got), "int64")
			}

			t.Errorf("Output %q not equal to expected %q", cast.ToString(got), cast.ToString(test.want))
		}
	}

	for _, test := range addTestsWithError {
		// these tests should fail
		if got := Add(test.arg...); got == test.want {
			t.Errorf("Negative test did not fail and instead returned %q", cast.ToString(got))
		}
	}
}
