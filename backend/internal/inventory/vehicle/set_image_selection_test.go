package vehicle

import (
	"context"
	"errors"
	"testing"
)

type setImageLookupProbe struct {
	selection                   SetImageSelection
	member, automatic           *VehicleImage
	memberError                 error
	memberReads, automaticReads int
}

func (p *setImageLookupProbe) Selection(context.Context, string) (SetImageSelection, error) {
	return p.selection, nil
}
func (p *setImageLookupProbe) MemberImage(context.Context, string, string) (*VehicleImage, error) {
	p.memberReads++
	return p.member, p.memberError
}
func (p *setImageLookupProbe) AutomaticImage(context.Context, string) (*VehicleImage, error) {
	p.automaticReads++
	return p.automatic, nil
}

func TestSetImageMissingMemberFallsBackButStorageFailureDoesNot(t *testing.T) {
	probe := &setImageLookupProbe{selection: SetImageSelection{Mode: VehicleSetMainImageModeMember, MemberImageID: "missing"}, automatic: &VehicleImage{ID: "automatic", VehicleID: "owner", URL: "recorded-url"}}
	mode, member, _, image, err := ResolveSetImage(t.Context(), probe, "set")
	if err != nil || mode != VehicleSetMainImageModeMember || member != "missing" || image == nil || image.Source != "automatic" || image.ImageID != "automatic" || probe.automaticReads != 1 {
		t.Fatal("missing selected image did not retain selection/fallback", mode, member, image, err)
	}
	failure := errors.New("selected image storage unavailable")
	probe.memberError = failure
	probe.automaticReads = 0
	_, _, _, image, err = ResolveSetImage(t.Context(), probe, "set")
	if !errors.Is(err, failure) || image != nil || probe.automaticReads != 0 {
		t.Fatal("storage failure hidden by fallback", image, err)
	}
}

func TestDedicatedSetImageUsesExistingVersionEncodingWithoutReadingMembers(t *testing.T) {
	probe := &setImageLookupProbe{selection: SetImageSelection{Mode: VehicleSetMainImageModeDedicated, MemberImageID: "recorded member", Dedicated: VehicleSetImage{
		BlobID: " version / + ", FileName: "recorded.png", CreatedAt: "original", UpdatedAt: "recorded update",
	}}}
	mode, member, dedicated, image, err := ResolveSetImage(t.Context(), probe, "set")
	if err != nil || mode != VehicleSetMainImageModeDedicated || member != "recorded member" || dedicated == nil || image == nil ||
		image.Source != "dedicated" || image.URL != "/api/v1/vehicle-sets/set/image/file?v=version+%2F+%2B" ||
		dedicated.BlobID != probe.selection.Dedicated.BlobID || dedicated.CreatedAt != "original" || probe.memberReads != 0 || probe.automaticReads != 0 {
		t.Fatal("dedicated selection or source identity changed", dedicated, image, err)
	}
}
