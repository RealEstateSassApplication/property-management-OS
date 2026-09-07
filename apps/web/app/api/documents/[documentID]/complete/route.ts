import { NextResponse } from "next/server";
import { completeDocumentUpload } from "../../../../../lib/documents";

export async function POST(_request: Request, { params }: { params: Promise<{ documentID: string }> }) {
  try {
    const { documentID } = await params;
    const data = await completeDocumentUpload(documentID);
    return NextResponse.json({ data });
  } catch (error) {
    return NextResponse.json({ error: error instanceof Error ? error.message : "Could not finalize document upload" }, { status: 409 });
  }
}
