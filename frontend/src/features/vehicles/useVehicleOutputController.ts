import type { Dispatch, FormEvent, SetStateAction } from "react";
import { useState } from "react";

import { api, type CreateVehicleRequest, type ExhibitionEntry, type Vehicle } from "../../shared/api";
import {
  buildBrandedQrPngDataUrl,
  buildQrSvg,
  downloadQrPngFile,
  downloadQrSvgFile,
  printQrSvgLabel,
  qrPayload
} from "./vehicleQr";
import {
  type InventoryReportAssets,
  inventoryReportHtml,
  openPrintDocument,
  reservePrintDocument
} from "./vehicleReports";
import { normalizedText, vehicleExhibitionEligible, vehicleToExhibitionEntry, vehicleToForm } from "./vehicleTransforms";
import type {
  ExhibitionAssignment,
  InventoryReportMode,
  InventoryReportSelection,
  SortDirection,
  SortKey
} from "./vehicleViewModel";
import { hasQrPayloadData } from "./vehicleViewModel";

type Translator = (key: string, values?: Record<string, string | number>) => string;

type UseVehicleOutputControllerOptions = {
  username: string;
  form: CreateVehicleRequest;
  selected: Vehicle | null;
  query: string;
  sort: { key: SortKey; direction: SortDirection };
  sortedVehicles: Vehicle[];
  selectedVisibleVehicles: Vehicle[];
  setVehicles: Dispatch<SetStateAction<Vehicle[]>>;
  setSelectedDetail: (vehicle: Vehicle) => void;
  onMessage: (message: string) => void;
  t: Translator;
};

