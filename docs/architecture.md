# Architecture

## 1. System boundary

The application is an interactive Go command-line program that creates and
manipulates virtual disk files. It models selected disk-management and
filesystem concepts in userspace; it does not mount a kernel filesystem.

The primary boundary is a `.mia` binary file whose offsets contain serialized Go
structures.

```text
User input
   |
   v
Analyzer
   |
   +--> DiskManagement
   +--> FileSystem
   +--> User
           |
           v
       Utilities
           |
           v
   Random-access binary file
```

## 2. Components

### Analyzer

Responsibilities:

- split an input line into command and parameter text;
- parse command parameters;
- dispatch supported operations;
- route `rep` requests into report generation.

The analyzer intentionally remains a thin orchestration layer. Storage
semantics belong to the domain packages below it.

### DiskManagement

Responsibilities:

- create and remove virtual disk files;
- initialize/read/write the MBR;
- create primary and extended partitions;
- maintain an EBR chain for logical partitions;
- enforce physical disk and extended-partition boundaries;
- mount validated top-level partitions;
- assign deterministic IDs such as `vd1a`;
- keep process-local mount state.

Virtual disk creation uses `os.File.Truncate` rather than allocating a byte slice
equal to the full virtual disk size.

### FileSystem

Responsibilities:

- locate a mounted partition;
- calculate the EXT2-style storage layout;
- initialize the superblock;
- create inode/block bitmaps;
- initialize the inode table;
- create the root directory;
- create the initial `/users.txt` file.

Bootstrap ownership/indexing used by the validated implementation:

```text
inode 0 -> root directory
block 0 -> root directory block

inode 1 -> /users.txt
block 1 -> /users.txt data

first free inode -> 2
first free block -> 2
```

### User

Responsibilities:

- traverse directory blocks by exact entry name;
- return `-1` for paths that do not exist;
- read `/users.txt`;
- perform exact username/password comparison;
- update process-local login state.

This is deliberately a filesystem simulation. Credentials stored in
`/users.txt` are plaintext and must not be interpreted as a production
authentication design.

### Structs

The structures package defines the serialized storage model, including:

- `MBR`
- `Partition`
- `EBR`
- `Superblock`
- `Inode`
- `Folderblock`
- `Content`
- `Fileblock`
- `Pointerblock`
- `MountedPartition`

Fixed-size byte arrays are used where the binary layout requires bounded fields.

### Utilities

Responsibilities:

- create/open files;
- write structures at explicit offsets;
- read structures at explicit offsets;
- encode/decode data using `encoding/binary` with little-endian ordering;
- generate Graphviz `.dot` sources for MBR and disk-layout reports.

## 3. Virtual disk model

### Top-level MBR layout

```text
offset 0
  |
  v
+--------------------------+
| MBR                      |
| - disk size              |
| - creation date          |
| - signature              |
| - fit metadata           |
| - 4 partition slots      |
+--------------------------+
| partition payloads       |
| ...                      |
+--------------------------+
disk end
```

Primary and extended partitions consume top-level disk capacity. Boundary
validation includes the serialized MBR size so no partition can extend beyond
the virtual disk.

## 4. Extended and logical partitions

A logical partition is represented through an EBR descriptor stored inside an
extended partition.

```text
extended.Start
      |
      v
+---------+-------------------+---------+-------------------+
| EBR #1  | logical payload 1 | EBR #2  | logical payload 2 |
+---------+-------------------+---------+-------------------+
```

For every logical creation:

1. traverse `PartNext` until the final EBR;
2. calculate the next EBR position;
3. calculate logical payload start after the descriptor;
4. verify logical end <= extended partition end;
5. update the previous `PartNext` when needed;
6. persist the new EBR.

I/O failures while traversing or writing EBRs are treated as operation failures.

## 5. EXT2-style partition layout

The validated filesystem initialization computes the following regions:

```text
partition.Start
      |
      v
+------------+
| Superblock |
+------------+
| inode BM   |
+------------+
| block BM   |
+------------+
| inode table|
+------------+
| data blocks|
+------------+
partition end
```

The superblock contains counts, free-space metadata, timestamps, magic
`0xEF53`, object sizes, and offsets for each region.

The current project models selected EXT2 concepts; it is not intended to be
binary-compatible with a host operating system's ext2 driver.

## 6. Mount model

Mount state is stored in memory as a map keyed by a normalized disk path.

A disk receives the first unused letter from `a` through `z`. Partitions on the
same disk reuse that disk letter. A mounted partition receives an ID in the
validated form:

```text
vd<partition-number><disk-letter>
```

Example:

```text
vd1a
```

Because the registry is process-local, restarting the CLI clears the mount
registry even though mount metadata may exist in the virtual disk structures.

## 7. Reports

Supported report helpers produce Graphviz source files:

- MBR report;
- physical disk-layout report.

The implementation writes `.dot` output. Rendering to PNG/JPG/SVG is an
optional external Graphviz step and is intentionally separate from the core
storage engine.

## 8. Quality model

The regression suite checks observable storage invariants rather than only
package compilation.

Current deterministic coverage includes:

- parser behavior;
- binary serialization round trips;
- report source generation;
- disk size and MBR metadata;
- primary partition metadata;
- physical disk bounds;
- logical partition bounds;
- persisted mount metadata;
- EXT2-style superblock invariants;
- bootstrap inode/block pointers;
- path-not-found semantics;
- exact login behavior.

The CI workflow executes the same local `scripts/check.sh` gate used by
developers.
