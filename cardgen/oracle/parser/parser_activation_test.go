package parser

import "testing"

func TestParseAbilityKinds(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source  string
		context Context
		want    AbilityKind
	}{
		"spell": {
			source:  "Destroy target creature.",
			context: Context{InstantOrSorcery: true},
			want:    AbilitySpell,
		},
		"activated": {
			source: "{T}: Add {G}.",
			want:   AbilityActivated,
		},
		"loyalty": {
			source:  "−2: Target creature you control fights target creature you don't control.",
			context: Context{Planeswalker: true},
			want:    AbilityLoyalty,
		},
		"variable loyalty": {
			source:  "+X: Draw X cards.",
			context: Context{Planeswalker: true},
			want:    AbilityLoyalty,
		},
		"numeric activated": {
			source: "2: Draw a card.",
			want:   AbilityActivated,
		},
		"triggered": {
			source: "Whenever you attack, draw a card.",
			want:   AbilityTriggered,
		},
		"ability word trigger": {
			source: "Formidable — Whenever you attack, draw a card.",
			want:   AbilityTriggered,
		},
		"saga chapter": {
			source: "I, II — Draw a card.",
			context: Context{
				Saga: true,
			},
			want: AbilityChapter,
		},
		"replacement": {
			source: "This land enters tapped.",
			want:   AbilityReplacement,
		},
		"static": {
			source: "Creatures you control have haste.",
			want:   AbilityStatic,
		},
		"reminder": {
			source: "(This creature can block creatures with flying.)",
			want:   AbilityReminder,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, test.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %d", len(document.Abilities))
			}
			if got := document.Abilities[0].Kind; got != test.want {
				t.Fatalf("kind = %s, want %s", got, test.want)
			}
		})
	}
}

func TestParseTypedActivationRestrictions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		restriction string
		kind        ActivationRestrictionKind
		count       ActivationFrequencyCountKind
		period      ActivationFrequencyPeriodKind
		quantifier  PhaseStepQuantifierKind
		player      TriggerPlayerSelectorKind
		phaseStep   PhaseStepNameKind
	}{
		{"sorcery timing", "Activate only as a sorcery.", ActivationRestrictionSorceryTiming, ActivationFrequencyCountUnknown, ActivationFrequencyPeriodUnknown, PhaseStepQuantifierUnknown, TriggerPlayerSelectorUnknown, PhaseStepNameUnknown},
		{"once each turn", "Activate only once each turn.", ActivationRestrictionFrequency, ActivationFrequencyCountOnce, ActivationFrequencyPeriodTurn, PhaseStepQuantifierUnknown, TriggerPlayerSelectorUnknown, PhaseStepNameUnknown},
		{"combat", "Activate only during combat.", ActivationRestrictionPhaseStep, ActivationFrequencyCountUnknown, ActivationFrequencyPeriodUnknown, PhaseStepQuantifierNone, TriggerPlayerSelectorAny, PhaseStepNameCombat},
		{"controller upkeep", "Activate only during your upkeep.", ActivationRestrictionPhaseStep, ActivationFrequencyCountUnknown, ActivationFrequencyPeriodUnknown, PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameUpkeep},
		{"typed unsupported phase", "Activate only during your end step.", ActivationRestrictionPhaseStep, ActivationFrequencyCountUnknown, ActivationFrequencyPeriodUnknown, PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameEndStep},
		{"explicit unsupported", "Activate only before combat.", ActivationRestrictionUnsupported, ActivationFrequencyCountUnknown, ActivationFrequencyPeriodUnknown, PhaseStepQuantifierUnknown, TriggerPlayerSelectorUnknown, PhaseStepNameUnknown},
		{"instant timing", "Activate only as an instant.", ActivationRestrictionInstantTiming, ActivationFrequencyCountUnknown, ActivationFrequencyPeriodUnknown, PhaseStepQuantifierUnknown, TriggerPlayerSelectorUnknown, PhaseStepNameUnknown},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source := "{1}: Draw a card. " + test.restriction
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			restrictions := document.Abilities[0].ActivationRestrictions
			if len(restrictions) != 1 {
				t.Fatalf("restrictions = %#v, want one", restrictions)
			}
			restriction := restrictions[0]
			if restriction.Kind != test.kind ||
				restriction.Frequency.Count.Kind != test.count ||
				restriction.Frequency.Period.Kind != test.period ||
				restriction.PhaseStep.Quantifier.Kind != test.quantifier ||
				restriction.PhaseStep.Player.Kind != test.player ||
				restriction.PhaseStep.Name.Kind != test.phaseStep {
				t.Fatalf("restriction = %#v", restriction)
			}
			assertTextSpan(t, "activation restriction", source, restriction.Span, test.restriction)
			switch restriction.Kind {
			case ActivationRestrictionSorceryTiming:
				assertSpanContains(t, "sorcery timing", restriction.Span, restriction.SorcerySpan)
			case ActivationRestrictionFrequency:
				assertSpanContains(t, "frequency count", restriction.Span, restriction.Frequency.Count.Span)
				assertSpanContains(t, "frequency period", restriction.Span, restriction.Frequency.Period.Span)
			case ActivationRestrictionPhaseStep:
				assertSpanContains(t, "phase/step name", restriction.Span, restriction.PhaseStep.Name.Span)
			default:
			}
		})
	}
}

