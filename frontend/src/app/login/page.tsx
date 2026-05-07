import { AuthFormShell } from "@/features/auth/components/auth-form-shell";
import { LoginForm } from "@/features/auth/components/login-form";

export default function LoginPage() {
  return (
    <AuthFormShell
      eyebrow="Access"
      title="Вход в рабочее пространство"
      subtitle="Авторизуйтесь, чтобы открыть документы, чат и страницу профиля. Неавторизованным пользователям доступны только вход и регистрация."
      footerText="Ещё нет аккаунта?"
      footerHref="/register"
      footerLabel="Зарегистрироваться"
    >
      <LoginForm />
    </AuthFormShell>
  );
}
