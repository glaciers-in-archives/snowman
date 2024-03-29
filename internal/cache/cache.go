package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/utils"
)

var CacheLocation string = ".snowman/cache/"

func Hash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

type CacheManager struct {
	CacheStrategy          string // "available", "never"
	StoredCacheHashes      map[string]bool
	CacheHashesUsedInBuild []string
}

func NewCacheManager(strategy string) (*CacheManager, error) {
	cm := CacheManager{
		CacheStrategy: strategy,
	}
	cm.StoredCacheHashes = make(map[string]bool)

	if err := os.MkdirAll(CacheLocation, 0770); err != nil {
		return nil, err
	}

	if strategy != "never" {
		if err := cm.readStoredHashes(); err != nil {
			return nil, err
		}
	}

	return &cm, nil
}

func (cm *CacheManager) readStoredHashes() error {
	// index cache hashes
	locationHashes, err := ioutil.ReadDir(CacheLocation)
	if err != nil {
		return err
	}

	for _, locationDirInfo := range locationHashes {
		contentDirInfo, err := ioutil.ReadDir(CacheLocation + locationDirInfo.Name())
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
	cm.CacheHashesUsedInBuild = append(cm.CacheHashesUsedInBuild, fullQueryHash)

	if !cm.StoredCacheHashes[fullQueryHash] || cm.CacheStrategy == "never" {
		return nil, nil
	}

	queryCacheLocation := CacheLocation + fullQueryHash + ".json"

	return os.Open(queryCacheLocation)
}

func (cm *CacheManager) SetCache(location string, query string, content string) error {
	if cm.CacheStrategy == "never" {
		return nil
	}

	fullQueryHash := Hash(location) + "/" + Hash(query)
	queryCacheLocation := CacheLocation + fullQueryHash + ".json"

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
	if err := utils.WriteLineSeperatedFile(cm.CacheHashesUsedInBuild, ".snowman/last_build_queries.txt"); err != nil {
		return err
	}

	return nil
}
