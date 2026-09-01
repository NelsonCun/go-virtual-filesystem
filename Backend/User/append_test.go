package User

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NelsonCun/go-virtual-filesystem/Structs"
	"github.com/NelsonCun/go-virtual-filesystem/Utilities"
)

func newAppendTestFile(t *testing.T) (*os.File, Structs.Superblock) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "append.bin")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatalf("open append test file: %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })

	inodeSize := int32(binary.Size(Structs.Inode{}))
	blockSize := int32(binary.Size(Structs.Fileblock{}))
	sb := Structs.Superblock{
		S_inodes_count: 4,
		S_blocks_count: 4,
		S_inode_start:  128,
		S_block_start:  128 + 4*inodeSize,
	}
	if err := file.Truncate(int64(sb.S_block_start + 4*blockSize)); err != nil {
		t.Fatalf("truncate append test file: %v", err)
	}
	return file, sb
}

func initializedAppendInode(size int32, blockIndex int32) Structs.Inode {
	var inode Structs.Inode
	for i := range inode.I_block {
		inode.I_block[i] = -1
	}
	inode.I_size = size
	inode.I_block[0] = blockIndex
	return inode
}

func TestAppendToFileBlockPersistsRequestedInode(t *testing.T) {
	file, sb := newAppendTestFile(t)
	inodeSize := int32(binary.Size(Structs.Inode{}))
	blockSize := int32(binary.Size(Structs.Fileblock{}))

	target := initializedAppendInode(3, 1)
	sentinel := initializedAppendInode(55, 2)
	sentinel.I_uid = 777

	if err := Utilities.WriteObject(file, sentinel, int64(sb.S_inode_start+inodeSize)); err != nil {
		t.Fatal(err)
	}
	if err := Utilities.WriteObject(file, target, int64(sb.S_inode_start+3*inodeSize)); err != nil {
		t.Fatal(err)
	}

	var block Structs.Fileblock
	copy(block.B_content[:], "abc")
	if err := Utilities.WriteObject(file, block, int64(sb.S_block_start+blockSize)); err != nil {
		t.Fatal(err)
	}

	if err := AppendToFileBlock(3, &target, "XYZ", file, sb); err != nil {
		t.Fatalf("append: %v", err)
	}

	var persistedTarget Structs.Inode
	if err := Utilities.ReadObject(file, &persistedTarget, int64(sb.S_inode_start+3*inodeSize)); err != nil {
		t.Fatal(err)
	}
	if persistedTarget.I_size != 6 {
		t.Fatalf("persisted target size: want 6 got %d", persistedTarget.I_size)
	}

	var persistedSentinel Structs.Inode
	if err := Utilities.ReadObject(file, &persistedSentinel, int64(sb.S_inode_start+inodeSize)); err != nil {
		t.Fatal(err)
	}
	if persistedSentinel.I_uid != 777 || persistedSentinel.I_size != 55 {
		t.Fatalf("unrelated inode modified: uid=%d size=%d",
			persistedSentinel.I_uid, persistedSentinel.I_size)
	}

	var persistedBlock Structs.Fileblock
	if err := Utilities.ReadObject(file, &persistedBlock, int64(sb.S_block_start+blockSize)); err != nil {
		t.Fatal(err)
	}
	if got := string(persistedBlock.B_content[:6]); got != "abcXYZ" {
		t.Fatalf("block content: want abcXYZ got %q", got)
	}
}

func TestAppendToFileBlockRejectsAdditionalBlockRequirement(t *testing.T) {
	file, sb := newAppendTestFile(t)
	inodeSize := int32(binary.Size(Structs.Inode{}))
	blockSize := int32(binary.Size(Structs.Fileblock{}))

	target := initializedAppendInode(63, 1)
	if err := Utilities.WriteObject(file, target, int64(sb.S_inode_start+3*inodeSize)); err != nil {
		t.Fatal(err)
	}

	var block Structs.Fileblock
	copy(block.B_content[:], strings.Repeat("a", 63))
	if err := Utilities.WriteObject(file, block, int64(sb.S_block_start+blockSize)); err != nil {
		t.Fatal(err)
	}

	err := AppendToFileBlock(3, &target, "XY", file, sb)
	if err == nil {
		t.Fatal("append requiring another block must be rejected")
	}
	if !strings.Contains(err.Error(), "additional file blocks") {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.I_size != 63 {
		t.Fatalf("failed append changed in-memory size: %d", target.I_size)
	}

	var persistedTarget Structs.Inode
	if err := Utilities.ReadObject(file, &persistedTarget, int64(sb.S_inode_start+3*inodeSize)); err != nil {
		t.Fatal(err)
	}
	if persistedTarget.I_size != 63 {
		t.Fatalf("failed append changed persisted size: %d", persistedTarget.I_size)
	}
}
