package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

func Test001_Conflict(t *testing.T) {
	err := NewConflictError("foo", "bar", "baz")
	assert.Equal(t, "foo with bar as baz already exists", err.Error())
}

func Test002_Validation(t *testing.T) {
	err := NewValidationError("foo", "bar")
	assert.Equal(t, "foo field is required as bar", err.Error())
}

func Test003_Mongo(t *testing.T) {
	var err error
	err = MongoErrorResponse(mgo.ErrNotFound)
	assert.Equal(t, "code=404, message=not found", err.Error())

	err = MongoErrorResponse(NewConflictError("foo", "bar", "baz"))
	assert.Equal(t, "code=409, message=foo with bar as baz already exists", err.Error())

	err = MongoErrorResponse(fmt.Errorf("foo"))
	assert.Equal(t, "code=500, message=foo", err.Error())
}
