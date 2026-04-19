package main

import (
	"file-server/internal/file_traverse"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"net/http"

	"errors"

	"log/slog"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

var (
	ROOT_PATH = "./"
)

func inTrustedRoot(path string, trustedRoot string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	absRoot, err := filepath.Abs(trustedRoot)
	if err != nil {
		return err
	}

	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return err
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		slog.Warn(fmt.Sprintf("path attempted: rel=%s, abs=%s,   absPath=%s", rel, absRoot, absPath))
		return errors.New("path is outside of trusted root")
	}
	return nil
}

func verifyPath(path string) (string, error) {

	// Read from config in the real world
	trustedRoot := ROOT_PATH

	c := filepath.Clean(path)
	fmt.Println("veryPath: Cleaned path: " + c)

	r, err := filepath.EvalSymlinks(c)
	if err != nil {
		fmt.Println("Error " + err.Error())
		return c, errors.New("Unsafe or invalid path specified")
	}

	err = inTrustedRoot(r, trustedRoot)
	if err != nil {
		fmt.Println("Error " + err.Error())
		return r, errors.New("Unsafe or invalid path specified")
	} else {
		fmt.Println("veryPath: Path is within trusted root: " + r)
		return r, nil
	}
}

func wrapTargetInLink(path, dirname string) string {
	return fmt.Sprintf("<a href=\"/files?path=%s\">%s</a>", path, dirname)
}

func startHTML(w http.ResponseWriter, title string) {
	fmt.Fprintf(w, "<html><head><title>%s</title></head><body>", title)
}

func endHTML(w http.ResponseWriter) {
	fmt.Fprintf(w, "</body></html>")
}

// Should fetch directory from e.g
// http://localhost:8080/files?path=/path/to/directory
func DisplayDirectoryContents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	path := query.Get("path")
	if path == "" {
		http.Error(w, "Missing 'path' query parameter", http.StatusBadRequest)
		return
	}

	path = filepath.Join(ROOT_PATH, path)
	path, err := verifyPath(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error verifying path: %v", err), http.StatusBadRequest)
		return
	}
	fmt.Println(" ---- ---- Verified path: " + path)

	// check if path is a directory, a regular file or neither
	info, err := os.Stat(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error accessing path: %v", err), http.StatusInternalServerError)
		return
	}
	// if path is a file, serve the file content
	if info.Mode().IsRegular() {
		content, err := file_traverse.GetFileContent(path)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading file: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", info.Name()))
		w.Write(content)
		return
	}

	files, err := file_traverse.ListFiles(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing files: %v", err), http.StatusInternalServerError)
		return
	}

	dirs, err := file_traverse.ListSubDirectories(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing directories: %v", err), http.StatusInternalServerError)
		return
	}

	startHTML(w, "Directory Contents")
	fmt.Fprintf(w, "<h2>Directories in directory %s:</h2>\n", path)
	for _, dir := range dirs {
		fmt.Fprintf(w, "<li>%s</li>\n", wrapTargetInLink(filepath.Join(path, dir), dir))
	}

	fmt.Fprintf(w, "<h2>Files in directory %s:</h2>\n", path)
	for _, file := range files {
		fmt.Fprintf(w, "<li>%s</li>\n", wrapTargetInLink(filepath.Join(path, file), file))
	}
	endHTML(w)
}

func Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", hello)
	mux.HandleFunc("/files", DisplayDirectoryContents)
	return mux
}

func main() {

	kek := filepath.Join("hey", "kek")

	fmt.Println(kek)

	current_working_dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	files, err := file_traverse.ListFiles(current_working_dir)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	dirs, err := file_traverse.ListSubDirectories(current_working_dir)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Directories:")
	for _, dir := range dirs {
		fmt.Println(dir)
	}
	fmt.Println("\nFiles:")
	for _, file := range files {
		fmt.Println(file)
	}

	port := "8080"
	fmt.Printf("Starting server on port %s...\n", port)
	http.ListenAndServe(":"+port, Router())
}
