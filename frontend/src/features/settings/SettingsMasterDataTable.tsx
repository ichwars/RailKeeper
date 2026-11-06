import type { ReactNode } from "react";
import { ChevronDown, ChevronUp, ExternalLink } from "lucide-react";
import type { MasterDataEntry } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { MasterDataBulkEntryCheckbox, MasterDataBulkSelectAllCheckbox } from "./MasterDataBulkSelection";
import { MasterDataLifecycleActions } from "./MasterDataLifecycleActions";
import {
  cv8NameText,
  cv8DecimalText,
  cv8BinaryText,
  cv8HexText,
  cv8CountryText,
  externalLink,
  manufacturerWebsite,
  manufacturerNeedsWebsiteReview,
  manufacturerSearchDomainsText,
  masterDataImage,
  metadataString,
  nominalScalesText
} from "./settingsModel";
import type { SettingsMasterDataController } from "./useSettingsMasterData";

type MasterDataTableColumn = {
  key: string;
  label: string;
  align?: "left" | "center" | "right";
  wrap?: boolean;
  sortable?: boolean;
  sortValue?: (entry: MasterDataEntry) => string | number;
  render: (entry: MasterDataEntry) => ReactNode;
};

const compareMasterDataValues = (left: string | number | undefined, right: string | number | undefined) => {
  if (typeof left === "number" && typeof right === "number") {
    return left - right;
  }
  return String(left ?? "").localeCompare(String(right ?? ""), "de-DE", {
    numeric: true,
    sensitivity: "base"
  });
};

