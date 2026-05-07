"use client";

import { useCallback, useEffect, useState } from "react";
import { ApiError } from "@/components/api/client";
import {
  listDocuments,
  type DocumentQueueItem,
  type DocumentQueueStatus,
} from "@/features/documents/api/list";
import { DocumentUploadPanel } from "@/features/documents/components/document-upload-panel";

const statusClassMap: Record<DocumentQueueStatus, string> = {
  pending_upload: "bg-slate-500/10 text-slate-700",
  uploaded: "bg-cyan-500/10 text-cyan-700",
  processing: "bg-amber-500/10 text-amber-700",
  indexed: "bg-emerald-500/10 text-emerald-700",
  failed: "bg-rose-500/10 text-rose-700",
};

const statusLabelMap: Record<DocumentQueueStatus, string> = {
  pending_upload: "ожидает загрузки",
  uploaded: "загружен",
  processing: "в обработке",
  indexed: "проиндексирован",
  failed: "ошибка",
};

function isKnownStatus(status: string): status is DocumentQueueStatus {
  return Object.prototype.hasOwnProperty.call(statusClassMap, status);
}

function resolveStatusClass(status: string): string {
  if (isKnownStatus(status)) {
    return statusClassMap[status];
  }

  return "bg-slate-500/10 text-slate-700";
}

function resolveStatusLabel(status: string): string {
  if (isKnownStatus(status)) {
    return statusLabelMap[status];
  }

  return status;
}

function formatCreatedAt(createdAt: string): string {
  const parsedDate = new Date(createdAt);

  if (Number.isNaN(parsedDate.getTime())) {
    return "—";
  }

  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "short",
    timeStyle: "short",
  }).format(parsedDate);
}

function toQueueErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return `Не удалось загрузить список документов (${error.status}): ${error.message}`;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return "Не удалось загрузить список документов.";
}

export default function DocumentsPage() {
  const [documents, setDocuments] = useState<DocumentQueueItem[]>([]);
  const [isInitialLoading, setIsInitialLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [queueError, setQueueError] = useState<string | null>(null);

  const fetchQueue = useCallback(async (silent = false) => {
    if (silent) {
      setIsRefreshing(true);
    } else {
      setIsInitialLoading(true);
    }

    setQueueError(null);

    try {
      const items = await listDocuments({ limit: 50, offset: 0 });
      setDocuments(items);
    } catch (error) {
      setQueueError(toQueueErrorMessage(error));
    } finally {
      if (silent) {
        setIsRefreshing(false);
      } else {
        setIsInitialLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    void fetchQueue();
  }, [fetchQueue]);

  const refreshQueue = () => {
    void fetchQueue(true);
  };

  return (
    <section className="space-y-4">
      <DocumentUploadPanel onUploadCompleted={refreshQueue} />

      <article className="rounded-2xl border border-border/70 bg-panel p-5 shadow-panel">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-lg font-semibold text-foreground">Очередь документов</h2>
          <button
            type="button"
            onClick={refreshQueue}
            disabled={isInitialLoading || isRefreshing}
            className="rounded-xl border border-border/70 bg-page/80 px-3 py-1.5 text-xs font-medium text-foreground transition hover:border-cyan-300 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {isRefreshing ? "Обновляем..." : "Обновить"}
          </button>
        </div>
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
              <th className="px-4 py-3">Создан</th>
            </tr>
          </thead>
          <tbody>
            {isInitialLoading && (
              <tr className="border-t border-border/60">
                <td colSpan={4} className="px-4 py-6 text-center text-sm text-muted">
                  Загружаем очередь документов...
                </td>
              </tr>
            )}

            {!isInitialLoading && queueError && (
              <tr className="border-t border-border/60">
                <td colSpan={4} className="px-4 py-6 text-center text-sm text-rose-700">
                  {queueError}
                </td>
              </tr>
            )}

            {!isInitialLoading && !queueError && documents.length === 0 && (
              <tr className="border-t border-border/60">
                <td colSpan={4} className="px-4 py-6 text-center text-sm text-muted">
                  В очереди пока нет документов.
                </td>
              </tr>
            )}

            {!isInitialLoading &&
              !queueError &&
              documents.map((document) => (
                <tr key={document.id} className="border-t border-border/60">
                  <td className="px-4 py-3 font-medium text-foreground">{document.file_name}</td>
                  <td className="px-4 py-3">
                    <span className={`rounded-full px-2.5 py-1 text-xs font-medium ${resolveStatusClass(document.status)}`}>
                      {resolveStatusLabel(document.status)}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-muted">—</td>
                  <td className="px-4 py-3 text-muted">{formatCreatedAt(document.created_at)}</td>
                </tr>
              ))}
          </tbody>
        </table>
      </article>
    </section>
  );
}
