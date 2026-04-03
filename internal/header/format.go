// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package header

import (
	"fmt"

	"github.com/chalindukodikara/licenseops/internal/language"
)

// Format defines how license headers are generated, detected, and stripped.
type Format interface {
	// Name returns the format identifier (e.g. "spdx", "reuse", "apache-long").
	Name() string
	// Generate produces the header text for the given parameters and comment style.
	Generate(style *language.Style, year, holder, license string) string
	// HasValid checks if a file has a valid header in this format.
	HasValid(path string, style *language.Style, holder, license string) (bool, error)
	// StripExisting removes an existing header in this format from file content.
	StripExisting(src []byte, style *language.Style) []byte
}

// FormatByName returns the Format implementation for the given name.
func FormatByName(name string) (Format, error) {
	switch name {
	case "spdx", "":
		return &SPDXFormat{}, nil
	case "reuse":
		return &ReuseFormat{}, nil
	case "apache-long":
		return &ApacheLongFormat{}, nil
	case "gpl-long":
		return &GPLLongFormat{}, nil
	case "custom":
		return &CustomFormat{}, nil
	default:
		return nil, fmt.Errorf("unknown header format: %q (valid: spdx, reuse, apache-long, gpl-long, custom)", name)
	}
}

// AllFormats returns all built-in format implementations for migration stripping.
func AllFormats() []Format {
	return []Format{
		&SPDXFormat{},
		&ReuseFormat{},
		&ApacheLongFormat{},
		&GPLLongFormat{},
	}
}
