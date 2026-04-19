package file_traverse

import "os"

func ListFiles(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var regularFiles []string
	for _, file := range files {
		if !file.IsDir() {
			regularFiles = append(regularFiles, file.Name())
		}
	}
	return regularFiles, nil
}

func ListSubDirectories(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var subdirs []string
	for _, file := range files {
		if file.IsDir() {
			subdirs = append(subdirs, file.Name())
		}
	}
	return subdirs, nil
}

func GetFileContent(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return content, nil
}
