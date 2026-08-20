import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import type {
  DataTransferArtifact,
  DataTransferIssue,
  DataTransferJob,
  DataTransferPreview,
  DataTransferProfile,
  DataTransferSummary
} from "./dataTransferModel";
import { ImportExportView } from "./ImportExportView";

const importProfile = profileFixture({
  id: "profile-import",
  name: "Fahrzeugimport",
  direction: "import",
  format: "csv",
  areas: ["vehicles"]
});

const exhibitionProfile = profileFixture({
  id: "profile-exhibition",
  name: "Messeliste Köln",
  areas: ["exhibitionLists"],
  options: { includeImages: false }
});

const exportProfile = profileFixture({
  id: "profile-export",
  name: "Werkstattbestand",
  areas: ["vehicles", "accessories"],
  options: { includeImages: true, notes: "Quartalsarchiv" }
});

const draftImportJob = jobFixture({
  id: "job-import",
  profileId: importProfile.id,
  profileName: importProfile.name,
  direction: "import",
  format: "csv",
  areas: ["vehicles"]
});

const importIssue: DataTransferIssue = {
  id: "issue-1",
  jobId: draftImportJob.id,
  area: "vehicles",
  recordKey: "RK-1001",
  rowNumber: 2,
  field: "inventoryNumber",
  severity: "error",
  code: "duplicate_inventory_number",
  message: "Inventarnummer ist bereits vorhanden.",
  proposedResolution: "replace_or_copy",
  selectedResolution: "",
  createdAt: "2026-08-20T08:00:00Z",
  updatedAt: "2026-08-20T08:00:00Z"
};

const previewFixture: DataTransferPreview = {
  job: jobFixture({
    ...draftImportJob,
    state: "review_required",
    stage: "review",
    sourceName: "fahrzeuge.csv",
    sourceSha256: "sha-preview",
    revision: 2,
    totalRecords: 1,
    errorRecords: 1
  }),
  records: [{
    area: "vehicles",
    recordKey: "RK-1001",
    rowNumber: 2,
    classification: "error",
    proposedAction: "replace",
    targetId: "vehicle-1",
    data: { inventoryNumber: "RK-1001", manufacturer: "Märklin", name: "BR 103" }
  }],
  issues: [importIssue],
  totalRecords: 1,
  readyRecords: 0,
  warningRecords: 0,
  errorRecords: 1
};

const summaryFixture: DataTransferSummary = {
  openJobs: 0,
  selectedRecords: 0,
  lastExportAt: "",
  artifactCount: 0,
  artifactBytes: 0,
  openFolderAvailable: false,
  artifactDirectory: "C:\\RailKeeper\\Exporte"
};

