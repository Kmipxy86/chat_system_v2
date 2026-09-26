package support

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"qrchat/internal/auth"
	"qrchat/internal/room"
)

type Service struct {
	repo   *Repository
	issuer *auth.Issuer
}

func NewService(repo *Repository, issuer *auth.Issuer) *Service {
	return &Service{repo: repo, issuer: issuer}
}

const defaultSubject = "แชทสนับสนุนลูกค้า"

type StartResult struct {
	auth.Session
	Subject string `json:"subject"`
}

// Start opens a new conversation; the customer connects with the returned
// session exactly like a guest who joined a room through an invite.
func (s *Service) Start(ctx context.Context, customerName, subject string) (StartResult, error) {
	name, err := room.CleanName(customerName, 32, "customer_name")
	if err != nil {
		return StartResult{}, err
	}
	if strings.TrimSpace(subject) == "" {
		subject = defaultSubject
	}
	subject, err = room.CleanName(subject, 80, "subject")
	if err != nil {
		return StartResult{}, err
	}

	rm := room.Room{ID: uuid.New(), Name: subject, OwnerTokenHash: auth.HashToken(auth.NewToken())}
	customer := room.Member{ID: uuid.New(), RoomID: rm.ID, DisplayName: name, Role: room.RoleMember}
	if err := s.repo.Start(ctx, rm, customer); err != nil {
		return StartResult{}, fmt.Errorf("start conversation: %w", err)
	}
	sess, err := s.issuer.Issue(auth.Claims{RoomID: rm.ID, MemberID: customer.ID, Role: room.RoleMember})
	if err != nil {
		return StartResult{}, fmt.Errorf("start conversation: %w", err)
	}
	return StartResult{Session: sess, Subject: subject}, nil
}

// Claim assigns a conversation to an agent (or refreshes their session if it
// is already theirs) and returns a normal room session with role "owner", so
// every existing chat/room endpoint works for the agent unchanged.
func (s *Service) Claim(ctx context.Context, roomID, agentID uuid.UUID, agentName string) (auth.Session, error) {
	agentMember := room.Member{ID: uuid.New(), RoomID: roomID, DisplayName: agentName, Role: room.RoleOwner}
	m, err := s.repo.Claim(ctx, roomID, agentID, agentMember)
	if err != nil {
		return auth.Session{}, fmt.Errorf("claim conversation: %w", err)
	}
	return s.issuer.Issue(auth.Claims{RoomID: roomID, MemberID: m.ID, Role: room.RoleOwner})
}

func (s *Service) ListOpen(ctx context.Context) ([]ConversationView, error) {
	views, err := s.repo.ListOpen(ctx)
	if err != nil {
		return nil, fmt.Errorf("list open conversations: %w", err)
	}
	return views, nil
}

func (s *Service) ListMine(ctx context.Context, agentID uuid.UUID) ([]ConversationView, error) {
	views, err := s.repo.ListMine(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("list my conversations: %w", err)
	}
	return views, nil
}
