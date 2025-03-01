package cache

import (
	"os"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/utils"
)

type SparqlCacheManager struct {
	CacheStrategy                 string // "available", "never"
	cacheLocation                 string
	cacheHistoryFile              string
	cacheFormat                   string
	StoredCacheHashes             map[string]bool
	cacheHashesUsedInCurrentBuild []string
	SnowmanDirectoryPath          string
}

func NewSparqlCacheManager(strategy string, snowmanDirectoryPath string) (*SparqlCacheManager, error) {
	cm := SparqlCacheManager{
		CacheStrategy:        strategy,
		cacheLocation:        "/cache/sparql/",
		cacheHistoryFile:     "/last_build_queries.txt",
		cacheFormat:          ".json",
		SnowmanDirectoryPath: snowmanDirectoryPath,
	}
	cm.StoredCacheHashes = make(map[string]bool)

	if err := os.MkdirAll(cm.SnowmanDirectoryPath+cm.cacheLocation, 0770); err != nil {
		return nil, err
	}

	if strategy != "never" {
		storedCacheItemPaths, err := loadStoredCacheItemHashes(cm.SnowmanDirectoryPath+cm.cacheLocation, cm.cacheFormat)
		if err != nil {
			return nil, err
		}
		cm.StoredCacheHashes = storedCacheItemPaths
	}

	return &cm, nil
}

// GetCacheItemsByResourceAndArguments returns a list of cache paths given a resource and arguments, like so:
// "myquery.rq", "arg1", "arg2"
// note that if only a location is provided, the resulting path might be a directory
func (cm *SparqlCacheManager) GetCacheItemsByResourceAndArguments(location string, arguments ...string) ([]string, error) {
	locationPathWithHash := cm.SnowmanDirectoryPath + cm.cacheLocation + Hash(location)
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

	cacheFilePath := locationPathWithHash + "/" + Hash(query) + cm.cacheFormat

	return []string{cacheFilePath}, nil
}

func (cm *SparqlCacheManager) GetUnusedCacheHashes() ([]string, error) {
	cachePaths, err := getUnusedCacheHashes(cm.SnowmanDirectoryPath+cm.cacheLocation, cm.SnowmanDirectoryPath+cm.cacheHistoryFile, cm.cacheFormat)
	if err != nil {
		return nil, err
	}

	return cachePaths, nil
}

func (cm *SparqlCacheManager) GetCache(location string, query string) (*os.File, error) {
	fullQueryHash := Hash(location) + "/" + Hash(query)
	cm.cacheHashesUsedInCurrentBuild = append(cm.cacheHashesUsedInCurrentBuild, fullQueryHash)

	if !cm.StoredCacheHashes[fullQueryHash] || cm.CacheStrategy == "never" {
		return nil, nil
	}

	queryCacheLocation := cm.SnowmanDirectoryPath + cm.cacheLocation + fullQueryHash + cm.cacheFormat

	return os.Open(queryCacheLocation)
}

func (cm *SparqlCacheManager) SetCache(location string, query string, content string) error {
	if cm.CacheStrategy == "never" {
		return nil
	}

	fullQueryHash := Hash(location) + "/" + Hash(query)
	queryCacheLocation := cm.SnowmanDirectoryPath + cm.cacheLocation + fullQueryHash + cm.cacheFormat

	cm.StoredCacheHashes[fullQueryHash] = true

	return saveToCacheFile(queryCacheLocation, content)
}

func (cm *SparqlCacheManager) Teardown() error {
	if err := utils.WriteLineSeperatedFile(cm.cacheHashesUsedInCurrentBuild, cm.SnowmanDirectoryPath+cm.cacheHistoryFile); err != nil {
		return err
	}

	return nil
}
