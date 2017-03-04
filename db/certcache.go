package db

import (
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2/bson"
)

type kv struct {
	Key   string `bson:"key"`
	Value []byte `bson:"value"`
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (db *DB) Get(ctx context.Context, key string) ([]byte, error) {
	session, coll := db.certs()
	defer session.Close()

	var result []kv
	err := coll.Find(bson.M{"key": key}).All(&result)
	if err != nil {
		panic(err)
	}
	if len(result) != 1 {
		return nil, autocert.ErrCacheMiss
	}
	return result[0].Value, nil
}

// Put stores the data in the cache under the specified key.
// Underlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (db *DB) Put(ctx context.Context, key string, data []byte) error {
	session, coll := db.certs()
	defer session.Close()

	_, err := coll.Upsert(bson.M{"key": key}, bson.M{"$set": bson.M{"value": data}})
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (db *DB) Delete(ctx context.Context, key string) error {
	session, coll := db.certs()
	defer session.Close()

	err := coll.Remove(bson.M{"key": key})
	if err != nil {
		return err
	}
	return nil
}
