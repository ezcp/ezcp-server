package db

import (
	"log"
	"time"

	"errors"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	dbName           = "ezcp"
	dbCollectionName = "tokens"
)

// Token is a stored Token
type Token struct {
	ID         bson.ObjectId `bson:"_id,omitempty"`
	Token      string        `bson:"token"`
	Length     int64         `bson:"len,omitempty"`
	Created    time.Time     `bson:"created"`
	Uploaded   *time.Time    `bson:"up,omitempty"`
	Downloaded *time.Time    `bson:"down,omitempty"`

	// only for permanent tokens
	Permanent bool   `bson:"permanent"`
	Creator   string `bson:"creator,omitempty"`
}

// DB is the interface to the DB module
type DB interface {
	CreateToken(token string) error
	TokenExists(token string, checkNoUpload bool) (bool, error)
	GetToken(token string) (*Token, error)
	TokenUploaded(token string, length int64, timestamp time.Time) error
	TokenDownloaded(token *Token, timestamp time.Time) error
	RemoveExpiredTokens() ([]string, error)
	Close()
}

type db struct {
	session *mgo.Session
}

// NewDB returns a new DB
func NewDB(dbHost string) (DB, error) {
	session, err := mgo.Dial(dbHost)
	if err != nil {
		return nil, err
	}
	log.Print("Connected to ", dbHost)

	err = session.DB(dbName).C(dbCollectionName).EnsureIndex(mgo.Index{
		Key:    []string{"token"},
		Unique: true,
	})
	if err != nil {
		return nil, err
	}
	err = session.DB(dbName).C(dbCollectionName).EnsureIndex(mgo.Index{
		Key: []string{"created"},
	})
	if err != nil {
		return nil, err
	}
	return &db{session}, nil
}

// CreateToken stores a new token
func (db *db) CreateToken(token string) error {
	session, coll := db.tokens()
	defer session.Close()
	error := coll.Insert(&Token{
		Token:     token,
		Created:   time.Now(),
		Permanent: false,
	})
	return error
}

// GetToken returns a token or nil if not found
func (db *db) GetToken(token string) (*Token, error) {
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
func (db *db) TokenExists(token string, checkNoUpload bool) (bool, error) {
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
func (db *db) TokenUploaded(token string, length int64, timestamp time.Time) error {
	session, coll := db.tokens()
	defer session.Close()
	err := coll.Update(bson.M{"token": token}, bson.M{"$set": bson.M{"length": length, "up": timestamp}})
	if err != nil {
		return err
	}
	return err
}

// TokenDownloaded is called once a file has been downloaded
func (db *db) TokenDownloaded(token *Token, timestamp time.Time) error {
	session, coll := db.tokens()
	defer session.Close()

	var err error
	if token.Permanent {
		update := bson.M{
			"$set": bson.M{
				"down": time.Now(),
			},
		}
		err = coll.Update(bson.M{"token": token.ID}, update)
	} else {
		err = coll.Remove(bson.M{"token": token.ID})
	}
	return err
}

// RemoveExpiredTokens removes old unused tokens and returns them
func (db *db) RemoveExpiredTokens() ([]string, error) {
	return nil, nil
}

// Close will definitely close the database
func (db *db) Close() {
	db.session.Close()
	db.session = nil
	log.Print("DB closed")
}

// tokens is used to quickly get hold of a session and collection
func (db *db) tokens() (*mgo.Session, *mgo.Collection) {
	if db.session == nil {
		panic(errors.New("DB is closed already"))
	}
	session := db.session.New()
	session.SetSafe(&mgo.Safe{})
	collection := session.DB(dbName).C(dbCollectionName)
	return session, collection
}
