package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/mana"
)

func TestParseCardsNamedSelfInGraveyardsManaTypedSyntax(t *testing.T) {
	t.Parallel()
	const name = "Rite of Flame"
	document, diagnostics := Parse(
		"Add {R} for each card named Rite of Flame in each graveyard.",
		Context{CardName: name},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v", effects)
	}
	effect := effects[0]
	if effect.Kind != EffectAddMana || !effect.Exact ||
		effect.Amount.DynamicKind != EffectDynamicAmountCardsNamedSelfInGraveyards ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
		effect.Amount.Multiplier != 1 {
		t.Fatalf("amount = %#v", effect.Amount)
	}
	if !effect.Mana.ColorsKnown || len(effect.Mana.Colors) != 1 ||
		effect.Mana.Colors[0] != mana.R || effect.Mana.Choice || effect.Mana.AnyColor {
		t.Fatalf("mana = %#v", effect.Mana)
	}
}

func TestParseCardsNamedSelfInGraveyardsManaFailsClosed(t *testing.T) {
	t.Parallel()
	// Without the matching self-name in Context, or with a multi-symbol output,
	// the recognizer must not type the dynamic count nor a single produced color.
	cases := []struct {
		name    string
		text    string
		context Context
	}{
		{
			name:    "name not the card's own",
			text:    "Add {R} for each card named Lightning Bolt in each graveyard.",
			context: Context{CardName: "Rite of Flame"},
		},
		{
			name:    "multi-symbol output",
			text:    "Add {R}{R} for each card named Rite of Flame in each graveyard.",
			context: Context{CardName: "Rite of Flame"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(tc.text, tc.context)
			if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
				return
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				return
			}
			effect := effects[0]
			if effect.Amount.DynamicKind == EffectDynamicAmountCardsNamedSelfInGraveyards &&
				effect.Mana.ColorsKnown && len(effect.Mana.Colors) == 1 {
				t.Fatalf("variant unexpectedly recognized:\n%s", tc.text)
			}
		})
	}
}
