package FileSystem

import (
	"encoding/binary"
	"path/filepath"
	"testing"

	"github.com/NelsonCun/go-virtual-filesystem/DiskManagement"
	"github.com/NelsonCun/go-virtual-filesystem/Structs"
	"github.com/NelsonCun/go-virtual-filesystem/Utilities"
)

func TestMkfsCreatesExt2Metadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "filesystem.mia")

	DiskManagement.Mkdisk(256, "ff", "k", path)
	DiskManagement.Fdisk(192, path, "data", "k", "p", "ff")
	DiskManagement.Mount(path, "data")

	id := firstMountedID(t, path)
	Mkfs(id, "full", "2fs")

	file, err := Utilities.OpenFile(path)
	if err != nil {
		t.Fatalf("open virtual disk: %v", err)
	}
	defer file.Close()

	var mbr Structs.MBR
	if err := Utilities.ReadObject(file, &mbr, 0); err != nil {
		t.Fatalf("read MBR: %v", err)
	}

	partition := mbr.Partitions[0]
	if partition.Size <= 0 {
		t.Fatal("expected initialized primary partition")
	}

	var sb Structs.Superblock
	if err := Utilities.ReadObject(file, &sb, int64(partition.Start)); err != nil {
		t.Fatalf("read superblock: %v", err)
	}

	if sb.S_filesystem_type != 2 {
		t.Fatalf("filesystem type mismatch: want 2 got %d", sb.S_filesystem_type)
	}
	if sb.S_magic != 0xEF53 {
		t.Fatalf("superblock magic mismatch: want 0xEF53 got 0x%X", sb.S_magic)
	}
	if sb.S_inodes_count <= 2 {
		t.Fatalf("expected more than two inodes, got %d", sb.S_inodes_count)
	}
	if sb.S_blocks_count != 3*sb.S_inodes_count {
		t.Fatalf("block count invariant failed: blocks=%d inodes=%d", sb.S_blocks_count, sb.S_inodes_count)
	}

	var root Structs.Inode
	if err := Utilities.ReadObject(file, &root, int64(sb.S_inode_start)); err != nil {
		t.Fatalf("read root inode: %v", err)
	}
	if root.I_block[0] != 0 {
		t.Fatalf("expected root inode to reference block 0, got %d", root.I_block[0])
	}

	var users Structs.Inode
	usersOffset := int64(sb.S_inode_start + int32(binary.Size(Structs.Inode{})))
	if err := Utilities.ReadObject(file, &users, usersOffset); err != nil {
		t.Fatalf("read users inode: %v", err)
	}
	if users.I_block[0] != 1 {
		t.Fatalf("expected users inode to reference block 1, got %d", users.I_block[0])
	}
}

func firstMountedID(t *testing.T, path string) string {
	t.Helper()
	for _, partitions := range DiskManagement.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.Path == path {
				return partition.ID
			}
		}
	}
	t.Fatalf("expected a mounted partition for path %q", path)
	return ""
}

func TestMkfsTracksFirstFreeInodeAndBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "first-free.mia")

	DiskManagement.Mkdisk(256, "ff", "k", path)
	DiskManagement.Fdisk(192, path, "data", "k", "p", "ff")
	DiskManagement.Mount(path, "data")

	id := firstMountedID(t, path)
	Mkfs(id, "full", "2fs")

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

	// mkfs reserves inode 0 for root, inode 1 for users.txt,
	// block 0 for the root directory and block 1 for users.txt.
	if sb.S_fist_ino != 2 {
		t.Fatalf("first free inode mismatch: want 2 got %d", sb.S_fist_ino)
	}
	if sb.S_first_blo != 2 {
		t.Fatalf("first free block mismatch: want 2 got %d", sb.S_first_blo)
	}
}
