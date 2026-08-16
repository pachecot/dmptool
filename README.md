# dmptool

Continuum Dump file tool

A command line tool that extracts data from dump files.

## Usage

`> dmptool [command] [options and additional args]`

### Commands

        pe         : export script
        list       : list objects
        ref        : get reference information
        tree       : display contents as a tree
        report     : generate structured reports

## pe command

`> dmptool pe <source dump file> <output directory>`

Extracts all the program files from the dump file and creates a matching tree file structure in the output directory with the individual program files.

All program files are named with a `.pe` extension

### Aliases:

                script, code, programs

### Options Flags:

```
        -f, --flatten                                       flatten the file path to a single name
        -h, --help                                         help for pe
                --prefix string                             prefix used when including file types (default "__")
                --separator string                         separator used when flattening file paths (default "~")
        -t, --type                                         include the typename as a folder for the files.

```

## ref Command

`> dmptool ref <source dump file>`

Scans all the program files from the dump file and creates a list of the external references.

### Aliases:

        references, refs

### Options Flags:

```
        -a, --all                                   return all the references
        -b, --bare                                 return just the references
        -c, --code                                 include the script code (default true)
        -g, --graphics                             include the graphics
        -h, --help                                 help for ref
        -l, --alarms                               include the alarm link sources
        -o, --output string                       output file to write to
        -s, --source                               show the source path
        -t, --typename                             show typename in path     
```

## list command

`>   dmptool list <dump file> [flags]`

Extracts list of objects into list and prints to the console.

Output files with the csv extension will be formatted as csv files.

### Option Flags:

```
        -d, --devices strings               filter with matching device ids / paths
        -f, --fields strings                 fields (default [DeviceId,Name,Type])
        -h, --help                         help for list
        -n, --names strings                 filter with matching names
        -o, --output string               output file to write to
        -s, --sort strings                 sort order list by fields
        -t, --types strings               types filter
        -w, --where string                 where filter
```
> --output string

Output results to a file. Uses file extension for supported file types (csv and xlsx). All other files are treated as plain text columnar data.

> --devices strings

Filters the results based on device id or path

> --fields string

Giving `?` to the field string will generate the list of fields available.

The argument is a comma separated list of fields to list in the result.

`--fields "DeviceId,Name,Type,Channel,Format,ElecType.ElecScaleBottom,ElecScaleTop,EngScaleBottom,EngScaleTop"`

> --sort strings

Fields to sort the output order of the results. Use DESC for descending order. Ascending (ASC) is default order.

`--sort "DeviceId,Type,Channel"`

> --where string

The where filter is a sql like where clause for filtering the results.

`--where "Type IN ('InfinityInput', 'InfinityOutput')"`

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
        -h, --help                                         help for tree
        -a, --ascii                                       use ascii characters for tree
        -n, --depth int                                   limit the depth of the tree
        -p, --parents                                     display parents or container objects only
        -o, --output string                               output file to write to
        -v, --view                                         page results in interactive view
```

## report command

`> dmptool report <subcommand> <dump file> [flags]`

Generates structured reports from a dump file. Each subcommand produces a fixed-column report for a specific object category. Output can be written to the console or to a file (csv or xlsx).

### Subcommands

        io                : IO points with electrical and engineering scale configuration
        points            : Point summary (String, Numeric, Input, Output, DateTime)
        alarms            : Points with alarms, expanded one row per alarm link
        logs              : Points with logging configured
        alarm-enrollments : AlarmEnrollment objects with parsed event parameters
        event-enrollments : EventEnrollment objects with parsed event parameters and object property reference

---

### report io

`> dmptool report io <dump file> [flags]`

Reports all InfinityInput and InfinityOutput points with their electrical and engineering scale configuration.

#### Aliases:

        inputs, outputs

#### Columns:

        DeviceId, Type, Name, Alias, Channel, ElecType,
        ElecScaleTop, ElecScaleBot, EngScaleTop, EngScaleBot, Conversion

#### Option Flags:

```
        -d, --devices strings               filter by device path
        -h, --help                         help for io
        -n, --names strings                 filter by point name
        -o, --output string               output file (csv, xlsx, or text)
```

---

### report points

`> dmptool report points <dump file> [flags]`

Reports a summary of all point objects: InfinityString, InfinityNumeric, InfinityInput, InfinityOutput, and InfinityDateTime.

#### Aliases:

        summary

#### Columns:

        DeviceId, Type, Name, Alias, Value, Setpoint

#### Option Flags:

```
        -d, --devices strings               filter by device path
        -h, --help                         help for points
        -n, --names strings                 filter by point name
        -o, --output string               output file (csv, xlsx, or text)
        -t, --types strings               filter by point type (e.g. InfinityNumeric)
```

---

### report alarms

`> dmptool report alarms <dump file> [flags]`

Reports all points that have alarm links configured. Each alarm link is expanded as an individual row.

#### Aliases:

        alarm

#### Columns:

        DeviceId, Name, Alarm, Enabled

#### Option Flags:

```
        -d, --devices strings               filter by device path
        -h, --help                         help for alarms
        -n, --names strings                 filter by point name
        -o, --output string               output file (csv, xlsx, or text)
```

---

### report logs

`> dmptool report logs <dump file> [flags]`

Reports all points that have logging configured.

#### Aliases:

        log

#### Columns:

        DeviceId, Name, LogType, Size, LogIntervalSecs

#### Option Flags:

```
        -d, --devices strings               filter by device path
        -h, --help                         help for logs
        -n, --names strings                 filter by point name
        -o, --output string               output file (csv, xlsx, or text)
```

---

### report alarm-enrollments

`> dmptool report alarm-enrollments <dump file> [flags]`

Reports all AlarmEnrollment objects with their event parameters parsed from the EventParameters CDT block. Supports both ACCExpression (type 340) and HiLimit/LoLimit (type 337) event types.

#### Aliases:

        ae, enrollments

#### Columns:

        Name, Alias, Description, EventType,
        TimeDelay, ACCExpression, Limit, Normal,
        EventNotificationId, MsgTextOffNormal, MsgTextNormal

#### Option Flags:

```
        -d, --devices strings               filter by device path
        -h, --help                         help for alarm-enrollments
        -n, --names strings                 filter by point name
        -o, --output string               output file (csv, xlsx, or text)
```

---

### report event-enrollments

`> dmptool report event-enrollments <dump file> [flags]`

Reports all EventEnrollment objects with their event parameters parsed from the EventParameters CDT block and the monitored point extracted from the ObjPropRef block. Handles ChangeOfState, ChangeOfValue, and OutOfRange event types.

#### Aliases:

        ee

#### Columns:

        Name, Alias, Description, EventType,
        TimeDelay, ObjPropRef, StateValues, COVIncrement, LowLimit, HighLimit, Deadband,
        EventNotificationId, MsgTextOffNormal, MsgTextNormal

| Column | Populated for |
|---|---|
| TimeDelay | all event types |
| ObjPropRef | all event types — the monitored point path |
| StateValues | ChangeOfState — comma-separated triggering state values |
| COVIncrement | ChangeOfValue — the COV increment threshold |
| LowLimit / HighLimit / Deadband | OutOfRange |

#### Option Flags:

```
        -d, --devices strings               filter by device path
        -h, --help                         help for event-enrollments
        -n, --names strings                 filter by point name
        -o, --output string               output file (csv, xlsx, or text)
```
