package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"service/app/services/sales-api/handlers"
	"service/domain/data/tests"
	"testing"
)

type UsersTest struct {
	app        http.Handler
	userToken  string
	adminToken string
}

func TestUsers(t *testing.T) {
	intTest := tests.NewIntegration(
		t,
		tests.DBContainer{
			Image: "postgres:14-alpine",
			Port:  "5432",
			Args:  []string{"-e", "POSTGRES_PASSWORD=postgres"},
		},
	)
	t.Cleanup(intTest.Teardown)

	shutdown := make(chan os.Signal, 1)
	userTests := UsersTest{
		app: handlers.AppAPIMux(handlers.APIMuxConfig{
			Shutdown: shutdown,
			Log:      intTest.Log,
			Auth:     intTest.Auth,
			DB:       intTest.DB,
		}),
		userToken:  intTest.Token("user@admin.com", "secret"),
		adminToken: intTest.Token("admin@admin.com", "secret"),
	}
	t.Run("genToken200", userTests.genToken200)
	t.Run("genToken404", userTests.genToken200)
}

func (ut *UsersTest) genToken200(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/users/token", nil)
	w := httptest.NewRecorder()

	r.SetBasicAuth("admin@example.com", "secret")
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to issue tokens to known users")
	{
		testID := 0
		t.Log("\tTest %d\t when fething a token with valid credentials", testID)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\t Test %d \t Should recieve a status code of 200 for response %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\t Test %d \t Should recieve a status code of 200 for response ", tests.Succeeded, testID)

			var got struct {
				Token string `json:"token"`
			}

			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d\tShould be able to unmarshal the response %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d\tShould be able to unmarshal the response", tests.Succeeded, testID)
		}
	}
}

func (ut *UsersTest) genToken404(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/users/token", nil)
	w := httptest.NewRecorder()

	r.SetBasicAuth("blahbalh@email.com", "random")
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to deny tokens to unknown users")
	{
		testID := 0
		t.Log("\tTest %d\t when fething a token with an unrecognized email", testID)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\t Test %d \t Should recieve a status code of 404 for response %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\t Test %d \t Should recieve a status code of 404 for response ", tests.Succeeded, testID)
		}
	}
}
