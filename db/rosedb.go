package db

import (
	"github.com/rosedblabs/rosedb/v2"
	"time"
)

var sdb *rosedb.DB

func NewRoseDb(DirPath string) error {
	var err error
	options := rosedb.DefaultOptions
	options.DirPath = DirPath
	sdb, err = rosedb.Open(options)
	if err != nil {
		return err

	}
	return nil
}
func PutWithTTL(k, v string, ttl int) error {
	err := sdb.PutWithTTL([]byte(k), []byte(v), time.Duration(ttl)*time.Second)
	return err
}
func Get(k string) (string, error) {
	val, err := sdb.Get([]byte(k))
	return string(val), err
}
func Exist(key []byte) (bool, error) {
	b, err := sdb.Exist(key)
	return b, err
}
func Close() error {
	err := sdb.Sync()
	if err != nil {
		return err
	}
	err = sdb.Close()
	if err != nil {
		return err
	}
	return nil
}
