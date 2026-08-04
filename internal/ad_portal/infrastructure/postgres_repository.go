package infrastructure

import (
	"context"
	"errors"

	"skykin-platform/internal/ad_portal/domain"
	"skykin-platform/internal/ad_portal/infrastructure/persistence"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) quiet(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Session(&gorm.Session{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

func (r *Repository) GetRoleBySlug(ctx context.Context, slug string) (*domain.Role, error) {
	var row persistence.RoleRow
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *Repository) CreateAdvertiser(ctx context.Context, a *domain.Advertiser) error {
	row := persistence.AdvertiserRowFromDomain(a)
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	*a = *row.ToDomain()
	return nil
}

func (r *Repository) CreatePortalUser(ctx context.Context, u *domain.PortalUser) error {
	row := persistence.PortalUserRowFromDomain(u)
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	*u = *row.ToDomain()
	return nil
}

func (r *Repository) GetPortalUserByEmail(ctx context.Context, email string) (*domain.PortalUser, error) {
	var row persistence.PortalUserRow
	err := r.quiet(ctx).
		Preload("Role").
		Preload("Advertiser").
		Where("email = ?", email).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *Repository) GetPortalUserByID(ctx context.Context, id string) (*domain.PortalUser, error) {
	var row persistence.PortalUserRow
	err := r.db.WithContext(ctx).
		Preload("Role").
		Preload("Advertiser").
		Where("id = ?", id).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *Repository) SeedRoles(ctx context.Context) error {
	roles := []persistence.RoleRow{
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

func (r *Repository) CreateAdvertiserAndPortalUser(ctx context.Context, adv *domain.Advertiser, u *domain.PortalUser) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		advRow := persistence.AdvertiserRowFromDomain(adv)
		var existing persistence.AdvertiserRow
		if err := tx.Where("company_name = ?", advRow.CompanyName).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(advRow).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			advRow.ID = existing.ID
		}

		u.AdvertiserID = &advRow.ID
		userRow := persistence.PortalUserRowFromDomain(u)
		if err := tx.Create(userRow).Error; err != nil {
			return err
		}
		*u = *userRow.ToDomain()
		*adv = *advRow.ToDomain()
		return nil
	})
}
