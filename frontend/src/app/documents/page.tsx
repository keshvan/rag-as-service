const tableRows = [
  { file: "onboarding-guide.pdf", status: "processing", chunks: "-", updatedAt: "2 мин назад" },
  { file: "support-playbook-v2.docx", status: "indexed", chunks: "254", updatedAt: "18 мин назад" },
  { file: "api-faq.txt", status: "failed", chunks: "0", updatedAt: "45 мин назад" },
];

const statusClassMap: Record<string, string> = {
  processing: "bg-amber-500/10 text-amber-700",
  indexed: "bg-emerald-500/10 text-emerald-700",
  failed: "bg-rose-500/10 text-rose-700",
};

const statusLabelMap: Record<string, string> = {
  processing: "в обработке",
  indexed: "проиндексирован",
  failed: "ошибка",
};

export default function DocumentsPage() {
  return (
    <section className="space-y-4">
      <article className="rounded-2xl border border-border/70 bg-panel p-5 shadow-panel">
        <h2 className="text-lg font-semibold text-foreground">Очередь документов</h2>
        <p className="mt-1 text-sm text-muted">
          Здесь отображаются все загруженные файлы и этапы их обработки. Как только документ будет проиндексирован, он
          станет доступен в чате.
        </p>
      </article>

      <article className="overflow-hidden rounded-2xl border border-border/70 bg-panel shadow-panel">
        <table className="w-full border-collapse text-left text-sm">
          <thead>
            <tr className="bg-slate-950/[0.03] text-xs uppercase tracking-[0.12em] text-muted">
              <th className="px-4 py-3">Файл</th>
              <th className="px-4 py-3">Статус</th>
              <th className="px-4 py-3">Чанки</th>
              <th className="px-4 py-3">Обновлено</th>
            </tr>
          </thead>
          <tbody>
            {tableRows.map((row) => (
              <tr key={row.file} className="border-t border-border/60">
                <td className="px-4 py-3 font-medium text-foreground">{row.file}</td>
                <td className="px-4 py-3">
                  <span className={`rounded-full px-2.5 py-1 text-xs font-medium ${statusClassMap[row.status]}`}>
                    {statusLabelMap[row.status]}
                  </span>
                </td>
                <td className="px-4 py-3 text-muted">{row.chunks}</td>
                <td className="px-4 py-3 text-muted">{row.updatedAt}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </article>
    </section>
  );
}
