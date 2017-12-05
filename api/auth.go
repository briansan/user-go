package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"

	"github.com/briansan/user-go/config"
	"github.com/briansan/user-go/errors"
	"github.com/briansan/user-go/store"
)

const (
	sessionDuration = time.Hour
)

var (
	secret = config.GetSecret()
)

func initSecret() {
	// Get user from db
	db, err := store.NewMongoStore()
	if err != nil {
		panic(err)
	}
	defer db.Cleanup()

	// Try to fetch user by creds
	if err := db.AdminExistsOrCreate(string(secret)); err != nil {
		panic(err)
	}
}

// NewJWTSession creates a jwt token with
//   aud = user
//   exp = now + sessionDuration
//   iss = now
func NewJWTSession(user string) (string, error) {
	claims := &jwt.StandardClaims{
		Audience:  user,
		ExpiresAt: time.Now().Add(sessionDuration).Unix(),
		IssuedAt:  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(secret)
	if err != nil {
		logger.Warn("issue signing jwt", "err", err)
		return "", err
	}
	return ss, nil
}

// AuthenticateJWT ensures that input jwt string matches
//   signature of secret and is valid within the given time
//   returns aud field (username)
func AuthenticateJWT(authString string) (string, error) {
	// Break up auth string by "Bearer" and "jwt"
	parts := strings.Split(authString, " ")
	if len(parts) != 2 {
		return "", fmt.Errorf("failed to split in 2")
	}
	jwtString := parts[1]

	// Try to parse jwtString
	token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		// Validate alg is HMAC
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method != jwt.SigningMethodHS256 {
			logger.Warn("bad jwt alg", "alg", token.Header["alg"])
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return "", err
	}

	// Type assert as a map
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to assert as map claims")
	}

	// Retrieve aud field
	user, ok := claims["aud"].(string)
	if !ok {
		return "", fmt.Errorf("no aud field")
	}
	return user, nil
}

// DoJWTAuth is a middleware function that will try to
//   validate the Authorization:Bearer token and fetch the
//   corresponding user
func DoJWTAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get Authorization header value
		values := c.Request().Header[echo.HeaderAuthorization]
		if len(values) != 1 {
			return echo.ErrUnauthorized
		}
		auth := values[0]

		// Get user id from token
		id, err := AuthenticateJWT(auth)
		if err != nil {
			logger.Warn("jwt auth failed", "reason", err.Error())
			return echo.ErrUnauthorized
		}

		// Get user from db
		db, err := store.NewMongoStore()
		if err != nil {
			return errors.MongoErrorResponse(err)
		}
		defer db.Cleanup()

		// Try to fetch user by creds
		user, err := db.GetUserByID(id)
		if err != nil {
			return errors.MongoErrorResponse(err)
		}
		if user == nil {
			return echo.ErrUnauthorized
		}

		c.Set("user", user)
		return next(c)
	}
}

func GetLogin(c echo.Context) error {
	// Get basic auth creds
	u, p, ok := c.Request().BasicAuth()
	if !ok {
		return echo.ErrUnauthorized
	}

	// Authenticate
	db, err := store.NewMongoStore()
	if err != nil {
		return errors.MongoErrorResponse(err)
	}
	defer db.Cleanup()

	// Try to fetch user by creds
	user, err := db.GetUserByCreds(u, p)
	if err != nil {
		if err.Error() == "not found" {
			return echo.ErrUnauthorized
		}
		if err != nil {
			return errors.MongoErrorResponse(err)
		}
	}

	// Create JWT token
	token, err := NewJWTSession(user.ID.Hex())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"session": token})
}

func initAuth(api *echo.Group) {
	initSecret()
	api.GET("/login", GetLogin)
}
