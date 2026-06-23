package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceBiteOfTheBlackRose exercises the voting
// construct (CR 701.32 "will of the council"): "Starting with you, each player
// votes for sickness or psychosis." lowers to a game.Vote primitive publishing
// the signed vote margin, with the two named arms gated on the margin's sign.
// The first option's arm gates on a positive margin and the tie-inclusive
// second option's arm gates on a non-positive margin.
func TestGenerateExecutableCardSourceBiteOfTheBlackRose(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Bite of the Black Rose",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{3}{B}",
		OracleText: "Will of the council — Starting with you, each player votes for sickness or psychosis. If sickness gets more votes, creatures your opponents control get -2/-2 until end of turn. If psychosis gets more votes or the vote is tied, each opponent discards two cards.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.Vote{",
		`"sickness"`,
		`"psychosis"`,
		`PublishResult: game.ResultKey("vote-result")`,
		`Key:         "vote-result"`,
		"AmountRange: opt.Val(game.IntRange{Min: 1, Max: 4})",
		"AmountRange: opt.Val(game.IntRange{Min: -4, Max: 0})",
	} {
		if !strings.Contains(normalizeSource(source), normalizeSource(wanted)) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
