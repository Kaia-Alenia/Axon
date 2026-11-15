package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestNormalizeVersionTag(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"v1.6.9", "1.6.9"},
		{"V1.7.0", "1.7.0"},
		{"1.7.0", "1.7.0"},
		{"  v2.0.1  ", "2.0.1"},
		{"", ""},
	} {
		if got := normalizeVersionTag(tc.in); got != tc.want {
			t.Errorf("normalizeVersionTag(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseSemver(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want [3]int
	}{
		{"v1.6.9", [3]int{1, 6, 9}},
		{"1.7.0", [3]int{1, 7, 0}},
		{"1.7", [3]int{1, 7, 0}},
		{"2", [3]int{2, 0, 0}},
		{"1.6.9-rc1", [3]int{1, 6, 9}},
		{"v10.20.30", [3]int{10, 20, 30}},
	} {
		got, err := parseSemver(tc.in)
		if err != nil {
			t.Errorf("parseSemver(%q) returned error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseSemver(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
	for _, bad := range []string{"", "v", "1.2.3.4", "a.b.c", "1.x.0", "-1.2.3", "1.2.-3"} {
		if _, err := parseSemver(bad); err == nil {
			t.Errorf("parseSemver(%q) = no error, want an error", bad)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"1.6.9", "1.6.9", 0},
		{"v1.6.9", "1.6.9", 0},
		{"1.7.0", "1.6.9", 1},
		{"1.6.9", "1.7.0", -1},
		{"1.6.10", "1.6.9", 1},
		{"1.10.0", "1.9.9", 1},
		{"2.0.0", "1.99.99", 1},
		{"1.6.9", "2.0.0", -1},
	} {
		got, err := compareVersions(tc.a, tc.b)
		if err != nil {
			t.Errorf("compareVersions(%q, %q) returned error: %v", tc.a, tc.b, err)
			continue
		}
		if got != tc.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
	if _, err := compareVersions("banana", "1.0.0"); err == nil {
		t.Errorf("compareVersions with an invalid version returned no error")
	}
}

func TestFetchLatestReleaseVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"tag_name":"v1.7.0"}`))
		case "/notfound":
			http.NotFound(w, r)
		case "/badjson":
			_, _ = w.Write([]byte(`{not json`))
		case "/notag":
			_, _ = w.Write([]byte(`{"tag_name":""}`))
		}
	}))
	defer srv.Close()

	got, err := fetchLatestReleaseVersion(http.DefaultClient, srv.URL+"/ok")
	if err != nil {
		t.Fatalf("fetchLatestReleaseVersion(ok) returned error: %v", err)
	}
	if got != "1.7.0" {
		t.Errorf("fetchLatestReleaseVersion(ok) = %q, want %q", got, "1.7.0")
	}
	for _, path := range []string{"/notfound", "/badjson", "/notag"} {
		if _, err := fetchLatestReleaseVersion(http.DefaultClient, srv.URL+path); err == nil {
			t.Errorf("fetchLatestReleaseVersion(%s) = no error, want an error", path)
		}
	}
}

// writeFakeBinary writes magic + zero padding so the file is size bytes.
func writeFakeBinary(t *testing.T, path string, magic []byte, size int) {
	t.Helper()
	buf := make([]byte, size)
	copy(buf, magic)
	if err := os.WriteFile(path, buf, 0755); err != nil {
		t.Fatalf("could not write fake binary %s: %v", path, err)
	}
}

func TestValidateDownloadedBinary(t *testing.T) {
	dir := t.TempDir()
	elf := []byte{0x7f, 'E', 'L', 'F'}
	mz := []byte{'M', 'Z'}
	macho64le := []byte{0xcf, 0xfa, 0xed, 0xfe}
	fat := []byte{0xca, 0xfe, 0xba, 0xbe}
	bad := []byte{'N', 'O', 'P', 'E'}

	valid := []struct {
		name string
		goos string
		magic []byte
	}{
		{"windows MZ", "windows", mz},
		{"linux ELF", "linux", elf},
		{"macOS Mach-O 64 LE", "darwin", macho64le},
		{"macOS fat binary", "darwin", fat},
		{"unknown platform size-only", "plan9", bad},
	}
	for _, tc := range valid {
		p := filepath.Join(dir, "valid-"+tc.goos)
		writeFakeBinary(t, p, tc.magic, 1024)
		if err := validateDownloadedBinary(p, tc.goos); err != nil {
			t.Errorf("validateDownloadedBinary(%s) returned error: %v", tc.name, err)
		}
	}

	invalid := []struct {
		name string
		goos string
		magic []byte
	}{
		{"windows without MZ", "windows", elf},
		{"linux without ELF", "linux", mz},
		{"macOS without Mach-O", "darwin", elf},
	}
	for _, tc := range invalid {
		p := filepath.Join(dir, "invalid-"+tc.goos)
		writeFakeBinary(t, p, tc.magic, 1024)
		if err := validateDownloadedBinary(p, tc.goos); err == nil {
			t.Errorf("validateDownloadedBinary(%s) = no error, want an error", tc.name)
		}
	}

	tooSmall := filepath.Join(dir, "small")
	writeFakeBinary(t, tooSmall, elf, 100)
	if err := validateDownloadedBinary(tooSmall, "linux"); err == nil {
		t.Errorf("validateDownloadedBinary(too small) = no error, want an error")
	}

	tooBig := filepath.Join(dir, "big")
	f, err := os.Create(tooBig)
	if err != nil {
		t.Fatalf("could not create sparse file: %v", err)
	}
	_, _ = f.Write(elf)
	if err := f.Truncate(maxDownloadSize + 1); err != nil {
		t.Fatalf("could not grow sparse file: %v", err)
	}
	_ = f.Close()
	if err := validateDownloadedBinary(tooBig, "linux"); err == nil {
		t.Errorf("validateDownloadedBinary(too large) = no error, want an error")
	}

	if err := validateDownloadedBinary(filepath.Join(dir, "missing"), "linux"); err == nil {
		t.Errorf("validateDownloadedBinary(missing file) = no error, want an error")
	}
}

func TestDownloadUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("fake-binary-bytes"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	p := filepath.Join(dir, "axon.new")
	if err := downloadUpdate(srv.URL+"/axon", p); err != nil {
		t.Fatalf("downloadUpdate returned error: %v", err)
	}
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("could not read downloaded file: %v", err)
	}
	if string(got) != "fake-binary-bytes" {
		t.Errorf("downloaded content = %q, want %q", got, "fake-binary-bytes")
	}
	if err := downloadUpdate(srv.URL+"/missing", p); err == nil {
		t.Errorf("downloadUpdate(404) = no error, want an error")
	}
}

func TestWaitAndSwapSuccess(t *testing.T) {
	dir := t.TempDir()
	exePath := filepath.Join(dir, "axon")
	newPath := filepath.Join(dir, "axon.new")
	if err := os.WriteFile(exePath, []byte("old binary"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte("new binary"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := waitAndSwap(12345, exePath, newPath); err != nil {
		t.Fatalf("waitAndSwap returned error: %v", err)
	}
	got, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatalf("could not read swapped binary: %v", err)
	}
	if string(got) != "new binary" {
		t.Errorf("exe content after swap = %q, want %q", got, "new binary")
	}
	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		t.Errorf(".new file still exists after a successful swap")
	}
}

func TestSwapViaRenameAway(t *testing.T) {
	dir := t.TempDir()
	exePath := filepath.Join(dir, "axon")
	newPath := filepath.Join(dir, "axon.new")
	if err := os.WriteFile(exePath, []byte("old binary"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte("new binary"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := swapViaRenameAway(exePath, newPath); err != nil {
		t.Fatalf("swapViaRenameAway returned error: %v", err)
	}
	got, _ := os.ReadFile(exePath)
	if string(got) != "new binary" {
		t.Errorf("exe content after rename-away swap = %q, want %q", got, "new binary")
	}
	old, err := os.ReadFile(exePath + ".old")
	if err != nil {
		t.Fatalf("could not read .old file: %v", err)
	}
	if string(old) != "old binary" {
		t.Errorf(".old content = %q, want %q", old, "old binary")
	}
}

// shortenSwapWait makes the polling loops in waitAndSwap fail fast.
func shortenSwapWait(t *testing.T) {
	t.Helper()
	oldTimeout, oldInterval := updateSwapTimeout, updateSwapInterval
	oldFallback, oldMaxInterval := updateSwapFallbackTimeout, updateSwapMaxInterval
	updateSwapTimeout = 300 * time.Millisecond
	updateSwapInterval = 50 * time.Millisecond
	updateSwapFallbackTimeout = 600 * time.Millisecond
	updateSwapMaxInterval = 100 * time.Millisecond
	t.Cleanup(func() {
		updateSwapTimeout, updateSwapInterval = oldTimeout, oldInterval
		updateSwapFallbackTimeout, updateSwapMaxInterval = oldFallback, oldMaxInterval
	})
}

func TestWaitAndSwapMissingNew(t *testing.T) {
	shortenSwapWait(t)
	dir := t.TempDir()
	exePath := filepath.Join(dir, "axon")
	if err := os.WriteFile(exePath, []byte("old binary"), 0755); err != nil {
		t.Fatal(err)
	}
	// No .new file: the rename can never succeed, and the fallback must
	// restore the old binary instead of leaving the exe missing.
	if err := waitAndSwap(999999, exePath, filepath.Join(dir, "axon.new")); err == nil {
		t.Errorf("waitAndSwap(missing .new) = no error, want an error")
	}
	got, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatalf("exe missing after failed swap: %v", err)
	}
	if string(got) != "old binary" {
		t.Errorf("exe content after failed swap = %q, want %q", got, "old binary")
	}
	if _, err := os.Stat(exePath + ".old"); !os.IsNotExist(err) {
		t.Errorf(".old file left behind after a failed swap")
	}
}

func TestApplyUpdateFailureWritesLog(t *testing.T) {
	shortenSwapWait(t)
	dir := t.TempDir()
	exePath := filepath.Join(dir, "axon")
	if err := os.WriteFile(exePath, []byte("old binary"), 0755); err != nil {
		t.Fatal(err)
	}
	// Missing .new file, so the swap must fail and update.log must appear.
	if err := applyUpdate(999999, exePath, filepath.Join(dir, "axon.new")); err == nil {
		t.Errorf("applyUpdate(missing .new) = no error, want an error")
	}
	logBytes, err := os.ReadFile(filepath.Join(dir, "update.log"))
	if err != nil {
		t.Fatalf("update.log was not written: %v", err)
	}
	if !strings.Contains(string(logBytes), "Update failed") {
		t.Errorf("update.log = %q, want it to mention the failure", logBytes)
	}
	got, err := os.ReadFile(exePath)
	if err != nil || string(got) != "old binary" {
		t.Errorf("exe after failed applyUpdate = %q (err %v), want it untouched", got, err)
	}
	if err := applyUpdate(0, exePath, filepath.Join(dir, "axon.new")); err == nil {
		t.Errorf("applyUpdate(pid 0) = no error, want an error")
	}
}

func TestApplyUpdateSuccessLaunchesNew(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script as the fake new binary")
	}
	dir := t.TempDir()
	exePath := filepath.Join(dir, "axon")
	newPath := filepath.Join(dir, "axon.new")
	marker := filepath.Join(dir, "launched")
	oldScript := "#!/bin/sh\nexit 0\n"
	// The new binary must answer --version (the launch check) and,
	// when started plainly, prove it ran by touching the marker.
	newScript := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then echo \"Axon Server v9.9.9\"; exit 0; fi\ntouch " + marker + "\n"
	if err := os.WriteFile(exePath, []byte(oldScript), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte(newScript), 0755); err != nil {
		t.Fatal(err)
	}
	if err := applyUpdate(os.Getpid(), exePath, newPath); err != nil {
		t.Fatalf("applyUpdate returned error: %v", err)
	}
	got, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatalf("could not read swapped binary: %v", err)
	}
	if string(got) != newScript {
		t.Errorf("exe content after swap = %q, want the new script", got)
	}
	// The relaunched binary runs async; wait briefly for its marker.
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("relaunched binary did not run (no marker file %s)", marker)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestFetchLatestReleaseInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v1.7.0","assets":[{"name":"axon-linux-amd64","size":1234567},{"name":"axon-windows-amd64.exe","size":7654321}]}`))
	}))
	defer srv.Close()

	tag, size, err := fetchLatestReleaseInfo(http.DefaultClient, srv.URL, "axon-linux-amd64")
	if err != nil {
		t.Fatalf("fetchLatestReleaseInfo returned error: %v", err)
	}
	if tag != "1.7.0" {
		t.Errorf("tag = %q, want %q", tag, "1.7.0")
	}
	if size != 1234567 {
		t.Errorf("asset size = %d, want %d", size, 1234567)
	}
	// An asset the release does not list is unknown size, not an error.
	if _, size, err := fetchLatestReleaseInfo(http.DefaultClient, srv.URL, "axon-darwin-arm64"); err != nil {
		t.Fatalf("fetchLatestReleaseInfo(unknown asset) returned error: %v", err)
	} else if size != -1 {
		t.Errorf("unknown asset size = %d, want -1", size)
	}
}

func TestDownloadUpdateExpectedSizeMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/good":
			_, _ = w.Write([]byte("0123456789abcdef")) // 16 bytes
		case "/truncated":
			// Declare more than is sent: the client must see a short body.
			w.Header().Set("Content-Length", "100")
			_, _ = w.Write([]byte("short"))
		}
	}))
	defer srv.Close()

	dir := t.TempDir()

	// Exact size: succeeds and the file stays.
	p := filepath.Join(dir, "exact.new")
	if written, err := downloadUpdateExpected(srv.URL+"/good", p, 16); err != nil {
		t.Errorf("downloadUpdateExpected(exact size) returned error: %v", err)
	} else if written != 16 {
		t.Errorf("downloadUpdateExpected wrote %d bytes, want 16", written)
	}

	// Asset size mismatch: rejected and the partial file is deleted.
	p = filepath.Join(dir, "mismatch.new")
	if _, err := downloadUpdateExpected(srv.URL+"/good", p, 999); err == nil {
		t.Errorf("downloadUpdateExpected(asset size mismatch) = no error, want an error")
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Errorf("mismatched download left a file behind at %s", p)
	}

	// Truncated body (fewer bytes than Content-Length): rejected and deleted.
	p = filepath.Join(dir, "truncated.new")
	if _, err := downloadUpdateExpected(srv.URL+"/truncated", p, -1); err == nil {
		t.Errorf("downloadUpdateExpected(truncated body) = no error, want an error")
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Errorf("truncated download left a file behind at %s", p)
	}
}

