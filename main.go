package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/raspi/audiogroup-extractor/reader"
	"io"
	"os"
	"path"
	"time"
)

func main() {

	flag.Usage = func() {
		f := path.Base(os.Args[0])
		fmt.Println(`Usage:`)
		fmt.Printf(`  %v <file>`+"\n", f)
		fmt.Println(`Example:`)
		fmt.Printf(`  %v audiogroup1.dat`+"\n", f)
		os.Exit(0)
	}

	flag.Parse()

	if flag.NArg() != 1 {
		_, _ = fmt.Fprintf(os.Stderr, `no file given, see --help for usage`)
		os.Exit(1)
	}

	sourceFile := flag.Arg(0)
	fi, err := os.Stat(sourceFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `could not stat file %q: %v`, sourceFile, err)
		os.Exit(1)
	}

	if fi.IsDir() {
		_, _ = fmt.Fprintf(os.Stderr, `file %q is a directory`, sourceFile)
		os.Exit(1)
	}

	startTime := time.Now()

	f, err := os.Open(sourceFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `could not open file: %v`, err)
		os.Exit(1)
	}
	defer f.Close()

	bname := path.Base(sourceFile)

	rdr, err := reader.New(f)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `error reading: %v`, err)
		os.Exit(1)
	}

	tracks := rdr.Tracks()

	fmt.Printf(`Found %d tracks in %q`+"\n", len(tracks), sourceFile)

	for idx, t := range tracks {

		offset, err := f.Seek(t.Offset, io.SeekStart)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, `could not seek to %v`, t.Offset)
			os.Exit(1)
		}

		if t.Size == 0 {
			_, _ = fmt.Fprintf(os.Stderr, `skipping (size 0) at %08d`, offset)
			continue
		}

		buffer := make([]byte, t.Size)

		readbytes, err := f.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			_, _ = fmt.Fprintf(os.Stderr, `could not read source file %q: %v`, sourceFile, err)
			os.Exit(1)
		}

		if readbytes <= 4 {
			_, _ = fmt.Fprintf(os.Stderr, `skipping (size not > 4) at %08d`, offset)
			continue
		}

		// Default extension (unknown)
		fext := `dat`

		// Guess some extension(s)
		switch string(buffer[:4]) {
		case `RIFF`:
			fext = `wav`
		}

		fname := fmt.Sprintf(`dump-%v-%03d-%08x.%v`, bname, idx, offset, fext)

		// Open new file
		newfile, err := os.Create(fname)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, `could not create %q: %v`, fname, err)
			os.Exit(1)
		}

		// Write into new file
		writtenbytes, err := newfile.Write(buffer[:readbytes])
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, `could not write to %q: %v`, fname, err)
			os.Exit(1)
		}

		err = newfile.Close()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, `could not close %q: %v`, fname, err)
			os.Exit(1)
		}

		fmt.Printf(`Wrote #%03d %q offset:%v size:%v bytes`+"\n", idx+1, fname, offset, writtenbytes)
	}

	fmt.Println()
	fmt.Printf(`Took %v`+"\n", time.Now().Sub(startTime))
	fmt.Println()
	fmt.Println("Use `file`, `ffprobe` or other such tool to determine the correct file extension(s) for dump-*.dat files")

}
