package ui

import (
	"html/template"
	"io"
	"net/url"
	"path"
)

// DirectoryEntry holds a rendered link entry for a directory or file.
type DirectoryEntry struct {
	Label string
	Href  string
	Type  string
}

// DirectoryListing is the template model for the directory listing page.
type DirectoryListing struct {
	Current     string
	Directories []DirectoryEntry
	Files       []DirectoryEntry
}

var directoryTemplate = template.Must(template.New("directory").Parse(`<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Directory Contents</title>
  <link rel="icon" href="/favicon.ico" type="image/x-icon">
  <link rel="stylesheet" href="/static/css/dirlist.css">
  <script src="/static/js/dirlist.js" defer></script>
</head>
<body>
  <h1>Directory listing for {{.Current}}</h1>
  {{if .Directories}}
    <h2>Directories</h2>
    <ul>
      {{- range .Directories}}
        <li class="directory" data-name="{{.Label}}">
          <a href="{{.Href}}"><span class="icon"></span><span class="label">{{.Label}}</span></a>
        </li>
      {{- end}}
    </ul>
  {{end}}
  {{if .Files}}
    <h2>Files</h2>
    <ul>
      {{- range .Files}}
        <li class="file" data-name="{{.Label}}">
          <a href="{{.Href}}"><span class="icon"></span><span class="label">{{.Label}}</span></a>
        </li>
      {{- end}}
    </ul>
  {{end}}
</body>
</html>`))

// RenderDirectoryListing renders the file listing page using a template.
// Called by DisplayDirectoryContents after directory contents are loaded.
func RenderDirectoryListing(w io.Writer, currentPath string, dirs, files []string) error {
	model := DirectoryListing{
		Current:     currentPath,
		Directories: buildEntries(currentPath, dirs, "directory"),
		Files:       buildEntries(currentPath, files, "file"),
	}
	return directoryTemplate.Execute(w, model)
}

func buildEntries(basePath string, names []string, entryType string) []DirectoryEntry {
	entries := make([]DirectoryEntry, 0, len(names))
	for _, name := range names {
		href := buildSafePathHref(basePath, name)
		if href == "" {
			continue
		}

		entries = append(entries, DirectoryEntry{
			Label: name,
			Href:  href,
			Type:  entryType,
		})
	}
	return entries
}

func buildSafePathHref(basePath, name string) string {
	basePath = path.Clean("/" + basePath)
	if basePath == "." {
		basePath = "/"
	}

	safeName := path.Clean(name)
	if safeName == "." || safeName == ".." || safeName == "" {
		return ""
	}
	if path.IsAbs(safeName) {
		safeName = path.Base(safeName)
	}

	targetPath := path.Join(basePath, safeName)
	return "/files?path=" + url.QueryEscape(targetPath)
}
