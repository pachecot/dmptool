# dmptool

Continuum Dump file tool

A command line tool that extracts data from dump files.

## Usage

`> dmptool [command] [options and additional args]`

### Commands

    pe   : export script        
    list : list objects          
    ref  : get reference information         
    tree : display contents as a tree          

## pe command 

`> dmptool pe <source dump file> <output directory>`

Extracts all the program files from the dump file and creates a matching tree file structure in the output directory with the individual program files.

All program files are named with a `.pe` extension

### Aliases: 

      script, code, programs

### Options Flags:

```
  -f, --flatten            flatten the file path to a single name
  -h, --help               help for pe
      --prefix string      prefix used when including file types (default "__")
      --separator string   separator used when flattening file paths (default "~")
  -t, --type               include the typename as a folder for the files.

```

## ref Command 

`> dmptool ref <source dump file>`

Scans all the program files from the dump file and creates a list of the external references.

### Aliases:

    references, refs

### Options Flags:

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

## list command

`>  dmptool list <dump file> [flags]`

extracts list of objects into list and prints to the console.

output files with the csv extension will be formatted as csv files.

### Option Flags:

```
  -d, --devices strings   filter with matching device ids / paths
  -f, --fields strings    fields (default [DeviceId,Name,Type])
  -h, --help              help for list
  -n, --names strings     filter with matching names
  -o, --output string     output file to write to
  -s, --sort strings      sort order list by fields
  -t, --types strings     types filter
  -w, --where string      where filter
```
> --output string

output results to a file. uses file extension for supported file types (csv and xlsx). all other files are treated as plain text columnar data.

> --devices strings

filters the results based on device id or path

> --fields string

giving "?" to the field string will generate the list of fields available.

the argument is a comma separated list of fields to list in the result.

`--fields "DeviceId,Name,Type,Channel,Format,ElecType.ElecScaleBottom,ElecScaleTop,EngScaleBottom,EngScaleTop"`

> --sort strings

fields to sort the output order of the results. Use DESC for descending order. Ascending (ASC) is default order.

`--sort "DeviceId,Type,Channel"`

> --where string

the where filter is a sql like where clause for filtering the results. 

`--where "Type IN ('InfinityInput','InfinityOutput')"`

#### supported operators

`IN` `BETWEEN` `LIKE` `<` `<=` `>` `>=` `!= ` `<>` `=` `==` `ISNULL`
`ISNOTNULL` `NOT` `AND` `OR` 

#### examples:

  - `<field> IN ('value1', 'value2', ...)`
  - `<field> BETWEEN 'value1' AND 'value2'`
  - `<field> LIKE <pattern>`
  - `<field> < 10`
  - `<field> <= 10`
  - `<field> > 10`
  - `<field> >= 10`
  - `<field> != 10`
  - `<field> <> 10`
  - `<field> ISNULL`
  - `<field> ISNOTNULL`
  - `NOT <filter>`
  - `<field> = 10 OR <field> = 1`
  - `<field> > 1 AND <field> < 5`
  - `Type IN ('InfinityOutput', 'InfinityInput') `
  - `Channel BETWEEN 1 AND 10`
  - `Name LIKE '%Temp%'`
  - `Type LIKE 'Infinity%put'`
  - `Type IN ('InfinityOutput', 'InfinityInput') AND ElecType ISNOTNULL`


## tree command

`> dmptool tree <source dump file> [flags]`

Display dump file as a tree.

### Options Flags:

```
  -h, --help            help for ref
  -a, --ascii           use ascii characters for tree
  -n, --depth int       limit the depth of the tree
  -p, --parent          display parents or container objects only
  -o, --output string   output file to write to
```

