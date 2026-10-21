package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

const (
	bucketName = "books"
)

type DBBook struct {
	Name         string
	Link         string
	LastDownload int
	Like         bool
	DownloadedAt time.Time
}

func OpenDB() (*bbolt.DB, error) {
	userCfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("Error finding config directory: %v", err)
	}

	appDir := filepath.Join(userCfgDir, "Dodeyo")
	dbPath := filepath.Join(appDir, "data.db")

	err = os.MkdirAll(appDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("Error creating db directory: %v", err)
	}

	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("Error opening database: %v", err)
	}
	return db, nil
}

func SaveBook(db *bbolt.DB, book DBBook) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("Error opening bucket: %v", err)
		}

		encodedBook, err := json.Marshal(book)
		if err != nil {
			return fmt.Errorf("Error marshiling book into db: %v", err)
		}

		return bucket.Put([]byte(book.Link), encodedBook)
	})
}

func GetBook(db *bbolt.DB, link string) (bool, *DBBook, error) {
	var book DBBook
	var exists bool

	err := db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))

		data := bucket.Get([]byte(link))
		if data == nil {
			exists = false
			return nil
		}

		exists = true
		return json.Unmarshal(data, &book)
	})

	if err != nil {
		return false, nil, fmt.Errorf("Error getting book from db: %v", err)
	}

	return exists, &book, nil
}
