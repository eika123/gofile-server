package main

import (
	"file-server/internal/auth"
	"file-server/internal/file_traverse"
	"file-server/internal/ui"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"errors"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func handleHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

var (
	ROOT_PATH = "./"
)

func loadEnv() {
	if runtimePath := os.Getenv("ROOT_PATH"); runtimePath != "" {
		ROOT_PATH = runtimePath
	}

	fmt.Println("Extracted ROOT_PATH from environment: " + ROOT_PATH)

	envMap, err := godotenv.Read()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Println("Error loading .env file:", err)
	}

	if ROOT_PATH == "" {
		if rootPath, ok := envMap["ROOT_PATH"]; ok && rootPath != "" {
			ROOT_PATH = rootPath
		}
	}

	if ROOT_PATH == "" {
		ROOT_PATH = "."
	}

	ROOT_PATH = filepath.Clean(ROOT_PATH)
	if !filepath.IsAbs(ROOT_PATH) {
		cwd, err := os.Getwd()
		if err == nil {
			ROOT_PATH = filepath.Join(cwd, ROOT_PATH)
		}
	}

	fmt.Printf("Using ROOT_PATH=%s\n", ROOT_PATH)
}

// inTrustedRoot ensures the given path is under the trusted root directory.
// Called by verifyPath during path validation.
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

// verifyPath resolves symlinks and checks that the path remains within ROOT_PATH.
// Called by resolveRequestedPath before file access.
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

// parseRequestedPath extracts the "path" query parameter from the request.
// Called by DisplayDirectoryContents.
func parseRequestedPath(r *http.Request) (string, error) {
	path := r.URL.Query().Get("path")
	if path == "" {
		return "", errors.New("Missing 'path' query parameter")
	}
	return path, nil
}

// resolveRequestedPath normalizes the requested path against ROOT_PATH and verifies it.
// Called by DisplayDirectoryContents after parsing the query.
func resolveRequestedPath(rawPath string) (string, error) {
	rawPath = path.Clean("/" + rawPath)
	if rawPath == "/" {
		rawPath = ""
	}

	rawPath = strings.TrimPrefix(rawPath, "/")
	cleanPath := filepath.Join(ROOT_PATH, rawPath)
	return verifyPath(cleanPath)
}

// getDisplayPath returns the path shown in the UI, with ROOT_PATH stripped.
// Called by DisplayDirectoryContents before rendering the listing.
func getDisplayPath(resolvedPath string) string {
	fmt.Println("Resolved path for display: " + resolvedPath)
	rootAbs, err := filepath.Abs(ROOT_PATH)
	if err != nil {
		return resolvedPath
	}

	resolvedAbs, err := filepath.Abs(resolvedPath)
	if err != nil {
		return resolvedPath
	}

	rel, err := filepath.Rel(rootAbs, resolvedAbs)
	if err != nil {
		return resolvedPath
	}
	if rel == "." {
		return "/"
	}
	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, "/") {
		rel = "/" + rel
	}
	return rel
}

// serveFile writes the contents of a regular file to the response.
// Called by DisplayDirectoryContents when the requested path is a file.
func serveFile(w http.ResponseWriter, path string, info os.FileInfo) error {
	content, err := file_traverse.GetFileContent(path)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", info.Name()))
	_, err = w.Write(content)
	return err
}

// listDirectoryContents returns the files and subdirectories for the given path.
// Called by DisplayDirectoryContents when the requested path is a directory.
func listDirectoryContents(path string) ([]string, []string, error) {
	files, err := file_traverse.ListFiles(path)
	if err != nil {
		return nil, nil, err
	}

	dirs, err := file_traverse.ListSubDirectories(path)
	if err != nil {
		return nil, nil, err
	}

	return files, dirs, nil
}

// Should fetch directory from e.g
// http://localhost:8080/files?path=/path/to/directory
func handleDisplayDirectoryContents(w http.ResponseWriter, r *http.Request) {
	rawPath, err := parseRequestedPath(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	path, err := resolveRequestedPath(rawPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error verifying path: %v", err), http.StatusBadRequest)
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error accessing path: %v", err), http.StatusInternalServerError)
		return
	}

	if info.Mode().IsRegular() {
		// serveFile streams the file content to the response with appropriate headers.
		if err := serveFile(w, path, info); err != nil {
			http.Error(w, fmt.Sprintf("Error reading file: %v", err), http.StatusInternalServerError)
		}
		return
	}

	files, dirs, err := listDirectoryContents(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing directory contents: %v", err), http.StatusInternalServerError)
		return
	}

	displayPath := getDisplayPath(path)
	if err := ui.RenderDirectoryListing(w, displayPath, dirs, files); err != nil {
		http.Error(w, fmt.Sprintf("Error rendering directory contents: %v", err), http.StatusInternalServerError)
	}
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

	r.Get("/hello", handleHello)
	r.Get("/files", handleDisplayDirectoryContents)
	r.Handle("/static/*", http.StripPrefix("/static/", ui.StaticHandler()))

	return r
}

func envOrDefault(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func initLogger() {
	defaultLogDir := "./logs/file-server"
	logDir := envOrDefault("LOG_PATH", defaultLogDir)

	if !filepath.IsAbs(logDir) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot resolve working directory: %v\n", err)
			return
		}
		logDir = filepath.Join(cwd, logDir)
	}
	logDir = filepath.Clean(logDir)

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cannot create log directory %s: %v\n", logDir, err)
		return
	}

	logFile, err := os.OpenFile(filepath.Join(logDir, "app.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cannot open log file: %v\n", err)
		return
	}

	w := io.MultiWriter(os.Stdout, logFile)
	slog.SetDefault(slog.New(slog.NewTextHandler(w, nil)))
	fmt.Fprintf(os.Stderr, "Logging to %s\n", filepath.Join(logDir, "app.log"))
}

func main() {
	loadEnv()
	initLogger()

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