func TestVerifyDownloadedSize(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bin")
	if err := os.WriteFile(p, make([]byte, 10), 0644); err != nil {
		t.Fatal(err)
	}
	if err := verifyDownloadedSize(p, 10); err != nil {
		t.Errorf("verifyDownloadedSize(exact) returned error: %v", err)
	}
	if err := verifyDownloadedSize(p, 11); err == nil {
		t.Errorf("verifyDownloadedSize(wrong size) = no error, want an error")
	}
	if err := verifyDownloadedSize(p, -1); err != nil {
		t.Errorf("verifyDownloadedSize(unknown size) returned error: %v", err)
	}
	if err := verifyDownloadedSize(filepath.Join(dir, "missing"), 10); err == nil {
		t.Errorf("verifyDownloadedSize(missing file) = no error, want an error")
	}
}

func TestParseVersionFromOutput(t *testing.T) {
	for _, tc := range []struct {
		in, want string
	}{
		{"Axon Server v1.7.0\nHigh-speed wireless input bridge", "1.7.0"},
		{"v2.0.1", "2.0.1"},
		{"Axon Server v1.7.0-rc1", "1.7.0-rc1"},
		{"version: 10.20.30 (build 4)", "10.20.30"},
	} {
		got, err := parseVersionFromOutput(tc.in)
		if err != nil {
			t.Errorf("parseVersionFromOutput(%q) returned error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseVersionFromOutput(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
	for _, bad := range []string{"", "no version here", "port 6969", "Axon Server"} {
		if _, err := parseVersionFromOutput(bad); err == nil {
			t.Errorf("parseVersionFromOutput(%q) = no error, want an error", bad)
		}
	}
}

func TestVerifyExeVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses shell scripts as fake binaries")
	}
	dir := t.TempDir()
	good := filepath.Join(dir, "axon-good")
	goodScript := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then echo \"Axon Server v9.9.9\"; exit 0; fi\nexit 1\n"
	if err := os.WriteFile(good, []byte(goodScript), 0755); err != nil {
		t.Fatal(err)
	}
	if err := verifyExeVersion(good, "9.9.9"); err != nil {
		t.Errorf("verifyExeVersion(matching version) returned error: %v", err)
	}
	if err := verifyExeVersion(good, ""); err != nil {
		t.Errorf("verifyExeVersion(any version) returned error: %v", err)
	}
	if err := verifyExeVersion(good, "1.0.0"); err == nil {
		t.Errorf("verifyExeVersion(wrong version) = no error, want an error")
	}
	silent := filepath.Join(dir, "axon-silent")
	if err := os.WriteFile(silent, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := verifyExeVersion(silent, ""); err == nil {
		t.Errorf("verifyExeVersion(no version in output) = no error, want an error")
	}
	if err := verifyExeVersion(filepath.Join(dir, "missing"), ""); err == nil {
		t.Errorf("verifyExeVersion(missing binary) = no error, want an error")
	}
}

func TestCheckFreeDiskSpace(t *testing.T) {
	dir := t.TempDir()
	if err := checkFreeDiskSpace(dir, 1); err != nil {
		t.Errorf("checkFreeDiskSpace(tiny requirement) returned error: %v", err)
	}
	if err := checkFreeDiskSpace(dir, 0); err != nil {
		t.Errorf("checkFreeDiskSpace(no requirement) returned error: %v", err)
	}
	if err := checkFreeDiskSpace(dir, 1<<62); err == nil {
		t.Errorf("checkFreeDiskSpace(impossible requirement) = no error, want an error")
	} else if !strings.Contains(err.Error(), "free disk space") {
		t.Errorf("checkFreeDiskSpace error = %q, want it to mention free disk space", err)
	}
	// A directory that cannot even be measured only warns and lets the
	// update continue (the download is verified byte-for-byte anyway).
	if err := checkFreeDiskSpace(filepath.Join(dir, "missing"), 1); err != nil {
		t.Errorf("checkFreeDiskSpace(unmeasurable dir) returned error: %v", err)
	}
}

func TestCheckInstallDirWritable(t *testing.T) {
	dir := t.TempDir()
	if err := checkInstallDirWritable(dir); err != nil {
		t.Errorf("checkInstallDirWritable(writable dir) returned error: %v", err)
	}
	if err := checkInstallDirWritable(filepath.Join(dir, "missing")); err == nil {
		t.Errorf("checkInstallDirWritable(missing dir) = no error, want an error")
	}
	file := filepath.Join(dir, "afile")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := checkInstallDirWritable(file); err == nil {
		t.Errorf("checkInstallDirWritable(file, not a dir) = no error, want an error")
	}
}

func TestRequiredDownloadBytes(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "axon")
	if err := os.WriteFile(exe, make([]byte, 100), 0755); err != nil {
		t.Fatal(err)
	}
	if got := requiredDownloadBytes(1000, exe); got != 1000+updateDiskSpaceMargin {
		t.Errorf("requiredDownloadBytes(known size) = %d, want %d", got, 1000+updateDiskSpaceMargin)
	}
	if got := requiredDownloadBytes(-1, exe); got != 100+updateDiskSpaceMargin {
		t.Errorf("requiredDownloadBytes(unknown size) = %d, want current exe size + margin (%d)", got, 100+updateDiskSpaceMargin)
	}
	if got := requiredDownloadBytes(0, filepath.Join(dir, "missing")); got != updateDiskSpaceMargin {
		t.Errorf("requiredDownloadBytes(unknown size, missing exe) = %d, want just the margin (%d)", got, updateDiskSpaceMargin)
	}
}

func TestBackupCurrentExeAndRestore(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "axon")
	if err := os.WriteFile(exe, []byte("old binary"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := backupCurrentExe(exe); err != nil {
		t.Fatalf("backupCurrentExe returned error: %v", err)
	}
	old, err := os.ReadFile(exe + ".old")
	if err != nil || string(old) != "old binary" {
		t.Fatalf("backup content = %q (err %v), want %q", old, err, "old binary")
	}
	if err := os.WriteFile(exe, []byte("new binary"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := restoreFromBackup(exe); err != nil {
		t.Fatalf("restoreFromBackup returned error: %v", err)
	}
	got, err := os.ReadFile(exe)
	if err != nil || string(got) != "old binary" {
		t.Fatalf("exe after restore = %q (err %v), want %q", got, err, "old binary")
	}
	if _, err := os.Stat(exe + ".old"); !os.IsNotExist(err) {
		t.Errorf("backup was not consumed by the restore")
	}
	if err := restoreFromBackup(exe); err == nil {
		t.Errorf("restoreFromBackup(no backup) = no error, want an error")
	}
}

func TestApplyUpdateLaunchCheckFailureRestoresBackup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses shell scripts as fake binaries")
	}
	dir := t.TempDir()
	exePath := filepath.Join(dir, "axon")
	newPath := filepath.Join(dir, "axon.new")
	oldScript := "#!/bin/sh\nexit 0\n"
	// Swaps in fine, but reports the wrong version, so the launch
	// check must fail and the old binary must come back.
	newScript := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then echo \"Axon Server v0.0.0\"; exit 0; fi\nexit 0\n"
	if err := os.WriteFile(exePath, []byte(oldScript), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte(newScript), 0755); err != nil {
		t.Fatal(err)
	}
	err := applyUpdateExpected(os.Getpid(), exePath, newPath, "9.9.9")
	if err == nil {
		t.Fatalf("applyUpdateExpected(wrong --version) = no error, want an error")
	}
	if !strings.Contains(err.Error(), "old version restored") {
		t.Errorf("error = %q, want it to say the old version was restored", err)
	}
	got, err := os.ReadFile(exePath)
	if err != nil || string(got) != oldScript {
		t.Fatalf("exe after failed launch check = %q (err %v), want the old binary back", got, err)
	}
	if _, err := os.Stat(exePath + ".old"); !os.IsNotExist(err) {
		t.Errorf(".old file left behind after a restore")
	}
	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		t.Errorf(".new file left behind after a failed launch check")
	}
	logBytes, err := os.ReadFile(filepath.Join(dir, "update.log"))
	if err != nil {
		t.Fatalf("update.log was not written: %v", err)
	}
	if !strings.Contains(string(logBytes), "Update failed") {
		t.Errorf("update.log = %q, want it to mention the failure", logBytes)
	}
}

func TestApplyUpdatePathWithSpaces(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses shell scripts as fake binaries")
	}
	dir := filepath.Join(t.TempDir(), "dir with spaces")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	exePath := filepath.Join(dir, "my axon")
	newPath := filepath.Join(dir, "my axon.new")
	marker := filepath.Join(dir, "launched marker")
	oldScript := "#!/bin/sh\nexit 0\n"
	newScript := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then echo \"Axon Server v9.9.9\"; exit 0; fi\ntouch '" + marker + "'\n"
	if err := os.WriteFile(exePath, []byte(oldScript), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte(newScript), 0755); err != nil {
		t.Fatal(err)
	}
	if err := applyUpdateExpected(os.Getpid(), exePath, newPath, "9.9.9"); err != nil {
		t.Fatalf("applyUpdateExpected(path with spaces) returned error: %v", err)
	}
	got, err := os.ReadFile(exePath)
	if err != nil || string(got) != newScript {
		t.Fatalf("exe after swap = %q (err %v), want the new script", got, err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("relaunched binary did not run (no marker file %s)", marker)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestCleanupStaleBackup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses shell scripts as fake binaries")
	}
	dir := t.TempDir()
	exe := filepath.Join(dir, "axon")
	goodScript := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then echo \"Axon Server v1.0.0\"; exit 0; fi\nexit 0\n"
	if err := os.WriteFile(exe, []byte(goodScript), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exe+".old", []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	cleanupStaleBackup(exe)
	if _, err := os.Stat(exe + ".old"); !os.IsNotExist(err) {
		t.Errorf(".old should be removed once the current binary passes the launch check")
	}
	// A current binary that cannot even answer --version: the backup
	// is the only way back, so it must be kept.
	if err := os.WriteFile(exe, []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exe+".old", []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	cleanupStaleBackup(exe)
	if _, err := os.Stat(exe + ".old"); err != nil {
		t.Errorf(".old should be kept when the current binary fails the launch check: %v", err)
	}
	// No backup at all: a no-op, never a crash.
	cleanupStaleBackup(filepath.Join(dir, "missing"))
}

func TestFormatBytes(t *testing.T) {
	for _, tc := range []struct {
		in   uint64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1024 * 1024, "1.0 MiB"},
		{5 * 1024 * 1024 * 1024, "5.00 GiB"},
	} {
		if got := formatBytes(tc.in); got != tc.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
