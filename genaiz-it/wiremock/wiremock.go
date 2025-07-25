package wiremock

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var (
	//go:embed all:wiremock_res
	embeddedRes  embed.FS
	embeddedPath = "wiremock_res"
)

type ProvisionStubbing interface {
	Build()
}

type provisionStubbing struct {
	wiremockHost string
	token        string
	realm        string
}

func CopyFiles(outputPath string) error {
	var filesPath = filepath.Join(embeddedPath, "__files")

	return copyDir(outputPath, filesPath)
}

func CopyMappings(outputPath string, filters []string) error {
	var mappingsPath = filepath.Join(embeddedPath, "mappings")
	var faviconPath = filepath.Join(outputPath, "favicon_ico.json")
	var err error

	// Required for the healthcheck and avoiding silly errors in the wiremock logs caused by browsers attempting to retrieve it.
	if err = copyFile(faviconPath, filepath.Join(mappingsPath, "favicon_ico.json")); err == nil {
		var filteredJson []fs.DirEntry

		if filteredJson, err = filterMappings(mappingsPath, filters...); err == nil {
			for _, filtered := range filteredJson {
				var path = filepath.Join(mappingsPath, filtered.Name())
				var output = filepath.Join(outputPath, filtered.Name())

				if err = copyFile(output, path); err != nil {
					break
				}
			}
		}
	}

	return err
}

func copyDir(outputPath string, dirPath string) error {
	var entries []fs.DirEntry
	var err error

	if entries, err = embeddedRes.ReadDir(dirPath); entries != nil {
		for _, entry := range entries {
			var path = filepath.Join(dirPath, entry.Name())
			var output = filepath.Join(outputPath, entry.Name())

			if entry.IsDir() {
				if err = os.MkdirAll(output, 0750); err == nil {
					if err = copyDir(outputPath, path); err != nil {
						break
					}
				}
			} else if err = copyFile(output, path); err != nil {
				break
			}
		}
	}

	return err
}

func copyFile(outputFile string, embeddedFile string) error {
	var bytes []byte
	var err error

	if bytes, err = embeddedRes.ReadFile(embeddedFile); err == nil {
		err = os.WriteFile(outputFile, bytes, 0600)
	}

	return err
}

func filterMappings(mappingsPath string, filters ...string) ([]fs.DirEntry, error) {
	var entries []fs.DirEntry
	var result []fs.DirEntry
	var err error

	if entries, err = embeddedRes.ReadDir(mappingsPath); entries != nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				if slices.ContainsFunc(filters, func(s string) bool {
					return strings.HasPrefix(entry.Name(), s)
				}) {
					result = append(result, entry)
				}
			}
		}
	}

	return result, err
}
