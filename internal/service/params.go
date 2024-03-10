package service

import (
	"fmt"
	"strings"

	"wonderful/internal/repository"

	"github.com/segmentio/ksuid"
)

// sanitize email substring to prevent SQL injection.
func sanitizeEmail(emailSub string) string {
	invalidChars := []string{"'", "\"", ";", " ", "*", "#"}
	for _, c := range invalidChars {
		emailSub = strings.ReplaceAll(emailSub, c, "")
	}
	return emailSub
}

// ConvertParams converts the API input parameters to the repository.Params type.
func ConvertParams(limit *int, startingAfter, endingBefore, email *string) (*repository.Params, error) {
	params := new(repository.Params)

	// Open API always validates the input, so we can safely assume that the input is valid.
	if limit != nil {
		params.Limit = *limit
	}
	if startingAfter != nil {
		id, err := ksuid.Parse(*startingAfter)
		if err != nil {
			return nil, fmt.Errorf("invalid startingAfter: %w", err)
		}
		params.StartingAfter = &id
	}
	if endingBefore != nil {
		id, err := ksuid.Parse(*endingBefore)
		if err != nil {
			return nil, fmt.Errorf("invalid endingBefore: %w", err)
		}
		params.EndingBefore = &id
	}
	if email != nil {
		sanitizedEmail := sanitizeEmail(*email)
		params.Email = &sanitizedEmail
	}
	return params, nil
}
