package function

import (
	"testing"

	"github.com/spf13/cast"
)

type mathTest struct {
	arg1 interface{}
	arg2 interface{}
	want int64
}

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

var addTests = []mathTest{
	{1, 2, 3},
	{"2", 3, 5},
	{2.0, 3, 5},
	{2.9, 3, 5},
	{2.1, 3, 5},
	{-1, 0, -1},
	{"-2", -1, -3},
}

var addTestsWithError = []mathTest{
	{1, 2, 4},
	{"yuyu", 3, 6},
	//{3, "", 6},
	//{3, nil, 6},
	{"~", 3, 6},
}

func TestAdd(t *testing.T) {
	for _, test := range addTests {
		if got := Add(test.arg1, test.arg2); got != test.want {
			if Type(got) != "int64" {
				t.Errorf("Output %q not equal to expected %q", Type(got), "int64")
			}

			t.Errorf("Output %q not equal to expected %q", cast.ToString(got), cast.ToString(test.want))
		}
	}

	for _, test := range addTestsWithError {
		// these tests should fail
		if got := Add(test.arg1, test.arg2); got == test.want {
			t.Errorf("Negative test did not fail and instead returned %q", cast.ToString(got))
		}
	}
}

var subTests = []mathTest{
	{1, 2, -1},
	{"2", 3, -1},
	{2.0, 3, -1},
	{2.9, 3, -1},
	{2.1, 3, -1},
	{-1, 0, -1},
	{"-2", -1, -1},
	{12, 2, 10},
	{-1234, -2345, 1111},
}

var subTestsWithError = []mathTest{
	{1, 2, 4},
	{"yuyu", 3, 6},
	//{3, 3, "", 6},
	//{3, 3, nil, 6},
	{"~", 3, 6},
}

func TestSub(t *testing.T) {
	for _, test := range subTests {
		if got := Sub(test.arg1, test.arg2); got != test.want {
			if Type(got) != "int64" {
				t.Errorf("Output %q not equal to expected %q", Type(got), "int64")
			}

			t.Errorf("Output %q not equal to expected %q", cast.ToString(got), cast.ToString(test.want))
		}
	}

	for _, test := range subTestsWithError {
		// these tests should fail
		if got := Sub(test.arg1, test.arg2); got == test.want {
			t.Errorf("Negative test did not fail and instead returned %q", cast.ToString(got))
		}
	}
}

var mulTests = []mathTest{
	{1, 2, 2},
	{"2", 3, 6},
	{2.0, 3, 6},
	{2.9, 3, 6},
	{2.1, 3, 6},
	{-1, 0, 0},
	{"-2", -1, 2},
}

var mulTestsWithError = []mathTest{
	{1, 2, 4},
	{"yuyu", 3, 6},
	//{3, 3, "", 6},
	//{3, 3, nil, 6},
	{"~", 3, 6},
}

func TestMul(t *testing.T) {
	for _, test := range mulTests {
		if got := Mul(test.arg1, test.arg2); got != test.want {
			if Type(got) != "int64" {
				t.Errorf("Output %q not equal to expected %q", Type(got), "int64")
			}

			t.Errorf("Output %q not equal to expected %q", cast.ToString(got), cast.ToString(test.want))
		}
	}

	for _, test := range mulTestsWithError {
		// these tests should fail
		if got := Mul(test.arg1, test.arg2); got == test.want {
			t.Errorf("Negative test did not fail and instead returned %q", cast.ToString(got))
		}
	}
}

var divTests = []mathTest{
	{2, 2, 1},
	{"2", 3, 0},
	{2.0, 3, 0},
	{2.9, 3, 0},
	{2.1, 3, 0},
	{5, 2, 2},
	{10, 2, 5},
	{"-2", -1, 2},
}

var divTestsWithError = []mathTest{
	{1, 2, 4},
	{"yuyu", 3, 6},
	//{3, 3, "", 6},
	//{3, 3, nil, 6},
	{"~", 3, 6},
	//{1, 0, 0},
	//{-1, 0, 0},
}

func TestDiv(t *testing.T) {
	for _, test := range divTests {
		if got := Div(test.arg1, test.arg2); got != test.want {
			if Type(got) != "int64" {
				t.Errorf("Output %q not equal to expected %q", Type(got), "int64")
			}

			t.Errorf("Output %q not equal to expected %q", cast.ToString(got), cast.ToString(test.want))
		}
	}

	for _, test := range divTestsWithError {
		// these tests should fail
		if got := Div(test.arg1, test.arg2); got == test.want {
			t.Errorf("Negative test did not fail and instead returned %q", cast.ToString(got))
		}
	}
}

var modTests = []mathTest{
	{2, 2, 0},
	{"2", 3, 2},
	{2.0, 3, 2},
	{2.9, 3, 2},
	{2.1, 3, 2},
	{5, 2, 1},
	{10, 2, 0},
	{"-2", -1, 0},
}

var modTestsWithError = []mathTest{
	{1, 2, 4},
	{"yuyu", 3, 6},
	//{3, 3, "", 6},
	//{3, 3, nil, 6},
	{"~", 3, 6},
	//{1, 0, 0},
	//{-1, 0, 0},
}

func TestMod(t *testing.T) {
	for _, test := range modTests {
		if got := Mod(test.arg1, test.arg2); got != test.want {
			if Type(got) != "int64" {
				t.Errorf("Output %q not equal to expected %q", Type(got), "int64")
			}

			t.Errorf("Output %q not equal to expected %q", cast.ToString(got), cast.ToString(test.want))
		}
	}

	for _, test := range modTestsWithError {
		// these tests should fail
		if got := Mod(test.arg1, test.arg2); got == test.want {
			t.Errorf("Negative test did not fail and instead returned %q", cast.ToString(got))
		}
	}
}
