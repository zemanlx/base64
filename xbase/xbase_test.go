package xbase

import (
	"bytes"
	"encoding/base64"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var update = flag.Bool("update", false, "update .golden files")

func Test_plainEncode(t *testing.T) {
	type args struct {
		input    io.Reader
		encoding *base64.Encoding
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantErr    bool
	}{
		{"empty input", args{strings.NewReader(""), base64.StdEncoding}, "", false},
		{"simple input", args{strings.NewReader("simple"), base64.StdEncoding}, "c2ltcGxl", false},
		{"日本 input", args{strings.NewReader("日本"), base64.StdEncoding}, "5pel5pys", false},
		{"standard encoding alphabet (/) with padding", args{strings.NewReader("lo£"), base64.StdEncoding}, "bG/Cow==", false},
		{"URL encoding alphabet (_) with padding", args{strings.NewReader("lo£"), base64.URLEncoding}, "bG_Cow==", false},
		{"standard encoding alphabet (/) with no padding", args{strings.NewReader("lo£"), base64.RawStdEncoding}, "bG/Cow", false},
		{"URL encoding alphabet (_) with no padding", args{strings.NewReader("lo£"), base64.RawURLEncoding}, "bG_Cow", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := plainEncode(tt.args.input, output, tt.args.encoding); (err != nil) != tt.wantErr {
				t.Errorf("plainEncode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotOutput := output.String(); gotOutput != tt.wantOutput {
				t.Errorf("plainEncode() = %v, want %v", gotOutput, tt.wantOutput)
			}
		})
	}
}

func Benchmark_plainEncode(b *testing.B) {
	testInput := "testdata/utf8.encode.input"
	input, err := os.Open(testInput)
	if err != nil {
		b.Fatalf("cannot open %s: %v", testInput, err)
	}
	defer input.Close()
	inputStats, err := input.Stat()
	if err != nil {
		b.Fatalf("cannot get stats for %s: %v", testInput, err)
	}
	inputStats.Size()

	buf := make([]byte, inputStats.Size())
	output := bytes.NewBuffer(buf)

	for i := 0; i < b.N; i++ {
		if _, err = input.Seek(0, io.SeekStart); err != nil {
			b.Fatalf("cannot seek input file %s: %v", testInput, err)
		}
		if err = plainEncode(input, output, base64.StdEncoding); err != nil {
			b.Fatalf("plainEncode(%s, output) = %v", input.Name(), err)
		}
	}
}

func Test_plainDecode(t *testing.T) {
	type args struct {
		input    io.Reader
		encoding *base64.Encoding
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantErr    bool
	}{
		{"empty output", args{strings.NewReader(""), base64.StdEncoding}, "", false},
		{"simple output", args{strings.NewReader("c2ltcGxl"), base64.StdEncoding}, "simple", false},
		{"日本 output", args{strings.NewReader("5pel5pys"), base64.StdEncoding}, "日本", false},
		{"standard encoding alphabet (/) with padding", args{strings.NewReader("bG/Cow=="), base64.StdEncoding}, "lo£", false},
		{"URL encoding alphabet (_) with padding", args{strings.NewReader("bG_Cow=="), base64.URLEncoding}, "lo£", false},
		{"standard encoding alphabet (/) with no padding", args{strings.NewReader("bG/Cow"), base64.RawStdEncoding}, "lo£", false},
		{"URL encoding alphabet (_) with no padding", args{strings.NewReader("bG_Cow"), base64.RawURLEncoding}, "lo£", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := plainDecode(tt.args.input, output, tt.args.encoding); (err != nil) != tt.wantErr {
				t.Errorf("plainDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOutput := output.String(); gotOutput != tt.wantOutput {
				t.Errorf("plainDecode() = %v, want %v", gotOutput, tt.wantOutput)
			}
		})
	}
}

func Benchmark_plainDecode(b *testing.B) {
	testInput := "testdata/utf8.decode.input"
	input, err := os.Open(testInput)
	if err != nil {
		b.Fatalf("cannot open %s: %v", testInput, err)
	}
	defer input.Close()
	inputStats, err := input.Stat()
	if err != nil {
		b.Fatalf("cannot get stats for %s: %v", testInput, err)
	}
	inputStats.Size()

	buf := make([]byte, inputStats.Size())
	output := bytes.NewBuffer(buf)

	for i := 0; i < b.N; i++ {
		if _, err = input.Seek(0, io.SeekStart); err != nil {
			b.Fatalf("cannot seek input file %s: %v", testInput, err)
		}
		if err = plainDecode(input, output, base64.StdEncoding); err != nil {
			b.Fatalf("plainDecode(%s, output) = %v", input.Name(), err)
		}
	}
}

func Test_Encode64(t *testing.T) {
	type args struct {
		fileName  string
		encoding  *base64.Encoding
		wrapAfter uint
	}
	tests := []struct {
		name         string
		args         args
		wantFileName string
		wantErr      bool
	}{
		{"Standard encoding with padding and no wrap", args{"testdata/100c.encode.input", base64.StdEncoding, 0}, "testdata/100c.encode.std.wrap-0.padded.golden", false},
		{"URL encoding with padding and no wrap", args{"testdata/100c.encode.input", base64.URLEncoding, 0}, "testdata/100c.encode.url.wrap-0.padded.golden", false},
		{"URL encoding with no padding and no wrap", args{"testdata/100c.encode.input", base64.RawURLEncoding, 0}, "testdata/100c.encode.url.wrap-0.no-padded.golden", false},
		{"Standard encoding with no padding and no wrap", args{"testdata/100c.encode.input", base64.RawStdEncoding, 0}, "testdata/100c.encode.std.wrap-0.no-padded.golden", false},
		{"Standard encoding with padding and wrap after 76 (default)", args{"testdata/100c.encode.input", base64.StdEncoding, 76}, "testdata/100c.encode.std.wrap-76.padded.golden", false},
		{"Standard encoding with padding and wrap after 137 (length of encoded line)", args{"testdata/100c.encode.input", base64.StdEncoding, 137}, "testdata/100c.encode.std.wrap-137.padded.golden", false},
		{"Standard encoding with padding and wrap after 200 (longer than length of encoded line)", args{"testdata/100c.encode.input", base64.StdEncoding, 200}, "testdata/100c.encode.std.wrap-200.padded.golden", false},
		{"Standard encoding with padding and no wrap", args{"testdata/utf8.encode.input", base64.StdEncoding, 0}, "testdata/utf8.encode.std.wrap-0.padded.golden", false},
		{"URL encoding with padding and no wrap", args{"testdata/utf8.encode.input", base64.URLEncoding, 0}, "testdata/utf8.encode.url.wrap-0.padded.golden", false},
		{"URL encoding with no padding and no wrap", args{"testdata/utf8.encode.input", base64.RawURLEncoding, 0}, "testdata/utf8.encode.url.wrap-0.no-padded.golden", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare file input
			file, err := os.Open(tt.args.fileName)
			if err != nil {
				t.Fatalf("cannot open %s: %v", tt.args.fileName, err)
			}
			defer file.Close()

			output := &bytes.Buffer{}

			// Execute
			err = Encode64(file, output, tt.args.encoding, tt.args.wrapAfter)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode64() error = %v, wantErr %v", err, tt.wantErr)
			}

			var gotOutput, wantOutput []byte
			gotOutput = output.Bytes()
			if !*update {
				wantOutput, err = ioutil.ReadFile(tt.wantFileName)
				if err != nil {
					t.Fatalf("Cannot read file %s: %v", tt.wantFileName, err)
				}
			} else {
				if err = ioutil.WriteFile(tt.wantFileName, gotOutput, 0644); err != nil {
					t.Fatalf("Cannot write to file %s: %v", tt.wantFileName, err)
				}
				return
			}

			if diff := cmp.Diff(string(gotOutput), string(wantOutput)); diff != "" {
				t.Errorf("Encode64() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func Benchmark_Encode64_noWrap(b *testing.B) {
	var wrapAfter uint
	testInput := "testdata/utf8.decode.input"
	input, err := os.Open(testInput)
	if err != nil {
		b.Fatalf("cannot open %s: %v", testInput, err)
	}
	defer input.Close()
	inputStats, err := input.Stat()
	if err != nil {
		b.Fatalf("cannot get stats for %s: %v", testInput, err)
	}
	inputStats.Size()

	buf := make([]byte, inputStats.Size())
	output := bytes.NewBuffer(buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		if _, err = input.Seek(0, io.SeekStart); err != nil {
			b.Fatalf("cannot seek input file %s: %v", testInput, err)
		}
		b.StartTimer()
		if err = Encode64(input, output, base64.StdEncoding, wrapAfter); err != nil {
			b.Fatalf("Encode64(%s, output, base64.StdEncoding, %d) = %v", input.Name(), wrapAfter, err)
		}
	}
}

func Benchmark_Encode64_wrap76(b *testing.B) {
	var wrapAfter uint = 76
	testInput := "testdata/utf8.decode.input"
	input, err := os.Open(testInput)
	if err != nil {
		b.Fatalf("cannot open %s: %v", testInput, err)
	}
	defer input.Close()
	inputStats, err := input.Stat()
	if err != nil {
		b.Fatalf("cannot get stats for %s: %v", testInput, err)
	}
	inputStats.Size()

	buf := make([]byte, inputStats.Size())
	output := bytes.NewBuffer(buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		if _, err = input.Seek(0, io.SeekStart); err != nil {
			b.Fatalf("cannot seek input file %s: %v", testInput, err)
		}
		b.StartTimer()
		if err = Encode64(input, output, base64.StdEncoding, wrapAfter); err != nil {
			b.Fatalf("Encode64(%s, output, base64.StdEncoding, %d) = %v", input.Name(), wrapAfter, err)
		}
	}
}

func Test_Decode64(t *testing.T) {
	type args struct {
		fileName      string
		encoding      *base64.Encoding
		ignoreGarbage bool
	}
	tests := []struct {
		name         string
		args         args
		wantFileName string
		wantErr      bool
	}{
		{"Standard encoding with padding and no garbage and no wrap", args{"testdata/100c.decode.std.wrap-0.no-garbage.padded.input", base64.StdEncoding, false}, "testdata/100c.decode.std.wrap-0.no-garbage.padded.gold", false},
		{"Standard encoding with padding and no garbage and wrap after 76", args{"testdata/100c.decode.std.wrap-76.no-garbage.padded.input", base64.StdEncoding, false}, "testdata/100c.decode.std.wrap-76.no-garbage.padded.gold", false},
		{"Standard encoding with padding and with garbage and wrap after 76", args{"testdata/100c.decode.std.wrap-76.std-garbage.padded.input", base64.StdEncoding, true}, "testdata/100c.decode.std.wrap-76.std-garbage.padded.gold", false},
		{"Standard encoding with padding and with garbage and wrap after 76 - fail (no ignore)", args{"testdata/100c.decode.std.wrap-76.std-garbage.padded.input", base64.StdEncoding, false}, "testdata/100c.decode.std.wrap-76.std-garbage.padded.fail.gold", true},
		{"Standard encoding with padding and no garbage and wrap after 137", args{"testdata/100c.decode.std.wrap-137.no-garbage.padded.input", base64.StdEncoding, false}, "testdata/100c.decode.std.wrap-137.no-garbage.padded.gold", false},
		{"Standard encoding with padding and no garbage and wrap after 200", args{"testdata/100c.decode.std.wrap-200.no-garbage.padded.input", base64.StdEncoding, false}, "testdata/100c.decode.std.wrap-200.no-garbage.padded.gold", false},
		{"URL encoding with padding and no wrap", args{"testdata/utf8.decode.url.wrap-0.padded.input", base64.URLEncoding, false}, "testdata/utf8.decode.url.wrap-0.padded.golden", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare file input
			file, err := os.Open(tt.args.fileName)
			if err != nil {
				t.Fatalf("cannot open %s: %v", tt.args.fileName, err)
			}
			defer file.Close()

			output := &bytes.Buffer{}

			// Execute
			err = Decode64(file, output, tt.args.encoding, tt.args.ignoreGarbage)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode64() error = %v, wantErr %v", err, tt.wantErr)
			}

			var gotOutput, wantOutput []byte
			gotOutput = output.Bytes()

			if !*update {
				wantOutput, err = ioutil.ReadFile(tt.wantFileName)
				if err != nil {
					t.Fatalf("Cannot read file %s: %v", tt.wantFileName, err)
				}
			} else {
				if err = ioutil.WriteFile(tt.wantFileName, gotOutput, 0644); err != nil {
					t.Fatalf("Cannot write to file %s: %v", tt.wantFileName, err)
				}
				return
			}

			if diff := cmp.Diff(string(gotOutput), string(wantOutput)); diff != "" {
				t.Errorf("Decode64() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func Benchmark_Decode64_noIgnoreGarbage(b *testing.B) {
	ignoreGarbage := true
	testInput := "testdata/utf8.decode.url.wrap-0.padded.input"
	input, err := os.Open(testInput)
	if err != nil {
		b.Fatalf("cannot open %s: %v", testInput, err)
	}
	defer input.Close()
	inputStats, err := input.Stat()
	if err != nil {
		b.Fatalf("cannot get stats for %s: %v", testInput, err)
	}
	inputStats.Size()

	buf := make([]byte, inputStats.Size())
	output := bytes.NewBuffer(buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		if _, err = input.Seek(0, io.SeekStart); err != nil {
			b.Fatalf("cannot seek input file %s: %v", testInput, err)
		}
		b.StartTimer()
		if err = Decode64(input, output, base64.URLEncoding, ignoreGarbage); err != nil {
			b.Fatalf("Decode64(%s, output, base64.URLEncoding, %v) = %v", input.Name(), ignoreGarbage, err)
		}
	}
}

func Benchmark_Decode64_ignoreGarbage(b *testing.B) {
	ignoreGarbage := true
	testInput := "testdata/utf8.decode.url.wrap-0.padded.input"
	input, err := os.Open(testInput)
	if err != nil {
		b.Fatalf("cannot open %s: %v", testInput, err)
	}
	defer input.Close()
	inputStats, err := input.Stat()
	if err != nil {
		b.Fatalf("cannot get stats for %s: %v", testInput, err)
	}
	inputStats.Size()

	buf := make([]byte, inputStats.Size())
	output := bytes.NewBuffer(buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		if _, err = input.Seek(0, io.SeekStart); err != nil {
			b.Fatalf("cannot seek input file %s: %v", testInput, err)
		}
		b.StartTimer()
		if err = Decode64(input, output, base64.URLEncoding, ignoreGarbage); err != nil {
			b.Fatalf("Decode64(%s, output, base64.URLEncoding, %v) = %v", input.Name(), ignoreGarbage, err)
		}
	}
}

func Test_wrapWriter_Write(t *testing.T) {
	type fields struct {
		leftover  int
		wrapAfter int
		w         io.Writer
	}
	type args struct {
		p []byte
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantN      int
		wantOutput []byte
		wantErr    bool
	}{
		{"no wrap for wrapAfter == 0", fields{wrapAfter: 0, w: &bytes.Buffer{}}, args{[]byte("1234567890")}, len("1234567890"), []byte("1234567890"), false},
		{"no wrap for wrapAfter longer than input", fields{wrapAfter: 50, w: &bytes.Buffer{}}, args{[]byte("1234567890")}, len("1234567890"), []byte("1234567890"), false},
		{"wrap for wrapAfter smaller than input with newline at the end - wrap 5 for len 10", fields{wrapAfter: 5, w: &bytes.Buffer{}}, args{[]byte("1234567890")}, len("1234567890"), []byte("12345\n67890\n"), false},
		{"wrap for wrapAfter smaller than input without newline at the end - wrap 7 for len 10", fields{wrapAfter: 7, w: &bytes.Buffer{}}, args{[]byte("1234567890")}, len("1234567890"), []byte("1234567\n890"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			ww := &wrapWriter{
				leftover:  tt.fields.leftover,
				wrapAfter: tt.fields.wrapAfter,
				w:         output,
			}
			gotN, err := ww.Write(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("wrapWriter.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("wrapWriter.Write() = %v, want %v", gotN, tt.wantN)
			}
			gotOutput := output.Bytes()
			if diff := cmp.Diff(string(gotOutput), string(tt.wantOutput)); diff != "" {
				t.Errorf("wrapWriter.Write() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func Test_garboReader_Read(t *testing.T) {
	type fields struct {
		alphabet      alphabet
		ignoreGarbage bool
	}
	tests := []struct {
		name          string
		fields        fields
		input         string
		wantP         string
		wantErrNotEOF bool
	}{
		{"no garbage to drop", fields{alphabet: base64std, ignoreGarbage: false}, "c2ltcGxl", "c2ltcGxl", false},
		{"no garbage to drop with ignore garbage", fields{alphabet: base64std, ignoreGarbage: true}, "c2ltcGxl", "c2ltcGxl", false},
		{"no garbage, drop newlines", fields{alphabet: base64std, ignoreGarbage: true}, "c2lt\ncGxl\n", "c2ltcGxl", false},
		{"no garbage, drop carriage returns with newlines", fields{alphabet: base64std, ignoreGarbage: true}, "c2lt\r\ncGxl\r\n", "c2ltcGxl", false},
		{"drop garbage for standard alphabet", fields{alphabet: base64std, ignoreGarbage: true}, `n_o-+£g$a%r^b&a*g(e)`, "no+garbage", false},
		{"drop garbage for URL alphabet", fields{alphabet: base64url, ignoreGarbage: true}, `n/o-+£g$a%r^b&a*g(e)`, "no-garbage", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gr := &garboReader{
				alphabet:      tt.fields.alphabet,
				ignoreGarbage: tt.fields.ignoreGarbage,
				r:             bytes.NewBufferString(tt.input),
			}
			p := make([]byte, len(tt.input))
			wantN := len(tt.wantP)
			gotN, err := gr.Read(p)
			if (err != nil) != tt.wantErrNotEOF && err != io.EOF {
				t.Errorf("garboReader.Read() error = %v, wantErrNotEOF %v", err, tt.wantErrNotEOF)
				return
			}
			if gotN != wantN {
				t.Errorf("garboReader.Read() = %v, want %v", gotN, wantN)
			}
			gotOutput := p[:gotN]
			if diff := cmp.Diff(string(gotOutput), tt.wantP); diff != "" {
				t.Errorf("garboReader.Read() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
