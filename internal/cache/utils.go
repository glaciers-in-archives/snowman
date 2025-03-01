package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/utils"
)

func Hash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

// GetUnusedCacheHashes returns a list of cache paths that were not used in the last build
// it's mainly used for cleaning up the cache directory, note that it returns the full path
func getUnusedCacheHashes(directory string, historyFilePath string, cacheFormat string) ([]string, error) {
	usedAtLastBuild, err := utils.ReadLineSeperatedFile(historyFilePath)
	if err != nil {
		return nil, err
	}

	// TODO: walking the file system like this is slower than checking the stored cache hashes
	// but it's more reliable in terms of finding everything odd in the cache directory
	// however, it might be worth rewriting this
	unusedCacheHashes := []string{}
	err = fs.WalkDir(os.DirFS("."), directory, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip the directory itself
		if path == directory {
			return nil
		}

		// last_build_queries.txt & last_build_recources.txt stores <hash>/<hash>
		pathAsCacheItem := strings.Replace(strings.Replace(path, cacheFormat, "", 1), directory, "", 1)
		isUsed := false
		for _, used := range usedAtLastBuild {
			if pathAsCacheItem == used || strings.HasPrefix(used, pathAsCacheItem) {
				isUsed = true
			}
		}

		if !isUsed {
			unusedCacheHashes = append(unusedCacheHashes, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return unusedCacheHashes, nil
}

func loadStoredCacheItemHashes(directory string, cacheFormat string) (map[string]bool, error) {
	storedCacheHashes := make(map[string]bool)

	locationHashes, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, locationHash := range locationHashes {
		contentDirInfo, err := os.ReadDir(directory + locationHash.Name())
		if err != nil {
			return nil, err
		}

		for _, contentFileInfo := range contentDirInfo {
			fullCacheHash := locationHash.Name() + "/" + strings.Replace(contentFileInfo.Name(), cacheFormat, "", 1)
			storedCacheHashes[fullCacheHash] = true
		}
	}

	return storedCacheHashes, nil
}

func saveToCacheFile(filePath string, content string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0770); err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}

	_, err = f.WriteString(content)
	if err != nil {
		return err
	}

	defer f.Close()
	f.Sync()

	return nil
}
