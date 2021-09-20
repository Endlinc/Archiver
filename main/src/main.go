package main

import (
	"Archiver/archiver"
	"Archiver/parser"
	"flag"
	"fmt"
)

var includeFiles string
var excludeFiles string
var outputPath string

func Init() {
	flag.StringVar(&includeFiles, "includes", "/usr", "declare files to be included")
	flag.StringVar(&excludeFiles, "excludes", "/usr", "declare files to be excluded")
	flag.StringVar(&outputPath, "output", "/tmp/archive.tar.gz", "declare files to be generated")
	flag.Parse()
}

func main() {
	fmt.Println("Let's see how it goes.")
	Init()
	fmt.Printf("includes: %s\n", includeFiles)
	fmt.Printf("excludes: %s\n", excludeFiles)
	includes, iErr := parser.ReadFileContent(includeFiles)
	if iErr != nil {
		fmt.Println(iErr)
		return
	}
	excludes, eErr := parser.ReadFileContent(excludeFiles)
	if eErr != nil {
		fmt.Println(eErr)
		return
	}
	compressor := new(archiver.Compressor)
	if err := compressor.Init(outputPath); err != nil {
		fmt.Println(err.Error())
		return
	}
	compressor.LoadPaths(includes, true)
	compressor.LoadPaths(excludes, false)
	compressor.AddAllPredecessors()
	if err := compressor.Archive(); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := compressor.Close(); err != nil {
		fmt.Println(err.Error())
		return
	}
	return
}
