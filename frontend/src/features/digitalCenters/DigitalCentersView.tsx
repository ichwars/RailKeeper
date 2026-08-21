export function DigitalCentersView({ roles }: { roles: string[] }) {
  return (
    <section
      className="digital-centers-workspace"
      data-can-administer={roles.includes("Admin")}
    >
      <header>
        <p className="eyebrow">DIGITALBETRIEB</p>
        <h1>Digitalzentralen</h1>
        <p>Zentralen, Live-Daten und Synchronisation in einer Arbeitsansicht.</p>
      </header>
    </section>
  );
}
