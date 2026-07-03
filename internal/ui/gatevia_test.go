package ui

import (
	"strings"
	"testing"

	"github.com/manticore-projects/aurscan/internal/scan"
)

func mal() []scan.Result {
	return []scan.Result{{Pkg: "evil", V: scan.Verdict{Verdict: "MALICIOUS", Confidence: 95, Summary: "bad"}}}
}

func TestGateViaRequiresINSTALL(t *testing.T) {
	var out strings.Builder
	// INSTALL proceeds directly, no "c" prefix needed
	if !GateVia(mal(), strings.NewReader("INSTALL\n"), &out, false) {
		t.Fatal("INSTALL should proceed")
	}
	out.Reset()
	// case-insensitive: lowercase install also proceeds
	if !GateVia(mal(), strings.NewReader("install\n"), &out, false) {
		t.Fatal("lowercase install should proceed")
	}
	out.Reset()
	// mixed case also proceeds
	if !GateVia(mal(), strings.NewReader("InStAlL\n"), &out, false) {
		t.Fatal("mixed-case INSTALL should proceed")
	}
}

func TestGateViaQuitExits(t *testing.T) {
	var out strings.Builder
	// q exits (only way out besides INSTALL)
	if GateVia(mal(), strings.NewReader("q\n"), &out, false) {
		t.Fatal("q should exit (return false)")
	}
	out.Reset()
	// quit exits, case-insensitive
	if GateVia(mal(), strings.NewReader("QUIT\n"), &out, false) {
		t.Fatal("QUIT should exit (return false)")
	}
}

func TestGateViaMisspellReprompts(t *testing.T) {
	var out strings.Builder
	// a typo re-prompts; a following INSTALL still proceeds (state is not lost)
	if !GateVia(mal(), strings.NewReader("INSTAL\nINSTALL\n"), &out, false) {
		t.Fatal("mis-spelled INSTALL should re-prompt, then INSTALL should proceed")
	}
}

func TestGateViaEmptyReprompts(t *testing.T) {
	var out strings.Builder
	// empty input re-prompts (does not abort); a following INSTALL proceeds
	if !GateVia(mal(), strings.NewReader("\nINSTALL\n"), &out, false) {
		t.Fatal("empty input should re-prompt, not abort; INSTALL should then proceed")
	}
	out.Reset()
	// empty input then q exits
	if GateVia(mal(), strings.NewReader("\nq\n"), &out, false) {
		t.Fatal("empty input should re-prompt; q should then exit")
	}
}

func TestGateViaWrongWordReprompts(t *testing.T) {
	var out strings.Builder
	// a wrong (non-quit) word re-prompts; a following INSTALL proceeds
	if !GateVia(mal(), strings.NewReader("nope\nINSTALL\n"), &out, false) {
		t.Fatal("wrong word should re-prompt, not abort; INSTALL should then proceed")
	}
	out.Reset()
	// a wrong word then q exits
	if GateVia(mal(), strings.NewReader("nope\nq\n"), &out, false) {
		t.Fatal("wrong word should re-prompt; q should then exit")
	}
}

func TestGateViaOKProceeds(t *testing.T) {
	ok := []scan.Result{{Pkg: "good", V: scan.Verdict{Verdict: "OK", Confidence: 90}}}
	var out strings.Builder
	if !GateVia(ok, strings.NewReader(""), &out, false) {
		t.Fatal("OK should proceed without prompting")
	}
}
