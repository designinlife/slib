package fs_test

import (
	"testing"

	"github.com/designinlife/slib/fs"
	"github.com/stretchr/testify/assert"
)

func TestZipFromDir(t *testing.T) {
	err := fs.ZipFromDir("D:\\tmp\\t2-dir.zip", "D:\\tmp\\t2", "")
	assert.NoError(t, err)
}

func TestZipFromFiles(t *testing.T) {
	err := fs.ZipFromFiles("D:\\tmp\\t2-files.zip", "D:\\tmp\\t2\\go.mod", "D:\\tmp\\t2\\vendor\\modules.txt")
	assert.NoError(t, err)
}

func TestUnzip(t *testing.T) {
	err := fs.Unzip("D:\\tmp\\t2-dir.zip", "D:\\tmp\\t2-dir")
	assert.NoError(t, err)
}
