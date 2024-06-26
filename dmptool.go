package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tpacheco/dmptool/cmds/list"
	"github.com/tpacheco/dmptool/cmds/pe"
	"github.com/tpacheco/dmptool/cmds/ref"
)

const version = "0.3.1"

func newCmdPE() *cobra.Command {
	pe := &pe.Command{}

	cc := &cobra.Command{
		Use:     "pe <dump file> <output directory>",
		Short:   "extract all PE program files into individual files a directory structure",
		Aliases: []string{"script", "code", "programs"},
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			pe.FileName = args[0]
			pe.OutDir = args[1]
			pe.Execute()
		},
	}

	cc.Flags().BoolVarP(&pe.TypeFolder, "type", "t", false, "incudes the typename in the folders")

	return cc
}

func newCmdList() *cobra.Command {

	listCmd := &list.Command{}

	cc := &cobra.Command{
		Use:     "list <dump file>",
		Short:   "extract all PE program files into individual files a directory structure",
		Aliases: []string{"script", "code", "programs"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			listCmd.FileName = args[0]
			listCmd.Execute()
		},
	}

	cc.Flags().StringVarP(&listCmd.OutFile, "output", "o", "", "output file to write to")
	cc.Flags().StringVarP(&listCmd.Filter, "where", "w", "", "where filter")
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

func main() {

	cc := &cobra.Command{
		Use:   "dmptool <command>",
		Short: "continuum dump file tool",
	}

	cc.AddCommand(
		newCmdPE(),
		newCmdRef(),
		newCmdList(),
		&cobra.Command{
			Use:   "version",
			Short: "print the version",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("version:", version)
			},
		},
	)
	cc.Execute()
}
