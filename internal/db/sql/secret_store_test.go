package sql

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/KarlGW/burnit/internal/db"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestCreateSecretQueries(t *testing.T) {
	tests := []struct {
		name  string
		input struct {
			driver Driver
			table  string
		}
		want    secretQueries
		wantErr error
	}{
		{
			name: "postgres",
			input: struct {
				driver Driver
				table  string
			}{
				driver: DriverPostgres,
				table:  "secrets",
			},
			want: secretQueries{
				selectByID:    "SELECT id, value, expires_at FROM secrets WHERE id = $1",
				insert:        "INSERT INTO secrets (id, value, expires_at) VALUES ($1, $2, $3)",
				delete:        "DELETE FROM secrets WHERE id = $1",
				deleteExpired: "DELETE FROM secrets WHERE expires_at < NOW() AT TIME ZONE 'UTC'",
			},
		},
		{
			name: "mssql",
			input: struct {
				driver Driver
				table  string
			}{
				driver: DriverMSSQL,
				table:  "secrets",
			},
			want: secretQueries{
				selectByID:    "SELECT ID, Value, ExpiresAt FROM Secrets WHERE ID = @p1",
				insert:        "INSERT INTO Secrets (ID, Value, ExpiresAt) VALUES (@p1, @p2, @p3)",
				delete:        "DELETE FROM Secrets WHERE ID = @p1",
				deleteExpired: "DELETE FROM Secrets WHERE ExpiresAt < GETUTCDATE()",
			},
		},
		{
			name: "sqlite",
			input: struct {
				driver Driver
				table  string
			}{
				driver: DriverSQLite,
				table:  "secrets",
			},
			want: secretQueries{
				selectByID:    "SELECT id, value, expires_at FROM secrets WHERE id = ?1",
				insert:        "INSERT INTO secrets (id, value, expires_at) VALUES (?1, ?2, ?3)",
				delete:        "DELETE FROM secrets WHERE id = ?1",
				deleteExpired: "DELETE FROM secrets WHERE expires_at < DATETIME('now')",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, gotErr := createSecretQueries(test.input.driver, test.input.table)

			if diff := cmp.Diff(test.want, got, cmp.AllowUnexported(secretQueries{})); diff != "" {
				t.Errorf("createSecretQueries() = unexpected result (-want +got)\n%s\n", diff)
			}

			if diff := cmp.Diff(test.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("createSecretQueries() = unexpected error (-want +got)\n%s\n", diff)
			}
		})
	}
}

func TestSecretStore_CreateAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	tests := []struct {
		name string
	}{
		{
			name: "create secret and retrieve secret",
		},
	}

	password := os.Getenv("DB_PASSWORD")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Run("postgres", func(t *testing.T) {
				store, err := setupSecretStore(DriverPostgres, password)
				if err != nil {
					t.Fatalf("could not setup secret store: %v", err)
				}

				ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
				defer cancel()

				id := uuid.New().String()
				val := uuid.New().String()
				expiresAt := time.Now().Add(5 * time.Second)

				gotSecret, createErr := store.Create(ctx, db.Secret{
					ID:        id,
					Value:     val,
					ExpiresAt: expiresAt,
				})
				if createErr != nil {
					t.Errorf("Create() = unexpected error: %v", createErr)
				}

				if gotSecret.ID != id {
					t.Errorf("Create() = unexpected ID, want: %s, got: %s\n", id, gotSecret.ID)
				}

				gotSecret, getErr := store.Get(ctx, gotSecret.ID)
				if getErr != nil {
					t.Errorf("Get() = unexpected error: %v", getErr)
				}

				if gotSecret.ID != id {
					t.Errorf("Create() = unexpected ID, want: %s, got: %s\n", id, gotSecret.ID)
				}
			})
		})
	}
}

func setupSecretStore(driver Driver, password string) (db.SecretStore, error) {
	var dsn string
	switch driver {
	case DriverPostgres:
		dsn = fmt.Sprintf("postgres://postgres:%s@localhost:5432/burnit", password)
	default:
		return nil, errors.New("could not determine driver")
	}

	client, err := NewClient(func(o *ClientOptions) {
		o.Driver = driver
		o.DSN = dsn
	})
	if err != nil {
		return nil, fmt.Errorf("database client: %w", err)
	}

	store, err := NewSecretStore(client)
	if err != nil {
		return nil, fmt.Errorf("secret store: %w", err)
	}

	return store, nil
}
