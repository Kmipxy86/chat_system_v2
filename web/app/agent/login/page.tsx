"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";

import { api } from "@/lib/api";
import { errorText } from "@/lib/errors";
import { saveAgentSession } from "@/lib/session";
import type { AgentSession } from "@/types/chat";

export default function AgentLoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      const sess = await api<AgentSession>("/agents/login", { method: "POST", body: { email, password } });
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
        <h1>เข้าสู่ระบบเจ้าหน้าที่</h1>
        <label>
          อีเมล
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoFocus
            placeholder="agent@example.com"
          />
        </label>
        <label>
          รหัสผ่าน
          <div className="input-with-action">
            <input
              type={showPassword ? "text" : "password"}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
            />
            <button type="button" className="link input-action" onClick={() => setShowPassword((v) => !v)}>
              {showPassword ? "ซ่อน" : "แสดง"}
            </button>
          </div>
        </label>
        <button className="primary wide" disabled={busy || !email.trim() || !password}>
          {busy ? "กำลังเข้าสู่ระบบ…" : "เข้าสู่ระบบ"}
        </button>
        {error && <p className="error">{error}</p>}
        <p className="muted small-text">
          ยังไม่มีบัญชี? <Link href="/agent/register">สมัครสมาชิกเจ้าหน้าที่</Link>
        </p>
      </form>
    </main>
  );
}
