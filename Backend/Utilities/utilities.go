package Utilities

import (
	"encoding/binary"
	"fmt"
	"github.com/NelsonCun/go-virtual-filesystem/Structs"
	"html"
	"os"
	"path/filepath"
	"strings"
)

// Funcion para crear un archivo binario
func CreateFile(name string) error {
	//Se asegura que el archivo existe
	dir := filepath.Dir(name)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Println("Err CreateFile dir==", err)
		return err
	}

	// Crear archivo
	if _, err := os.Stat(name); os.IsNotExist(err) {
		file, err := os.Create(name)
		if err != nil {
			fmt.Println("Err CreateFile create==", err)
			return err
		}
		defer file.Close()
	}
	return nil
}

// Funcion para abrir un archivo binario ead/write mode
func OpenFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Err OpenFile==", err)
		return nil, err
	}
	return file, nil
}

// Funcion para escribir un objecto en un archivo binario
func WriteObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Write(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Err WriteObject==", err)
		return err
	}
	return nil

}

// Funcion para leer un objeto de un archivo binario
func ReadObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Read(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Err ReadObject==", err)
		return err
	}
	return nil
}

func cleanFixedString(data []byte) string {
	return strings.TrimRight(string(data), "\x00")
}

func dotEscape(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"\n", "\\n",
		"\r", "",
	)
	return replacer.Replace(value)
}

func reportDotPath(reportPath string) string {
	ext := filepath.Ext(reportPath)
	if ext == "" {
		return reportPath + ".dot"
	}
	return strings.TrimSuffix(reportPath, ext) + ".dot"
}

// GenerateMBRReport writes a Graphviz representation of the MBR, its primary/
// extended partition table, and any EBRs discovered by the caller.
func GenerateMBRReport(mbr Structs.MBR, ebrs []Structs.EBR, reportPath string, _ *os.File) error {
	var b strings.Builder

	b.WriteString("digraph MBR {\n")
	b.WriteString("  graph [rankdir=TB, bgcolor=\"white\"];\n")
	b.WriteString("  node [shape=plaintext, fontname=\"Helvetica\"];\n")
	b.WriteString("  mbr [label=<\n")
	b.WriteString("    <table border=\"1\" cellborder=\"1\" cellspacing=\"0\" cellpadding=\"6\">\n")
	b.WriteString("      <tr><td colspan=\"2\"><b>Master Boot Record</b></td></tr>\n")
	fmt.Fprintf(&b, "      <tr><td>Disk size</td><td>%d bytes</td></tr>\n", mbr.MbrSize)
	fmt.Fprintf(&b, "      <tr><td>Created</td><td>%s</td></tr>\n", html.EscapeString(cleanFixedString(mbr.CreationDate[:])))
	fmt.Fprintf(&b, "      <tr><td>Signature</td><td>%d</td></tr>\n", mbr.Signature)
	fmt.Fprintf(&b, "      <tr><td>Fit</td><td>%s</td></tr>\n", html.EscapeString(cleanFixedString(mbr.Fit[:])))
	b.WriteString("    </table>\n")
	b.WriteString("  >];\n")

	for i, p := range mbr.Partitions {
		if p.Size <= 0 {
			continue
		}

		name := html.EscapeString(cleanFixedString(p.Name[:]))
		ptype := html.EscapeString(cleanFixedString(p.Type[:]))
		fit := html.EscapeString(cleanFixedString(p.Fit[:]))
		status := html.EscapeString(cleanFixedString(p.Status[:]))
		id := html.EscapeString(cleanFixedString(p.Id[:]))

		fmt.Fprintf(&b, "  p%d [label=<\n", i)
		b.WriteString("    <table border=\"1\" cellborder=\"1\" cellspacing=\"0\" cellpadding=\"6\">\n")
		fmt.Fprintf(&b, "      <tr><td colspan=\"2\"><b>Partition %d</b></td></tr>\n", i+1)
		fmt.Fprintf(&b, "      <tr><td>Name</td><td>%s</td></tr>\n", name)
		fmt.Fprintf(&b, "      <tr><td>Type</td><td>%s</td></tr>\n", ptype)
		fmt.Fprintf(&b, "      <tr><td>Fit</td><td>%s</td></tr>\n", fit)
		fmt.Fprintf(&b, "      <tr><td>Start</td><td>%d</td></tr>\n", p.Start)
		fmt.Fprintf(&b, "      <tr><td>Size</td><td>%d bytes</td></tr>\n", p.Size)
		fmt.Fprintf(&b, "      <tr><td>Status</td><td>%s</td></tr>\n", status)
		fmt.Fprintf(&b, "      <tr><td>ID</td><td>%s</td></tr>\n", id)
		b.WriteString("    </table>\n")
		b.WriteString("  >];\n")
		fmt.Fprintf(&b, "  mbr -> p%d;\n", i)
	}

	for i, ebr := range ebrs {
		name := html.EscapeString(cleanFixedString(ebr.PartName[:]))
		fmt.Fprintf(&b, "  ebr%d [label=<\n", i)
		b.WriteString("    <table border=\"1\" cellborder=\"1\" cellspacing=\"0\" cellpadding=\"6\">\n")
		fmt.Fprintf(&b, "      <tr><td colspan=\"2\"><b>EBR %d</b></td></tr>\n", i+1)
		fmt.Fprintf(&b, "      <tr><td>Name</td><td>%s</td></tr>\n", name)
		fmt.Fprintf(&b, "      <tr><td>Fit</td><td>%c</td></tr>\n", ebr.PartFit)
		fmt.Fprintf(&b, "      <tr><td>Start</td><td>%d</td></tr>\n", ebr.PartStart)
		fmt.Fprintf(&b, "      <tr><td>Size</td><td>%d bytes</td></tr>\n", ebr.PartSize)
		fmt.Fprintf(&b, "      <tr><td>Next</td><td>%d</td></tr>\n", ebr.PartNext)
		b.WriteString("    </table>\n")
		b.WriteString("  >];\n")

		if i == 0 {
			b.WriteString("  mbr -> ebr0 [style=dashed, label=\"logical chain\"];\n")
		} else {
			fmt.Fprintf(&b, "  ebr%d -> ebr%d;\n", i-1, i)
		}
	}

	b.WriteString("}\n")

	dotPath := reportDotPath(reportPath)
	dir := filepath.Dir(dotPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create report directory: %w", err)
		}
	}
	if err := os.WriteFile(dotPath, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("write MBR report: %w", err)
	}
	return nil
}

