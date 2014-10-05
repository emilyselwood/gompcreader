# Minor Planet Center Reader #

## Overview ##

Simple Go library to read the minor planet center data files.

The go docs should be reasonable. If some thing doesn't seem to work please raise a bug.

The expected input files can be obtained here: http://www.minorplanetcenter.net/iau/MPCORB.html

## Example ##

```
package main

import (
  "fmt"
  "io"
  "strconv"
  "flag"
  "log"
  "github.com/wselwood/gompcreader"
)


var inputfile = flag.String("in", "", "the minor planet center file to read")

func main() {

  flag.Parse()

  if *inputfile == "" {
    log.Fatal("No input file provided. Use the -in /path/to/file")
  }


  mpcReader, err := gompcreader.NewMpcReader(inputfile)
  if err != nil {
    log.Fatal("error creating mpcReader ", err)
  }

  var count int64 = 0
  result, err := mpcReader.ReadEntry()
  for err == nil {
    fmt.Printf("%s:%s\n", result.Id, result.ReadableDesignation)
    result, err = mpcReader.ReadEntry()
    count = count + 1
  }

  if err != nil && err != io.EOF {
    log.Fatalf("error reading line %d\n", count, err)
  }

  fmt.Printf("read %d records\n", count)
}
```
