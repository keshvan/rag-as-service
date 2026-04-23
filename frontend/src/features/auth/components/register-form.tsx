"use client";

import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/features/auth/providers/auth-provider";

export function RegisterForm() {
  const { register } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [organizationName, setOrganizationName] = useState("");
  const [organizationUrl, setOrganizationUrl] = useState("");
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setMessage(null);
    setIsSubmitting(true);

    try {
      await register({
        email,
        password,
        organization_name: organizationName,
        organization_url: organizationUrl,
      });
      setMessage("Регистрация отправлена. Введите код подтверждения, который пришёл на почту.");
      router.push(`/verification?email=${encodeURIComponent(email)}`);
    } catch (submissionError) {
      const nextError = submissionError instanceof Error ? submissionError.message : "Не удалось зарегистрироваться";
      setError(nextError);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form className="space-y-4" onSubmit={onSubmit}>
      <div>
        <label htmlFor="org-name" className="mb-2 block text-sm font-medium text-foreground">
          Название организации
        </label>
        <input
          id="org-name"
          value={organizationName}
          onChange={(event) => setOrganizationName(event.target.value)}
          className="w-full rounded-2xl border border-border/70 bg-page/70 px-4 py-3 outline-none transition focus:border-accent"
          placeholder="Acme Corp"
          required
        />
      </div>

      <div>
        <label htmlFor="org-url" className="mb-2 block text-sm font-medium text-foreground">
          URL организации
        </label>
        <input
          id="org-url"
          value={organizationUrl}
          onChange={(event) => setOrganizationUrl(event.target.value)}
          className="w-full rounded-2xl border border-border/70 bg-page/70 px-4 py-3 outline-none transition focus:border-accent"
          placeholder="https://acme.example"
          required
        />
      </div>

      <div>
        <label htmlFor="register-email" className="mb-2 block text-sm font-medium text-foreground">
          Email
        </label>
        <input
          id="register-email"
          type="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="w-full rounded-2xl border border-border/70 bg-page/70 px-4 py-3 outline-none transition focus:border-accent"
          placeholder="team@company.com"
          required
        />
      </div>

      <div>
        <label htmlFor="register-password" className="mb-2 block text-sm font-medium text-foreground">
          Пароль
        </label>
        <input
          id="register-password"
          type="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          className="w-full rounded-2xl border border-border/70 bg-page/70 px-4 py-3 outline-none transition focus:border-accent"
          placeholder="Придумайте пароль"
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
        {isSubmitting ? "Создаём аккаунт..." : "Зарегистрироваться"}
      </button>
    </form>
  );
}
