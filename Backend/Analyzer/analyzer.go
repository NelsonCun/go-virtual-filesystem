package Analyzer

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/NelsonCun/go-virtual-filesystem/DiskManagement"
	"github.com/NelsonCun/go-virtual-filesystem/FileSystem"
	"github.com/NelsonCun/go-virtual-filesystem/Structs"
	"github.com/NelsonCun/go-virtual-filesystem/User"
	"github.com/NelsonCun/go-virtual-filesystem/Utilities" // Import the Utilities package
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`-(\w+)=("[^"]+"|\S+)`)

//-nombre=valor

//input := "mkdisk -size=3000 -unit=K -fit=BF -path=/tmp/virtual-disks/disk1.mia"
// "mkdisk -size=3000 -unit=K -fit=BF -path=/tmp/virtual-disks/disk1.mia"

/*
parts[0] es "mkdisk"
*/

func getCommandAndParams(input string) (string, string) {
	parts := strings.Fields(input)
	if len(parts) > 0 {
		command := strings.ToLower(parts[0])
		params := strings.Join(parts[1:], " ")
		return command, params
	}
	return "", input

	/*Después de procesar la entrada:
	command será "mkdisk".
	params será "-size=3000 -unit=K -fit=BF -path=/tmp/virtual-disks/disk1.mia".*/
}

