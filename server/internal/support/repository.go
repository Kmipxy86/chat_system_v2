package support

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"qrchat/internal/apperr"
	"qrchat/internal/room"
)

type Repository struct {
	db    *pgxpool.Pool
	rooms *room.Repository
}

func NewRepository(db *pgxpool.Pool, rooms *room.Repository) *Repository {
	return &Repository{db: db, rooms: rooms}
}

// Start creates a new conversation room with the customer as its first member.
func (r *Repository) Start(ctx context.Context, rm room.Room, customer room.Member) error {
	if err := r.rooms.CreateWithOwner(ctx, rm, customer); err != nil {
		return fmt.Errorf("start conversation: %w", err)
	}
	return nil
}

// Claim assigns roomID to agentID and inserts agentMember as its owner-role
// member, unless the conversation is already assigned to that same agent (the
// existing member is returned, so an expired JWT can be refreshed by claiming
// again) or to a different agent (apperr.ConversationClaimed).
func (r *Repository) Claim(ctx context.Context, roomID, agentID uuid.UUID, agentMember room.Member) (room.Member, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return room.Member{}, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	var assigned *uuid.UUID
	err = tx.QueryRow(ctx, `SELECT status, assigned_agent_id FROM rooms WHERE id = $1 FOR UPDATE`, roomID).
		Scan(&status, &assigned)
	if errors.Is(err, pgx.ErrNoRows) {
		return room.Member{}, apperr.NotFound
	}
	if err != nil {
		return room.Member{}, fmt.Errorf("lock conversation: %w", err)
	}
	if status == room.StatusClosed {
		return room.Member{}, apperr.RoomClosed
	}
	if assigned != nil {
		if *assigned != agentID {
			return room.Member{}, apperr.ConversationClaimed
		}
		m, err := r.rooms.OwnerMember(ctx, roomID)
		if err != nil {
			return room.Member{}, fmt.Errorf("reload claim: %w", err)
		}
		return m, nil
	}

	agentMember.RoomID = roomID
	if err := room.InsertMember(ctx, tx, agentMember); err != nil {
		return room.Member{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE rooms SET assigned_agent_id = $2 WHERE id = $1`, roomID, agentID); err != nil {
		return room.Member{}, fmt.Errorf("assign conversation: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return room.Member{}, fmt.Errorf("commit: %w", err)
	}
	return agentMember, nil
}

const conversationSelect = `
	SELECT r.id, r.name, r.status, r.created_at, m.display_name
	FROM rooms r
	JOIN room_members m ON m.room_id = r.id AND m.role = '` + room.RoleMember + `'
`

func scanConversation(row pgx.Row) (ConversationView, error) {
	var v ConversationView
	err := row.Scan(&v.RoomID, &v.Subject, &v.Status, &v.CreatedAt, &v.CustomerName)
	v.CreatedAt = v.CreatedAt.UTC()
	return v, err
}

// ListOpen returns active conversations no agent has claimed yet, oldest first
// so the longest-waiting customer surfaces at the top of the queue.
func (r *Repository) ListOpen(ctx context.Context) ([]ConversationView, error) {
	rows, err := r.db.Query(ctx,
		conversationSelect+`WHERE r.assigned_agent_id IS NULL AND r.status = $1 ORDER BY r.created_at ASC`,
		room.StatusActive)
	if err != nil {
		return nil, fmt.Errorf("list open conversations: %w", err)
	}
	views, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ConversationView, error) { return scanConversation(row) })
	if err != nil {
		return nil, fmt.Errorf("scan open conversations: %w", err)
	}
	return views, nil
}

// ListMine returns every conversation (active or closed) assigned to agentID,
// most recent first.
func (r *Repository) ListMine(ctx context.Context, agentID uuid.UUID) ([]ConversationView, error) {
	rows, err := r.db.Query(ctx,
		conversationSelect+`WHERE r.assigned_agent_id = $1 ORDER BY r.created_at DESC`, agentID)
	if err != nil {
		return nil, fmt.Errorf("list my conversations: %w", err)
	}
	views, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ConversationView, error) { return scanConversation(row) })
	if err != nil {
		return nil, fmt.Errorf("scan my conversations: %w", err)
	}
	return views, nil
}
