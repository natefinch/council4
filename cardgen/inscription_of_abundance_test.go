package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

const inscriptionOfAbundanceOracle = "Kicker {2}{G}\n" +
	"Choose one. If this spell was kicked, choose any number instead.\n" +
	"• Put two +1/+1 counters on target creature.\n" +
	"• Target player gains X life, where X is the greatest power among creatures they control.\n" +
	"• Target creature you control fights target creature you don't control."

func TestLowerInscriptionOfAbundance(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Inscription of Abundance",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{G}",
		OracleText: inscriptionOfAbundanceOracle,
	})
	content := face.SpellAbility.Val
	if content.MinModes != 1 || content.MaxModes != 1 ||
		content.ModeChoiceBonus != (game.ModeChoiceBonus{
			Condition:    game.ModeChoiceConditionSpellKicked,
			ReplaceRange: true,
			MinModes:     0,
			MaxModes:     3,
		}) ||
		len(content.Modes) != 3 {
		t.Fatalf("modal content = %#v", content)
	}

	counters, ok := content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
	if !ok || counters.CounterKind != counter.PlusOnePlusOne ||
		counters.Amount.Value() != 2 ||
		counters.Object != game.TargetPermanentReference(0) {
		t.Fatalf("counter mode = %#v", content.Modes[0])
	}
	gain, ok := content.Modes[1].Sequence[0].Primitive.(game.GainLife)
	if !ok || gain.Player != game.TargetPlayerReference(0) {
		t.Fatalf("life mode = %#v", content.Modes[1])
	}
	dynamic := gain.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountGreatestPowerInGroup {
		t.Fatalf("life amount = %#v", gain.Amount)
	}
	anchor, ok := dynamic.Val.Group.PlayerAnchor()
	if !ok || anchor != game.TargetPlayerReference(0) {
		t.Fatalf("life group = %#v", dynamic.Val.Group)
	}
	fight, ok := content.Modes[2].Sequence[0].Primitive.(game.Fight)
	if !ok ||
		fight.Object != game.TargetPermanentReference(0) ||
		fight.RelatedObject != game.TargetPermanentReference(1) ||
		len(content.Modes[2].Targets) != 2 {
		t.Fatalf("fight mode = %#v", content.Modes[2])
	}
}

func TestGenerateInscriptionOfAbundanceSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Inscription of Abundance",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{G}",
		OracleText: inscriptionOfAbundanceOracle,
	}, "i")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.ModeChoiceConditionSpellKicked",
		"ReplaceRange: true",
		"MaxModes: 3",
		"game.DynamicAmountGreatestPowerInGroup",
		"game.TargetPlayerReference(0)",
		"game.Fight{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestKickedModalReplacementRejectsNonSpellAbility(t *testing.T) {
	t.Parallel()
	ctx := contentCtx{
		enclosingKind: compiler.AbilityTriggered,
		content: compiler.AbilityContent{Modes: []compiler.CompiledMode{{
			Modal: &compiler.CompiledModalSemantics{
				MinModes: 1,
				MaxModes: 1,
				Bonus: compiler.CompiledModeChoiceBonus{
					Condition:    compiler.ModeChoiceBonusConditionSpellKicked,
					ReplaceRange: true,
					MinModes:     0,
					MaxModes:     1,
				},
			},
		}}},
	}
	_, diagnostic := lowerModalContent("", ctx, &parser.Ability{Modal: &parser.Modal{}})
	if diagnostic == nil {
		t.Fatal("kicked modal replacement lowered on a triggered ability")
	}
}
