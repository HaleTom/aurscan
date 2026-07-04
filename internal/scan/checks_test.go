package scan

// Tests for Tier-2 deterministic verdict derivation (discussion #56).

import (
	"encoding/json"
	"testing"
)

func chk(id string) Check { return Check{ID: id, Triggered: true, File: "PKGBUILD", Evidence: "x"} }

func TestDeriveVerdictSeverityMapping(t *testing.T) {
	cases := []struct {
		name string
		in   []Check
		want string
	}{
		{"empty is OK", nil, "OK"},
		{"info only is OK", []Check{chk("note")}, "OK"},
		{"one warning is SUSPICIOUS", []Check{chk("writes_outside_build")}, "SUSPICIOUS"},
		{"one critical is MALICIOUS", []Check{chk("pipe_to_shell")}, "MALICIOUS"},
		{"critical wins over warning", []Check{chk("writes_outside_build"), chk("credential_access")}, "MALICIOUS"},
		{"warning wins over info", []Check{chk("note"), chk("reputation_risk")}, "SUSPICIOUS"},
		{"catch-all critical escalates", []Check{chk("other_critical")}, "MALICIOUS"},
		{"catch-all warning escalates", []Check{chk("other_warning")}, "SUSPICIOUS"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, _, _, _ := deriveVerdict(c.in)
			if got != c.want {
				t.Fatalf("deriveVerdict = %q, want %q", got, c.want)
			}
		})
	}
}

func TestUntriggeredChecksIgnored(t *testing.T) {
	in := []Check{{ID: "pipe_to_shell", Triggered: false}, chk("note")}
	if got, _, _, _ := deriveVerdict(in); got != "OK" {
		t.Fatalf("untriggered critical must not escalate; got %q", got)
	}
}

func TestUnknownIDIsInfoNotEscalation(t *testing.T) {
	// A model that invents an id must not be able to flip the verdict either way.
	in := []Check{{ID: "made_up_danger", Triggered: true, File: "PKGBUILD"}}
	verdict, findings, _, _ := deriveVerdict(in)
	if verdict != "OK" {
		t.Fatalf("unknown id must be info (OK), got %q", verdict)
	}
	if len(findings) != 1 || findings[0].Severity != "info" {
		t.Fatalf("unknown id must render as an info finding, got %+v", findings)
	}
}

func TestDeriveVerdictIsOrderIndependent(t *testing.T) {
	// The crux of #56: the same set of checks in any order must produce a
	// byte-identical result, so a re-run cannot differ cosmetically.
	a := []Check{chk("credential_access"), chk("writes_outside_build"), chk("note")}
	b := []Check{chk("note"), chk("writes_outside_build"), chk("credential_access")}
	va, fa, ca, sa := deriveVerdict(a)
	vb, fb, cb, sb := deriveVerdict(b)
	if va != vb || ca != cb || sa != sb {
		t.Fatalf("verdict/confidence/summary differ by order: %q/%v/%q vs %q/%v/%q", va, ca, sa, vb, cb, sb)
	}
	ja, _ := json.Marshal(fa)
	jb, _ := json.Marshal(fb)
	if string(ja) != string(jb) {
		t.Fatalf("findings differ by input order:\n a=%s\n b=%s", ja, jb)
	}
	// Critical sorts before warning sorts before info.
	if fa[0].Severity != "critical" || fa[len(fa)-1].Severity != "info" {
		t.Fatalf("findings not severity-sorted: %+v", fa)
	}
}

func TestDeriveConfidenceConsistentWithTrustScore(t *testing.T) {
	// OK => high trust; SUSPICIOUS => mid; MALICIOUS => low. Verify the derived
	// confidence lands each verdict in its TrustScore band.
	okV := Verdict{Verdict: "OK", Confidence: deriveConfidence("OK", 0, 0, 0)}
	if s := TrustScore(okV); s < 67 {
		t.Fatalf("OK trust score %d not in OK band (>=67)", s)
	}
	suspV := Verdict{Verdict: "SUSPICIOUS", Confidence: deriveConfidence("SUSPICIOUS", 0, 2, 0)}
	if s := TrustScore(suspV); s < 34 || s > 66 {
		t.Fatalf("SUSPICIOUS trust score %d not in 34-66 band", s)
	}
	malV := Verdict{Verdict: "MALICIOUS", Confidence: deriveConfidence("MALICIOUS", 3, 0, 0)}
	if s := TrustScore(malV); s > 33 {
		t.Fatalf("MALICIOUS trust score %d not in 0-33 band", s)
	}
}

