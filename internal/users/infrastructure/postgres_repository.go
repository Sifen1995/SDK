package infrastructure

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"skykin-platform/internal/users/domain"
	"skykin-platform/internal/users/infrastructure/persistence"

	"gorm.io/gorm"
)

type postgresUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID == 0 {
		id, err := randomUserID()
		if err != nil {
			return err
		}
		user.ID = id
	}
	row := persistence.UserRowFromDomain(user)
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	user.CreatedAt = row.CreatedAt
	return nil
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var row persistence.UserRow
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *postgresUserRepository) FindAll(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.User, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&persistence.UserRow{}).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []persistence.UserRow
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, len(rows))
	for i := range rows {
		users[i] = rows[i].ToDomain()
	}
	return users, total, nil
}

func randomUserID() (int64, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return 0, fmt.Errorf("generate user id: %w", err)
	}
	n := int64(binary.BigEndian.Uint64(buf[:]) & 0x7fffffffffffffff)
	if n == 0 {
		n = 1
	}
	return n, nil
}
