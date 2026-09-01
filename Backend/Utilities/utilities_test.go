package Utilities

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"proyecto1/Structs"
)

func TestWriteReadObjectRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "roundtrip.mia")

	if err := CreateFile(path); err != nil {
		t.Fatalf("CreateFile: %v", err)
	}

	file, err := OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer file.Close()

	var original Structs.MBR
	original.MbrSize = 64 * 1024
	original.Signature = 12345
	copy(original.CreationDate[:], "2026-09-01")
	copy(original.Fit[:], "f")

	if err := WriteObject(file, original, 0); err != nil {
		t.Fatalf("WriteObject: %v", err)
	}

	var decoded Structs.MBR
	if err := ReadObject(file, &decoded, 0); err != nil {
		t.Fatalf("ReadObject: %v", err)
	}

	if decoded.MbrSize != original.MbrSize {
		t.Fatalf("MbrSize mismatch: want %d got %d", original.MbrSize, decoded.MbrSize)
	}
	if decoded.Signature != original.Signature {
		t.Fatalf("Signature mismatch: want %d got %d", original.Signature, decoded.Signature)
	}
	if string(decoded.CreationDate[:]) != string(original.CreationDate[:]) {
		t.Fatalf("CreationDate mismatch: want %q got %q", original.CreationDate, decoded.CreationDate)
	}
}

func TestGenerateMBRReportCreatesDotFile(t *testing.T) {
	var mbr Structs.MBR
	mbr.MbrSize = 128 * 1024
	mbr.Signature = 77
	copy(mbr.CreationDate[:], "2026-09-01")
	copy(mbr.Fit[:], "f")

	mbr.Partitions[0].Start = 256
	mbr.Partitions[0].Size = 32 * 1024
	copy(mbr.Partitions[0].Name[:], "data")
	copy(mbr.Partitions[0].Type[:], "p")
	copy(mbr.Partitions[0].Fit[:], "f")

	reportPath := filepath.Join(t.TempDir(), "mbr.jpg")
	if err := GenerateMBRReport(mbr, nil, reportPath, nil); err != nil {
		t.Fatalf("GenerateMBRReport: %v", err)
	}

	dotPath := filepath.Join(filepath.Dir(reportPath), "mbr.dot")
	content, err := os.ReadFile(dotPath)
	if err != nil {
		t.Fatalf("read generated dot file: %v", err)
	}

	text := string(content)
	for _, expected := range []string{"Master Boot Record", "Partition 1", "data"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("generated MBR report does not contain %q", expected)
		}
	}
}

func TestGenerateDiskReportCreatesDotFile(t *testing.T) {
	var mbr Structs.MBR
	mbr.MbrSize = 128 * 1024
	mbr.Partitions[0].Start = 256
	mbr.Partitions[0].Size = 32 * 1024
	copy(mbr.Partitions[0].Name[:], "data")
	copy(mbr.Partitions[0].Type[:], "p")

	reportPath := filepath.Join(t.TempDir(), "disk.png")
	if err := GenerateDiskReport(mbr, nil, reportPath, nil, mbr.MbrSize); err != nil {
		t.Fatalf("GenerateDiskReport: %v", err)
	}

	dotPath := filepath.Join(filepath.Dir(reportPath), "disk.dot")
	content, err := os.ReadFile(dotPath)
	if err != nil {
		t.Fatalf("read generated dot file: %v", err)
	}

	text := string(content)
	for _, expected := range []string{"digraph DISK", "data", "Free"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("generated disk report does not contain %q", expected)
		}
	}
}
