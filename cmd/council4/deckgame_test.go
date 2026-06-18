package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func testCommanderCard(name string) *game.CardDef {
	pt := game.PT{Value: 2}
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       name,
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Power:      opt.Val(pt),
			Toughness:  opt.Val(pt),
		},
	}
}

func testBasicLand(name string) *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       name,
			Supertypes: []types.Super{types.Basic},
			Types:      []types.Card{types.Land},
		},
	}
}

func testDeckRegistry() *cards.Registry {
	return cards.NewRegistry([]*game.CardDef{
		testCommanderCard("Test Commander"),
		testBasicLand("Forest"),
	})
}

func writeFourDecklists(t *testing.T, content string) []string {
	t.Helper()
	dir := t.TempDir()
	paths := make([]string, 0, game.NumPlayers)
	for i := range game.NumPlayers {
		path := filepath.Join(dir, fmt.Sprintf("deck%d.txt", i+1))
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		paths = append(paths, path)
	}
	return paths
}

func writeValidDecklists(t *testing.T) []string {
	t.Helper()
	return writeFourDecklists(t, "// Commander\n1 Test Commander\n// Deck\n99 Forest\n")
}

func TestRunDeckGameLoadsAndRunsRealDecklists(t *testing.T) {
	paths := writeValidDecklists(t)

	var buf bytes.Buffer
	if err := runDeckGame(&buf, paths, 2, 1, false, false, testDeckRegistry()); err != nil {
		t.Fatalf("runDeckGame: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"Council4 deck game", "(under test)", "Players:", "Turns:"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	// Player 2 is under test, so the marker must sit on that line.
	if !strings.Contains(out, "Player 2: "+paths[1]+" (under test)") {
		t.Errorf("under-test marker not on Player 2 line:\n%s", out)
	}
}

func TestRunDeckGameWrongDeckCount(t *testing.T) {
	err := runDeckGame(&bytes.Buffer{}, []string{"a.txt", "b.txt", "c.txt"}, 1, 1, false, false, testDeckRegistry())
	if err == nil || !strings.Contains(err.Error(), "exactly") {
		t.Fatalf("err = %v, want an 'exactly four' error", err)
	}
}

func TestRunDeckGameInvalidTested(t *testing.T) {
	err := runDeckGame(&bytes.Buffer{}, []string{"a", "b", "c", "d"}, 5, 1, false, false, testDeckRegistry())
	if err == nil || !strings.Contains(err.Error(), "-tested") {
		t.Fatalf("err = %v, want a -tested range error", err)
	}
}

func TestRunDeckGameMissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope.txt")
	paths := []string{missing, missing, missing, missing}
	err := runDeckGame(&bytes.Buffer{}, paths, 1, 1, false, false, testDeckRegistry())
	if err == nil || !strings.Contains(err.Error(), "nope.txt") {
		t.Fatalf("err = %v, want a missing-file error", err)
	}
}

func TestRunDeckGameUnknownCard(t *testing.T) {
	paths := writeFourDecklists(t, "// Commander\n1 Test Commander\n// Deck\n98 Forest\n1 Bogus Card\n")

	err := runDeckGame(&bytes.Buffer{}, paths, 1, 1, false, false, testDeckRegistry())
	if err == nil || !strings.Contains(err.Error(), "Bogus Card") {
		t.Fatalf("err = %v, want an unknown-card error mentioning Bogus Card", err)
	}
}

func TestRunDeckGameIllegalDeck(t *testing.T) {
	// Every card resolves, but 100 main-deck cards violate the 99-card rule, so
	// the load must surface a Commander legality error instead of running.
	paths := writeFourDecklists(t, "// Commander\n1 Test Commander\n// Deck\n100 Forest\n")

	err := runDeckGame(&bytes.Buffer{}, paths, 1, 1, false, false, testDeckRegistry())
	if err == nil || !strings.Contains(err.Error(), "99") {
		t.Fatalf("err = %v, want a deck-size legality error", err)
	}
}

func TestDeckName(t *testing.T) {
	cases := map[string]string{
		"/decks/atraxa.txt": "atraxa",
		"krenko.dec":        "krenko",
		"noext":             "noext",
	}
	for path, want := range cases {
		if got := deckName(path); got != want {
			t.Errorf("deckName(%q) = %q, want %q", path, got, want)
		}
	}
}
