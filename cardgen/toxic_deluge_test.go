package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

const toxicDelugeText = "As an additional cost to cast this spell, pay X life.\nAll creatures get -X/-X until end of turn."

func toxicDelugeCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Toxic Deluge",
		Layout:     "normal",
		ManaCost:   "{2}{B}",
		TypeLine:   "Sorcery",
		OracleText: toxicDelugeText,
		Colors:     []string{"B"},
	}
}

func TestLowerToxicDelugeLinkedLifeX(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, toxicDelugeCard())
	if len(face.AdditionalCosts) != 1 {
		t.Fatalf("additional costs = %#v, want one", face.AdditionalCosts)
	}
	additional := face.AdditionalCosts[0]
	if additional.Kind != cost.AdditionalPayLife || !additional.AmountFromX || additional.Amount != 0 {
		t.Fatalf("additional cost = %#v, want pay X life", additional)
	}
	mode := face.SpellAbility.Val.Modes[0]
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	continuous := apply.ContinuousEffects[0]
	selection := continuous.Group.Selection()
	if continuous.Layer != game.LayerPowerToughnessModify ||
		!continuous.PowerDeltaDynamic.Exists ||
		continuous.PowerDeltaDynamic.Val.Kind != game.DynamicAmountX ||
		continuous.PowerDeltaDynamic.Val.Multiplier != -1 ||
		!continuous.ToughnessDeltaDynamic.Exists ||
		continuous.ToughnessDeltaDynamic.Val.Kind != game.DynamicAmountX ||
		continuous.ToughnessDeltaDynamic.Val.Multiplier != -1 ||
		selection.Controller != game.ControllerAny ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("continuous effect = %#v, want all creatures -X/-X", continuous)
	}
	if issues := game.ValidateCardDef(&game.CardDef{CardFace: game.CardFace{
		Name:            "Toxic Deluge",
		Types:           []types.Card{types.Sorcery},
		AdditionalCosts: face.AdditionalCosts,
		SpellAbility:    face.SpellAbility,
	}}); len(issues) != 0 {
		t.Fatalf("validation issues = %#v", issues)
	}
}

func TestRenderToxicDelugeLinkedLifeX(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(toxicDelugeCard(), "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Kind:        cost.AdditionalPayLife",
		"AmountFromX: true",
		"PowerDeltaDynamic: opt.Val(game.DynamicAmount{",
		"ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{",
		"game.DynamicAmountX",
		"Multiplier: -1",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestPayXLifeCostComposesWithExistingTargetXEffect(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Hatred",
		Layout:     "normal",
		ManaCost:   "{3}{B}{B}",
		TypeLine:   "Instant",
		OracleText: "As an additional cost to cast this spell, pay X life.\nTarget creature gets +X/+0 until end of turn.",
		Colors:     []string{"B"},
	})
	if len(face.AdditionalCosts) != 1 || !face.AdditionalCosts[0].AmountFromX {
		t.Fatalf("additional costs = %#v, want pay X life", face.AdditionalCosts)
	}
}

func TestToxicDelugeFamilyFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		manaCost   string
		oracleText string
	}{
		{
			name:       "mana X",
			manaCost:   "{X}{B}",
			oracleText: toxicDelugeText,
		},
		{
			name:       "sacrifice X cost",
			manaCost:   "{2}{B}",
			oracleText: "As an additional cost to cast this spell, sacrifice X creatures.\nAll creatures get -X/-X until end of turn.",
		},
		{
			name:       "discard X cost",
			manaCost:   "{2}{B}",
			oracleText: "As an additional cost to cast this spell, discard X cards.\nAll creatures get -X/-X until end of turn.",
		},
		{
			name:       "controlled group",
			manaCost:   "{2}{B}",
			oracleText: "As an additional cost to cast this spell, pay X life.\nCreatures you control get -X/-X until end of turn.",
		},
		{
			name:       "opponent group",
			manaCost:   "{2}{B}",
			oracleText: "As an additional cost to cast this spell, pay X life.\nCreatures your opponents control get -X/-X until end of turn.",
		},
		{
			name:       "permanent",
			manaCost:   "{2}{B}",
			oracleText: "As an additional cost to cast this spell, pay X life.\nAll creatures get -X/-X.",
		},
		{
			name:       "asymmetric",
			manaCost:   "{2}{B}",
			oracleText: "As an additional cost to cast this spell, pay X life.\nAll creatures get -X/+X until end of turn.",
		},
		{
			name:       "different duration",
			manaCost:   "{2}{B}",
			oracleText: "As an additional cost to cast this spell, pay X life.\nAll creatures get -X/-X this turn.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Reject",
				Layout:     "normal",
				ManaCost:   test.manaCost,
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
				Colors:     []string{"B"},
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("expected fail-closed diagnostic for %q", test.oracleText)
			}
		})
	}
}
