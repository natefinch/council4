package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerTargetPlayerGraveyardShuffle lowers "Target player shuffles their
// graveyard into their library." (Reminisce) to a ShuffleGraveyardIntoLibrary
// primitive naming the targeted player, with a single player target.
func TestLowerTargetPlayerGraveyardShuffle(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Reminisce",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target player shuffles their graveyard into their library.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Allow != game.TargetAllowPlayer {
		t.Fatalf("targets = %+v, want one player target", mode.Targets)
	}
	shuffle, ok := mode.Sequence[0].Primitive.(game.ShuffleGraveyardIntoLibrary)
	if !ok {
		t.Fatalf("primitive = %T, want game.ShuffleGraveyardIntoLibrary", mode.Sequence[0].Primitive)
	}
	if shuffle.Player != game.TargetPlayerReference(0) {
		t.Fatalf("shuffle player = %#v, want target player 0", shuffle.Player)
	}
}

// TestLowerTargetPlayerHandShuffleRejected keeps a target-player shuffle whose
// source zone is not the graveyard fail-closed, so the graveyard-into-library
// recognizer does not over-match other zone shuffles.
func TestLowerTargetPlayerHandShuffleRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Hand Shuffle",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target player shuffles their hand into their library.",
		Games:      []string{"paper"},
		Legalities: map[string]string{"legacy": "legal"},
	})
	if len(diagnostics) == 0 {
		t.Fatalf("diagnostics = %+v, want a non-graveyard shuffle to fail closed", diagnostics)
	}
}
