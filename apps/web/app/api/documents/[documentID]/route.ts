import { NextResponse } from "next/server";
import { deleteDocument } from "../../../lib/documents";

export async function DELETE(_request: Request, { params }: { params: Promise<{ documentID: string }> }) {
  try {
    const { documentID } = await params;
    const data = await deleteDocument(documentID);
    return NextResponse.json({ data });
  } catch (error) {
    return NextResponse.json({ error: error instanceof Error ? error.message : "Could not delete document" }, { status: 502 });
  }
}
