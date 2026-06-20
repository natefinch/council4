package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const otawaraOracleText = "{T}: Add {U}.\nChannel — {3}{U}, Discard this card: Return target artifact, creature, enchantment, or planeswalker to its owner's hand. This ability costs {1} less to activate for each legendary creature you control."

func TestGenerateOtawaraSoaringCity(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Otawara, Soaring City",
		Layout:     "normal",
		TypeLine:   "Legendary Land",
		OracleText: otawaraOracleText,
	}
	face := lowerSingleFace(t, card)
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if ability.ZoneOfFunction != zone.Hand ||
		!ability.ManaCost.Exists ||
		!slices.Equal(ability.ManaCost.Val, cost.Mana{cost.O(3), cost.U}) {
		t.Fatalf("zone/cost = %v/%v, want hand/{3}{U}", ability.ZoneOfFunction, ability.ManaCost)
	}
	if len(ability.AdditionalCosts) != 1 ||
		ability.AdditionalCosts[0].Kind != cost.AdditionalDiscard ||
		ability.AdditionalCosts[0].Source != zone.Hand ||
		ability.AdditionalCosts[0].Text != "Discard this card" {
		t.Fatalf("additional costs = %#v, want discard source from hand", ability.AdditionalCosts)
	}
	if len(ability.CostModifiers) != 1 {
		t.Fatalf("cost modifiers = %#v, want one", ability.CostModifiers)
	}
	modifier := ability.CostModifiers[0]
	if modifier.Kind != game.CostModifierAbility || modifier.PerObjectReduction != 1 ||
		!slices.Equal(modifier.CountSelection.RequiredTypes, []types.Card{types.Creature}) ||
		!slices.Equal(modifier.CountSelection.Supertypes, []types.Super{types.Legendary}) ||
		modifier.CountSelection.Controller != game.ControllerYou {
		t.Fatalf("cost modifier = %#v", modifier)
	}
	targets := game.BodyTargets(&ability)
	if len(targets) != 1 ||
		!slices.Equal(targets[0].Predicate.PermanentTypes, []types.Card{
			types.Artifact, types.Creature, types.Enchantment, types.Planeswalker,
		}) {
		t.Fatalf("targets = %#v", targets)
	}
	content := game.BodyContent(&ability)
	if len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 1 {
		t.Fatalf("content = %#v, want one instruction", content)
	}
	bounce, ok := content.Modes[0].Sequence[0].Primitive.(game.Bounce)
	if !ok || bounce.Object != game.TargetPermanentReference(0) {
		t.Fatalf("primitive = %#v, want target permanent bounce", content.Modes[0].Sequence[0].Primitive)
	}
}

func TestGenerateOtawaraSourceRendersTypedChannelData(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Otawara, Soaring City",
		Layout:     "normal",
		TypeLine:   "Legendary Land",
		OracleText: otawaraOracleText,
	}, "o")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"ZoneOfFunction: zone.Hand",
		`"Discard this card"`,
		"PerObjectReduction: 1",
		"Supertypes: []types.Super{types.Legendary}",
		"PermanentTypes: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Planeswalker}",
		"Primitive: game.Bounce",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateChannelVariantsFailClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
	}{
		{
			name: "discard another card",
			text: "Channel — {3}{U}, Discard a card: Return target artifact, creature, enchantment, or planeswalker to its owner's hand. This ability costs {1} less to activate for each legendary creature you control.",
		},
		{
			name: "different destination",
			text: "Channel — {3}{U}, Discard this card: Return target artifact, creature, enchantment, or planeswalker to your hand. This ability costs {1} less to activate for each legendary creature you control.",
		},
		{
			name: "conjunctive union",
			text: "Channel — {3}{U}, Discard this card: Return target artifact, creature, enchantment, and planeswalker to its owner's hand. This ability costs {1} less to activate for each legendary creature you control.",
		},
		{
			name: "additional activation cost",
			text: "Channel — {3}{U}, Pay 1 life, Discard this card: Return target artifact, creature, enchantment, or planeswalker to its owner's hand. This ability costs {1} less to activate for each legendary creature you control.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Channel",
				Layout:     "normal",
				TypeLine:   "Legendary Land",
				OracleText: test.text,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("source %q unexpectedly generated", test.text)
			}
		})
	}
}

func TestGenerateActivationCostReductionPreservesUnsupportedMainSentenceContent(t *testing.T) {
	t.Parallel()
	tests := []string{
		"{1}: Draw a card, then you become the monarch. This ability costs {1} less to activate for each legendary creature you control.",
		"{1}: Draw a card, then venture into the dungeon. This ability costs {1} less to activate for each legendary creature you control.",
	}
	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: text,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("source %q unexpectedly generated", text)
			}
		})
	}
}
