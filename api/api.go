package api

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/briansan/user-go/config"
)

func New() *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{config.GetWWWHost()},
		AllowCredentials: true,
	}))

	// setup /api
	api := e.Group("/api/v1")

	// setup /api/service
	svc := api.Group("/service")

	// ping pong
	svc.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "pong")
	})

	// setup users
	initAuth(api)
	initUsers(api)

	// setup the rest
	return e
}
