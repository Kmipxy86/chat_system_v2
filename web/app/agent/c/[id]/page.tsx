"use client";

import Link from "next/link";
import { use, useEffect, useState } from "react";

import ChatRoom from "@/components/ChatRoom";
import { api } from "@/lib/api";
import { isExpired, loadAgentSession, loadSession, saveSession } from "@/lib/session";
import type { Session } from "@/types/chat";

type State = { kind: "loading" } | { kind: "ready"; session: Session } | { kind: "missing" } | { kind: "expired" };

export default function AgentConversationPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const [state, setState] = useState<State>({ kind: "loading" });

  useEffect(() => {
    let cancelled = false;
    const resolve = async (): Promise<State> => {
      const s = loadSession(id);
      if (!s) return { kind: "missing" };
      if (!isExpired(s)) return { kind: "ready", session: s };
      // The room JWT expired; re-claim with the agent account session to get a fresh one.
      const agent = loadAgentSession();
      if (!agent || isExpired(agent)) return { kind: "expired" };
      try {
        const fresh = await api<Session>(`/support/conversations/${id}/claim`, { method: "POST", jwt: agent.jwt });
        saveSession(fresh);
        return { kind: "ready", session: fresh };
      } catch {
        return { kind: "expired" };
      }
    };
    resolve().then((st) => !cancelled && setState(st));
    return () => {
      cancelled = true;
    };
  }, [id]);

  if (state.kind === "ready") return <ChatRoom session={state.session} variant="agent" />;
  if (state.kind === "loading") return <main className="center muted">กำลังโหลด…</main>;
  return (
    <main className="center">
      <div className="card">
        <h1>{state.kind === "expired" ? "สิทธิ์เข้าถึงหมดอายุ" : "ยังไม่ได้รับเรื่องการสนทนานี้"}</h1>
        <Link href="/agent">กลับไปแดชบอร์ด</Link>
      </div>
    </main>
  );
}