// GenerateDiskReport writes a compact Graphviz view of the physical disk layout.
func GenerateDiskReport(
	mbr Structs.MBR,
	ebrs []Structs.EBR,
	reportPath string,
	_ *os.File,
	totalDiskSize int32,
) error {
	var b strings.Builder
	var used int64

	mbrSize := int64(binary.Size(mbr))
	if mbrSize < 0 {
		mbrSize = 0
	}
	used += mbrSize

	b.WriteString("digraph DISK {\n")
	b.WriteString("  graph [rankdir=LR, bgcolor=\"white\"];\n")
	b.WriteString("  node [shape=record, fontname=\"Helvetica\"];\n")
	b.WriteString("  disk [label=\"{MBR")

	for _, p := range mbr.Partitions {
		if p.Size <= 0 {
			continue
		}
		name := dotEscape(cleanFixedString(p.Name[:]))
		ptype := dotEscape(cleanFixedString(p.Type[:]))
		fmt.Fprintf(&b, "|%s (%s)\\n%d bytes", name, ptype, p.Size)
		used += int64(p.Size)
	}

	free := int64(totalDiskSize) - used
	if free < 0 {
		free = 0
	}
	fmt.Fprintf(&b, "|Free\\n%d bytes", free)
	b.WriteString("}\"];\n")

	for i, ebr := range ebrs {
		if ebr.PartSize <= 0 {
			continue
		}
		name := dotEscape(cleanFixedString(ebr.PartName[:]))
		fmt.Fprintf(
			&b,
			"  logical%d [shape=box, label=\"Logical: %s\\nStart: %d\\nSize: %d bytes\"];\n",
			i,
			name,
			ebr.PartStart,
			ebr.PartSize,
		)
		fmt.Fprintf(&b, "  disk -> logical%d [style=dashed];\n", i)
	}

	b.WriteString("}\n")

	dotPath := reportDotPath(reportPath)
	dir := filepath.Dir(dotPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create report directory: %w", err)
		}
	}
	if err := os.WriteFile(dotPath, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("write disk report: %w", err)
	}
	return nil
}
