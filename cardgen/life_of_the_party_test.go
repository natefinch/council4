package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerLifeOfTheParty proves Life of the Party's enter trigger lowers to the
// full typed shape: an intervening "if it's not a token" non-token condition, a
// group-recipient copy-of-source token creation that publishes each created token
// under a link key, and a following rest-of-game goad bound to exactly those
// linked tokens. The card composes three newly supported components — a non-token
// self-ETB condition, an each-opponent copy of the source, and a rest-of-game
// goad of the created tokens — so it must lower without diagnostics.
func TestLowerLifeOfTheParty(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Life of the Party",
		Layout:   "normal",
		ManaCost: "{3}{R}",
		TypeLine: "Creature — Elemental",
		OracleText: "First strike, trample, haste\n" +
			"Whenever this creature attacks, it gets +X/+0 until end of turn, where X is the number of creatures you control.\n" +
			"When this creature enters, if it's not a token, each opponent creates a token that's a copy of it. The tokens are goaded for the rest of the game.",
		Power:     new("2"),
		Toughness: new("2"),
	}
	face := lowerSingleFace(t, card)
	source, diagnostics, err := GenerateExecutableCardSource(card, "l")
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("generated source: err=%v diagnostics=%#v", err, diagnostics)
	}
	for _, want := range []string{
		`game.LinkedObjectsGroup(game.LinkedKey("goad-created-tokens"))`,
		"RestOfGame: true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}

	var enter *game.TriggeredAbility
	for i := range face.TriggeredAbilities {
		if face.TriggeredAbilities[i].Trigger.InterveningCondition.Exists {
			enter = &face.TriggeredAbilities[i]
			break
		}
	}
	if enter == nil {
		t.Fatal("no triggered ability with an intervening condition lowered")
	}

	if !enter.Trigger.InterveningCondition.Exists {
		t.Fatal("enter trigger has no intervening condition")
	}
	if !enter.Trigger.InterveningCondition.Val.ObjectMatches.Exists ||
		!enter.Trigger.InterveningCondition.Val.ObjectMatches.Val.NonToken {
		t.Fatalf("intervening condition is not a non-token match: %#v", enter.Trigger.InterveningCondition.Val)
	}

	if len(enter.Content.Modes) != 1 {
		t.Fatalf("enter trigger modes = %d, want 1", len(enter.Content.Modes))
	}
	seq := enter.Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("enter trigger sequence length = %d, want 2 (create then goad)", len(seq))
	}

	create, ok := seq[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("first instruction is %T, want game.CreateToken", seq[0].Primitive)
	}
	if create.RecipientGroup.Kind == game.PlayerGroupReferenceNone {
		t.Fatal("create token has no recipient group; want each opponent")
	}
	if create.PublishLinked == "" {
		t.Fatal("create token does not publish its created tokens under a link key")
	}
	if _, isCopy := create.Source.TokenCopy(); !isCopy {
		t.Fatal("create token source is not a copy of the source permanent")
	}

	goad, ok := seq[1].Primitive.(game.Goad)
	if !ok {
		t.Fatalf("second instruction is %T, want game.Goad", seq[1].Primitive)
	}
	if !goad.RestOfGame {
		t.Fatal("goad is not rest-of-game")
	}
	key, linked := goad.Group.LinkedKey()
	if !linked {
		t.Fatal("goad does not target a linked-objects group")
	}
	if key != create.PublishLinked {
		t.Fatalf("goad linked key %q does not match create publish key %q", key, create.PublishLinked)
	}
}
