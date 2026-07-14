package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EditorSettings struct {
	UseGitRootProjectName bool `json:"useGitRootProjectName"`
}

func GetEditorSettings(ctx context.Context, pool *pgxpool.Pool, userID string) (*EditorSettings, error) {
	var settings EditorSettings

	err := pool.QueryRow(ctx, `
		SELECT use_git_root_project_name
		FROM editor_settings
		WHERE user_id = $1
	`, userID).Scan(&settings.UseGitRootProjectName)
	if err == nil {
		return &settings, nil
	}

	_, insertErr := pool.Exec(ctx, `
		INSERT INTO editor_settings (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
	`, userID)
	if insertErr != nil {
		return nil, insertErr
	}

	return &EditorSettings{UseGitRootProjectName: true}, nil
}

func UpdateEditorSettings(ctx context.Context, pool *pgxpool.Pool, userID string, updates EditorSettings) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO editor_settings (user_id, use_git_root_project_name, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (user_id)
		DO UPDATE SET
			use_git_root_project_name = EXCLUDED.use_git_root_project_name,
			updated_at = now()
	`, userID, updates.UseGitRootProjectName)
	return err
}
