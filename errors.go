package ipmi

import (
	"errors"
	"fmt"
)

var (
	ErrUnpackedDataTooShort = errors.New("unpacked data is too short")
)

func ErrUnpackedDataTooShortWith(actual int, expected int) error {
	return fmt.Errorf("%s (%d/%d)", ErrUnpackedDataTooShort, actual, expected)
}

func ErrNotEnoughDataWith(msg string, actual int, expected int) error {
	return fmt.Errorf("not enough data for %s (%d/%d)", msg, actual, expected)
}

func ErrDCMIGroupExtensionIdentificationMismatch(expected uint8, actual uint8) error {
	return fmt.Errorf("DCMI group extension ID mismatch: expected %#02x, got %#02x", expected, actual)
}

func CheckDCMIGroupExenstionMatch(grpExt uint8) error {
	if grpExt != GroupExtensionDCMI {
		return ErrDCMIGroupExtensionIdentificationMismatch(GroupExtensionDCMI, grpExt)
	}
	return nil
}
