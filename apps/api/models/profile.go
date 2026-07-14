package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileStats struct {
	TotalSeconds  *int    `json:"totalSeconds,omitempty"`
	TopProject    *string `json:"topProject,omitempty"`
	TopLanguage   *string `json:"topLanguage,omitempty"`
	TopOS         *string `json:"topOS,omitempty"`
	TopEditor     *string `json:"topEditor,omitempty"`
	CurrentStreak *int    `json:"currentStreak,omitempty"`
}

type ProfileVisibility struct {
	HideTime      bool `json:"hideTime"`
	HideProjects  bool `json:"hideProjects"`
	HideLanguages bool `json:"hideLanguages"`
	HideOS        bool `json:"hideOS"`
	HideEditor    bool `json:"hideEditor"`
}

type PublicProfile struct {
	Username     string            `json:"username"`
	DisplayName  string            `json:"displayName"`
	AccountEmail *string           `json:"accountEmail,omitempty"`
	ContactEmail string            `json:"contactEmail"`
	Bio          string            `json:"bio"`
	Website      string            `json:"website"`
	ProfileImage *string           `json:"profileImage"`
	Country      *string           `json:"country"`
	CreatedAt    time.Time         `json:"createdAt"`
	IsOwner      bool              `json:"isOwner"`
	Visibility   ProfileVisibility `json:"visibility"`
	Stats        ProfileStats      `json:"stats"`
	Heatmap      []HeatmapDay      `json:"heatmap,omitempty"`
	Badges       []BadgeInfo       `json:"badges,omitempty"`
}

type UpdateProfileInput struct {
	Username       string
	DisplayName    string
	Bio            string
	Website        string
	ContactEmail   string
	AvatarURL      *string
	AvatarPublicID *string
}

type BadgeInfo struct {
	Type     string    `json:"type"`
	EarnedAt time.Time `json:"earnedAt"`
}

type BadgeVisibilityInfo struct {
	Type     string    `json:"type"`
	EarnedAt time.Time `json:"earnedAt"`
	Hidden   bool      `json:"hidden"`
}

