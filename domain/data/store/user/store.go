package user

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"service/domain/sys/auth"
	"service/domain/sys/database"
	"service/domain/sys/validate"
	"time"
)

type Store struct {
	logger *zap.SugaredLogger
	db     *sqlx.DB
}

func NewStore(log *zap.SugaredLogger, db *sqlx.DB) Store {
	return Store{
		logger: log,
		db:     db,
	}
}

func (s Store) Create(ctx context.Context, nu NewUser, now time.Time) (User, error) {
	if err := validate.Check(nu); err != nil {
		return User{}, err
	}
	hashPass, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("generate hash %w", err)
	}

	usr := User{
		ID:           validate.GenerateUID(),
		Name:         nu.Name,
		Email:        nu.Email,
		PasswordHash: hashPass,
		Roles:        nu.Roles,
		DateCreated:  now,
		DateUpdated:  now,
	}

	q := `INSERT INTO users
	(user_id, name, email, password_hash, roles, date_created, date_updated)
	VALUES
	(:user_id, :name, :email, :password_hash, :roles, :date_created, :date_updated)`

	if err := database.NamedExecContext(ctx, s.logger, s.db, q, usr); err != nil {
		return User{}, fmt.Errorf("inserting user %w", err)
	}

	return usr, nil
}

func (s Store) Update(ctx context.Context, claims auth.Claims, userID string, uu UpdateUser, now time.Time) error {

	if err := validate.CheckID(userID); err != nil {
		return database.ErrInvalidID
	}

	usr, err := s.QueryByID(ctx, claims, userID)
	if err != nil {
		return fmt.Errorf("inserting user %s - %w", userID, err)
	}

	if uu.Name != nil {
		usr.Name = *uu.Name
	}

	if uu.Email != nil {
		usr.Email = *uu.Email
	}

	if uu.Roles != nil {
		usr.Roles = uu.Roles
	}

	if uu.Password != nil {
		pw, err := bcrypt.GenerateFromPassword([]byte(*uu.PasswordConfirm), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("generating password hash - %w", err)
		}
		usr.PasswordHash = pw
	}
	usr.DateUpdated = now

	q := `UPDATE 
		users 
	SET
		"name" = :name
		"email" = :email 
		"roles" = :roles
		"password_hash" = :password_hash
		"date_updated"" = :date_updated
	WHERE
		user_id = :user_id`

	if err := database.NamedExecContext(ctx, s.logger, s.db, q, usr); err != nil {
		return fmt.Errorf("updating user %w", err)
	}

	return nil
}

func (s Store) Delete(ctx context.Context, claims auth.Claims, userID string) error {
	if err := validate.CheckID(userID); err != nil {
		return database.ErrInvalidID
	}

	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return database.ErrForbidden
	}

	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID,
	}

	q := `DELETE FROM  
		users 
			WHERE
		user_id = :user_id`

	if err := database.NamedExecContext(ctx, s.logger, s.db, q, data); err != nil {
		return fmt.Errorf("deleting user %s - %w", userID, err)
	}

	return nil
}

func (s Store) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]User, error) {

	data := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	q := `
	SELECT *
	FROM  
		users 
	ORDER BY
		user_id
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	var users []User
	if err := database.NamedQuerySlice(ctx, s.logger, s.db, q, data, &users); err != nil {
		if err == database.ErrNotFound {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("selecting user %w", err)
	}
	return users, nil
}

func (s Store) QueryByID(ctx context.Context, claims auth.Claims, userID string) (User, error) {
	if err := validate.CheckID(userID); err != nil {
		return User{}, database.ErrInvalidID
	}

	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return User{}, database.ErrForbidden
	}

	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID,
	}

	q :=
		`SELECT *
	FROM  
		users 
	WHERE
		user_id = :user_id`

	var usr User
	if err := database.NamedQueryStruct(ctx, s.logger, s.db, q, data, &usr); err != nil {
		return User{}, fmt.Errorf("deleting user %s - %w", userID, err)
	}

	return usr, nil
}

func (s Store) QueryByEmail(ctx context.Context, claims auth.Claims, email string) (User, error) {

	//TODO: validate the email

	data := struct {
		Email string `db:"email"`
	}{
		Email: email,
	}

	q :=
		`SELECT *
	FROM  
		users 
	WHERE
		email = :email`

	var usr User
	if err := database.NamedQueryStruct(ctx, s.logger, s.db, q, data, &usr); err != nil {
		return User{}, fmt.Errorf("selecting user by email %s - %w", email, err)
	}

	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != usr.ID {
		return User{}, database.ErrForbidden
	}

	return usr, nil
}

func (s Store) Authenticate(ctx context.Context, now time.Time, email, password string) (auth.Claims, error) {

	//TODO: validate the email

	data := struct {
		Email string `db:"email"`
	}{
		Email: email,
	}

	q :=
		`SELECT *
	FROM  
		users 
	WHERE
		email = :email`

	var usr User
	if err := database.NamedQueryStruct(ctx, s.logger, s.db, q, data, &usr); err != nil {
		if err == database.ErrNotFound {
			return auth.Claims{}, database.ErrNotFound
		}
		return auth.Claims{}, fmt.Errorf("selecting user %q %w", email, err)
	}

	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(password)); err != nil {
		return auth.Claims{}, err
	}

	claims := auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    "service project",
			Subject:   usr.ID,
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
			IssuedAt:  time.Now().UTC().Unix(),
		},
		Roles: usr.Roles,
	}
	return claims, nil

}
