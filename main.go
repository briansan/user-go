package main

import (
	"github.com/briansan/user-go/api"
	"github.com/briansan/user-go/store"
)

const ()

func main() {
	err := store.InitMongoSession()
	if err != nil {
		panic(err)
	}

	api.New().Start(":8888")
}
