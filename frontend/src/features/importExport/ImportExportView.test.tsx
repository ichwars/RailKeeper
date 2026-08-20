import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import type {
  DataTransferArtifact,
  DataTransferJob,
  DataTransferJobDetails,
  DataTransferProfile,
  DataTransferSummary
} from "./dataTransferModel";
import { ImportExportView } from "./ImportExportView";

const reviewJob = jobFixture({
  id: "job-review",
  profileId: "profile-import",
  profileName: "Fahrzeugimport",
  format: "csv",
  state: "review_required",
  stage: "review",
  sourceName: "fahrzeuge_august.csv",
  totalRecords: 48,
  readyRecords: 42,
  warningRecords: 5,
  errorRecords: 1
});

const exhibitionExportJob = jobFixture({
  id: "job-exhibition",
  profileId: "profile-exhibition",
  profileName: "Messeliste Köln",
  direction: "export",
  areas: ["exhibitionLists"],
  state: "completed",
  stage: "completed",
  sourceName: "messeliste-koeln.json",
  totalRecords: 37,
  readyRecords: 37,
  completedAt: "2026-08-20T12:32:00Z",
  createdAt: "2026-08-20T12:31:00Z"
});

const summaryFixture: DataTransferSummary = {
  openJobs: 3,
  selectedRecords: 429,
  lastExportAt: "2026-08-20T12:32:00Z",
  artifactCount: 12,
  artifactBytes: 84 * 1024 * 1024,
  openFolderAvailable: true,
  artifactDirectory: "C:\\RailKeeper\\Exporte"
};

const artifactFixture: DataTransferArtifact = {
  id: "artifact-review",
  jobId: reviewJob.id,
  relativePath: "exports/fahrzeuge_august.csv",
  displayName: "fahrzeuge_august.csv",
  mimeType: "text/csv",
  sizeBytes: 4096,
  sha256: "abc",
  deletedAt: "",
  createdAt: "2026-08-20T12:31:00Z"
};

describe("ImportExportView", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(api, "dataTransferSummary").mockResolvedValue(summaryFixture);
    vi.spyOn(api, "dataTransferProfiles").mockResolvedValue([
      profileFixture({
        id: "profile-backup",
        name: "Vollständige Sicherung",
        direction: "export",
        areas: ["vehicles", "accessories", "exhibitionLists"]
      }),
      profileFixture({ id: "profile-vehicles", name: "Fahrzeugliste für Excel", direction: "export", format: "csv" }),
      profileFixture({
        id: "profile-accessories",
        name: "Zubehör-Inventur",
        direction: "export",
        format: "csv",
        areas: ["accessories"]
      }),
      profileFixture({
        id: "profile-exhibition",
        name: "Messeliste kompakt",
        direction: "export",
        areas: ["exhibitionLists"]
      })
    ]);
    vi.spyOn(api, "dataTransferJobs").mockResolvedValue([reviewJob, exhibitionExportJob]);
    vi.spyOn(api, "dataTransferJob").mockImplementation(async (id) => detailsFixture(
      id === exhibitionExportJob.id ? exhibitionExportJob : reviewJob
    ));
    vi.spyOn(api, "openDataTransferArtifactFolder").mockResolvedValue(undefined);
  });

  it("renders the reference dashboard topology without master data", async () => {
    render(<ImportExportView roles={["Admin"]} />);

    expect(await screen.findByText("DATENTRANSFER")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Import/Export" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Neuer Import" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "Neuer Export" })).toBeEnabled();
    expect(screen.getByRole("heading", { name: "Aufträge" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Exportprofile" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Transferverlauf" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Auftragsdetails" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Lokale Ablage" })).toBeInTheDocument();
    expect(screen.getByText("Messeliste kompakt")).toBeInTheDocument();
    expect(screen.queryByText("Stammdaten")).not.toBeInTheDocument();
  });

  it("updates job details and progress stages after selection and applies job filters", async () => {
    render(<ImportExportView roles={["Admin"]} />);

    expect(await screen.findByRole("heading", { name: "Fahrzeugimport" })).toBeInTheDocument();
    expect(screen.getByText("Zuordnung")).toBeInTheDocument();
    fireEvent.click(screen.getByTitle("Messeliste Köln"));

    expect(await screen.findByRole("heading", { name: "Messeliste Köln" })).toBeInTheDocument();
    expect(screen.getByText("Vorbereitung")).toBeInTheDocument();
    expect(screen.getByRole("progressbar", { name: "Fortschritt Messeliste Köln" })).toHaveAttribute(
      "aria-valuenow",
      "100"
    );

    fireEvent.click(screen.getByRole("button", { name: "Erledigt" }));
    expect(screen.getByRole("button", { name: "Erledigt" })).toHaveAttribute("aria-pressed", "true");
    await waitFor(() => expect(api.dataTransferJobs).toHaveBeenLastCalledWith({
      states: ["completed", "completed_with_warnings", "failed", "cancelled"],
      limit: 100
    }));
  });

  it("opens the local artifact folder when the capability is available", async () => {
    render(<ImportExportView roles={["Admin"]} />);

    fireEvent.click(await screen.findByRole("button", { name: "Ordner öffnen" }));

    await waitFor(() => expect(api.openDataTransferArtifactFolder).toHaveBeenCalledOnce());
  });

  it("shows the artifact directory and downloads without an unavailable folder action", async () => {
    vi.mocked(api.dataTransferSummary).mockResolvedValue({ ...summaryFixture, openFolderAvailable: false });
    render(<ImportExportView roles={["Admin"]} />);

    expect(await screen.findByText("C:\\RailKeeper\\Exporte")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Ordner öffnen" })).not.toBeInTheDocument();
    expect(await screen.findByRole("link", { name: "fahrzeuge_august.csv herunterladen" })).toHaveAttribute(
      "href",
      "/api/v1/data-transfer/artifacts/artifact-review/download"
    );
  });
});

function profileFixture(overrides: Partial<DataTransferProfile> = {}): DataTransferProfile {
  return {
    id: "profile-import",
    name: "Fahrzeugimport",
    direction: "import",
    format: "railkeeper-json",
    areas: ["vehicles"],
    options: {},
    enabled: true,
    createdByUserId: "user-1",
    lastUsedAt: "2026-08-20T12:32:00Z",
    createdAt: "2026-08-20T08:00:00Z",
    updatedAt: "2026-08-20T08:00:00Z",
    ...overrides
  };
}

function jobFixture(overrides: Partial<DataTransferJob> = {}): DataTransferJob {
  return {
    id: "job-1",
    profileId: "profile-import",
    profileName: "Fahrzeugimport",
    direction: "import",
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

function detailsFixture(job: DataTransferJob): DataTransferJobDetails {
  return {
    job,
    issues: [],
    artifacts: job.id === reviewJob.id ? [artifactFixture] : [{
      ...artifactFixture,
      id: "artifact-exhibition",
      jobId: exhibitionExportJob.id,
      displayName: "messeliste-koeln.json"
    }]
  };
}
