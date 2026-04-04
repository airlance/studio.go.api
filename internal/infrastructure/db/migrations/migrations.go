package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/resoul/studio.go.api/internal/domain"
	"gorm.io/gorm"
)

func All() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "202404041700_initial_schema",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(
					&domain.Profile{},
					&domain.Workspace{},
					&domain.WorkspaceMember{},
					&domain.WorkspaceInvite{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
					&domain.WorkspaceInvite{},
					&domain.WorkspaceMember{},
					&domain.Workspace{},
					&domain.Profile{},
				)
			},
		},
	}
}
