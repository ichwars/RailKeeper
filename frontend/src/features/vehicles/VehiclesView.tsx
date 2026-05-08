import { FormEvent, useEffect, useState } from "react";
import { Eye, Pencil, Plus, Search, Trash2, X } from "lucide-react";
import { api, CreateVehicleRequest, Vehicle } from "../../shared/api";

const emptyVehicle: CreateVehicleRequest = {
  manufacturer: "",
  articleNumber: "",
  name: "",
  gauge: "H0",
  epoch: "",
  railwayCompany: "",
  category: ""
};

function vehicleToForm(vehicle: Vehicle): CreateVehicleRequest {
  return {
    inventoryNumber: vehicle.inventoryNumber,
    manufacturer: vehicle.manufacturer,
    articleNumber: vehicle.articleNumber || "",
    name: vehicle.name,
    gauge: vehicle.gauge,
    epoch: vehicle.epoch || "",
    railwayCompany: vehicle.railwayCompany || "",
    category: vehicle.category || ""
  };
}

export function VehiclesView() {
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [form, setForm] = useState<CreateVehicleRequest>(emptyVehicle);
  const [query, setQuery] = useState("");
  const [message, setMessage] = useState("");
  const [saving, setSaving] = useState(false);
  const [selected, setSelected] = useState<Vehicle | null>(null);
  const [mode, setMode] = useState<"create" | "view" | "edit">("create");
  const [deleteCandidate, setDeleteCandidate] = useState<Vehicle | null>(null);

  const load = () => {
    api
      .vehicles(query)
      .then(setVehicles)
      .catch((error: Error) => setMessage(error.message));
  };

  useEffect(() => {
    load();
  }, [query]);

  const update = (patch: Partial<CreateVehicleRequest>) => {
    setForm((current) => ({ ...current, ...patch }));
  };

  const resetCreate = () => {
    setSelected(null);
    setMode("create");
    setForm(emptyVehicle);
    setMessage("");
  };

  const openDetail = (vehicle: Vehicle) => {
    api
      .vehicle(vehicle.id)
      .then((detail) => {
        setSelected(detail);
        setForm(vehicleToForm(detail));
        setMode("view");
        setMessage("");
      })
      .catch((error: Error) => setMessage(error.message));
  };

  const openEdit = (vehicle: Vehicle) => {
    api
      .vehicle(vehicle.id)
      .then((detail) => {
        setSelected(detail);
        setForm(vehicleToForm(detail));
        setMode("edit");
        setMessage("");
      })
      .catch((error: Error) => setMessage(error.message));
  };

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setSaving(true);
    setMessage("");

    const action = mode === "edit" && selected
      ? api.updateVehicle(selected.id, form)
      : api.createVehicle(form);

    action
      .then((vehicle) => {
        setSelected(vehicle);
        setForm(mode === "edit" ? vehicleToForm(vehicle) : emptyVehicle);
        setMode(mode === "edit" ? "view" : "create");
        load();
      })
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
  };

  const confirmDelete = () => {
    if (!deleteCandidate) return;

    api
      .deleteVehicle(deleteCandidate.id)
      .then(() => {
        if (selected?.id === deleteCandidate.id) {
          resetCreate();
        }
        setDeleteCandidate(null);
        load();
      })
      .catch((error: Error) => setMessage(error.message));
  };

  const readonly = mode === "view";

  return (
    <>
      <section className="page-head vehicles-head">
        <div>
          <p className="eyebrow">Fahrzeuge</p>
          <h1>Bestand aufbauen</h1>
          <p>Lege Fahrzeuge mit den wichtigsten Stammdaten an. Bilder, Wartung und CV-Daten folgen auf diesem Modell.</p>
        </div>
      </section>

      <section className="work-grid">
        <form className="panel vehicle-form" onSubmit={submit}>
          <div className="panel-head form-head">
            <div>
              {mode === "create" && <Plus size={18} aria-hidden="true" />}
              {mode === "view" && <Eye size={18} aria-hidden="true" />}
              {mode === "edit" && <Pencil size={18} aria-hidden="true" />}
              <h2>{mode === "create" ? "Fahrzeug anlegen" : mode === "edit" ? "Fahrzeug bearbeiten" : "Fahrzeugdetail"}</h2>
            </div>
            {mode !== "create" && (
              <button type="button" className="icon-button" onClick={resetCreate} aria-label="Schliessen" title="Schliessen">
                <X size={17} />
              </button>
            )}
          </div>

          {mode !== "create" && (
            <label>
              Inventar-Nr.
              <input value={form.inventoryNumber || ""} onChange={(event) => update({ inventoryNumber: event.target.value })} disabled={readonly} />
            </label>
          )}

          <label>
            Hersteller
            <input value={form.manufacturer} onChange={(event) => update({ manufacturer: event.target.value })} disabled={readonly} required />
          </label>

          <label>
            Bezeichnung
            <input value={form.name} onChange={(event) => update({ name: event.target.value })} disabled={readonly} required />
          </label>

          <div className="form-row">
            <label>
              Spur
              <input value={form.gauge} onChange={(event) => update({ gauge: event.target.value })} disabled={readonly} required />
            </label>
            <label>
              Artikel-Nr.
              <input value={form.articleNumber || ""} onChange={(event) => update({ articleNumber: event.target.value })} disabled={readonly} />
            </label>
          </div>

          <div className="form-row">
            <label>
              Epoche
              <input value={form.epoch || ""} onChange={(event) => update({ epoch: event.target.value })} disabled={readonly} />
            </label>
            <label>
              Bahngesellschaft
              <input value={form.railwayCompany || ""} onChange={(event) => update({ railwayCompany: event.target.value })} disabled={readonly} />
            </label>
          </div>

          <label>
            Kategorie
            <input value={form.category || ""} onChange={(event) => update({ category: event.target.value })} disabled={readonly} />
          </label>

          {selected && readonly && (
            <dl className="detail-meta">
              <div>
                <dt>Erstellt</dt>
                <dd>{new Date(selected.createdAt).toLocaleString("de-DE")}</dd>
              </div>
              <div>
                <dt>Aktualisiert</dt>
                <dd>{new Date(selected.updatedAt).toLocaleString("de-DE")}</dd>
              </div>
            </dl>
          )}

          {readonly ? (
            <button type="button" className="primary-button" onClick={() => setMode("edit")}>
              Bearbeiten
            </button>
          ) : (
            <button className="primary-button" disabled={saving}>
              {saving ? "Wird gespeichert..." : mode === "edit" ? "Aenderungen speichern" : "Fahrzeug speichern"}
            </button>
          )}

          {message && <p className="form-message">{message}</p>}
        </form>

        <section className="panel vehicle-list">
          <div className="panel-head list-head">
            <h2>Fahrzeuge</h2>
            <label className="search-field">
              <Search size={16} aria-hidden="true" />
              <input
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Suchen"
              />
            </label>
          </div>

          {vehicles.length === 0 ? (
            <p className="empty-state">Noch keine Fahrzeuge vorhanden.</p>
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Inventar</th>
                    <th>Hersteller</th>
                    <th>Artikel</th>
                    <th>Bezeichnung</th>
                    <th>Spur</th>
                    <th className="actions-cell">Aktionen</th>
                  </tr>
                </thead>
                <tbody>
                  {vehicles.map((vehicle) => (
                    <tr key={vehicle.id}>
                      <td>{vehicle.inventoryNumber}</td>
                      <td>{vehicle.manufacturer}</td>
                      <td>{vehicle.articleNumber || "-"}</td>
                      <td>{vehicle.name}</td>
                      <td>{vehicle.gauge}</td>
                      <td className="actions-cell">
                        <div className="table-actions">
                          <button type="button" className="icon-button" onClick={() => openDetail(vehicle)} aria-label="Anzeigen" title="Anzeigen">
                            <Eye size={16} />
                          </button>
                          <button type="button" className="icon-button" onClick={() => openEdit(vehicle)} aria-label="Bearbeiten" title="Bearbeiten">
                            <Pencil size={16} />
                          </button>
                          <button type="button" className="icon-button danger" onClick={() => setDeleteCandidate(vehicle)} aria-label="Loeschen" title="Loeschen">
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </section>

      {deleteCandidate && (
        <div className="confirm-layer" role="dialog" aria-modal="true" aria-label="Fahrzeug loeschen">
          <section className="confirm-card">
            <div className="panel-head form-head">
              <h2>Fahrzeug loeschen?</h2>
              <button type="button" className="icon-button" onClick={() => setDeleteCandidate(null)} aria-label="Schliessen">
                <X size={17} />
              </button>
            </div>
            <p>
              {deleteCandidate.inventoryNumber} - {deleteCandidate.name}
            </p>
            <div className="confirm-actions">
              <button type="button" className="secondary-button" onClick={() => setDeleteCandidate(null)}>
                Abbrechen
              </button>
              <button type="button" className="danger-button" onClick={confirmDelete}>
                Loeschen
              </button>
            </div>
          </section>
        </div>
      )}
    </>
  );
}
