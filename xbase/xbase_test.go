package xbase

import (
	"bytes"
	"encoding/base64"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"regexp"
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

func Test_wrap(t *testing.T) {
	type args struct {
		wrapAfter uint
		input     io.Reader
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantErr    bool
	}{
		{"no wrapping", args{0, strings.NewReader("1234567890")}, "1234567890", false},
		{"wrap after 5th character", args{5, strings.NewReader("1234567890")}, "12345\n67890\n", false},
		{"wrap after 7th character", args{7, strings.NewReader("1234567890")}, "1234567\n890\n", false},
		{"wrap after end of file", args{20, strings.NewReader("1234567890")}, "1234567890\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := wrap(tt.args.wrapAfter, tt.args.input, output); (err != nil) != tt.wantErr {
				t.Errorf("wrap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOutput := output.String(); gotOutput != tt.wantOutput {
				t.Errorf("wrap() = %v, want %v", gotOutput, tt.wantOutput)
			}
		})
	}
}

func Benchmark_wrap(b *testing.B) {
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

	for i := 0; i < b.N; i++ {
		if _, err = input.Seek(0, io.SeekStart); err != nil {
			b.Fatalf("cannot seek input file %s: %v", testInput, err)
		}
		if err = wrap(wrapAfter, input, output); err != nil {
			b.Fatalf("wrap(%d, %s, output) = %v", wrapAfter, input.Name(), err)
		}
	}
}

func Test_dropGarbage(t *testing.T) {
	type args struct {
		input   io.Reader
		garbage *regexp.Regexp
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantErr    bool
	}{
		{"no garbage to drop", args{strings.NewReader("c2ltcGxl"), stdGarbage64}, "c2ltcGxl", false},
		{"no garbage, keep newlines", args{strings.NewReader("c2lt\ncGxl\n"), stdGarbage64}, "c2lt\ncGxl\n", false},
		{"no garbage, keep carriage returns with newlines", args{strings.NewReader("c2lt\r\ncGxl\r\n"), stdGarbage64}, "c2lt\r\ncGxl\r\n", false},
		{"drop garbage for standard alphabet", args{strings.NewReader(`n_o-+£g$a%r^b&a*g(e)`), stdGarbage64}, "no+garbage", false},
		{"drop garbage for URL alphabet", args{strings.NewReader(`n/o-+£g$a%r^b&a*g(e)`), urlGarbage64}, "no-garbage", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := dropGarbage(tt.args.garbage, tt.args.input, output); (err != nil) != tt.wantErr {
				t.Errorf("dropGarbage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOutput := output.String(); gotOutput != tt.wantOutput {
				t.Errorf("dropGarbage() = %v, want %v", gotOutput, tt.wantOutput)
			}
		})
	}
}

func Benchmark_dropGarbage(b *testing.B) {
	testInput := "testdata/utf8.decode.input"
	input, err := os.Open(testInput)
	if err != nil {
		b.Fatalf("cannot open %s: %v", testInput, err)
	}
	defer input.Close()
	output, err := ioutil.TempFile("", "output")
	if err != nil {
		b.Fatalf("cannot create tempfile %s: %v", testInput, err)
	}
	defer os.Remove(output.Name())

	for i := 0; i < b.N; i++ {
		if _, err = input.Seek(0, io.SeekStart); err != nil {
			b.Fatalf("cannot seek input file %s: %v", testInput, err)
		}
		if err = dropGarbage(stdGarbage64, input, output); err != nil {
			b.Fatalf("dropGarbage(%s, %s) = %v", input.Name(), output.Name(), err)
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

func Benchmark_Encode64(b *testing.B) {
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

func Test_wrapWriter_Write(t *testing.T) {
	type fields struct {
		count     int
		wrapAfter int
		notUsed   bool
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
				count:     tt.fields.count,
				wrapAfter: tt.fields.wrapAfter,
				notUsed:   tt.fields.notUsed,
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
