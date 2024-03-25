package user

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-cmp/cmp"
	"service/domain/data/tests"
	"service/domain/sys/auth"
	"service/domain/sys/database"
	"testing"
	"time"
)

var dbContainer = tests.DBContainer{
	Image: "postgres:14-alpine",
	Port:  "5432",
	Args:  []string{"-e", "POSTGRES_PASS=postgres"},
}

//TestUser After Create query it to make sure it exists !

func TestUser(t *testing.T) {

	logger, db, fn := tests.NewUnit(t, dbContainer)
	t.Cleanup(fn)

	store := NewStore(logger, db)

	t.Log("Given the need to work with user records")
	{
		testID := 0
		t.Logf("\t Test %d \t Whem Handling a single User", testID)
		{
			ctx := context.Background()
			now := time.Date(2023, time.August, 1, 0, 0, 0, 0, time.UTC)

			nu := NewUser{
				Name:            "Omid h",
				Email:           "omid.hosseini777@gmail.com",
				Roles:           []string{auth.RoleAdmin},
				Password:        "gopher",
				PasswordConfirm: "gopher",
			}

			usr, err := store.Create(ctx, nu, now)
			if err != nil {
				t.Fatalf("\t%s\t Test %d Should be able to create user %s", tests.Failed, testID, err)
			}
			t.Logf("\t%s\t Test %d Should be able to create user %s", tests.Succeeded, testID, err)

			claims := auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service",
					Subject:   usr.ID,
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
					IssuedAt:  time.Now().UTC().Unix(),
				},
				Roles: []string{auth.RoleUser},
			}

			saved, err := store.QueryByID(ctx, claims, usr.ID)
			if err != nil {
				t.Fatalf("\t%s\t Test %d should be able to retrieve user %s", tests.Failed, testID, err)
			}
			t.Logf("\t%s\t Test %d Should be able to retrieve user %s", tests.Succeeded, testID, err)

			diff := cmp.Diff(usr, saved)
			if diff != "" {
				t.Fatalf("\t%s\t Test %d should be able to match user %s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\t Test %d Should be able to match user %s", tests.Succeeded, testID, diff)

			upd := UpdateUser{
				Name:  tests.StringPointer("nika"),
				Email: tests.StringPointer("stalkeromid2142@gmail.com"),
				Roles: []string{auth.RoleUser},
			}

			claims = auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service",
					Subject:   usr.ID,
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
					IssuedAt:  time.Now().UTC().Unix(),
				},
				//Change me to Admin
				Roles: []string{auth.RoleAdmin},
			}

			err = store.Update(ctx, claims, usr.ID, upd, now)
			if err != nil {
				t.Fatalf("\t%s\t Test %d should be able to update user %s", tests.Failed, testID, err)
			}
			t.Logf("\t%s\t Test %d Should be able to update user %s", tests.Succeeded, testID, err)

			// min 7
			usrByMail, err := store.QueryByEmail(ctx, claims, *upd.Email)
			if err != nil {
				t.Fatalf("\t%s\t Test %d should be able to query by email %s", tests.Failed, testID, err)
			}
			t.Logf("\t%s\t Test %d Should be able to query by email %s", tests.Succeeded, testID, err)

			if usrByMail.Name != *upd.Name {
				t.Errorf("\t%s\t Test %d :\t should be able to see updates to Name", tests.Failed, testID)
				t.Logf("\t\t Test %d Expected %s", testID, *upd.Name)
				t.Logf("\t\t Test %d Got %s", testID, usrByMail.Name)
			} else {
				t.Errorf("\t%s\t Test %d :\t should be able to see updates to Name", tests.Succeeded, testID)
			}

			if usrByMail.Email != *upd.Email {
				t.Errorf("\t%s\t Test %d :\t should be able to see updates to Email", tests.Failed, testID)
				t.Logf("\t\t Test %d Expected %s", testID, *upd.Email)
				t.Logf("\t\t Test %d Got %s", testID, usrByMail.Email)
			} else {
				t.Errorf("\t%s\t Test %d :\t should be able to see updates to Email", tests.Succeeded, testID)
			}

			err = store.Delete(ctx, claims, usr.ID)
			if err != nil {
				t.Fatalf("\t%s\t Test %d should be able to Delete user %s", tests.Failed, testID, err)
			}
			t.Logf("\t%s\t Test %d Should be able to delete user %s", tests.Succeeded, testID, err)

			_, err = store.QueryByID(ctx, claims, usr.ID)
			if !errors.Is(err, database.ErrNotFound) {
				t.Fatalf("\t%s\t Test %d should not be able to query by id anymore %s", tests.Failed, testID, err)
			}
			t.Logf("\t%s\t Test %d Should not be able to query by by id anymore %s", tests.Succeeded, testID, err)
		}
	}
}
