package parser

import "testing"

// TestParseChosenColorCountMana covers Three Tree City's third ability "Choose a
// color. Add an amount of mana of that color equal to the number of creatures you
// control of the chosen type.": the add-mana body is recognized as the
// chosen-color dynamic-count form, its amount is a battlefield creature count
// restricted to the source's chosen creature type, the leading "Choose a color."
// sentence is credited onto the add-mana effect span, and no diagnostic is
// raised.
func TestParseChosenColorCountMana(t *testing.T) {
	t.Parallel()
	text := "{2}, {T}: Choose a color. Add an amount of mana of that color equal to the number of creatures you control of the chosen type."
	document, diagnostics := Parse(text, Context{CardName: "Three Tree City"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	ability := document.Abilities[0]
	var manaEffect *EffectSyntax
	for i := range ability.Sentences {
		for j := range ability.Sentences[i].Effects {
			if ability.Sentences[i].Effects[j].Kind == EffectAddMana {
				manaEffect = &ability.Sentences[i].Effects[j]
			}
		}
	}
	if manaEffect == nil {
		t.Fatalf("no add-mana effect: %#v", ability.Sentences)
	}
	if !manaEffect.Mana.ChosenColorDynamic {
		t.Fatalf("ChosenColorDynamic not set: %#v", manaEffect.Mana)
	}
	if !manaEffect.Exact {
		t.Fatal("add-mana effect not exact")
	}
	if manaEffect.HasUnrecognizedSibling {
		t.Fatal("Choose a color sentence left the effect with an unrecognized sibling")
	}
	if manaEffect.Amount.DynamicKind != EffectDynamicAmountCount {
		t.Fatalf("amount kind = %v, want count", manaEffect.Amount.DynamicKind)
	}
	if manaEffect.Amount.Selection == nil || !manaEffect.Amount.Selection.SubtypeFromEntryChoice {
		t.Fatalf("amount selection SubtypeFromEntryChoice not set: %#v", manaEffect.Amount.Selection)
	}
	if manaEffect.Amount.Selection.Kind != SelectionCreature ||
		manaEffect.Amount.Selection.Controller != SelectionControllerYou {
		t.Fatalf("amount selection = %#v, want your creatures", manaEffect.Amount.Selection)
	}
	// The effect span must widen to cover the leading "Choose a color." sentence
	// so the mana ability's coverage scan credits the choice.
	if manaEffect.Span.Start.Offset > 12 {
		t.Fatalf("effect span start = %d, want it widened over \"Choose a color.\"", manaEffect.Span.Start.Offset)
	}
}

// TestParseChosenTypeDynamicCount covers the generic "of the chosen type"
// qualifier on a dynamic count subject outside the mana context, confirming the
// trailing qualifier extends a draw amount and sets the chosen-type predicate.
func TestParseChosenTypeDynamicCount(t *testing.T) {
	t.Parallel()
	text := "Draw cards equal to the number of creatures you control of the chosen type."
	document, diagnostics := Parse(text, Context{CardName: "Probe"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	var drawEffect *EffectSyntax
	for i := range document.Abilities {
		for j := range document.Abilities[i].Sentences {
			for k := range document.Abilities[i].Sentences[j].Effects {
				if document.Abilities[i].Sentences[j].Effects[k].Kind == EffectDraw {
					drawEffect = &document.Abilities[i].Sentences[j].Effects[k]
				}
			}
		}
	}
	if drawEffect == nil {
		t.Fatalf("no draw effect: %#v", document.Abilities)
	}
	if !drawEffect.Exact {
		t.Fatal("draw effect not exact")
	}
	if drawEffect.Amount.Selection == nil || !drawEffect.Amount.Selection.SubtypeFromEntryChoice {
		t.Fatalf("draw amount selection SubtypeFromEntryChoice not set: %#v", drawEffect.Amount.Selection)
	}
}

// TestParseAnyOneColorDynamicMana covers Kami of Whispered Hopes' "{T}: Add X
// mana of any one color, where X is this creature's power.": the add-mana body
// is recognized as the any-one-color dynamic form, its amount is the source's
// power, and the effect is exact. It also covers the "an amount of mana of any
// one color equal to <dynamic>" wording and the devotion amount variant, and
// confirms a plain fixed "any color" body does not gain the dynamic flag.
func TestParseAnyOneColorDynamicMana(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		text     string
		wantKind EffectDynamicAmountKind
	}{
		{"where X is power", "{T}: Add X mana of any one color, where X is this creature's power.", EffectDynamicAmountSourcePower},
		{"equal to power", "{T}: Add an amount of mana of any one color equal to this creature's power.", EffectDynamicAmountSourcePower},
		{"devotion", "{T}: Add X mana of any one color, where X is your devotion to green.", EffectDynamicAmountDevotion},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tc.text, Context{CardName: "Probe"})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			var manaEffect *EffectSyntax
			for i := range document.Abilities {
				for j := range document.Abilities[i].Sentences {
					for k := range document.Abilities[i].Sentences[j].Effects {
						if document.Abilities[i].Sentences[j].Effects[k].Kind == EffectAddMana {
							manaEffect = &document.Abilities[i].Sentences[j].Effects[k]
						}
					}
				}
			}
			if manaEffect == nil {
				t.Fatalf("no add-mana effect: %#v", document.Abilities)
			}
			if !manaEffect.Mana.AnyOneColorDynamic {
				t.Fatalf("AnyOneColorDynamic not set: %#v", manaEffect.Mana)
			}
			if !manaEffect.Exact {
				t.Fatal("add-mana effect not exact")
			}
			if manaEffect.Amount.DynamicKind != tc.wantKind {
				t.Fatalf("amount kind = %v, want %v", manaEffect.Amount.DynamicKind, tc.wantKind)
			}
		})
	}
}

// TestParseAnyOneColorFixedNotDynamic confirms the fixed "Add one mana of any
// color." body keeps its plain AnyColor typing and does not gain the dynamic
// flag, so the dynamic recognizer stays fail-closed without a dynamic amount.
func TestParseAnyOneColorFixedNotDynamic(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("{T}: Add one mana of any color.", Context{CardName: "Probe"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	var manaEffect *EffectSyntax
	for i := range document.Abilities {
		for j := range document.Abilities[i].Sentences {
			for k := range document.Abilities[i].Sentences[j].Effects {
				if document.Abilities[i].Sentences[j].Effects[k].Kind == EffectAddMana {
					manaEffect = &document.Abilities[i].Sentences[j].Effects[k]
				}
			}
		}
	}
	if manaEffect == nil {
		t.Fatalf("no add-mana effect: %#v", document.Abilities)
	}
	if manaEffect.Mana.AnyOneColorDynamic {
		t.Fatal("fixed any-color body wrongly typed as dynamic")
	}
	if !manaEffect.Mana.AnyColor {
		t.Fatalf("AnyColor not set: %#v", manaEffect.Mana)
	}
}
