package function

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"

	"github.com/spf13/cast"
)

func MD5 (arg interface{}) string {
	data := cast.ToString(arg)
	hash := md5.Sum([]byte(data))
	return string(hash[:])
}

func SHA1(arg interface{}) string {
	data := cast.ToString(arg)
	hash := sha1.Sum([]byte(data))
	return string(hash[:])
}

func SHA256(arg interface{}) string {
	data := cast.ToString(arg)
	hash := sha256.Sum256([]byte(data))
	return string(hash[:])
}
