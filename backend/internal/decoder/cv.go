// Package decoder owns decoder and CV rules without transport or storage.
package decoder

func ValidCVNumber(value int) bool {
	return value >= 1 && value <= 1024
}

func ValidCVValue(value int) bool {
	return value >= 0 && value <= 255
}
