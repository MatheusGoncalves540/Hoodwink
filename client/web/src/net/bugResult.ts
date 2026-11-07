/**
 * Resultado de uma chamada de relato de bug, classificado pelo status HTTP.
 * O 0 representa falha de rede (fetch nem chegou a responder).
 */
export type BugResult = "ok" | "limit" | "unauthorized" | "error";

export function classifyBugResult(status: number): BugResult {
  if (status === 200 || status === 201) return "ok";
  if (status === 429) return "limit";
  if (status === 401) return "unauthorized";
  return "error";
}
