package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/heroiclabs/nakama/v3/server/evr"
)

var (
	version  = "dev"
	commitID = "none"

	flags = Flags{}

	symbolPattern = regexp.MustCompile(`0x[0-9a-fA-F]{16}`)
)

// Define flags
type Flags struct {
	showHelp    bool
	showVersion bool
}

func init() {

	versionInfo := fmt.Sprintf("%s %s", version, commitID)
	// Initialize flags
	flag.BoolVar(&flags.showHelp, "help", false, "Display help information")
	flag.BoolVar(&flags.showVersion, "version", false, fmt.Sprintf("Display version information: %s", versionInfo))
	flag.Parse()
}
func main() {
	if flags.showHelp {
		fmt.Println("Usage: evrcat [OPTION]... [FILE]...")
		fmt.Println("\nConcatenate FILE(s) to standard output, replacing hexadecimal numbers with tokens.")
		fmt.Println("")
		fmt.Println("With no FILE, or when FILE is -, read standard input.")
		fmt.Println("")
		fmt.Println("--help: Display help information")
		fmt.Println("--version: Display version information")

		os.Exit(0)
	}

	if flags.showVersion {
		fmt.Printf("evrcat %s, commit %s\n", version, commitID)
		os.Exit(0)
	}

	files := flag.Args()
	if len(files) == 0 {
		files = append(files, "-")
	}

	for _, file := range files {
		var scanner *bufio.Scanner
		if file == "-" {
			scanner = bufio.NewScanner(os.Stdin)
		} else {

			file, err := os.Open(file)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error opening file:", err)
				continue
			}
			defer file.Close()
			scanner = bufio.NewScanner(file)
		}

		for scanner.Scan() {
			line := scanner.Text()
			matches := symbolPattern.FindAllString(line, -1)
			replacements := make([]string, 0, len(matches)*2)
			for _, match := range matches {
				replacement := evr.ToSymbol(match).Token().String()
				if !strings.HasPrefix(replacement, "0x") {
					replacements = append(replacements, match, replacement)
				}

			}
			line = strings.NewReplacer(replacements...).Replace(line)
			fmt.Println(line)
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading file:", err)
		}
	}
}

func replaceSymbols(line string) string {
	matches := symbolPattern.FindAllString(line, -1)
	replacements := make([]string, 0, len(matches)*2)
	for _, match := range matches {
		replacement := evr.ToSymbol(match).Token().String()
		if strings.HasPrefix(replacement, "0x") {
			replacement = strings.ToUpper(replacement)
		}
		replacements = append(replacements, match, replacement)
	}
	return strings.NewReplacer(replacements...).Replace(line)
}