// GetPublicProfile returns a user's public profile data, respecting
// their privacy settings. viewerID is the currently authenticated
// user's ID, or "" if the request is unauthenticated — used to
// determine isOwner and whether to include the private account email.
//
// Fields gated by a privacy toggle (TotalSeconds, TopProject,
// TopLanguage, Heatmap) are only ever populated when the viewer is
// allowed to see them — for a non-owner viewing a profile with that
// toggle on, the field stays nil/empty and is omitted from the JSON
// response entirely (via omitempty), so no hidden data is ever sent
// over the wire, not even as a null placeholder.
func GetPublicProfile(ctx context.Context, pool *pgxpool.Pool, username string, viewerID string) (*PublicProfile, error) {
	var userID string
	var accountEmail string
	var displayName, bio, website, contactEmail *string
	var p PublicProfile

	err := pool.QueryRow(ctx, `
		SELECT id, username, display_name, email, contact_email, bio, website, country, avatar_url, created_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL
	`, username).Scan(
		&userID, &p.Username, &displayName, &accountEmail, &contactEmail,
		&bio, &website, &p.Country, &p.ProfileImage, &p.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	isOwner := viewerID != "" && viewerID == userID
	p.IsOwner = isOwner

	if displayName != nil {
		p.DisplayName = *displayName
	}
	if bio != nil {
		p.Bio = *bio
	}
	if website != nil {
		p.Website = *website
	}
	if contactEmail != nil {
		p.ContactEmail = *contactEmail
	}
	if isOwner {
		p.AccountEmail = &accountEmail
	}

	privacy, err := GetPrivacySettings(ctx, pool, userID)
	if err != nil {
		return nil, err
	}

	if !privacy.ProfilePublic && !isOwner {
		return nil, nil
	}

	p.Visibility = ProfileVisibility{
		HideTime:      privacy.HideTime,
		HideProjects:  privacy.HideProjects,
		HideLanguages: privacy.HideLanguages,
		HideOS:        privacy.HideOS,
		HideEditor:    privacy.HideEditor,
	}

	streak, err := GetCurrentStreak(ctx, pool, userID)
	if err != nil {
		return nil, err
	}
	p.Stats.CurrentStreak = &streak

	showTime := isOwner || !privacy.HideTime
	showProjects := isOwner || !privacy.HideProjects
	showLanguages := isOwner || !privacy.HideLanguages
	showOS := isOwner || !privacy.HideOS
	showEditor := isOwner || !privacy.HideEditor

	summary, err := GetStatsSummary(ctx, pool, userID, "1=1")
	if err != nil {
		return nil, err
	}
	if showOS {
		p.Stats.TopOS = summary.TopOS
	}
	if showEditor {
		p.Stats.TopEditor = summary.TopEditor
	}

	if showTime {
		p.Stats.TotalSeconds = &summary.TotalSeconds

		heatmap, err := GetHeatmap(ctx, pool, userID)
		if err != nil {
			return nil, err
		}
		p.Heatmap = heatmap
	}
	// else: p.Heatmap and p.Stats.TotalSeconds stay nil/empty and are
	// dropped from the JSON response by omitempty — not sent at all.

	if showLanguages {
		languages, err := GetLanguageBreakdown(ctx, pool, userID, "1=1")
		if err != nil {
			return nil, err
		}
		if len(languages) > 0 {
			p.Stats.TopLanguage = &languages[0].Language
		}
	}

	if showProjects {
		projects, err := GetProjectBreakdown(ctx, pool, userID, "1=1")
		if err != nil {
			return nil, err
		}
		if len(projects) > 0 {
			p.Stats.TopProject = &projects[0].Project
		}
	}

	badges, err := GetUserBadges(ctx, pool, userID)
	if err != nil {
		return nil, err
	}

	hidden, err := GetHiddenBadgeTypes(ctx, pool, userID)
	if err != nil {
		return nil, err
	}
	var visible []BadgeInfo
	for _, b := range badges {
		if !hidden[b.Type] {
			visible = append(visible, b)
		}
	}
	badges = visible
	p.Badges = badges

	return &p, nil
}

// UpdateProfile updates the editable profile fields for a user,
// including the avatar URL/public ID from Cloudinary.
func UpdateProfile(ctx context.Context, pool *pgxpool.Pool, userID string, input UpdateProfileInput) error {
	_, err := pool.Exec(ctx, `
		UPDATE users
		SET username = $1, display_name = $2, bio = $3, website = $4,
		    contact_email = $5, avatar_url = $6, avatar_public_id = $7
		WHERE id = $8
	`, input.Username, input.DisplayName, input.Bio, input.Website,
		input.ContactEmail, input.AvatarURL, input.AvatarPublicID, userID)
	return err
}

func GetUserBadges(ctx context.Context, pool *pgxpool.Pool, userID string) ([]BadgeInfo, error) {
	rows, err := pool.Query(ctx, `
		SELECT badge_type, earned_at FROM badges WHERE user_id = $1 ORDER BY earned_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var badges []BadgeInfo
	for rows.Next() {
		var b BadgeInfo
		if err := rows.Scan(&b.Type, &b.EarnedAt); err != nil {
			return nil, err
		}
		badges = append(badges, b)
	}
	return badges, nil
}

func GetUserBadgeVisibility(ctx context.Context, pool *pgxpool.Pool, userID string) ([]BadgeVisibilityInfo, error) {
	rows, err := pool.Query(ctx, `
		SELECT b.badge_type, b.earned_at, hb.user_id IS NOT NULL AS hidden
		FROM badges b
		LEFT JOIN hidden_badges hb
			ON hb.user_id = b.user_id AND hb.badge_type = b.badge_type
		WHERE b.user_id = $1
		ORDER BY b.earned_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var badges []BadgeVisibilityInfo
	for rows.Next() {
		var b BadgeVisibilityInfo
		if err := rows.Scan(&b.Type, &b.EarnedAt, &b.Hidden); err != nil {
			return nil, err
		}
		badges = append(badges, b)
	}
	return badges, nil
}

func GetHiddenBadgeTypes(ctx context.Context, pool *pgxpool.Pool, userID string) (map[string]bool, error) {
	rows, err := pool.Query(ctx, `SELECT badge_type FROM hidden_badges WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hidden := make(map[string]bool)
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		hidden[t] = true
	}
	return hidden, nil
}

func SetBadgeHidden(ctx context.Context, pool *pgxpool.Pool, userID, badgeType string, hidden bool) error {
	if hidden {
		_, err := pool.Exec(ctx, `
			INSERT INTO hidden_badges (user_id, badge_type) VALUES ($1, $2)
			ON CONFLICT (user_id, badge_type) DO NOTHING
		`, userID, badgeType)
		return err
	}
	_, err := pool.Exec(ctx, `DELETE FROM hidden_badges WHERE user_id = $1 AND badge_type = $2`, userID, badgeType)
	return err
}
