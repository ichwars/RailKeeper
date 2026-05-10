import { useEffect, useMemo, useState } from "react";
import { AlertTriangle, BarChart3, Box, CalendarClock, Gauge, Image, Wrench } from "lucide-react";
import { api, Vehicle, VehicleMaintenance } from "../../shared/api";

function numberValue(value?: string) {
  if (!value) {
    return 0;
  }
  const normalized = value.replace(/[^\d,.-]/g, "").replace(/\./g, "").replace(",", ".");
  const parsed = Number.parseFloat(normalized);
  return Number.isFinite(parsed) ? parsed : 0;
}

function currency(value: number) {
  return new Intl.NumberFormat("de-DE", { style: "currency", currency: "EUR", maximumFractionDigits: 0 }).format(value);
}

function formatDate(value?: string) {
  if (!value) {
    return "-";
  }
  return new Intl.DateTimeFormat("de-DE", { day: "2-digit", month: "2-digit", year: "numeric" }).format(new Date(`${value}T00:00:00`));
}

function dateDistance(entry: VehicleMaintenance) {
  if (!entry.dueDate || entry.status === "erledigt") {
    return null;
  }
  const now = new Date();
  const due = new Date(`${entry.dueDate}T00:00:00`);
  return Math.ceil((due.getTime() - now.getTime()) / 86400000);
}

function topEntries(values: string[]) {
  const counts = new Map<string, number>();
  for (const value of values.filter(Boolean)) {
    counts.set(value, (counts.get(value) || 0) + 1);
  }
  return [...counts.entries()].sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0])).slice(0, 5);
}

function maintenanceDistanceText(days: number) {
  if (days < 0) {
    return `${Math.abs(days)} Tage überfällig`;
  }
  if (days === 0) {
    return "heute fällig";
  }
  if (days === 1) {
    return "morgen fällig";
  }
  return `in ${days} Tagen`;
}

