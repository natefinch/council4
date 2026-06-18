package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestParseTemporaryKeywordSubjectExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		{"This creature gains flying until end of turn.", true},
		{"This creature gains trample and haste until end of turn.", true},
		{"Target creature gains flying until end of turn.", true},
		// Unknown keyword stays fail-closed.
		{"This creature gains banding until end of turn.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].Exact != test.exact {
				t.Fatalf("effect Exact = %v, want %v", effects[0].Exact, test.exact)
			}
		})
	}
}

func TestParseExcludedColorTypeTargetExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		{"Destroy target nonblack creature.", true},
		{"Destroy target nonwhite permanent.", true},
		{"Destroy target noncreature artifact.", true},
		{"Destroy target nonartifact creature.", true},
		{"Destroy target nonwhite creature you control.", true},
		{"Destroy target creature.", true},
		// Two excluded colors are not reconstructed and must stay fail-closed.
		{"Destroy target nonblack nonred creature.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || len(effects[0].Targets) != 1 {
				t.Fatalf("effects = %#v, want one effect with one target", effects)
			}
			if effects[0].Targets[0].Exact != test.exact {
				t.Fatalf("target Exact = %v, want %v", effects[0].Targets[0].Exact, test.exact)
			}
		})
	}
}

func TestParseColorSpellTargetExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		{"Counter target blue spell.", true},
		{"Counter target nonblue spell.", true},
		{"Counter target colorless spell.", true},
		{"Counter target multicolored spell.", true},
		// Monocolored spells have no canonical predicate yet and stay fail-closed.
		{"Counter target monocolored spell.", false},
		// A color combined with a card-type filter is not reconstructed.
		{"Counter target blue creature spell.", false},
		// Two colors are not reconstructed and must stay fail-closed.
		{"Counter target blue and white spell.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || len(effects[0].Targets) != 1 {
				t.Fatalf("effects = %#v, want one effect with one target", effects)
			}
			if effects[0].Targets[0].Exact != test.exact {
				t.Fatalf("target Exact = %v, want %v", effects[0].Targets[0].Exact, test.exact)
			}
		})
	}
}

func TestParseResolvingEffectKinds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		kind   EffectKind
	}{
		{"Add {G}.", EffectAddMana},
		{"Attach target Equipment to target creature.", EffectAttach},
		{"Cast that card.", EffectCast},
		{"Counter target spell.", EffectCounter},
		{"Create a token.", EffectCreate},
		{"Deal 2 damage to any target.", EffectDealDamage},
		{"Destroy target creature.", EffectDestroy},
		{"Discard a card.", EffectDiscard},
		{"Discover 3.", EffectDiscover},
		{"Double its power.", EffectDouble},
		{"Draw a card.", EffectDraw},
		{"This land enters tapped.", EffectEnterTapped},
		{"This creature enters prepared.", EffectEnterPrepared},
		{"Exile target creature.", EffectExile},
		{"Target creature fights target creature.", EffectFight},
		{"Gain 2 life.", EffectGain},
		{"Gain control of target creature.", EffectGainControl},
		{"Target creature has flying.", EffectGrantKeyword},
		{"Investigate.", EffectInvestigate},
		{"Target creature explores.", EffectExplore},
		{"Lose 2 life.", EffectLose},
		{"Manifest the top card of your library.", EffectManifest},
		{"Manifest dread.", EffectManifestDread},
		{"Look at the top two cards of your library.", EffectManifestDread},
		{"Mill two cards.", EffectMill},
		{"Target creature gets +2/+2.", EffectModifyPT},
		{"Put a +1/+1 counter on target creature.", EffectPut},
		{"Proliferate.", EffectProliferate},
		{"Regenerate target creature.", EffectRegenerate},
		{"Return target creature to its owner's hand.", EffectReturn},
		{"Reveal that card.", EffectReveal},
		{"Sacrifice a creature.", EffectSacrifice},
		{"Scry 2.", EffectScry},
		{"Surveil 2.", EffectSurveil},
		{"Search your library for a card.", EffectSearch},
		{"Shuffle your library.", EffectShuffle},
		{"Tap target creature.", EffectTap},
		{"Untap target creature.", EffectUntap},
		{"Transform target creature.", EffectTransform},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) == 0 || effects[0].Kind != test.kind {
				t.Fatalf("effects = %#v, want first kind %v", effects, test.kind)
			}
		})
	}
}

func TestParseResolvingEffectExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		{"Scry 2.", true},
		{"Scry 2, then celebrate.", false},
		{"Surveil two.", true},
		{"Surveil two, then celebrate.", false},
		{"Investigate.", true},
		{"Investigate twice.", true},
		{"Investigate, then celebrate.", false},
		{"Proliferate.", true},
		{"Proliferate two times.", true},
		{"Proliferate, then celebrate.", false},
		{"Creatures you control get +2/+2 until end of turn.", true},
		{"Creatures you control get +2/+2 until end of turn, then celebrate.", false},
		{"This creature gets +2/+0 until end of turn.", true},
		{"This creature gets +1/+1 until end of turn, then celebrate.", false},
		{"Put a +1/+1 counter on this creature.", true},
		{"Put a +1/+1 counter on this creature, then celebrate.", false},
		{"Gain control of target creature.", true},
		{"Gain control of target creature until end of turn.", true},
		{"Gain control of target creature for as long as you control this creature.", true},
		{"Gain control of target creature until end of turn, then celebrate.", false},
		{"Sacrifice a creature.", true},
		{"You sacrifice a creature.", true},
		{"Sacrifice two permanents.", true},
		{"Each opponent sacrifices a creature.", true},
		{"Sacrifice a creature, then celebrate.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || effects[0].Exact != test.exact {
				t.Fatalf("effects = %#v, want one effect with Exact=%v", effects, test.exact)
			}
		})
	}
}

func TestParseCreateCopyOfTargetToken(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		copy   bool
	}{
		{"Create a token that's a copy of target creature you control.", true},
		{"Create a token that's a copy of target artifact.", true},
		{"Create a 1/1 white Soldier creature token.", false},
		{"Create a token that's a copy of target creature you control, then celebrate.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].TokenCopyOfTarget != test.copy {
				t.Fatalf("TokenCopyOfTarget = %v, want %v", effects[0].TokenCopyOfTarget, test.copy)
			}
			if test.copy && !effects[0].Exact {
				t.Fatalf("copy token effect should be exact: %#v", effects[0])
			}
		})
	}
}

func TestParseGainControlSequenceExactness(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Untap target creature and gain control of it until end of turn. That creature gains haste until end of turn.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 2 || !effects[0].Exact || !effects[1].Exact {
		t.Fatalf("effects = %#v, want two exact effects", effects)
	}
}

func TestParseGainControlFollowOnExactness(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Gain control of target creature until end of turn. Untap that creature. It gains haste until end of turn.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, sentence := range document.Abilities[0].Sentences {
		for _, effect := range sentence.Effects {
			if !effect.Exact {
				t.Errorf("%v effect is not exact: %#v", effect.Kind, effect)
			}
		}
	}
}

func TestParseSupportedGainControlEffectsExact(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		context Context
	}{
		{
			name:   "kicked trigger",
			source: "When this creature enters, if it was kicked, gain control of target creature until end of turn. Untap that creature. It gains haste until end of turn.",
		},
		{
			name:    "optional source duration trigger",
			source:  "Whenever a land you control enters, you may gain control of target creature for as long as you control this creature.",
			context: Context{CardName: "Control Creature"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, test.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, sentence := range document.Abilities[0].Sentences {
				for _, effect := range sentence.Effects {
					if !effect.Exact {
						t.Errorf("%v effect is not exact: %#v", effect.Kind, effect)
					}
				}
			}
		})
	}
}

