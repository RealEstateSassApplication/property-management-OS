const modules = [
  { name: "Portfolio", description: "Properties, units, owners, and occupancy." },
  { name: "Leasing", description: "Tenants, tenancies, leases, deposits, and renewals." },
  { name: "Rent", description: "Obligations, payments, arrears, adjustments, and ledger history." },
  { name: "Maintenance", description: "Requests, work orders, vendors, approvals, and proof of work." },
];

export default function Home() {
  return (
    <main className="shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">PROPERTY MANAGEMENT OS</p>
          <h1>One operating layer for every property.</h1>
        </div>
        <span className="status">Foundation · V0.1</span>
      </header>

      <section className="hero">
        <p>
          The management platform for the part of real estate that starts after a property is
          onboarded: occupancy, leasing, rent, maintenance, documents, and owner reporting.
        </p>
      </section>

      <section className="moduleGrid" aria-label="Core modules">
        {modules.map((module, index) => (
          <article className="moduleCard" key={module.name}>
            <span className="moduleIndex">0{index + 1}</span>
            <div>
              <h2>{module.name}</h2>
              <p>{module.description}</p>
            </div>
          </article>
        ))}
      </section>

      <footer className="footer">
        <span>Next.js web</span>
        <span>Go API</span>
        <span>PostgreSQL</span>
        <span>OpenAPI</span>
      </footer>
    </main>
  );
}
