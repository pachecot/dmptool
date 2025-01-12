
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

#### Options Flags:
```
  -a, --all             return all the references
  -b, --bare            return just the references
  -o, --output string   output file to write to
```

### list

`>  dmptool list <dump file> [flags]`

extracts list of objects into list and prints to the console.

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