func TestParseRejectsNamedSourceDurationWithTrailingText(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"{T}: Gain control of target creature for as long as you control Merieke Ri Berit, then celebrate.",
		Context{CardName: "Merieke Ri Berit"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Exact {
		t.Fatalf("effects = %#v, want one inexact effect", effects)
	}
}

func TestParseComposedResolvingSyntax(t *testing.T) {
	t.Parallel()
	source := "Return up to two target cards with cycling from your graveyard to your hand, then draw a card."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	sentence := document.Abilities[0].Sentences[0]
	if len(sentence.Effects) != 2 || sentence.Effects[0].Kind != EffectReturn || sentence.Effects[1].Kind != EffectDraw {
		t.Fatalf("effects = %#v", sentence.Effects)
	}
	if sentence.Effects[0].FromZone != zone.Graveyard || sentence.Effects[0].ToZone != zone.Hand ||
		sentence.Effects[1].Amount.Value != 1 || !sentence.Effects[1].Amount.Known {
		t.Fatalf("typed effects = %#v", sentence.Effects)
	}
	if len(sentence.Targets) != 1 ||
		sentence.Targets[0].Cardinality != (TargetCardinalitySyntax{Min: 0, Max: 2}) ||
		sentence.Targets[0].Selection.Kind != SelectionCard ||
		sentence.Targets[0].Selection.Keyword != KeywordCycling {
		t.Fatalf("targets = %#v", sentence.Targets)
	}
}

func TestParseResolvingDurationDynamicAmountAndPayment(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Counter target spell unless its controller pays {2}{U}.\nTarget creature gets +2/+2 for each creature you control until end of turn.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	counterEffect := document.Abilities[0].Sentences[0].Effects[0]
	if counterEffect.Payment.Payer != EffectPaymentPayerTargetController ||
		!slices.Equal(counterEffect.Payment.ManaCost, cost.Mana{cost.O(2), cost.U}) {
		t.Fatalf("payment = %#v", counterEffect.Payment)
	}
	buff := document.Abilities[1].Sentences[0].Effects[0]
	if buff.Duration != EffectDurationUntilEndOfTurn ||
		buff.Amount.DynamicKind != EffectDynamicAmountCount ||
		buff.Amount.DynamicForm != EffectDynamicAmountFormForEach {
		t.Fatalf("buff = %#v", buff)
	}
}

func TestParseResolvingCreateForEachIterator(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"When this enchantment enters, for each Shrine you control, create a 1/1 red Monk creature token.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectCreate || !effect.Exact {
		t.Fatalf("effect = %#v", effect)
	}
	if effect.Amount.DynamicKind != EffectDynamicAmountCount ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
		effect.Amount.Multiplier != 1 {
		t.Fatalf("amount = %#v", effect.Amount)
	}
	if effect.Amount.Selection == nil ||
		len(effect.Amount.Selection.SubtypesAny) != 1 ||
		effect.Amount.Selection.SubtypesAny[0] != "Shrine" {
		t.Fatalf("for-each selection = %#v", effect.Amount.Selection)
	}
}

func TestParseResolvingReplacementAndManaMeaning(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.\n"+
			"Add {G}, {W}, or {U}.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	replacement := document.Abilities[0].Sentences[0].Effects[1].Replacement
	if replacement.Kind != EffectReplacementTwiceThatMany || replacement.Span.Start == replacement.Span.End {
		t.Fatalf("replacement = %#v", replacement)
	}
	if got := document.Abilities[0].Sentences[0].Effects[0].Replacement.Kind; got != EffectReplacementNone {
		t.Fatalf("replaced event modifier = %v, want none", got)
	}
	mana := document.Abilities[1].Sentences[0].Effects[0].Mana
	if !mana.Choice || mana.AnyColor || !slices.Equal(mana.Symbols, []string{"{G}", "{W}", "{U}"}) {
		t.Fatalf("mana = %#v", mana)
	}

	nearMiss, _ := Parse(
		"If an effect would create one or more tokens under your control, it creates twice those tokens instead.\n"+
			"Add {G} and {W}.",
		Context{InstantOrSorcery: true},
	)
	if got := nearMiss.Abilities[0].Sentences[0].Effects[1].Replacement.Kind; got != EffectReplacementInstead {
		t.Fatalf("near-miss replacement = %v, want plain instead", got)
	}
	if got := nearMiss.Abilities[1].Sentences[0].Effects[0].Mana; len(got.Symbols) != 0 || got.AnyColor {
		t.Fatalf("near-miss mana = %#v, want unknown", got)
	}

	modified, _ := Parse(
		"If an effect would create one or more tokens under your control, it creates twice that many tapped tokens instead.",
		Context{InstantOrSorcery: true},
	)
	if got := modified.Abilities[0].Sentences[0].Effects[1].Replacement.Kind; got != EffectReplacementInstead {
		t.Fatalf("modified replacement = %v, want plain instead", got)
	}
	treasure, _ := Parse(
		"If an effect would create one or more tokens under your control, it creates twice that many Treasure tokens instead.",
		Context{InstantOrSorcery: true},
	)
	if got := treasure.Abilities[0].Sentences[0].Effects[1].Replacement.Kind; got != EffectReplacementInstead {
		t.Fatalf("Treasure replacement = %v, want plain instead", got)
	}
}

