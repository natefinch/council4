package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

const jeskasWillOracle = "Choose one. If you control a commander as you cast this spell, you may choose both instead.\n" +
	"• Add {R} for each card in target opponent's hand.\n" +
	"• Exile the top three cards of your library. You may play them this turn."

func TestLowerJeskasWill(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Jeska's Will",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{R}",
		OracleText: jeskasWillOracle,
	})
	content := face.SpellAbility.Val
	if content.MinModes != 1 || content.MaxModes != 1 ||
		content.ModeChoiceBonus.Condition != game.ModeChoiceConditionControlsCommander ||
		content.ModeChoiceBonus.AdditionalMaxModes != 1 ||
		len(content.Modes) != 2 {
		t.Fatalf("modal content = %+v", content)
	}

	manaMode := content.Modes[0]
	if len(manaMode.Targets) != 1 ||
		manaMode.Targets[0].Selection.Val.Player != game.PlayerOpponent ||
		len(manaMode.Sequence) != 1 {
		t.Fatalf("mana mode = %+v", manaMode)
	}
	add, ok := manaMode.Sequence[0].Primitive.(game.AddMana)
	if !ok || add.ManaColor != mana.R || !add.Amount.IsDynamic() {
		t.Fatalf("mana primitive = %+v", manaMode.Sequence[0].Primitive)
	}
	dynamic := add.Amount.DynamicAmount().Val
	if dynamic.Kind != game.DynamicAmountCountCardsInZone ||
		dynamic.Player == nil ||
		*dynamic.Player != game.TargetPlayerReference(0) ||
		dynamic.CardZone != zone.Hand ||
		dynamic.Multiplier != 1 {
		t.Fatalf("dynamic mana amount = %+v", dynamic)
	}

	impulseMode := content.Modes[1]
	if len(impulseMode.Targets) != 0 || len(impulseMode.Sequence) != 1 {
		t.Fatalf("impulse mode = %+v", impulseMode)
	}
	impulse, ok := impulseMode.Sequence[0].Primitive.(game.ImpulseExile)
	if !ok ||
		impulse.Player != game.ControllerReference() ||
		impulse.Amount.Value() != 3 ||
		impulse.Duration != game.DurationThisTurn {
		t.Fatalf("impulse primitive = %+v", impulseMode.Sequence[0].Primitive)
	}
}

func TestGenerateJeskasWillSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Jeska's Will",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{R}",
		OracleText: jeskasWillOracle,
	}, "k")
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("GenerateExecutableCardSource() err=%v diagnostics=%#v", err, diagnostics)
	}

	for _, want := range []string{
		"ModeChoiceConditionControlsCommander",
		"DynamicAmountCountCardsInZone",
		"TargetPlayerReference(0)",
		"game.ImpulseExile",
		"game.DurationThisTurn",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestLowerImpulseExileOutsideCommanderModal(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Standalone Impulse",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{R}",
		OracleText: "Exile the top three cards of your library. You may play them this turn.",
	})
	content := face.SpellAbility.Val
	if len(content.Modes) != 1 {
		t.Fatalf("content = %+v", content)
	}
	sequence := content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence = %+v", sequence)
	}
	impulse, ok := sequence[0].Primitive.(game.ImpulseExile)
	if !ok || impulse.Amount.Value() != 3 || impulse.Duration != game.DurationThisTurn {
		t.Fatalf("primitive = %+v", sequence[0].Primitive)
	}
}
