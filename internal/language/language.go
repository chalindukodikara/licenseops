// Copyright 2026 The LicenseOps Authors
// SPDX-License-Identifier: Apache-2.0

package language

import (
	"path/filepath"
	"strings"
)

// Style represents how comments are written in a given language.
type Style struct {
	// LinePrefix is the single-line comment prefix (e.g. "//", "#", "--").
	// Empty if the language only supports block comments.
	LinePrefix string
	// BlockStart and BlockEnd delimit block comments (e.g. "/*" and "*/").
	// Empty if line comments are used.
	BlockStart string
	BlockEnd   string
}

// IsBlock returns true if this style uses block comments.
func (s Style) IsBlock() bool {
	return s.BlockStart != ""
}

var (
	slashSlash = Style{LinePrefix: "//"}
	hash       = Style{LinePrefix: "#"}
	dashDash   = Style{LinePrefix: "--"}
	cBlock     = Style{BlockStart: "/*", BlockEnd: "*/"}
	xmlBlock   = Style{BlockStart: "<!--", BlockEnd: "-->"}
)

// extensionStyles maps file extensions (with leading dot) to comment styles.
var extensionStyles = map[string]Style{
	// // style
	".go":    slashSlash,
	".rs":    slashSlash,
	".java":  slashSlash,
	".js":    slashSlash,
	".jsx":   slashSlash,
	".ts":    slashSlash,
	".tsx":   slashSlash,
	".c":     slashSlash,
	".h":     slashSlash,
	".cpp":   slashSlash,
	".hpp":   slashSlash,
	".cc":    slashSlash,
	".cs":    slashSlash,
	".swift": slashSlash,
	".kt":    slashSlash,
	".kts":   slashSlash,
	".scala": slashSlash,
	".dart":  slashSlash,
	".proto": slashSlash,
	".groovy": slashSlash,
	".zig":   slashSlash,
	".m":     slashSlash,
	".mm":    slashSlash,
	".v":     slashSlash,
	".sv":    slashSlash,

	// # style
	".py":         hash,
	".pyi":        hash,
	".rb":         hash,
	".sh":         hash,
	".bash":       hash,
	".zsh":        hash,
	".fish":       hash,
	".pl":         hash,
	".pm":         hash,
	".yaml":       hash,
	".yml":        hash,
	".toml":       hash,
	".dockerfile": hash,
	".mk":         hash,
	".tf":         hash,
	".hcl":        hash,
	".r":          hash,
	".R":          hash,
	".ex":         hash,
	".exs":        hash,
	".nix":        hash,
	".conf":       hash,
	".cmake":      hash,
	".ps1":        hash,
	".psm1":       hash,
	".tcl":        hash,

	// -- style
	".hs":  dashDash,
	".lua": dashDash,
	".sql": dashDash,
	".ada": dashDash,
	".elm": dashDash,

	// block comment only
	".css":  cBlock,
	".scss": cBlock,
	".less": cBlock,

	// XML-style block
	".html": xmlBlock,
	".htm":  xmlBlock,
	".xml":  xmlBlock,
	".svg":  xmlBlock,
	".vue":  xmlBlock,
}

// filenameStyles maps specific filenames (no path) to comment styles.
var filenameStyles = map[string]Style{
	"Dockerfile": hash,
	"Makefile":   hash,
	"Rakefile":   hash,
	"Gemfile":    hash,
	"Vagrantfile": hash,
	"Jenkinsfile": slashSlash,
}

// skipExtensions are extensions that should never have headers.
var skipExtensions = map[string]bool{
	".json": true, ".lock": true, ".sum": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".ico": true, ".bmp": true, ".webp": true, ".tiff": true,
	".wasm": true, ".pdf": true, ".zip": true, ".tar": true,
	".gz": true, ".bz2": true, ".xz": true, ".zst": true,
	".bin": true, ".exe": true, ".dll": true, ".so": true, ".dylib": true,
	".o": true, ".a": true, ".lib": true,
	".woff": true, ".woff2": true, ".ttf": true, ".otf": true, ".eot": true,
	".mp3": true, ".mp4": true, ".wav": true, ".avi": true,
	".mod": true, ".patch": true, ".diff": true,
	".min.js": true, ".min.css": true,
	".map": true,
	".pb.go": true,
}

// ForPath returns the comment style for the given file path, or nil if
// the file type is not supported or should be skipped.
func ForPath(path string) *Style {
	base := filepath.Base(path)

	// Check skip list first (binary, lock files, etc.)
	for ext := range skipExtensions {
		if strings.HasSuffix(base, ext) {
			return nil
		}
	}

	// Check filename-based styles
	if s, ok := filenameStyles[base]; ok {
		return &s
	}

	// Check extension-based styles
	ext := strings.ToLower(filepath.Ext(path))
	if s, ok := extensionStyles[ext]; ok {
		return &s
	}

	return nil
}

// Supported returns true if the file at path has a known comment style.
func Supported(path string) bool {
	return ForPath(path) != nil
}
