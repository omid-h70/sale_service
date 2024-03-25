package commands

import (
	"errors"
)

var ErrHelp = errors.New("Help")

/*
func GenToken(logger *zap.SugaredLogger, cfg database.config, userID string, kid string) error {
	if userID == '' || kid == '' {
		fmt.Println("help: gentoken <user_id> <kid>")
		return ErrHelp
	}
	db, err := database.open(cfg)
	if err != nil {
		return fmt.Errorf("connect db err: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	store := user.NewUser(log, db)

	claims := auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Subject: userID,
		},
		Roles: []string{auth.RoleAdmin},
	}

	usr, err := store.QueryById(ctx, claims, userID)
	if err != nil {
		return fmt.Errorf("retrieve user error %w", err)
	}

	keysFolder := "/zarf/keys/"
	ks, err := keystore.NewFS(os.DirFS(keysFolder))
	if err != nil {
		return fmt.Errorf("reading key error %w", err)
	}

	activeKID := "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"
	a, err := auth.New(activeKID, ks)
	if err != nil {
		return fmt.Errorf("constructing auth error %w", err)
	}

	claims := auth.Claims{
		StandardClaim: jwt.StandardClaims{
			Issuer:    "service project",
			Subject:   usr.ID,
			ExpiresAt: time.Now().Add(7860 * time.Hour).Unix(),
			IssuedAt:  time.Now().UTC().Unix(),
		},
		Roles: usr.Roles,
	}

	token, err := a.GenerateKey(claims)
	if err != nil {
		return fmt.Errorf("constructing auth error %w", err)
	}
	return nil
} */
