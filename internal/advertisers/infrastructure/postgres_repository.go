package infrastructure

import (
	"context"

	"skykin-platform/internal/advertisers/model"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetRoleBySlug(ctx context.Context, slug string) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repository) CreateAdvertiser(ctx context.Context, a *model.Advertiser) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *Repository) CreatePortalUser(ctx context.Context, u *model.PortalUser) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *Repository) GetPortalUserByEmail(ctx context.Context, email string) (*model.PortalUser, error) {
	var u model.PortalUser
	err := r.db.WithContext(ctx).
		Preload("Role").
		Preload("Advertiser").
		Where("email = ?", email).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetPortalUserByID(ctx context.Context, id string) (*model.PortalUser, error) {
	var u model.PortalUser
	err := r.db.WithContext(ctx).
		Preload("Role").
		Preload("Advertiser").
		Where("id = ?", id).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) SeedRoles(ctx context.Context) error {
	roles := []model.Role{
		{Slug: "operator_admin", DisplayName: "Operator Admin"},
		{Slug: "advertiser", DisplayName: "Advertiser"},
		{Slug: "read_only_analyst", DisplayName: "Read-Only Analyst"},
	}
	for _, role := range roles {
		if err := r.db.WithContext(ctx).Where("slug = ?", role.Slug).FirstOrCreate(&role).Error; err != nil {
			return err
		}
	}
	return nil
}
