package service

import "errors"

// ErrRandomUserAPI is an error when fetching random users from the RandomUserAPI.
var ErrRandomUserAPI = errors.New("error fetching random users from the RandomUserAPI")
