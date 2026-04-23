import { AuthFormShell } from "@/features/auth/components/auth-form-shell";
import { VerificationForm } from "@/features/auth/components/verification-form";

export default function VerificationPage() {
  return (
    <AuthFormShell
      eyebrow="Verification"
      title="Подтвердите email кодом из письма"
      subtitle="После подтверждения учётная запись станет доступна для входа, а защищённые страницы откроются через обычный login flow."
      footerText="Уже подтвердили почту?"
      footerHref="/login"
      footerLabel="Войти"
    >
      <VerificationForm />
    </AuthFormShell>
  );
}