func TestParseActivationRestrictionGrammarVariants(t *testing.T) {
	t.Parallel()
	for _, restriction := range []string{
		"Activate only at sorcery speed.",
		"Activate only any time you could cast a sorcery.",
		"Activate only as an instant.",
		"Activate only at instant speed.",
		"Activate only any time you could cast an instant.",
		"Activate only once per turn.",
		"Activate only one time every turn.",
		"Activate only during each combat.",
		"Activate only during each of your upkeeps.",
	} {
		t.Run(restriction, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse("{1}: Draw a card. "+restriction, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			restrictions := document.Abilities[0].ActivationRestrictions
			if len(restrictions) != 1 || restrictions[0].Kind == ActivationRestrictionUnsupported {
				t.Fatalf("restrictions = %#v, want one supported typed restriction", restrictions)
			}
		})
	}
}

func TestParseComposedActivationRestrictions(t *testing.T) {
	t.Parallel()
	source := "{1}: Draw a card. (Before.) Activate only once per turn. (Between.) Activate only at sorcery speed. (After.)"
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	restrictions := document.Abilities[0].ActivationRestrictions
	if len(restrictions) != 2 ||
		restrictions[0].Kind != ActivationRestrictionFrequency ||
		restrictions[1].Kind != ActivationRestrictionSorceryTiming {
		t.Fatalf("restrictions = %#v", restrictions)
	}
}

func TestParseConjoinedActivationRestrictions(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		text  string
		kinds []ActivationRestrictionKind
	}{
		{
			"sorcery and once each turn",
			"{1}: Draw a card. Activate only as a sorcery and only once each turn.",
			[]ActivationRestrictionKind{ActivationRestrictionSorceryTiming, ActivationRestrictionFrequency},
		},
		{
			"player turn and once each turn",
			"{1}: Draw a card. Activate only during your turn and only once each turn.",
			[]ActivationRestrictionKind{ActivationRestrictionPlayerTurn, ActivationRestrictionFrequency},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.text, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			restrictions := document.Abilities[0].ActivationRestrictions
			if len(restrictions) != len(test.kinds) {
				t.Fatalf("restrictions = %#v, want %d clauses", restrictions, len(test.kinds))
			}
			for i, kind := range test.kinds {
				if restrictions[i].Kind != kind {
					t.Fatalf("restriction[%d].Kind = %v, want %v", i, restrictions[i].Kind, kind)
				}
			}
		})
	}
}

func TestParseActivationPlayerTurnRestriction(t *testing.T) {
	t.Parallel()
	source := "{1}: Draw a card. Activate only during your turn."
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	restrictions := document.Abilities[0].ActivationRestrictions
	if len(restrictions) != 1 {
		t.Fatalf("restrictions = %#v, want one", restrictions)
	}
	restriction := restrictions[0]
	if restriction.Kind != ActivationRestrictionPlayerTurn ||
		restriction.PlayerTurn.Player.Kind != TriggerPlayerSelectorYou {
		t.Fatalf("restriction = %#v, want your-turn player restriction", restriction)
	}
	assertTextSpan(t, "activation restriction", source, restriction.Span, "Activate only during your turn.")
	assertSpanContains(t, "player turn player", restriction.Span, restriction.PlayerTurn.Player.Span)
}

func TestParseActivationTurnBeforeAttackersRestriction(t *testing.T) {
	t.Parallel()
	source := "{T}: Target creature gets +2/+2 until end of turn. " +
		"Activate only during your turn, before attackers are declared."
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	restrictions := document.Abilities[0].ActivationRestrictions
	if len(restrictions) != 1 {
		t.Fatalf("restrictions = %#v, want one", restrictions)
	}
	restriction := restrictions[0]
	if restriction.Kind != ActivationRestrictionTurnBeforeAttackers ||
		restriction.PlayerTurn.Player.Kind != TriggerPlayerSelectorYou {
		t.Fatalf("restriction = %#v, want turn-before-attackers your restriction", restriction)
	}
	assertTextSpan(t, "activation restriction", source, restriction.Span,
		"Activate only during your turn, before attackers are declared.")
}

func TestParseActivationTurnBeforeAttackersFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"{1}: Draw a card. Activate only during an opponent's turn, before attackers are declared.",
		"{1}: Draw a card. Activate only during your turn, before blockers are declared.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			restrictions := document.Abilities[0].ActivationRestrictions
			if len(restrictions) != 1 || restrictions[0].Kind == ActivationRestrictionTurnBeforeAttackers {
				t.Fatalf("restrictions = %#v, want non-turn-before-attackers restriction", restrictions)
			}
		})
	}
}

func TestParseActivationPlayerTurnFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"{1}: Draw a card. Activate only during an opponent's turn.",
		"{1}: Draw a card. Activate only during each turn.",
		"{1}: Draw a card. Activate only during your turns.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			restrictions := document.Abilities[0].ActivationRestrictions
			if len(restrictions) != 1 || restrictions[0].Kind == ActivationRestrictionPlayerTurn {
				t.Fatalf("restrictions = %#v, want non-player-turn restriction", restrictions)
			}
		})
	}
}

func TestParseActivationRestrictionsFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"{1}: Draw a card. Activate only once each round.",
		"{1}: Draw a card. Activate only during your next upkeep.",
		"{1}: Draw a card. Activate only during combat on your turn.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			restrictions := document.Abilities[0].ActivationRestrictions
			if len(restrictions) != 1 || restrictions[0].Kind != ActivationRestrictionUnsupported {
				t.Fatalf("restrictions = %#v, want explicit unsupported restriction", restrictions)
			}
		})
	}
	for _, source := range []string{
		"{1}: Draw a card. Activate only if you control a creature.",
		"{1}: Activate only as a sorcery. Draw a card.",
		"Activate only as a sorcery.",
		"{1}: Draw a card. Activate as a sorcery.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if restrictions := document.Abilities[0].ActivationRestrictions; len(restrictions) != 0 {
				t.Fatalf("restrictions = %#v, want none", restrictions)
			}
		})
	}
}
