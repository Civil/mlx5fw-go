//go:build pcie_enabled
// +build pcie_enabled

package pcie

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var bdfRe = regexp.MustCompile(`^[0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-7]$`)

func isBDF(s string) bool { return bdfRe.MatchString(s) }
func IsBDF(s string) bool { return isBDF(s) }

// findMSTForBDF tries to locate /dev/mst pciconf device for a BDF.
// This is best-effort: it scans /dev/mst and returns the first readable entry.
func findMSTForBDF(bdf string) (string, bool) {
	entries, err := os.ReadDir("/dev/mst")
	if err != nil {
		return "", false
	}
	// Normalize to lowercase
	bdf = strings.ToLower(bdf)
	// Prefer exact-named entries like 0000:07:00.0_pciconf0
	exact := bdf + "_pciconf0"
	for _, e := range entries {
		name := strings.ToLower(e.Name())
		if name == exact {
			path := filepath.Join("/dev/mst", e.Name())
			if _, err := os.Stat(path); err == nil {
				return path, true
			}
		}
	}
	// Fallback: any entry starting with BDF and containing pciconf
	for _, e := range entries {
		name := strings.ToLower(e.Name())
		if strings.HasPrefix(name, bdf) && strings.Contains(name, "pciconf") {
			path := filepath.Join("/dev/mst", e.Name())
			if _, err := os.Stat(path); err == nil {
				return path, true
			}
		}
	}
	// Symlink-based mapping: resolve links and look for BDF in target
	for _, e := range entries {
		name := e.Name()
		full := filepath.Join("/dev/mst", name)
		fi, err := os.Lstat(full)
		if err != nil {
			continue
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			if target, err := os.Readlink(full); err == nil {
				tgt := strings.ToLower(target)
				if strings.Contains(tgt, bdf) && strings.Contains(tgt, "pciconf") {
					if _, err := os.Stat(full); err == nil {
						return full, true
					}
				}
			}
		}
	}
	// As a last resort, open each /dev/mst entry and query MST_PARAMS to match the BDF
	for _, e := range entries {
		name := e.Name()
		path := filepath.Join("/dev/mst", name)
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if p, err := mstParamsOnPath(path); err == nil {
			got := fmt.Sprintf("%04x:%02x:%02x.%d", p.Domain, p.Bus, p.Slot, p.Func)
			if strings.ToLower(got) == bdf {
				return path, true
			}
		}
	}
	return "", false
}

// ListMSTNodes returns all /dev/mst entries containing "pciconf".
func ListMSTNodes() ([]string, error) {
	entries, err := os.ReadDir("/dev/mst")
	if err != nil {
		return nil, err
	}
	out := []string{}
	for _, e := range entries {
		out = append(out, filepath.Join("/dev/mst", e.Name()))
	}
	return out, nil
}
