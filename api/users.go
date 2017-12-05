package api

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/mgutz/logxi/v1"

	"github.com/briansan/user-go/errors"
	"github.com/briansan/user-go/schema"
	"github.com/briansan/user-go/store"
)

var (
	logger = log.New("api")
	allows = schema.RoleHasPermission
)

// GetUsers retrieves all users
//   available to roles with ModifyAllUsersRestricted permission
func GetUsers(c echo.Context) error {
	// Type assert user from context and authorize
	user, ok := c.Get("user").(*schema.UserSecure)
	if !ok {
		return echo.ErrUnauthorized
	}
	if !allows(user.Role, schema.PermissionModifyAllUsersRestricted) {
		return echo.ErrForbidden
	}

	// Get db connection
	db, err := store.NewMongoStore()
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, err.Error())
	}
	defer db.Cleanup()

	// Try to get users
	users, err := db.GetAllUsers()
	if err != nil {
		return errors.MongoErrorResponse(err)
	}

	//
	if c.QueryParam("mapped") == "true" {
		m := map[string]*schema.UserSecure{}
		for _, u := range users {
			m[u.ID.Hex()] = u
		}
		return c.JSON(http.StatusOK, m)
	}
	return c.JSON(http.StatusOK, users)
}

func PostUsers(c echo.Context) error {
	u := schema.User{}
	c.Bind(&u)

	// Validate
	if err := u.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get user from db
	db, err := store.NewMongoStore()
	if err != nil {
		return errors.MongoErrorResponse(err)
	}
	defer db.Cleanup()

	// Try to fetch JWT header (for admin)
	var user *schema.UserSecure
	values, ok := c.Request().Header[echo.HeaderAuthorization]
	if ok && len(values) == 1 {
		auth := values[0]
		// Get user id from token
		id, err := AuthenticateJWT(auth)
		if err != nil {
			logger.Warn("jwt auth failed", "reason", err.Error())
			return echo.ErrUnauthorized
		}
		// Try to fetch user by creds
		user, err = db.GetUserByID(id)
		if err != nil {
			return errors.MongoErrorResponse(err)
		}
		if user == nil {
			return echo.ErrUnauthorized
		}
	}

	// If not admin, default role to user
	if user == nil || !allows(user.Role, schema.PermissionModifyAllUsers) {
		u.Role = &schema.RoleUser
	}

	// Try to add user
	if err = db.CreateUser(&u); err != nil {
		return errors.MongoErrorResponse(err)
	}

	return c.JSON(http.StatusCreated, u)
}

func GetUserByUserID(c echo.Context) error {
	userID := c.Param("userID")

	// Type assert user from context and try to return that if user id's match
	user, ok := c.Get("user").(*schema.UserSecure)
	if ok && (user.ID.Hex() == userID || user.Username == userID) {
		return c.JSON(http.StatusOK, user)
	}

	// Don't go any further if user role doesn't have permission to view other users
	if !ok || !allows(user.Role, schema.PermissionModifyAllUsersRestricted) {
		return echo.ErrForbidden
	}

	// Establish db connection
	db, err := store.NewMongoStore()
	if err != nil {
		return errors.MongoErrorResponse(err)
	}
	defer db.Cleanup()

	// Try to fetch by username
	u, err := db.GetUserByUsername(userID)
	if err != nil && err.Error() != "not found" {
		return errors.MongoErrorResponse(err)
	}

	// Return if found
	if u != nil {
		return c.JSON(http.StatusOK, u)
	}

	// Try to fetch by ID
	u, err = db.GetUserByID(userID)
	if err != nil {
		return errors.MongoErrorResponse(err)
	}

	return c.JSON(http.StatusOK, u)
}

func PatchUser(c echo.Context) error {
	userID := c.Param("userID")

	// Type assert user from context and authorize
	user, ok := c.Get("user").(*schema.UserSecure)
	if !ok {
		return echo.ErrUnauthorized
	}

	// Only allow roles who have permission to ModifyAllUsersRestricted
	if user.ID.Hex() != userID && user.Username != userID && !allows(user.Role, schema.PermissionModifyAllUsersRestricted) {
		return echo.ErrForbidden
	}

	// Get user patch doc
	userPatch := &schema.User{}
	c.Bind(userPatch)

	// Only allow ModifyAllUsers permission to touch the Role field
	if userPatch.Role != nil && !allows(user.Role, schema.PermissionModifyAllUsers) {
		return echo.ErrForbidden
	}

	// Establish db connection
	db, err := store.NewMongoStore()
	if err != nil {
		return errors.MongoErrorResponse(err)
	}
	defer db.Cleanup()

	// Authenticate user if password is being touched and isn't an admin
	if userPatch.Password != nil && !allows(user.Role, schema.PermissionModifyAllUsers) {
		if userPatch.OldPassword == nil {
			return echo.NewHTTPError(http.StatusBadRequest, errors.NewValidationError("oldPassword", "string"))
		}
		if user, err = db.GetUserByCreds(user.Username, *userPatch.OldPassword); user == nil {
			return echo.ErrUnauthorized
		}
	}

	// Try to update user
	if user, err = db.UpdateUser(user.ID.Hex(), userPatch); err != nil {
		return errors.MongoErrorResponse(err)
	}
	return c.JSON(http.StatusOK, user)
}

func DeleteUser(c echo.Context) error {
	userID := c.Param("userID")

	// Type assert user from context and authorize
	user, ok := c.Get("user").(*schema.UserSecure)
	if !ok {
		return echo.ErrUnauthorized
	}

	// Only allow roles who have permission to ModifyAllUsers
	if user.ID.Hex() != userID && !allows(user.Role, schema.PermissionModifyAllUsers) {
		return echo.ErrForbidden
	}

	// Establish db connection
	db, err := store.NewMongoStore()
	if err != nil {
		return errors.MongoErrorResponse(err)
	}
	defer db.Cleanup()

	// Try to delete user
	u, err := db.DeleteUser(user.ID.Hex())
	if err != nil {
		return errors.MongoErrorResponse(err)
	}

	return c.JSON(http.StatusOK, u)
}

func initUsers(api *echo.Group) {
	api.GET("/users", GetUsers, DoJWTAuth)
	api.POST("/users", PostUsers)
	api.GET("/users/:userID", GetUserByUserID, DoJWTAuth)
	api.PATCH("/users/:userID", PatchUser, DoJWTAuth)
	api.DELETE("/users/:userID", DeleteUser, DoJWTAuth)
}
