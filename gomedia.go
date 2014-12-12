package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"os"
)

var INITIALIZED bool = false
var MEDIA_EXTS map[string]bool

func make_string_set(strings ...string) map[string]bool {
	set := make(map[string]bool)
	for _, s := range strings {
		set[s] = true
	}
	return set
}

func _print_dir_html(buf io.Writer, dir string, indent string) {
	fmt.Fprintln(buf, indent, "<ul>")

	files, _ := ioutil.ReadDir(dir)
	for _, file := range files {
		fmt.Fprintln(buf, indent, "    <li>")

		// Determine class names based on file type (well, extension)
		class := "filename"
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if len(ext) > 0 {
			ext = ext[1:len(ext)]
		}

		if _, exists := MEDIA_EXTS[ext]; exists {
			class += " type-video"
		}

		fmt.Fprintln(buf, "<span class=\"", class, "\">", file.Name(), "</span>")

		if file.IsDir() {
			_print_dir_html(buf, filepath.Join(dir, file.Name()), indent+"        ")
		}

		fmt.Fprintln(buf, indent, "    </li>")
	}

	fmt.Fprintln(buf, indent, "</ul>")
}

func print_dir_html(buf io.Writer, dir string) {
	_print_dir_html(buf, dir, "")
}

func list_files(w http.ResponseWriter, r *http.Request) {
	if !INITIALIZED {
		fmt.Fprintln(w, "not initialized")
		return
	}

	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintln(w, `
<html>
<head>
	<title>GoMedia: media directory listing</title>
	<style type="text/css">
		ul {
			list-style: none;
		}

		.type-video {
			color: blue;
		}
	</style>
</head>
<body>`)
	_print_dir_html(w, os.Args[1], "    ")
	fmt.Fprint(w, "</body>\n</html>")
}

func init() {
	MEDIA_EXTS = make_string_set("mp4", "avi", "mkv")
	INITIALIZED = true
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "<path to directory>")
		os.Exit(2);
	}

	http.Handle("/", http.HandlerFunc(list_files))
	log.Fatal(http.ListenAndServe(":7000", nil))
}
