export type ECoSImageTarget = { objectId: number; name: string; address: number; protocol: string };
export type ECoSImageOptions = { host: string; locomotives: ECoSImageTarget[] };
export type ECoSImagePreview = {
  token: string;
  expiresAt: string;
  host: string;
  target: ECoSImageTarget;
  current: { locomotiveType: string; imageType: "IMAGE_TYPE_INT" | "IMAGE_TYPE_USER"; index: number };
  slot: number;
  description: string;
  sha256: string;
  currentImage: string;
  newImage: string;
  existingImage: boolean;
};
export type ECoSImageResult = {
  status: "verified" | "upload_unknown" | "image_not_stored" | "image_unverified" | "assignment_unknown" | "control_busy";
  slot: number;
  objectId: number;
  imageVerified: boolean;
  assignmentVerified: boolean;
  auditRecorded: boolean;
  liveMonitor: { wasRunning: boolean; restarted: boolean };
};
