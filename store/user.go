package store

import (
	"crypto/sha256"
	"encoding/base64"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/briansan/user-go/errors"
	"github.com/briansan/user-go/schema"
)

const (
	usersCollectionName = "users"
)

var (
	adminUsername = "boss"
	adminEmail    = "bk@breadtech.com"
)

func hash(s string) string {
	hash := sha256.Sum256([]byte(s))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func ensureUserIndex() {
	c := mongo.DB(databaseName).C(usersCollectionName)
	if err := c.EnsureIndex(mgo.Index{
		Key:      []string{"username"},
		Unique:   true,
		DropDups: true,
	}); err != nil {
		panic(err)
	}

}

func newUserQueryByID(id string) bson.M {
	return bson.M{"_id": bson.ObjectIdHex(id)}
}

func newUserQueryByUsername(username string) bson.M {
	return bson.M{"username": username}
}

func newUserQueryByEmail(email string) bson.M {
	return bson.M{"email": email}
}

func newUserQueryByCreds(user, pw string) bson.M {
	return bson.M{"username": user, "password": hash(pw)}
}

// GetUsersCollection returns an mgo instance to the users collection
func (m *MongoStore) GetUsersCollection() *mgo.Collection {
	return m.GetDatabase().C(usersCollectionName)
}

// CreateUser inserts user object into db
// error is 500 if mongo fails, 409 if user exists, else nil
func (m *MongoStore) CreateUser(user *schema.User) error {
	uname := *user.Username
	if user, err := m.GetUserByUsername(uname); err != nil && err != mgo.ErrNotFound {
		return err
	} else if user != nil {
		return errors.NewConflictError("user", "username", uname)
	}

	// Hash the password
	pw := hash(*user.Password)
	user.Password = &pw

	// Try to insert and return error
	user.ID = bson.NewObjectId()
	if err := m.GetUsersCollection().Insert(user); err != nil {
		return err
	}
	return nil
}

// GetAllUsers retrieves all users
func (m *MongoStore) GetAllUsers() ([]*schema.UserSecure, error) {
	users := []*schema.UserSecure{}
	err := m.GetUsersCollection().Find(nil).All(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetUser looks up user in db with given query for entire object (excpet password)
// error is 500 if mongo fails, else nil
func (m *MongoStore) GetUser(q bson.M) (*schema.UserSecure, error) {
	user := schema.UserSecure{}
	err := m.GetUsersCollection().Find(q).One(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID looks up user with given object id
func (m *MongoStore) GetUserByID(id string) (*schema.UserSecure, error) {
	return m.GetUser(newUserQueryByID(id))
}

// GetUserByUsername looks up user with given username
func (m *MongoStore) GetUserByUsername(username string) (*schema.UserSecure, error) {
	return m.GetUser(newUserQueryByUsername(username))
}

// GetUserByCreds looks up user with given username, password
func (m *MongoStore) GetUserByCreds(user, pw string) (*schema.UserSecure, error) {
	return m.GetUser(newUserQueryByCreds(user, pw))
}

// GetUserByEmail looks up user with given email
func (m *MongoStore) GetUserByEmail(email string) (*schema.UserSecure, error) {
	return m.GetUser(newUserQueryByEmail(email))
}

// UpdateUser ...
// TODO check for username exists and email exists
func (m *MongoStore) UpdateUser(userID string, user *schema.User) (*schema.UserSecure, error) {
	// Hash the password if provided
	if user.Password != nil {
		h := hash(*user.Password)
		user.Password = &h
	}

	// Try to update the user
	q := newUserQueryByID(userID)
	changeInfo := mgo.Change{
		Update:    bson.M{"$set": user},
		Upsert:    false,
		ReturnNew: true,
	}
	safeUser := schema.UserSecure{}
	_, err := m.GetUsersCollection().Find(q).Apply(changeInfo, &safeUser)
	if err != nil {
		return nil, err
	}
	return &safeUser, nil
}

// DeleteUser removes user from db with given username
// error is 500 if mongo fails, else nil
func (m *MongoStore) DeleteUser(userID string) (*schema.UserSecure, error) {
	user, err := m.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	if err = m.GetUsersCollection().Remove(newUserQueryByID(userID)); err != nil {
		return user, err
	}

	return user, nil
}

// AdminExistsOrCreate checks for existence of admin account
// and creates one with given key if it doesn't exist
func (m *MongoStore) AdminExistsOrCreate(secret string) error {
	// Try to fetch admin user
	user, err := m.GetUserByUsername(adminUsername)
	if err != nil && err != mgo.ErrNotFound {
		return err
	}

	// Return if admin exists
	if user != nil {
		return nil
	}

	// Create admin
	return m.CreateUser(&schema.User{
		Username: &adminUsername,
		Password: &secret,
		Email:    &adminEmail,
		Role:     &schema.RoleAdmin,
	})
}
