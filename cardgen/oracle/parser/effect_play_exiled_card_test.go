package parser

import "testing"

func TestParsePlayHideawayCardEffect(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"You may play the exiled card without paying its mana cost if you control three or more creatures.",
		"Then if you control ten or more creatures, you may play the exiled card without paying its mana cost.",
	} {
		document, diagnostics := Parse(source, Context{})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", source, diagnostics)
		}
		if len(document.Abilities) != 1 ||
			len(document.Abilities[0].Sentences) != 1 ||
			len(document.Abilities[0].Sentences[0].Effects) != 1 {
			t.Fatalf("%q document = %#v", source, document)
		}
		effect := document.Abilities[0].Sentences[0].Effects[0]
		if effect.Kind != EffectPlay ||
			!effect.PlayHideawayExiledCard ||
			!effect.CastWithoutPayingManaCost ||
			!effect.Optional ||
			!effect.Exact {
			t.Fatalf("%q effect = %#v", source, effect)
		}
	}
}
