import { Activity, CircleAlert, Database, FileClock } from "lucide-react";
import { DashboardCard } from "@/components/ui/dashboard-card";

const recentActivity = [
  { event: "Загружен файл onboarding-guide.pdf", time: "2 мин назад", status: "в обработке" },
  { event: "Проиндексирован policy-v4.docx", time: "18 мин назад", status: "готово" },
  { event: "Алерт по задержке запроса устранен", time: "31 мин назад", status: "решено" },
];

export default function DashboardPage() {
  return (
    <section className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <DashboardCard title="Документов в хранилище" value="1,284" trend="+6.2% за неделю">
          <div className="flex items-center gap-2 text-xs text-muted">
            <Database className="h-3.5 w-3.5" />
            248 ожидают индексации
          </div>
        </DashboardCard>
        <DashboardCard title="Пропускная способность" value="84 / час" trend="Пайплайн стабилен">
          <div className="flex items-center gap-2 text-xs text-muted">
            <FileClock className="h-3.5 w-3.5" />
            Среднее время чанкинга 1.8 c
          </div>
        </DashboardCard>
        <DashboardCard title="RAG-запросов за сегодня" value="392" trend="P95 задержка 1.2 c">
          <div className="flex items-center gap-2 text-xs text-muted">
            <Activity className="h-3.5 w-3.5" />
            99.1% успешных ответов
          </div>
        </DashboardCard>
        <DashboardCard title="Ошибок индексации" value="12" trend="-3 за последние сутки">
          <div className="flex items-center gap-2 text-xs text-muted">
            <CircleAlert className="h-3.5 w-3.5" />
            Требуют повторной обработки
          </div>
        </DashboardCard>
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <article className="rounded-2xl border border-border/70 bg-panel p-5 shadow-panel lg:col-span-2">
          <p className="text-xs uppercase tracking-[0.16em] text-muted">Качество ответов</p>
          <h2 className="mt-2 text-xl font-semibold text-foreground">Точность ответов выросла на 9%</h2>
          <p className="mt-2 text-sm leading-6 text-muted">
            После последнего обновления система лучше подбирает фрагменты в ответах и реже возвращает лишний
            контекст. В первую очередь стоит проверить документы со сложной версткой и таблицами.
          </p>
          <div className="mt-5 flex flex-wrap gap-2">
            <span className="rounded-full bg-cyan-500/10 px-3 py-1 text-xs font-medium text-cyan-700">
              Точность поиска: 82%
            </span>
            <span className="rounded-full bg-emerald-500/10 px-3 py-1 text-xs font-medium text-emerald-700">
              Релевантные источники: 94%
            </span>
            <span className="rounded-full bg-amber-500/10 px-3 py-1 text-xs font-medium text-amber-700">
              Повторных запусков задач: 12
            </span>
          </div>
        </article>

        <article className="rounded-2xl border border-border/70 bg-panel p-5 shadow-panel">
          <p className="text-xs uppercase tracking-[0.16em] text-muted">Последняя активность</p>
          <ul className="mt-4 space-y-3">
            {recentActivity.map((item) => (
              <li key={item.event} className="rounded-xl border border-border/70 p-3">
                <p className="text-sm font-medium text-foreground">{item.event}</p>
                <div className="mt-1 flex items-center justify-between text-xs text-muted">
                  <span>{item.time}</span>
                  <span className="rounded bg-slate-950/5 px-2 py-0.5 uppercase tracking-wide">{item.status}</span>
                </div>
              </li>
            ))}
          </ul>
        </article>
      </div>
    </section>
  );
}
