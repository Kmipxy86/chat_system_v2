"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useMemo, useState, type FormEvent } from "react";

import { api } from "@/lib/api";
import { errorText } from "@/lib/errors";
import { saveAgentSession } from "@/lib/session";
import type { AgentSession } from "@/types/chat";

function validateEmail(email: string): string {
  if (email === "") return "";
  return /^[^\s@]+@[^\s@]+$/.test(email) ? "" : "รูปแบบอีเมลไม่ถูกต้อง";
}

function validatePassword(password: string): string {
  if (password === "") return "";
  return password.length >= 8 ? "" : "รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร";
}

export default function AgentRegisterPage() {
  return (
    <Suspense fallback={<main className="center muted">กำลังโหลด…</main>}>
      <RegisterForm />
    </Suspense>
  );
}

function RegisterForm() {
  const router = useRouter();
  // A link like /agent/register?key=... skips retyping the invite key by hand.
  const keyFromLink = useSearchParams().get("key") ?? "";

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [signupKey, setSignupKey] = useState(keyFromLink);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const emailError = useMemo(() => validateEmail(email), [email]);
  const passwordError = useMemo(() => validatePassword(password), [password]);
  const canSubmit =
    name.trim() !== "" &&
    email.trim() !== "" &&
    !emailError &&
    !passwordError &&
    password !== "" &&
    signupKey.trim() !== "";

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
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
        {!keyFromLink && <p className="muted small-text">ต้องมีรหัสเชิญจากผู้ดูแลระบบจึงจะสมัครได้</p>}
        <label>
          ชื่อ
          <input value={name} onChange={(e) => setName(e.target.value)} maxLength={80} required autoFocus />
        </label>
        <label>
          อีเมล
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        </label>
        {emailError && <p className="error field-hint">{emailError}</p>}
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
        {passwordError && <p className="error field-hint">{passwordError}</p>}
        {keyFromLink ? (
          <input type="hidden" value={signupKey} readOnly />
        ) : (
          <label>
            รหัสเชิญจากผู้ดูแลระบบ
            <input value={signupKey} onChange={(e) => setSignupKey(e.target.value)} required />
          </label>
        )}
        <button className="primary wide" disabled={busy || !canSubmit}>
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
