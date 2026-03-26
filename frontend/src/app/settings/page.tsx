const settingsSections = [
  {
    title: "Рабочее пространство",
    description: "Вы работаете в пространстве компании Acme. Здесь будут общие параметры команды и прав доступа.",
    value: "Acme",
  },
  {
    title: "Уведомления",
    description: "Включите уведомления о завершении индексации, ошибках обработки и новых ответах в чате.",
    value: "Email и в приложении",
  },
  {
    title: "Безопасность",
    description: "Управляйте активными сессиями и подтвержденными устройствами для безопасного доступа к данным.",
    value: "2 активные сессии",
  },
];

export default function SettingsPage() {
  return (
    <section className="space-y-4">
      {settingsSections.map((section) => (
        <article key={section.title} className="rounded-2xl border border-border/70 bg-panel p-5 shadow-panel">
          <p className="text-xs uppercase tracking-[0.16em] text-muted">{section.title}</p>
          <p className="mt-2 text-sm leading-6 text-foreground">{section.description}</p>
          <p className="mt-2 inline-flex rounded-full bg-slate-950/5 px-3 py-1 text-xs font-medium text-muted">
            {section.value}
          </p>
        </article>
      ))}
    </section>
  );
}
