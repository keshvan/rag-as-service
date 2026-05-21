"use client";

import {
  Bot,
  Loader2,
  MessageSquareQuote,
  Send,
  SlidersHorizontal,
  Trash2,
  User,
} from "lucide-react";
import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type FormEvent,
  type KeyboardEvent,
} from "react";
import { ApiError } from "@/components/api/client";
import { queryRag, type RagSource } from "@/features/chat/api/query";
import { cn } from "@/lib/cn";

type ChatMessage = {
  id: string;
  role: "user" | "assistant";
  content: string;
  sources?: RagSource[];
  isError?: boolean;
};

const MIN_CHUNKS = 1;
const MAX_CHUNKS = 20;
const DEFAULT_CHUNKS = 5;

function createMessageId() {
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

function toChatErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return `Не удалось получить ответ (${error.status}): ${error.message}`;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return "Не удалось получить ответ.";
}

function getSourceTitle(source: RagSource, index: number): string {
  return source.metadata?.object_key || source.document_id || source.id || `Источник ${index + 1}`;
}

function getSourceMeta(source: RagSource): string {
  const parts = [
    source.metadata?.content_type,
    source.chunk_id || source.metadata?.chunk_index
      ? `фрагмент ${source.chunk_id || source.metadata?.chunk_index}`
      : null,
  ].filter(Boolean);

  return parts.join(" · ");
}

function formatScore(score?: number): string | null {
  if (typeof score !== "number" || Number.isNaN(score)) {
    return null;
  }

  return `${Math.round(score * 100)}%`;
}

