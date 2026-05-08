"use client";

import { useRef, useState, type ChangeEvent } from "react";
import { ApiError } from "@/components/api/client";
import {
  confirmDocumentUpload,
  initDocumentUpload,
  StorageUploadError,
  uploadFileToStorage,
} from "@/features/documents/api/upload";
import { cn } from "@/lib/cn";

type StepName = "select" | "presign" | "upload" | "confirm";
type UploadStepName = Exclude<StepName, "select">;
type StepStatus = "pending" | "active" | "done" | "error";
type StepState = Record<StepName, StepStatus>;

const ACCEPTED_EXTENSIONS = new Set(["pdf", "txt", "docx"]);
const ACCEPT_ATTRIBUTE = ".pdf,.txt,.docx";

const STEP_DEFINITIONS: Array<{ key: StepName; label: string }> = [
  { key: "select", label: "1. Выбор документа" },
  { key: "presign", label: "2. Подготовка загрузки" },
  { key: "upload", label: "3. Загрузка документа" },
  { key: "confirm", label: "4. Передача в обработку" },
];

const STEP_STATUS_LABEL: Record<StepStatus, string> = {
  pending: "не начато",
  active: "выполняется",
  done: "завершено",
  error: "ошибка",
};

const STEP_STATUS_SYMBOL: Record<StepStatus, string> = {
  pending: "-",
  active: "...",
  done: "ok",
  error: "!",
};

const STEP_STATUS_CLASS: Record<StepStatus, string> = {
  pending: "border-border/60 bg-page/50 text-muted",
  active: "border-cyan-300 bg-cyan-500/10 text-cyan-800",
  done: "border-emerald-300 bg-emerald-500/10 text-emerald-800",
  error: "border-rose-300 bg-rose-500/10 text-rose-800",
};

const INITIAL_STEP_STATE: StepState = {
  select: "pending",
  presign: "pending",
  upload: "pending",
  confirm: "pending",
};

function getFileExtension(fileName: string): string {
  return fileName.split(".").pop()?.toLowerCase() ?? "";
}

function isSupportedFile(file: File): boolean {
  return ACCEPTED_EXTENSIONS.has(getFileExtension(file.name));
}

function markFailedStep(state: StepState, step: UploadStepName): StepState {
  switch (step) {
    case "presign":
      return { ...state, presign: "error" };
    case "upload":
      return { ...state, upload: "error" };
    case "confirm":
      return { ...state, confirm: "error" };
    default:
      return state;
  }
}

