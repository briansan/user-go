package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/suite"

	"github.com/briansan/user-go/schema"
	"github.com/briansan/user-go/store"
)

type APITestSuite struct {
	suite.Suite
	e *echo.Echo
}

func (suite *APITestSuite) SetupTest() {
	os.Setenv("BT_MONGO_DATABASE", "test")
	os.Setenv("BT_SECRET", "test_secret")
	os.Setenv("BT_TESTING", "true")

	store.InitMongoSession()
	store.Nuke()
	suite.e = New()
}

func (suite *APITestSuite) Test001_NormalUsage() {
	// 0a. GET /api/service/ping
	code, pong := suite.request("GET", "/api/v1/service/ping", "", nil, nil)
	suite.Equal(http.StatusOK, code)
	suite.Equal("pong", pong)

	// 0b. GET /api/login (as admin)
	var token map[string]string
	code, _ = suite.request(
		"GET", "/api/v1/login",
		basicAuthString("boss", "test_secret"),
		nil, &token,
	)
	suite.Equal(http.StatusOK, code)

	adminSession, ok := token["session"]
	suite.True(ok)

	// 1. POST /api/users
	username, password, email := "foo", "bar", "foo@bar.com"
	user := &schema.User{
		Username: &username,
		Password: &password,
		Email:    &email,
	}

	secureUser := &schema.UserSecure{}
	code, _ = suite.request("POST", "/api/v1/users", "", user, secureUser)
	suite.Equal(http.StatusCreated, code)
	suite.Equal(username, secureUser.Username)
	suite.Equal(email, secureUser.Email)

	// save user id
	uid := secureUser.ID.Hex()

	// 2. GET /api/login
	code, _ = suite.request(
		"GET", "/api/v1/login",
		basicAuthString(username, password),
		nil, &token,
	)
	suite.Equal(http.StatusOK, code)

	session, ok := token["session"]
	suite.True(ok)
	jwtAuth := jwtAuthString(session)

	// 2a. GET /api/users/{userID} (as user)
	secureUser = &schema.UserSecure{}
	code, _ = suite.request("GET", "/api/v1/users/"+uid, jwtAuth, nil, secureUser)
	suite.Equal(http.StatusOK, code)
	suite.Equal(username, secureUser.Username)
	suite.Equal(email, secureUser.Email)

	// 2b. GET /api/users/{username} (as user)
	code, _ = suite.request("GET", "/api/v1/users/"+username, jwtAuth, nil, secureUser)
	suite.Equal(http.StatusOK, code)
	suite.Equal(username, secureUser.Username)
	suite.Equal(email, secureUser.Email)

	// 2c. GET /api/users/{userID} (as admin)
	code, _ = suite.request("GET", "/api/v1/users/"+uid, jwtAuthString(adminSession), nil, secureUser)
	suite.Equal(http.StatusOK, code)
	suite.Equal(username, secureUser.Username)
	suite.Equal(email, secureUser.Email)

	// 2d. GET /api/users/{username} (as admin)
	code, _ = suite.request("GET", "/api/v1/users/"+username, jwtAuthString(adminSession), nil, secureUser)
	suite.Equal(http.StatusOK, code)
	suite.Equal(username, secureUser.Username)
	suite.Equal(email, secureUser.Email)

	// 3a. GET /api/users (as admin)
	users := []*schema.UserSecure{}
	code, _ = suite.request("GET", "/api/v1/users", jwtAuthString(adminSession), nil, &users)
	suite.Equal(http.StatusOK, code)
	suite.Equal(2, len(users))

	// 3b. GET /api/users (fails as user)
	secureUser = &schema.UserSecure{}
	code, _ = suite.request("GET", "/api/v1/users", jwtAuth, nil, secureUser)
	suite.Equal(http.StatusForbidden, code)

	// 4. PATCH /api/users/{userID}
	// 4c. PATCH /api/users/{userID}.Role (fails)
	user = &schema.User{Role: &schema.RoleAdmin}
	secureUser = &schema.UserSecure{}
	code, _ = suite.request("PATCH", "/api/v1/users/"+username, jwtAuth, user, secureUser)
	suite.Equal(http.StatusForbidden, code)

	user = &schema.User{Role: &schema.RoleAdmin}
	secureUser = &schema.UserSecure{}
	code, _ = suite.request("PATCH", "/api/v1/users/"+uid, jwtAuth, user, secureUser)
	suite.Equal(http.StatusForbidden, code)

	// 5. DELETE /api/users/{userID}
	secureUser = &schema.UserSecure{}
	code, _ = suite.request("DELETE", "/api/v1/users/"+uid, jwtAuth, user, secureUser)
	suite.Equal(http.StatusOK, code)
	suite.Equal(username, secureUser.Username)
	suite.Equal(email, secureUser.Email)
}

func (suite *APITestSuite) request(method, path, auth string, body, response interface{}) (int, string) {
	var req *http.Request
	var err error

	if body != nil {
		// interface to json string
		buf, err := json.Marshal(body)
		suite.Nil(err)

		// create request
		req, err = http.NewRequest(method, path, bytes.NewReader(buf))
		suite.Nil(err)

		// set content-type
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		// create request with no body
		req, err = http.NewRequest(method, path, nil)
		suite.Nil(err)
	}
	// set auth
	if len(auth) > 0 {
		req.Header.Set(echo.HeaderAuthorization, auth)
	}

	// record response
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
	resp, _ := ioutil.ReadAll(rec.Body)

	// json string to interface if response
	if resp != nil {
		json.Unmarshal(resp, response)
	}

	return rec.Code, string(resp)
}

func basicAuthString(user, pass string) string {
	b64auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, pass)))
	return fmt.Sprintf("Basic %s", b64auth)
}

func jwtAuthString(jwt string) string {
	return fmt.Sprintf("Bearer %s", jwt)
}

func TestAPI(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
