package commands

import (
	"archive/zip"
	"efa/infra"
	sett "efa/infra/cli/commands/fabric/settings"
	"efa/infra/cli/utils"
	"efa/infra/constants"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//SupportSaveCommand provides command to take snapshot of running instance e.g. db, logs etc
var SupportSaveCommand = &cobra.Command{
	Use:   "supportsave",
	Short: "Initiate the support save",
	RunE:  utils.TimedRunE(runSupportSave),
}

func init() {
}

func runSupportSave(cmd *cobra.Command, args []string) error {
	fmt.Println("Version :", infra.Version)
	fmt.Println("Time Stamp:", infra.BuildStamp)
	err := CopyAppInfo()
	SSInputs := []string{constants.DBLocation, constants.DBLocation + ".autobk", constants.LogPathToArchove, constants.AppInfoLocation}
	SSOutput := constants.SSArchive
	err = SupportSaveArchive(SSInputs, SSOutput)
	CleanupPostSS()
	return err
}

// CleanupPostSS does cleanup of temp files created by support save command
func CleanupPostSS() {
	err := os.Remove(constants.AppInfoLocation)
	if err != nil {
		fmt.Println("Failed to cleanup")
	}
}

// CopyAppInfo take snapshot of application specific informations
func CopyAppInfo() error {
	_, err := os.Stat(constants.AppInfoLocation)
	if os.IsExist(err) {
		// if the file exist delete it
		err = os.Remove(constants.AppInfoLocation)
		if err != nil {
			return err
		}
	}
	file, err := os.Create(constants.AppInfoLocation)
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString("Version Information")
	file.WriteString("\n====================\n\n")
	fmt.Fprintf(file, "Version :%s", infra.Version)
	fmt.Fprintf(file, "\nTime Stamp %s:", infra.BuildStamp)

	file.WriteString("\n\nFabric Settings")
	file.WriteString("\n=================\n\n")
	table := tablewriter.NewWriter(file)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"Name", "Value"})

	sett.GetFabricShowOut(table)
	table.Render()
	return nil
}

// SupportSaveArchive archives the files to create support save data
func SupportSaveArchive(SSInputs []string, SSOutput string) error {
	newfile, err := os.Create(SSOutput)
	if err != nil {
		return err
	}
	defer newfile.Close()

	zipWriter := zip.NewWriter(newfile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range SSInputs {
		err = AddToArchive(file, zipWriter)
		if err != nil {
			fmt.Printf("Failed to add file %s in SS", file)
		}
	}
	return nil
}

// AddToArchive adds a file or directory to an archive
func AddToArchive(source string, archive *zip.Writer) error {

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}
