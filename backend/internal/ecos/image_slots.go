package ecos

import (
	"context"
	"fmt"
	"net/http"
)

// Check all advertised slots before selecting a free one. An identical image
// beyond the first free slot must be reused, including after a restart or lost reply.
func (web *ImageWeb) FindImageSlot(ctx context.Context, source []byte) (int, bool, error) {
	if err := ValidateImageBMP(source); err != nil {
		return 0, false, err
	}
	states := map[int]bool{}
	pages, visited := map[int]bool{0: true}, map[int]bool{}
	for range 10 {
		page := -1
		for candidate := 0; candidate < 10; candidate++ {
			if pages[candidate] && !visited[candidate] {
				page = candidate
				break
			}
		}
		if page < 0 {
			break
		}
		visited[page] = true
		data, err := web.request(ctx, http.MethodGet, imageListPath(page), "", nil, 256<<10)
		if err != nil {
			return 0, false, fmt.Errorf("%w: page %d: %w", ErrImageInventory, page, err)
		}
		entries, available, err := parseImagePage(data, page)
		if err != nil {
			return 0, false, fmt.Errorf("%w: page %d: %w", ErrImageInventory, page, err)
		}
		for index, occupied := range entries {
			states[index] = occupied
		}
		for next := range available {
			pages[next] = true
		}
	}
	free := -1
	for index := 0; index < 500; index++ {
		occupied, listed := states[index]
		if !listed {
			continue
		}
		if !occupied {
			if free < 0 {
				free = index
			}
			continue
		}
		actual, err := web.ReadImage(ctx, "user", index)
		if err != nil {
			return 0, false, err
		}
		if VerifyImage(source, actual) {
			return index, true, nil
		}
	}
	if free < 0 {
		return 0, false, ErrImageSlotsFull
	}
	return free, false, nil
}
