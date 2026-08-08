package app

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/alirezaudev/ttype/internal/storage"
	"github.com/alirezaudev/ttype/internal/text/langcache"
)

func RunDownloadLanguages(out io.Writer, in io.Reader, yes bool, workers int) error {
	dirs, err := storage.DefaultDirs()
	if err != nil {
		return fmt.Errorf("resolve data dir: %w", err)
	}
	cache := langcache.New(dirs.Data)

	fmt.Fprintln(out, "Fetching language catalog...")
	plan, err := cache.PlanDownloadAll()
	if err != nil {
		return err
	}

	if plan.TotalRemote == 0 {
		return fmt.Errorf("no languages found in remote catalog")
	}
	if len(plan.Entries) == 0 {
		fmt.Fprintf(out, "All %d languages are already cached in %s\n", plan.TotalRemote, cache.Dir())
		return nil
	}

	if workers <= 0 {
		workers = 8
	}

	fmt.Fprintf(out, "Languages available: %d\n", plan.TotalRemote)
	fmt.Fprintf(out, "Already cached:      %d\n", plan.AlreadyCached)
	fmt.Fprintf(out, "To download:         %d\n", len(plan.Entries))
	fmt.Fprintf(out, "Estimated size:      ~%s\n", langcache.FormatSize(plan.TotalBytes))
	fmt.Fprintf(out, "Cache directory:     %s\n", cache.Dir())
	fmt.Fprintf(out, "Parallel downloads:  %d\n", workers)
	fmt.Fprintln(out)

	if !yes {
		confirmed, err := confirm(out, in, "Download all languages? [y/N] ")
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintln(out, "Cancelled.")
			return nil
		}
	}

	fmt.Fprintln(out, "Downloading...")
	var progressMu sync.Mutex
	failures := cache.DownloadAll(plan, workers, func(done, total int, id string, err error) {
		progressMu.Lock()
		defer progressMu.Unlock()

		if err != nil {
			fmt.Fprintf(out, "[%d/%d] %s ... failed: %v\n", done, total, id, err)
			return
		}
		fmt.Fprintf(out, "[%d/%d] %s\n", done, total, id)
	})

	if len(failures) > 0 {
		return fmt.Errorf("downloaded %d/%d languages (%d failed)", len(plan.Entries)-len(failures), len(plan.Entries), len(failures))
	}

	fmt.Fprintf(out, "Done. Cached %d languages in %s\n", plan.TotalRemote, cache.Dir())
	return nil
}

func confirm(out io.Writer, in io.Reader, prompt string) (bool, error) {
	fmt.Fprint(out, prompt)

	line, err := readConfirmationLine(in)
	if err != nil {
		return false, err
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func readConfirmationLine(in io.Reader) (string, error) {
	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		return "", scanner.Err()
	}
	return scanner.Text(), nil
}
