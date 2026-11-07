package decoder

import (
	"encoding/hex"
	"time"
)

func ValidDecoderID(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 16 && hex.EncodeToString(decoded) == value
}

func ValidDecoderDates(installed, removed *string) bool {
	for _, date := range []*string{installed, removed} {
		if date != nil {
			if _, err := time.Parse(time.DateOnly, *date); err != nil {
				return false
			}
		}
	}
	return installed == nil || removed == nil || *installed <= *removed
}

func InstanceIn(result *DecoderConfigurationContext, id string) *DecoderInstance {
	for index := range result.Instances {
		if result.Instances[index].ID == id {
			return &result.Instances[index]
		}
	}
	return nil
}

func InstallationIn(result *DecoderConfigurationContext, id string) *DecoderInstallation {
	for index := range result.Installations {
		if result.Installations[index].ID == id {
			return &result.Installations[index]
		}
	}
	return nil
}

func WorkspaceIn(result *DecoderConfigurationContext, id string) *DecoderCVWorkspace {
	for index := range result.Workspaces {
		if result.Workspaces[index].ID == id {
			return &result.Workspaces[index]
		}
	}
	return nil
}
