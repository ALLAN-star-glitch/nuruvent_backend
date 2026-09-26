// internal/modules/video/infrastructure/postgres/connection_repository.go

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// ConnectionRepository implements videodomain.ConnectionRepository
// against Postgres.
type ConnectionRepository struct {
	db     *gorm.DB
	cipher videodomain.TokenCipher
}

func NewConnectionRepository(db *gorm.DB, cipher videodomain.TokenCipher) *ConnectionRepository {
	return &ConnectionRepository{db: db, cipher: cipher}
}

// ============================================================
// WRITE
// ============================================================

// Create encrypts the tokens and inserts a new connection.
func (r *ConnectionRepository) Create(ctx context.Context, c *videodomain.Connection) error {
	model, err := r.encryptForWrite(c)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return translateError(err, "create connection", videodomain.ErrConnectionNotFound)
	}
	return nil
}

// Update re-encrypts the tokens and updates mutable fields.
func (r *ConnectionRepository) Update(ctx context.Context, c *videodomain.Connection) error {
	accessEnc, err := r.cipher.Encrypt(c.AccessToken)
	if err != nil {
		return fmt.Errorf("encrypt access token: %w", err)
	}
	refreshEnc, err := r.cipher.Encrypt(c.RefreshToken)
	if err != nil {
		return fmt.Errorf("encrypt refresh token: %w", err)
	}

	res := r.db.WithContext(ctx).
		Model(&VideoConnectionModel{}).
		Where("id = ?", c.ID).
		Updates(map[string]any{
			"external_user_id":        c.ExternalUserID,
			"external_email":          c.ExternalEmail,
			"external_org_id":         c.ExternalOrgID,
			"access_token_encrypted":  accessEnc,
			"refresh_token_encrypted": refreshEnc,
			"token_expires_at":        c.TokenExpiresAt,
			"scopes":                  c.Scopes,
			"revoked_at":              c.RevokedAt,
			"updated_at":              time.Now().UTC(),
		})
	if res.Error != nil {
		return translateError(res.Error, "update connection", videodomain.ErrConnectionNotFound)
	}
	if res.RowsAffected == 0 {
		return videodomain.ErrConnectionNotFound
	}
	return nil
}

// Revoke marks the connection as revoked. Idempotent.
func (r *ConnectionRepository) Revoke(ctx context.Context, id string, now time.Time) error {
	res := r.db.WithContext(ctx).
		Model(&VideoConnectionModel{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Updates(map[string]any{
			"revoked_at": now,
			"updated_at": now,
		})
	if res.Error != nil {
		return translateError(res.Error, "revoke connection", videodomain.ErrConnectionNotFound)
	}
	if res.RowsAffected == 0 {
		// Already revoked or not found — treat as not found for
		// callers that want to distinguish.
		return videodomain.ErrConnectionNotFound
	}
	return nil
}

// ============================================================
// READ
// ============================================================

func (r *ConnectionRepository) FindByID(ctx context.Context, id string) (*videodomain.Connection, error) {
	var m VideoConnectionModel
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, videodomain.ErrConnectionNotFound
	}
	if err != nil {
		return nil, translateError(err, "find connection", videodomain.ErrConnectionNotFound)
	}
	return r.decryptFromRead(&m)
}

func (r *ConnectionRepository) FindActiveByUserAndPlatform(
	ctx context.Context,
	userID string,
	platform videodomain.Platform,
) (*videodomain.Connection, error) {
	var m VideoConnectionModel
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND platform = ? AND revoked_at IS NULL",
			userID, string(platform)).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, videodomain.ErrConnectionNotFound
	}
	if err != nil {
		return nil, translateError(err, "find active connection", videodomain.ErrConnectionNotFound)
	}
	return r.decryptFromRead(&m)
}

func (r *ConnectionRepository) ListByUser(
	ctx context.Context,
	userID string,
) ([]*videodomain.Connection, error) {
	var models []VideoConnectionModel
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "list connections", videodomain.ErrConnectionNotFound)
	}
	out := make([]*videodomain.Connection, 0, len(models))
	for i := range models {
		conn, err := r.decryptFromRead(&models[i])
		if err != nil {
			return nil, err
		}
		out = append(out, conn)
	}
	return out, nil
}

// ============================================================
// HELPERS
// ============================================================

// encryptForWrite maps a domain connection to a model, encrypting
// tokens in the process.
func (r *ConnectionRepository) encryptForWrite(c *videodomain.Connection) (*VideoConnectionModel, error) {
	accessEnc, err := r.cipher.Encrypt(c.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("encrypt access token: %w", err)
	}
	refreshEnc, err := r.cipher.Encrypt(c.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("encrypt refresh token: %w", err)
	}
	model := toConnectionModel(c)
	model.AccessTokenEncrypted = accessEnc
	model.RefreshTokenEncrypted = refreshEnc
	return model, nil
}

// decryptFromRead maps a model to a domain connection, decrypting
// tokens in the process.
func (r *ConnectionRepository) decryptFromRead(m *VideoConnectionModel) (*videodomain.Connection, error) {
	access, err := r.cipher.Decrypt(m.AccessTokenEncrypted)
	if err != nil {
		return nil, fmt.Errorf("decrypt access token: %w", err)
	}
	refresh, err := r.cipher.Decrypt(m.RefreshTokenEncrypted)
	if err != nil {
		return nil, fmt.Errorf("decrypt refresh token: %w", err)
	}
	conn := toConnectionDomain(m)
	conn.AccessToken = access
	conn.RefreshToken = refresh
	return conn, nil
}

// compile-time assertion
var _ videodomain.ConnectionRepository = (*ConnectionRepository)(nil)