package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AgentSession is what an agent receives after registering or logging in.
// Unlike a room Session it is not scoped to any room.
type AgentSession struct {
	AgentID   uuid.UUID `json:"agent_id"`
	Name      string    `json:"name"`
	JWT       string    `json:"jwt"`
	ExpiresAt time.Time `json:"expires_at"`
}

type agentClaims struct {
	Typ string `json:"typ"`
	jwt.RegisteredClaims
}

// IssueAgent signs an account-scoped JWT, independent of any room.
func (i *Issuer) IssueAgent(agentID uuid.UUID, name string) (AgentSession, error) {
	now := i.now().UTC()
	exp := now.Add(i.ttl)
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, agentClaims{
		Typ: "agent",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   agentID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	})
	signed, err := tok.SignedString(i.secret)
	if err != nil {
		return AgentSession{}, fmt.Errorf("sign agent jwt: %w", err)
	}
	return AgentSession{AgentID: agentID, Name: name, JWT: signed, ExpiresAt: exp}, nil
}

// ParseAgent validates an account-scoped JWT and returns the agent id it was issued for.
func (i *Issuer) ParseAgent(token string) (uuid.UUID, error) {
	var ac agentClaims
	_, err := jwt.ParseWithClaims(token, &ac,
		func(*jwt.Token) (any, error) { return i.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(i.now),
	)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse agent jwt: %w", err)
	}
	if ac.Typ != "agent" {
		return uuid.UUID{}, fmt.Errorf("parse agent jwt: wrong token type")
	}
	agentID, err := uuid.Parse(ac.Subject)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse agent jwt subject: %w", err)
	}
	return agentID, nil
}
