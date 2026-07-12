package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PublicProfile struct {
	Username      string         `json:"username"`
	Bio           *string        `json:"bio"`
	Website       *string        `json:"website"`
	Country       *string        `json:"country"`
	AvatarURL     *string        `json:"avatarUrl"`
	CreatedAt     time.Time      `json:"createdAt"`
	TotalSeconds  *int           `json:"totalSeconds"` // nil if hidden
	TopLanguage   *string        `json:"topLanguage"`  // nil if hidden
	TopProject    *string        `json:"topProject"`   // nil if hidden
	CurrentStreak int            `json:"currentStreak"`
	Languages     []LanguageStat `json:"languages,omitempty"` // nil if hidden
}

// GetPublicProfile returns a user's public profile data,
// respecting their privacy settings. Returns nil if the
// user doesn't exist, is deleted, or has profile_public=false.
func GetPublicProfile(ctx context.Context, pool *pgxpool.Pool, username string) (*PublicProfile, error) {
	var userID string
	var p PublicProfile

	err := pool.QueryRow(ctx, `
		SELECT id, username, bio, website, country, avatar_url, created_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL
	`, username).Scan(&userID, &p.Username, &p.Bio, &p.Website, &p.Country, &p.AvatarURL, &p.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	privacy, err := GetPrivacySettings(ctx, pool, userID)
	if err != nil {
		return nil, err
	}

	if !privacy.ProfilePublic {
		return nil, nil // treat as not found, don't leak that the account exists
	}

	streak, err := GetCurrentStreak(ctx, pool, userID)
	if err != nil {
		return nil, err
	}
	p.CurrentStreak = streak

	if !privacy.HideTime {
		summary, err := GetStatsSummary(ctx, pool, userID, "1=1") // all time
		if err != nil {
			return nil, err
		}
		p.TotalSeconds = &summary.TotalSeconds
	}

	if !privacy.HideLanguages {
		languages, err := GetLanguageBreakdown(ctx, pool, userID, "1=1")
		if err != nil {
			return nil, err
		}
		p.Languages = languages
		if len(languages) > 0 {
			p.TopLanguage = &languages[0].Language
		}
	}

	if !privacy.HideProjects {
		projects, err := GetProjectBreakdown(ctx, pool, userID, "1=1")
		if err != nil {
			return nil, err
		}
		if len(projects) > 0 {
			p.TopProject = &projects[0].Project
		}
	}

	return &p, nil
}
