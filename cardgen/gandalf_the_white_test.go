package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func gandalfTheWhiteCard() *ScryfallCard {
	return &ScryfallCard{
		Name:          "Gandalf the White",
		Layout:        "normal",
		ManaCost:      "{3}{W}{W}",
		TypeLine:      "Legendary Creature — Avatar Wizard",
		OracleText:    "Flash\nYou may cast legendary spells and artifact spells as though they had flash.\nIf a legendary permanent or an artifact entering or leaving the battlefield causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
		Colors:        []string{"W"},
		ColorIdentity: []string{"W"},
		Power:         new("4"),
		Toughness:     new("5"),
	}
}

func TestLowerGandalfTheWhite(t *testing.T) {
	face := lowerSingleFace(t, gandalfTheWhiteCard())
	if len(face.StaticAbilities) != 3 {
		t.Fatalf("static abilities = %#v, want flash plus two rule statics", face.StaticAbilities)
	}
	var flash, multiplier *game.RuleEffect
	for i := range face.StaticAbilities {
		for j := range face.StaticAbilities[i].Body.RuleEffects {
			effect := &face.StaticAbilities[i].Body.RuleEffects[j]
			switch effect.Kind {
			case game.RuleEffectCastSpellsAsThoughFlash:
				flash = effect
			case game.RuleEffectAdditionalTriggerForControlledPermanent:
				multiplier = effect
			default:
			}
		}
	}
	if flash == nil || len(flash.SpellCharacteristicFilters) != 2 {
		t.Fatalf("flash effect = %#v", flash)
	}
	if len(flash.SpellCharacteristicFilters[0].Supertypes) != 1 ||
		flash.SpellCharacteristicFilters[0].Supertypes[0] != types.Legendary ||
		len(flash.SpellCharacteristicFilters[1].Types) != 1 ||
		flash.SpellCharacteristicFilters[1].Types[0] != types.Artifact {
		t.Fatalf("flash filters = %#v", flash.SpellCharacteristicFilters)
	}
	if multiplier == nil ||
		!multiplier.TriggerCausePermanentEnters ||
		!multiplier.TriggerCausePermanentLeaves ||
		len(multiplier.TriggerCausePermanentFilters) != 2 {
		t.Fatalf("multiplier = %#v", multiplier)
	}
}

func TestGenerateExecutableGandalfTheWhite(t *testing.T) {
	source, diagnostics, err := GenerateExecutableCardSource(gandalfTheWhiteCard(), "g")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.FlashStaticBody",
		"game.RuleEffectCastSpellsAsThoughFlash",
		"SpellCharacteristicFilters",
		"game.RuleEffectAdditionalTriggerForControlledPermanent",
		"TriggerCausePermanentEnters: true",
		"TriggerCausePermanentLeaves: true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
