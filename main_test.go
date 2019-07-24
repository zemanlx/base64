package main

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func init() {
	version = "0.0.0"
	commit = "deadbeef"
	date = "2000-02-20"
}
func Test_getEncoding(t *testing.T) {
	type args struct {
		noPadding bool
		url       bool
	}
	tests := []struct {
		name         string
		args         args
		wantEncoding *base64.Encoding
	}{
		{"with padding and standard encoding = StdEncoding", args{false, false}, base64.StdEncoding},
		{"with padding and URL encoding = URLEncoding", args{false, true}, base64.URLEncoding},
		{"no padding and standard encoding = RawStdEncoding", args{true, false}, base64.RawStdEncoding},
		{"no padding and URL encoding = RawURLEncoding", args{true, true}, base64.RawURLEncoding},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEncoding := getEncoding(tt.args.noPadding, tt.args.url)
			if !reflect.DeepEqual(gotEncoding, tt.wantEncoding) {
				t.Errorf("getEncoding() gotEncoding = %v, want %v", gotEncoding, tt.wantEncoding)
			}
		})
	}
}

func Test_getFile_stdin(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name     string
		args     args
		wantFile *os.File
		wantErr  bool
	}{
		{"no file as input should return os.Stdin", args{""}, os.Stdin, false},
		{"- as input should return os.Stdin", args{"-"}, os.Stdin, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, err := getFile(tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFile.Name(), tt.wantFile.Name()) {
				t.Errorf("getFile() = %v, want %v", gotFile.Name(), tt.wantFile.Name())
			}
		})
	}
}

func Test_getFile_realFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	gotFile, err := getFile(tmpfile.Name())
	if err != nil {
		t.Errorf("getFile(%s) error = %v", tmpfile.Name(), err)
	}
	if gotFile.Name() != tmpfile.Name() {
		t.Errorf("getFile(%s) = %s, want %s", tmpfile.Name(), gotFile.Name(), tmpfile.Name())
	}
}

func Test_printHelp(t *testing.T) {
	type args struct {
		programName string
	}
	tests := []struct {
		name        string
		args        args
		wantMessage string
	}{
		{
			"print help for program hulahop",
			args{"hulahop"},
			`Usage: hulahop [OPTION]... [FILE]

Base64 encode or decode FILE, or standard input, to standard output.
With no FILE, or when FILE is -, read standard input.


The data are encoded as described for the base64 alphabet in RFC 4648.
When decoding, the input may contain newlines in addition to the bytes of
the formal base64 alphabet.  Use --ignore-garbage to attempt to recover
from any other non-alphabet bytes in the encoded stream.
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMessage []byte
			stderr := os.Stderr
			defer func() { os.Stderr = stderr }()
			r, fakeStderr, err := os.Pipe()
			if err != nil {
				t.Fatalf("cannot create os.Pipe(): %v", err)
			}
			defer r.Close()
			os.Stderr = fakeStderr
			printHelp(tt.args.programName)
			fakeStderr.Close()
			gotMessage, err = ioutil.ReadAll(r)
			if err != nil {
				t.Fatalf("cannot read fakeStderr: %v", err)
			}
			if diff := cmp.Diff(string(gotMessage), tt.wantMessage); diff != "" {
				t.Errorf("printHelp() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func Test_printVersion(t *testing.T) {
	tests := []struct {
		name        string
		wantMessage string
	}{
		{
			"print help",
			"Version:  0.0.0\nCommit:   deadbeef\nDate:     2000-02-20\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			printVersion(buf)
			gotMessage := buf.String()
			if diff := cmp.Diff(string(gotMessage), tt.wantMessage); diff != "" {
				t.Errorf("printHelp() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
