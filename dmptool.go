package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tpacheco/dmptool/cmds/list"
	"github.com/tpacheco/dmptool/cmds/pe"
	"github.com/tpacheco/dmptool/cmds/ref"
)

const version = "0.5.4"

var (
	// Version is the version of the tool
	Version = version

	// Date is the build date of the tool
	Date = ""
)

func newCmdPE() *cobra.Command {
	pe := &pe.Command{}

	cc := &cobra.Command{
		Use:     "pe <dump file> <output directory>",
		Short:   "extracts all PE program files into individual files",
		Aliases: []string{"script", "code", "programs"},
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			pe.FileName = args[0]
			pe.OutDir = args[1]
			pe.Execute()
		},
	}

	cf := cc.Flags()
	cf.BoolVarP(&pe.TypeFolder, "type", "t", false, "include the typename as a folder for the files")
	cf.BoolVarP(&pe.Flatten, "flatten", "f", false, "flatten the file path to a single name")
	cf.StringVar(&pe.FlattenSep, "separator", "~", "separator used when flattening file paths")
	cf.StringVar(&pe.TypePrefix, "prefix", "__", "prefix used when including file types")

	return cc
}

func newCmdList() *cobra.Command {

	listCmd := &list.Command{}

	cc := &cobra.Command{
		Use:     "list <dump file>",
		Short:   "extracts list of objects into list",
		Aliases: []string{},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			listCmd.FileName = args[0]
			listCmd.Execute()
		},
	}

	cc.Flags().StringVarP(&listCmd.OutFile, "output", "o", "", "output file to write to")
	cc.Flags().StringSliceVarP(&listCmd.Filters, "where", "w", []string{}, "where like filter")
	cc.Flags().StringSliceVarP(&listCmd.Names, "names", "n", []string{}, "filter with matching names")
	cc.Flags().StringSliceVarP(&listCmd.Devices, "devices", "d", []string{}, "filter with matching device ids / paths")
	cc.Flags().BoolVarP(&listCmd.Record, "records", "r", false, "list in a record format")
	cc.Flags().StringSliceVarP(&listCmd.Fields, "fields", "f", []string{"DeviceId", "Name", "Type"}, "fields")
	cc.Flags().StringSliceVarP(&listCmd.Types, "types", "t", []string{}, "types filter")

	return cc
}

func newCmdRef() *cobra.Command {
	cmdRef := &ref.Command{}
	cc := &cobra.Command{
		Use:     "ref <dump file>",
		Short:   "list external references in the dump file",
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"references", "refs"},
		Run: func(cmd *cobra.Command, args []string) {
			cmdRef.FileName = args[0]
			cmdRef.Execute()
		},
	}

	cc.Flags().StringVarP(&cmdRef.OutFile, "output", "o", "", "output file to write to")
	cc.Flags().BoolVarP(&cmdRef.Bare, "bare", "b", false, "return just the references")
	cc.Flags().BoolVarP(&cmdRef.All, "all", "a", false, "return all the references")
	cc.Flags().BoolVarP(&cmdRef.Sources, "source", "s", false, "show the source path")
	cc.Flags().BoolVarP(&cmdRef.Code, "code", "c", true, "include the script code")
	cc.Flags().BoolVarP(&cmdRef.Graphics, "graphics", "g", false, "include the graphics")
	cc.Flags().BoolVarP(&cmdRef.ShowType, "typename", "t", false, "show typename in path")

	return cc
}

func newCmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("version: %s  build date: %s\n", Version, Date)
		},
	}
}

func main() {

	cc := &cobra.Command{
		Use:   "dmptool <command>",
		Short: "continuum dump file tool",
		Args:  cobra.MinimumNArgs(2),
	}
	cc.AddCommand(
		newCmdPE(),
		newCmdRef(),
		newCmdList(),
		newCmdVersion(),
	)
	cc.Execute()
}
