import Link from "next/link";
import type { ReactNode } from "react";

export function PortalShell({ children, mode }: { children: ReactNode; mode: "owner" | "tenant" }) {
  const title = mode === "owner" ? "Owner Portal" : "Tenant Portal";
  return (
    <div className="portalFrame">
      <header className="portalTopbar">
        <div>
          <p className="brandMark">PROPERTY OS</p>
          <p className="brandSubline">{title}</p>
        </div>
        <nav className="portalNav" aria-label={`${title} navigation`}>
          <Link className="portalNavItem" href={mode === "owner" ? "/owner-portal" : "/tenant-portal"}>Overview</Link>
        </nav>
      </header>
      <main className="portalWorkspace">{children}</main>
    </div>
  );
}
