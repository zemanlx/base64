package xbase

import (
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"sync"
)

var (
	stdGarbage64 *regexp.Regexp = regexp.MustCompile(`[^a-zA-Z0-9\+/=\n\r]+`)
	urlGarbage64 *regexp.Regexp = regexp.MustCompile(`[^a-zA-Z0-9\-_=\n\r]+`)
)

// Encode64 read stream from input and encode it to base64 with optional wrapping
func Encode64(input io.Reader, output io.Writer, encoding *base64.Encoding, wrapAfter uint) error {
	var wg sync.WaitGroup
	errc := make(chan error, 2) // one per worker goroutine
	pr, pw := io.Pipe()

	wg.Add(1)
	go func() { // encode
		defer wg.Done()
		defer func() {
			if err := pw.Close(); err != nil {
				errc <- fmt.Errorf("cannot close pipe writer: %v", err)
			}
		}()
		if err := plainEncode(input, pw, encoding); err != nil {
			errc <- fmt.Errorf("cannot encode: %v\n", err)
		}
	}()

	wg.Add(1)
	go func() { // wrap
		defer wg.Done()
		defer func() {
			if err := pr.Close(); err != nil {
				errc <- fmt.Errorf("cannot close pipe reader: %v", err)
			}
		}()
		if err := wrap(wrapAfter, pr, output); err != nil {
			errc <- fmt.Errorf("cannot wrap: %v\n", err)
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

func wrap(wrapAfter uint, input io.Reader, output io.Writer) (err error) {
	if wrapAfter == 0 {
		if _, err = io.Copy(output, input); err != nil {
			return err
		}
		return nil
	}
	wrapSymbol := []byte("\n")
	var written int64
	for {
		written, err = io.CopyN(output, input, int64(wrapAfter))
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if _, err = output.Write(wrapSymbol); err != nil {
			return err
		}
	}
	// To be backward compatible with linux base64
	// add one newline after wrapping if there isn't newline from for loop
	if written > 0 && written < int64(wrapAfter) {
		if _, err = output.Write(wrapSymbol); err != nil {
			return err
		}
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
