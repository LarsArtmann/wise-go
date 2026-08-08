package wise

import (
	"context"
	"fmt"

	id "github.com/larsartmann/go-branded-id"
	"github.com/larsartmann/wise-go/internal/raw"
)

// ListProfiles returns all profiles for the authenticated user.
func (c *Client) ListProfiles(ctx context.Context) ([]ProfileResult, error) {
	var profiles []raw.Profile

	err := c.get(ctx, "/v2/profiles", &profiles)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}

	results := make([]ProfileResult, 0, len(profiles))
	for _, p := range profiles {
		result, mapErr := mapProfile(p)
		if mapErr != nil {
			return nil, fmt.Errorf("map profile %d: %w", p.ID, mapErr)
		}

		results = append(results, result)
	}

	return results, nil
}

func mapProfile(p raw.Profile) (ProfileResult, error) {
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
		ID:        id.NewID[ProfileBrand](p.ID),
		Type:      profileType,
		Name:      name,
		Email:     p.Email,
		CreatedAt: createdAt,
	}, nil
}

func parseProfileType(s string) (ProfileType, error) {
	return parseEnum(map[string]ProfileType{
		"PERSONAL": ProfileTypePersonal,
		"BUSINESS": ProfileTypeBusiness,
	}, s, "profile type")
}
