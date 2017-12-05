package store

import (
	"fmt"
	"time"

	"github.com/mgutz/logxi/v1"
	"gopkg.in/mgo.v2"

	"github.com/briansan/user-go/config"
)

var (
	databaseName = config.GetMongoDatabase()
	logger       = log.New("store")

	mongo *mgo.Session
)

type MongoStore struct {
	s *mgo.Session
}

// InitMongoSession resets the mongo session pointer with updated connection info
func InitMongoSession() error {
	// To avoid a socket leak
	if mongo != nil {
		CleanupMongoSession()
	}

	// Establish new session
	url := config.GetMongoURL()
	logger.Debug("init mongo", "url", url)
	var err error
	if mongo, err = mgo.Dial(url); err != nil {
		return err
	}

	// Ensure indicies
	ensureUserIndex()

	return nil
}

// CleanupMongoSession closes the current session and sets the pointer to nil
func CleanupMongoSession() {
	if mongo == nil {
		return
	}
	mongo.Close()
	time.Sleep(time.Second)

	mongo = nil
}

// Nuke destroys the database if it is in a test environment
func Nuke() error {
	if config.IsTesting() && mongo != nil {
		return mongo.DB(databaseName).DropDatabase()
	}
	return fmt.Errorf("env.TESTING must be set to true")
}

// NewMongoStore returns an instance of the store with a copied mongo session
// error is 500 if mongo ping fails
func NewMongoStore() (*MongoStore, error) {
	if err := mongo.Ping(); err != nil {
		return nil, err
	}
	return &MongoStore{s: mongo.Copy()}, nil
}

// Cleanup closes the mongo session of this store object
func (m *MongoStore) Cleanup() {
	m.s.Close()
}

// GetDatabase returns a pointer to an mgo database object
func (m *MongoStore) GetDatabase() *mgo.Database {
	return m.s.DB(databaseName)
}
