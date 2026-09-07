"use server";

import { revalidatePath } from "next/cache";
import { decideAgentAction } from "../lib/agent-actions";

export async function decideAgentActionAction(formData: FormData) {
  const id = String(formData.get("actionId") ?? "").trim();
  const decision = String(formData.get("decision") ?? "reject") as "approve" | "reject";
  const reason = String(formData.get("reason") ?? "").trim();
  if (!reason) {
    throw new Error("A human review reason is required.");
  }
  await decideAgentAction(id, decision, reason);
  revalidatePath("/agent-actions");
  revalidatePath("/maintenance");
}
