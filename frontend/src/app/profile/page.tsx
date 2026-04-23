"use client";

import { useAuth } from "@/features/auth/providers/auth-provider";

export default function ProfilePage() {
  const { user, logout } = useAuth();

  return (
    <section className="space-y-4">
      <article className="rounded-2xl border border-border/70 bg-panel p-6 shadow-panel">
        <p className="text-xs uppercase tracking-[0.16em] text-muted">Текущий пользователь</p>
        <h2 className="mt-2 text-2xl font-semibold text-foreground">Профиль</h2>
        <p className="mt-3 text-sm leading-6 text-muted">
          Auth service сейчас отдаёт на фронт идентификатор пользователя. Этого достаточно, чтобы привязать доступ,
          переходы и защищённые роуты.
        </p>

        <div className="mt-6 grid gap-4 md:grid-cols-2">
          <div className="rounded-2xl bg-page/70 p-4">
            <p className="text-xs uppercase tracking-[0.12em] text-muted">GUID</p>
            <p className="mt-2 break-all text-sm font-medium text-foreground">{user?.guid ?? "Сессия не найдена"}</p>
          </div>

          <div className="rounded-2xl bg-page/70 p-4">
            <p className="text-xs uppercase tracking-[0.12em] text-muted">Статус</p>
            <p className="mt-2 text-sm font-medium text-foreground">{user ? "Авторизован" : "Не авторизован"}</p>
          </div>
        </div>

        <button
          type="button"
          onClick={() => void logout()}
          className="mt-6 rounded-2xl border border-border/70 bg-white px-4 py-3 text-sm font-medium text-foreground transition hover:border-rose-300 hover:text-rose-700"
        >
          Выйти из аккаунта
        </button>
      </article>
    </section>
  );
}
