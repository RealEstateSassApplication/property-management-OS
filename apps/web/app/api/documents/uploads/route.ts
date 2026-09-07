import { NextResponse } from "next/server";
import { initiateDocumentUpload } from "../../../lib/documents";

export async function POST(request: Request) {
  try {
    const input = await request.json();
    const data = await initiateDocumentUpload(input);
    return NextResponse.json({ data }, { status: 201 });
  } catch (error) {
    return NextResponse.json({ error: error instanceof Error ? error.message : "Could not create document upload" }, { status: 400 });
  }
}
