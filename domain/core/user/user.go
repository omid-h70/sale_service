package user

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"service/domain/data/store/user"
	"service/domain/sys/auth"
	"time"
)

// Role represents a role in the system.
type Role struct {
	name string
}

type Core struct {
	logger *zap.SugaredLogger
	user   user.Store
}

func NewCore(log *zap.SugaredLogger, db *sqlx.DB) Core {
	return Core{
		logger: log,
		user:   user.NewStore(log, db),
	}
}

func (c Core) Create(ctx context.Context, nu user.NewUser, now time.Time) (user.User, error) {

	usr, err := c.user.Create(ctx, nu, now)
	if err != nil {
		return user.User{}, fmt.Errorf("create: %w", err)
	}
	return usr, nil
}

func (c Core) Update(ctx context.Context, claims auth.Claims, userID string, uu user.UpdateUser, now time.Time) error {
	err := c.user.Update(ctx, claims, userID, uu, now)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	return nil
}

func (c Core) Delete(ctx context.Context, claims auth.Claims, userID string) error {
	err := c.user.Delete(ctx, claims, userID)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	return nil
}

func (c Core) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]user.User, error) {
	usr, err := c.user.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return []user.User{}, fmt.Errorf("Query: %w", err)
	}
	return usr, nil
}

func (c Core) QueryByID(ctx context.Context, claims auth.Claims, userID string) (user.User, error) {
	usr, err := c.user.QueryByID(ctx, claims, userID)
	if err != nil {
		return user.User{}, fmt.Errorf("QueryByID: %w", err)
	}
	return usr, nil
}

func (c Core) QueryByEmail(ctx context.Context, claims auth.Claims, email string) (user.User, error) {
	usr, err := c.user.QueryByEmail(ctx, claims, email)
	if err != nil {
		return user.User{}, fmt.Errorf("QueryByEmail: %w", err)
	}
	return usr, nil
}

func (c Core) Authenticate(ctx context.Context, now time.Time, email, password string) (auth.Claims, error) {
	claims, err := c.user.Authenticate(ctx, now, email, password)
	if err != nil {
		return auth.Claims{}, fmt.Errorf("create: %w", err)
	}
	return claims, nil
}