export function SettingsMasterDataTable({ controller }: { controller: SettingsMasterDataController }) {
  const { t } = useI18n();
  const {
    activeType, canEditMasterData, masterDataBatchBusy, search, masterDataSort, loading, isSymbolData,
    isManufacturerData, isCV8ManufacturerData, filteredItems, masterDataSelection, startEdit,
    setEntryActive, deactivateEntry, deleteEntry, toggleMasterDataSort
  } = controller;

  const masterDataColumns: MasterDataTableColumn[] = [];
  if (canEditMasterData) {
    masterDataColumns.push({
      key: "selection",
      label: "",
      align: "center",
      sortable: false,
      render: (entry) => <MasterDataBulkEntryCheckbox
        selection={masterDataSelection}
        entry={entry}
        label={isCV8ManufacturerData ? cv8NameText(entry) : entry.label}
        disabled={masterDataBatchBusy}
      />
    });
  }
  if (isSymbolData) {
    masterDataColumns.push({
      key: "symbol",
      label: t("settings.data.symbol"),
      align: "center",
      sortable: false,
      render: (entry) => {
        const symbolImage = masterDataImage(entry);
        return symbolImage ? <img className="master-data-symbol-preview" src={symbolImage} alt="" /> : "-";
      }
    });
  }
  masterDataColumns.push({
    key: "name",
    label: isCV8ManufacturerData ? t("settings.data.cv8Manufacturer") : t("settings.data.name"),
    sortValue: (entry) => isCV8ManufacturerData ? cv8NameText(entry) : entry.label,
    render: (entry) => (
      <>
        <strong>{isCV8ManufacturerData ? cv8NameText(entry) : entry.label}</strong>
        {isManufacturerData && manufacturerNeedsWebsiteReview(entry) && <span className="master-data-warning">{t("settings.data.wikiSourceWarning")}</span>}
      </>
    )
  });
  if (isManufacturerData) {
    masterDataColumns.push(
      {
        key: "nominal-scales",
        label: t("settings.data.nominalScales"),
        align: "center",
        sortValue: (entry) => nominalScalesText(entry),
        render: (entry) => nominalScalesText(entry) || "-"
      },
      {
        key: "manufacturer-website",
        label: t("settings.data.manufacturerWebsite"),
        wrap: true,
        sortValue: (entry) => manufacturerWebsite(entry),
        render: (entry) => manufacturerWebsite(entry) || "-"
      },
      {
        key: "search-domains",
        label: t("settings.data.searchDomains"),
        wrap: true,
        sortValue: (entry) => manufacturerSearchDomainsText(entry),
        render: (entry) => manufacturerSearchDomainsText(entry) || "-"
      }
    );
  }
  if (isCV8ManufacturerData) {
    masterDataColumns.push(
      {
        key: "cv8-decimal",
        label: t("settings.data.cv8Decimal"),
        align: "center",
        sortValue: (entry) => Number(cv8DecimalText(entry)) || 0,
        render: (entry) => <code>{cv8DecimalText(entry) || "-"}</code>
      },
      {
        key: "cv8-binary",
        label: t("settings.data.cv8Binary"),
        align: "center",
        sortValue: (entry) => cv8BinaryText(entry),
        render: (entry) => <code>{cv8BinaryText(entry) || "-"}</code>
      },
      {
        key: "cv8-hex",
        label: t("settings.data.cv8Hex"),
        align: "center",
        sortValue: (entry) => cv8HexText(entry),
        render: (entry) => <code>{cv8HexText(entry) || "-"}</code>
      },
      {
        key: "cv8-country",
        label: t("settings.data.cv8Country"),
        align: "center",
        sortValue: (entry) => cv8CountryText(entry),
        render: (entry) => cv8CountryText(entry) || "-"
      }
    );
  }
  if (!isSymbolData && !isCV8ManufacturerData) {
    masterDataColumns.push({
      key: "link",
      label: t("settings.data.link"),
      align: "center",
      sortValue: (entry) => isManufacturerData ? manufacturerWebsite(entry) : entry.sourceUrl || "",
      render: (entry) => {
        const website = manufacturerWebsite(entry);
        const link = isManufacturerData ? (website ? { href: website, title: t("settings.data.openManufacturerWebsite") } : null) : externalLink(entry);
        return link ? (
          <a className="table-icon-link" href={link.href} target="_blank" rel="noreferrer" aria-label={link.title} title={link.title}>
            <ExternalLink size={16} />
          </a>
        ) : "-";
      }
    });
  }
  if (isSymbolData) {
    masterDataColumns.push({
      key: "description",
      label: t("settings.data.description"),
      wrap: true,
      sortValue: (entry) => metadataString(entry, "description"),
      render: (entry) => metadataString(entry, "description") || "-"
    });
  }
  masterDataColumns.push(
    {
      key: "status",
      label: t("settings.master.status"),
      align: "center",
      sortValue: (entry) => entry.active ? 1 : 0,
      render: (entry) => t(entry.active ? "settings.articleManagement.active" :
        "settings.articleManagement.inactive")
    },
    {
      key: "origin",
      label: t("settings.master.origin"),
      align: "center",
      sortValue: (entry) => entry.origin || "custom",
      render: (entry) => t(`settings.master.origin.${entry.origin || "custom"}`)
    }
  );
  masterDataColumns.push({
    key: "actions",
    label: t("settings.data.actions"),
    align: "right",
    sortable: false,
    render: (entry) => (
      <MasterDataLifecycleActions
        entry={entry}
        displayLabel={isCV8ManufacturerData ? cv8NameText(entry) : entry.label}
        onEdit={() => startEdit(entry)}
        onDeactivate={() => deactivateEntry(entry)}
        onReactivate={() => setEntryActive(entry, true)}
        onDelete={() => deleteEntry(entry)}
        disabled={masterDataBatchBusy}
      />
    )
  });
  const activeSortColumn =
    masterDataColumns.find((column) => column.key === masterDataSort.key && column.sortable !== false && column.sortValue) ||
    masterDataColumns.find((column) => column.key === "name" && column.sortValue);
  const sortedItems = activeSortColumn?.sortValue
    ? [...filteredItems].sort((left, right) => {
      const result = compareMasterDataValues(activeSortColumn.sortValue?.(left), activeSortColumn.sortValue?.(right));
      return masterDataSort.direction === "asc" ? result : -result;
    })
    : filteredItems;

  return (
    <div className={`table-wrap master-data-table master-data-table-${activeType}`}>
      <table>
        <thead>
          <tr>
            {masterDataColumns.map((column) => {
              const isActiveSort = activeSortColumn?.key === column.key;
              const className = [
                `master-data-col-${column.key}`,
                column.align ? `is-${column.align}` : "",
                column.wrap ? "is-wrapped" : ""
              ].filter(Boolean).join(" ");
              return (
                <th
                  key={column.key}
                  className={className}
                  aria-sort={isActiveSort ? (masterDataSort.direction === "asc" ? "ascending" : "descending") : "none"}
                >
                  {column.key === "selection" ? (
                    <MasterDataBulkSelectAllCheckbox
                      selection={masterDataSelection}
                      disabled={masterDataBatchBusy}
                    />
                  ) : column.sortable === false || !column.sortValue ? (
                    column.label
                  ) : (
                    <button
                      type="button"
                      className={`sort-button ${isActiveSort ? "active" : ""}`}
                      onClick={() => toggleMasterDataSort(column.key)}
                    >
                      <span>{column.label}</span>
                      {isActiveSort && masterDataSort.direction === "asc" ? <ChevronUp size={13} /> : <ChevronDown size={13} />}
                    </button>
                  )}
                </th>
              );
            })}
          </tr>
        </thead>
        <tbody>
          {loading ? (
            <tr>
              <td colSpan={masterDataColumns.length} className="loading-cell">Lade aus lokaler Stammdatenbank...</td>
            </tr>
          ) : filteredItems.length === 0 ? (
            <tr>
              <td colSpan={masterDataColumns.length} className="loading-cell">{t("settings.data.noEntries")}</td>
            </tr>
          ) : (
            sortedItems.map((entry) => (
              <tr key={entry.id} className={entry.active ? "" : "muted-row"}>
                {masterDataColumns.map((column) => {
                  const className = [
                    `master-data-col-${column.key}`,
                    column.align ? `is-${column.align}` : "",
                    column.wrap ? "is-wrapped" : ""
                  ].filter(Boolean).join(" ");
                  return <td key={column.key} className={className}>{column.render(entry)}</td>;
                })}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}
