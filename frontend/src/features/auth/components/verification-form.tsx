"use client";

import { useEffect, useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "@/features/auth/providers/auth-provider";

export function VerificationForm() {
  const { verifyEmail } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    const nextEmail = new URLSearchParams(window.location.search).get("email") ?? "";
    setEmail(nextEmail);
  }, []);

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setMessage(null);
    setIsSubmitting(true);

    try {
      await verifyEmail(email, code);
      setMessage("Email подтверждён. Теперь можно войти в аккаунт.");
      router.push(`/login?email=${encodeURIComponent(email)}`);
    } catch (submissionError) {
      const nextError = submissionError instanceof Error ? submissionError.message : "Не удалось подтвердить email";
      setError(nextError);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form className="space-y-4" onSubmit={onSubmit}>
      <div>
        <label htmlFor="verification-email" className="mb-2 block text-sm font-medium text-foreground">
          Email
        </label>
        <input
          id="verification-email"
          type="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="w-full rounded-2xl border border-border/70 bg-page/70 px-4 py-3 outline-none transition focus:border-accent"
          placeholder="team@company.com"
          required
        />
      </div>

      <div>
        <label htmlFor="verification-code" className="mb-2 block text-sm font-medium text-foreground">
          Код подтверждения
        </label>
        <input
          id="verification-code"
          value={code}
          onChange={(event) => setCode(event.target.value)}
          className="w-full rounded-2xl border border-border/70 bg-page/70 px-4 py-3 outline-none transition focus:border-accent"
          placeholder="Введите код из письма"
          required
        />
      </div>

      {error ? <p className="rounded-2xl bg-rose-500/10 px-4 py-3 text-sm text-rose-700">{error}</p> : null}
      {message ? <p className="rounded-2xl bg-emerald-500/10 px-4 py-3 text-sm text-emerald-700">{message}</p> : null}

      <button
        type="submit"
        disabled={isSubmitting}
        className="w-full rounded-2xl bg-accent px-4 py-3 text-sm font-medium text-white transition hover:bg-accent-strong disabled:cursor-not-allowed disabled:opacity-70"
      >
        {isSubmitting ? "Подтверждаем..." : "Подтвердить email"}
      </button>

      <p className="text-sm text-muted">
        Код не пришёл? Сначала завершите регистрацию на{" "}
        <Link href="/register" className="font-medium text-accent-strong transition-colors hover:text-foreground">
          странице создания аккаунта
        </Link>
        .
      </p>
    </form>
  );
}
