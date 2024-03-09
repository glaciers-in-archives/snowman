package utils

import (
	"testing"
)

var validatePathSectionTests = []struct {
	pathSection     string
	expectedIsValid bool
}{
	{"a", true},
	{"a b", true},
	{"a-b", true},

	// the path can't be empty
	{"", false},

	// the path can't start with a slash or a dot
	{".test", false},
	{"./test/test", false},
	{"/test/test", false},

	// the path can't end with a slash or a dot
	{"test/", false},
	{"test.", false},

	// the path can't conatain a slash or a dot next to each other
	{"test//test", false},
	{"test..test", false},
	{"test./test", false},
	{"test/.test", false},
	{"test/./test", false},
}

func TestValidatePathSection(t *testing.T) {
	for _, test := range validatePathSectionTests {
		// a valid path section returns nil and an invalid path section returns an error
		if test.expectedIsValid {
			if err := ValidatePathSection(test.pathSection); err != nil {
				t.Errorf("Expected path section \"%s\" to be valid, but got error: %v", test.pathSection, err)
			}
		} else {
			if err := ValidatePathSection(test.pathSection); err == nil {
				t.Errorf("Expected path section \"%s\" to be invalid, but got nil", test.pathSection)
			}
		}
	}
}
