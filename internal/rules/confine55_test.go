package rules

// Tests for BLD-001 / BLD-002: go/cargo builds whose caches are not confined
// to $srcdir write into the invoking user's $HOME (issue #55).

import "testing"

func scanCodes(t *testing.T, pkgbuild string) map[string]bool {
	t.Helper()
	codes := map[string]bool{}
	for _, h := range Scan(map[string]string{"PKGBUILD": pkgbuild}) {
		codes[h.Code] = true
	}
	return codes
}

func TestGoBuildWithoutGopathFires(t *testing.T) {
	codes := scanCodes(t, `pkgname=foo
build() {
  cd "$srcdir/$pkgname"
  go build -o foo .
}`)
	if !codes["BLD-001"] {
		t.Fatalf("expected BLD-001, got %v", codes)
	}
	if codes["BLD-002"] {
		t.Fatalf("BLD-002 must not fire on a go-only PKGBUILD, got %v", codes)
	}
}

func TestGoBuildWithExportedGopathClean(t *testing.T) {
	codes := scanCodes(t, `pkgname=foo
build() {
  export GOPATH="$srcdir/gopath"
  go build -o foo .
}`)
	if codes["BLD-001"] {
		t.Fatalf("BLD-001 must not fire when GOPATH is confined, got %v", codes)
	}
}

func TestGoBuildWithInlineGomodcacheClean(t *testing.T) {
	codes := scanCodes(t, `pkgname=foo
build() {
  GOMODCACHE="$srcdir/gomod" go build -o foo .
}`)
	if codes["BLD-001"] {
		t.Fatalf("BLD-001 must not fire on inline GOMODCACHE prefix, got %v", codes)
	}
}

func TestGoBuildVendoredClean(t *testing.T) {
	// A vendored build does not touch the module cache.
	codes := scanCodes(t, `pkgname=foo
build() {
  export GOFLAGS="-buildmode=pie -trimpath -mod=vendor"
  go build -o foo .
}`)
	if codes["BLD-001"] {
		t.Fatalf("BLD-001 must not fire with -mod=vendor, got %v", codes)
	}
}

func TestGopathOnlyInCommentStillFires(t *testing.T) {
	// A commented-out assignment is inert and must not count as confinement.
	codes := scanCodes(t, `pkgname=foo
build() {
  # export GOPATH="$srcdir/gopath"
  go install ./...
}`)
	if !codes["BLD-001"] {
		t.Fatalf("expected BLD-001 despite commented GOPATH, got %v", codes)
	}
}

func TestEchoGoBuildDoesNotFire(t *testing.T) {
	// echo arguments are printed data, not a command position.
	codes := scanCodes(t, `pkgname=foo
build() {
  echo "run: go build -o foo ." > README
  make
}`)
	if codes["BLD-001"] {
		t.Fatalf("BLD-001 must not fire on echo'd text, got %v", codes)
	}
}

func TestGoVersionDoesNotFire(t *testing.T) {
	// Non-writing subcommands (version/env/clean) do not touch the caches.
	codes := scanCodes(t, `pkgname=foo
build() {
  go version
  go env GOOS
  make
}`)
	if codes["BLD-001"] {
		t.Fatalf("BLD-001 must not fire on go version/env, got %v", codes)
	}
}

func TestCargoWithoutCargoHomeFires(t *testing.T) {
	codes := scanCodes(t, `pkgname=bar
prepare() {
  cd "$srcdir/$pkgname"
  cargo fetch --locked
}
build() {
  cargo build --release --frozen
}`)
	if !codes["BLD-002"] {
		t.Fatalf("expected BLD-002, got %v", codes)
	}
	if codes["BLD-001"] {
		t.Fatalf("BLD-001 must not fire on a cargo-only PKGBUILD (\"cargo build\" is not \"go build\"), got %v", codes)
	}
}

func TestCargoWithCargoHomeClean(t *testing.T) {
	// An export in prepare() covers build(): same makepkg process.
	codes := scanCodes(t, `pkgname=bar
prepare() {
  export CARGO_HOME="$srcdir/cargo-home"
  cargo fetch --locked
}
build() {
  cargo build --release --frozen
}`)
	if codes["BLD-002"] {
		t.Fatalf("BLD-002 must not fire when CARGO_HOME is confined, got %v", codes)
	}
}

func TestNoBuildToolsNoHits(t *testing.T) {
	codes := scanCodes(t, `pkgname=baz
build() {
  ./configure --prefix=/usr
  make
}`)
	if codes["BLD-001"] || codes["BLD-002"] {
		t.Fatalf("BLD rules must not fire without go/cargo, got %v", codes)
	}
}

func TestBldSeverityIsInformational(t *testing.T) {
	hits := Scan(map[string]string{"PKGBUILD": "build() { go build .; }"})
	for _, h := range hits {
		if h.Code == "BLD-001" && h.Severity != Medium {
			t.Fatalf("BLD-001 severity = %q, want %q", h.Severity, Medium)
		}
	}
}

func TestRawFallbackStillDetects(t *testing.T) {
	// Unparseable shell (unclosed function body) falls back to raw-text
	// matching; detection must not be lost.
	src := `pkgname=foo
build() {
  go build -o foo .
`
	if _, err := extractCommands(src); err == nil {
		t.Fatal("fixture must fail shell parsing so the raw path is exercised")
	}
	codes := scanCodes(t, src)
	if !codes["BLD-001"] {
		t.Fatalf("expected BLD-001 via raw fallback, got %v", codes)
	}
}
