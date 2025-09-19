package main

// Proxy for opening local files in the explorer/finder via link from the browser
// run as:
// go run .\proxy.go -base C:\Users\Sophie\SomeFolder -token foo -port 1234
// and put a link onto a website:
// http://localhost:1234/open?name=subDir&token=foo
// when the link is clicked (fetched by the browser) the subDir is open in the explorer locally

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var basePath string
var port string
var token string

func HasFileWithPrefixExceptExt(dir, prefix, excludeExt string) (bool, error) {
	pattern := filepath.Join(dir, prefix)
	files, err := filepath.Glob(pattern)
	if err != nil {
		return false, err
	}

	for _, f := range files {
		if !strings.HasSuffix(f, excludeExt) {
			return true, nil
		}
	}
	return false, nil
}

func openPath(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		//		cmd = exec.Command("explorer", "/c", "start", "", path)
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
}

func main() {
	flag.StringVar(&basePath, "base", ".", "Base directory for allowed paths (required)")
	flag.StringVar(&port, "port", "4455", "Port to listen on")
	flag.StringVar(&token, "token", "", "Secret token to check for (&token=...) in request")
	flag.Parse()

	abs, err := filepath.Abs(basePath)
	if err != nil {
		fmt.Println("Error resolving base path:", err)
		os.Exit(1)
	}
	basePath = abs
	info, err := os.Stat(basePath)
	if os.IsNotExist(err) {
		fmt.Println("Error: base path does not exist:", basePath)
		os.Exit(1)
	}
	if err != nil {
		fmt.Println("Error checking base path:", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Println("Error: base path is not a directory:", basePath)
		os.Exit(1)
	}

	http.HandleFunc("/open", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		givenToken := r.URL.Query().Get("token")
		if name == "" {
			http.Error(w, "Missing ?name= parameter", http.StatusBadRequest)
			return
		}
		if token != "" && token != givenToken {
			http.Error(w, "Invalid Token", http.StatusBadRequest)
			return
		}

		cleanName := filepath.Clean(name)
		if filepath.IsAbs(name) {
			fmt.Printf("Base path: %s\n", cleanName)
			fmt.Printf("Base path: %s\n", name)
			http.Error(w, "Invalid folder/file name", http.StatusBadRequest)
			return
		}

		fullPath := filepath.Join(basePath, cleanName)
		info, err := os.Stat(fullPath)

		if err != nil || !info.IsDir() {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if err := openPath(fullPath); err != nil {
			http.Error(w, fmt.Sprintf("Failed to open: %v", err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		glob := r.URL.Query().Get("glob")
		givenToken := r.URL.Query().Get("token")
		if name == "" {
			http.Error(w, "Missing ?name= parameter", http.StatusBadRequest)
			return
		}
		if token != "" && token != givenToken {
			http.Error(w, "Invalid Token", http.StatusBadRequest)
			return
		}

		cleanName := filepath.Clean(name)
		if filepath.IsAbs(name) {
			fmt.Printf("Base path: %s\n", cleanName)
			fmt.Printf("Base path: %s\n", name)
			http.Error(w, "Invalid folder/file name", http.StatusBadRequest)
			return
		}
		svgTpl := "<svg viewBox='0 0 16 16' xmlns='http://www.w3.org/2000/svg' font-family='monospace'><rect width='16' height='16' fill='%s' /><text x='1' y='12'>%s</text></svg>"
		fullPath := filepath.Join(basePath, cleanName)
		_, err := os.Stat(fullPath)
		if err != nil {
			w.Header().Add("Content-Type", "image/svg+xml")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, svgTpl, "red", glob)
			return
		} else {
			if glob != "" {
				globBase := filepath.Base(glob)

				ok, err := HasFileWithPrefixExceptExt(fullPath, globBase, ".txt")
				if err != nil {
					http.Error(w, "Bad Glob", http.StatusBadRequest)
					return
				}
				if ok {
					w.Header().Add("Content-Type", "image/svg+xml")
					w.WriteHeader(http.StatusOK)
					fmt.Fprintf(w, svgTpl, "green", glob)
					return
				} else {
					w.Header().Add("Content-Type", "image/svg+xml")
					w.WriteHeader(http.StatusOK)
					fmt.Fprintf(w, svgTpl, "orange", glob)
					return
				}

			} else {

				w.Header().Add("Content-Type", "image/svg+xml")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("<svg viewBox='0 0 16 16' xmlns='http://www.w3.org/2000/svg'><rect width='16' height='16' fill='green' /></svg>"))
				return
			}
		}
	})
	http.HandleFunc("/style", func(w http.ResponseWriter, r *http.Request) {
		class := r.URL.Query().Get("class")
		givenToken := r.URL.Query().Get("token")
		if class == "" {
			http.Error(w, "Missing ?class= parameter", http.StatusBadRequest)
			return
		}
		if token != "" && token != givenToken {
			http.Error(w, "Invalid Token", http.StatusBadRequest)
			return
		}
		w.Header().Add("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		cssTpl := ".%s { display: initial !important; }"

		fmt.Fprintf(w, cssTpl, class)
	})

	fmt.Printf("Base path: %s\n", basePath)
	fmt.Printf("Listening on http://localhost:%s\n", port)
	fmt.Printf("Example:\n http://localhost:%s/open?name=.&token=%s\n", port, token)
	if err := http.ListenAndServe("localhost:"+port, nil); err != nil {
		fmt.Println("Server error:", err)
		os.Exit(1)
	}
}
