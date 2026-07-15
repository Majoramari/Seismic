package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PrivacySettings struct {
	HideProjects    bool `json:"hideProjects"`
	HideTime        bool `json:"hideTime"`
	HideLanguages   bool `json:"hideLanguages"`
	HideOS          bool `json:"hideOS"`
	HideEditor      bool `json:"hideEditor"`
	HideLeaderboard bool `json:"hideLeaderboard"`
	ProfilePublic   bool `json:"profilePublic"`
}

// GetPrivacySettings returns a user's privacy settings,
// creating a default row first if one doesn't exist yet.
func GetPrivacySettings(ctx context.Context, pool *pgxpool.Pool, userID string) (*PrivacySettings, error) {
	var p PrivacySettings

	err := pool.QueryRow(ctx, `
	SELECT
		hide_projects,
		hide_time,
		hide_languages,
		hide_os,
		hide_editor,
		hide_leaderboard,
		profile_public
	FROM privacy_settings
	WHERE user_id = $1
	`, userID).Scan(
		&p.HideProjects,
		&p.HideTime,
		&p.HideLanguages,
		&p.HideOS,
		&p.HideEditor,
		&p.HideLeaderboard,
		&p.ProfilePublic,
	)

	if err != nil {
		// No row yet, create default settings
		_, insertErr := pool.Exec(ctx, `
			INSERT INTO privacy_settings (user_id) VALUES ($1)
		`, userID)
		if insertErr != nil {
			return nil, insertErr
		}
		return &PrivacySettings{ProfilePublic: true}, nil
	}

	return &p, nil
}

// UpdatePrivacySettings updates only the fields provided,
// leaving others unchanged.
func UpdatePrivacySettings(ctx context.Context, pool *pgxpool.Pool, userID string, updates map[string]bool) error {
	for field, value := range updates {
		column := ""
		switch field {
		case "hideProjects":
			column = "hide_projects"
		case "hideTime":
			column = "hide_time"
		case "hideLanguages":
			column = "hide_languages"
		case "hideOS":
			column = "hide_os"
		case "hideEditor":
			column = "hide_editor"
		case "hideLeaderboard":
			column = "hide_leaderboard"
		case "profilePublic":
			column = "profile_public"
		default:
			continue
		}

		_, err := pool.Exec(ctx, `
			UPDATE privacy_settings SET `+column+` = $1, updated_at = now() WHERE user_id = $2
		`, value, userID)
		if err != nil {
			return err
		}
	}
	return nil
}

// ResetUserTimers deletes all tracked coding activity and activity-earned badges
// for a user, giving them a fresh start while preserving manually granted role badges.
func ResetUserTimers(ctx context.Context, pool *pgxpool.Pool, userID string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	preservedBadges := []string{"supporter", "contributor", "maintainer"}

	_, err = tx.Exec(ctx, `
		DELETE FROM hidden_badges
		WHERE user_id = $1
			AND badge_type <> ALL($2)
	`, userID, preservedBadges)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM badges
		WHERE user_id = $1
			AND badge_type <> ALL($2)
	`, userID, preservedBadges)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `DELETE FROM heartbeats WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// DeleteUserAccount soft-deletes an account by anonymizing
// their username/email and setting deleted_at.
func DeleteUserAccount(ctx context.Context, pool *pgxpool.Pool, userID string) error {
	_, err := pool.Exec(ctx, `
		UPDATE users
		SET username = 'deleted_' || substr(id::text, 1, 8),
		    email = 'deleted_' || id || '@deleted.seismic.icu',
		    api_key = gen_random_uuid(),
		    deleted_at = now()
		WHERE id = $1
	`, userID)
	return err
}
