package wise

import (
	"context"
	"fmt"
)

// ListProfiles returns all profiles for the authenticated user.
func (c *Client) ListProfiles(ctx context.Context) ([]ProfileResult, error) {
	var profiles []Profile

	err := c.get(ctx, "/v2/profiles", &profiles)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}

	results := make([]ProfileResult, len(profiles))
	for i, p := range profiles {
		result, mapErr := mapProfile(p)
		if mapErr != nil {
			return nil, fmt.Errorf("map profile %d: %w", p.ID, mapErr)
		}

		results[i] = result
	}

	return results, nil
}

func mapProfile(p Profile) (ProfileResult, error) {
	name := p.FirstName + " " + p.LastName
	if p.Type == "BUSINESS" {
		name = p.BusinessName
	}

	createdAt, err := parseRFC3339(p.CreatedAt)
	if err != nil {
		return ProfileResult{}, fmt.Errorf("parse created_at %q: %w", p.CreatedAt, err)
	}

	profileType, err := parseProfileType(p.Type)
	if err != nil {
		return ProfileResult{}, fmt.Errorf("parse type %q: %w", p.Type, err)
	}

	return ProfileResult{
		ID:        p.ID,
		Type:      profileType,
		Name:      name,
		Email:     p.Email,
		CreatedAt: createdAt,
	}, nil
}

func parseProfileType(s string) (ProfileType, error) {
	switch s {
	case "PERSONAL":
		return ProfileTypePersonal, nil
	case "BUSINESS":
		return ProfileTypeBusiness, nil
	default:
		return "", fmt.Errorf("unknown profile type %q", s)
	}
}
