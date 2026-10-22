import type { ECoSImagePreview } from "../../shared/ecosImageModel";

export function sameECoSImageReview(previous: ECoSImagePreview, current: ECoSImagePreview) {
  return previous.host === current.host && previous.sha256 === current.sha256 &&
    previous.description === current.description && previous.slot === current.slot &&
    previous.existingImage === current.existingImage && previous.newImage === current.newImage &&
    previous.currentImage === current.currentImage && previous.currentImageMissing === current.currentImageMissing &&
    previous.target.objectId === current.target.objectId && previous.target.name === current.target.name &&
    previous.target.address === current.target.address && previous.target.protocol === current.target.protocol &&
    previous.current.locomotiveType === current.current.locomotiveType &&
    previous.current.imageType === current.current.imageType && previous.current.index === current.current.index;
}
