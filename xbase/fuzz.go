// +build gofuzz

package xbase

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

func Fuzz(data []byte) int {
	// DropGarbage tests
	outDropGarbageStd := &bytes.Buffer{}
	if err := dropGarbage(stdGarbage64, strings.NewReader(fmt.Sprintf("%q", data)), outDropGarbageStd); err != nil {
		panic(err)
	}
	outDropGarbageURL := &bytes.Buffer{}
	if err := dropGarbage(urlGarbage64, strings.NewReader(fmt.Sprintf("%q", data)), outDropGarbageURL); err != nil {
		panic(err)
	}

	// Standard encoding, padded, 0 wrap
	outEncStd0 := &bytes.Buffer{}
	if err := Encode64(bytes.NewReader(data), outEncStd0, base64.StdEncoding, 0); err != nil {
		panic(err)
	}
	outDecStd := &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncStd0.Bytes()), outDecStd, base64.StdEncoding, false); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecStd.Bytes()) {
		panic("data != outDecStd.Bytes()")
	}
	outDecStdIgn := &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncStd0.Bytes()), outDecStdIgn, base64.StdEncoding, true); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecStdIgn.Bytes()) {
		panic("data != outDecStd.Bytes()")
	}

	// Standard encoding, padded, 76 wrap
	outEncStd76 := &bytes.Buffer{}
	if err := Encode64(bytes.NewReader(data), outEncStd76, base64.StdEncoding, 76); err != nil {
		panic(err)
	}
	outDecStd = &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncStd76.Bytes()), outDecStd, base64.StdEncoding, false); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecStd.Bytes()) {
		panic("data != outDecStd.Bytes()")
	}
	outDecStdIgn = &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncStd76.Bytes()), outDecStdIgn, base64.StdEncoding, true); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecStdIgn.Bytes()) {
		panic("data != outDecStd.Bytes()")
	}

	// Standard encoding, no padding, 0 wrap
	outEncRawStd0 := &bytes.Buffer{}
	if err := Encode64(bytes.NewReader(data), outEncRawStd0, base64.RawStdEncoding, 0); err != nil {
		panic(err)
	}
	outDecStd = &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncRawStd0.Bytes()), outDecStd, base64.RawStdEncoding, false); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecStd.Bytes()) {
		panic("data != outDecStd.Bytes()")
	}
	outDecStdIgn = &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncRawStd0.Bytes()), outDecStdIgn, base64.RawStdEncoding, true); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecStdIgn.Bytes()) {
		panic("data != outDecStd.Bytes()")
	}

	// Standard encoding, no padding, 76 wrap
	outEncRawStd76 := &bytes.Buffer{}
	if err := Encode64(bytes.NewReader(data), outEncRawStd76, base64.RawStdEncoding, 76); err != nil {
		panic(err)
	}
	outDecStd = &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncRawStd76.Bytes()), outDecStd, base64.RawStdEncoding, false); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecStd.Bytes()) {
		panic("data != outDecStd.Bytes()")
	}
	outDecStdIgn = &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncRawStd76.Bytes()), outDecStdIgn, base64.RawStdEncoding, true); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecStdIgn.Bytes()) {
		panic("data != outDecStd.Bytes()")
	}

	// URL encoding, padded, 0 wrap
	outEncURL0 := &bytes.Buffer{}
	if err := Encode64(bytes.NewReader(data), outEncURL0, base64.URLEncoding, 0); err != nil {
		panic(err)
	}
	outDecURL := &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncURL0.Bytes()), outDecURL, base64.URLEncoding, false); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecURL.Bytes()) {
		panic("data != outDecURL.Bytes()")
	}
	outDecURLIgn := &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncURL0.Bytes()), outDecURLIgn, base64.URLEncoding, true); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecURLIgn.Bytes()) {
		panic("data != outDecURL.Bytes()")
	}

	// URL encoding, padded, 76 wrap
	outEncURL76 := &bytes.Buffer{}
	if err := Encode64(bytes.NewReader(data), outEncURL76, base64.URLEncoding, 76); err != nil {
		panic(err)
	}
	outDecURL = &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncURL76.Bytes()), outDecURL, base64.URLEncoding, false); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecURL.Bytes()) {
		panic("data != outDecURL.Bytes()")
	}
	outDecURLIgn = &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncURL76.Bytes()), outDecURLIgn, base64.URLEncoding, true); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecURLIgn.Bytes()) {
		panic("data != outDecURL.Bytes()")
	}

	// URL encoding, no padding, 0 wrap
	outEncRawURL0 := &bytes.Buffer{}
	if err := Encode64(bytes.NewReader(data), outEncRawURL0, base64.RawURLEncoding, 0); err != nil {
		panic(err)
	}
	outDecURL = &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncRawURL0.Bytes()), outDecURL, base64.RawURLEncoding, false); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecURL.Bytes()) {
		panic("data != outDecURL.Bytes()")
	}
	outDecURLIgn = &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncRawURL0.Bytes()), outDecURLIgn, base64.RawURLEncoding, true); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecURLIgn.Bytes()) {
		panic("data != outDecURL.Bytes()")
	}

	// URL encoding, no padding, 76 wrap
	outEncRawURL76 := &bytes.Buffer{}
	if err := Encode64(bytes.NewReader(data), outEncRawURL76, base64.RawURLEncoding, 76); err != nil {
		panic(err)
	}
	outDecURL = &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncRawURL76.Bytes()), outDecURL, base64.RawURLEncoding, false); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecURL.Bytes()) {
		panic("data != outDecURL.Bytes()")
	}
	outDecURLIgn = &bytes.Buffer{}
	if err := Decode64(bytes.NewReader(outEncRawURL76.Bytes()), outDecURLIgn, base64.RawURLEncoding, true); err != nil {
		panic(err)
	}
	if !bytes.Equal(data, outDecURLIgn.Bytes()) {
		panic("data != outDecURL.Bytes()")
	}

	return 1
}
