package database

import (
	"encoding/json"
	"errors"
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

func StartDB() error {
	configDir, _ := os.UserConfigDir()
	appDir := filepath.Join(configDir, "Dodeyo")
	_ = os.MkdirAll(appDir, 0755)
	dbPath := filepath.Join(appDir, "data.db")

	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return fmt.Errorf("Could not open database: %v", err)
	}
	defer db.Close()

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	if err != nil {
		return fmt.Errorf("Could not create bucket: %v", err)
	}
	return nil
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

func GetAllBooks(db *bbolt.DB) ([]DBBook, error) {
	var books []DBBook
	err := db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return errors.New("Bucket not found")
		}

		return bucket.ForEach(func(key, value []byte) error {
			var book DBBook

			if err := json.Unmarshal(value, &book); err != nil {
				return err
			}

			books = append(books, book)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return books, nil
}
