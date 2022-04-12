// Package cuectx provides a single, central CUE context (runtime) and Thema
// library that can be used uniformly across Grafana, and related helper
// functions for loading Thema lineages.

package cuectx

import (
	"io"
	"io/fs"
	"path/filepath"
	"testing/fstest"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/grafana/thema"
	"github.com/grafana/thema/load"
)

var ctx *cue.Context = cuecontext.New()
var lib thema.Library = thema.NewLibrary(ctx)

// ProvideCUEContext is a wire service provider of a central cue.Context.
func ProvideCUEContext() *cue.Context {
	return ctx
}

// ProvideThemaLibrary is a wire service provider of a central thema.Library.
func ProvideThemaLibrary() thema.Library {
	return lib
}

// LoadGrafanaInstancesWithThema loads CUE files containing a lineage
// representing some Grafana core model schema. It is expected to be used when
// implementing a thema.LineageFactory.
//
// This function primarily juggles paths to make CUE's loader happy. Provide the
// path from the grafana root to the directory containing the lineage.cue. The
// lineage.cue file must be the sole contents of the provided fs.FS.
//
// More details on underlying behavior can be found in the docs for github.com/grafana/thema/load.InstancesWithThema.
func LoadGrafanaInstancesWithThema(path string, cueFS fs.FS, lib thema.Library, opts ...thema.BindOption) (thema.Lineage, error) {
	prefix := filepath.FromSlash(path)
	fs, err := prefixWithGrafanaCUE(prefix, cueFS)
	if err != nil {
		return nil, err
	}
	inst, err := load.InstancesWithThema(fs, prefix)

	// Need to trick loading by creating the embedded file and
	// making it look like a module in the root dir.
	if err != nil {
		return nil, err
	}

	val := lib.Context().BuildInstance(inst)

	lin, err := thema.BindLineage(val, lib, opts...)
	if err != nil {
		return nil, err
	}

	return lin, nil
}

func prefixWithGrafanaCUE(prefix string, inputfs fs.FS) (fs.FS, error) {
	m := fstest.MapFS{
		filepath.Join("cue.mod", "module.cue"): &fstest.MapFile{Data: []byte(`module: "github.com/grafana/grafana"`)},
	}

	prefix = filepath.FromSlash(prefix)
	err := fs.WalkDir(inputfs, ".", (func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		f, err := inputfs.Open(path)
		if err != nil {
			return err
		}
		defer f.Close() // nolint: errcheck

		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		m[filepath.Join(prefix, path)] = &fstest.MapFile{Data: b}
		return nil
	}))

	return m, err
}
