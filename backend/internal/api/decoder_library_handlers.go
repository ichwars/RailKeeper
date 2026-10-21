package api

import (
	"context"
	"net/http"

	"railkeeper/backend/internal/decoderlibrary"
)

func (a *App) decoderLibrary(w http.ResponseWriter, r *http.Request) {
	catalog, err := a.masterDataService.DecoderLibrary(r.Context(), r.URL.Query().Get("includeInactive") != "true")
	if err != nil {
		a.logger.Error("decoder library read failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "decoder_library_failed", "Decoder library could not be loaded.")
		return
	}
	respondJSON(w, http.StatusOK, catalog)
}

func (a *App) decoderReferences(w http.ResponseWriter, r *http.Request) {
	references, err := a.masterDataService.DecoderReferences(r.Context(), r.PathValue("modelID"), true)
	if err != nil {
		a.logger.Error("decoder CV references read failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "decoder_references_failed", "Decoder CV references could not be loaded.")
		return
	}
	respondJSON(w, http.StatusOK, references)
}

func (a *App) enrichDecoderPreview(ctx context.Context, preview *vehicleCVFilePreviewResponse) {
	if a.masterDataService == nil || len(preview.SuggestedCVValues) == 0 || preview.CVImportError != "" {
		return
	}
	catalog, err := a.masterDataService.DecoderLibrary(ctx, true)
	if err != nil {
		a.logger.Warn("decoder preview library unavailable", "error", err)
		return
	}
	groups := map[string][]decoderlibrary.Reading{}
	for _, value := range preview.SuggestedCVValues {
		profile := value.DecoderProfile
		if profile == "" {
			profile = preview.SuggestedDecoderProfile
		}
		groups[profile] = append(groups[profile], decoderlibrary.Reading{
			CVNumber: value.CVNumber, Value: value.Value, IndexCV31: value.IndexCV31, IndexCV32: value.IndexCV32,
			Protocol: value.Protocol, DecoderProfile: profile, Description: value.Description,
		})
	}
	preview.DecoderLibraryMatches = map[string]decoderlibrary.Match{}
	for profile, rows := range groups {
		preview.DecoderLibraryMatches[profile] = decoderlibrary.Resolve(catalog.Profiles, profile, rows, catalog.Inventory)
	}
}
