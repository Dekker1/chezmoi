package chezmoi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

var _ System = &DumpSystem{}

func TestDumpSystem(t *testing.T) {
	chezmoitest.WithTestFS(t, map[string]interface{}{
		"/home/user/.local/share/chezmoi": map[string]interface{}{
			".chezmoiignore":  "README.md\n",
			".chezmoiremove":  "*.txt\n",
			".chezmoiversion": "1.2.3\n",
			".chezmoitemplates": map[string]interface{}{
				"template": "# contents of .chezmoitemplates/template\n",
			},
			"README.md": "",
			"dot_dir": map[string]interface{}{
				"file": "# contents of .dir/file\n",
			},
			"run_script":      "# contents of script\n",
			"symlink_symlink": ".dir/subdir/file\n",
		},
	}, func(fileSystem vfs.FS) {
		ctx := context.Background()
		system := NewRealSystem(fileSystem)
		s := NewSourceState(
			WithBaseSystem(system),
			WithSourceDir("/home/user/.local/share/chezmoi"),
			WithSystem(system),
		)
		require.NoError(t, s.Read(ctx, nil))
		requireEvaluateAll(t, s, system)

		dumpSystem := NewDumpSystem()
		persistentState := NewMockPersistentState()
		require.NoError(t, s.applyAll(dumpSystem, system, persistentState, "", ApplyOptions{
			Include: NewEntryTypeSet(EntryTypesAll),
		}))
		expectedData := map[AbsPath]interface{}{
			".dir": &dirData{
				Type: dataTypeDir,
				Name: ".dir",
				Perm: 0o777 &^ chezmoitest.Umask,
			},
			".dir/file": &fileData{
				Type:     dataTypeFile,
				Name:     ".dir/file",
				Contents: "# contents of .dir/file\n",
				Perm:     0o666 &^ chezmoitest.Umask,
			},
			"script": &scriptData{
				Type:     dataTypeScript,
				Name:     "script",
				Contents: "# contents of script\n",
			},
			"symlink": &symlinkData{
				Type:     dataTypeSymlink,
				Name:     "symlink",
				Linkname: ".dir/subdir/file",
			},
		}
		actualData := dumpSystem.Data()
		assert.Equal(t, expectedData, actualData)
	})
}
