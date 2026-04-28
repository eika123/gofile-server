package main

import (
	"file-server/internal/auth"
	"file-server/internal/file_traverse"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"errors"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

var (
	ROOT_PATH = "./"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}
	rootPath := os.Getenv("ROOT_PATH")
	if rootPath != "" {
		ROOT_PATH = rootPath
	}
	fmt.Printf("Loaded ROOT_PATH from .env: %s\n", ROOT_PATH)
}

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

	fmt.Printf("received path = %s\n", path)

	// trim ROOT_PATH from the beginning of path if it exists to avoid double join
	if after, ok := strings.CutPrefix(path, ROOT_PATH); ok {
		path = after
		fmt.Printf("trimmed path = %s\n", path)
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
	fmt.Fprintf(w, "<h2>Directories in directory \"%s\":</h2>\n", path)
	for _, dir := range dirs {
		fmt.Fprintf(w, "<li>%s</li>\n", wrapTargetInLink(filepath.Join(path, dir), dir))
	}

	fmt.Fprintf(w, "<h2>Files in directory \"%s\":</h2>\n", path)
	for _, file := range files {
		fmt.Println("file: " + file)
		fmt.Println("path: " + filepath.Join(path, file))
		fmt.Fprintf(w, "<li>%s</li>\n", wrapTargetInLink(filepath.Join(path, file), file))
	}
	endHTML(w)
}

func Router(authMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	if authMiddleware != nil {
		r.Use(authMiddleware)
	}

	r.Get("/hello", hello)
	r.Get("/files", DisplayDirectoryContents)

	return r
}

func envOrDefault(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func main() {
	loadEnv()

	authUser := os.Getenv("BASIC_AUTH_USER")
	authPass := os.Getenv("BASIC_AUTH_PASS")
	if authUser == "" || authPass == "" {
		fmt.Fprintln(os.Stderr, "BASIC_AUTH_USER and BASIC_AUTH_PASS must be set")
		os.Exit(1)
	}

	bindAddr := envOrDefault("BIND_ADDR", "")
	if bindAddr == "" {
		bindAddr = ":" + envOrDefault("PORT", "8080")
	}

	fmt.Printf("Starting server on %s...\n", bindAddr)
	router := Router(auth.NewBasicAuth(authUser, authPass))
	if err := http.ListenAndServe(bindAddr, router); err != nil {
		slog.Error("server stopped", "err", err)
	}
}
