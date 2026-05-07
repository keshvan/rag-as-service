import { apiClient } from "@/features/auth/lib/auth-api";

type InitUploadRequest = {
  file_name: string;
  content_type: string;
};

export type InitUploadResponse = {
  document_id: string;
  upload_url: string;
  headers: Record<string, string>;
  object_key: string;
  expires_in: number;
};

type ConfirmUploadRequest = {
  document_id: string;
  size_bytes: number;
};

export type ConfirmUploadResponse = {
  status: string;
};

const CONTENT_TYPE_BY_EXTENSION: Record<string, string> = {
  pdf: "application/pdf",
  txt: "text/plain",
  docx: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
};

export class StorageUploadError extends Error {
  constructor(public readonly status: number) {
    super(`Storage upload failed with status ${status}`);
    this.name = "StorageUploadError";
  }
}

function resolveContentType(file: File): string {
  const type = file.type.trim();
  if (type.length > 0) {
    return type;
  }

  const extension = file.name.split(".").pop()?.toLowerCase() ?? "";
  return CONTENT_TYPE_BY_EXTENSION[extension] ?? "application/octet-stream";
}

export async function initDocumentUpload(file: File): Promise<InitUploadResponse> {
  const payload: InitUploadRequest = {
    file_name: file.name,
    content_type: resolveContentType(file),
  };

  return apiClient.post<InitUploadResponse>("/documents/presign-upload", {
    body: JSON.stringify(payload),
  });
}

export async function uploadFileToStorage(file: File, uploadURL: string, headers: Record<string, string>): Promise<void> {
  const response = await fetch(uploadURL, {
    method: "PUT",
    headers,
    body: file,
  });

  // Product requirement: backend confirmation should happen only after explicit 200 from storage.
  if (response.status !== 200) {
    throw new StorageUploadError(response.status);
  }
}

export async function confirmDocumentUpload(documentID: string, sizeBytes: number): Promise<ConfirmUploadResponse> {
  const payload: ConfirmUploadRequest = {
    document_id: documentID,
    size_bytes: sizeBytes,
  };

  return apiClient.post<ConfirmUploadResponse>("/documents/confirm", {
    body: JSON.stringify(payload),
  });
}
