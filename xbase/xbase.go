package xbase

import (
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"sync"
)

var (
	stdGarbage64 = regexp.MustCompile(`[^a-zA-Z0-9\+/=\n\r]+`)
	urlGarbage64 = regexp.MustCompile(`[^a-zA-Z0-9\-_=\n\r]+`)
)

// Encode64 read stream from input and encode it to base64 with optional wrapping
func Encode64(input io.Reader, output io.Writer, encoding *base64.Encoding, wrapAfter uint) error {

	wrapper := NewWrapWriter(output, int(wrapAfter))

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
	count     int
	wrapAfter int
	notUsed   bool

	w io.Writer
}

func (ww *wrapWriter) Write(p []byte) (n int, err error) {
	if ww.wrapAfter == 0 {
		ww.notUsed = true
		return ww.w.Write(p)
	}

	// if we have less bytes than the wrapAfter number then just write the bytes and add to the counter
	if len(p)+ww.count < ww.wrapAfter {
		ww.count += len(p)
		return ww.w.Write(p)
	}

	ns := (len(p) + ww.count) / ww.wrapAfter // how many \n's we will need taking into account any partial lines previously written
	b := make([]byte, 0, len(p)+ns)          // allocate how much we will write including the newlines

	var x int
	for i := 0; i < ns; i++ {
		b = append(b, p[x:x+ww.wrapAfter-ww.count]...)
		b = append(b, []byte("\n")...)
		x = x + ww.wrapAfter - ww.count
		ww.count = 0 // reset the counter to 0 so we only use it the first time
	}
	b = append(b, p[x:]...) // write any remaining bytes after the last newline was added
	ww.count = len(p[x:])   // if we have any bytes left over then add to the counter
	n, err = ww.w.Write(b)  // write to the underlining writer (checking for errors)
	return n - ns, err      // return the bytes written (minus the newlines... otherwise the written bytes won't matchup with the upstream writer) and any errors
}

// AddMissingNewline write newline to internal writer if last character is not newline (ww.count != 0)
// and wrapping is used
func (ww *wrapWriter) AddMissingNewline() (err error) {
	if ww.count != 0 && !ww.notUsed {
		_, err = ww.w.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func NewWrapWriter(w io.Writer, at int) *wrapWriter {
	return &wrapWriter{wrapAfter: at, w: w}
}

func wrap(wrapAfter uint, input io.Reader, output io.Writer) (err error) {
	wrapOutput := NewWrapWriter(output, int(wrapAfter))
	if _, err = io.Copy(wrapOutput, input); err != nil {
		return err
	}

	if err := wrapOutput.AddMissingNewline(); err != nil {
		return err
	}

	return nil
}

// Decode64 read stream from input and decode it output with optional garbade ignoring
func Decode64(input io.Reader, output io.Writer, encoding *base64.Encoding, ignoreGarbage bool) error {
	var (
		garbage          *regexp.Regexp
		plainDecodeInput io.Reader
		wg               sync.WaitGroup
	)
	errc := make(chan error, 2) // one per worker goroutine
	pr, pw := io.Pipe()

	switch encoding {
	case base64.StdEncoding, base64.RawStdEncoding:
		garbage = stdGarbage64
	case base64.URLEncoding, base64.RawURLEncoding:
		garbage = urlGarbage64
	default:
		return fmt.Errorf("encoding is not supported")
	}

	plainDecodeInput = input
	if ignoreGarbage {
		plainDecodeInput = pr
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if err := pw.Close(); err != nil {
					errc <- fmt.Errorf("cannot close pipe writer: %v", err)
				}
			}()
			if err := dropGarbage(garbage, input, pw); err != nil {
				errc <- fmt.Errorf("cannot drop garbage: %v", err)
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if pr, ok := plainDecodeInput.(*io.PipeReader); ok {
				if err := pr.Close(); err != nil {
					errc <- fmt.Errorf("cannot close pipe reader: %v", err)
				}
			}
		}()
		if err := plainDecode(plainDecodeInput, output, encoding); err != nil {
			errc <- fmt.Errorf("cannot decode: %v", err)
		}
	}()

	go func() {
		wg.Wait()
		close(errc)
	}()

	for err := range errc {
		return err
	}
	return nil
}

func dropGarbage(garbage *regexp.Regexp, input io.Reader, output io.Writer) (err error) {
	buffer := make([]byte, 32*1024)
	for {
		n, err := input.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("cannot read from input: %v", err)
		}
		if _, err = output.Write(garbage.ReplaceAllLiteral(buffer[:n], []byte{})); err != nil {
			return err
		}
	}
	return nil
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
