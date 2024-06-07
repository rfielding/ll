package main

import (
	"fmt"
	"net/http"
	"exec"
)

func translateLL(input string, direction string) (string, error) {
	cmd := exec.Command(
		direction,
		"lou_translate",
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

/*
*

	GET /ueb2?s=Hello+World!
*/
func toUEB2(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	translation, err := translateLL(q.Get("s"), "--forward")
	if err != nil {
		w.WriteHeader(500)
		w.Write(fmt.Sprintf("%v", err))
		return
	}
	fmt.Fprintf(w, translation)
}

func main() {
	http.HandleFunc("/ueb2", toUEB2)
	http.ListenAndServe(":8090", nil)
}
