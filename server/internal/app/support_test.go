package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

type agentSession struct {
	AgentID string `json:"agent_id"`
	Name    string `json:"name"`
	JWT     string `json:"jwt"`
}

func (a *testAPI) registerAgent(name, email, password string) agentSession {
	a.t.Helper()
	body, _ := json.Marshal(map[string]string{"name": name, "email": email, "password": password})
	req, _ := http.NewRequestWithContext(context.Background(), "POST", a.srv.URL+"/agents/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Signup-Key", "test-signup-key")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		a.t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusCreated {
		a.t.Fatalf("register agent: status %d", res.StatusCode)
	}
	var s agentSession
	if err := json.NewDecoder(res.Body).Decode(&s); err != nil {
		a.t.Fatal(err)
	}
	return s
}

func TestSupportConversationFlow(t *testing.T) {
	a := newAPI(t)
	email := uuid.NewString() + "@example.com"
	ag := a.registerAgent("เอ", email, "hunter22222")
	if ag.AgentID == "" || ag.JWT == "" {
		t.Fatalf("register agent = %+v", ag)
	}

	// Registering the same email again must fail.
	{
		body, _ := json.Marshal(map[string]string{"name": "อีกคน", "email": email, "password": "hunter22222"})
		req, _ := http.NewRequestWithContext(context.Background(), "POST", a.srv.URL+"/agents/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Agent-Signup-Key", "test-signup-key")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		if res.StatusCode != http.StatusConflict {
			t.Fatalf("duplicate register status %d, want 409", res.StatusCode)
		}
	}

	var login agentSession
	if code := a.do("POST", "/agents/login", "", map[string]string{"email": email, "password": "hunter22222"}, &login); code != 200 {
		t.Fatalf("login status %d", code)
	}
	if login.AgentID != ag.AgentID {
		t.Fatalf("login agent id = %q, want %q", login.AgentID, ag.AgentID)
	}
	if code := a.do("POST", "/agents/login", "", map[string]string{"email": email, "password": "wrong"}, nil); code != 401 {
		t.Fatalf("bad password status %d", code)
	}

	var started struct {
		RoomID   string `json:"room_id"`
		MemberID string `json:"member_id"`
		JWT      string `json:"jwt"`
		Subject  string `json:"subject"`
	}
	if code := a.do("POST", "/support/conversations", "", map[string]string{"customer_name": "ลูกค้า A", "subject": "ปัญหาการชำระเงิน"}, &started); code != 201 {
		t.Fatalf("start conversation status %d", code)
	}
	if started.Subject != "ปัญหาการชำระเงิน" {
		t.Fatalf("subject = %q", started.Subject)
	}

	// Not yet claimed: appears in the open queue.
	var open struct {
		Conversations []struct {
			RoomID       string `json:"room_id"`
			CustomerName string `json:"customer_name"`
		} `json:"conversations"`
	}
	if code := a.do("GET", "/support/conversations", login.JWT, nil, &open); code != 200 {
		t.Fatalf("list open status %d", code)
	}
	found := false
	for _, c := range open.Conversations {
		if c.RoomID == started.RoomID {
			found = true
			if c.CustomerName != "ลูกค้า A" {
				t.Fatalf("customer name = %q", c.CustomerName)
			}
		}
	}
	if !found {
		t.Fatalf("started conversation not in open queue: %+v", open)
	}

	var claimed session
	if code := a.do("POST", "/support/conversations/"+started.RoomID+"/claim", login.JWT, nil, &claimed); code != 200 {
		t.Fatalf("claim status %d", code)
	}
	if claimed.RoomID != started.RoomID {
		t.Fatalf("claimed room = %q", claimed.RoomID)
	}

	// Once claimed it drops out of the open queue...
	a.do("GET", "/support/conversations", login.JWT, nil, &open)
	for _, c := range open.Conversations {
		if c.RoomID == started.RoomID {
			t.Fatalf("claimed conversation still open: %+v", open)
		}
	}
	// ...and shows up under "mine".
	var mine struct {
		Conversations []struct {
			RoomID string `json:"room_id"`
		} `json:"conversations"`
	}
	if code := a.do("GET", "/support/conversations?filter=mine", login.JWT, nil, &mine); code != 200 {
		t.Fatalf("list mine status %d", code)
	}
	if len(mine.Conversations) != 1 || mine.Conversations[0].RoomID != started.RoomID {
		t.Fatalf("mine = %+v", mine)
	}

	// Re-claiming (e.g. after the room JWT expired) is idempotent for the same agent.
	var reclaimed session
	if code := a.do("POST", "/support/conversations/"+started.RoomID+"/claim", login.JWT, nil, &reclaimed); code != 200 {
		t.Fatalf("reclaim status %d", code)
	}
	if reclaimed.MemberID != claimed.MemberID {
		t.Fatalf("reclaim minted a new member: %q vs %q", reclaimed.MemberID, claimed.MemberID)
	}

	// A second agent cannot steal an already-claimed conversation.
	ag2 := a.registerAgent("บี", uuid.NewString()+"@example.com", "hunter22222")
	if code := a.do("POST", "/support/conversations/"+started.RoomID+"/claim", ag2.JWT, nil, nil); code != 409 {
		t.Fatalf("steal claim status %d, want 409", code)
	}

	// The customer and the agent can now chat over the normal room/WS API.
	custDial := a.dial(session{RoomID: started.RoomID, MemberID: started.MemberID, JWT: started.JWT})
	agentDial := a.dial(claimed)
	custDial.next("presence")
	agentDial.next("presence")
	custDial.send(map[string]string{"type": "send", "client_msg_id": "11111111-1111-1111-1111-111111111111", "body": "สวัสดีค่ะ"})
	custDial.next("ack")
	if m := agentDial.next("message"); m["body"] != "สวัสดีค่ะ" {
		t.Fatalf("agent got %v", m)
	}
}

func TestAgentSignupRequiresKey(t *testing.T) {
	a := newAPI(t)
	body, _ := json.Marshal(map[string]string{"name": "x", "email": uuid.NewString() + "@example.com", "password": "hunter22222"})
	req, _ := http.NewRequestWithContext(context.Background(), "POST", a.srv.URL+"/agents/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("register without key status %d, want 403", res.StatusCode)
	}
}
