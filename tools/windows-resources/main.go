package main

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/tc-hib/winres"
	"github.com/tc-hib/winres/version"
)

var releaseVersion = regexp.MustCompile(`^[0-9]+(\.[0-9]+){2,3}(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$`)

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintln(os.Stderr, "usage: windows-resources VERSION ICON.png INPUT.exe OUTPUT.exe")
		os.Exit(2)
	}
	if err := writeResources(os.Args[1], os.Args[2], os.Args[3], os.Args[4]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func writeResources(release, iconPath, inputPath, outputPath string) error {
	info, err := versionInfo(release)
	if err != nil {
		return err
	}
	source, err := os.Open(iconPath)
	if err != nil {
		return err
	}
	defer source.Close()
	img, err := png.Decode(source)
	if err != nil {
		return err
	}
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	size := max(width, height)
	square := image.NewNRGBA(image.Rect(0, 0, size, size))
	centered := image.Rect(0, 0, width, height).Add(image.Pt((size-width)/2, (size-height)/2))
	draw.Draw(square, centered, img, img.Bounds().Min, draw.Src)
	icon, err := winres.NewIconFromResizedImage(square, []int{256, 64, 48, 32, 16})
	if err != nil {
		return err
	}
	input, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer input.Close()
	resources, err := winres.LoadFromEXE(input)
	if err != nil && !errors.Is(err, winres.ErrNoResources) {
		return err
	}
	if _, err := input.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := resources.SetIcon(winres.ID(1), icon); err != nil {
		return err
	}
	resources.SetVersionInfo(info)
	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	if err := resources.WriteToEXE(output, input); err != nil {
		_ = output.Close()
		return err
	}
	return output.Close()
}

func versionInfo(release string) (version.Info, error) {
	var info version.Info
	if !releaseVersion.MatchString(release) {
		return info, fmt.Errorf("invalid RailKeeper release version %q", release)
	}
	numeric, _, prerelease := strings.Cut(release, "-")
	for index, part := range strings.Split(numeric, ".") {
		number, err := strconv.ParseUint(part, 10, 16)
		if err != nil {
			return info, fmt.Errorf("invalid Windows version component %q: %w", part, err)
		}
		info.FileVersion[index] = uint16(number)
	}
	info.ProductVersion = info.FileVersion
	info.Flags.Prerelease = prerelease
	for _, field := range []struct{ key, value string }{
		{version.ProductName, "RailKeeper"},
		{version.ProductVersion, release},
		{version.FileVersion, release},
		{version.FileDescription, "RailKeeper"},
		{version.InternalName, "RailKeeper.exe"},
		{version.OriginalFilename, "RailKeeper.exe"},
	} {
		if err := info.Set(version.LangDefault, field.key, field.value); err != nil {
			return info, err
		}
	}
	return info, nil
}
