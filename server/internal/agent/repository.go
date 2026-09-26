package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"qrchat/internal/apperr"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

const agentCols = `id, name, email, password_hash, created_at`

func scanAgent(row pgx.Row) (Agent, error) {
	var a Agent
	err := row.Scan(&a.ID, &a.Name, &a.Email, &a.PasswordHash, &a.CreatedAt)
	return a, err
}

// Create inserts a new agent account, returning apperr.EmailTaken on a duplicate email.
func (r *Repository) Create(ctx context.Context, a Agent) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO agents (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`,
		a.ID, a.Name, a.Email, a.PasswordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperr.EmailTaken
		}
		return fmt.Errorf("insert agent: %w", err)
	}
	return nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (Agent, error) {
	a, err := scanAgent(r.db.QueryRow(ctx, `SELECT `+agentCols+` FROM agents WHERE email = $1`, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return Agent{}, apperr.NotFound
	}
	if err != nil {
		return Agent{}, fmt.Errorf("get agent by email: %w", err)
	}
	return a, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Agent, error) {
	a, err := scanAgent(r.db.QueryRow(ctx, `SELECT `+agentCols+` FROM agents WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Agent{}, apperr.NotFound
	}
	if err != nil {
		return Agent{}, fmt.Errorf("get agent by id: %w", err)
	}
	return a, nil
}
