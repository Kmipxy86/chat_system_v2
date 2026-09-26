"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";

import { api } from "@/lib/api";
import { errorText } from "@/lib/errors";
import { saveAgentSession } from "@/lib/session";
import type { AgentSession } from "@/types/chat";

export default function AgentRegisterPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [signupKey, setSignupKey] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      const sess = await api<AgentSession>("/agents/register", {
        method: "POST",
        body: { name, email, password },
        headers: { "X-Agent-Signup-Key": signupKey },
      });
      saveAgentSession(sess);
      router.push("/agent");
    } catch (err) {
      setError(errorText(err));
      setBusy(false);
    }
  };

  return (
    <main className="center">
      <form className="card" onSubmit={submit}>
        <h1>สมัครบัญชีเจ้าหน้าที่</h1>
        <p className="muted small-text">ต้องมีรหัสเชิญจากผู้ดูแลระบบจึงจะสมัครได้</p>
        <label>
          ชื่อ
          <input value={name} onChange={(e) => setName(e.target.value)} maxLength={80} required autoFocus />
        </label>
        <label>
          อีเมล
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        </label>
        <label>
          รหัสผ่าน
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
          />
        </label>
        <label>
          รหัสเชิญจากผู้ดูแลระบบ
          <input value={signupKey} onChange={(e) => setSignupKey(e.target.value)} required />
        </label>
        <button className="primary wide" disabled={busy || !name.trim() || !email.trim() || !password || !signupKey}>
          {busy ? "กำลังสมัคร…" : "สมัครสมาชิก"}
        </button>
        {error && <p className="error">{error}</p>}
        <p className="muted small-text">
          มีบัญชีแล้ว? <Link href="/agent/login">เข้าสู่ระบบ</Link>
        </p>
      </form>
    </main>
  );
}
