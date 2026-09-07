"use client";

import { useState, type FormEvent } from "react";

export type ResourceOption = { value: string; label: string };

type UploadIntent = {
  document: { id: string };
  uploadUrl: string;
  headers: Record<string, string>;
};

function inferContentType(file: File) {
  if (file.type) return file.type;
  const extension = file.name.toLowerCase().split(".").pop();
  const byExtension: Record<string, string> = {
    pdf: "application/pdf",
    jpg: "image/jpeg",
    jpeg: "image/jpeg",
    png: "image/png",
    webp: "image/webp",
    txt: "text/plain",
    csv: "text/csv",
  };
  return extension ? byExtension[extension] ?? "application/octet-stream" : "application/octet-stream";
}

export function DocumentUpload({ resources }: { resources: ResourceOption[] }) {
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const data = new FormData(form);
    const file = data.get("file");
    const resource = String(data.get("resource") ?? "");
    const kind = String(data.get("kind") ?? "other");
    if (!(file instanceof File) || !file.size || !resource) return;
    const separator = resource.indexOf(":");
    const resourceType = separator >= 0 ? resource.slice(0, separator) : resource;
    const resourceId = separator >= 0 ? resource.slice(separator + 1) : "";
    const contentType = inferContentType(file);

    setBusy(true);
    setMessage("Creating secure upload…");
    try {
      const intentResponse = await fetch("/api/documents/uploads", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ resourceType, resourceId, kind, fileName: file.name, contentType, sizeBytes: file.size }),
      });
      const intentPayload = (await intentResponse.json()) as { data?: UploadIntent; error?: string };
      if (!intentResponse.ok || !intentPayload.data) throw new Error(intentPayload.error ?? "Could not create upload intent");

      setMessage("Uploading directly to object storage…");
      const uploadHeaders = new Headers(intentPayload.data.headers);
      uploadHeaders.set("Content-Type", contentType);
      const uploadResponse = await fetch(intentPayload.data.uploadUrl, { method: "PUT", headers: uploadHeaders, body: file });
      if (!uploadResponse.ok) throw new Error(`Object storage returned ${uploadResponse.status}`);

      setMessage("Verifying uploaded object…");
      const completeResponse = await fetch(`/api/documents/${intentPayload.data.document.id}/complete`, { method: "POST" });
      const completePayload = (await completeResponse.json()) as { error?: string };
      if (!completeResponse.ok) throw new Error(completePayload.error ?? "Upload verification failed");

      setMessage("Upload verified. Refreshing document register…");
      form.reset();
      window.location.reload();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Document upload failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form className="stackedForm" onSubmit={submit}>
      <label>
        Attach to
        <select name="resource" required defaultValue="">
          <option value="" disabled>Select a property record</option>
          {resources.map((resource) => <option value={resource.value} key={resource.value}>{resource.label}</option>)}
        </select>
      </label>
      <label>
        Document type
        <select name="kind" defaultValue="other">
          <option value="lease">Lease</option>
          <option value="identity">Identity</option>
          <option value="inspection">Inspection</option>
          <option value="invoice">Invoice</option>
          <option value="receipt">Receipt</option>
          <option value="maintenance">Maintenance</option>
          <option value="statement">Statement</option>
          <option value="other">Other</option>
        </select>
      </label>
      <label>
        File
        <input name="file" type="file" required accept=".pdf,.jpg,.jpeg,.png,.webp,.txt,.csv,application/pdf,image/jpeg,image/png,image/webp,text/plain,text/csv" />
      </label>
      <button className="primaryButton" type="submit" disabled={busy}>{busy ? "Uploading…" : "Upload securely"}</button>
      {message ? <p className="formHint" aria-live="polite">{message}</p> : null}
    </form>
  );
}

export function DocumentActions({ id, available }: { id: string; available: boolean }) {
  const [busy, setBusy] = useState(false);

  async function remove() {
    if (!window.confirm("Delete this document from Property OS and object storage?")) return;
    setBusy(true);
    try {
      const response = await fetch(`/api/documents/${id}`, { method: "DELETE" });
      if (!response.ok) throw new Error("Delete failed");
      window.location.reload();
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="tableActions">
      {available ? <a className="textLink" href={`/api/documents/${id}/download`} target="_blank" rel="noreferrer">Download</a> : null}
      <button className="textButton" type="button" onClick={remove} disabled={busy}>{busy ? "Deleting…" : "Delete"}</button>
    </div>
  );
}
