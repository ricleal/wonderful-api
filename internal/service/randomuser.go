package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	randomUserURL = "https://randomuser.me/api/?results=5000"
)

// RandomUser represents a random user from RandomUser API.
type RandomUser struct {
	Results []Results `json:"results"`
	Info    Info      `json:"info"`
}

// Name represents a name of a random user.
type Name struct {
	Title string `json:"title"`
	First string `json:"first"`
	Last  string `json:"last"`
}

// Login represents a login of a random user.
type Login struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Password string `json:"password"`
	Salt     string `json:"salt"`
	Md5      string `json:"md5"`
	Sha1     string `json:"sha1"`
	Sha256   string `json:"sha256"`
}

// Dob represents a date of birth of a random user.
type Dob struct {
	Date time.Time `json:"date"`
	Age  int       `json:"age"`
}

// Registered represents a registered date of a random user.
type Registered struct {
	Date time.Time `json:"date"`
	Age  int       `json:"age"`
}

// ID represents an ID of a random user.
type ID struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Picture represents a picture of a random user.
type Picture struct {
	Large     string `json:"large"`
	Medium    string `json:"medium"`
	Thumbnail string `json:"thumbnail"`
}

// Results represents a result of a random user.
type Results struct {
	Gender     string     `json:"gender"`
	Name       Name       `json:"name"`
	Email      string     `json:"email"`
	Login      Login      `json:"login"`
	Dob        Dob        `json:"dob"`
	Registered Registered `json:"registered"`
	Phone      string     `json:"phone"`
	Cell       string     `json:"cell"`
	ID         ID         `json:"id"`
	Picture    Picture    `json:"picture"`
	Nat        string     `json:"nat"`
}

// Info represents an info of a random user.
type Info struct {
	Seed    string `json:"seed"`
	Results int    `json:"results"`
	Page    int    `json:"page"`
	Version string `json:"version"`
}

// FetchRandomUsers fetches random users from RandomUser API.
func FetchRandomUsers(ctx context.Context, client http.Client) (*RandomUser, error) {
	var out RandomUser
	if err := fetch(ctx, client, randomUserURL, &out); err != nil {
		return nil, errors.Join(ErrRandomUserAPI, err)
	}
	return &out, nil
}

func fetch(ctx context.Context, c http.Client, url string, r any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to GET request: %w", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to GET response: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to GET HTTP status OK: %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}
