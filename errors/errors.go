package errors

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"gopkg.in/mgo.v2"
)

type ConflictError struct {
	error
	Type  string
	Field string
	Value string
}

func (err ConflictError) Error() string {
	return fmt.Sprintf("%v with %v as %v already exists",
		err.Type, err.Field, err.Value)
}

func NewConflictError(typ, field, value string) error {
	return &ConflictError{
		Type:  typ,
		Field: field,
		Value: value,
	}
}

type ValidationError struct {
	error
	Field string
	Type  string
}

func NewValidationError(f, t string) *ValidationError {
	return &ValidationError{Field: f, Type: t}
}

func (err ValidationError) Error() string {
	return fmt.Sprintf("%v field is required as %v", err.Field, err.Type)
}

func MongoErrorResponse(err error) error {
	if err == mgo.ErrNotFound {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	if mgo.IsDup(err) {
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	}
	if conflict, ok := err.(*ConflictError); ok {
		return echo.NewHTTPError(http.StatusConflict, conflict.Error())
	}
	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
}
