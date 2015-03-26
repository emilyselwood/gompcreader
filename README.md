# Minor Planet Center Reader #

[![Build Status](https://travis-ci.org/wselwood/gompcreader.svg?branch=master)](https://travis-ci.org/wselwood/gompcreader)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/wselwood/gompcreader/blob/master/LICENSE.md)
[![Docs](https://img.shields.io/badge/license-MIT-blue.svg)]https://godoc.org/github.com/wselwood/gompcreader)

## Overview ##

Simple Go library to read the minor planet center data files.

The go docs should be reasonable. If some thing doesn't seem to work please raise a bug.

The expected input files can be obtained here: http://www.minorplanetcenter.net/iau/MPCORB.html Either the gziped or unzipped files should work automatically, file type detection is done by simple file extension only.

## Example ##

```
package main

import (
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/wselwood/gompcreader"
)

var inputfile = flag.String("in", "", "the minor planet center file to read")

func main() {

	flag.Parse()

	if *inputfile == "" {
		log.Fatal("No input file provided. Use the -in /path/to/file")
	}

	mpcReader, err := gompcreader.NewMpcReader(*inputfile)
	if err != nil {
		log.Fatal("error creating mpcReader ", err)
	}

	var count int64
	result, err := mpcReader.ReadEntry()
	for err == nil {
		fmt.Printf("%s:%s\n", result.ID, result.ReadableDesignation)
		result, err = mpcReader.ReadEntry()
		count = count + 1
	}

	if err != nil && err != io.EOF {
		log.Fatal(fmt.Sprintf("error reading line %d\n", count), err)
	}

	fmt.Printf("read %d records\n", count)
}

```
