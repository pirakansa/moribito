package webhook

import (
	"os"
	"path/filepath"
	"runtime"
)

func readFixture(name string) ([]byte, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return nil, os.ErrNotExist
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	path := filepath.Join(root, "test", "fixtures", "webhook", name)
	return os.ReadFile(path)
}
