import { Database, MoreVertical, Play, Plus } from "lucide-react";

import { formatDateTime, type Language } from "../../shared/i18n";
import type { DataTransferArea, DataTransferProfile } from "./dataTransferModel";

type Translate = (key: string, values?: Record<string, string | number>) => string;

type TransferProfilesTableProps = {
  canCreate: boolean;
  canEdit: boolean;
  language: Language;
  mutating: boolean;
  onCreate: () => void;
  onEdit: (profileId: string) => void;
  onRun: (profileId: string) => void;
  profiles: DataTransferProfile[];
  t: Translate;
};

export function TransferProfilesTable({
  canCreate,
  canEdit,
  language,
  mutating,
  onCreate,
  onEdit,
  onRun,
  profiles,
  t
}: TransferProfilesTableProps) {
  const exportProfiles = profiles.filter((profile) => profile.direction === "export" && profile.enabled);

  return (
    <section className="panel data-transfer-panel transfer-profiles-panel">
      <header className="data-transfer-panel-head">
        <h2><Database size={20} aria-hidden="true" />{t("importExport.dashboard.profiles.title")}</h2>
        {canCreate && (
          <button type="button" className="secondary-button data-transfer-small-action" onClick={onCreate}>
            <Plus size={16} aria-hidden="true" />{t("importExport.dashboard.profiles.create")}
          </button>
        )}
      </header>
      <div className="data-transfer-table-wrap">
        <table className="data-transfer-table transfer-profiles-table">
          <thead>
            <tr>
              <th>{t("importExport.dashboard.profiles.profile")}</th>
              <th>{t("importExport.dashboard.profiles.areas")}</th>
              <th>{t("importExport.dashboard.profiles.format")}</th>
              <th>{t("importExport.dashboard.profiles.lastUsed")}</th>
              <th>{t("importExport.dashboard.profiles.action")}</th>
            </tr>
          </thead>
          <tbody>
            {exportProfiles.length === 0 ? (
              <tr><td className="data-transfer-empty" colSpan={5}>{t("importExport.dashboard.profiles.empty")}</td></tr>
            ) : exportProfiles.map((profile) => (
              <tr key={profile.id}>
                <td><strong className="data-transfer-truncate" title={profile.name}>{profile.name}</strong></td>
                <td className="data-transfer-truncate" title={areaLabels(profile.areas, t)}>
                  {areaLabels(profile.areas, t)}
                </td>
                <td>{profile.format === "csv" ? "CSV" : "JSON"}</td>
                <td>{profile.lastUsedAt ? formatDateTime(profile.lastUsedAt, language) : "–"}</td>
                <td>
                  <span className="transfer-profile-actions">
                    <button
                      type="button"
                      className="transfer-inline-run"
                      disabled={mutating}
                      onClick={() => onRun(profile.id)}
                    >
                      <Play size={13} fill="currentColor" aria-hidden="true" />
                      {t("importExport.dashboard.profiles.run")}
                    </button>
                    {canEdit && (
                      <button
                        type="button"
                        className="icon-button transfer-row-menu"
                        aria-label={`${profile.name} ${t("importExport.dashboard.profiles.edit")}`}
                        title={t("importExport.dashboard.profiles.edit")}
                        onClick={() => onEdit(profile.id)}
                      >
                        <MoreVertical size={16} aria-hidden="true" />
                      </button>
                    )}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function areaLabels(areas: DataTransferArea[], t: Translate) {
  return areas.map((area) => t(`importExport.dashboard.area.${area}`)).join(", ");
}
