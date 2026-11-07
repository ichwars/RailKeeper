package ecosmodel

import (
	"fmt"
	"strconv"
	"strings"
)

func ECoSLocomotiveDetailCommand(objectID int) string {
	return fmt.Sprintf("get(%d, profile, protocol, name, addr, funcdesc)", objectID)
}

func ECoSRawProbeFields() []string {
	return []string{
		"cv",
		"cvs",
		"cvlist",
		"functionmapping",
	}
}

func InterestingECoSFields(attributes map[string][]string) []string {
	interesting := []string{}
	for key := range attributes {
		normalized := strings.ToLower(key)
		if strings.Contains(normalized, "cv") || normalized == "functionmapping" {
			interesting = append(interesting, key)
		}
	}
	SortStrings(interesting)
	return interesting
}

func ParseECoSLocomotives(lines []string) []ECoSLocomotive {
	locomotives := []ECoSLocomotive{}
	byID := map[int]int{}
	for _, line := range lines {
		if strings.HasPrefix(line, "<") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		objectID, err := strconv.Atoi(fields[0])
		if err != nil || objectID <= 0 {
			continue
		}
		locomotiveIndex, exists := byID[objectID]
		if !exists {
			locomotives = append(locomotives, ECoSLocomotive{
				ObjectID:  objectID,
				Functions: []ECoSFunction{},
			})
			locomotiveIndex = len(locomotives) - 1
			byID[objectID] = locomotiveIndex
		}
		for _, arg := range ParseECoSArgumentList(line) {
			ApplyECoSArgument(&locomotives[locomotiveIndex], arg)
		}
	}
	return locomotives
}

func MergeECoSRawLocomotive(target *ECoSRawLocomotive, source ECoSLocomotive) {
	if source.Name != "" {
		target.Name = source.Name
	}
	if source.Address != 0 {
		target.Address = source.Address
	}
	if source.Protocol != "" {
		target.Protocol = source.Protocol
	}
	if source.Profile != "" {
		target.Profile = source.Profile
	}
	if len(source.Functions) > 0 {
		target.Functions = source.Functions
	}
}

func ApplyECoSArgument(locomotive *ECoSLocomotive, arg Argument) {
	key := strings.ToLower(arg.Key)
	if !IsAllowedECoSLocomotiveAttribute(key) {
		return
	}
	if locomotive.Attributes == nil {
		locomotive.Attributes = map[string][]string{}
	}
	locomotive.Attributes[key] = append(locomotive.Attributes[key], CleanECoSValue(arg.Value))
	switch key {
	case "name":
		locomotive.Name = CleanECoSValue(arg.Value)
	case "protocol":
		locomotive.Protocol = CleanECoSValue(arg.Value)
	case "profile":
		locomotive.Profile = CleanECoSValue(arg.Value)
	case "addr":
		locomotive.Address = ParseECoSInt(arg.Value)
	case "funcdesc":
		if len(arg.Params) >= 2 {
			index := ParseECoSInt(arg.Params[0])
			if index < 0 || index > 31 {
				return
			}
			function := EnsureECoSFunction(locomotive, index)
			function.Description = ParseECoSInt(arg.Params[1])
		}
	}
}

func EnsureECoSFunction(locomotive *ECoSLocomotive, index int) *ECoSFunction {
	for functionIndex := range locomotive.Functions {
		if locomotive.Functions[functionIndex].Index == index {
			return &locomotive.Functions[functionIndex]
		}
	}
	locomotive.Functions = append(locomotive.Functions, ECoSFunction{Index: index})
	position := len(locomotive.Functions) - 1
	for position > 0 && locomotive.Functions[position-1].Index > index {
		locomotive.Functions[position-1], locomotive.Functions[position] =
			locomotive.Functions[position], locomotive.Functions[position-1]
		position--
	}
	return &locomotive.Functions[position]
}