describe("data transfer operational dialogs", () => {
  beforeEach(() => {
    vi.spyOn(api, "dataTransferSummary").mockResolvedValue(summaryFixture);
    vi.spyOn(api, "dataTransferProfiles").mockResolvedValue([importProfile, exhibitionProfile, exportProfile]);
    vi.spyOn(api, "dataTransferJobs").mockResolvedValue([]);
    vi.spyOn(api, "createDataTransferProfile").mockImplementation(async (input) => profileFixture(input));
    vi.spyOn(api, "updateDataTransferProfile").mockImplementation(async (id, input) => profileFixture({ id, ...input }));
    vi.spyOn(api, "createDataTransferImportJob").mockResolvedValue(draftImportJob);
    vi.spyOn(api, "uploadDataTransferImport").mockResolvedValue(previewFixture);
    vi.spyOn(api, "resolveDataTransferIssue").mockResolvedValue(jobFixture({
      ...previewFixture.job,
      state: "ready",
      revision: 3,
      readyRecords: 1,
      errorRecords: 0
    }));
    vi.spyOn(api, "confirmDataTransferImport").mockResolvedValue(jobFixture({
      ...previewFixture.job,
      state: "completed",
      stage: "completed",
      readyRecords: 1,
      errorRecords: 0
    }));
    vi.spyOn(api, "createDataTransferExportJob").mockImplementation(async ({ profileId }) => {
      const profile = [exhibitionProfile, exportProfile].find((item) => item.id === profileId) || exportProfile;
      return jobFixture({
        id: "job-export",
        profileId: profile.id,
        profileName: profile.name,
        direction: "export",
        format: profile.format,
        areas: profile.areas,
        options: profile.options
      });
    });
    vi.spyOn(api, "executeDataTransferExportJob").mockResolvedValue({
      job: jobFixture({
        id: "job-export",
        profileId: exportProfile.id,
        profileName: exportProfile.name,
        direction: "export",
        format: exportProfile.format,
        areas: exportProfile.areas,
        options: exportProfile.options,
        state: "completed",
        stage: "completed",
        sourceName: "railkeeper-transfer.json",
        totalRecords: 37,
        readyRecords: 37
      }),
      artifact: artifactFixture(),
      openFolderAvailable: false
    });
    vi.spyOn(api, "deleteDataTransferArtifact").mockResolvedValue(undefined);
  });

  it("disables CSV for exhibition lists and explains why", async () => {
    render(<ImportExportView roles={["Admin"]} />);

    fireEvent.click(await screen.findByRole("button", { name: "Profil anlegen" }));
    const dialog = screen.getByRole("dialog", { name: "Transferprofil anlegen" });
    fireEvent.click(within(dialog).getByRole("checkbox", { name: "Fahrzeuge" }));
    fireEvent.click(within(dialog).getByRole("checkbox", { name: "Ausstellungslisten" }));

    expect(within(dialog).getByRole("radio", { name: "CSV" })).toBeDisabled();
    expect(within(dialog).getByText("Ausstellungslisten sind nur als JSON verfügbar.")).toBeInTheDocument();
  });

  it("keeps the uploaded file and server preview visible for review", async () => {
    const user = userEvent.setup();
    render(<ImportExportView roles={["Editor"]} />);

    fireEvent.click(await screen.findByRole("button", { name: "Neuer Import" }));
    const dialog = screen.getByRole("dialog", { name: "Import prüfen" });
    await user.selectOptions(within(dialog).getByLabelText("Importprofil"), importProfile.id);
    await user.upload(within(dialog).getByLabelText("Importdatei"), new File(["a;b"], "fahrzeuge.csv", {
      type: "text/csv"
    }));

    expect(await within(dialog).findByRole("heading", { name: "Vorschau" })).toBeInTheDocument();
    expect(within(dialog).getByText("fahrzeuge.csv")).toBeInTheDocument();
    expect(within(dialog).getByText("RK-1001")).toBeInTheDocument();
    expect(api.uploadDataTransferImport).toHaveBeenCalledWith("job-import", expect.any(File));
  });

  it("hydrates a persisted preview when job details finish loading after the dialog opens", async () => {
    const persistedJob = jobFixture({
      ...previewFixture.job,
      state: "ready",
      readyRecords: 1,
      errorRecords: 0,
      preview: { records: previewFixture.records }
    });
    let resolveDetails: ((details: { job: DataTransferJob; issues: DataTransferIssue[]; artifacts: DataTransferArtifact[] }) => void)
      | undefined;
    vi.mocked(api.dataTransferJobs).mockResolvedValue([persistedJob]);
    vi.spyOn(api, "dataTransferJob").mockImplementation(() => new Promise((resolve) => {
      resolveDetails = resolve;
    }));

    render(<ImportExportView roles={["Editor"]} />);
    fireEvent.click(await screen.findByRole("button", { name: "Import bestätigen" }));
    const dialog = screen.getByRole("dialog", { name: "Import prüfen" });
    expect(within(dialog).queryByRole("heading", { name: "Vorschau" })).not.toBeInTheDocument();

    resolveDetails?.({ job: persistedJob, issues: [], artifacts: [] });

    expect(await within(dialog).findByRole("heading", { name: "Vorschau" })).toBeInTheDocument();
    expect(within(dialog).getByText("RK-1001")).toBeInTheDocument();
  });

  it("blocks unresolved errors and only confirms after an explicit resolution", async () => {
    const user = userEvent.setup();
    render(<ImportExportView roles={["Editor"]} />);

    fireEvent.click(await screen.findByRole("button", { name: "Neuer Import" }));
    const dialog = screen.getByRole("dialog", { name: "Import prüfen" });
    await user.selectOptions(within(dialog).getByLabelText("Importprofil"), importProfile.id);
    await user.upload(within(dialog).getByLabelText("Importdatei"), new File(["a;b"], "fahrzeuge.csv", {
      type: "text/csv"
    }));

    const blockedConfirm = await within(dialog).findByRole("button", { name: "0 Datensätze importieren" });
    expect(blockedConfirm).toBeDisabled();
    expect(api.confirmDataTransferImport).not.toHaveBeenCalled();

    await user.selectOptions(within(dialog).getByLabelText("Auflösung für RK-1001"), "replace");
    const enabledConfirm = await within(dialog).findByRole("button", { name: "1 Datensatz importieren" });
    expect(enabledConfirm).toBeEnabled();
    await user.click(enabledConfirm);

    await waitFor(() => expect(api.confirmDataTransferImport).toHaveBeenCalledWith("job-import"));
  });

  it("retries into a fresh draft flow without confirming the historical import", async () => {
    const failedJob = jobFixture({
      id: "job-failed",
      profileId: "",
      profileName: "Fehlgeschlagener Fahrzeugimport",
      direction: "import",
      format: "csv",
      areas: ["vehicles"],
      state: "failed",
      stage: "failed"
    });
    const retry = jobFixture({
      ...failedJob,
      id: "job-retry",
      state: "draft",
      stage: "created",
      revision: 1
    });
    vi.mocked(api.dataTransferJobs).mockResolvedValue([failedJob]);
    vi.spyOn(api, "dataTransferJob").mockResolvedValue({ job: failedJob, issues: [], artifacts: [] });
    vi.spyOn(api, "retryDataTransferJob").mockResolvedValue(retry);

    render(<ImportExportView roles={["Editor"]} />);
    fireEvent.click(await screen.findByRole("button", { name: "Erneut versuchen" }));

    const retryDialog = await screen.findByRole("dialog", { name: "Import prüfen" });
    expect(within(retryDialog).getByText("Fehlgeschlagener Fahrzeugimport")).toBeInTheDocument();
    expect(api.retryDataTransferJob).toHaveBeenCalledWith("job-failed");
    expect(api.confirmDataTransferImport).not.toHaveBeenCalled();
  });

  it("shows export areas, format, options, and the final exported count", async () => {
    const user = userEvent.setup();
    render(<ImportExportView roles={["Viewer"]} />);

    fireEvent.click(await screen.findByRole("button", { name: "Neuer Export" }));
    const dialog = screen.getByRole("dialog", { name: "Export erstellen" });
    await user.selectOptions(within(dialog).getByLabelText("Exportprofil"), exportProfile.id);

    expect(within(dialog).getByText("Fahrzeuge, Zubehör")).toBeInTheDocument();
    expect(within(dialog).getByText("JSON")).toBeInTheDocument();
    expect(within(dialog).getByText("includeImages: ja")).toBeInTheDocument();
    await user.click(within(dialog).getByRole("button", { name: "Export ausführen" }));

    expect(await within(dialog).findByText("37 Datensätze exportiert")).toBeInTheDocument();
    expect(api.createDataTransferExportJob).toHaveBeenCalledWith({ profileId: exportProfile.id });
    expect(api.executeDataTransferExportJob).toHaveBeenCalledWith("job-export");
  });

  it("updates a profile without rewriting the historical job snapshot", async () => {
    const user = userEvent.setup();
    const historicalJob = jobFixture({
      id: "job-history",
      profileId: exportProfile.id,
      profileName: "Werkstattbestand (alt)",
      direction: "export",
      state: "completed",
      stage: "completed",
      sourceName: "railkeeper-transfer.json"
    });
    vi.mocked(api.dataTransferJobs).mockResolvedValue([historicalJob]);
    vi.spyOn(api, "dataTransferJob").mockResolvedValue({ job: historicalJob, issues: [], artifacts: [] });

    render(<ImportExportView roles={["Editor"]} />);
    fireEvent.click(await screen.findByRole("button", { name: "Werkstattbestand bearbeiten" }));
    const dialog = screen.getByRole("dialog", { name: "Transferprofil bearbeiten" });
    const name = within(dialog).getByLabelText("Profilname");
    await user.clear(name);
    await user.type(name, "Werkstattbestand neu");
    await user.click(within(dialog).getByRole("button", { name: "Änderungen speichern" }));

    await waitFor(() => expect(api.updateDataTransferProfile).toHaveBeenCalledWith(
      exportProfile.id,
      expect.objectContaining({ name: "Werkstattbestand neu" })
    ));
    expect((await screen.findAllByText("Werkstattbestand (alt)")).length).toBeGreaterThan(0);
  });

  it("requires confirmation before deleting a local artifact", async () => {
    const completedJob = jobFixture({
      id: "job-completed",
      profileName: "Abgeschlossener Export",
      state: "completed",
      stage: "completed",
      sourceName: "railkeeper-transfer.json"
    });
    vi.mocked(api.dataTransferJobs).mockResolvedValue([completedJob]);
    vi.spyOn(api, "dataTransferJob").mockResolvedValue({
      job: completedJob,
      issues: [],
      artifacts: [artifactFixture({ jobId: completedJob.id })]
    });
    const confirm = vi.spyOn(window, "confirm").mockReturnValue(true);

    render(<ImportExportView roles={["Admin"]} />);
    fireEvent.click(await screen.findByRole("button", { name: "railkeeper-transfer.json löschen" }));

    expect(confirm).toHaveBeenCalledWith(expect.stringContaining("railkeeper-transfer.json"));
    await waitFor(() => expect(api.deleteDataTransferArtifact).toHaveBeenCalledWith("artifact-1"));
  });
});

