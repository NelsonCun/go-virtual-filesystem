# Virtual Filesystem & Disk Storage Engine in Go

A systems-programming project that models virtual disks, MBR/EBR partition
structures, an EXT2-style filesystem layout, binary metadata persistence, mount
state, basic user authentication, and Graphviz report generation.

The project is intentionally focused on the storage engine and command-line
workflow. A disconnected prototype frontend was removed so the repository
communicates one clear technical signal:

**Go · Systems Programming · Filesystem Design · Binary Storage · Disk Structures**

## Why this project matters

This repository demonstrates work below the typical web-application layer:

- binary serialization with `encoding/binary` and little-endian layouts;
- explicit byte offsets and random-access file I/O;
- MBR partition metadata and EBR-linked logical partitions;
- virtual disk boundary and extended-partition capacity validation;
- an EXT2-style superblock, bitmaps, inode table, and data blocks;
- deterministic in-memory mount identifiers;
- filesystem path traversal and a simulated `/users.txt` authentication flow;
- Graphviz `.dot` generation for MBR and disk-layout reports;
- regression tests for storage invariants and edge cases.

The implementation started as a systems-programming academic project and was
subsequently hardened, tested, and documented as a professional portfolio
artifact.

## Architecture

```text
Interactive CLI
      |
      v
Command Analyzer
      |
      +-------------------+---------------------+
      |                   |                     |
      v                   v                     v
Disk Management       Filesystem              User
      |                   |                     |
      +---------+---------+---------------------+
                |
                v
            Utilities
      binary read/write + reports
                |
                v
        Virtual .mia disk file
```

At the storage layer:

```text
Virtual disk
+-----+----------------------+----------------------+------+
| MBR | Primary / Extended   | Primary / Extended   | ...  |
+-----+----------------------+----------------------+------+

Extended partition
+-----+----------------+-----+----------------+------+
| EBR | Logical data   | EBR | Logical data   | ...  |
+-----+----------------+-----+----------------+------+

EXT2-style partition
+------------+--------------+--------------+-------------+--------+
| Superblock | Inode bitmap | Block bitmap | Inode table | Blocks |
+------------+--------------+--------------+-------------+--------+
```

See [`docs/architecture.md`](docs/architecture.md) for the component and storage
model in more detail.

## Implemented commands

| Command | Purpose |
| --- | --- |
| `mkdisk` | Create a virtual disk and initialize its MBR. |
| `rmdisk` | Remove a virtual disk. |
| `fdisk` | Create primary, extended, or logical partitions. |
| `mount` | Mount a top-level partition and assign an ID such as `vd1a`. |
| `mounted` | List partitions mounted in the current process. |
| `mkfs` | Initialize the validated EXT2-style filesystem layout. |
| `login` | Authenticate against the simulated `/users.txt` file. |
| `rep` | Generate supported Graphviz report sources. |

## Quick start

### Requirements

- Go 1.22+
- Linux, macOS, or another environment with standard Go file I/O support
- Graphviz is optional if you want to render generated `.dot` files yourself

### Run the CLI

```bash
cd Backend
go run .
```

Example interactive flow:

```text
mkdisk -size=256 -unit=k -fit=ff -path="/tmp/virtual-disks/demo.mia"
fdisk -size=192 -unit=k -type=p -fit=ff -path="/tmp/virtual-disks/demo.mia" -name=data
mount -path="/tmp/virtual-disks/demo.mia" -name=data
mounted
mkfs -id=vd1a -type=full -fs=2fs
login -user=root -pass=123 -id=vd1a
```

A longer portable command sample is available in [`examples/commands.txt`](examples/commands.txt). It uses `/tmp/virtual-disks/` rather than machine-specific paths.

The `root / 123` credential above belongs only to the filesystem simulation
created by `mkfs`; it is not an external service credential.

## Quality gates

Run the complete local quality gate from the repository root:

```bash
./scripts/check.sh
```

It checks:

1. Go formatting;
2. `go vet ./...`;
3. the full regression suite with `go test -count=1 ./...`.

The repository currently contains **14 deterministic Go tests** covering:

- command parsing;
- binary structure round trips;
- MBR and disk Graphviz output;
- virtual disk creation;
- primary partition creation and physical disk bounds;
- logical partition capacity inside an extended partition;
- mount IDs and persisted mount state;
- EXT2-style metadata and first-free pointers;
- missing-path semantics;
- exact password matching in the simulated user store.

## Key invariants

The hardened implementation enforces several storage invariants:

- top-level partitions cannot extend beyond the physical disk;
- the MBR itself is included in disk-capacity calculations;
- logical partitions must fit inside the extended partition;
- EBR reads and writes are checked for errors;
- mount letters are assigned deterministically;
- the first free inode and block after filesystem bootstrap are both index `2`;
- missing filesystem paths return `-1` rather than aliasing the root inode;
- credential matching is exact rather than substring-based.

## Repository structure

```text
.
├── Backend/
│   ├── Analyzer/         # command parsing and dispatch
│   ├── DiskManagement/   # virtual disks, MBR, partitions, EBR, mounts
│   ├── FileSystem/       # EXT2-style initialization and metadata
│   ├── Structs/          # binary storage structures
│   ├── User/             # path traversal and simulated login
│   ├── Utilities/        # binary I/O and Graphviz report generation
│   ├── go.mod
│   └── main.go
├── docs/
│   └── architecture.md
├── scripts/
│   └── check.sh
└── .github/
    └── workflows/
        └── go-ci.yml
```

## Current scope and limitations

This is a filesystem and storage **simulation**, not a production filesystem or
security subsystem.

Validated scope:

- MBR metadata;
- primary and extended partitions;
- EBR-linked logical partition creation;
- top-level partition mounting;
- EXT2-style filesystem initialization;
- root directory and `/users.txt` bootstrap;
- exact simulated login;
- MBR and disk-layout `.dot` report generation.

Known limitations:

- EXT3 is not implemented.
- Indirect inode blocks are not implemented.
- Appending data across additional file blocks is incomplete.
- `file` and `ls` report modes are placeholders.
- Fit values are accepted and persisted, but allocation does not yet implement a
  complete best-fit/first-fit/worst-fit free-hole strategy.
- Logical-partition mounting is not part of the validated feature set.
- Mount state is process-local and is not restored after restarting the CLI.
- Simulated user credentials are stored as plaintext inside `/users.txt`; this is
  deliberately not presented as production authentication.

These limitations are documented explicitly to distinguish implemented behavior
from future work.

## Design priorities

The portfolio hardening work favors:

- correctness of byte boundaries over permissive behavior;
- deterministic state over map-order-dependent behavior;
- bounded memory usage for virtual disk creation;
- explicit failure semantics;
- tests for storage invariants rather than only happy-path execution;
- honest documentation of unsupported features.

## License

No open-source license has been assigned yet. Source code is visible for
portfolio and review purposes once the repository is published.
