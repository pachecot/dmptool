# dmptool

Continuum Dump file tool

A command line tool that extracts data from dump files.

## Usage

`> dmptool [command] [option and additional args]`

## Commands

### pe

`> dmptool pe <source dumpfile> <output directory>`

Extracts all the program files from the dump file and creates a matching tree file structure in the output directory with the individual program files.

All program files are named with a `.pe` extension

