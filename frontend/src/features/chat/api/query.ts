import { apiClient } from "@/features/auth/lib/auth-api";

export type RagSource = {
  id?: string;
  document_id?: string;
  chunk_id?: string;
  text?: string;
  score?: number;
  metadata?: Record<string, string>;
};

type RagQueryResponse = {
  answer: string;
  sources?: RagSource[];
};

type RagQueryPayload = {
  query: string;
  limit: number;
};

export async function queryRag(payload: RagQueryPayload): Promise<RagQueryResponse> {
  return apiClient.post<RagQueryResponse>("/rag/query", {
    body: JSON.stringify({
      query: payload.query,
      limit: payload.limit,
    }),
  });
}
