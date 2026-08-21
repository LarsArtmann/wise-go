package wise

import (
	"context"
	"fmt"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// User is the parsed representation of a Wise user account — the owner of an
// API token and its profiles. Details is nil when Wise has no personal
// details on file.
type User struct {
	ID      UserID
	Name    string
	Email   string
	Active  bool
	Details *UserDetails
}

// UserDetails is the personal-details block of a User. PrimaryAddress is the
// address object ID for Wise's addresses endpoints; DateOfBirth is parsed
// from Wise's date-only wire format and is the zero time when unset.
type UserDetails struct {
	FirstName      string
	LastName       string
	PhoneNumber    string
	DateOfBirth    time.Time
	Occupation     string
	Avatar         string
	PrimaryAddress int64
	Address        *UserAddress
}

// UserAddress is the address Wise has on file for a user.
type UserAddress struct {
	CountryCode string
	FirstLine   string
	PostCode    string
	City        string
	State       string
	Occupation  string
}

// GetMe returns the user account that owns the client's API token — the
// identity behind ListProfiles' userId values.
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	var user raw.User

	if err := c.get(ctx, "/v1/me", &user); err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}

	result, err := mapUser(user)
	if err != nil {
		return nil, fmt.Errorf("map current user %d: %w", user.ID, err)
	}

	return &result, nil
}

// GetUser returns a user account by ID. Personal API tokens can only read
// their own user; platform tokens may read their customers'.
func (c *Client) GetUser(ctx context.Context, userID UserID) (*User, error) {
	if userID.Get() == 0 {
		return nil, errorfamily.NewRejection(
			"wise.user.invalid_request",
			"userID is required",
		)
	}

	path := fmt.Sprintf("/v1/users/%d", userID.Get())

	var user raw.User

	if err := c.get(ctx, path, &user); err != nil {
		return nil, fmt.Errorf("get user %d: %w", userID.Get(), err)
	}

	result, err := mapUser(user)
	if err != nil {
		return nil, fmt.Errorf("map user %d: %w", user.ID, err)
	}

	return &result, nil
}

// mapUser converts a raw wire user into the parsed User type.
func mapUser(user raw.User) (User, error) {
	result := User{
		ID:      NewUserID(user.ID),
		Name:    user.Name,
		Email:   user.Email,
		Active:  user.Active,
		Details: nil,
	}

	if user.Detail == nil {
		return result, nil
	}

	dateOfBirth, err := parseWiseDate(user.Detail.DateOfBirth)
	if err != nil {
		return User{}, errorfamily.WrapCorruption(
			err,
			"wise.user.parse_date_of_birth",
			fmt.Sprintf("parse dateOfBirth %q", user.Detail.DateOfBirth),
		)
	}

	var address *UserAddress

	if user.Detail.Address != nil {
		address = &UserAddress{
			CountryCode: user.Detail.Address.CountryCode,
			FirstLine:   user.Detail.Address.FirstLine,
			PostCode:    user.Detail.Address.PostCode,
			City:        user.Detail.Address.City,
			State:       user.Detail.Address.State,
			Occupation:  user.Detail.Address.Occupation,
		}
	}

	result.Details = &UserDetails{
		FirstName:      user.Detail.FirstName,
		LastName:       user.Detail.LastName,
		PhoneNumber:    user.Detail.PhoneNumber,
		DateOfBirth:    dateOfBirth,
		Occupation:     user.Detail.Occupation,
		Avatar:         user.Detail.Avatar,
		PrimaryAddress: user.Detail.PrimaryAddress,
		Address:        address,
	}

	return result, nil
}

// parseWiseDate parses a Wise date-only value ("1977-01-01"). An empty value
// yields the zero time; zoneless dates are interpreted as UTC.
func parseWiseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date %q: %w", s, err)
	}

	return t, nil
}
