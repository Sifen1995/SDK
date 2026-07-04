package infrastructure

import (
	"context"
	"time"

	"skykin-platform/internal/permissions/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PermissionModel struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(100)"`
	Resource    string    `gorm:"type:varchar(50)"`
	Action      string    `gorm:"type:varchar(50)"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time
}

func (PermissionModel) TableName() string { return "rbac_permissions" }

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

var _ domain.PermissionRepository = (*PermissionRepository)(nil)

func (r *PermissionRepository) FindAll(ctx context.Context) ([]*domain.Permission, error) {
	var rows []PermissionModel
	if err := r.db.WithContext(ctx).Order("resource, action").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Permission, len(rows))
	for i := range rows {
		out[i] = permissionToDomain(&rows[i])
	}
	return out, nil
}

func (r *PermissionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	var row PermissionModel
	if err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&row).Error; err != nil {
		return nil, err
	}
	return permissionToDomain(&row), nil
}

func permissionToDomain(row *PermissionModel) *domain.Permission {
	id, _ := uuid.Parse(row.ID)
	return &domain.Permission{
		ID: id, Name: row.Name, Resource: row.Resource,
		Action: row.Action, Description: row.Description, CreatedAt: row.CreatedAt,
	}
}
