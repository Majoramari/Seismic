package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RefreshToken represents a row in the refresh_tokens table.
type RefreshToken struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}

// GenerateRawToken creates a random 32-byte token, returned
// as a hex string. This is what gets sent to the client and
// stored in their browser
func GenerateRawToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken hashes a raw token before storing or looking it
// up, same idea as hashing a password.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// CreateRefreshToken generates a new raw token, stores its
// hash, and returns the raw token to send to the client.
func CreateRefreshToken(ctx context.Context, pool *pgxpool.Pool, userID string) (string, error) {
	raw, err := GenerateRawToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	_, err = pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, hashToken(raw), expiresAt)
	if err != nil {
		return "", err
	}

	return raw, nil
}

// FindValidRefreshToken looks up a refresh token by its raw
// value. Returns nil if not found, revoked, or expired.
func FindValidRefreshToken(ctx context.Context, pool *pgxpool.Pool, raw string) (*RefreshToken, error) {
	var rt RefreshToken

	err := pool.QueryRow(ctx, `
		SELECT id, user_id, expires_at, revoked, created_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked = false
	`, hashToken(raw)).Scan(
		&rt.ID, &rt.UserID, &rt.ExpiresAt, &rt.Revoked, &rt.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, nil
	}

	return &rt, nil
}

// RevokeRefreshToken marks a token as revoked, used during
// rotation or logout.
func RevokeRefreshToken(ctx context.Context, pool *pgxpool.Pool, id string) error {
	_, err := pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked = true WHERE id = $1
	`, id)
	return err
}

// RevokeAllUserRefreshTokens revokes every non-revoked token for a user.
// Useful when rotating all tokens (e.g. password change, account compromise).
func RevokeAllUserRefreshTokens(ctx context.Context, pool *pgxpool.Pool, userID string) error {
	_, err := pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked = true
		WHERE user_id = $1 AND revoked = false
	`, userID)
	return err
}

// CleanupExpiredRefreshTokens deletes all expired tokens from the database.
// Should be called periodically to prevent the table from growing unbounded.
func CleanupExpiredRefreshTokens(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	res, err := pool.Exec(ctx, `
		DELETE FROM refresh_tokens WHERE expires_at < NOW()
	`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}
