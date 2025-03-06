package main

import (
	"fmt"

	_ "embed"

	"github.com/spf13/cobra"
	"github.com/tpacheco/dmptool/cmds/list"
	"github.com/tpacheco/dmptool/cmds/pe"
	"github.com/tpacheco/dmptool/cmds/ref"
	"github.com/tpacheco/dmptool/cmds/tree"
)

var (
	// version is the version of the tool
	//go:embed VERSION
	version string

	// Date is the build date of the tool
	Date = ""
)

func newCmdPE() *cobra.Command {
	pe := &pe.Command{}

	cc := &cobra.Command{
		Use:   "pe <dump file> [output directory]",
		Short: "extracts all PE program files into individual files",
		Long: `This command will extract all the PE program files from the dump file. 

The PE program files will be written to the output directory. The output directory 
can be specified as the second argument. If the, output directory is not specified, 
then the files will be written to the current directory. 

A directory will be created for each object. The directory will be named with the 
device path of the object. The PE program file will be written to the directory. 
The PE program file will be named with the object name and the extension .pe.

The output files can be flattened to a single name. If the --flatten flag is set, 
then the file path will be flattened to a single name. The separator used to flatten 
the file path can be specified with the --separator flag. The default separator is "~". 

The output files can be organized by the type of object. If the --type flag is set, 
then the files will be organized by the type of object. The type of object will be used 
as a sub folder for the files.

The --prefix flag set a prefix for the type folders. The default prefix is "__". 

		DeviceName
			__Program
			__InfinityProgram
			__InfinityFunction
`,
		Aliases: []string{"script", "code", "programs"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pe.FileName = args[0]
			if len(args) > 1 {
				pe.OutDir = args[1]
			}
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
		Use:   "list <dump file>",
		Short: "extracts list of objects into list",
		Long: `This command will extract the list of objects from the dump file. The list can be filtered by the type of object,
the name of the object, the device id or path, and the properties of the object. The output can be written to a file in csv or xlsx format.	

The output file can be specified with the --output flag. If the file extension is .csv, then the output will be in csv format.
If the file extension is .xlsx, then the output will be in xlsx format. If the file extension is not recognized, then the output
will be written as a text. If the output is not specified then the output is to stout.

The fields to include in the output can be specified with the --fields flag. The default fields are DeviceId, Name, and Type. The fields
can be any of the properties of the object. The fields can be specified multiple times to include multiple fields. The fields can also be
specified with the --fields flag in the format of "field1,field2,field3".

The types of objects to include in the output can be specified with the --types flag. The types can be specified multiple times to include
multiple types. The types can also be specified with the --types flag in the format of "type1,type2,type3".

The names of the objects to include in the output can be specified with the --names flag. The names can be specified multiple times to include
multiple names. The names can also be specified with the --names flag in the format of "name1,name2,name3". The names are matched with the Name
property of the object.

The device ids or paths of the objects to include in the output can be specified with the --devices flag. The device ids or paths can be specified
multiple times to include multiple device ids or paths. The device ids or paths can also be specified with the --devices flag in the format of
"device1,device2,device3". The device ids or paths are matched with the Path property of the object. The device ids or paths can be partial matches.

The properties of the objects to include in the output can be specified with the --where flag. The where filter can be specified multiple times to
include multiple filters. The where filter can also be specified with the --where flag in the format of "field1 op1 value1,field2 op2 value2,field3 op3 value3".
The where filter can be used to filter the objects based on the properties of the object. The where filter can be used to filter the objects based on
the properties of the object. The operators that can be used are "=", ">", "<", ">=", "<=", "like", and "@". The "like" operator is used to match a
substring of the property value. The "@" operator is used to match a substring of the Name or Path properties. The where filter can be used to filter
the objects based on the properties of the object. The where filter can be used to filter the objects based on the properties of the object. The operators
that can be used are "=", ">", "<", ">=", "<=", "like", and "@". The "like" operator is used to match a substring of the property value. The "@" operator
is used to match a substring of the Name or Path properties. The where filter can be used to filter the objects based on the properties of the object. The
where filter can be used to filter the objects based on the properties of the object. The operators that can be used are "=", ">", "<", ">=", "<=", and "like".
The "like" operator is used to match a substring of the property value. If no operator is used the parameter will try to match the substring of the Name and Path
 properties.
`,
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
	cc.Flags().StringSliceVarP(&listCmd.Fields, "fields", "f", []string{"DeviceId", "Name", "Type"}, "list of fields to include")
	cc.Flags().StringSliceVarP(&listCmd.Types, "types", "t", []string{}, "types filter")

	return cc
}

func newCmdRef() *cobra.Command {
	cmdRef := &ref.Command{}
	cc := &cobra.Command{
		Use:   "ref <dump file>",
		Short: "list external references in the dump file",
		Long: `This command will list all the external references in the dump file. The references
can be filtered by the type of reference. The output can be written to a file in csv or xlsx format.

If the --all flag is set, all the internal and external references will be listed.

If the --bare flag is set, only the references will be listed to the console. 

If the --code, --graphics, or --alarms flags are set, the references will be filtered by the type of reference.
The flags can be combined to include multiple types of references. If none of the flags are set, then just the 
code references will be listed.

If the --source flag is set, the source path will be included in the output.

The output file can be specified with the --output flag. If the file extension is .csv, then the output will be in csv format.
If the file extension is .xlsx, then the output will be in xlsx format. If the file extension is not recognized, then the output
will be written as a text. If the output is not specified then the output is to stout.
`,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"references", "refs"},
		Run: func(cmd *cobra.Command, args []string) {
			if !(cmdRef.Code || cmdRef.Graphics || cmdRef.Alarms) {
				cmdRef.Code = true
			}
			cmdRef.FileName = args[0]
			cmdRef.Execute()
		},
	}

	cc.Flags().StringVarP(&cmdRef.OutFile, "output", "o", "", "output file to write to. default is stdout")
	cc.Flags().BoolVarP(&cmdRef.Bare, "bare", "b", false, "return just the references")
	cc.Flags().BoolVarP(&cmdRef.All, "all", "a", false, "return all the references")
	cc.Flags().BoolVarP(&cmdRef.Sources, "source", "s", false, "show the source path")
	cc.Flags().BoolVarP(&cmdRef.Code, "code", "c", false, "include the script code sources (default)")
	cc.Flags().BoolVarP(&cmdRef.Graphics, "graphics", "g", false, "include the graphics sources")
	cc.Flags().BoolVarP(&cmdRef.Alarms, "alarms", "l", false, "include the alarm link sources")
	cc.Flags().BoolVarP(&cmdRef.ShowType, "typename", "t", false, "show typename in path")

	return cc
}

func newCmdTree() *cobra.Command {
	cmdTree := &tree.Command{}
	cc := &cobra.Command{
		Use:   "tree <dump file>",
		Short: "display the object tree",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmdTree.FileName = args[0]
			cmdTree.Execute()
		},
	}

	cc.Flags().StringVarP(&cmdTree.OutFile, "output", "o", "", "output file to write to. default is stdout")
	cc.Flags().BoolVarP(&cmdTree.Ascii, "ascii", "a", false, "use ascii characters for tree")
	cc.Flags().BoolVarP(&cmdTree.Parents, "parents", "p", false, "container objects only")
	cc.Flags().IntVarP(&cmdTree.Depth, "depth", "n", 0, "max depth of tree")
	return cc
}

func newCmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("version: %s  build date: %s\n", version, Date)
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
		newCmdTree(),
		newCmdList(),
		newCmdVersion(),
	)
	cc.Execute()
}
