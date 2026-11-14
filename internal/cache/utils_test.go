package cache

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
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

func TestGetUnusedCacheHashes(t *testing.T) {
	snowmanDirectory := t.TempDir()
	// the cache managers pass the directory with a trailing slash
	cacheDirectory := snowmanDirectory + "/cache/sparql/"
	historyFilePath := snowmanDirectory + "/last_build_queries.txt"

	for _, cacheItem := range []string{"used/1", "used/2", "partlyused/1", "partlyused/2", "unused/1"} {
		if err := saveToCacheFile(cacheDirectory+cacheItem+".json", "{}"); err != nil {
			t.Fatal(err)
		}
	}
	if err := utils.WriteLineSeperatedFile([]string{"used/1", "used/2", "partlyused/1"}, historyFilePath); err != nil {
		t.Fatal(err)
	}

	unused, err := getUnusedCacheHashes(cacheDirectory, historyFilePath, ".json")
	if err != nil {
		t.Fatal(err)
	}

	// a directory is unused when none of the cache items in it are used
	expected := []string{
		filepath.Join(cacheDirectory, "partlyused", "2.json"),
		filepath.Join(cacheDirectory, "unused"),
		filepath.Join(cacheDirectory, "unused", "1.json"),
	}
	sort.Strings(unused)
	if !reflect.DeepEqual(unused, expected) {
		t.Errorf("Expected unused cache items %v, got %v", expected, unused)
	}

	for _, path := range unused {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Expected the unused cache item %s to exist: %v", path, err)
		}
	}
}
