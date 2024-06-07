package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"bufio"
	"bytes"
)

/* START MY OWN BRAILLE MAPPING THAT CAN DO BINARY BYTES */
// 1) start with braille ascii for 0x20 to 0x5F
// 2) flip dot 7 for all letters in 0x40 to 0x5F
// 3) copy 0x20..0x3F to 0x00, and set dot 7 on 0x00..0x1F
// 4) copy 0x40..0x5F to 0x60, and set dot 7 on 0x60..0x7F
// 5) flip 7 dot for all letters in 0x60..0x7F
// 6) copy all of 0x00..0x7F to 0x80, and set dot 8 on 0x80..0xFF
//
// Note that in this table, you can read the braille directly:
// using dot number:
//
//	0b01011001
//	  87654321  .. the braille dots... imagine re-laid out:
//
//	   14
//	   25
//	   36
//	   78
//
// by looking at it sideways.  after some practice, you can read dots from binary.
var brailleAsciiPattern = []int{
	0b00000000, 0b00101110, 0b00010000, 0b00111100, 0b00101011, 0b00101001, 0b00101111, 0b00000100, 0b00110111, 0b00111110, 0b00100001, 0b00101100, 0b00100000, 0b00100100, 0b00101000, 0b00001100,
	0b00110100, 0b00000010, 0b00000110, 0b00010010, 0b00110010, 0b00100010, 0b00010110, 0b00110110, 0b00100110, 0b00010100, 0b00110001, 0b00110000, 0b00100011, 0b00111111, 0b00011100, 0b00111001,
	0b00001000, 0b00000001, 0b00000011, 0b00001001, 0b00011001, 0b00010001, 0b00001011, 0b00011011, 0b00010011, 0b00001010, 0b00011010, 0b00000101, 0b00000111, 0b00001101, 0b00011101, 0b00010101,
	0b00001111, 0b00011111, 0b00010111, 0b00001110, 0b00011110, 0b00100101, 0b00100111, 0b00111010, 0b00101101, 0b00111101, 0b00110101, 0b00101010, 0b00110011, 0b00111011, 0b00011000, 0b00111000,
}

// When looking at 0x20 through 0x5F as 6-dot, mask off dot 7 first,
// as there are only upper-case letters in braille ascii
var braillePerm = make([]int, 256)
var asciiPerm = make([]int, 256)
var present = make([]int, 256)

func brailleInit() {
	// Copy in the standard braille ascii patern
	for i := 0; i < 64; i++ {
		braillePerm[0x20+i] = brailleAsciiPattern[i]
	}
	// Flip the case of the alphabet
	//for i := 0x41; i <= 0x5A; i++ {
	for i := 0x40; i <= 0x60; i++ {
		braillePerm[i] = braillePerm[i] ^ 0x40
	}
	// Copy lower half of standard to cover control codes
	for i := 0; i < 32; i++ {
		braillePerm[i] = (braillePerm[i+0x20]) ^ 0x40
	}
	// Copy upper half of standard to cover upper case
	for i := 0; i < 32; i++ {
		braillePerm[0x60+i] = braillePerm[0x40+i] ^ 0x40
	}
	// Swap 124 and 127 the underscore and delete,
	// a strange exception logically, but I see it in real terminals
	braillePermTmp := braillePerm[0x7F]
	braillePerm[0x7F] = braillePerm[0x5F]
	braillePerm[0x5F] = braillePermTmp

	// Duplicated it all in high bits
	for i := 0; i < 128; i++ {
		braillePerm[0x80+i] = braillePerm[i] ^ 0x80
	}
	// Reverse mapping
	for i := 0; i < 256; i++ {
		asciiPerm[braillePerm[i]] = i
		present[i]++
	}
	// Panic if codes are missing or duplicated
	for i := 0; i < 256; i++ {
		if present[i] != 1 {
			panic(fmt.Sprintf("inconsistency at %d", i))
		}
	}

}

func asciiToComputerBRL(s string) string {
	var b strings.Builder
	for _,n := range []byte(s) {
		// make a brl rune unless we literally translate it
		if n == 0x09 {
			b.WriteString("        ")
			continue
		}
		c := braillePerm[n] + 0x2800
		b.WriteString(fmt.Sprintf("%c", c))
	}
	return b.String()
}

func computerBRLToASCII(s string) string {
	var b bytes.Buffer
	for _,c := range s {
		if 0x2800 <= c && c <= 0x28FF { 
			c = rune(asciiPerm[ int(c) - 0x2800 ])
		}
		b.WriteString(fmt.Sprintf("%c", c))
	}
	return b.String()
}


