package config

import (
	"os"
	"path/filepath"
	"strings"
)

// devCheckout describes a local LingTai source tree that the managed Python
// runtime should be installed from in editable mode instead of from PyPI.
//
// On a dev machine the `lingtai` PyPI package's source lives in the
// `lingtai-kernel` repo (its pyproject declares `name = "lingtai"`). The
// sibling `lingtai` repo is the Go TUI/portal workspace and is NOT a Python
// package, so it is never an editable install target — it only serves as a
// marker that this is a real LingTai dev workspace.
type devCheckout struct {
	// KernelSrc is the lingtai-kernel directory (contains pyproject.toml).
	KernelSrc string
	// TUISrc is the sibling lingtai/ workspace dir. Recorded for diagnostics;
	// it is not a pip install target.
	TUISrc string
}

// installTargets returns the directories to pass to `pip install -e`. Only the
// kernel is a Python package, so this is always a single-element slice.
func (d devCheckout) installTargets() []string {
	return []string{d.KernelSrc}
}

// isEditableForKernel reports whether the given PEP 610 direct_url source
// (e.g. "file:///Users/dev/work/GitHub/lingtai-kernel") points at this dev
// checkout's kernel. Used to decide whether an existing editable install
// already matches the local checkout and therefore needs no reinstall.
func (d devCheckout) isEditableForKernel(source string) bool {
	if source == "" || d.KernelSrc == "" {
		return false
	}
	got := normalizeDirURL(source)
	want := normalizeDirURL(d.KernelSrc)
	return got != "" && got == want
}

// normalizeDirURL strips a file:// scheme and trailing slash so editable
// source URLs and plain paths compare equal.
func normalizeDirURL(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "file://")
	s = strings.TrimRight(s, "/")
	if s == "" {
		return ""
	}
	return filepath.Clean(s)
}

// findDevCheckouts looks for an explicitly configured local LingTai dev
// workspace and returns it if valid. lookupEnv defaults to os.LookupEnv when
// nil; tests inject a stub.
//
// $LINGTAI_DEV_ROOT is the dev-mode contract: it must point at a directory that
// directly contains sibling lingtai-kernel/ and lingtai/ checkouts. We do not
// auto-scan common GitHub folders here, because ordinary users can have source
// clones without intending the managed runtime to replace PyPI with editable
// local code.
//
// A checkout is valid only when lingtai-kernel/pyproject.toml exists AND a
// sibling lingtai/ directory exists; this avoids treating a bare clone or an
// unrelated directory as a dev workspace.
func findDevCheckouts(home string, lookupEnv func(string) (string, bool)) (devCheckout, bool) {
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}

	if override, ok := lookupEnv("LINGTAI_DEV_ROOT"); ok && strings.TrimSpace(override) != "" {
		return devCheckoutAt(strings.TrimSpace(override))
	}

	return devCheckout{}, false
}

// devCheckoutAt validates a single candidate root directory.
func devCheckoutAt(root string) (devCheckout, bool) {
	kernel := filepath.Join(root, "lingtai-kernel")
	tui := filepath.Join(root, "lingtai")
	if _, err := os.Stat(filepath.Join(kernel, "pyproject.toml")); err != nil {
		return devCheckout{}, false
	}
	if info, err := os.Stat(tui); err != nil || !info.IsDir() {
		return devCheckout{}, false
	}
	return devCheckout{
		KernelSrc: filepath.Clean(kernel),
		TUISrc:    filepath.Clean(tui),
	}, true
}
