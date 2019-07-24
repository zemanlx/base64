package xbase

import (
	"encoding/base64"
	"fmt"
	"io"
)

// Encode64 read stream from input and encode it to base64 with optional wrapping
func Encode64(input io.Reader, output io.Writer, encoding *base64.Encoding, wrapAfter uint) error {

	wrapper := &wrapWriter{wrapAfter: int(wrapAfter), w: output}

	if err := plainEncode(input, wrapper, encoding); err != nil {
		return fmt.Errorf("cannot encode: %v", err)
	}

	// To be backward compatible with linux base64
	// add one newline after wrapping if there isn't newline
	if err := wrapper.AddMissingNewline(); err != nil {
		return fmt.Errorf("cannot add missing newline: %v", err)
	}
	return nil
}

func plainEncode(input io.Reader, output io.Writer, encoding *base64.Encoding) (err error) {
	buffer := make([]byte, 32*1024)
	encoder := base64.NewEncoder(encoding, output)
	defer func() {
		if derr := encoder.Close(); derr != nil {
			err = fmt.Errorf("cannot close encoder: %v, %v", derr, err)
		}
		fmt.Fprint(output)
	}()

	for {
		n, err := input.Read(buffer)
		if err != nil {
			if err == io.EOF && n == 0 {
				break
			}
			return fmt.Errorf("cannot read from input: %v", err)
		}
		if _, err = encoder.Write(buffer[:n]); err != nil {
			return fmt.Errorf("encoder cannot write to buffer: %v", err)
		}
	}
	return nil
}

type wrapWriter struct {
	leftover  int
	wrapAfter int

	w io.Writer
}

func (ww *wrapWriter) Write(p []byte) (n int, err error) {
	if ww.wrapAfter == 0 {
		return ww.w.Write(p)
	}

	// if we have less bytes than the wrapAfter number then just write the bytes and incease leftover
	if len(p)+ww.leftover < ww.wrapAfter {
		ww.leftover += len(p)
		return ww.w.Write(p)
	}

	ns := (len(p) + ww.leftover) / ww.wrapAfter // how many \n's will be needed
	b := make([]byte, 0, len(p)+ns)             // how much will be written including the newlines

	var x int
	for i := 0; i < ns; i++ {
		b = append(b, p[x:x+ww.wrapAfter-ww.leftover]...)
		b = append(b, []byte("\n")...)
		x = x + ww.wrapAfter - ww.leftover
		ww.leftover = 0
	}
	if len(p[x:]) > 0 {
		b = append(b, p[x:]...) // write any remaining bytes after the last newline was added
		ww.leftover = len(p[x:])
	}

	n, err = ww.w.Write(b)
	n -= ns // the bytes written minus the newlines to match len(p) if everying was OK
	return n, err
}

// AddMissingNewline write newline to internal writer
func (ww *wrapWriter) AddMissingNewline() (err error) {
	if ww.leftover != 0 && ww.wrapAfter != 0 {
		_, err = ww.w.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

// Decode64 read stream from input and decode it output with optional garbade ignoring
func Decode64(input io.Reader, output io.Writer, encoding *base64.Encoding, ignoreGarbage bool) error {
	var (
		alphabet alphabet
	)

	switch encoding {
	case base64.StdEncoding, base64.RawStdEncoding:
		alphabet = base64std
	case base64.URLEncoding, base64.RawURLEncoding:
		alphabet = base64url
	default:
		return fmt.Errorf("encoding is not supported")
	}

	sweeper := &garboReader{alphabet: alphabet, ignoreGarbage: ignoreGarbage, r: input}

	if err := plainDecode(sweeper, output, encoding); err != nil {
		return fmt.Errorf("cannot decode: %v", err)
	}

	return nil
}

type garboReader struct {
	alphabet      alphabet
	ignoreGarbage bool
	n             int

	r io.Reader
}

func (gr *garboReader) Read(p []byte) (n int, err error) {
	n, err = gr.r.Read(p)
	if err != nil && n == 0 {
		return n, err
	}
	if !gr.ignoreGarbage {
		return n, err
	}

	gr.n = 0
	for _, char := range p[:n] {
		if gr.alphabet[char] {
			p[gr.n] = char
			gr.n++
		}
	}
	copy(p, p[:gr.n])
	return gr.n, err // may contain io.EOF
}

func plainDecode(input io.Reader, output io.Writer, encoding *base64.Encoding) (err error) {
	buffer := make([]byte, 32*1024)
	decoder := base64.NewDecoder(encoding, input)
	for {
		n, err := decoder.Read(buffer)
		if err != nil {
			if err == io.EOF && n == 0 {
				break
			}
			return fmt.Errorf("decoder cannot read from buffer: %v", err)
		}
		if _, err = output.Write(buffer[:n]); err != nil {
			return fmt.Errorf("cannot write to output: %v", err)
		}
	}

	return nil
}
