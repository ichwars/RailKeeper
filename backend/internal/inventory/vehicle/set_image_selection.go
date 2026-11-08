package vehicle

import (
	"context"
	"net/url"
	"strings"
)

func mainImageFromVehicleImage(source string, image *VehicleImage) *VehicleSetMainImage {
	return &VehicleSetMainImage{
		Source: source, ImageID: image.ID, VehicleID: image.VehicleID, URL: image.URL,
		ThumbnailURL: image.ThumbnailURL, Title: image.Title,
	}
}
func SetImageURLs(setID, version string) (string, string) {
	suffix := ""
	if version = strings.TrimSpace(version); version != "" {
		suffix = "?v=" + url.QueryEscape(version)
	}
	base := "/api/v1/vehicle-sets/" + setID + "/image/"
	return base + "file" + suffix, base + "thumbnail" + suffix
}
func distinctSetImageIDs(values ...string) []string {
	result := []string{}
	seen := map[string]struct{}{}
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
func (s *SetImageService) GetSetImage(ctx context.Context, setID string) (*VehicleSetImage, error) {
	image, err := s.repository.DedicatedImage(ctx, strings.TrimSpace(setID))
	if err != nil {
		return nil, err
	}
	if image.BlobID == "" {
		return nil, ErrVehicleSetImageNotFound
	}
	image.URL, image.ThumbnailURL = SetImageURLs(strings.TrimSpace(setID), image.BlobID)
	return &image, nil
}
func ResolveSetImage(
	ctx context.Context, reader SetImageLookup,
	setID string,
) (VehicleSetMainImageMode, string, *VehicleSetImage, *VehicleSetMainImage, error) {
	selection, err := reader.Selection(ctx, setID)
	if err != nil {
		return "", "", nil, nil, err
	}
	mode, memberImageID := selection.Mode, selection.MemberImageID
	fileName, mimeType, blobID, thumbnailBlobID, createdAt, updatedAt := selection.Dedicated.FileName, selection.Dedicated.MimeType, selection.Dedicated.BlobID, selection.Dedicated.ThumbnailBlobID, selection.Dedicated.CreatedAt, selection.Dedicated.UpdatedAt
	var dedicated *VehicleSetImage
	if blobID != "" {
		fileURL, thumbnailURL := SetImageURLs(setID, blobID)
		dedicated = &VehicleSetImage{
			URL:          fileURL,
			ThumbnailURL: thumbnailURL,
			FileName:     fileName, MimeType: mimeType, BlobID: blobID, ThumbnailBlobID: thumbnailBlobID,
			CreatedAt: createdAt, UpdatedAt: updatedAt,
		}
	}
	if mode == VehicleSetMainImageModeDedicated && dedicated != nil {
		return mode, memberImageID, dedicated, &VehicleSetMainImage{
			Source: "dedicated", URL: dedicated.URL, ThumbnailURL: dedicated.ThumbnailURL,
		}, nil
	}
	if mode == VehicleSetMainImageModeMember && memberImageID != "" {
		image, err := reader.MemberImage(ctx, setID, memberImageID)
		if err == nil && image != nil {
			return mode, memberImageID, dedicated, mainImageFromVehicleImage("member", image), nil
		}
		if err != nil {
			return "", "", nil, nil, err
		}
	}
	automatic, err := reader.AutomaticImage(ctx, setID)
	if err != nil {
		return "", "", nil, nil, err
	}
	if automatic == nil {
		return mode, memberImageID, dedicated, nil, nil
	}
	return mode, memberImageID, dedicated, mainImageFromVehicleImage("automatic", automatic), nil
}
func (s *SetImageService) LoadMainImages(ctx context.Context, ids []string) (map[string]*VehicleSetMainImage, error) {
	result := make(map[string]*VehicleSetMainImage, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	rows, err := s.repository.Candidates(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.Mode == VehicleSetMainImageModeDedicated && row.DedicatedBlobID != "" {
			fileURL, thumbnailURL := SetImageURLs(row.SetID, row.DedicatedBlobID)
			result[row.SetID] = &VehicleSetMainImage{Source: "dedicated", URL: fileURL, ThumbnailURL: thumbnailURL}
			continue
		}
		if row.Image == nil {
			continue
		}
		image := WithImageURLs(*row.Image)
		source := "automatic"
		if row.Mode == VehicleSetMainImageModeMember && image.ID == row.MemberImageID {
			source = "member"
		}
		result[row.SetID] = mainImageFromVehicleImage(source, &image)
	}
	return result, nil
}
