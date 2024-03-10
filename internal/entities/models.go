package entities

import "time"

// User is a struct that holds the user information.
type User struct {
	ID           string
	Name         string
	Email        string
	Phone        string
	Cell         string
	Picture      map[string]string
	Registration time.Time
}
