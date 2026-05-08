import { AuthFormShell } from "@/features/auth/components/auth-form-shell";
import { LoginForm } from "@/features/auth/components/login-form";

export default function LoginPage() {
  return (
    <AuthFormShell
      title="Вход"
      footerText="Ещё нет аккаунта?"
      footerHref="/register"
      footerLabel="Зарегистрироваться"
    >
      <LoginForm />
    </AuthFormShell>
  );
}
