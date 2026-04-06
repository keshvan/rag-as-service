const mockSources = [
  {
    document: "onboarding-guide.pdf",
    snippet: "Чтобы файл появился в чате, дождитесь статуса «проиндексирован» в разделе «Документы».",
  },
  {
    document: "support-playbook-v2.docx",
    snippet: "Указывайте в вопросе конкретный процесс или термин, чтобы ответ был точнее.",
  },
];

export default function ChatPage() {
  return (
    <section className="grid gap-4 lg:grid-cols-[minmax(0,2fr)_minmax(280px,1fr)]">
      <article className="rounded-2xl border border-border/70 bg-panel p-5 shadow-panel">
        <h2 className="text-lg font-semibold text-foreground">RAG-чат</h2>
        <p className="mt-1 text-sm text-muted">Задавайте вопросы по загруженным документам и получайте ответ с источниками.</p>

        <div className="mt-5 rounded-xl border border-border/70 bg-page/70 p-4">
          <p className="text-sm font-medium text-foreground">Вопрос</p>
          <p className="mt-1 text-sm text-muted">Как добавить новый регламент в базу знаний и проверить, что он доступен в чате?</p>
        </div>

        <div className="mt-4 rounded-xl border border-border/70 bg-cyan-500/5 p-4">
          <p className="text-sm font-medium text-foreground">Ответ</p>
          <p className="mt-1 text-sm leading-6 text-muted">
            Загрузите документ в разделе «Документы», дождитесь статуса «проиндексирован», затем вернитесь в чат и
            задайте вопрос по ключевым терминам из файла. В ответе появятся ссылки на источники, чтобы проверить
            точность.
          </p>
        </div>
      </article>

      <article className="rounded-2xl border border-border/70 bg-panel p-5 shadow-panel">
        <h2 className="text-base font-semibold text-foreground">Источники</h2>
        <ul className="mt-3 space-y-3">
          {mockSources.map((source) => (
            <li key={source.document} className="rounded-xl border border-border/70 p-3">
              <p className="text-xs uppercase tracking-[0.12em] text-muted">{source.document}</p>
              <p className="mt-1 text-sm text-foreground">{source.snippet}</p>
            </li>
          ))}
        </ul>
      </article>
    </section>
  );
}
