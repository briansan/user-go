package store

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"gopkg.in/mgo.v2"

	"github.com/briansan/user-go/schema"
)

type StoreTestSuite struct {
	suite.Suite
	store *MongoStore
}

func (suite *StoreTestSuite) SetupTest() {
	// Use test database and reestablish session
	os.Setenv(envDatabaseName, "test")
	InitMongoSession()

	var err error
	suite.store, err = NewMongoStore()
	suite.Nil(err)

	suite.store.GetUsersCollection().RemoveAll(nil)
}

func (suite *StoreTestSuite) TearDownTest() {
	// Use test database and reestablish session
	suite.store.Cleanup()
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

// Test001_User asserts proper CRUD functionality of user object with mongo
func (suite *StoreTestSuite) Test001_User() {
	username := "foo"
	email := "bar"
	pw := "baz"
	role := 1337

	// Test CreateUser
	newUser := &schema.User{
		Username: &username,
		Email:    &email,
		Password: &pw,
		Role:     &role,
	}
	err := suite.store.CreateUser(newUser)
	suite.Nil(err)

	// Test GetUserByUsername
	user, err := suite.store.GetUserByUsername(username)
	suite.Nil(err)
	suite.NotNil(user)
	suite.Equal(username, user.Username)
	suite.Equal(email, user.Email)

	id := user.ID.Hex()

	// Test GetUserByEmail
	user, err = suite.store.GetUserByEmail(email)
	suite.Nil(err)
	suite.NotNil(user)
	suite.Equal(username, user.Username)
	suite.Equal(email, user.Email)

	// Test GetUserByPassword
	user, err = suite.store.GetUserByCreds(username, pw)
	suite.Nil(err)
	suite.NotNil(user)
	suite.Equal(username, user.Username)
	suite.Equal(email, user.Email)

	// Test CreateUser with conflict
	err = suite.store.CreateUser(newUser)
	suite.NotNil(err)
	suite.Equal("user with username as foo already exists", err.Error())

	// Test UpdateUser
	newUsername := "foobar"
	userPatch := &schema.User{Username: &newUsername}
	user, err = suite.store.UpdateUser(id, userPatch)
	suite.Nil(err)
	suite.Equal(newUsername, user.Username)
	suite.Equal(email, user.Email)
	suite.Equal(role, user.Role)

	// Add second user
	err = suite.store.CreateUser(newUser)
	suite.Nil(err)

	// Try to update second user
	u, err := suite.store.GetUserByUsername(*newUser.Username)
	suite.Nil(err)

	user, err = suite.store.UpdateUser(u.ID.Hex(), userPatch)
	suite.Nil(user)
	suite.True(mgo.IsDup(err))

	// Test GetAllUsers
	users, err := suite.store.GetAllUsers()
	suite.Nil(err)
	suite.Equal(len(users), 2)

	// Test DeleteUser
	user, err = suite.store.DeleteUser(id)
	suite.Nil(err)
	suite.NotNil(user)
	suite.Equal(newUsername, user.Username)
	suite.Equal(email, user.Email)
}

// Test002_Admin asserts proper admin insertion
func (suite *StoreTestSuite) Test002_Admin() {
	// Create admin
	err := suite.store.AdminExistsOrCreate("test_secret")
	suite.Nil(err)

	// Fetch admin
	u, err := suite.store.GetUserByCreds("boss", "test_secret")
	suite.Nil(err)
	suite.Equal(schema.RoleAdmin, u.Role)
}