export function useVehicleOutputController({
  username,
  form,
  selected,
  query,
  sort,
  sortedVehicles,
  selectedVisibleVehicles,
  setVehicles,
  setSelectedDetail,
  onMessage,
  t
}: UseVehicleOutputControllerOptions) {
  const [qrDialogOpen, setQrDialogOpen] = useState(false);
  const [qrSvg, setQrSvg] = useState("");
  const [qrError, setQrError] = useState("");
  const [reportDialogOpen, setReportDialogOpen] = useState(false);
  const [reportMode, setReportMode] = useState<InventoryReportMode>("summary");
  const [reportTitle, setReportTitle] = useState("Fahrzeugsammlung");
  const [reportSelection, setReportSelection] = useState<InventoryReportSelection>("all");
  const [reportIncludeQRCode, setReportIncludeQRCode] = useState(true);
  const [reportIncludeImages, setReportIncludeImages] = useState(true);
  const [reportCreating, setReportCreating] = useState(false);
  const [exhibitionAssignment, setExhibitionAssignment] = useState<ExhibitionAssignment | null>(null);

  const generateQr = async () => {
    if (!hasQrPayloadData(selected, form)) {
      setQrDialogOpen(false);
      setQrSvg("");
      setQrError("");
      onMessage(t("vehicles.qr.missingInput"));
      return;
    }

    setQrDialogOpen(true);
    setQrSvg("");
    setQrError("");
    try {
      setQrSvg(await buildQrSvg(selected, form));
    } catch (error) {
      setQrError(error instanceof Error ? error.message : "QR-Code konnte nicht erstellt werden.");
    }
  };

  const downloadQrSvg = () => {
    downloadQrSvgFile(qrSvg, form.inventoryNumber || "railkeeper");
  };

  const downloadQrPng = async () => {
    await downloadQrPngFile(qrPayload(selected, form), form.inventoryNumber || "railkeeper");
  };

  const printQr = () => {
    try {
      printQrSvgLabel(qrSvg, form);
    } catch (error) {
      setQrError(error instanceof Error ? error.message : "Druckfenster konnte nicht geöffnet werden.");
    }
  };

  const buildInventoryReportAssets = async (reportVehicles: Vehicle[], includeQRCode = reportIncludeQRCode) => {
    const assets: InventoryReportAssets = {};
    if (!includeQRCode) return assets;

    await Promise.all(
      reportVehicles.map(async (vehicle) => {
        assets[vehicle.id] = {
          qrCode: await buildBrandedQrPngDataUrl(qrPayload(vehicle, vehicleToForm(vehicle)), 192)
        };
      })
    );
    return assets;
  };

  const loadCompleteReportVehicles = (reportVehicles: Vehicle[]) => {
    return Promise.all(reportVehicles.map((vehicle) => api.vehicle(vehicle.id)));
  };

  const createInventoryReport = async (event?: FormEvent) => {
    event?.preventDefault();
    const reportVehicles = reportSelection === "selected" ? selectedVisibleVehicles : sortedVehicles;
    const reportName = `railkeeper-inventory-${reportMode}`;

    if (reportVehicles.length === 0) {
      onMessage("Es gibt keine Fahrzeuge für den PDF-Report.");
      return;
    }

    setReportCreating(true);
    const reportWindow = reservePrintDocument(reportName, reportTitle.trim() || "Fahrzeugsammlung");
    try {
      const completeReportVehicles = await loadCompleteReportVehicles(reportVehicles);
      const assets = await buildInventoryReportAssets(completeReportVehicles);
      const html = inventoryReportHtml(completeReportVehicles, query, sort, {
        mode: reportMode,
        title: reportTitle.trim() || "Fahrzeugsammlung",
        includeQRCode: reportIncludeQRCode,
        includeImages: reportIncludeImages
      }, assets);
      openPrintDocument(html, reportName, reportWindow);
      setReportDialogOpen(false);
    } catch (error) {
      onMessage(error instanceof Error ? error.message : "Report konnte nicht erstellt werden.");
    } finally {
      setReportCreating(false);
    }
  };

  const printVehicleReport = async (vehicle: Vehicle) => {
    const reportName = `railkeeper-vehicle-${vehicle.id}`;
    const reportWindow = reservePrintDocument(
      reportName,
      vehicle.name || vehicle.inventoryNumber || "Fahrzeugsammlung"
    );

    try {
      const completeVehicle = await api.vehicle(vehicle.id);
      const assets = await buildInventoryReportAssets([completeVehicle], true);
      const html = inventoryReportHtml(
        [completeVehicle],
        completeVehicle.inventoryNumber || completeVehicle.name,
        sort,
        {
          mode: "details",
          title: completeVehicle.name || completeVehicle.inventoryNumber || "Fahrzeugsammlung",
          includeQRCode: true,
          includeImages: true
        },
        assets
      );
      openPrintDocument(html, reportName, reportWindow);
    } catch (error) {
      onMessage(error instanceof Error ? error.message : "Report konnte nicht erstellt werden.");
    }
  };

  const updateVehicleExhibitionFlag = (vehicle: Vehicle, exhibition: boolean) => {
    onMessage("");
    return api.updateVehicle(vehicle.id, { ...vehicleToForm(vehicle), exhibition }).then((updated) => {
      const nextVehicle = {
        ...updated,
        images: updated.images ?? vehicle.images,
        attachments: updated.attachments ?? vehicle.attachments
      };
      setVehicles((current) => current.map((item) => (item.id === nextVehicle.id ? nextVehicle : item)));
      if (selected?.id === nextVehicle.id) setSelectedDetail(nextVehicle);
      return nextVehicle;
    });
  };

  const loadAssignmentEntries = (listID: string) => {
    if (!listID) {
      setExhibitionAssignment((current) => current
        ? { ...current, selectedListID: "", entries: [], loadingEntries: false, error: "" }
        : current);
      return;
    }

    setExhibitionAssignment((current) => current
      ? { ...current, selectedListID: listID, loadingEntries: true, error: "" }
      : current);
    api.exhibitionEntries(listID)
      .then((entries) => {
        setExhibitionAssignment((current) => current && current.selectedListID === listID
          ? { ...current, entries, loadingEntries: false, error: "" }
          : current);
      })
      .catch((error: Error) => {
        setExhibitionAssignment((current) => current
          ? { ...current, entries: [], loadingEntries: false, error: error.message }
          : current);
      });
  };

  const openExhibitionAssignment = (vehicle: Vehicle) => {
    if (!vehicleExhibitionEligible(vehicle)) {
      onMessage(t("vehicles.exhibition.requiresDecoder"));
      return;
    }

    onMessage("");
    setExhibitionAssignment({
      vehicle,
      lists: [],
      selectedListID: "",
      entries: [],
      loadingEntries: true,
      saving: false,
      error: ""
    });
    api.exhibitionLists()
      .then((lists) => {
        const availableLists = lists.filter((list) => !list.locked);
        const firstListID = availableLists[0]?.id || "";
        setExhibitionAssignment((current) => current ? {
          ...current,
          lists: availableLists,
          selectedListID: firstListID,
          loadingEntries: Boolean(firstListID),
          error: firstListID ? "" : t("vehicles.exhibition.noOpenLists")
        } : current);
        if (!firstListID) return undefined;
        return api.exhibitionEntries(firstListID).then((entries) => {
          setExhibitionAssignment((current) => current && current.selectedListID === firstListID
            ? { ...current, entries, loadingEntries: false, error: "" }
            : current);
        });
      })
      .catch((error: Error) => {
        setExhibitionAssignment((current) => current
          ? { ...current, loadingEntries: false, error: error.message }
          : current);
      });
  };

  const duplicateAssignmentVehicle = exhibitionAssignment
    ? exhibitionAssignment.entries.find((entry) =>
        normalizedText(entry.owner) === normalizedText(username) &&
        normalizedText(entry.locomotiveName) === normalizedText(exhibitionAssignment.vehicle.name)
      )
    : undefined;
  const duplicateAssignmentDecoder: ExhibitionEntry | undefined = exhibitionAssignment?.vehicle.digitalDecoderNumber
    ? exhibitionAssignment.entries.find((entry) =>
        normalizedText(entry.decoderNumber) === normalizedText(exhibitionAssignment.vehicle.digitalDecoderNumber)
      )
    : undefined;

  const saveExhibitionAssignment = () => {
    if (!exhibitionAssignment || !exhibitionAssignment.selectedListID) return;
    if (duplicateAssignmentVehicle || duplicateAssignmentDecoder) return;

    setExhibitionAssignment((current) => current ? { ...current, saving: true, error: "" } : current);
    api.vehicle(exhibitionAssignment.vehicle.id)
      .then((detail) => api.createExhibitionEntry(
        exhibitionAssignment.selectedListID,
        vehicleToExhibitionEntry(detail, username)
      ).then(() => detail))
      .then((detail) => updateVehicleExhibitionFlag(detail, true))
      .then(() => {
        setExhibitionAssignment(null);
        onMessage(t("vehicles.exhibition.assigned"));
      })
      .catch((error: Error) => {
        setExhibitionAssignment((current) => current ? { ...current, saving: false, error: error.message } : current);
      });
  };

  const toggleVehicleExhibition = (vehicle: Vehicle, checked: boolean) => {
    if (checked) {
      openExhibitionAssignment(vehicle);
      return;
    }
    updateVehicleExhibitionFlag(vehicle, false)
      .then(() => onMessage(t("vehicles.exhibition.disabled")))
      .catch((error: Error) => onMessage(error.message));
  };

  const openQrForVehicle = (vehicle: Vehicle) => {
    setQrSvg("");
    setQrError("");
    api.vehicle(vehicle.id)
      .then(async (detail) => {
        setSelectedDetail(detail);
        setQrDialogOpen(true);
        setQrSvg(await buildQrSvg(detail, vehicleToForm(detail)));
      })
      .catch((error: Error) => setQrError(error.message));
  };

  return {
    qr: {
      state: { open: qrDialogOpen, svg: qrSvg, error: qrError },
      commands: { close: () => setQrDialogOpen(false), generate: generateQr, downloadPng: downloadQrPng, downloadSvg: downloadQrSvg, print: printQr, openForVehicle: openQrForVehicle }
    },
    report: {
      state: { open: reportDialogOpen, mode: reportMode, title: reportTitle, selection: reportSelection, includeQRCode: reportIncludeQRCode, includeImages: reportIncludeImages, creating: reportCreating },
      setters: { setOpen: setReportDialogOpen, setMode: setReportMode, setTitle: setReportTitle, setSelection: setReportSelection, setIncludeQRCode: setReportIncludeQRCode, setIncludeImages: setReportIncludeImages },
      commands: { create: createInventoryReport, printVehicle: printVehicleReport }
    },
    exhibition: {
      state: { assignment: exhibitionAssignment, duplicateVehicle: duplicateAssignmentVehicle, duplicateDecoder: duplicateAssignmentDecoder },
      commands: { close: () => setExhibitionAssignment(null), loadEntries: loadAssignmentEntries, save: saveExhibitionAssignment, toggle: toggleVehicleExhibition }
    }
  };
}
