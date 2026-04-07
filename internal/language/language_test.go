// Copyright 2026 Chalindu Kodikara
// SPDX-License-Identifier: Apache-2.0

package language

import "testing"

func TestForPath_LineComment(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"main.go", "//"},
		{"app.rs", "//"},
		{"App.java", "//"},
		{"index.js", "//"},
		{"index.jsx", "//"},
		{"index.ts", "//"},
		{"index.tsx", "//"},
		{"hello.c", "//"},
		{"hello.h", "//"},
		{"hello.cpp", "//"},
		{"hello.hpp", "//"},
		{"hello.cc", "//"},
		{"App.cs", "//"},
		{"View.swift", "//"},
		{"App.kt", "//"},
		{"App.kts", "//"},
		{"App.scala", "//"},
		{"main.dart", "//"},
		{"api.proto", "//"},
		{"build.groovy", "//"},
		{"main.zig", "//"},
		{"AppDelegate.m", "//"},
		{"AppDelegate.mm", "//"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			s := ForPath(tc.path)
			if s == nil {
				t.Fatalf("ForPath(%q) = nil", tc.path)
			}
			if s.LinePrefix != tc.want {
				t.Errorf("LinePrefix = %q, want %q", s.LinePrefix, tc.want)
			}
			if s.IsBlock() {
				t.Errorf("IsBlock() = true, want false")
			}
		})
	}
}

func TestForPath_HashComment(t *testing.T) {
	cases := []string{
		"script.py",
		"types.pyi",
		"app.rb",
		"deploy.sh",
		"deploy.bash",
		"setup.zsh",
		"plugin.fish",
		"perl.pl",
		"perl.pm",
		"config.yaml",
		"config.yml",
		"pyproject.toml",
		"build.mk",
		"main.tf",
		"vars.hcl",
		"plot.r",
		"plot.R",
		"app.ex",
		"test.exs",
		"shell.nix",
		"app.conf",
		"build.cmake",
		"deploy.ps1",
		"module.psm1",
		"app.tcl",
	}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			s := ForPath(p)
			if s == nil {
				t.Fatalf("ForPath(%q) = nil", p)
			}
			if s.LinePrefix != "#" {
				t.Errorf("LinePrefix = %q, want %q", s.LinePrefix, "#")
			}
		})
	}
}

func TestForPath_DashDashComment(t *testing.T) {
	cases := []string{"lib.hs", "init.lua", "query.sql", "package.ada", "Main.elm"}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			s := ForPath(p)
			if s == nil {
				t.Fatalf("ForPath(%q) = nil", p)
			}
			if s.LinePrefix != "--" {
				t.Errorf("LinePrefix = %q, want %q", s.LinePrefix, "--")
			}
		})
	}
}

func TestForPath_BlockComment(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"style.css", "/*"},
		{"app.scss", "/*"},
		{"app.less", "/*"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			s := ForPath(tc.path)
			if s == nil {
				t.Fatalf("ForPath(%q) = nil", tc.path)
			}
			if s.BlockStart != tc.want {
				t.Errorf("BlockStart = %q, want %q", s.BlockStart, tc.want)
			}
			if s.BlockEnd != "*/" {
				t.Errorf("BlockEnd = %q, want %q", s.BlockEnd, "*/")
			}
			if !s.IsBlock() {
				t.Error("IsBlock() = false, want true")
			}
		})
	}
}

func TestForPath_XMLBlockComment(t *testing.T) {
	cases := []string{"page.html", "page.htm", "doc.xml", "icon.svg", "App.vue"}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			s := ForPath(p)
			if s == nil {
				t.Fatalf("ForPath(%q) = nil", p)
			}
			if s.BlockStart != "<!--" || s.BlockEnd != "-->" {
				t.Errorf("got %q/%q, want <!--/-->", s.BlockStart, s.BlockEnd)
			}
		})
	}
}

