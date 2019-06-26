package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	flag "github.com/spf13/pflag"

	"github.com/zemanlx/base64/xbase"
)

var (
	version string // build time variable
	commit  string // build time variable
)

func main() {
	var (
		decode        = flag.BoolP("decode", "d", false, "decode data")
		ignoreGarbage = flag.BoolP("ignore-garbage", "i", false, "when decoding, ignore non-alphabet characters")
		noPadding     = flag.BoolP("no-padding", "n", false, "omit padding")
		url           = flag.BoolP("url", "u", false, "use URL encoding according RFC4648")
		wrapAfter     = flag.UintP("wrap", "w", 76, "wrap encoded lines after COLS character,\nuse 0 to disable line wrapping")
		showVersion   = flag.BoolP("version", "v", false, "output version information and exit")
		help          = flag.BoolP("help", "h", false, "print this help")
	)
	flag.Parse()

	if *help {
		printHelp(filepath.Base(os.Args[0]))
		os.Exit(0)
	}

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	encoding := getEncoding(*noPadding, *url)

	file, err := getFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer file.Close()

	if !*decode { // encode
		if err = xbase.Encode64(file, os.Stdout, encoding, *wrapAfter); err != nil {
			fmt.Fprintf(os.Stderr, "encode pipeline error: %v\n", err)
			os.Exit(1)
		}
	}

	if *decode {
		if err = xbase.Decode64(file, os.Stdout, encoding, *ignoreGarbage); err != nil {
			fmt.Fprintf(os.Stderr, "encode pipeline error: %v\n", err)
			os.Exit(1)
		}
	}
}

func printHelp(programName string) {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]... [FILE]\n", programName)
	fmt.Fprintf(os.Stderr, `
Base64 encode or decode FILE, or standard input, to standard output.
With no FILE, or when FILE is -, read standard input.

`)
	flag.PrintDefaults() // print to STDERR
	fmt.Fprintf(os.Stderr, `
The data are encoded as described for the base64 alphabet in RFC 4648.
When decoding, the input may contain newlines in addition to the bytes of
the formal base64 alphabet.  Use --ignore-garbage to attempt to recover
from any other non-alphabet bytes in the encoded stream.
`)
}

func printVersion() {
	const padding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	fmt.Fprintf(w, "Version:\t%s\n", version)
	fmt.Fprintf(w, "Commit:\t%s\n", commit)
	w.Flush()
}

func getEncoding(noPadding, url bool) (encoding *base64.Encoding) {
	switch {
	case noPadding && url:
		encoding = base64.RawURLEncoding
	case !noPadding && url:
		encoding = base64.URLEncoding
	case noPadding && !url:
		encoding = base64.RawStdEncoding
	case !noPadding && !url:
		encoding = base64.StdEncoding
	}
	return
}

func getFile(fileName string) (file *os.File, err error) {
	if fileName == "" || fileName == "-" {
		return os.Stdin, nil
	}
	if file, err = os.Open(fileName); err != nil {
		return nil, fmt.Errorf("cannot open %s: %v", fileName, err)
	}
	return file, nil
}
