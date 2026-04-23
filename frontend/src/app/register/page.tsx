import { AuthFormShell } from "@/features/auth/components/auth-form-shell";
import { RegisterForm } from "@/features/auth/components/register-form";

export default function RegisterPage() {
  return (
    <AuthFormShell
      eyebrow="Workspace"
      title="Создайте организацию и аккаунт"
      subtitle="После регистрации auth service отправит подтверждение, а после входа фронт откроет защищённые разделы и профиль пользователя."
      footerText="Уже есть аккаунт?"
      footerHref="/login"
      footerLabel="Войти"
    >
      <RegisterForm />
    </AuthFormShell>
  );
}
