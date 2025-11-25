package database

import (
	"errors"
	"os"
	"path/filepath"

	bolt "go.etcd.io/bbolt"
)

func Open(username string) (*bolt.DB, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "SentryVault", "users", username+".db")
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	if err = createBuckets(db); err != nil {
		return nil, err
	}
	return db, nil
}

func createBuckets(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Header"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("Content"))
		if err != nil {
			return err
		}
		return nil
	})
}

func SetHeaders(db *bolt.DB, combinedTitle, salt []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Header"))
		if b == nil {
			return errors.New("header bucket not found")
		}
		if err := b.Put([]byte("combinedTitle"), combinedTitle); err != nil {
			return err
		}
		if err := b.Put([]byte("salt"), salt); err != nil {
			return err
		}
		return nil
	})
}

func GetHeaders(db *bolt.DB) ([]byte, []byte, error) {
	var combinedTitle, salt []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Header"))
		if b == nil {
			return errors.New("header bucket not found")
		}
		combinedTitle = b.Get([]byte("combinedTitle"))
		if combinedTitle == nil {
			return errors.New("combined title not found")
		}
		salt = b.Get([]byte("salt"))
		if salt == nil {
			return errors.New("salt not found")
		}
		return nil
	})
	return combinedTitle, salt, err
}

func CreateEntry(db *bolt.DB, entry []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Content"))
		if b == nil {
			return errors.New("bucket \"content\" not found")
		}
		_, err := b.CreateBucket(entry)
		if err != nil {
			return err
		}
		return nil
	})
}

func RemoveEntry(db *bolt.DB, entry []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Content"))
		if b == nil {
			return errors.New("bucket \"content\" not found")
		}
		return b.DeleteBucket(entry)
	})
}

func GetEntries(db *bolt.DB) ([][]byte, error) {
	var entries [][]byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Content"))
		if b == nil {
			return errors.New("bucket \"content\" not found")
		}
		c := b.Cursor()
		for key, _ := c.First(); key != nil; key, _ = c.Next() {
			entries = append(entries, key)
		}
		return nil
	})
	return entries, err
}

func Insert(db *bolt.DB, entry, key, value []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Content"))
		if b == nil {
			return errors.New("bucket \"content\" not found")
		}
		b = b.Bucket(entry)
		if b == nil {
			return errors.New("bucket \"" + string(entry) + "\" not found")
		}
		if err := b.Put(key, value); err != nil {
			return err
		}
		return nil
	})
}

func RetrieveAll(db *bolt.DB, entry []byte) ([][][]byte, error) {
	var pairs [][][]byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Content"))
		if b == nil {
			return errors.New("bucket \"content\" not found")
		}
		b = b.Bucket(entry)
		if b == nil {
			return errors.New("bucket \"" + string(entry) + "\" not found")
		}
		c := b.Cursor()
		for key, value := c.First(); key != nil && value != nil; key, value = c.Next() {
			pairs = append(pairs, [][]byte{key, value})
		}
		return nil
	})
	return pairs, err
}

func Remove(db *bolt.DB, entry, key []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Content"))
		if b == nil {
			return errors.New("bucket \"content\" not found")
		}
		b = b.Bucket(entry)
		if b == nil {
			return errors.New("bucket \"" + string(entry) + "\" not found")
		}
		err := b.Delete(key)
		return err
	})
}
