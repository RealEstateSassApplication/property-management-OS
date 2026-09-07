import Link from "next/link";
import type { ReactNode } from "react";

const navigation = [
  { href: "/properties", label: "Portfolio", ready: true },
  { href: "/tenants", label: "Tenants", ready: false },
  { href: "/leases", label: "Leasing", ready: false },
  { href: "/rent", label: "Rent", ready: false },
  { href: "/maintenance", label: "Maintenance", ready: false },
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
          {navigation.map((item) => {
            const className = item.label === section ? "navItem navItemActive" : "navItem";

            if (!item.ready) {
              return (
                <span
                  aria-disabled="true"
                  className={`${className} navItemDisabled`}
                  key={item.href}
                  title="Coming in the next development tranche"
                >
                  {item.label}
                  <small>Next</small>
                </span>
              );
            }

            return (
              <Link className={className} href={item.href} key={item.href}>
                {item.label}
              </Link>
            );
          })}
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
