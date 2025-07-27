package fs_test

import (
	"testing"

	"github.com/designinlife/slib/fs"
	"github.com/stretchr/testify/assert"
)

func TestTarFromDir(t *testing.T) {
	err := fs.TarFromDir("D:\\tmp\\t2-dir.tar", "D:\\tmp\\t2", "t2-tar")
	assert.NoError(t, err)

	err = fs.TarFromDir("D:\\tmp\\t2-dir.tar.gz", "D:\\tmp\\t2", "t2-tar-gz")
	assert.NoError(t, err)
}

func TestTarFromFiles(t *testing.T) {
	err := fs.TarFromFiles("D:\\tmp\\t2-files.tar", "D:\\tmp\\t2\\go.mod", "D:\\tmp\\t2\\vendor\\modules.txt")
	assert.NoError(t, err)

	err = fs.TarFromFiles("D:\\tmp\\t2-files.tar.gz", "D:\\tmp\\t2\\go.mod", "D:\\tmp\\t2\\vendor\\modules.txt")
	assert.NoError(t, err)
}

func TestUntar(t *testing.T) {
	err := fs.Untar("D:\\tmp\\t2-dir.tar", "D:\\tmp\\t2-dir-tar")
	assert.NoError(t, err)

	err = fs.Untar("D:\\tmp\\t2-dir.tar.gz", "D:\\tmp\\t2-dir-tar-gz")
	assert.NoError(t, err)
}
