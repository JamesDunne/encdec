package main

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
)

type Algorithm struct {
	Encode func(dst io.Writer, src io.Reader) error
	Decode func(dst io.Writer, src io.Reader) error
}

var algorithms map[string]Algorithm = map[string]Algorithm{
	"base64": Algorithm{
		Encode: func(dst io.Writer, src io.Reader) error {
			encoder := base64.NewEncoder(base64.URLEncoding, dst)
			defer encoder.Close()
			_, err := io.Copy(encoder, src)
			return err
		},
		Decode: func(dst io.Writer, src io.Reader) error {
			decoder := base64.NewDecoder(base64.URLEncoding, src)
			_, err := io.Copy(dst, decoder)
			return err
		},
	},
	"base32": Algorithm{
		Encode: func(dst io.Writer, src io.Reader) error {
			encoder := base32.NewEncoder(base32.StdEncoding, dst)
			defer encoder.Close()
			_, err := io.Copy(encoder, src)
			return err
		},
		Decode: func(dst io.Writer, src io.Reader) error {
			decoder := base32.NewDecoder(base32.StdEncoding, src)
			_, err := io.Copy(dst, decoder)
			return err
		},
	},
	"hex": Algorithm{
		Encode: func(dst io.Writer, src io.Reader) error {
			b, err := ioutil.ReadAll(src)
			if err != nil {
				return err
			}
			o := hex.EncodeToString(b)
			_, err = io.Copy(dst, strings.NewReader(o))
			return err
		},
		Decode: func(dst io.Writer, src io.Reader) error {
			b, err := ioutil.ReadAll(src)
			if err != nil {
				return err
			}
			o, err := hex.DecodeString(string(b))
			if err != nil {
				return err
			}
			_, err = io.Copy(dst, bytes.NewReader(o))
			return err
		},
	},
	"uri": Algorithm{
		Encode: func(dst io.Writer, src io.Reader) error {
			b, err := ioutil.ReadAll(src)
			if err != nil {
				return err
			}
			o := url.QueryEscape(string(b))
			_, err = io.Copy(dst, strings.NewReader(o))
			return err
		},
		Decode: func(dst io.Writer, src io.Reader) error {
			b, err := ioutil.ReadAll(src)
			if err != nil {
				return err
			}
			o, err := url.QueryUnescape(string(b))
			if err != nil {
				return err
			}
			_, err = io.Copy(dst, strings.NewReader(o))
			return err
		},
	},
	"html": Algorithm{
		Encode: func(dst io.Writer, src io.Reader) error {
			b, err := ioutil.ReadAll(src)
			if err != nil {
				return err
			}
			o := html.EscapeString(string(b))
			_, err = io.Copy(dst, strings.NewReader(o))
			return err
		},
		Decode: func(dst io.Writer, src io.Reader) error {
			b, err := ioutil.ReadAll(src)
			if err != nil {
				return err
			}
			o := html.UnescapeString(string(b))
			_, err = io.Copy(dst, strings.NewReader(o))
			return err
		},
	},
}

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "encdec <-e | -d> <algorithm> <data | ->\n")
		fmt.Fprintf(os.Stderr, "\nAlgorithms:\n")
		for name, _ := range algorithms {
			fmt.Fprintf(os.Stderr, "  %s\n", name)
		}
		return
	}

	// Parse args:
	encoding := true
	if args[0] == "-e" {
		encoding = true
	} else if args[0] == "-d" {
		encoding = false
	} else {
		fmt.Fprintln(os.Stderr, "-e or -d expected as first argument")
		return
	}

	// Get algorithm
	algorithm_name := strings.ToLower(args[1])
	algorithm, ok := algorithms[algorithm_name]
	if !ok {
		fmt.Fprintln(os.Stderr, "Unknown algorithm name")
		fmt.Fprintf(os.Stderr, "\nAlgorithms:\n")
		for name, _ := range algorithms {
			fmt.Fprintf(os.Stderr, "  %s\n", name)
		}
		return
	}

	// Data taken from stdin or args:
	var src io.Reader

	if len(args) == 2 {
		src = bytes.NewReader([]byte{})
	} else {
		if args[2] != "-" {
			data := strings.Join(args[2:], " ")
			src = bytes.NewReader([]byte(data))
		} else {
			src = os.Stdin
		}
	}

	// Encode or decode data:
	if encoding {
		algorithm.Encode(os.Stdout, src)
	} else {
		algorithm.Decode(os.Stdout, src)
	}
}
