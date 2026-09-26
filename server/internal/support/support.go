// Package support layers customer-support semantics on top of room/chat/invite:
// a customer opens a conversation (a plain room with no owner yet) and any
// logged-in agent can claim it, becoming that room's owner member. From then
// on every existing room/chat endpoint works unchanged for both sides.
package support

import (
	"time"

	"github.com/google/uuid"
)

type ConversationView struct {
	RoomID       uuid.UUID `json:"room_id"`
	Subject      string    `json:"subject"`
	Status       string    `json:"status"`
	CustomerName string    `json:"customer_name"`
	CreatedAt    time.Time `json:"created_at"`
}
