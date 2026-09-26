"use client";

import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";

import { useStored } from "@/hooks/useStored";
import { api } from "@/lib/api";
import { errorText } from "@/lib/errors";
import { loadDisplayName, saveDisplayName, saveSession } from "@/lib/session";
import type { StartedConversation } from "@/types/chat";

export default function SupportStartPage() {
  const router = useRouter();
  const savedName = useStored(loadDisplayName, "");
  const [typedName, setName] = useState<string | null>(null);
  const name = typedName ?? savedName;
  const [subject, setSubject] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      const started = await api<StartedConversation>("/support/conversations", {
        method: "POST",
        body: { customer_name: name, subject },
      });
      saveDisplayName(name.trim());
      saveSession({ ...started, room_name: started.subject, kind: "support" });
      router.push(`/r/${started.room_id}`);
    } catch (err) {
      setError(errorText(err));
      setBusy(false);
    }
  };

  return (
    <main className="center">
      <form className="card" onSubmit={submit}>
        <h1>ติดต่อฝ่ายสนับสนุน</h1>
        <p className="muted">แจ้งปัญหาหรือคำถาม แล้วเจ้าหน้าที่จะเข้ามาคุยกับคุณโดยตรง ไม่ต้องสมัครสมาชิก</p>
        <label>
          ชื่อของคุณ
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            maxLength={32}
            required
            autoFocus
            placeholder="เช่น สมชาย"
          />
        </label>
        <label>
          เรื่องที่ต้องการติดต่อ
          <input
            value={subject}
            onChange={(e) => setSubject(e.target.value)}
            maxLength={80}
            placeholder="เช่น ปัญหาการชำระเงิน"
          />
        </label>
        <button className="primary wide" disabled={busy || !name.trim()}>
          {busy ? "กำลังเริ่มการสนทนา…" : "เริ่มแชท"}
        </button>
        {error && <p className="error">{error}</p>}
      </form>
    </main>
  );
}
