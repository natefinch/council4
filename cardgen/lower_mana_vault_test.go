package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
)

const manaVaultOracleText = "This artifact doesn't untap during your untap step.\n" +
	"At the beginning of your upkeep, you may pay {4}. If you do, untap this artifact.\n" +
	"At the beginning of your draw step, if this artifact is tapped, it deals 1 damage to you.\n" +
	"{T}: Add {C}{C}{C}."

func TestLowerManaVaultEndToEnd(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Mana Vault",
		Layout:     "normal",
		ManaCost:   "{1}",
		TypeLine:   "Artifact",
		OracleText: manaVaultOracleText,
	}
	face := lowerSingleFace(t, card)
	if len(face.StaticAbilities) != 1 ||
		len(face.StaticAbilities[0].Body.RuleEffects) != 1 ||
		face.StaticAbilities[0].Body.RuleEffects[0].Kind != game.RuleEffectDoesntUntap ||
		!face.StaticAbilities[0].Body.RuleEffects[0].AffectedSource {
		t.Fatalf("static abilities = %#v", face.StaticAbilities)
	}
	if len(face.ManaAbilities) != 1 ||
		len(face.ManaAbilities[0].Content.Modes) != 1 ||
		len(face.ManaAbilities[0].Content.Modes[0].Sequence) != 3 {
		t.Fatalf("mana abilities = %#v", face.ManaAbilities)
	}
	for _, instruction := range face.ManaAbilities[0].Content.Modes[0].Sequence {
		add, ok := instruction.Primitive.(game.AddMana)
		if !ok || add.Amount != game.Fixed(1) || add.ManaColor != mana.C {
			t.Fatalf("mana instruction = %#v", instruction)
		}
	}
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %#v", face.TriggeredAbilities)
	}
	upkeep := face.TriggeredAbilities[0]
	if upkeep.Optional ||
		upkeep.Trigger.Pattern.Step != game.StepUpkeep ||
		len(upkeep.Content.Modes) != 1 ||
		len(upkeep.Content.Modes[0].Sequence) != 2 {
		t.Fatalf("upkeep trigger = %#v", upkeep)
	}
	pay, ok := upkeep.Content.Modes[0].Sequence[0].Primitive.(game.Pay)
	if !ok || !pay.Payment.ManaCost.Exists ||
		pay.Payment.ManaCost.Val.String() != (cost.Mana{cost.O(4)}).String() {
		t.Fatalf("payment instruction = %#v", upkeep.Content.Modes[0].Sequence[0])
	}
	untap, ok := upkeep.Content.Modes[0].Sequence[1].Primitive.(game.Untap)
	if !ok || untap.Object != game.SourcePermanentReference() ||
		!upkeep.Content.Modes[0].Sequence[1].ResultGate.Exists ||
		upkeep.Content.Modes[0].Sequence[1].ResultGate.Val.Succeeded != game.TriTrue {
		t.Fatalf("untap instruction = %#v", upkeep.Content.Modes[0].Sequence[1])
	}
	draw := face.TriggeredAbilities[1]
	if draw.Trigger.Pattern.Step != game.StepDraw ||
		!draw.Trigger.InterveningCondition.Exists ||
		!draw.Trigger.InterveningCondition.Val.ObjectMatches.Exists ||
		draw.Trigger.InterveningCondition.Val.ObjectMatches.Val.Tapped != game.TriTrue {
		t.Fatalf("draw trigger = %#v", draw)
	}
	damage, ok := draw.Content.Modes[0].Sequence[0].Primitive.(game.Damage)
	if !ok || damage.Amount != game.Fixed(1) ||
		damage.Recipient != game.PlayerDamageRecipient(game.ControllerReference()) ||
		!damage.DamageSource.Exists ||
		damage.DamageSource.Val != game.SourcePermanentReference() {
		t.Fatalf("damage instruction = %#v", draw.Content.Modes[0].Sequence[0])
	}
}

func TestGenerateManaVaultExecutableSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Mana Vault",
		Layout:     "normal",
		ManaCost:   "{1}",
		TypeLine:   "Artifact",
		OracleText: manaVaultOracleText,
	}, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.RuleEffectDoesntUntap",
		"game.Pay",
		"cost.O(4)",
		"game.SourcePermanentReference()",
		"game.StepDraw",
		"game.TriTrue",
		"DamageSource: opt.Val(game.SourcePermanentReference())",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
