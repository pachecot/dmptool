
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

### ref

`> dmptool ref <source dump file>`

Scans all the program files from the dump file and creates a list of the external references.

#### Options Flags:
```
  -a, --all             return all the references
  -b, --bare            return just the references
  -o, --output string   output file to write to
```
