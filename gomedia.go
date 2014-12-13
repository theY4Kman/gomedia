package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var INITIALIZED bool = false
var MEDIA_EXTS map[string]bool
var ROOT_PATH string

func make_string_set(strings ...string) map[string]bool {
	set := make(map[string]bool)
	for _, s := range strings {
		set[s] = true
	}
	return set
}

func _print_dir_html(buf io.Writer, dir string, indent string) (int, int) {
	num_files, num_videos := 0, 0

	fmt.Fprintln(buf, indent, "<ul>")

	files, _ := ioutil.ReadDir(dir)
	for _, file := range files {
		num_files++

		fmt.Fprintln(buf, indent, "    <li>")

		path := filepath.Join(dir, file.Name())
		class := ""

		if file.IsDir() {
			var child_buf bytes.Buffer
			child_writer := bufio.NewWriter(&child_buf)

			class += " folder"

			child_files, child_videos := _print_dir_html(child_writer, path, indent+"        ")
			fmt.Fprintf(buf, "<span class=\"%s\" data-path=\"%s\">%s <span class=\"num-files\">(%d files, %d videos)</span></span>", class, path, file.Name(), child_files, child_videos)

			// Flush the buffered writer before writing to our output buf
			child_writer.Flush()
			child_buf.WriteTo(buf)

			num_files += child_files
			num_videos += child_videos
		} else {
			class += " file"

			// Determine class names based on file type (well, extension)
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if len(ext) > 0 {
				ext = ext[1:len(ext)]
			}

			if _, exists := MEDIA_EXTS[ext]; exists {
				class += " type-video"
				num_videos++
			}

			fmt.Fprintf(buf, "<span class=\"%s\" data-path=\"%s\">%s</span>", class, path, file.Name())
		}

		fmt.Fprintln(buf, indent, "    </li>")
	}

	fmt.Fprintln(buf, indent, "</ul>")
	return num_files, num_videos
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
		body {
			font-family: sans-serif;
		}

		ul {
			list-style: none;
		}

		ul ul {
			padding-top: 5px;
		}

		li {
			padding: 5px 0;
		}

		.info {
			background-color: lightgray;
			border: 1px solid gray;
			margin-left: 15px;
			padding: 5px;
		}

		.type-video {
			color: green;
			cursor: pointer;
		}

		.folder {
			color: gray;
			cursor: pointer;
		}

		.info {
			font-family: monospace;
		}

		.header {
			font-weight: bold;
			font-size: large;
			padding-top: 20px;
		}
		tr:first-child .header {
			padding-top: 0;
		}

		.title {
			padding-right: 10px;
			position: relative;
		}
		.title::after {
			content: ":";
			position: absolute;
			right: 0px;
		}
	</style>
	<script type="text/javascript" src="//code.jquery.com/jquery-2.1.1.min.js"></script>
	<script type="text/javascript">
		$(function() {
			$('ul ul').hide();
			$('.folder').click(function() {
				var $children = $(this).parent().children('ul');
				$children.slideToggle(300);
			});

			$('.type-video').click(function() {
				var $info = $(this).siblings('.info');
				if ($info.length > 0) {
					$info.slideToggle(300);
					return;
				}

				$info = $('<div>', {class: 'info'}).text('Loading...');
				$info.appendTo($(this).parent())

				$.post('/info', {path: $(this).data('path')}, function(info) {
					$info.empty();
					var $table = $('<table>').appendTo($info);
					$.each(info.split(/\r?\n/), function(_, line) {
						line = line.trim();
						if (line.length == 0) return;

						var $tr = $('<tr>').appendTo($table);
						if (line.indexOf(':') !== -1) {
							var value_line = line.split(/\s*:\s*/),
							    title = value_line[0],
							    value = value_line[1];
									$tr.append($('<td>', {class: 'title'}).text(title));
									$tr.append($('<td>', {class: 'value'}).text(value));
						} else {
							$tr.append($('<td>', {class: 'header', colspan: 2}).text(line));
						}
					});
				});
			});
		});
	</script>
</head>
<body>`)
	_print_dir_html(w, ROOT_PATH, "    ")
	fmt.Fprint(w, "</body>\n</html>")
}

func media_info(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		w.Write([]byte("405 Method Not Allowed"))
		return
	}

	path := r.PostFormValue("path")
	cmd := exec.Command("mediainfo", path)
	info, _ := cmd.CombinedOutput()

	w.Write(info)
}

func init() {
	MEDIA_EXTS = make_string_set("mp4", "avi", "mkv")

	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "<path to directory>")
		os.Exit(2)
	}
	ROOT_PATH = os.Args[1]

	INITIALIZED = true
}
func main() {
	http.Handle("/", http.HandlerFunc(list_files))
	http.Handle("/info", http.HandlerFunc(media_info))

	log.Fatal(http.ListenAndServe(":7000", nil))
}
