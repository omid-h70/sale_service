package schema

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/ardanlabs/darwin"
	"github.com/jmoiron/sqlx"
	"service/domain/sys/database"
)

var (
	//go:embed sql/schema.sql
	schemaDoc string
	//embed above file into that variable, so its within binary !

	//go:embed sql/seed.sql
	schemaSeed string

	//go:embed sql/delete.sql
	deleteDoc string
)

func Migrate(ctx context.Context, db *sqlx.DB) error {
	if err := database.StatusCheck(ctx, db); err != nil {
		return fmt.Errorf("status chech database %w", err)
	}

	driver, err := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})
	if err != nil {
		return fmt.Errorf("construct darwin driver %w", err)
	}

	d := darwin.New(driver, darwin.ParseMigrations(schemaDoc))
	return d.Migrate()
}

func Seed(ctx context.Context, db *sqlx.DB) error {

	if err := database.StatusCheck(ctx, db); err != nil {
		return fmt.Errorf("status chech database %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(schemaSeed); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}

func DeleteAll(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(deleteDoc); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}
