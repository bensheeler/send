package scanner

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFindRequestFileFindsExplicitHTTPPath(t *testing.T) {
	cwd := t.TempDir()
	writeFile(t, cwd, "requests/users.http", "GET http://example.com\n")

	result, err := FindRequestFile(cwd, "requests/users.http", Options{})
	if err != nil {
		t.Fatalf("FindRequestFile returned error: %v", err)
	}

	assertBasePath(t, result.Path, filepath.Join(cwd, "requests", "users.http"))
}

func TestFindRequestFileFindsExplicitRESTPath(t *testing.T) {
	cwd := t.TempDir()
	writeFile(t, cwd, "requests/users.rest", "GET http://example.com\n")

	result, err := FindRequestFile(cwd, "requests/users.rest", Options{})
	if err != nil {
		t.Fatalf("FindRequestFile returned error: %v", err)
	}

	assertBasePath(t, result.Path, filepath.Join(cwd, "requests", "users.rest"))
}

func TestFindRequestFileReturnsNotFoundForMissingExplicitPath(t *testing.T) {
	dir := t.TempDir()

	_, err := FindRequestFile(dir, "missing.http", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileReturnsNotFoundForExplicitDirectoryPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "requests.http")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	_, err := FindRequestFile(dir, "requests.http", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileReturnsNotFoundForUppercaseExplicitExtension(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "users.HTTP", "GET http://example.com\n")

	_, err := FindRequestFile(dir, "users.HTTP", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileRejectsNegativeLookupDepth(t *testing.T) {
	dir := t.TempDir()

	_, err := FindRequestFile(dir, "users", Options{LookupDepth: -1})
	var invalid *InvalidOptionsError
	if !errors.As(err, &invalid) {
		t.Fatalf("error = %v, want InvalidOptionsError", err)
	}
}

func TestFindRequestFileReturnsNotFoundForExplicitPathEscapingCWD(t *testing.T) {
	baseDir := t.TempDir()
	cwd := filepath.Join(baseDir, "cwd")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeFile(t, baseDir, "outside.http", "GET http://example.com\n")

	_, err := FindRequestFile(cwd, "../outside.http", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileReturnsNotFoundForExplicitPathContainingDotDot(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "users.http", "GET http://example.com\n")

	_, err := FindRequestFile(dir, "api/../users.http", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileReturnsNotFoundForExplicitPathContainingDotOrEmptyComponent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "api/users.http", "GET http://example.com\n")

	for _, selector := range []string{"api/./users.http", "api//users.http"} {
		t.Run(selector, func(t *testing.T) {
			_, err := FindRequestFile(dir, selector, Options{LookupDepth: 3})
			if !errors.Is(err, ErrNotFound) {
				t.Fatalf("error = %v, want ErrNotFound", err)
			}
		})
	}
}

func TestFindRequestFileRequiresExactCaseForExplicitPath(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "api/Users.http", "GET http://example.com\n")

	_, err := FindRequestFile(dir, "api/users.http", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileReturnsNotFoundForExplicitSymlinkPath(t *testing.T) {
	baseDir := t.TempDir()
	cwd := filepath.Join(baseDir, "cwd")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeFile(t, baseDir, "outside.http", "GET http://example.com\n")
	if err := os.Symlink(filepath.Join(baseDir, "outside.http"), filepath.Join(cwd, "link.http")); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	_, err := FindRequestFile(cwd, "link.http", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileReturnsNotFoundForExplicitPathThroughSymlinkedDirectory(t *testing.T) {
	baseDir := t.TempDir()
	cwd := filepath.Join(baseDir, "cwd")
	outsideDir := filepath.Join(baseDir, "outside")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeFile(t, outsideDir, "outside.http", "GET http://example.com\n")
	if err := os.Symlink(outsideDir, filepath.Join(cwd, "linked")); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	_, err := FindRequestFile(cwd, "linked/outside.http", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileFindsExtensionlessPath(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "api/users.rest", "GET http://example.com\n")

	result, err := FindRequestFile(dir, "api/users", Options{LookupDepth: 3})
	if err != nil {
		t.Fatalf("FindRequestFile returned error: %v", err)
	}

	assertBasePath(t, result.Path, filepath.Join(dir, "api", "users.rest"))
}

func TestFindRequestFileRejectsAmbiguousExtensionlessPath(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "api/users.http", "GET http://example.com\n")
	writeFile(t, dir, "api/users.rest", "GET http://example.com\n")

	_, err := FindRequestFile(dir, "api/users", Options{LookupDepth: 3})
	var ambiguous *AmbiguousError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("error = %v, want AmbiguousError", err)
	}
	want := []string{
		filepath.Join(dir, "api", "users.http"),
		filepath.Join(dir, "api", "users.rest"),
	}
	assertPaths(t, ambiguous.Matches, want)
}

func TestFindRequestFileReturnsNotFoundForMissingExtensionlessPath(t *testing.T) {
	dir := t.TempDir()

	_, err := FindRequestFile(dir, "api/users", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileReturnsNotFoundForExtensionlessPathEscapingCWD(t *testing.T) {
	baseDir := t.TempDir()
	cwd := filepath.Join(baseDir, "cwd")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeFile(t, baseDir, "outside.rest", "GET http://example.com\n")

	_, err := FindRequestFile(cwd, "../outside", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileReturnsNotFoundForExtensionlessPathContainingDotDot(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "users.rest", "GET http://example.com\n")

	_, err := FindRequestFile(dir, "api/../users", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileReturnsNotFoundForExtensionlessPathContainingDotOrEmptyComponent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "api/users.rest", "GET http://example.com\n")

	for _, selector := range []string{"api/./users", "api//users"} {
		t.Run(selector, func(t *testing.T) {
			_, err := FindRequestFile(dir, selector, Options{LookupDepth: 3})
			if !errors.Is(err, ErrNotFound) {
				t.Fatalf("error = %v, want ErrNotFound", err)
			}
		})
	}
}

func TestFindRequestFileRequiresExactCaseForExtensionlessPath(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "api/Users.rest", "GET http://example.com\n")

	_, err := FindRequestFile(dir, "api/users", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileReturnsNotFoundForExtensionlessSymlinkPath(t *testing.T) {
	baseDir := t.TempDir()
	cwd := filepath.Join(baseDir, "cwd")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeFile(t, baseDir, "outside.rest", "GET http://example.com\n")
	if err := os.Symlink(filepath.Join(baseDir, "outside.rest"), filepath.Join(cwd, "link.rest")); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	_, err := FindRequestFile(cwd, "link", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileReturnsNotFoundForExtensionlessPathThroughSymlinkedDirectory(t *testing.T) {
	baseDir := t.TempDir()
	cwd := filepath.Join(baseDir, "cwd")
	outsideDir := filepath.Join(baseDir, "outside")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeFile(t, outsideDir, "outside.http", "GET http://example.com\n")
	if err := os.Symlink(outsideDir, filepath.Join(cwd, "linked")); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	_, err := FindRequestFile(cwd, "linked/outside", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileFindsBareNameWithinLookupDepth(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "api/users.http", "GET http://example.com\n")

	result, err := FindRequestFile(dir, "users", Options{LookupDepth: 3})
	if err != nil {
		t.Fatalf("FindRequestFile returned error: %v", err)
	}
	assertBasePath(t, result.Path, filepath.Join(dir, "api", "users.http"))
}

func TestFindRequestFileMatchesBareNameCaseInsensitively(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "api/Users.http", "GET http://example.com\n")

	result, err := FindRequestFile(dir, "users", Options{LookupDepth: 3})
	if err != nil {
		t.Fatalf("FindRequestFile returned error: %v", err)
	}
	assertBasePath(t, result.Path, filepath.Join(dir, "api", "Users.http"))
}

func TestFindRequestFileRejectsCaseVariantAmbiguity(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "api/Users.http", "GET http://example.com\n")
	writeFile(t, dir, "api/users.rest", "GET http://example.com\n")

	_, err := FindRequestFile(dir, "users", Options{LookupDepth: 3})
	var ambiguous *AmbiguousError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("error = %v, want AmbiguousError", err)
	}
}

func TestFindRequestFileRejectsRootCaseVariantAmbiguity(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "users.http", "GET http://example.com\n")
	writeFile(t, dir, "Users.rest", "GET http://example.com\n")

	_, err := FindRequestFile(dir, "users", Options{LookupDepth: 3})
	var ambiguous *AmbiguousError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("error = %v, want AmbiguousError", err)
	}
}

func TestFindRequestFilePrefersShallowestBareNameMatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "api/v1/users.http", "GET http://example.com\n")
	writeFile(t, dir, "api/v1/internal/users.http", "GET http://example.com\n")

	result, err := FindRequestFile(dir, "users", Options{LookupDepth: 3})
	if err != nil {
		t.Fatalf("FindRequestFile returned error: %v", err)
	}
	assertBasePath(t, result.Path, filepath.Join(dir, "api", "v1", "users.http"))
}

func TestFindRequestFileRejectsSameDepthBareNameAmbiguity(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "api/users.http", "GET http://example.com\n")
	writeFile(t, dir, "web/users.rest", "GET http://example.com\n")
	writeFile(t, dir, "api/v1/users.http", "GET http://example.com\n")

	_, err := FindRequestFile(dir, "users", Options{LookupDepth: 3})
	var ambiguous *AmbiguousError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("error = %v, want AmbiguousError", err)
	}
	want := []string{
		filepath.Join(dir, "api", "users.http"),
		filepath.Join(dir, "web", "users.rest"),
	}
	assertPaths(t, ambiguous.Matches, want)
}

func TestFindRequestFileRespectsLookupDepth(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "api/v1/users.http", "GET http://example.com\n")

	_, err := FindRequestFile(dir, "users", Options{LookupDepth: 1})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileLookupDepthZeroOnlyIncludesDirectChildren(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "users.http", "GET http://example.com\n")
	writeFile(t, dir, "api/orders.http", "GET http://example.com\n")

	result, err := FindRequestFile(dir, "users", Options{LookupDepth: 0})
	if err != nil {
		t.Fatalf("FindRequestFile returned error: %v", err)
	}
	assertBasePath(t, result.Path, filepath.Join(dir, "users.http"))

	_, err = FindRequestFile(dir, "orders", Options{LookupDepth: 0})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFilePathSelectorDoesNotUseFuzzySuffixLookup(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "api/v1/users.http", "GET http://example.com\n")

	_, err := FindRequestFile(dir, "v1/users", Options{LookupDepth: 3})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestFindRequestFileSkipsIgnoredDirectoriesDuringBareNameLookup(t *testing.T) {
	skippedDirs := []string{".git", "node_modules", "vendor", "dist", "build", "target"}

	for _, skippedDir := range skippedDirs {
		t.Run(skippedDir, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, filepath.Join(skippedDir, "users.http"), "GET http://example.com\n")

			_, err := FindRequestFile(dir, "users", Options{LookupDepth: 3})
			if !errors.Is(err, ErrNotFound) {
				t.Fatalf("error = %v, want ErrNotFound", err)
			}
		})
	}
}

func TestFindRequestFileAllowsExplicitPathInsideSkippedDirectory(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "node_modules/users.http", "GET http://example.com\n")

	result, err := FindRequestFile(dir, "node_modules/users.http", Options{LookupDepth: 3})
	if err != nil {
		t.Fatalf("FindRequestFile returned error: %v", err)
	}
	assertBasePath(t, result.Path, filepath.Join(dir, "node_modules", "users.http"))
}

func writeFile(t *testing.T, basePath, name, contents string) {
	t.Helper()

	path := filepath.Join(basePath, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func assertBasePath(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}

func assertPaths(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("paths = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("paths = %v, want %v", got, want)
		}
	}
}
