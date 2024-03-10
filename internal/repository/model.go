package repository

import (
	"time"

	"github.com/segmentio/ksuid"
)

// Params is a struct that holds the parameters for the ListUsers method.
type Params struct {
	Email         *string
	StartingAfter *ksuid.KSUID
	EndingBefore  *ksuid.KSUID
	Limit         int
}

// User is a struct that holds the user information.
type User struct {
	ID           ksuid.KSUID
	Name         string
	Email        string
	Phone        string
	Cell         string
	Picture      map[string]string
	Registration time.Time
}