func TestForPath_FilenameStyles(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"Dockerfile", "#"},
		{"Makefile", "#"},
		{"Rakefile", "#"},
		{"Gemfile", "#"},
		{"Vagrantfile", "#"},
		{"Jenkinsfile", "//"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := ForPath(tc.name)
			if s == nil {
				t.Fatalf("ForPath(%q) = nil", tc.name)
			}
			if s.LinePrefix != tc.want {
				t.Errorf("LinePrefix = %q, want %q", s.LinePrefix, tc.want)
			}
		})
	}
}

func TestForPath_FilenameStyles_WithPath(t *testing.T) {
	// filenames must work even when nested in a directory
	if s := ForPath("infra/docker/Dockerfile"); s == nil || s.LinePrefix != "#" {
		t.Error("nested Dockerfile should be detected")
	}
	if s := ForPath("ci/Jenkinsfile"); s == nil || s.LinePrefix != "//" {
		t.Error("nested Jenkinsfile should be detected")
	}
}

func TestForPath_SkippedExtensions(t *testing.T) {
	skipped := []string{
		"data.json",
		"package.lock",
		"go.sum",
		"image.png",
		"image.jpg",
		"image.jpeg",
		"image.gif",
		"icon.ico",
		"image.bmp",
		"image.webp",
		"image.tiff",
		"app.wasm",
		"doc.pdf",
		"archive.zip",
		"archive.tar",
		"archive.gz",
		"archive.bz2",
		"archive.xz",
		"archive.zst",
		"binary.bin",
		"binary.exe",
		"lib.dll",
		"lib.so",
		"lib.dylib",
		"obj.o",
		"lib.a",
		"font.woff",
		"font.woff2",
		"font.ttf",
		"font.otf",
		"font.eot",
		"sound.mp3",
		"video.mp4",
		"sound.wav",
		"video.avi",
		"go.mod",
		"changes.patch",
		"changes.diff",
		"app.min.js",
		"app.min.css",
		"app.js.map",
		"types.pb.go",
	}
	for _, p := range skipped {
		t.Run(p, func(t *testing.T) {
			if s := ForPath(p); s != nil {
				t.Errorf("ForPath(%q) = %+v, want nil", p, s)
			}
		})
	}
}

func TestForPath_UnknownExtension(t *testing.T) {
	cases := []string{
		"readme.txt",
		"data.csv",
		"unknown.xyz",
		"noext",
		"strange.qqq",
	}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			if s := ForPath(p); s != nil {
				t.Errorf("ForPath(%q) = %+v, want nil", p, s)
			}
		})
	}
}

func TestForPath_CaseInsensitiveExt(t *testing.T) {
	// extension lookup is lowercased before matching
	if s := ForPath("MAIN.GO"); s == nil || s.LinePrefix != "//" {
		t.Errorf("MAIN.GO should be detected as Go")
	}
	if s := ForPath("Script.PY"); s == nil || s.LinePrefix != "#" {
		t.Errorf("Script.PY should be detected as Python")
	}
}

func TestForPath_NestedPath(t *testing.T) {
	if s := ForPath("path/to/deep/file.ts"); s == nil || s.LinePrefix != "//" {
		t.Errorf("nested .ts file should be detected")
	}
	if s := ForPath("a/b/c/style.css"); s == nil || !s.IsBlock() {
		t.Errorf("nested .css file should be detected as block")
	}
}

func TestSupported(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"main.go", true},
		{"script.py", true},
		{"style.css", true},
		{"Dockerfile", true},
		{"data.json", false},
		{"unknown.xyz", false},
		{"image.png", false},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			if got := Supported(tc.path); got != tc.want {
				t.Errorf("Supported(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

func TestStyle_IsBlock(t *testing.T) {
	if (Style{LinePrefix: "//"}).IsBlock() {
		t.Error("line style should not be block")
	}
	if !(Style{BlockStart: "/*", BlockEnd: "*/"}).IsBlock() {
		t.Error("block style should be block")
	}
	if (Style{}).IsBlock() {
		t.Error("empty style should not be block")
	}
}
