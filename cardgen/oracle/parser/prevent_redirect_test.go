package parser

import "testing"

// TestFoldPreventNextSourceRedirect covers the Deflecting Palm redirect rider
// ("If damage is prevented this way, Deflecting Palm deals that much damage to
// that source's controller."). The fold couples the rider onto the preceding
// one-shot prevent-next-from-source shield as a single annotated prevention
// effect, so the shield carries PreventDamageRedirectToSourceController, the
// redirect sentence's effects are consumed, and the "if damage is prevented this
// way" ability condition is cleared.
func TestFoldPreventNextSourceRedirect(t *testing.T) {
	t.Parallel()
	const source = "The next time a source of your choice would deal damage to you this turn, prevent that damage. If damage is prevented this way, Deflecting Palm deals that much damage to that source's controller."
	document, _ := Parse(source, Context{InstantOrSorcery: true, CardName: "Deflecting Palm"})
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	ability := document.Abilities[0]
	if len(ability.Sentences) != 2 {
		t.Fatalf("sentences = %d, want 2", len(ability.Sentences))
	}
	shield := ability.Sentences[0].Effects
	if len(shield) != 1 || shield[0].Kind != EffectPreventDamage || !shield[0].PreventDamageNextFromSource {
		t.Fatalf("shield effect = %#v, want lone prevent-next-from-source", shield)
	}
	if !shield[0].PreventDamageRedirectToSourceController {
		t.Fatal("PreventDamageRedirectToSourceController = false, want true after fold")
	}
	if shield[0].RequiresOrderedLowering {
		t.Fatal("RequiresOrderedLowering = true, want false after fold")
	}
	if got := ability.Sentences[1].Effects; got != nil {
		t.Fatalf("redirect sentence effects = %#v, want nil after fold", got)
	}
	if got := ability.ConditionSegments; got != nil {
		t.Fatalf("condition segments = %#v, want nil after fold", got)
	}
}

// TestFoldPreventNextSourceRedirectDoesNotFire proves the fold is fail-closed:
// the plain one-shot shield without the redirect rider keeps
// PreventDamageRedirectToSourceController unset, and a shield followed by a
// non-redirect sentence is left untouched.
func TestFoldPreventNextSourceRedirectDoesNotFire(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "plain shield without rider",
			source: "The next time a source of your choice would deal damage to you this turn, prevent that damage.",
		},
		{
			name:   "shield with unrelated second sentence",
			source: "The next time a source of your choice would deal damage to you this turn, prevent that damage. Draw a card.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true, CardName: "Deflecting Palm"})
			if len(document.Abilities) == 0 {
				t.Fatal("abilities = 0, want at least 1")
			}
			shield := document.Abilities[0].Sentences[0].Effects
			if len(shield) != 1 || !shield[0].PreventDamageNextFromSource {
				t.Fatalf("shield effect = %#v, want lone prevent-next-from-source", shield)
			}
			if shield[0].PreventDamageRedirectToSourceController {
				t.Fatal("PreventDamageRedirectToSourceController = true, want false (no rider)")
			}
		})
	}
}
