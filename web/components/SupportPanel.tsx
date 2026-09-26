"use client";

interface Props {
  roomClosed: boolean;
  onClose: () => void;
}

export default function SupportPanel({ roomClosed, onClose }: Props) {
  if (roomClosed) {
    return (
      <section className="panel">
        <h2>การสนทนานี้จบแล้ว</h2>
        <p className="muted">ประวัติแชทยังเปิดอ่านได้</p>
      </section>
    );
  }

  return (
    <section className="panel">
      <h2>จัดการการสนทนา</h2>
      <button
        className="danger wide"
        onClick={() => {
          if (confirm("จบการสนทนานี้? ลูกค้าจะไม่สามารถส่งข้อความเพิ่มได้")) onClose();
        }}
      >
        จบการสนทนา
      </button>
    </section>
  );
}
