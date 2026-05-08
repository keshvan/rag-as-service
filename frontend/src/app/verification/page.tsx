import { AuthFormShell } from "@/features/auth/components/auth-form-shell";
import { VerificationForm } from "@/features/auth/components/verification-form";

export default function VerificationPage() {
  return (
    <AuthFormShell
      title="Подтверждение email"
      footerText="Уже подтвердили почту?"
      footerHref="/login"
      footerLabel="Войти"
    >
      <VerificationForm />
    </AuthFormShell>
  );
}
