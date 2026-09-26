package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"qrchat/internal/apperr"
	"qrchat/internal/auth"
)

type Service struct {
	repo   *Repository
	issuer *auth.Issuer
}

func NewService(repo *Repository, issuer *auth.Issuer) *Service {
	return &Service{repo: repo, issuer: issuer}
}

// Register creates a new agent account and returns a session, same as Login would.
func (s *Service) Register(ctx context.Context, name, email, password string) (auth.AgentSession, error) {
	name, err := CleanName(name)
	if err != nil {
		return auth.AgentSession{}, err
	}
	email, err = CleanEmail(email)
	if err != nil {
		return auth.AgentSession{}, err
	}
	if err := ValidatePassword(password); err != nil {
		return auth.AgentSession{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return auth.AgentSession{}, fmt.Errorf("hash password: %w", err)
	}
	a := Agent{ID: uuid.New(), Name: name, Email: email, PasswordHash: string(hash)}
	if err := s.repo.Create(ctx, a); err != nil {
		return auth.AgentSession{}, fmt.Errorf("register: %w", err)
	}
	return s.issuer.IssueAgent(a.ID, a.Name)
}

// Login verifies credentials and returns a fresh session.
func (s *Service) Login(ctx context.Context, email, password string) (auth.AgentSession, error) {
	email, err := CleanEmail(email)
	if err != nil {
		return auth.AgentSession{}, apperr.InvalidCredentials
	}
	a, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apperr.NotFound) {
			return auth.AgentSession{}, apperr.InvalidCredentials
		}
		return auth.AgentSession{}, fmt.Errorf("login: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password)) != nil {
		return auth.AgentSession{}, apperr.InvalidCredentials
	}
	return s.issuer.IssueAgent(a.ID, a.Name)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Agent, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Agent{}, fmt.Errorf("get agent: %w", err)
	}
	return a, nil
}