func TestParseResolvingEffectCompositionOwnership(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Gain control of target creature, then it gains haste until end of turn.\n"+
			"They discard a card, then draw a card.\n"+
			"Add {R}, then draw a card.\n"+
			"Put a charge counter on target artifact with mana value X.\n"+
			"Untap target creature and gain control of it until end of turn.\n"+
			"Tap target creature that entered this turn.\n"+
			"Tap up to X target creatures.\n"+
			"Tap target creature named Bob.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	control := document.Abilities[0].Sentences[0].Effects
	if len(control) != 2 ||
		control[0].Duration != EffectDurationNone ||
		control[1].Duration != EffectDurationUntilEndOfTurn ||
		len(control[0].Targets) != 1 || len(control[1].References) != 1 {
		t.Fatalf("control effects = %#v", control)
	}

	discardDraw := document.Abilities[1].Sentences[0].Effects
	if len(discardDraw) != 2 ||
		discardDraw[0].Context != EffectContextEventPlayer ||
		discardDraw[1].Context != EffectContextPriorSubject {
		t.Fatalf("discard/draw contexts = %#v", discardDraw)
	}

	manaDraw := document.Abilities[2].Sentences[0].Effects
	if len(manaDraw) != 2 || !slices.Equal(manaDraw[0].Mana.Symbols, []string{"{R}"}) {
		t.Fatalf("mana/draw effects = %#v", manaDraw)
	}

	counterEffect := document.Abilities[3].Sentences[0].Effects[0]
	if !counterEffect.Amount.Known || counterEffect.Amount.Value != 1 ||
		len(counterEffect.Targets) != 1 || counterEffect.Targets[0].Selection.Kind != SelectionUnknown {
		t.Fatalf("counter effect = %#v", counterEffect)
	}

	untapControl := document.Abilities[4].Sentences[0].Effects
	if len(untapControl) != 2 ||
		untapControl[0].Duration != EffectDurationNone ||
		untapControl[1].Duration != EffectDurationUntilEndOfTurn {
		t.Fatalf("untap/control durations = %#v", untapControl)
	}
	if target := document.Abilities[5].Sentences[0].Targets[0]; target.Selection.Kind != SelectionUnknown {
		t.Fatalf("relative-clause target = %#v, want unknown selection", target)
	}
	if target := document.Abilities[6].Sentences[0].Targets[0]; target.Cardinality != (TargetCardinalitySyntax{}) {
		t.Fatalf("variable target cardinality = %#v, want unknown", target.Cardinality)
	}
	if target := document.Abilities[7].Sentences[0].Targets[0]; target.Selection.Kind != SelectionUnknown {
		t.Fatalf("unrecognized target qualifier = %#v, want unknown selection", target)
	}
}

func TestParseResolvingSyntaxFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"The counter remains on it.",
		"It was cast this turn.",
		"Double strike is useful.",
		"{1}: Draw a card. Activate only any time you could cast a sorcery.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, sentence := range document.Abilities[0].Sentences {
				if len(sentence.Effects) != 0 && source != "{1}: Draw a card. Activate only any time you could cast a sorcery." {
					t.Fatalf("effects = %#v, want none", sentence.Effects)
				}
				if source == "{1}: Draw a card. Activate only any time you could cast a sorcery." &&
					len(sentence.Effects) > 0 && sentence.Effects[0].Kind != EffectDraw {
					t.Fatalf("activation restriction emitted effect: %#v", sentence.Effects)
				}
			}
		})
	}

	document, _ := Parse("Draw a card for each creatures you control.", Context{InstantOrSorcery: true})
	amount := document.Abilities[0].Sentences[0].Effects[0].Amount
	if amount.Known || amount.DynamicKind != EffectDynamicAmountNone {
		t.Fatalf("ambiguous amount = %#v, want unknown", amount)
	}

	document, _ = Parse("Draw a card, 5 mill.", Context{InstantOrSorcery: true})
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 2 || effects[1].Context != EffectContextUnknown {
		t.Fatalf("non-word subject effects = %#v, want unknown context", effects)
	}
}
