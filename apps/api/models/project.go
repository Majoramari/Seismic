package models

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectLanguageStat struct {
	Language string `json:"language"`
	Seconds  int    `json:"seconds"`
}

type ProjectOverview struct {
	Project      string                `json:"project"`
	Seconds      int                   `json:"seconds"`
	RepoURL      *string               `json:"repoUrl"`
	WebsiteURL   *string               `json:"websiteUrl"`
	LastWorkedAt *time.Time            `json:"lastWorkedAt"`
	LastCommitAt *time.Time            `json:"lastCommitAt"`
	Languages    []ProjectLanguageStat `json:"languages"`
	Archived     bool                  `json:"archived"`
}

type ProjectCommitSync struct {
	Hash        string  `json:"hash"`
	Message     *string `json:"message"`
	AuthorName  *string `json:"authorName"`
	AuthorEmail *string `json:"authorEmail"`
	CommittedAt *int64  `json:"committedAt"`
}

type ProjectSync struct {
	Project      string              `json:"project"`
	RepoURL      *string             `json:"repoUrl"`
	WebsiteURL   *string             `json:"websiteUrl"`
	LastCommitAt *int64              `json:"lastCommitAt"`
	Commits      []ProjectCommitSync `json:"commits"`
}

func GetProjectOverviews(ctx context.Context, pool *pgxpool.Pool, userID string, archived bool, rangeSQL string, limit int, offset int) ([]ProjectOverview, error) {
	heartbeatRangeSQL := strings.ReplaceAll(rangeSQL, "start_time", "received_at")
	if limit <= 0 {
		limit = 12
	}
	if limit > 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := pool.Query(ctx, `
		WITH project_names AS (
			SELECT project FROM sessions WHERE user_id = $1 AND `+rangeSQL+`
			UNION
			SELECT project FROM heartbeats WHERE user_id = $1 AND `+heartbeatRangeSQL+`
			UNION
			SELECT project_name AS project FROM project_metadata WHERE user_id = $1
		),
		session_totals AS (
			SELECT
				project,
				SUM(duration_seconds)::INT AS seconds,
				MAX(end_time) AS last_worked_at
			FROM sessions
			WHERE user_id = $1 AND `+rangeSQL+`
			GROUP BY project
		),
		heartbeat_meta AS (
			SELECT
				project,
				MAX(repo_url) FILTER (WHERE repo_url IS NOT NULL AND repo_url <> '') AS repo_url,
				MAX(website_url) FILTER (WHERE website_url IS NOT NULL AND website_url <> '') AS website_url,
				MAX(last_commit_at) AS last_commit_at,
				MAX(received_at) AS last_heartbeat_at
			FROM heartbeats
			WHERE user_id = $1 AND `+heartbeatRangeSQL+`
			GROUP BY project
		)
		SELECT
			pn.project,
			COALESCE(st.seconds, 0) AS seconds,
			COALESCE(pm.repo_url, hm.repo_url) AS repo_url,
			COALESCE(pm.website_url, hm.website_url) AS website_url,
			COALESCE(st.last_worked_at, hm.last_heartbeat_at) AS last_worked_at,
			COALESCE(pm.last_commit_at, hm.last_commit_at) AS last_commit_at,
			COALESCE(ps.archived, false) AS archived
		FROM project_names pn
		LEFT JOIN session_totals st ON st.project = pn.project
		LEFT JOIN heartbeat_meta hm ON hm.project = pn.project
		LEFT JOIN project_metadata pm
			ON pm.user_id = $1 AND pm.project_name = pn.project
		LEFT JOIN project_settings ps
			ON ps.user_id = $1 AND ps.project_name = pn.project
		WHERE COALESCE(ps.archived, false) = $2
		ORDER BY COALESCE(st.seconds, 0) DESC, COALESCE(st.last_worked_at, hm.last_heartbeat_at) DESC
		LIMIT $3 OFFSET $4
	`, userID, archived, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []ProjectOverview
	for rows.Next() {
		var p ProjectOverview
		if err := rows.Scan(&p.Project, &p.Seconds, &p.RepoURL, &p.WebsiteURL, &p.LastWorkedAt, &p.LastCommitAt, &p.Archived); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range projects {
		languages, err := getProjectLanguages(ctx, pool, userID, projects[i].Project, rangeSQL)
		if err != nil {
			return nil, err
		}
		projects[i].Languages = languages
	}

	return projects, nil
}

func SyncProject(ctx context.Context, pool *pgxpool.Pool, userID string, sync ProjectSync) error {
	var lastCommitHash *string
	for _, commit := range sync.Commits {
		if commit.Hash == "" {
			continue
		}
		lastCommitHash = &commit.Hash
		break
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO project_metadata (
			user_id, project_name, repo_url, website_url, last_commit_hash, last_commit_at, last_synced_at
		)
		VALUES (
			$1, $2, $3, $4, $5,
			CASE WHEN $6::BIGINT IS NULL THEN NULL ELSE to_timestamp($6::DOUBLE PRECISION / 1000) END,
			now()
		)
		ON CONFLICT (user_id, project_name)
		DO UPDATE SET
			repo_url = COALESCE(EXCLUDED.repo_url, project_metadata.repo_url),
			website_url = COALESCE(EXCLUDED.website_url, project_metadata.website_url),
			last_commit_hash = COALESCE(EXCLUDED.last_commit_hash, project_metadata.last_commit_hash),
			last_commit_at = COALESCE(EXCLUDED.last_commit_at, project_metadata.last_commit_at),
			last_synced_at = now()
	`, userID, sync.Project, sync.RepoURL, sync.WebsiteURL, lastCommitHash, sync.LastCommitAt)
	if err != nil {
		return err
	}

	for _, commit := range sync.Commits {
		if commit.Hash == "" {
			continue
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO project_commits (
				user_id, project_name, repo_url, hash, message, author_name, author_email, committed_at
			)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7,
				CASE WHEN $8::BIGINT IS NULL THEN NULL ELSE to_timestamp($8::DOUBLE PRECISION / 1000) END
			)
			ON CONFLICT (user_id, project_name, hash)
			DO UPDATE SET
				repo_url = COALESCE(EXCLUDED.repo_url, project_commits.repo_url),
				message = COALESCE(EXCLUDED.message, project_commits.message),
				author_name = COALESCE(EXCLUDED.author_name, project_commits.author_name),
				author_email = COALESCE(EXCLUDED.author_email, project_commits.author_email),
				committed_at = COALESCE(EXCLUDED.committed_at, project_commits.committed_at)
		`, userID, sync.Project, sync.RepoURL, commit.Hash, commit.Message, commit.AuthorName, commit.AuthorEmail, commit.CommittedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func SetProjectArchived(ctx context.Context, pool *pgxpool.Pool, userID string, projectName string, archived bool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO project_settings (user_id, project_name, archived, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id, project_name)
		DO UPDATE SET archived = EXCLUDED.archived, updated_at = now()
	`, userID, projectName, archived)
	return err
}

func getProjectLanguages(ctx context.Context, pool *pgxpool.Pool, userID string, projectName string, rangeSQL string) ([]ProjectLanguageStat, error) {
	rows, err := pool.Query(ctx, `
		SELECT language, SUM(duration_seconds)::INT AS seconds
		FROM sessions
		WHERE user_id = $1 AND project = $2 AND `+rangeSQL+`
			AND language NOT IN ('textmate', 'unknown', 'log')
		GROUP BY language
		HAVING SUM(duration_seconds) > 0
		ORDER BY seconds DESC
	`, userID, projectName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var languages []ProjectLanguageStat
	for rows.Next() {
		var language ProjectLanguageStat
		if err := rows.Scan(&language.Language, &language.Seconds); err != nil {
			return nil, err
		}
		languages = append(languages, language)
	}

	if len(languages) == 0 {
		heartbeatRangeSQL := strings.ReplaceAll(rangeSQL, "start_time", "received_at")
		rows, err := pool.Query(ctx, `
			SELECT language, 0 AS seconds
			FROM heartbeats
			WHERE user_id = $1 AND project = $2 AND `+heartbeatRangeSQL+`
				AND language NOT IN ('textmate', 'unknown', 'log')
			GROUP BY language
			ORDER BY language ASC
		`, userID, projectName)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var language ProjectLanguageStat
			if err := rows.Scan(&language.Language, &language.Seconds); err != nil {
				return nil, err
			}
			languages = append(languages, language)
		}
	}

	return languages, rows.Err()
}
