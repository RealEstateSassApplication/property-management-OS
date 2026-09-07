import type { Metadata } from "next";
import "./globals.css";
import "./portal.css";
import "./admin.css";

export const metadata: Metadata = {
  title: "Property Management OS",
  description: "Operate properties, leases, rent, maintenance, and owner reporting from one system.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
