package db

import (
	"log"
	"time"

	"errors"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	dbName               = "ezcp"
	tokensCollectionName = "tokens"
	txCollectionName     = "tx"
	certsCollectionName  = "certs"
)

// Token is a stored Token
type Token struct {
	Token      string     `bson:"token"`
	Length     int64      `bson:"len,omitempty"`
	Created    time.Time  `bson:"created"`
	Uploaded   *time.Time `bson:"up,omitempty"`
	Downloaded *time.Time `bson:"down,omitempty"`

	// only for permanent tokens
	Permanent bool   `bson:"permanent"`
	Creator   string `bson:"creator,omitempty"`
}

// DB models our db
type DB struct {
	session *mgo.Session

	bitgoToken  BitgoToken
	bitgoWallet string
}

// BitgoToken exist to avoid confusion
type BitgoToken string

// NewDB returns a new DB
func NewDB(dbHost string, token BitgoToken, wallet string) (*DB, error) {
	session, err := mgo.Dial(dbHost)
	if err != nil {
		return nil, err
	}
	log.Print("Connected to ", dbHost)

	err = session.DB(dbName).C(tokensCollectionName).EnsureIndex(mgo.Index{
		Key:    []string{"token"},
		Unique: true,
	})
	if err != nil {
		return nil, err
	}
	err = session.DB(dbName).C(tokensCollectionName).EnsureIndex(mgo.Index{
		Key: []string{"created"},
	})
	if err != nil {
		return nil, err
	}
	err = session.DB(dbName).C(txCollectionName).EnsureIndex(mgo.Index{
		Key:    []string{"id"},
		Unique: true,
	})
	if err != nil {
		return nil, err
	}
	err = session.DB(dbName).C(certsCollectionName).EnsureIndex(mgo.Index{
		Key:    []string{"key"},
		Unique: true,
	})
	if err != nil {
		return nil, err
	}
	return &DB{session, token, wallet}, nil
}

// CreateToken stores a new token
func (db *DB) CreateToken(token string) error {
	session, coll := db.tokens()
	defer session.Close()
	error := coll.Insert(&Token{
		Token:     token,
		Created:   time.Now(),
		Permanent: false,
	})
	return error
}

// CreateDurableToken stores a new token
func (db *DB) CreateDurableToken(token string, creator string) error {
	session, coll := db.tokens()
	defer session.Close()
	error := coll.Insert(&Token{
		Token:     token,
		Created:   time.Now(),
		Permanent: true,
		Creator:   creator,
	})
	return error
}

// GetToken returns a token or nil if not found
func (db *DB) GetToken(token string) (*Token, error) {
	session, coll := db.tokens()
	defer session.Close()

	var tokens []Token
	err := coll.Find(bson.M{"token": token}).All(&tokens)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, nil
	}
	return &tokens[0], nil
}

// TokenExists checks if a token exists
func (db *DB) TokenExists(token string, checkNoUpload bool) (bool, error) {
	session, coll := db.tokens()
	defer session.Close()
	var q interface{}
	if checkNoUpload {
		q = bson.M{"token": token, "up": nil}
	} else {
		q = bson.M{"token": token}
	}
	query := coll.Find(q)
	count, err := query.Count()
	if err != nil {
		return false, err
	}
	return count == 1, nil
}

// TokenUploaded is called once a file has been uploaded
func (db *DB) TokenUploaded(token string, length int64, timestamp time.Time) error {
	session, coll := db.tokens()
	defer session.Close()
	err := coll.Update(bson.M{"token": token}, bson.M{"$set": bson.M{"length": length, "up": timestamp}})
	if err != nil {
		return err
	}
	return err
}

// TokenDownloaded is called once a file has been downloaded
func (db *DB) TokenDownloaded(token *Token, timestamp time.Time) error {
	session, coll := db.tokens()
	defer session.Close()

	var err error
	if token.Permanent {
		update := bson.M{
			"$set": bson.M{
				"down": time.Now(),
			},
		}
		err = coll.Update(bson.M{"token": token.Token}, update)
	} else {
		err = coll.Remove(bson.M{"token": token.Token})
	}
	return err
}

// RemoveExpiredTokens removes old unused tokens and returns them
func (db *DB) RemoveExpiredTokens() ([]string, error) {
	return nil, nil
}

// Close will definitely close the database
func (db *DB) Close() {
	db.session.Close()
	db.session = nil
	log.Print("DB closed")
}

// tokens is used to quickly get hold of a session and collection
func (db *DB) tokens() (*mgo.Session, *mgo.Collection) {
	if db.session == nil {
		panic(errors.New("DB is closed already"))
	}
	session := db.session.New()
	session.SetSafe(&mgo.Safe{})
	collection := session.DB(dbName).C(tokensCollectionName)
	return session, collection
}

// tx is used to quickly get hold of a session and TX collection
func (db *DB) tx() (*mgo.Session, *mgo.Collection) {
	if db.session == nil {
		panic(errors.New("DB is closed already"))
	}
	session := db.session.New()
	session.SetSafe(&mgo.Safe{})
	collection := session.DB(dbName).C(txCollectionName)
	return session, collection
}

// certs is used to quickly get hold of a session and certs collection
func (db *DB) certs() (*mgo.Session, *mgo.Collection) {
	if db.session == nil {
		panic(errors.New("DB is closed already"))
	}
	session := db.session.New()
	session.SetSafe(&mgo.Safe{})
	collection := session.DB(dbName).C(certsCollectionName)
	return session, collection
}

// StoreTransaction stores a transaction to mongodb
func (db *DB) StoreTransaction(tx *Transaction) error {
	session, coll := db.tx()
	defer session.Close()
	error := coll.Insert(tx)
	return error
}

// LoadTransaction loads a transaction from mongodb
func (db *DB) LoadTransaction(txid string) (*Transaction, error) {
	session, coll := db.tx()
	defer session.Close()
	var transactions []Transaction
	err := coll.Find(bson.M{"id": txid}).All(&transactions)
	if err != nil {
		return nil, err
	}
	if len(transactions) == 0 {
		return nil, nil
	}
	return &transactions[0], nil

}
