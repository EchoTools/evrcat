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
	reverseMode bool
	uppercase   bool
}

func main() {

	versionInfo := fmt.Sprintf("%s %s", version, commitID)
	// Initialize flags
	flag.BoolVar(&flags.showHelp, "help", false, "Display help information")
	flag.BoolVar(&flags.reverseMode, "reverse", false, "Reverse mode: replace tokens with hashes")
	flag.BoolVar(&flags.uppercase, "uppercase", false, "Uppercase hexadecimal numbers")
	flag.BoolVar(&flags.showVersion, "version", false, fmt.Sprintf("Display version information: %s", versionInfo))
	flag.Parse()

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
				fmt.Fprintf(os.Stderr, "Error opening file `%s`: %s", file.Name(), err.Error())
				continue
			}
			defer file.Close()
			scanner = bufio.NewScanner(file)
		}

		for scanner.Scan() {
			line := scanner.Text()
			if flags.reverseMode {
				line = replaceTokens(line, flags.uppercase)
			} else {
				line = replaceSymbols(line)
			}
			// print out the line
			if _, err := fmt.Println(line); err != nil {
				fmt.Fprintln(os.Stderr, "Error writing to stdout:", err)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading file:", err)
		}
	}
}

// replaceSymbols replaces all (known) symbol hashes in a line with their token representation.
func replaceSymbols(line string) string {
	matches := symbolPattern.FindAllString(line, -1)
	replacements := make([]string, 0, len(matches)*2)
	for _, match := range matches {
		replacement := evr.ToSymbol(match).Token().String()
		if !strings.HasPrefix(replacement, "0x") {
			replacements = append(replacements, match, replacement)
		}
	}
	line = strings.NewReplacer(replacements...).Replace(line)
	return line
}

// replaceTokens replaces all tokens in a line with their hashed representation.
func replaceTokens(line string, uppercase bool) string {
	tokens := strings.Split(line, " ")
	hashes := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if len(t) == 0 {
			continue
		}
		sym := evr.ToSymbol(t)
		s := sym.HexString()
		if uppercase {
			s = strings.ToUpper(s)
		}
		hashes = append(hashes, s)
	}
	return strings.Join(hashes, " ")
}
