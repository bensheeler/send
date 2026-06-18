package scanner

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Options struct {
	LookupDepth int
}

type Result struct {
	Path string
}

// AmbiguousError preserves the exact candidate paths so the CLI can show the
// user why a selector could not be resolved safely.
type AmbiguousError struct {
	Selector string
	Matches  []string
}

// candidate is internal match bookkeeping. depth is the relative filesystem
// depth from cwd: direct children are 0, api/users.http is 1, and so on.
type candidate struct {
	path  string
	depth int
}

func (e *AmbiguousError) Error() string {
	return "ambiguous request file selector"
}

type InvalidOptionsError struct {
	Message string
}

func (e *InvalidOptionsError) Error() string {
	return e.Message
}

var ErrNotFound = errors.New("request file not found")

// FindRequestFile resolves the user's file selector to exactly one request
// file. It intentionally does not parse request contents; this package only
// answers "which .http/.rest file did the user mean?"
func FindRequestFile(cwd, selector string, opts Options) (Result, error) {
	if opts.LookupDepth < 0 {
		return Result{}, &InvalidOptionsError{Message: "lookup depth must be non-negative"}
	}

	if strings.TrimSpace(selector) == "" {
		return Result{}, ErrNotFound
	}

	if hasRequestFileExtension(selector) {
		// Explicit extensions mean exact relative paths. They do not get the
		// case-insensitive recursive lookup used for bare names.
		path, ok := selectorPath(cwd, selector)
		if !ok {
			return Result{}, ErrNotFound
		}
		_, ok, err := requestFileInfo(cwd, path)
		if err != nil {
			return Result{}, err
		}
		if !ok {
			return Result{}, ErrNotFound
		}

		return Result{Path: path}, nil
	}

	var matches []candidate
	var err error
	if isBareName(selector) {
		// Bare names are convenience selectors, so they search recursively and
		// case-insensitively within the configured lookup depth.
		matches, err = findBareNameMatches(cwd, selector, opts.LookupDepth)
	} else {
		// Selectors with path separators are exact relative stems, not fuzzy
		// suffix searches. api/users may match api/users.http or api/users.rest.
		matches, err = findExtensionlessPathMatches(cwd, selector)
	}
	if err != nil {
		return Result{}, err
	}

	match, err := chooseMatch(selector, matches)
	if err != nil {
		return Result{}, err
	}

	return Result{Path: match.path}, nil
}

func findExtensionlessPathMatches(cwd, selector string) ([]candidate, error) {
	basePath, ok := selectorPath(cwd, selector)
	if !ok {
		return nil, nil
	}
	matches := make([]candidate, 0, 2)

	// Extensionless path selectors may match either supported request-file
	// extension. If both exist, chooseMatch will report ambiguity.
	for _, ext := range []string{".http", ".rest"} {
		path := basePath + ext
		_, ok, err := requestFileInfo(cwd, path)
		if err != nil {
			return nil, err
		}
		if ok {
			matches = append(matches, candidate{path: path, depth: 0})
		}
	}

	return matches, nil
}

func findBareNameMatches(cwd, selector string, lookupDepth int) ([]candidate, error) {
	matches := make([]candidate, 0, 2)

	err := filepath.WalkDir(cwd, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == cwd {
			return nil
		}

		rel, err := filepath.Rel(cwd, path)
		if err != nil {
			return err
		}
		depth := strings.Count(rel, string(filepath.Separator))

		if entry.IsDir() {
			// Discovery skips dependency/build directories and never walks deeper
			// than lookupDepth. Explicit selectors are handled elsewhere and can
			// still point inside skipped directories.
			if shouldSkipDir(entry.Name()) || depth >= lookupDepth {
				return filepath.SkipDir
			}
			return nil
		}

		if depth > lookupDepth || entry.Type()&os.ModeSymlink != 0 || !hasRequestFileExtension(path) {
			return nil
		}
		stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if !strings.EqualFold(stem, selector) {
			return nil
		}
		_, ok, err := requestFileInfo(cwd, path)
		if err != nil {
			return err
		}
		if ok {
			matches = append(matches, candidate{path: path, depth: depth})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func chooseMatch(selector string, matches []candidate) (candidate, error) {
	if len(matches) == 0 {
		return candidate{}, ErrNotFound
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].depth != matches[j].depth {
			return matches[i].depth < matches[j].depth
		}
		return matches[i].path < matches[j].path
	})
	// Prefer the shallowest match so nearby files beat deeper duplicates. If the
	// shallowest depth has multiple matches, failing is safer than guessing.
	if len(matches) == 1 || matches[0].depth < matches[1].depth {
		return matches[0], nil
	}

	paths := []string{matches[0].path}
	for _, match := range matches[1:] {
		if match.depth != matches[0].depth {
			break
		}
		paths = append(paths, match.path)
	}
	return candidate{}, &AmbiguousError{Selector: selector, Matches: paths}
}

func requestFileInfo(cwd, path string) (os.FileInfo, bool, error) {
	// This helper centralizes the filesystem safety policy for selector-derived
	// paths: the candidate must remain inside cwd, match exact path casing, and
	// contain no symlink components. Returning ok=false means "not a match";
	// real filesystem errors are returned so callers can surface them later.
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return nil, false, err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, false, nil
	}

	current := cwd
	components := strings.Split(rel, string(filepath.Separator))
	for i, component := range components {
		// os.Lstat may succeed case-insensitively on some filesystems. Check the
		// directory entries first so explicit path selectors stay exact.
		if !componentExistsWithExactCase(current, component) {
			return nil, false, nil
		}
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			// Symlinks can point outside cwd even when the lexical path looks safe.
			return nil, false, nil
		}
		if i < len(components)-1 && !info.IsDir() {
			return nil, false, nil
		}
		if i == len(components)-1 {
			if info.IsDir() {
				return nil, false, nil
			}
			return info, true, nil
		}
	}

	return nil, false, nil
}

func selectorPath(cwd, selector string) (string, bool) {
	path := filepath.FromSlash(selector)
	// Reject paths that filepath.Join would normalize into something different.
	// Exact selectors should not accept traversal, dot components, repeated
	// separators, or absolute paths.
	if filepath.IsAbs(path) || hasInvalidPathComponent(path) {
		return "", false
	}
	return filepath.Join(cwd, path), true
}

func hasInvalidPathComponent(path string) bool {
	for _, component := range strings.Split(path, string(filepath.Separator)) {
		if component == "" || component == "." || component == ".." {
			return true
		}
	}
	return false
}

func componentExistsWithExactCase(dir, name string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.Name() == name {
			return true
		}
	}
	return false
}

func hasRequestFileExtension(selector string) bool {
	ext := filepath.Ext(selector)
	return ext == ".http" || ext == ".rest"
}

func isBareName(selector string) bool {
	return filepath.Base(filepath.FromSlash(selector)) == selector
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", "dist", "build", "target":
		return true
	default:
		return false
	}
}
