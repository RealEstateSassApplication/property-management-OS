import { NextResponse } from "next/server";
import { getDocumentDownload } from "../../../../lib/documents";

export async function GET(_request: Request, { params }: { params: Promise<{ documentID: string }> }) {
  try {
    const { documentID } = await params;
    const grant = await getDocumentDownload(documentID);
    return NextResponse.redirect(grant.url, 307);
  } catch (error) {
    return NextResponse.json({ error: error instanceof Error ? error.message : "Could not create document download" }, { status: 409 });
  }
}
