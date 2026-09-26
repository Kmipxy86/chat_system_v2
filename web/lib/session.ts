import type { AgentSession, CreatedInvite, Session } from "@/types/chat";

// Sessions live in localStorage so a refresh or a closed tab can come back.
// Every access is guarded: storage can be unavailable (private mode, blocked).

const sessionKey = (roomId: string) => `qrchat:session:${roomId}`;
const inviteKey = (roomId: string) => `qrchat:invite:${roomId}`;
const NAME_KEY = "qrchat:display_name";
const AGENT_SESSION_KEY = "qrchat:agent_session";

function read<T>(key: string): T | null {
  try {
    const raw = localStorage.getItem(key);
    return raw ? (JSON.parse(raw) as T) : null;
  } catch {
    return null;
  }
}

function write(key: string, value: unknown) {
  try {
    if (value === null) localStorage.removeItem(key);
    else localStorage.setItem(key, JSON.stringify(value));
  } catch {
    // storage unavailable: the session only lasts for this page
  }
}

export const loadSession = (roomId: string) => read<Session>(sessionKey(roomId));
export const saveSession = (s: Session) => write(sessionKey(s.room_id), s);

export function isExpired(s: { expires_at: string }): boolean {
  return Date.parse(s.expires_at) - 60_000 < Date.now();
}

/** The logged-in support agent, if any; global (not per-room), unlike room sessions. */
export const loadAgentSession = () => read<AgentSession>(AGENT_SESSION_KEY);
export const saveAgentSession = (s: AgentSession) => write(AGENT_SESSION_KEY, s);
export const clearAgentSession = () => write(AGENT_SESSION_KEY, null);

/** Rooms this browser hosts, newest first. */
export function hostedRooms(): Session[] {
  const out: Session[] = [];
  try {
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (!key?.startsWith("qrchat:session:")) continue;
      const s = read<Session>(key);
      if (s?.owner_token) out.push(s);
    }
  } catch {
    return [];
  }
  return out.reverse();
}

export const loadInvite = (roomId: string) => read<CreatedInvite>(inviteKey(roomId));
export const saveInvite = (roomId: string, inv: CreatedInvite | null) => write(inviteKey(roomId), inv);

export const loadDisplayName = () => read<string>(NAME_KEY) ?? "";
export const saveDisplayName = (name: string) => write(NAME_KEY, name);

/** Absolute join link; the server returns a path when APP_BASE_URL is unset. */
export function inviteLink(inv: CreatedInvite): string {
  return inv.invite_url.startsWith("/") ? window.location.origin + inv.invite_url : inv.invite_url;
}
