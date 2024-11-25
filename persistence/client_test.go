package persistence

import (
	"github.com/malinowskip/pal/constants"
	"github.com/malinowskip/pal/util"
	"path"
	"testing"
)

func TestStartClientCreatesDatabaseFile(t *testing.T) {
	projectPath := t.TempDir()

	_, err := StartClient(projectPath)

	if err != nil {
		t.Error(err)
	}

	if !util.FileExists(path.Join(projectPath, constants.AppDir, "db.sqlite")) {
		t.Error("Starting a database client should create a database")
	}
}