func Analyze() {

	for {
		var input string
		fmt.Println("======================")
		fmt.Println("Ingrese comando: ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input = scanner.Text()

		command, params := getCommandAndParams(input)

		fmt.Println("Comando: ", command, " - ", "Parametro: ", params)

		AnalyzeCommand(command, params)

		//mkdisk -size=3000 -unit=K -fit=BF -path="/tmp/virtual-disks/disk1.mia"
	}
}

func AnalyzeCommand(command string, params string) {

	switch command {
	case "mkdisk":
		fn_mkdisk(params)
	case "rmdisk":
		fn_rmdisk(params)
	case "fdisk":
		fn_fdisk(params)
	case "mount":
		fn_mount(params)
	case "mounted":
		fn_mounted()
	case "mkfs":
		fn_mkfs(params)
	case "login":
		fn_login(params)
	case "rep":
		fn_rep(params)
	default:
		fmt.Println("Error: Comando inválido o no encontrado")
	}

}

func fn_mkdisk(params string) {
	// Definir flag
	fs := flag.NewFlagSet("mkdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	fit := fs.String("fit", "ff", "Ajuste")
	unit := fs.String("unit", "m", "Unidad")
	path := fs.String("path", "", "Ruta")

	// Parse flag
	fs.Parse(strings.Fields(params))

	// Encontrar la flag en el input
	matches := re.FindAllStringSubmatch(params, -1)

	// Process the input
	for _, match := range matches {
		flagName := match[1]                   // match[1]: Captura y guarda el nombre del flag (por ejemplo, "size", "unit", "fit", "path")
		flagValue := strings.ToLower(match[2]) //strings.ToLower(match[2]): Captura y guarda el valor del flag, asegurándose de que esté en minúsculas

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "size", "fit", "unit", "path":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	/*
			Primera Iteración :
		    flagName es "size".
		    flagValue es "3000".
		    El switch encuentra que "size" es un flag reconocido, por lo que se ejecuta fs.Set("size", "3000").
		    Esto asigna el valor 3000 al flag size.

	*/

	// Validaciones
	if *size <= 0 {
		fmt.Println("Error: Size must be greater than 0")
		return
	}

	if *fit != "bf" && *fit != "ff" && *fit != "wf" {
		fmt.Println("Error: Fit must be 'bf', 'ff', or 'wf'")
		return
	}

	if *unit != "k" && *unit != "m" {
		fmt.Println("Error: Unit must be 'k' or 'm'")
		return
	}

	if *path == "" {
		fmt.Println("Error: Path is required")
		return
	}

	// LLamamos a la funcion
	DiskManagement.Mkdisk(*size, *fit, *unit, *path)
}

func fn_rmdisk(params string) {
	// Definir flag
	fs := flag.NewFlagSet("mrdisk", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")

	fs.Parse(strings.Fields(params))

	// Encontrar la flag en el input
	matches := re.FindAllStringSubmatch(params, -1)

	// Process the input
	for _, match := range matches {
		flagName := match[1]                   // match[1]: Captura y guarda el nombre del flag (por ejemplo, "path")
		flagValue := strings.ToLower(match[2]) //trings.ToLower(match[2]): Captura y guarda el valor del flag, asegurándose de que esté en minúsculas

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "path":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	// Validaciones

	if *path == "" {
		fmt.Println("Error: Path is required")
		return
	}

	// LLamamos a la funcion
	DiskManagement.Mrdisk(*path)
}

func fn_fdisk(input string) {
	// Definir flags
	fs := flag.NewFlagSet("fdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre")
	unit := fs.String("unit", "k", "Unidad")
	type_ := fs.String("type", "p", "Tipo")
	fit := fs.String("fit", "wf", "Ajuste") // Dejar fit vacío por defecto

	// Parsear los flags
	fs.Parse(strings.Fields(input))

	// Encontrar los flags en el input
	matches := re.FindAllStringSubmatch(input, -1)

	// Procesar el input
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])

		flagValue = strings.Trim(flagValue, "\"")

		switch strings.ToLower(flagName) {
		case "size", "fit", "unit", "path", "name", "type":
			fs.Set(strings.ToLower(flagName), flagValue)
		default:
			fmt.Println("Error: Flag " + flagName + " not found")
		}
	}

	// Validaciones
	if *size <= 0 {
		fmt.Println("Error: Size must be greater than 0")
		return
	}

	if *path == "" {
		fmt.Println("Error: Path is required")
		return
	}

	// Si no se proporcionó un fit, usar el valor predeterminado "w"
	if *fit == "" {
		*fit = "w"
	}

	// Validar fit (b/w/f)
	if *fit != "bf" && *fit != "ff" && *fit != "wf" {
		fmt.Println("Error: Fit must be 'bf', 'ff', or 'wf'")
		return
	}

	if *unit != "b" && *unit != "k" && *unit != "m" {
		fmt.Println("Error: Unit must be 'b', 'k' or 'm'")
		return
	}

	if *type_ != "p" && *type_ != "e" && *type_ != "l" {
		fmt.Println("Error: Type must be 'p', 'e', or 'l'")
		return
	}

	// Llamar a la función
	DiskManagement.Fdisk(*size, *path, *name, *unit, *type_, *fit)
}

func fn_mount(params string) {
	fs := flag.NewFlagSet("mount", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre de la partición")

	//fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2]) // Convertir todo a minúsculas
		flagValue = strings.Trim(flagValue, "\"")
		fs.Set(flagName, flagValue)
	}

	if *path == "" || *name == "" {
		fmt.Println("Error: Path y Name son obligatorios")
		return
	}

	// Convertir el nombre a minúsculas antes de pasarlo al Mount
	lowercaseName := strings.ToLower(*name)
	DiskManagement.Mount(*path, lowercaseName)
}

func fn_mounted() {
	DiskManagement.Mounted()
}

func fn_mkfs(input string) {
	fs := flag.NewFlagSet("mkfs", flag.ExitOnError)
	id := fs.String("id", "", "Id")
	type_ := fs.String("type", "", "Tipo")
	fs_ := fs.String("fs", "2fs", "Fs")

	// Parse the input string, not os.Args
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "id", "type", "fs":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	// Verifica que se hayan establecido todas las flags necesarias
	if *id == "" {
		fmt.Println("Error: id es un parámetro obligatorio.")
		return
	}

	if *type_ == "" { //2fs 3fs
		fmt.Println("Error: type es un parámetro obligatorio.")
		return
	}

	// Llamar a la función
	FileSystem.Mkfs(*id, *type_, *fs_)
}

func fn_login(input string) {
	// Definir las flags
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	user := fs.String("user", "", "Usuario")
	pass := fs.String("pass", "", "Contraseña")
	id := fs.String("id", "", "Id")

	// Parsearlas
	fs.Parse(os.Args[1:])

	// Match de flags en el input
	matches := re.FindAllStringSubmatch(input, -1)

	// Procesar el input
	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "user", "pass", "id":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	User.Login(*user, *pass, *id)

}

func fn_rep(input string) {
	fs := flag.NewFlagSet("rep", flag.ExitOnError)
	name := fs.String("name", "", "Nombre del reporte a generar (mbr, disk, inode, block, bm_inode, bm_block, sb, file, ls)")
	path := fs.String("path", "", "Ruta donde se generará el reporte")
	id := fs.String("id", "", "ID de la partición")
	pathFileLs := fs.String("path_file_ls", "", "Nombre del archivo o carpeta para reportes file o ls") // Parámetro opcional

	// Parsear los parámetros de entrada
	matches := re.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.Trim(match[2], "\"")

		switch flagName {
		case "name", "path", "id", "path_file_ls":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag no encontrada:", flagName)
		}
	}

	// Verificar los parámetros obligatorios
	if *name == "" || *path == "" || *id == "" {
		fmt.Println("Error: 'name', 'path' y 'id' son parámetros obligatorios.")
		return
	}

	// Verificar si el disco está montado usando DiskManagement
	mounted := false
	var diskPath string
	for _, partitions := range DiskManagement.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.ID == *id {
				mounted = true
				diskPath = partition.Path
				break
			}
		}
	}

	if !mounted {
		fmt.Println("Error: La partición con ID", *id, "no está montada.")
		return
	}

	// Crear la carpeta si no existe
	reportsDir := filepath.Dir(*path)
	err := os.MkdirAll(reportsDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error al crear la carpeta:", reportsDir)
		return
	}

	// Generar el reporte según el tipo de reporte solicitado
	switch *name {
	case "mbr":
		// Abrir el archivo binario del disco montado
		file, err := Utilities.OpenFile(diskPath)
		if err != nil {
			fmt.Println("Error: No se pudo abrir el archivo en la ruta:", diskPath)
			return
		}
		defer file.Close()

		// Leer el objeto MBR desde el archivo binario
		var TempMBR Structs.MBR
		if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
			fmt.Println("Error: No se pudo leer el MBR desde el archivo")
			return
		}

		// Leer y procesar los EBRs si hay particiones extendidas
		var ebrs []Structs.EBR
		for i := 0; i < 4; i++ {
			if string(TempMBR.Partitions[i].Type[:]) == "e" { // Partición extendida
				fmt.Println("Partición extendida encontrada: ", string(TempMBR.Partitions[i].Name[:]))

				// El primer EBR está al inicio de la partición extendida
				ebrPosition := TempMBR.Partitions[i].Start
				ebrCounter := 1

				// Leer todos los EBRs dentro de la partición extendida
				for ebrPosition != -1 {
					fmt.Printf("Leyendo EBR en posición: %d\n", ebrPosition)
					var tempEBR Structs.EBR
					if err := Utilities.ReadObject(file, &tempEBR, int64(ebrPosition)); err != nil {
						fmt.Println("Error: No se pudo leer el EBR desde el archivo")
						break
					}

					// Añadir el EBR a la lista
					ebrs = append(ebrs, tempEBR)
					fmt.Printf("EBR %d leído. Start: %d, Size: %d, Next: %d, Name: %s\n", ebrCounter, tempEBR.PartStart, tempEBR.PartSize, tempEBR.PartNext, string(tempEBR.PartName[:]))

					// Depuración: Mostrar el EBR leído
					Structs.PrintEBR(tempEBR)

					// Mover a la siguiente posición de EBR
					ebrPosition = tempEBR.PartNext
					ebrCounter++

					// Si no hay más EBRs, salir del bucle
					if ebrPosition == -1 {
						fmt.Println("No hay más EBRs en esta partición extendida.")
					}
				}
			}
		}

		// Generar el archivo .dot del MBR con EBRs
		reportPath := *path
		if err := Utilities.GenerateMBRReport(TempMBR, ebrs, reportPath, file); err != nil {
			fmt.Println("Error al generar el reporte MBR:", err)
		} else {
			fmt.Println("Reporte MBR generado exitosamente en:", reportPath)

			// Renderizar el archivo .dot a .jpg usando Graphviz
			dotFile := strings.TrimSuffix(reportPath, filepath.Ext(reportPath)) + ".dot"
			outputJpg := reportPath
			cmd := exec.Command("dot", "-Tjpg", dotFile, "-o", outputJpg)
			err = cmd.Run()
			if err != nil {
				fmt.Println("Error al renderizar el archivo .dot a imagen:", err)
			} else {
				fmt.Println("Imagen generada exitosamente en:", outputJpg)
			}
		}

	//CASE PARA EL REPORTE DISK
	case "disk":
		// Abrir el archivo binario del disco montado
		file, err := Utilities.OpenFile(diskPath)
		if err != nil {
			fmt.Println("Error: No se pudo abrir el archivo en la ruta:", diskPath)
			return
		}
		defer file.Close()

		// Leer el objeto MBR desde el archivo binario
		var TempMBR Structs.MBR
		if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
			fmt.Println("Error: No se pudo leer el MBR desde el archivo")
			return
		}

		// Leer y procesar los EBRs si hay particiones extendidas
		var ebrs []Structs.EBR
		for i := 0; i < 4; i++ {
			if string(TempMBR.Partitions[i].Type[:]) == "e" { // Partición extendida
				ebrPosition := TempMBR.Partitions[i].Start
				for ebrPosition != -1 {
					var tempEBR Structs.EBR
					if err := Utilities.ReadObject(file, &tempEBR, int64(ebrPosition)); err != nil {
						break
					}
					ebrs = append(ebrs, tempEBR)
					ebrPosition = tempEBR.PartNext
				}
			}
		}

		// Calcular el tamaño total del disco
		totalDiskSize := TempMBR.MbrSize

		// Generar el archivo .dot del DISK
		reportPath := *path
		if err := Utilities.GenerateDiskReport(TempMBR, ebrs, reportPath, file, totalDiskSize); err != nil {
			fmt.Println("Error al generar el reporte DISK:", err)
		} else {
			fmt.Println("Reporte DISK generado exitosamente en:", reportPath)

			// Renderizar el archivo .dot a .jpg usando Graphviz
			dotFile := strings.TrimSuffix(reportPath, filepath.Ext(reportPath)) + ".dot"
			outputJpg := reportPath
			cmd := exec.Command("dot", "-Tjpg", dotFile, "-o", outputJpg)
			err = cmd.Run()
			if err != nil {
				fmt.Println("Error al renderizar el archivo .dot a imagen:", err)
			} else {
				fmt.Println("Imagen generada exitosamente en:", outputJpg)
			}
		}

	case "file", "ls":
		// Para los reportes "file" y "ls", pathFileLs es obligatorio
		if *pathFileLs == "" {
			fmt.Println("Error: 'path_file_ls' es obligatorio para los reportes 'file' y 'ls'.")
			return
		}

		// Lógica para generar los reportes de tipo 'file' y 'ls'
		fmt.Println("Generando reporte", *name, "con archivo/carpeta:", *pathFileLs)
		// Aquí iría la lógica adicional para generar estos reportes

	default:
		fmt.Println("Error: Tipo de reporte no válido.")
	}
}
