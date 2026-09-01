package Analyzer

import "testing"

func TestGetCommandAndParams(t *testing.T) {
	command, params := getCommandAndParams(`MKDISK -size=64 -unit=k -path="/tmp/virtual disk.mia"`)

	if command != "mkdisk" {
		t.Fatalf("expected normalized command mkdisk, got %q", command)
	}

	expected := `-size=64 -unit=k -path="/tmp/virtual disk.mia"`
	if params != expected {
		t.Fatalf("expected params %q, got %q", expected, params)
	}
}

func TestGetCommandAndParamsEmptyInput(t *testing.T) {
	command, params := getCommandAndParams("")

	if command != "" || params != "" {
		t.Fatalf("expected empty command and params, got command=%q params=%q", command, params)
	}
}
