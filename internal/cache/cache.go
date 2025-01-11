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

type CacheManager struct {
	CacheStrategy                 string // "available", "never"
	StoredCacheHashes             map[string]bool
	cacheHashesUsedInCurrentBuild []string
	cacheHashesUsedInLastBuild    []string
	SnowmanDirectoryPath          string
}

func NewCacheManager(strategy string, snowmanDirectoryPath string) (*CacheManager, error) {
	cm := CacheManager{
		CacheStrategy:        strategy,
		SnowmanDirectoryPath: snowmanDirectoryPath,
	}
	cm.StoredCacheHashes = make(map[string]bool)

	if err := os.MkdirAll(cm.SnowmanDirectoryPath+"/cache/", 0770); err != nil {
		return nil, err
	}

	if strategy != "never" {
		if err := cm.readStoredHashes(); err != nil {
			return nil, err
		}
	}

	return &cm, nil
}

// GetCacheItemsByResourceAndArguments returns a list of cache paths given a resource and arguments, like so:
// "myquery.rq", "arg1", "arg2"
// note that if only a location is provided, the resulting path might be a directory
func (cm *CacheManager) GetCacheItemsByResourceAndArguments(location string, arguments ...string) ([]string, error) {
	locationPathWithHash := cm.SnowmanDirectoryPath + "/cache/" + Hash(location)
	if len(arguments) == 0 { // just the location
		files, err := os.ReadDir(locationPathWithHash)
		if err != nil {
			return nil, err
		}

		if len(files) > 1 {
			paths := []string{}
			for _, file := range files {
				paths = append(paths, locationPathWithHash+"/"+file.Name())
			}
			return paths, nil
		}

		return []string{locationPathWithHash + "/" + files[0].Name()}, nil
	}

	// to build the cache hash, we need to inject the arguments into the query
	sparqlBytes, err := os.ReadFile("queries/" + location)
	if err != nil {
		return nil, err
	}

	query := string(sparqlBytes)
	for _, arg := range arguments {
		query = strings.Replace(query, "{{.}}", arg, 1)
	}

	cacheFilePath := locationPathWithHash + "/" + Hash(query) + ".json"

	return []string{cacheFilePath}, nil
}

func (cm *CacheManager) loadCacheHashesUsedInLastBuild() error {
	lastBuildQueries, err := utils.ReadLineSeperatedFile(cm.SnowmanDirectoryPath + "/last_build_queries.txt")
	if err != nil {
		return err
	}

	cm.cacheHashesUsedInLastBuild = lastBuildQueries
	return nil
}

// GetUnusedCacheHashes returns a list of cache paths that were not used in the last build
// it's mainly used for cleaning up the cache directory, note that it returns the full path
func (cm *CacheManager) GetUnusedCacheHashes() ([]string, error) {
	err := cm.loadCacheHashesUsedInLastBuild()
	if err != nil {
		return nil, err
	}

	// TODO: walking the file system like this is slower than checking the stored cache hashes
	// but it's more reliable in terms of finding everything odd in the cache directory
	// however, it might be worth rewriting this
	unusedCacheHashes := []string{}
	err = fs.WalkDir(os.DirFS("."), cm.SnowmanDirectoryPath+"/cache", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip the directory itself
		if path == cm.SnowmanDirectoryPath+"/cache" {
			return nil
		}

		// last_build_queries.txt stores <hash>/<hash>
		pathAsCacheItem := strings.Replace(strings.Replace(path, ".json", "", 1), cm.SnowmanDirectoryPath+"/cache/", "", 1)
		isUsed := false
		for _, used := range cm.cacheHashesUsedInLastBuild {
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

func (cm *CacheManager) readStoredHashes() error {
	locationHashes, err := os.ReadDir(cm.SnowmanDirectoryPath + "/cache/")
	if err != nil {
		return err
	}

	for _, locationDirInfo := range locationHashes {
		contentDirInfo, err := os.ReadDir(cm.SnowmanDirectoryPath + "/cache/" + locationDirInfo.Name())
		if err != nil {
			return err
		}
		for _, contentFileInfo := range contentDirInfo {
			fullCacheHash := locationDirInfo.Name() + "/" + strings.Replace(contentFileInfo.Name(), ".json", "", 1)
			cm.StoredCacheHashes[fullCacheHash] = true
		}
	}

	return nil
}

func (cm *CacheManager) GetCache(location string, query string) (*os.File, error) {
	fullQueryHash := Hash(location) + "/" + Hash(query)
	cm.cacheHashesUsedInCurrentBuild = append(cm.cacheHashesUsedInCurrentBuild, fullQueryHash)

	if !cm.StoredCacheHashes[fullQueryHash] || cm.CacheStrategy == "never" {
		return nil, nil
	}

	queryCacheLocation := cm.SnowmanDirectoryPath + "/cache/" + fullQueryHash + ".json"

	return os.Open(queryCacheLocation)
}

func (cm *CacheManager) SetCache(location string, query string, content string) error {
	if cm.CacheStrategy == "never" {
		return nil
	}

	fullQueryHash := Hash(location) + "/" + Hash(query)
	queryCacheLocation := cm.SnowmanDirectoryPath + "/cache/" + fullQueryHash + ".json"

	if err := os.MkdirAll(filepath.Dir(queryCacheLocation), 0770); err != nil {
		return err
	}

	f, err := os.Create(queryCacheLocation)
	if err != nil {
		return err
	}

	_, err = f.WriteString(content)
	if err != nil {
		return err
	}

	defer f.Close()
	f.Sync()

	cm.StoredCacheHashes[fullQueryHash] = true

	return nil
}

func (cm *CacheManager) Teardown() error {
	if err := utils.WriteLineSeperatedFile(cm.cacheHashesUsedInCurrentBuild, cm.SnowmanDirectoryPath+"/last_build_queries.txt"); err != nil {
		return err
	}

	return nil
}
