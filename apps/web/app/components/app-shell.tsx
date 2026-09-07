import Link from "next/link";
import type { ReactNode } from "react";

const navigation = [
  { href: "/properties", label: "Portfolio" },
  { href: "/tenants", label: "Tenants" },
  { href: "/leases", label: "Leasing" },
  { href: "/rent", label: "Rent" },
  { href: "/maintenance", label: "Maintenance" },
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
          {navigation.map((item) => (
            <Link className={item.label === section ? "navItem navItemActive" : "navItem"} href={item.href} key={item.href}>
              {item.label}
            </Link>
          ))}
        </nav>
        <div className="sidebarFoot">
          <span className="liveDot" />
          <span>Development workspace</span>
        </div>
      </aside>
      <main className="workspace">{children}</main>
    </div>
  );
}
