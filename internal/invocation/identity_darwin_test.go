//go:build darwin

package invocation

import "testing"

func TestEnclosingAppBundles(t *testing.T) {
	got := enclosingAppBundles("/Applications/Doubao.app/Contents/Helpers/Doubao Browser.app/Contents/MacOS/Doubao Browser")
	want := []string{"/Applications/Doubao.app", "/Applications/Doubao.app/Contents/Helpers/Doubao Browser.app"}
	if len(got) != len(want) {
		t.Fatalf("enclosingAppBundles() = %#v, want %#v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("enclosingAppBundles()[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}

func TestParseCodesignDetails(t *testing.T) {
	fields := parseCodesignDetails([]byte("Executable=/Applications/ChatGPT.app/Contents/MacOS/ChatGPT\nIdentifier=com.openai.codex\nTeamIdentifier=2DC432GLL2\n"))
	if fields["Identifier"] != "com.openai.codex" || fields["TeamIdentifier"] != "2DC432GLL2" {
		t.Fatalf("parseCodesignDetails() = %#v", fields)
	}
}

func TestPlatformApplicationIdentitiesHandlesEmptyChain(t *testing.T) {
	identities, warnings := platformApplicationIdentities(nil)
	if len(identities) != 0 || len(warnings) != 0 {
		t.Fatalf("identities = %#v, warnings = %#v", identities, warnings)
	}
}
