package User

import (
	"path/filepath"
	"testing"

	"github.com/NelsonCun/go-virtual-filesystem/DiskManagement"
	"github.com/NelsonCun/go-virtual-filesystem/FileSystem"
	"github.com/NelsonCun/go-virtual-filesystem/Structs"
	"github.com/NelsonCun/go-virtual-filesystem/Utilities"
)

func TestInitSearchMissingPathReturnsNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-path.mia")

	DiskManagement.Mkdisk(256, "ff", "k", path)
	DiskManagement.Fdisk(192, path, "users", "k", "p", "ff")
	DiskManagement.Mount(path, "users")

	id := mountedIDForLoginTest(t, path)
	FileSystem.Mkfs(id, "full", "2fs")

	file, err := Utilities.OpenFile(path)
	if err != nil {
		t.Fatalf("open virtual disk: %v", err)
	}
	defer file.Close()

	var mbr Structs.MBR
	if err := Utilities.ReadObject(file, &mbr, 0); err != nil {
		t.Fatalf("read MBR: %v", err)
	}

	var sb Structs.Superblock
	if err := Utilities.ReadObject(file, &sb, int64(mbr.Partitions[0].Start)); err != nil {
		t.Fatalf("read superblock: %v", err)
	}

	got := InitSearch("/does-not-exist.txt", file, sb)
	if got != -1 {
		t.Fatalf("missing path should return -1, got %d", got)
	}
}
