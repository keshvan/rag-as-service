import { AuthFormShell } from "@/features/auth/components/auth-form-shell";
import { RegisterForm } from "@/features/auth/components/register-form";

export default function RegisterPage() {
  return (
    <AuthFormShell
      title="Регистрация"
      footerText="Уже есть аккаунт?"
      footerHref="/login"
      footerLabel="Войти"
    >
      <RegisterForm />
    </AuthFormShell>
  );
}