/* END MY OWN BRAILLE MAPPING THAT CAN DO BINARY BYTES */













 
func translateLL(input string, direction string) (string, error) {
	cmd := exec.Command(
		"lou_translate",
		direction,
		"unicode.dis,en-ueb-g2.ctb",
	)

	// Provide the input to the command
	cmd.Stdin = strings.NewReader(input)

	// Capture the output
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func downloadURL(url string) (string, error) {
	// Send GET request
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch URL: %s, status code: %d", url, resp.StatusCode)
	}

	// Read the response body
	// yup, this can be BIG
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func SplitLines(s string) []string {
    var lines []string
    sc := bufio.NewScanner(strings.NewReader(s))
    for sc.Scan() {
        lines = append(lines, sc.Text())
    }
    return lines
}

/*
GET /ueb2?u=https://www.gutenberg.org/cache/epub/8001/pg8001.txt
GET /ueb2?s=Hello+World!
*/
func toBRL(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	original := q.Get("original") == "true"
	computer := q.Get("computer") == "true"
	var s string
	if len(q.Get("s")) > 0 {
		s = q.Get("s")
	}
	if len(q.Get("u")) > 0 {
		u, err := downloadURL(q.Get("u"))
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("%v", err)))
			return
		}
		s = u
	}
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	w.Write([]byte("<html>\n"))
	w.Write([]byte("  <head>\n"))
	w.Write([]byte("    <link rel='stylesheet' href='static/mystyle.css'>\n"))
	w.Write([]byte("  </head>\n"))
	w.Write([]byte("  <body>\n"))
	w.Write([]byte("<pre>"))
	lines := SplitLines(s)
	for _,line := range lines {
		var translation string
		var err error
		if computer {
			translation = asciiToComputerBRL(line)
		} else {
			translation, err = translateLL(line, "--forward")
		}
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("%v", err)))
			return
		}
		if original {
			fmt.Fprintf(w, "%s\n", line)
		}
		fmt.Fprintf(w, "%s\n", translation)
	}
	w.Write([]byte("    </pre>"))
	w.Write([]byte("  </body>\n"))
	w.Write([]byte("</html>\n"))
}

func fromBRL(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	q := req.URL.Query()
	original := q.Get("original") == "true"
	computer := q.Get("computer") == "true"
	var s string
	if len(q.Get("s")) > 0 {
		s = q.Get("s")
	}
	if len(q.Get("u")) > 0 {
		u, err := downloadURL(q.Get("u"))
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("%v", err)))
			return
		}
		s = u
	}
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	w.Write([]byte("<html>\n"))
	w.Write([]byte("  <head>\n"))
	w.Write([]byte("    <link rel='stylesheet' href='static/mystyle.css'>\n"))
	w.Write([]byte("  </head>\n"))
	w.Write([]byte("  <body>\n"))
	w.Write([]byte("    <pre>"))
	lines := SplitLines(s)
	for _,line := range lines {
		var translation string
		var err error
		if computer {
			translation = computerBRLToASCII(line)
		} else {
			translation, err = translateLL(line, "--backward")
		}
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("%v", err)))
			return
		}
		w.Header().Set("Content-type", "text/html; charset=utf-8")
		if original {
			fmt.Fprintf(w, "%s\n", line)
		}
		fmt.Fprintf(w, "%s\n", translation)
	}
	w.Write([]byte("    </pre>"))
	w.Write([]byte("  </body>\n"))
	w.Write([]byte("</html>\n"))
}

// examples at the root page
func defaultPage(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	w.Write([]byte("<html>\n"))
	w.Write([]byte("  <head>\n"))
	w.Write([]byte("    <link rel='stylesheet' href='static/mystyle.css'>\n"))
	w.Write([]byte("  </head>\n"))
	w.Write([]byte("  <body>\n"))
	w.Write([]byte("    <h2>Examples</h2>\n"))
	w.Write([]byte("    <ul>\n"))
	w.Write([]byte("      <li><a href='/tobrl?u=https://www.gutenberg.org/cache/epub/8001/pg8001.txt'>Genesis</a></li>\n"))
	w.Write([]byte("      <li><a href='/tobrl?original=true&u=https://www.gutenberg.org/cache/epub/8001/pg8001.txt'>Genesis, with original text</a></li>\n"))
	w.Write([]byte("      <li><a href='/tobrl?original=true&computer=true&u=http://localhost:8090/static/usl.py'>Python, with computer braille of original text</a></li>\n"))
	w.Write([]byte("      <li><a href='/tobrl?original=true&computer=false&u=http://localhost:8090/static/usl.py'>Python, with UEB braille of original text</a></li>\n"))
	w.Write([]byte("      <li><a href='/tobrl?s=Hello,+World!'>Hello, World!</a></li>\n"))
	w.Write([]byte("      <li><a href='/frombrl?s=⠠⠓⠑⠇⠇⠕⠂+⠠⠸⠺⠖'>⠠⠓⠑⠇⠇⠕⠂⠀⠠⠸⠺⠖</a></li>\n"))
	w.Write([]byte("    </ul>\n"))
	w.Write([]byte("  </body>\n"))
	w.Write([]byte("</html>\n"))
}

func main() {
	brailleInit()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/tobrl", toBRL)
	http.HandleFunc("/frombrl", fromBRL)
	http.HandleFunc("/", defaultPage)
	http.ListenAndServe(":8090", nil)
}
