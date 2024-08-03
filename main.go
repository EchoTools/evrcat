package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/echotools/evrcat/cat"
	"github.com/heroiclabs/nakama/v3/server/evr"
	bolt "go.etcd.io/bbolt"
)

var (
	version  = "dev"
	commitID = "none"

	HashBucketName = []byte("hashes")

	flags = Flags{}
)

// Define flags
type Flags struct {
	showHelp         bool
	showVersion      bool
	reverseMode      bool
	uppercase        bool
	databasePath     string
	updateDB         bool
	files            []string
	serverListenPort string
}

func parseFlags() {
	// Initialize flags
	flag.BoolVar(&flags.showHelp, "help", false, "Display help information")
	flag.BoolVar(&flags.reverseMode, "reverse", false, "Replace tokens with hashes")
	flag.StringVar(&flags.databasePath, "db-path", "~/.cache/evrcat/lookup.db", "load cache database from `PATH`")
	flag.BoolVar(&flags.updateDB, "update-db", false, "Update the database (only works with --reverse)")
	flag.BoolVar(&flags.uppercase, "uppercase", false, "Use uppercase hexadecimal strings")
	flag.StringVar(&flags.serverListenPort, "server", "", "Run as a server on `PORT` (reverse not supported)")
	flag.BoolVar(&flags.showVersion, "version", false, "Display version")
	flag.Parse()

	if flags.showHelp {
		fmt.Fprintln(os.Stderr, "Usage: evrcat [OPTION]... [FILE]...")
		fmt.Fprintln(os.Stderr, "Concatenate FILE(s) to standard output, replacing hashes with tokens.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "With no FILE, or when FILE is -, read standard input.")
		fmt.Fprintln(os.Stderr, "")
		flag.Usage()
		fmt.Fprintln(os.Stderr, "")
		os.Exit(0)
	}

	if flags.showVersion {
		fmt.Fprintln(os.Stdout, version)
		os.Exit(0)
	}

	if flags.serverListenPort != "" {
		// Build the map from the built-in symbol cache
		hashmap := make(map[string]string, len(evr.SymbolCache))
		for k, v := range evr.SymbolCache {
			hashmap[k.HexString()] = string(v)
			hashmap[k.HexStringUpper()] = string(v)
		}
		fmt.Fprintf(os.Stderr, "Loaded %d entries from built-in symbol cache\n", len(hashmap))

		server := cat.NewEVRCatServer(hashmap)
		server.Start(flags.serverListenPort)
	}

	flags.files = flag.Args()
	if len(flags.files) == 0 {
		flags.files = append(flags.files, "-")
	}

}

func main() {
	var err error
	parseFlags()

	// The hash map is only populated if the user wants to update the database.
	var hashmap map[evr.Symbol]evr.SymbolToken
	var db *bolt.DB

	if flags.databasePath != "" {
		if db, err = openDB(flags.databasePath); err != nil {
			fmt.Fprintln(os.Stderr, "Error opening database:", err.Error())
		}
		hashmap, err = readDB(db)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading database:", err.Error())
		}
		fmt.Fprintf(os.Stderr, "Loaded %d entries from %s\n", len(hashmap), flags.databasePath)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c
		if db != nil {
			// If reversing values, then update the cache with new values
			if flags.updateDB && flags.reverseMode {
				if err := updateDB(db, hashmap); err != nil {
					fmt.Fprintln(os.Stderr, "Error updating database:", err.Error())
				}
			}
			if err := closeDB(db); err != nil {
				fmt.Fprintln(os.Stderr, "Error closing database:", err.Error())
			}
		}
		os.Exit(0)
	}()

	cat := cat.NewEVRCat()

	if flags.reverseMode && flags.updateDB {
		// Start with the built-in symbol cache
		hashmap, err = readDB(db)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading database:", err.Error())
			os.Exit(1)
		}
		// Update the map with the built-in symbol cache
		for k, v := range evr.SymbolCache {
			hashmap[k] = v
		}
	}

	for _, file := range flags.files {
		var scanner *bufio.Scanner
		if file == "-" {
			scanner = bufio.NewScanner(os.Stdin)
		} else {

			file, err := os.Open(file)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error opening file:", file.Name(), err.Error())
				continue
			}
			defer file.Close()
			scanner = bufio.NewScanner(file)
		}

		for scanner.Scan() {
			line := scanner.Text()
			if flags.reverseMode {
				line = cat.ReplaceTokens(line, flags.uppercase, hashmap)
			} else {
				line = cat.ReplaceHashes(line)
			}
			// print out the line
			if _, err := fmt.Println(line); err != nil {
				fmt.Fprintln(os.Stderr, "Error writing to stdout:", err)
				os.Exit(1)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading file:", err)
			continue
		}

		if db != nil {
			if err := db.Sync(); err != nil {
				fmt.Fprintln(os.Stderr, "Error syncing database:", err.Error())
			}
		}
	}
	// If reversing values, then update the cache with new values
	if flags.updateDB && flags.reverseMode {
		if err := updateDB(db, hashmap); err != nil {
			fmt.Fprintln(os.Stderr, "Error updating database:", err.Error())
		}
	}
	if err := closeDB(db); err != nil {
		fmt.Fprintln(os.Stderr, "Error closing database:", err.Error())
	}
}

func openDB(path string) (*bolt.DB, error) {
	// Translate the path to an absolute path, including home directory expansion
	path = filepath.Clean(path)
	if path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}

func readDB(db *bolt.DB) (map[evr.Symbol]evr.SymbolToken, error) {
	hashes := make(map[evr.Symbol]evr.SymbolToken)
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(HashBucketName)
		if bucket == nil {
			// Database doesn't exist (yet)
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			hash := binary.LittleEndian.Uint64(k)
			token := evr.SymbolToken(v)
			hashes[evr.Symbol(hash)] = token
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read database: %w", err)
	}

	return hashes, nil
}

func updateDB(db *bolt.DB, hashes map[evr.Symbol]evr.SymbolToken) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(HashBucketName)
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}

		for hash, token := range hashes {
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(hash))
			if err := bucket.Put(b, []byte(token)); err != nil {
				return err
			}
		}
		return nil
	})
}

func closeDB(db *bolt.DB) error {
	if db == nil {
		return nil
	}
	return db.Close()
}
