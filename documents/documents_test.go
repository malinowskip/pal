package documents

import (
	"os"
	"pal/testutil"
	"path"
	"path/filepath"
	"testing"
)

func TestDocumentLoading(t *testing.T) {
	projectPath := t.TempDir()

	// Create a test file in the temporary directory.
	addTestFile := func(filePath, fileContent string) {
		abs := path.Join(projectPath, filePath)

		if err := os.MkdirAll(filepath.Dir(abs), os.ModePerm); err != nil {
			panic(err)
		}

		file, err := os.Create(abs)

		if err != nil {
			panic(err)
		}

		_, err = file.WriteString(fileContent)

		if err != nil {
			panic(err)
		}
	}

	toInclude := map[string]string{
		"README.md":              "Hello, world!",
		"subdir/README.md":       "",
		"src/main.go":            "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		"subdir/.aimportant.txt": "should be included", // New test case
	}

	toExclude := map[string]string{
		"subdir/backup.sql":              "",
		"subdir/node_modules/leftPad.js": "",
		"subdir1/access.log":             "",
		"subdir1/nested/access.log":      "",
		"subdir2/*.xml":                  "",
		".git/hello.md":                  "",
		"build/compiled/bin/output.exe":  "",
		".afile.txt":                     "should be excluded", // New test case
		"subdir/other.txt":               "should be excluded", // New test case
	}

	toInclude[".gitignore"] = "*.txt\n*.sql\nnode_modules\nbuild" // Modified to include *.txt
	toInclude["subdir/.gitignore"] = "!.aimportant.txt\n"
	toExclude["image.png"] = "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x10\x00\x00\x00\x10\x08\x06\x00\x00\x00\x1f\xf3\xffa\x00\x00\x00\x01sRGB\x00\xae\xce\x1c\xe9\x00\x00\x00\x04gAMA\x00\x00\xb1\x8f\x0b\xfca\x05\x00\x00\x00\tpHYs\x00\x00\x0e\xc3\x00\x00\x0e\xc3\x01\xc7o\xa8d\x00\x00\x00\x0cIDAT8Oc\xf8\xff\xff?\x03\x00\x05\xfe\x02\xfe\xdc\xccY\xe7\x00\x00\x00\x00IEND\xaeB`\x82"
	toExclude["adir/.gitignore"] = "*"
	toInclude["subdir1/.gitignore"] = "*.log"

	additionalExcludePatterns := []string{"*.xml"}

	for p, content := range toInclude {
		addTestFile(p, content)
	}

	for p, content := range toExclude {
		addTestFile(p, content)
	}

	docs, _ := LoadDocuments(projectPath, additionalExcludePatterns, 10000)

	for _, doc := range docs {
		content, exists := toInclude[doc.Path]
		if !exists {
			t.Errorf("LoadDocuments included a path that was not expected: %s", doc.Path)
		}

		if doc.Content != content {
			t.Errorf("LoadDocuments included content that was not expected: %s", doc.Content)
		}
	}

outerloop:
	for expectedPath := range toInclude {
		for _, item := range docs {
			if item.Path == expectedPath {
				break outerloop
			}
		}
		t.Errorf("The collection does not contain the specified item: %v", expectedPath)
	}

}

func TestDocumentLoadingRespectsFileSizeRestriction(t *testing.T) {
	projectPath := t.TempDir()
	testFileName := "README.md"
	testFilePath := path.Join(projectPath, testFileName)
	f, _ := os.Create(testFilePath)
	f.WriteString("Hello, world!")

	t.Run("Loads file normally", func(t *testing.T) {
		maxFileSize := int64(10_000)
		d, _ := LoadDocuments(projectPath, []string{}, maxFileSize)
		testutil.AssertContains(t, d, func(doc Document) bool {
			return doc.Path == testFileName
		})
	})

	t.Run("Does not load the file if it’s too large", func(t *testing.T) {
		maxFileSize := int64(1)
		d, _ := LoadDocuments(projectPath, []string{}, maxFileSize)
		testutil.AssertNotContains(t, d, func(doc Document) bool {
			return doc.Path == testFileName
		})
	})
}