export function OverviewView() {
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");

  useEffect(() => {
    api
      .vehicles()
      .then(setVehicles)
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setLoading(false));
  }, []);

  const stats = useMemo(() => {
    const totalValue = vehicles.reduce((sum, vehicle) => sum + numberValue(vehicle.listPrice), 0);
    const digital = vehicles.filter((vehicle) => vehicle.digital).length;
    const analog = vehicles.length - digital;
    const withImages = vehicles.filter((vehicle) => (vehicle.images || []).length > 0).length;
    const allMaintenance = vehicles.flatMap((vehicle) => (vehicle.maintenance || []).map((entry) => ({ vehicle, entry, days: dateDistance(entry) })));
    const scheduledMaintenance = allMaintenance.filter((item): item is { vehicle: Vehicle; entry: VehicleMaintenance; days: number } => item.days !== null && item.entry.status !== "erledigt");
    const due = scheduledMaintenance.filter((item) => item.days <= 0).length;
    const upcoming = scheduledMaintenance.filter((item) => item.days > 0 && item.days <= 30).length;
    const openMaintenance = allMaintenance.filter((item) => item.entry.status !== "erledigt").length;
    const completedMaintenance = allMaintenance.filter((item) => item.entry.status === "erledigt").length;
    const maintenanceCost = allMaintenance.reduce((sum, item) => sum + numberValue(item.entry.cost), 0);
    const nextMaintenance = [...scheduledMaintenance].sort((a, b) => a.days - b.days).slice(0, 4);
    const conditions = topEntries(allMaintenance.map((item) => item.entry.conditionRating || "").filter(Boolean));
    const categories = topEntries(vehicles.map((vehicle) => vehicle.category || "Ohne Kategorie"));
    const gauges = topEntries(vehicles.map((vehicle) => vehicle.gauge || "Ohne Spur"));
    const manufacturers = topEntries(vehicles.map((vehicle) => vehicle.manufacturer || "Ohne Hersteller"));
    return { totalValue, digital, analog, withImages, due, upcoming, openMaintenance, completedMaintenance, maintenanceCost, nextMaintenance, conditions, categories, gauges, manufacturers };
  }, [vehicles]);

  const digitalShare = vehicles.length ? Math.round((stats.digital / vehicles.length) * 100) : 0;
  const imageShare = vehicles.length ? Math.round((stats.withImages / vehicles.length) * 100) : 0;

  return (
    <>
      <section className="page-head overview-head">
        <p className="eyebrow">RailKeeper Cockpit</p>
        <h1>Übersicht</h1>
        <p>Der schnelle Blick auf Bestand, Wert, Digitalisierung und offene Aufgaben.</p>
      </section>

      {message && <p className="form-message">{message}</p>}

      <section className="overview-hero panel">
        <div>
          <span className="overview-icon"><Box size={20} aria-hidden="true" /></span>
          <p>Gesamtbestand</p>
          <strong>{loading ? "..." : vehicles.length}</strong>
          <small>{stats.categories.length} Kategorien, {stats.gauges.length} Spurweiten</small>
        </div>
        <div>
          <span className="overview-icon"><Gauge size={20} aria-hidden="true" /></span>
          <p>Digitalisierung</p>
          <strong>{digitalShare}%</strong>
          <small>{stats.digital} digital · {stats.analog} analog</small>
        </div>
        <div>
          <span className="overview-icon"><BarChart3 size={20} aria-hidden="true" /></span>
          <p>Erfasster Listenwert</p>
          <strong>{currency(stats.totalValue)}</strong>
          <small>Basis: gepflegte Listenpreise</small>
        </div>
        <div className={stats.due > 0 ? "attention" : ""}>
          <span className="overview-icon">{stats.due > 0 ? <AlertTriangle size={20} aria-hidden="true" /> : <Wrench size={20} aria-hidden="true" />}</span>
          <p>Wartung</p>
          <strong>{stats.due}</strong>
          <small>{stats.upcoming} in 30 Tagen · {stats.openMaintenance} offen</small>
        </div>
      </section>

      <section className="overview-grid">
        <article className="panel insight-card">
          <div className="panel-head">
            <div>
              <h2>Bestandsmix</h2>
              <p>Kategorien mit den meisten Fahrzeugen.</p>
            </div>
          </div>
          <div className="bar-list">
            {stats.categories.map(([label, count]) => (
              <div key={label}>
                <span>{label}</span>
                <strong>{count}</strong>
                <i style={{ width: `${vehicles.length ? Math.max(8, (count / vehicles.length) * 100) : 0}%` }} />
              </div>
            ))}
          </div>
        </article>

        <article className="panel insight-card">
          <div className="panel-head">
            <div>
              <h2>Datenqualität</h2>
              <p>Was schon gut gepflegt ist.</p>
            </div>
          </div>
          <div className="quality-list">
            <div><span>Bilder</span><strong>{imageShare}%</strong><i style={{ width: `${imageShare}%` }} /></div>
            <div><span>Decoder-Nummern</span><strong>{vehicles.filter((vehicle) => vehicle.digitalDecoderNumber).length}</strong><i style={{ width: `${vehicles.length ? (vehicles.filter((vehicle) => vehicle.digitalDecoderNumber).length / vehicles.length) * 100 : 0}%` }} /></div>
            <div><span>Artikelnummern</span><strong>{vehicles.filter((vehicle) => vehicle.articleNumber).length}</strong><i style={{ width: `${vehicles.length ? (vehicles.filter((vehicle) => vehicle.articleNumber).length / vehicles.length) * 100 : 0}%` }} /></div>
          </div>
        </article>

        <article className="panel insight-card">
          <div className="panel-head">
            <div>
              <h2>Hersteller</h2>
              <p>Die stärksten Hersteller im Bestand.</p>
            </div>
          </div>
          <div className="rank-list">
            {stats.manufacturers.map(([label, count], index) => (
              <div key={label}><span>{index + 1}</span><strong>{label}</strong><em>{count}</em></div>
            ))}
          </div>
        </article>

        <article className="panel insight-card maintenance-insight-card">
          <div className="panel-head">
            <div>
              <h2>Wartungsradar</h2>
              <p>Die nächsten fälligen Arbeiten im Blick.</p>
            </div>
            <CalendarClock size={18} aria-hidden="true" />
          </div>
          {stats.nextMaintenance.length === 0 ? (
            <p className="empty-mini">Keine geplanten Wartungen mit Fälligkeitsdatum.</p>
          ) : (
            <div className="maintenance-overview-list">
              {stats.nextMaintenance.map(({ vehicle, entry, days }) => (
                <div key={`${vehicle.id}-${entry.id}`} className={days <= 0 ? "due" : ""}>
                  <span>
                    <strong>{vehicle.inventoryNumber}</strong>
                    <small>{vehicle.name || entry.kind}</small>
                  </span>
                  <em>{entry.kind}</em>
                  <b>{maintenanceDistanceText(days)}</b>
                  <small>{formatDate(entry.dueDate)}</small>
                </div>
              ))}
            </div>
          )}
          <div className="maintenance-kpi-row">
            <span><small>Erledigt</small><strong>{stats.completedMaintenance}</strong></span>
            <span><small>Kosten</small><strong>{currency(stats.maintenanceCost)}</strong></span>
            <span><small>Zustände</small><strong>{stats.conditions.length}</strong></span>
          </div>
        </article>

        <article className="panel insight-card">
          <div className="panel-head">
            <div>
              <h2>Nächster Mehrwert</h2>
              <p>Automatisch aus deinen Daten abgeleitet.</p>
            </div>
            <Image size={18} aria-hidden="true" />
          </div>
          <p className="recommendation">
            {vehicles.length === 0
              ? "Lege die ersten Fahrzeuge an oder importiere eine Bestandsliste."
              : imageShare < 70
                ? "Mehr Hauptbilder würden Kartenansicht, Drucklisten und QR-Etiketten deutlich nützlicher machen."
                : stats.due > 0
                  ? "Die fälligen Wartungen sind der beste nächste Arbeitspunkt."
                  : "Der Bestand wirkt stabil. Als nächstes lohnen sich Ersatzteile und strukturierte Preis-/Wertpflege."}
          </p>
        </article>
      </section>
    </>
  );
}