function profileFixture(overrides: Partial<DataTransferProfile> = {}): DataTransferProfile {
  return {
    id: "profile-1",
    name: "Exportprofil",
    direction: "export",
    format: "railkeeper-json",
    areas: ["vehicles"],
    options: {},
    enabled: true,
    createdByUserId: "user-1",
    createdAt: "2026-08-20T08:00:00Z",
    updatedAt: "2026-08-20T08:00:00Z",
    ...overrides
  };
}

function jobFixture(overrides: Partial<DataTransferJob> = {}): DataTransferJob {
  return {
    id: "job-1",
    profileId: "profile-1",
    profileName: "Exportprofil",
    direction: "export",
    format: "railkeeper-json",
    areas: ["vehicles"],
    options: {},
    state: "draft",
    stage: "created",
    sourceName: "",
    sourceSha256: "",
    packageVersion: 1,
    revision: 1,
    totalRecords: 0,
    readyRecords: 0,
    warningRecords: 0,
    errorRecords: 0,
    preview: {},
    createdByUserId: "user-1",
    confirmedByUserId: "",
    confirmedAt: "",
    completedAt: "",
    resultMessage: "",
    createdAt: "2026-08-20T08:00:00Z",
    updatedAt: "2026-08-20T08:00:00Z",
    ...overrides
  };
}

function artifactFixture(overrides: Partial<DataTransferArtifact> = {}): DataTransferArtifact {
  return {
    id: "artifact-1",
    jobId: "job-export",
    relativePath: "exports/railkeeper-transfer.json",
    displayName: "railkeeper-transfer.json",
    mimeType: "application/json",
    sizeBytes: 4096,
    sha256: "artifact-sha",
    deletedAt: "",
    createdAt: "2026-08-20T08:01:00Z",
    ...overrides
  };
}
