import Link from "next/link";
import type { ReactNode } from "react";

const navigation = [
  { href: "/properties", label: "Portfolio", enabled: true },
  { href: "/owners", label: "Owners", enabled: true },
  { href: "/tenants", label: "Tenants", enabled: true },
  { href: "/leases", label: "Leasing", enabled: true },
  { href: "/rent", label: "Rent", enabled: true },
  { href: "/accounting", label: "Accounting", enabled: true },
  { href: "/maintenance", label: "Maintenance", enabled: true },
  { href: "/documents", label: "Documents", enabled: true },
  { href: "/notifications", label: "Notifications", enabled: true },
  { href: "/agent-actions", label: "Agent Review", enabled: true },
  { href: "/reports", label: "Reports", enabled: true },
  { href: "/settings", label: "Settings", enabled: true },
];

export function AppShell({ children, section = "Portfolio" }: { children: ReactNode; section?: string }) {
  return (
    <div className="appFrame">
      <aside className="sidebar">
        <div>
          <p className="brandMark">PROPERTY OS</p>
          <p className="brandSubline">Real estate operations</p>
        </div>
        <nav className="navList" aria-label="Primary navigation">
          {navigation.map((item) =>
            item.enabled ? (
              <Link className={item.label === section ? "navItem navItemActive" : "navItem"} href={item.href} key={item.href}>
                {item.label}
              </Link>
            ) : (
              <span className="navItem navItemDisabled" key={item.href} title="Coming next">
                {item.label}<small>Next</small>
              </span>
            ),
          )}
        </nav>
        <div className="sidebarFoot"><span className="liveDot" /><span>Development workspace</span></div>
      </aside>
      <main className="workspace">{children}</main>
    </div>
  );
}
