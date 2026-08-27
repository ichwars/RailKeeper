import { Database, FileDown, FileUp, MoreVertical, Plus } from "lucide-react";
import type { KeyboardEvent } from "react";

import { formatDateTime, type Language } from "../../shared/i18n";
import type { DataTransferArea, DataTransferProfile } from "./dataTransferModel";

type Translate = (key: string, values?: Record<string, string | number>) => string;

type TransferProfilesTableProps = {
  canCreate: boolean;
  canEdit: boolean;
  canExport: boolean;
  canImport: boolean;
  language: Language;
  mutating: boolean;
  onCreate: () => void;
  onEdit: (profileId: string) => void;
  onRun: (profile: DataTransferProfile) => void;
  profiles: DataTransferProfile[];
  t: Translate;
};

export function TransferProfilesTable({
  canCreate,
  canEdit,
  canExport,
  canImport,
  language,
  mutating,
  onCreate,
  onEdit,
  onRun,
  profiles,
  t
}: TransferProfilesTableProps) {
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
              <th>{t("importExport.dashboard.profiles.direction")}</th>
              <th>{t("importExport.dashboard.profiles.status")}</th>
              <th>{t("importExport.dashboard.profiles.areas")}</th>
              <th>{t("importExport.dashboard.profiles.format")}</th>
              <th>{t("importExport.dashboard.profiles.lastUsed")}</th>
              <th>{t("importExport.dashboard.profiles.action")}</th>
            </tr>
          </thead>
          <tbody>
            {profiles.length === 0 ? (
              <tr><td className="data-transfer-empty" colSpan={7}>{t("importExport.dashboard.profiles.empty")}</td></tr>
            ) : profiles.map((profile) => {
              const RunIcon = profile.direction === "import" ? FileUp : FileDown;
              const runTitle = t(`importExport.dashboard.profiles.run.${profile.direction}`);
              return <tr
                className={canEdit ? "transfer-profile-row" : undefined}
                key={profile.id}
                onClick={canEdit ? () => onEdit(profile.id) : undefined}
                onKeyDown={canEdit ? (event) => editFromKeyboard(event, profile.id, onEdit) : undefined}
                tabIndex={canEdit ? 0 : undefined}
              >
                <td><strong className="data-transfer-truncate" title={profile.name}>{profile.name}</strong></td>
                <td>{t(`importExport.dashboard.profiles.${profile.direction}`)}</td>
                <td>
                  <span className={`transfer-profile-status ${profile.enabled ? "enabled" : "disabled"}`}>
                    {t(`importExport.dashboard.profiles.${profile.enabled ? "enabled" : "disabled"}`)}
                  </span>
                </td>
                <td title={areaLabels(profile.areas, t)}>
                  <span className="data-transfer-truncate">{areaLabels(profile.areas, t)}</span>
                </td>
                <td>{profile.format === "csv" ? "CSV" : "JSON"}</td>
                <td>{profile.lastUsedAt ? formatDateTime(profile.lastUsedAt, language) : "–"}</td>
                <td>
                  <span className="transfer-profile-actions">
                    {profile.enabled && (profile.direction === "import" ? canImport : canExport) ? (
                      <button
                        type="button"
                        className="icon-button transfer-profile-run"
                        aria-label={`${profile.name} ${t("importExport.dashboard.profiles.start")}`}
                        title={runTitle}
                        disabled={mutating}
                        onClick={(event) => {
                          event.stopPropagation();
                          onRun(profile);
                        }}
                      >
                        <RunIcon size={16} aria-hidden="true" />
                      </button>
                    ) : null}
                    {canEdit && (
                      <button
                        type="button"
                        className="icon-button transfer-row-menu"
                        aria-label={`${profile.name} ${t("importExport.dashboard.profiles.edit")}`}
                        title={t("importExport.dashboard.profiles.edit")}
                        onClick={(event) => {
                          event.stopPropagation();
                          onEdit(profile.id);
                        }}
                      >
                        <MoreVertical size={16} aria-hidden="true" />
                      </button>
                    )}
                  </span>
                </td>
              </tr>;
            })}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function editFromKeyboard(
  event: KeyboardEvent<HTMLTableRowElement>,
  profileId: string,
  onEdit: (profileId: string) => void
) {
  if (event.key !== "Enter" && event.key !== " ") return;
  event.preventDefault();
  onEdit(profileId);
}

function areaLabels(areas: DataTransferArea[], t: Translate) {
  return areas.map((area) => t(`importExport.dashboard.area.${area}`)).join(", ");
}
