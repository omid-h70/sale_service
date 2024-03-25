package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"io"
	"os"
	"service/domain/data/schema"
	"service/domain/data/store/user"
	"service/domain/sys/auth"
	"service/domain/sys/database"
	"service/foundation/docker"
	"service/foundation/keystore"
	"service/foundation/logger"
	"testing"
	"time"
)

const (
	Succeeded = "\u2713"
	Failed    = "\u2717"
)

type DBContainer struct {
	Image string
	Port  string
	Args  []string
}

func NewUnit(t *testing.T, dbc DBContainer) (*zap.SugaredLogger, *sqlx.DB, func()) {

	r, w, _ := os.Pipe() // to have reader and writer ?
	old := os.Stdout     // make copy of it
	os.Stdout = w

	c := docker.StartContainer(t, dbc.Image, dbc.Port, dbc.Args...)

	db, err := database.Open(database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       c.Host,
		DisableTLS: true,
	})

	if err != nil {
		t.Fatalf("Opening Database Connection %v", err)
	}
	t.Log("waiting for database to be ready")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Migrate(ctx, db); err != nil {
		docker.DumpContainerLogs(t, c.ID)
		docker.StopContainer(t, c.ID)
		t.Fatalf("Migrating error %v", err)
	}

	if err := schema.Seed(ctx, db); err != nil {
		docker.DumpContainerLogs(t, c.ID)
		docker.StopContainer(t, c.ID)
		t.Fatalf("seeding error %v", err)
	}

	log, err := logger.New("TEST")
	if err != nil {
		t.Fatalf("logger error %s", err)
	}

	teardown := func() {
		t.Helper()
		db.Close()
		docker.StopContainer(t, c.ID)
		//log.s
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		os.Stdout = old
		fmt.Println("********************** LOGS *****************************")
		fmt.Print(buf.String())
		fmt.Println("********************** LOGS *****************************")
	}

	return log, db, teardown
}

type Test struct {
	DB       *sqlx.DB
	Log      *zap.SugaredLogger
	Auth     *auth.Auth
	Teardown func()
	t        *testing.T
}

func NewIntegration(t *testing.T, dbc DBContainer) *Test {
	log, db, teardown := NewUnit(t, dbc)

	keyID := "4754d86b-7a6d-4df5-9c65-224741361492"
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	auth, err := auth.New(keyID, keystore.NewMap(map[string]*rsa.PrivateKey{keyID: privateKey}))
	if err != nil {
		t.Fatal(err)
	}

	test := Test{
		DB:       db,
		Log:      log,
		Auth:     auth,
		t:        t,
		Teardown: teardown,
	}
	return &test
}

func (test *Test) Token(email, pass string) string {
	test.t.Log("Generating token for Test ...")

	store := user.NewStore(test.Log, test.DB)
	claims, err := store.Authenticate(context.Background(), time.Now(), email, pass)
	if err != nil {
		test.t.Fatal(err)
	}

	token, err := test.Auth.GenerateToken(claims)
	if err != nil {
		test.t.Fatal(err)
	}
	return token
}

// StringPointer some helper functions in tests
func StringPointer(s string) *string {
	return &s
}

func IntPointer(n int) *int {
	return &n
}
