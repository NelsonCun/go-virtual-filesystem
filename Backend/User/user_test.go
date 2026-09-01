package User

import (
	"path/filepath"
	"testing"

	"github.com/NelsonCun/go-virtual-filesystem/DiskManagement"
	"github.com/NelsonCun/go-virtual-filesystem/FileSystem"
)

func TestLoginRequiresExactPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "login.mia")

	DiskManagement.Mkdisk(256, "ff", "k", path)
	DiskManagement.Fdisk(192, path, "users", "k", "p", "ff")
	DiskManagement.Mount(path, "users")

	id := mountedIDForLoginTest(t, path)
	FileSystem.Mkfs(id, "full", "2fs")

	Login("root", "12", id)
	if isLoggedInForTest(id) {
		t.Fatal("partial password must not authenticate")
	}

	Login("root", "123", id)
	if !isLoggedInForTest(id) {
		t.Fatal("exact bootstrap credentials should authenticate")
	}
}

func mountedIDForLoginTest(t *testing.T, path string) string {
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

func isLoggedInForTest(id string) bool {
	for _, partitions := range DiskManagement.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.ID == id {
				return partition.LoggedIn
			}
		}
	}
	return false
}
