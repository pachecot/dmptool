package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tpacheco/dmptool/cmds/pe"
	"github.com/tpacheco/dmptool/cmds/ref"
)

const version = "0.1.2"

func newCmdPE() *cobra.Command {
	pe := &pe.Command{}
	cc := &cobra.Command{
		Use:   "pe <dump file> <output directory>",
		Short: "extract all PE program files into individual files a directory structure",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			pe.FileName = args[0]
			pe.OutDir = args[1]
			pe.Execute()
		},
	}

	cc.Flags().BoolVarP(&pe.TypeFolder, "type", "t", false, "incudes the typename in the folders")

	return cc
}

func newCmdRef() *cobra.Command {
	cmdRef := &ref.Command{}
	cc := &cobra.Command{
		Use:   "ref <dump file>",
		Short: "list external references in the dump file",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmdRef.FileName = args[0]
			cmdRef.Execute()
		},
	}

	cc.Flags().StringVarP(&cmdRef.OutFile, "output", "o", "", "output file to write to")
	cc.Flags().BoolVarP(&cmdRef.Bare, "bare", "b", false, "return just the references")
	cc.Flags().BoolVarP(&cmdRef.All, "all", "a", false, "return all the references")
	cc.Flags().BoolVarP(&cmdRef.Sources, "sources", "s", false, "return all the sources")

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
