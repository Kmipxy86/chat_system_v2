"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { useStored } from "@/hooks/useStored";
import { api } from "@/lib/api";
import { errorText } from "@/lib/errors";
import { clearAgentSession, isExpired, loadAgentSession, saveSession } from "@/lib/session";
import type { AgentSession, ConversationView, Session } from "@/types/chat";

const dateFmt = new Intl.DateTimeFormat("th-TH", { dateStyle: "short", timeStyle: "short" });

// undefined: not hydrated yet; null: confirmed logged out.
const notHydrated = undefined as AgentSession | null | undefined;

export default function AgentDashboardPage() {
  const router = useRouter();
  const agent = useStored(loadAgentSession, notHydrated);
  const loggedOut = agent !== undefined && (agent === null || isExpired(agent));

  const [open, setOpen] = useState<ConversationView[]>([]);
  const [mine, setMine] = useState<ConversationView[]>([]);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (loggedOut) router.replace("/agent/login");
  }, [loggedOut, router]);

  useEffect(() => {
    if (!agent || loggedOut) return;
    let cancelled = false;
    const tick = () => {
      Promise.all([
        api<{ conversations: ConversationView[] }>("/support/conversations", { jwt: agent.jwt }),
        api<{ conversations: ConversationView[] }>("/support/conversations?filter=mine", { jwt: agent.jwt }),
      ])
        .then(([o, m]) => {
          if (cancelled) return;
          setOpen(o.conversations);
          setMine(m.conversations.filter((c) => c.status === "active"));
        })
        .catch((e) => {
          if (cancelled) return;
          setError(errorText(e));
        });
    };
    tick();
    const t = setInterval(tick, 8000);
    return () => {
      cancelled = true;
      clearInterval(t);
    };
  }, [agent, loggedOut]);

  const claim = async (roomId: string) => {
    if (!agent) return;
    setBusyId(roomId);
    setError("");
    try {
      const sess = await api<Session>(`/support/conversations/${roomId}/claim`, { method: "POST", jwt: agent.jwt });
      saveSession(sess);
      router.push(`/agent/c/${roomId}`);
    } catch (e) {
      setError(errorText(e));
      setBusyId(null);
    }
  };

  const logout = () => {
    clearAgentSession();
    router.replace("/agent/login");
  };

  if (!agent || loggedOut) return <main className="center muted">กำลังโหลด…</main>;

  return (
    <main className="dashboard">
      <header className="dashboard-header">
        <h1>แดชบอร์ดเจ้าหน้าที่</h1>
        <div className="dashboard-who">
          <span className="muted">{agent.name}</span>
          <button className="small" onClick={logout}>
            ออกจากระบบ
          </button>
        </div>
      </header>

      {error && <p className="error">{error}</p>}

      <section className="panel">
        <h2>กำลังสนทนาอยู่ ({mine.length})</h2>
        {mine.length === 0 && <p className="muted">ยังไม่มีการสนทนาที่คุณรับเรื่อง</p>}
        <ul className="conversations">
          {mine.map((c) => (
            <li key={c.room_id}>
              <Link href={`/agent/c/${c.room_id}`}>
                <strong>{c.customer_name}</strong> · {c.subject}
              </Link>
              <span className="muted small-text">{dateFmt.format(new Date(c.created_at))}</span>
            </li>
          ))}
        </ul>
      </section>

      <section className="panel">
        <h2>รอเจ้าหน้าที่รับเรื่อง ({open.length})</h2>
        {open.length === 0 && <p className="muted">ยังไม่มีลูกค้ารอสาย</p>}
        <ul className="conversations">
          {open.map((c) => (
            <li key={c.room_id}>
              <span>
                <strong>{c.customer_name}</strong> · {c.subject}
                <br />
                <span className="muted small-text">{dateFmt.format(new Date(c.created_at))}</span>
              </span>
              <button className="primary small" disabled={busyId === c.room_id} onClick={() => claim(c.room_id)}>
                {busyId === c.room_id ? "กำลังรับเรื่อง…" : "รับเรื่อง"}
              </button>
            </li>
          ))}
        </ul>
      </section>
    </main>
  );
}