function toErrorMessage(error: unknown): string {
  if (error instanceof StorageUploadError) {
    return `Ошибка загрузки в хранилище (HTTP ${error.status}). Подтверждение на бэкенд не отправлено.`;
  }

  if (error instanceof ApiError) {
    return `Ошибка запроса к бэкенду (${error.status}): ${error.message}`;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return "Загрузка завершилась с неизвестной ошибкой.";
}

type DocumentUploadPanelProps = {
  onUploadCompleted?: (documentID: string) => void | Promise<void>;
};

export function DocumentUploadPanel({ onUploadCompleted }: DocumentUploadPanelProps) {
  const inputRef = useRef<HTMLInputElement | null>(null);

  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [stepState, setStepState] = useState<StepState>(INITIAL_STEP_STATE);
  const [isUploading, setIsUploading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const onFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    const nextFile = event.target.files?.[0] ?? null;

    setErrorMessage(null);
    setSuccessMessage(null);

    if (!nextFile) {
      setSelectedFile(null);
      setStepState(INITIAL_STEP_STATE);
      return;
    }

    if (!isSupportedFile(nextFile)) {
      setSelectedFile(null);
      setStepState(INITIAL_STEP_STATE);
      setErrorMessage("Поддерживаются только файлы .pdf, .txt и .docx.");
      event.target.value = "";
      return;
    }

    setSelectedFile(nextFile);
    setStepState({
      ...INITIAL_STEP_STATE,
      select: "done",
    });
  };

  const clearSelection = () => {
    if (isUploading) {
      return;
    }

    setSelectedFile(null);
    setStepState(INITIAL_STEP_STATE);
    setErrorMessage(null);
    setSuccessMessage(null);

    if (inputRef.current) {
      inputRef.current.value = "";
    }
  };

  const onStartUpload = async () => {
    if (!selectedFile || isUploading) {
      return;
    }

    setIsUploading(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    let currentStep: UploadStepName = "presign";

    try {
      setStepState({
        select: "done",
        presign: "active",
        upload: "pending",
        confirm: "pending",
      });
      const initUploadResponse = await initDocumentUpload(selectedFile);

      currentStep = "upload";
      setStepState({
        select: "done",
        presign: "done",
        upload: "active",
        confirm: "pending",
      });
      await uploadFileToStorage(
        selectedFile,
        initUploadResponse.upload_url,
        initUploadResponse.headers,
      );

      currentStep = "confirm";
      setStepState({
        select: "done",
        presign: "done",
        upload: "done",
        confirm: "active",
      });
      await confirmDocumentUpload(initUploadResponse.document_id, selectedFile.size);

      setStepState({
        select: "done",
        presign: "done",
        upload: "done",
        confirm: "done",
      });
      void onUploadCompleted?.(initUploadResponse.document_id);
      setSuccessMessage(`Загрузка завершена. ID документа: ${initUploadResponse.document_id}`);
    } catch (error) {
      setStepState((prevState) => markFailedStep(prevState, currentStep));
      setErrorMessage(toErrorMessage(error));
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <article className="rounded-2xl border border-border/70 bg-panel p-5 shadow-panel">
      <h2 className="text-lg font-semibold text-foreground">Загрузка документа</h2>

      <div className="mt-4 flex flex-col gap-3 lg:flex-row lg:items-center">
        <input
          ref={inputRef}
          type="file"
          accept={ACCEPT_ATTRIBUTE}
          onChange={onFileChange}
          disabled={isUploading}
          className="w-full rounded-xl border border-border/70 bg-page/60 px-3 py-2 text-sm text-foreground file:mr-3 file:rounded-lg file:border-0 file:bg-white file:px-3 file:py-1.5 file:text-sm file:font-medium file:text-foreground"
        />

        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => void onStartUpload()}
            disabled={!selectedFile || isUploading}
            className="rounded-xl border border-border/70 bg-white px-4 py-2 text-sm font-medium text-foreground transition hover:border-cyan-300 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {isUploading ? "Загрузка..." : "Начать загрузку"}
          </button>

          <button
            type="button"
            onClick={clearSelection}
            disabled={isUploading}
            className="rounded-xl border border-border/70 bg-page/80 px-4 py-2 text-sm font-medium text-foreground transition hover:border-border disabled:cursor-not-allowed disabled:opacity-60"
          >
            Сброс
          </button>
        </div>
      </div>

      {selectedFile && (
        <p className="mt-3 text-xs text-muted">
          Выбран файл: {selectedFile.name} ({selectedFile.size} байт)
        </p>
      )}

      <ul className="mt-4 space-y-2">
        {STEP_DEFINITIONS.map((step) => {
          const status = stepState[step.key];

          return (
            <li
              key={step.key}
              className={cn(
                "flex items-center justify-between rounded-xl border px-3 py-2 text-sm",
                STEP_STATUS_CLASS[status],
              )}
            >
              <span className="font-medium">{step.label}</span>
              <span className="text-xs uppercase tracking-[0.08em]">
                {STEP_STATUS_SYMBOL[status]} {STEP_STATUS_LABEL[status]}
              </span>
            </li>
          );
        })}
      </ul>

      {errorMessage && (
        <p className="mt-4 rounded-xl border border-rose-300 bg-rose-500/10 px-3 py-2 text-sm text-rose-800">
          {errorMessage}
        </p>
      )}

      {successMessage && (
        <p className="mt-4 rounded-xl border border-emerald-300 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-800">
          {successMessage}
        </p>
      )}
    </article>
  );
}
