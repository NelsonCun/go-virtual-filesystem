package DiskManagement

import (
	"os"
	"path/filepath"
	"testing"

	"proyecto1/Structs"
	"proyecto1/Utilities"
)

func resetMountedPartitionsForTest() {
	mountedPartitions = make(map[string][]MountedPartition)
}

func createTestDisk(t *testing.T, sizeKB int) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "disk.mia")
	Mkdisk(sizeKB, "ff", "k", path)

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat virtual disk: %v", err)
	}

	wantSize := int64(sizeKB * 1024)
	if info.Size() != wantSize {
		t.Fatalf("virtual disk size mismatch: want %d got %d", wantSize, info.Size())
	}

	return path
}

func readMBRForTest(t *testing.T, path string) Structs.MBR {
	t.Helper()

	file, err := Utilities.OpenFile(path)
	if err != nil {
		t.Fatalf("open virtual disk: %v", err)
	}
	defer file.Close()

	var mbr Structs.MBR
	if err := Utilities.ReadObject(file, &mbr, 0); err != nil {
		t.Fatalf("read MBR: %v", err)
	}
	return mbr
}

func TestMkdiskCreatesVirtualDiskAndMBR(t *testing.T) {
	resetMountedPartitionsForTest()
	path := createTestDisk(t, 64)

	mbr := readMBRForTest(t, path)
	if mbr.MbrSize != 64*1024 {
		t.Fatalf("MBR disk size mismatch: want %d got %d", 64*1024, mbr.MbrSize)
	}
	if mbr.Signature < 0 {
		t.Fatalf("expected non-negative MBR signature, got %d", mbr.Signature)
	}
}

func TestFdiskCreatesPrimaryPartition(t *testing.T) {
	resetMountedPartitionsForTest()
	path := createTestDisk(t, 128)

	Fdisk(32, path, "data", "k", "p", "ff")

	mbr := readMBRForTest(t, path)
	p := mbr.Partitions[0]

	if p.Size != 32*1024 {
		t.Fatalf("partition size mismatch: want %d got %d", 32*1024, p.Size)
	}
	if got := string(p.Type[:]); got != "p" {
		t.Fatalf("partition type mismatch: want p got %q", got)
	}
	if got := cleanTestString(p.Name[:]); got != "data" {
		t.Fatalf("partition name mismatch: want data got %q", got)
	}
}

func TestMountTracksPartitionWithNeutralID(t *testing.T) {
	resetMountedPartitionsForTest()
	path := createTestDisk(t, 128)
	Fdisk(32, path, "data", "k", "p", "ff")

	Mount(path, "data")

	partitionsByDisk := GetMountedPartitions()
	if len(partitionsByDisk) != 1 {
		t.Fatalf("expected one mounted disk, got %d", len(partitionsByDisk))
	}

	var mounted MountedPartition
	found := false
	for _, partitions := range partitionsByDisk {
		if len(partitions) == 1 {
			mounted = partitions[0]
			found = true
		}
	}

	if !found {
		t.Fatal("expected one mounted partition")
	}
	if mounted.ID != "vd1a" {
		t.Fatalf("expected neutral partition ID vd1a, got %q", mounted.ID)
	}
	if mounted.Status != '1' {
		t.Fatalf("expected mounted status 1, got %q", mounted.Status)
	}

	mbr := readMBRForTest(t, path)
	if got := cleanTestString(mbr.Partitions[0].Id[:]); got != "vd1a" {
		t.Fatalf("persisted partition ID mismatch: want vd1a got %q", got)
	}
	if mbr.Partitions[0].Status[0] != '1' {
		t.Fatalf("persisted mount status mismatch: got %q", mbr.Partitions[0].Status[0])
	}
}

func cleanTestString(data []byte) string {
	end := len(data)
	for end > 0 && data[end-1] == 0 {
		end--
	}
	return string(data[:end])
}

func TestFdiskKeepsPrimaryPartitionWithinDiskBounds(t *testing.T) {
	resetMountedPartitionsForTest()
	path := createTestDisk(t, 64)

	// A partition cannot consume the entire disk because the MBR itself
	// occupies bytes at the beginning of the virtual disk.
	Fdisk(64, path, "too-large", "k", "p", "ff")

	mbr := readMBRForTest(t, path)
	p := mbr.Partitions[0]

	if p.Size == 0 {
		return // Correct behavior: the invalid partition was rejected.
	}

	end := int64(p.Start) + int64(p.Size)
	if end > int64(mbr.MbrSize) {
		t.Fatalf(
			"partition exceeds disk boundary: start=%d size=%d end=%d disk=%d",
			p.Start,
			p.Size,
			end,
			mbr.MbrSize,
		)
	}
}

func TestLogicalPartitionUsesExtendedPartitionCapacity(t *testing.T) {
	resetMountedPartitionsForTest()
	path := createTestDisk(t, 512)

	Fdisk(400, path, "extended", "k", "e", "ff")
	Fdisk(128, path, "logical1", "k", "l", "ff")

	mbr := readMBRForTest(t, path)
	extended := mbr.Partitions[0]

	if extended.Type[0] != 'e' {
		t.Fatalf("expected first partition to be extended, got %q", extended.Type[0])
	}

	file, err := Utilities.OpenFile(path)
	if err != nil {
		t.Fatalf("open virtual disk: %v", err)
	}
	defer file.Close()

	var ebr Structs.EBR
	if err := Utilities.ReadObject(file, &ebr, int64(extended.Start)); err != nil {
		t.Fatalf("read first EBR: %v", err)
	}

	if ebr.PartSize != 128*1024 {
		t.Fatalf(
			"logical partition should fit inside extended partition: want size=%d got=%d",
			128*1024,
			ebr.PartSize,
		)
	}

	logicalEnd := int64(ebr.PartStart) + int64(ebr.PartSize)
	extendedEnd := int64(extended.Start) + int64(extended.Size)
	if logicalEnd > extendedEnd {
		t.Fatalf(
			"logical partition exceeds extended boundary: logicalEnd=%d extendedEnd=%d",
			logicalEnd,
			extendedEnd,
		)
	}
}
