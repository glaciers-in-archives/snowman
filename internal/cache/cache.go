package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var CacheLocation string = ".snowman/cache/"

type CacheManager struct {
	CacheStrategy string // "available", "never"
	CacheHashes   map[string]bool
}

func NewCacheManager(strategy string) (*CacheManager, error) {
	cm := CacheManager{
		CacheStrategy: strategy,
	}
	cm.CacheHashes = make(map[string]bool)

	if err := os.MkdirAll(CacheLocation, 0770); err != nil {
		return nil, err
	}

	if strategy != "never" {
		// index cache hashes
		locationHashes, err := ioutil.ReadDir(CacheLocation)
		if err != nil {
			return nil, err
		}

		for _, locationDirInfo := range locationHashes {
			contentDirInfo, err := ioutil.ReadDir(CacheLocation + locationDirInfo.Name())
			if err != nil {
				return nil, err
			}
			for _, contentFileInfo := range contentDirInfo {
				fullCacheHash := locationDirInfo.Name() + "/" + strings.Replace(contentFileInfo.Name(), ".json", "", 1)
				cm.CacheHashes[fullCacheHash] = true
			}
		}
	}

	return &cm, nil
}

func (cm *CacheManager) GetCache(location string, query string) (*os.File, error) {
	fullQueryHash := hash(location) + "/" + hash(query)

	if !cm.CacheHashes[fullQueryHash] || cm.CacheStrategy == "never" {
		return nil, nil
	}

	queryCacheLocation := CacheLocation + fullQueryHash + ".json"

	return os.Open(queryCacheLocation)
}

func (cm *CacheManager) SetCache(location string, query string, content string) error {
	if cm.CacheStrategy == "never" {
		return nil
	}

	fullQueryHash := hash(location) + "/" + hash(query)
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

	cm.CacheHashes[fullQueryHash] = true

	return nil
}

func hash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}
