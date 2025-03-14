
# dmptool

Continuum Dump file tool

A command line tool that extracts data from dump files.

## Usage

`> dmptool [command] [option and additional args]`

## Commands

### pe

`> dmptool pe <source dump file> <output directory>`

Extracts all the program files from the dump file and creates a matching tree file structure in the output directory with the individual program files.

All program files are named with a `.pe` extension

#### Aliases: 

      script, code, programs

#### Options Flags:

```
  -f, --flatten            flatten the file path to a single name
  -h, --help               help for pe
      --prefix string      prefix used when including file types (default "__")
      --separator string   separator used when flattening file paths (default "~")
  -t, --type               include the typename as a folder for the files.

```

### ref

`> dmptool ref <source dump file>`

Scans all the program files from the dump file and creates a list of the external references.

#### Aliases:

    references, refs

#### Options Flags:

```
  -a, --all             return all the references
  -b, --bare            return just the references
  -c, --code            include the script code (default true)
  -g, --graphics        include the graphics
  -h, --help            help for ref
  -o, --output string   output file to write to
  -s, --source          show the source path
  -t, --typename        show typename in path  
```

### list

`>  dmptool list <dump file> [flags]`

extracts list of objects into list and prints to the console.

output files with the csv extension will be formatted as csv files.

#### Option Flags:

```
  -d, --devices strings   filter with matching device ids / paths
  -f, --fields strings    fields (default [DeviceId,Name,Type])
  -h, --help              help for list
  -n, --names strings     filter with matching names
  -o, --output string     output file to write to
  -r, --records           list in a record format
  -t, --types strings     types filter
  -w, --where strings     where like filter
```

### tree

`> dmptool tree <source dump file> [flags]`

Display dump file as a tree.

#### Options Flags:

```
  -h, --help            help for ref
  -a, --ascii           use ascii characters for tree
  -n, --depth int       limit the depth of the tree
  -p, --parent          display parents or container objects only
  -o, --output string   output file to write to
```
