package cache

import (
	"net/url"
	"os"

	"github.com/glaciers-in-archives/snowman/internal/utils"
)

var CurrentResourcesCacheManager *ResourcesCacheManager

type ResourcesCacheManager struct {
	CacheStrategy                 string // "available", "never"
	cacheLocation                 string
	cacheHistoryFile              string
	cacheFormat                   string
	StoredCacheHashes             map[string]bool
	cacheHashesUsedInCurrentBuild []string
	SnowmanDirectoryPath          string
}

func NewResourcesCacheManager(strategy string, snowmanDirectoryPath string) error {
	cm := ResourcesCacheManager{
		CacheStrategy:        strategy,
		cacheLocation:        "/cache/resources/",
		cacheHistoryFile:     "/last_build_resources.txt",
		cacheFormat:          ".txt",
		SnowmanDirectoryPath: snowmanDirectoryPath,
	}
	cm.StoredCacheHashes = make(map[string]bool)
	CurrentResourcesCacheManager = &cm

	if err := os.MkdirAll(cm.SnowmanDirectoryPath+cm.cacheLocation, 0770); err != nil {
		return err
	}

	if strategy != "never" {
		storedCacheItemPaths, err := loadStoredCacheItemHashes(cm.SnowmanDirectoryPath+cm.cacheLocation, cm.cacheFormat)
		if err != nil {
			return err
		}
		cm.StoredCacheHashes = storedCacheItemPaths
	}

	return nil
}

func getFullResourceHashFromUrl(itemUrl string) (string, error) {
	parsedUrl, err := url.Parse(itemUrl)
	if err != nil {
		return "", err
	}

	urlHostnameHash := Hash(parsedUrl.Hostname())
	urlFullHash := Hash(itemUrl)
	return urlHostnameHash + "/" + urlFullHash, nil
}

// TODO: we propebly need another one for induvidual items?
func (cm *ResourcesCacheManager) GetCacheItemByHostname(hostname string) ([]string, error) {
	urlHostnameHash := Hash(hostname)

	cacheLocation := cm.SnowmanDirectoryPath + cm.cacheLocation + urlHostnameHash + "/"

	files, err := os.ReadDir(cacheLocation)
	if err != nil {
		return nil, err
	}

	paths := []string{}
	for _, file := range files {
		paths = append(paths, cacheLocation+"/"+file.Name())
	}
	return paths, nil
}

func (cm *ResourcesCacheManager) GetUnusedCacheHashes() ([]string, error) {
	cachePaths, err := getUnusedCacheHashes(cm.SnowmanDirectoryPath+cm.cacheLocation, cm.SnowmanDirectoryPath+cm.cacheHistoryFile, cm.cacheFormat)
	if err != nil {
		return nil, err
	}

	return cachePaths, nil
}

func (cm *ResourcesCacheManager) GetCache(itemUrl string) (*os.File, error) {
	fullItemHash, err := getFullResourceHashFromUrl(itemUrl)
	if err != nil {
		return nil, err
	}
	cm.cacheHashesUsedInCurrentBuild = append(cm.cacheHashesUsedInCurrentBuild, fullItemHash)

	if !cm.StoredCacheHashes[fullItemHash] || cm.CacheStrategy == "never" {
		return nil, nil
	}

	itemPath := cm.SnowmanDirectoryPath + cm.cacheLocation + fullItemHash + cm.cacheFormat

	return os.Open(itemPath)
}

func (cm *ResourcesCacheManager) SetCache(itemUrl string, content string) error {
	if cm.CacheStrategy == "never" {
		return nil
	}

	fullItemHash, err := getFullResourceHashFromUrl(itemUrl)
	if err != nil {
		return err
	}

	itemPath := cm.SnowmanDirectoryPath + cm.cacheLocation + fullItemHash + cm.cacheFormat
	cm.StoredCacheHashes[fullItemHash] = true

	return saveToCacheFile(itemPath, content)
}

func (cm *ResourcesCacheManager) Teardown() error {
	if err := utils.WriteLineSeperatedFile(cm.cacheHashesUsedInCurrentBuild, cm.SnowmanDirectoryPath+cm.cacheHistoryFile); err != nil {
		return err
	}

	return nil
}
