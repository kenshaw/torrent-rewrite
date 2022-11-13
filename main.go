package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	bencode "github.com/IncSW/go-bencode"
)

func main() {
	dir := flag.String("dir", "./resume", "dir")
	flag.Parse()
	if err := run(*dir); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(dir string) error {
	var files []string
	if err := filepath.Walk(dir, func(n string, fi fs.FileInfo, err error) error {
		switch {
		case err != nil:
			return err
		case fi.IsDir(), filepath.Ext(n) != ".resume":
			return nil
		}
		files = append(files, n)
		return nil
	}); err != nil {
		return err
	}
	for _, file := range files {
		buf, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		x, err := bencode.Unmarshal(buf)
		if err != nil {
			return err
		}
		v, ok := x.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid type %T", x)
		}
		dest, ok := v["destination"]
		if !ok {
			return fmt.Errorf("bad file: %s", file)
		}
		s, ok := dest.([]uint8)
		switch {
		case !ok:
			return fmt.Errorf("bad file 2: %s", file)
		case strings.HasPrefix(string(s), "/media/stuff/"):
			v["destination"] = []byte("/media/vol0/" + strings.TrimPrefix(string(s), "/media/stuff/"))
			buf, err := bencode.Marshal(v)
			if err != nil {
				return fmt.Errorf("unable to encode %s: %v", file, err)
			}
			if err := os.WriteFile(file, buf, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}
