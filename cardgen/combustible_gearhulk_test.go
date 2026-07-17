package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func combustibleGearhulkCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Combustible Gearhulk",
		Layout:     "normal",
		ManaCost:   "{4}{R}{R}",
		TypeLine:   "Artifact Creature — Construct",
		Power:      new("6"),
		Toughness:  new("6"),
		OracleText: "First strike\nWhen this creature enters, target opponent may have you draw three cards. If the player doesn't, you mill three cards, then this creature deals damage to that player equal to the total mana value of those cards.",
	}
}

func TestLowerCombustibleGearhulk(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, combustibleGearhulkCard())
	if len(face.StaticAbilities) == 0 || face.StaticAbilities[0].VarName != "game.FirstStrikeStaticBody" {
		t.Fatal("first strike was not lowered")
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 3 {
		t.Fatalf("mode = %#v", mode)
	}
	draw, ok := mode.Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Player != game.ControllerReference() ||
		!mode.Sequence[0].Optional ||
		!mode.Sequence[0].OptionalActor.Exists ||
		mode.Sequence[0].OptionalActor.Val != game.TargetPlayerReference(0) {
		t.Fatalf("draw offer = %#v", mode.Sequence[0])
	}
	mill, ok := mode.Sequence[1].Primitive.(game.Mill)
	if !ok || mill.Player != game.ControllerReference() || mill.PublishLinked == "" {
		t.Fatalf("mill = %#v", mode.Sequence[1])
	}
	damage, ok := mode.Sequence[2].Primitive.(game.Damage)
	recipient, recipientOK := damage.Recipient.PlayerReference()
	if !ok || !recipientOK || recipient != game.TargetPlayerReference(0) ||
		!damage.DamageSource.Exists ||
		damage.DamageSource.Val != game.SourcePermanentReference() {
		t.Fatalf("damage = %#v", mode.Sequence[2])
	}
	dynamic := damage.Amount.DynamicAmount()
	if !dynamic.Exists ||
		dynamic.Val.Kind != game.DynamicAmountReferencedCardsTotalManaValue ||
		dynamic.Val.LinkedKey != mill.PublishLinked {
		t.Fatalf("damage amount = %#v, mill link = %q", damage.Amount, mill.PublishLinked)
	}
	for _, instruction := range mode.Sequence[1:] {
		if !instruction.ResultGate.Exists ||
			instruction.ResultGate.Val.Accepted != game.TriFalse {
			t.Fatalf("decline branch gate = %#v", instruction.ResultGate)
		}
	}
}

func TestGenerateExecutableCombustibleGearhulk(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(combustibleGearhulkCard(), "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.FirstStrike",
		"OptionalActor: opt.Val(game.TargetPlayerReference(0))",
		`PublishLinked: game.LinkedKey("may-have-milled-cards")`,
		"game.DynamicAmountReferencedCardsTotalManaValue",
		"game.PlayerDamageRecipient(game.TargetPlayerReference(0))",
		"opt.Val(game.SourcePermanentReference())",
		"game.TriFalse",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	if got := strings.Count(source, `game.LinkedKey("may-have-milled-cards")`); got != 2 {
		t.Fatalf("linked mill key occurrences = %d, want 2:\n%s", got, source)
	}
}
