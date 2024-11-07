package documents

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/go-git/go-git/plumbing/format/gitignore"
)

// A Document represents a text file, such as `README.md`, which is included in
// the context sent to the LLM.
type Document struct {
	// Relative path within the project directory.
	Path    string
	Content string
}

// Given the project path, loads documents that will be included in the context
// sent to the LLM.
//
// It accepts additional .gitignore glob patterns for files to be excluded and
// respects any .gitignore files in the project directory.
//
// It excludes files whose size exceeds the maxFileSize argument.
func LoadDocuments(projectPath string, excludePatterns []string, maxFileSize int64) ([]Document, error) {
	var documents []Document

	// Hardcoded patterns that should always be applied.
	patterns := []gitignore.Pattern{
		gitignore.ParsePattern(".git", []string{}),
		gitignore.ParsePattern(".pal", []string{}),
	}

	// Include custom exclude patterns provided by the user.
	for _, pattern := range excludePatterns {
		patterns = append(patterns, gitignore.ParsePattern(pattern, []string{}))
	}

	// The matcher applies .gitignore patterns to filter out files and directories.
	// It will be updated with additional patterns whenever a .gitignore file is
	// found.
	matcher := gitignore.NewMatcher(patterns)

	err := filepath.WalkDir(projectPath, func(path string, dirEntry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate the relative path from the project root to the current file. This
		// is needed to properly handle gitignore .patterns and create correct document
		// paths.
		relPath, err := filepath.Rel(projectPath, path)
		if err != nil {
			return err
		}

		pathElements := strings.Split(relPath, string(os.PathSeparator))

		// For directory entries, search for and process any .gitignore files.
		// If found, parse their patterns and update the gitignore matcher.
		if dirEntry.IsDir() {
			gitignorePath := filepath.Join(path, ".gitignore")

			// At this point, attempt to read a .gitignore file in the current directory.
			// If one exists, parse its patterns; if not, continue with the current patterns.
			if content, err := os.ReadFile(gitignorePath); err == nil {
				// The domain specifies the scope where gitignore patterns should apply. When
				// processing the root .gitignore file (at projectPath), we leave the domain
				// empty so patterns apply project-wide. For .gitignore files in subdirectories,
				// patterns only apply within that subdirectory's tree.
				var domain []string
				if relPath != "." {
					domain = strings.Split(relPath, string(os.PathSeparator))
				}

				for _, line := range strings.Split(string(content), "\n") {
					if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
						pattern := gitignore.ParsePattern(line, domain)
						// Add the pattern to the global list by appending it at the end. Patterns
						// added later have higher precedence and will override any earlier
						// conflicting patterns during matching.
						patterns = append(patterns, pattern)
					}
				}

				// Update the gitignore matcher with all accumulated patterns, allowing
				// newly added patterns from this directory’s .gitignore to take effect.
				matcher = gitignore.NewMatcher(patterns)
			}
		}

		// If the matcher matches a file against its .gitignore patterns, the entry
		// should be excluded.
		if matcher.Match(pathElements, dirEntry.IsDir()) {
			// When a directory matches .gitignore patterns, tell WalkDir to skip the
			// entire directory and its subdirectories.
			if dirEntry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := dirEntry.Info()
		if err != nil {
			return err
		}

		// Exclude files that exceed the file size limit.
		if info.Size() > maxFileSize {
			return nil
		}

		// Exclude non-UTF-8 files like images.
		isTextFile, err := IsValidUTF8File(path)
		if !isTextFile || err != nil {
			return nil
		}

		// Now if the entry is a file, include its contents.
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			documents = append(documents, Document{
				Path:    relPath,
				Content: string(content),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return documents, nil
}

// Based on a sample of the file, checks if the file contains valid UTF-8
// encoding.
func IsValidUTF8File(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}

	if fileInfo.IsDir() {
		return false, nil
	}

	buffer := make([]byte, 8*1024) // 8KB should be enough for a initial check
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}

	return utf8.Valid(buffer[:n]), nil
}