export default function ChatPage() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [query, setQuery] = useState("");
  const [chunkLimit, setChunkLimit] = useState(DEFAULT_CHUNKS);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  const latestSources = useMemo(() => {
    return [...messages].reverse().find((message) => message.role === "assistant")?.sources ?? [];
  }, [messages]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
  }, [messages, isSubmitting]);

  const onSubmit = async (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault();

    const trimmedQuery = query.trim();
    if (!trimmedQuery || isSubmitting) {
      return;
    }

    const nextUserMessage: ChatMessage = {
      id: createMessageId(),
      role: "user",
      content: trimmedQuery,
    };

    setMessages((currentMessages) => [...currentMessages, nextUserMessage]);
    setQuery("");
    setIsSubmitting(true);

    try {
      const response = await queryRag({
        query: trimmedQuery,
        limit: chunkLimit,
      });

      setMessages((currentMessages) => [
        ...currentMessages,
        {
          id: createMessageId(),
          role: "assistant",
          content: response.answer || "Ответ пустой.",
          sources: response.sources ?? [],
        },
      ]);
    } catch (error) {
      setMessages((currentMessages) => [
        ...currentMessages,
        {
          id: createMessageId(),
          role: "assistant",
          content: toChatErrorMessage(error),
          isError: true,
        },
      ]);
    } finally {
      setIsSubmitting(false);
    }
  };

  const onQuestionKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key !== "Enter" || event.shiftKey) {
      return;
    }

    event.preventDefault();
    void onSubmit();
  };

  return (
    <section className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_360px]">
      <article className="flex min-h-[calc(100vh-9rem)] flex-col overflow-hidden rounded-2xl border border-border/70 bg-panel shadow-panel">
        <div className="border-b border-border/70 p-5">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <div className="flex items-center gap-2">
                <MessageSquareQuote className="h-5 w-5 text-accent-strong" />
                <h2 className="text-lg font-semibold text-foreground">RAG-чат</h2>
              </div>
              <p className="mt-1 text-sm text-muted">Задайте вопрос по проиндексированным документам.</p>
            </div>

            <button
              type="button"
              onClick={() => setMessages([])}
              disabled={messages.length === 0 || isSubmitting}
              className="inline-flex items-center gap-2 rounded-xl border border-border/70 bg-page/80 px-3 py-2 text-sm font-medium text-foreground transition hover:border-cyan-300 disabled:cursor-not-allowed disabled:opacity-60"
              aria-label="Очистить чат"
            >
              <Trash2 className="h-4 w-4" />
              Очистить
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto bg-page/40 p-4 sm:p-5">
          {messages.length === 0 && !isSubmitting ? (
            <div className="flex min-h-[18rem] items-center justify-center rounded-2xl border border-dashed border-border/80 bg-panel/70 p-6 text-center">
              <div>
                <Bot className="mx-auto h-8 w-8 text-accent-strong" />
                <p className="mt-3 text-sm font-medium text-foreground">Пока нет сообщений</p>
                <p className="mt-1 max-w-sm text-sm leading-6 text-muted">
                  Напишите вопрос ниже, и ответ появится здесь вместе с найденными фрагментами.
                </p>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              {messages.map((message) => {
                const isUser = message.role === "user";
                const Icon = isUser ? User : Bot;

                return (
                  <div
                    key={message.id}
                    className={cn("flex gap-3", isUser ? "justify-end" : "justify-start")}
                  >
                    {!isUser && (
                      <div className="mt-1 flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-accent/10 text-accent-strong">
                        <Icon className="h-4 w-4" />
                      </div>
                    )}

                    <div
                      className={cn(
                        "max-w-[min(42rem,100%)] whitespace-pre-wrap rounded-2xl border px-4 py-3 text-sm leading-6",
                        isUser
                          ? "border-accent/30 bg-accent text-white"
                          : message.isError
                            ? "border-rose-300 bg-rose-500/10 text-rose-800"
                            : "border-border/70 bg-panel text-foreground",
                      )}
                    >
                      {message.content}
                    </div>

                    {isUser && (
                      <div className="mt-1 flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-slate-950/5 text-muted">
                        <Icon className="h-4 w-4" />
                      </div>
                    )}
                  </div>
                );
              })}

              {isSubmitting && (
                <div className="flex gap-3">
                  <div className="mt-1 flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-accent/10 text-accent-strong">
                    <Bot className="h-4 w-4" />
                  </div>
                  <div className="inline-flex items-center gap-2 rounded-2xl border border-border/70 bg-panel px-4 py-3 text-sm text-muted">
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Готовлю ответ...
                  </div>
                </div>
              )}

              <div ref={messagesEndRef} />
            </div>
          )}
        </div>

        <form onSubmit={(event) => void onSubmit(event)} className="border-t border-border/70 bg-panel p-4 sm:p-5">
          <div className="mb-4 rounded-xl border border-border/70 bg-page/70 p-3">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <label htmlFor="chunk-limit" className="flex items-center gap-2 text-sm font-medium text-foreground">
                <SlidersHorizontal className="h-4 w-4 text-accent-strong" />
                Максимум фрагментов
              </label>
              <span className="rounded-full bg-cyan-500/10 px-3 py-1 text-xs font-medium text-cyan-700">
                {chunkLimit}
              </span>
            </div>
            <input
              id="chunk-limit"
              type="range"
              min={MIN_CHUNKS}
              max={MAX_CHUNKS}
              step={1}
              value={chunkLimit}
              onChange={(event) => setChunkLimit(Number(event.target.value))}
              disabled={isSubmitting}
              className="mt-3 w-full accent-cyan-700"
            />
          </div>

          <div className="flex flex-col gap-3 lg:flex-row">
            <textarea
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              onKeyDown={onQuestionKeyDown}
              disabled={isSubmitting}
              rows={3}
              placeholder="Введите вопрос..."
              className="min-h-24 flex-1 resize-none rounded-xl border border-border/70 bg-page/60 px-4 py-3 text-sm leading-6 text-foreground outline-none transition placeholder:text-muted/75 focus:border-cyan-400 focus:ring-4 focus:ring-cyan-500/10 disabled:cursor-not-allowed disabled:opacity-70"
            />
            <button
              type="submit"
              disabled={!query.trim() || isSubmitting}
              className="inline-flex items-center justify-center gap-2 rounded-xl border border-accent/30 bg-accent px-5 py-3 text-sm font-semibold text-white transition hover:bg-accent-strong disabled:cursor-not-allowed disabled:opacity-60 lg:w-36"
            >
              {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
              Отправить
            </button>
          </div>
        </form>
      </article>

      <article className="rounded-2xl border border-border/70 bg-panel p-5 shadow-panel">
        <h2 className="text-base font-semibold text-foreground">Источники</h2>

        {latestSources.length === 0 ? (
          <p className="mt-3 rounded-xl border border-dashed border-border/80 bg-page/60 p-4 text-sm leading-6 text-muted">
            После ответа здесь появятся найденные фрагменты.
          </p>
        ) : (
          <ul className="mt-4 space-y-3">
            {latestSources.map((source, index) => {
              const score = formatScore(source.score);
              const meta = getSourceMeta(source);

              return (
                <li key={source.id || `${source.document_id}-${index}`} className="rounded-xl border border-border/70 p-3">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="truncate text-xs uppercase tracking-[0.12em] text-muted">
                        {getSourceTitle(source, index)}
                      </p>
                      {meta && <p className="mt-1 text-xs text-muted">{meta}</p>}
                    </div>
                    {score && (
                      <span className="shrink-0 rounded-full bg-emerald-500/10 px-2 py-1 text-xs font-medium text-emerald-700">
                        {score}
                      </span>
                    )}
                  </div>

                  {source.text && (
                    <p className="mt-3 max-h-32 overflow-y-auto rounded-lg bg-page/70 p-3 text-sm leading-6 text-foreground">
                      {source.text}
                    </p>
                  )}
                </li>
              );
            })}
          </ul>
        )}
      </article>
    </section>
  );
}
