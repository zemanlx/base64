package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
)

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
