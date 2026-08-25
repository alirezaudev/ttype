package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alirezaudev/ttype/internal/storage"
)

type ClearTarget string

const (
	ClearHistory   ClearTarget = "history"
	ClearLanguages ClearTarget = "languages"
	ClearAll       ClearTarget = "all"
)

func ParseClearTarget(s string) (ClearTarget, error) {
	switch ClearTarget(s) {
	case ClearHistory, ClearLanguages, ClearAll:
		return ClearTarget(s), nil
	default:
		return "", fmt.Errorf("unknown target %q (history, languages or all)", s)
	}
}

func (t ClearTarget) description() string {
	switch t {
	case ClearHistory:
		return "your test history, personal bests and replays"
	case ClearLanguages:
		return "every downloaded language list"
	default:
		return "your test history, personal bests, replays and downloaded languages"
	}
}

func RunClear(store storage.Store, target ClearTarget, yes bool, out io.Writer, in io.Reader) error {
	if !yes {
		confirmed, err := confirm(out, in, fmt.Sprintf("Delete %s? [y/N] ", target.description()))
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintln(out, "Nothing was deleted.")
			return nil
		}
	}

	dirs := store.Paths()
	var removed []string

	if target == ClearHistory || target == ClearAll {
		paths := []string{
			filepath.Join(dirs.Data, "history.json"),
			filepath.Join(dirs.Data, "bests.json"),
			filepath.Join(dirs.Data, "replays"),
		}
		for _, path := range paths {
			gone, err := removePath(path)
			if err != nil {
				return err
			}
			if gone {
				removed = append(removed, path)
			}
		}
	}

	if target == ClearLanguages || target == ClearAll {
		path := filepath.Join(dirs.Data, "languages")
		gone, err := removePath(path)
		if err != nil {
			return err
		}
		if gone {
			removed = append(removed, path)
		}
	}

	if len(removed) == 0 {
		fmt.Fprintln(out, "Nothing to delete.")
		return nil
	}
	for _, path := range removed {
		fmt.Fprintf(out, "  removed %s\n", path)
	}
	return nil
}

func removePath(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if err := os.RemoveAll(path); err != nil {
		return false, fmt.Errorf("remove %s: %w", path, err)
	}
	return true, nil
}
