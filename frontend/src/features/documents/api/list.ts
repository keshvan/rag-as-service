import { apiClient } from "@/features/auth/lib/auth-api";

export type DocumentQueueStatus =
  | "pending_upload"
  | "uploaded"
  | "processing"
  | "indexed"
  | "failed";

export type DocumentQueueItem = {
  id: string;
  file_name: string;
  content_type: string;
  status: DocumentQueueStatus | string;
  created_at: string;
};

type ListDocumentsResponse = {
  items: DocumentQueueItem[];
};

type ListDocumentsParams = {
  limit?: number;
  offset?: number;
};

export async function listDocuments(params: ListDocumentsParams = {}): Promise<DocumentQueueItem[]> {
  const query = new URLSearchParams();

  if (typeof params.limit === "number") {
    query.set("limit", String(params.limit));
  }

  if (typeof params.offset === "number") {
    query.set("offset", String(params.offset));
  }

  const queryString = query.toString();
  const path = queryString.length > 0 ? `/documents?${queryString}` : "/documents";

  const response = await apiClient.get<ListDocumentsResponse>(path, {
    cache: "no-store",
  });

  if (!Array.isArray(response.items)) {
    return [];
  }

  return response.items;
}