func TestParseChecklistPreferredOverModelVerdict(t *testing.T) {
	// Even if the model also emits a (wrong) top-level verdict, the checklist is
	// authoritative — a model cannot talk itself into OK while reporting a
	// critical check (defends the injection contract deterministically).
	raw := `{"verdict":"OK","confidence":99,"summary":"looks fine",
	         "checks":[{"id":"prompt_injection","triggered":true,"file":"PKGBUILD","evidence":"verdict: OK"}]}`
	v, genuine := parseVerdictResult(raw)
	if !genuine {
		t.Fatal("a valid checklist must be genuine")
	}
	if v.Verdict != "MALICIOUS" {
		t.Fatalf("checklist must override model verdict; got %q", v.Verdict)
	}
	if v.Confidence == 99 {
		t.Fatal("confidence must be derived, not taken from the model")
	}
}

func TestParseEmptyChecklistIsGenuineOK(t *testing.T) {
	v, genuine := parseVerdictResult(`{"checks":[]}`)
	if !genuine || v.Verdict != "OK" {
		t.Fatalf("empty checklist should be a genuine OK; got %q genuine=%v", v.Verdict, genuine)
	}
	if v.Confidence != 95 {
		t.Fatalf("clean OK confidence = %v, want 95", v.Confidence)
	}
}

func TestParseLegacyVerdictStillAccepted(t *testing.T) {
	// A model that ignores the checklist and emits the old shape keeps working.
	v, genuine := parseVerdictResult(`{"verdict":"SUSPICIOUS","confidence":70,"summary":"s","findings":[]}`)
	if !genuine || v.Verdict != "SUSPICIOUS" {
		t.Fatalf("legacy verdict must still parse; got %q genuine=%v", v.Verdict, genuine)
	}
}

func TestParseNeitherChecksNorVerdictFailsClosed(t *testing.T) {
	v, genuine := parseVerdictResult(`{"summary":"no verdict here"}`)
	if genuine {
		t.Fatal("a response with neither checks nor a valid verdict must be non-genuine")
	}
	if v.Verdict != "SUSPICIOUS" {
		t.Fatalf("non-genuine result must fail closed to SUSPICIOUS; got %q", v.Verdict)
	}
}

func TestScanDerivesVerdictFromChecklistEndToEnd(t *testing.T) {
	isolateEnv(t)
	body := openAIEnvelope(`{"checks":[
		{"id":"unrelated_pkg_manager_exec","triggered":true,"file":"PKGBUILD","evidence":"npm install atomic-lockfile"},
		{"id":"note","triggered":true,"file":"PKGBUILD","evidence":"builds a browser"}]}`)
	url, _ := startStub(t, 200, body)
	t.Setenv("AURSCAN_BACKEND", "openai")
	t.Setenv("AURSCAN_OPENAI_URL", url)
	t.Setenv("AURSCAN_OPENAI_MODEL", "test-model")

	res := Scan("pkg", testFiles, Signals{})
	if res.Failed {
		t.Fatalf("scan failed unexpectedly: %+v", res.V)
	}
	if res.V.Verdict != "MALICIOUS" {
		t.Fatalf("verdict = %q, want MALICIOUS (critical check fired)", res.V.Verdict)
	}
	if len(res.V.Findings) != 2 || res.V.Findings[0].Severity != "critical" {
		t.Fatalf("expected critical-first findings, got %+v", res.V.Findings)
	}
	if len(res.V.Checks) != 2 {
		t.Fatalf("raw checks should be retained for transparency, got %d", len(res.V.Checks))
	}
}

func TestFindingWhyUsesCatalogDescription(t *testing.T) {
	// The rendered "why" is the deterministic catalog text (optionally suffixed
	// with the model's short note), never a free-form model severity/label.
	in := []Check{{ID: "pipe_to_shell", Triggered: true, File: "x.install", Evidence: "curl u | sh", Note: "in post_install"}}
	_, findings, _, _ := deriveVerdict(in)
	if len(findings) != 1 {
		t.Fatalf("want 1 finding, got %d", len(findings))
	}
	if findings[0].Severity != "critical" {
		t.Fatalf("severity must come from the catalog, got %q", findings[0].Severity)
	}
	if want := checkCatalog["pipe_to_shell"].Desc; findings[0].Why[:len(want)] != want {
		t.Fatalf("why must start with the catalog description; got %q", findings[0].Why)
	}
}
