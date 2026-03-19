package templates

import "embed"

// Assets stores local generator templates for stacks that do not have an upstream initializer.
//
//go:embed all:assets/**
var Assets embed.FS

func Exists(path string) bool {
	file, err := Assets.Open(path)
	if err != nil {
		return false
	}

	defer file.Close()

	return true
}
