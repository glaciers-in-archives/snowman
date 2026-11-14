package cache

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glaciers-in-archives/snowman/internal/utils"
)

func TestWriteCacheHistory(t *testing.T) {
	tests := []struct {
		name         string
		previous     []string // nil means there is no history file
		used         []string
		partialBuild bool
		expected     []string
	}{
		{"full build replaces the history", []string{"a/1", "a/2"}, []string{"b/1"}, false, []string{"b/1"}},
		{"full build removes duplicates", nil, []string{"a/1", "a/1", "a/2"}, false, []string{"a/1", "a/2"}},
		{"partial build adds to the history", []string{"a/1", "a/2"}, []string{"a/2", "b/1"}, true, []string{"a/1", "a/2", "b/1"}},
		{"partial build without history", nil, []string{"a/1"}, true, []string{"a/1"}},
		{"partial build with empty history", []string{}, []string{"a/1"}, true, []string{"a/1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			historyFilePath := filepath.Join(t.TempDir(), "history.txt")
			if tt.previous != nil {
				if err := utils.WriteLineSeperatedFile(tt.previous, historyFilePath); err != nil {
					t.Fatal(err)
				}
			}

			if err := writeCacheHistory(historyFilePath, tt.used, tt.partialBuild); err != nil {
				t.Fatal(err)
			}

			history, err := utils.ReadLineSeperatedFile(historyFilePath)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(history, tt.expected) {
				t.Errorf("Expected history %v, got %v", tt.expected, history)
			}
		})
	}
}
