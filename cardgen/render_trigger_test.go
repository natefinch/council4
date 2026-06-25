package cardgen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestRenderCombatTriggerPattern(t *testing.T) {
	ctx := newRenderCtx()
	lit, err := (Renderer{}).renderTriggerPattern(ctx, &game.TriggerPattern{
		Event:                game.EventDamageDealt,
		Source:               game.TriggerSourceSelf,
		Subject:              game.TriggerSubjectDamageSource,
		DamageRecipient:      game.DamageRecipientPermanent,
		DamageRecipientTypes: []types.Card{types.Creature},
		RequireCombatDamage:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Event: game.EventDamageDealt",
		"Source: game.TriggerSourceSelf",
		"Subject: game.TriggerSubjectDamageSource",
		"DamageRecipient: game.DamageRecipientPermanent",
		"DamageRecipientTypes: []types.Card{types.Creature}",
		"RequireCombatDamage: true",
	} {
		if !strings.Contains(lit, want) {
			t.Fatalf("trigger pattern literal %q does not contain %q", lit, want)
		}
	}
}

func TestRenderSaturatedCombatTriggerPattern(t *testing.T) {
	ctx := newRenderCtx()
	lit, err := (Renderer{}).renderTriggerPattern(ctx, &game.TriggerPattern{
		Event:                    game.EventAttackerDeclared,
		RelatedSubjectSelection:  game.Selection{RequiredTypes: []types.Card{types.Creature}},
		DamageRecipientSelection: game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}},
		DamageRecipientIsSource:  true,
		DamageSourceSelection:    game.Selection{Controller: game.ControllerYou},
		AttackRecipient:          game.AttackRecipientPlayer | game.AttackRecipientPlaneswalker,
		AttackRecipientSelection: game.Selection{Controller: game.ControllerYou},
		RequireNonCombatDamage:   true,
		OneOrMore:                true,
		OneOrMorePerAttackTarget: true,
		StepPlayerSourceAttachedSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"RelatedSubjectSelection:",
		"DamageRecipientSelection:",
		"DamageRecipientIsSource: true",
		"DamageSourceSelection:",
		"AttackRecipient: game.AttackRecipientPlayer | game.AttackRecipientPlaneswalker",
		"AttackRecipientSelection:",
		"RequireNonCombatDamage: true",
		"OneOrMorePerAttackTarget: true",
		"StepPlayerSourceAttachedSelection:",
	} {
		if !strings.Contains(lit, want) {
			t.Fatalf("trigger pattern literal %q does not contain %q", lit, want)
		}
	}
}

func TestRenderTriggerPatternRecipientTypesWithoutRecipient(t *testing.T) {
	ctx := newRenderCtx()
	lit, err := (Renderer{}).renderTriggerPattern(ctx, &game.TriggerPattern{
		Event:                game.EventDamageDealt,
		DamageRecipientTypes: []types.Card{types.Creature},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(lit, "DamageRecipientTypes: []types.Card{types.Creature}") {
		t.Fatalf("trigger pattern literal %q does not contain recipient types", lit)
	}
}

func TestRenderLifeTriggerPattern(t *testing.T) {
	ctx := newRenderCtx()
	lit, err := (Renderer{}).renderTriggerPattern(ctx, &game.TriggerPattern{
		Event:  game.EventLifeGained,
		Player: game.TriggerPlayerOpponent,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Event: game.EventLifeGained",
		"Player: game.TriggerPlayerOpponent",
	} {
		if !strings.Contains(lit, want) {
			t.Fatalf("trigger pattern literal %q does not contain %q", lit, want)
		}
	}
}

func TestRenderTriggerPatternNonSelfFields(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	pattern := game.TriggerPattern{
		Event:                 game.EventPermanentEnteredBattlefield,
		Controller:            game.TriggerControllerYou,
		ExcludeSelf:           true,
		Player:                game.TriggerPlayerOpponent,
		RequirePermanentTypes: []types.Card{types.Creature},
		ExcludePermanentTypes: []types.Card{types.Artifact},
		RequireNonToken:       true,
	}
	rendered, err := (Renderer{}).renderTriggerPattern(ctx, &pattern)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.EventPermanentEnteredBattlefield",
		"Controller: game.TriggerControllerYou",
		"ExcludeSelf: true",
		"Player: game.TriggerPlayerOpponent",
		"RequirePermanentTypes: []types.Card{types.Creature}",
		"ExcludePermanentTypes: []types.Card{types.Artifact}",
		"RequireNonToken: true",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered pattern missing %q:\n%s", want, rendered)
		}
	}
	if _, ok := ctx.imports[importTypes]; !ok {
		t.Fatal("renderTriggerPattern with RequirePermanentTypes did not request types import")
	}
	src := "package p\nvar _ = " + rendered
	if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
		t.Fatalf("rendered pattern is not valid Go: %v\n%s", err, rendered)
	}
}

func TestRenderTriggerPatternOneOrMore(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	pattern := game.TriggerPattern{
		Event:                 game.EventPermanentEnteredBattlefield,
		OneOrMore:             true,
		RequirePermanentTypes: []types.Card{types.Artifact},
		Controller:            game.TriggerControllerYou,
	}
	rendered, err := (Renderer{}).renderTriggerPattern(ctx, &pattern)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"OneOrMore: true",
		"RequirePermanentTypes: []types.Card{types.Artifact}",
		"Controller: game.TriggerControllerYou",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered pattern missing %q:\n%s", want, rendered)
		}
	}
	src := "package p\nvar _ = " + rendered
	if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
		t.Fatalf("rendered pattern is not valid Go: %v\n%s", err, rendered)
	}
}

func TestRenderTriggerPatternOpponentController(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	pattern := game.TriggerPattern{
		Event:                 game.EventPermanentEnteredBattlefield,
		Controller:            game.TriggerControllerOpponent,
		RequirePermanentTypes: []types.Card{types.Land},
	}
	rendered, err := (Renderer{}).renderTriggerPattern(ctx, &pattern)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "Controller: game.TriggerControllerOpponent") {
		t.Fatalf("rendered pattern missing opponent controller:\n%s", rendered)
	}
}

func TestRenderTriggerControllerRejectsUnknown(t *testing.T) {
	t.Parallel()
	if _, err := renderTriggerController(game.TriggerControllerFilter(99)); err == nil {
		t.Fatal("expected error for unknown controller filter")
	}
}

func TestRenderTriggerPlayerRejectsUnknown(t *testing.T) {
	t.Parallel()
	if _, err := renderTriggerPlayer(game.TriggerPlayerFilter(99)); err == nil {
		t.Fatal("expected error for unknown player filter")
	}
}

func TestRenderTriggerPatternRejectsUnsupportedFields(t *testing.T) {
	t.Parallel()
	pattern := game.TriggerPattern{
		Event:         game.EventPermanentEnteredBattlefield,
		MatchFromZone: true,
	}
	if _, err := (Renderer{}).renderTriggerPattern(newRenderCtx(), &pattern); err == nil {
		t.Fatal("expected unsupported trigger pattern field error")
	}
}

func TestRenderTriggerPatternRejectsUnrestrictedAbilityActivatedEvent(t *testing.T) {
	t.Parallel()
	pattern := game.TriggerPattern{Event: game.EventAbilityActivated}
	if _, err := (Renderer{}).renderTriggerPattern(newRenderCtx(), &pattern); err == nil {
		t.Fatal("unrestricted ability-activated trigger pattern rendered")
	}

	pattern.ExcludeManaAbility = true
	if _, err := (Renderer{}).renderTriggerPattern(newRenderCtx(), &pattern); err != nil {
		t.Fatalf("non-mana ability-activated trigger pattern: %v", err)
	}
}

func TestRenderTriggerPatternBeginningOfStep(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	pattern := game.TriggerPattern{
		Event:      game.EventBeginningOfStep,
		Controller: game.TriggerControllerYou,
		Step:       game.StepUpkeep,
	}
	rendered, err := (Renderer{}).renderTriggerPattern(ctx, &pattern)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.EventBeginningOfStep",
		"Controller: game.TriggerControllerYou",
		"Step: game.StepUpkeep",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered pattern missing %q:\n%s", want, rendered)
		}
	}
	src := "package p\nvar _ = " + rendered
	if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
		t.Fatalf("rendered pattern is not valid Go: %v\n%s", err, rendered)
	}
}

func TestRenderStepHelperAcceptsKnownSteps(t *testing.T) {
	t.Parallel()
	tests := []struct {
		step game.Step
		want string
	}{
		{game.StepUpkeep, "game.StepUpkeep"},
		{game.StepDraw, "game.StepDraw"},
		{game.StepBeginningOfCombat, "game.StepBeginningOfCombat"},
		{game.StepEnd, "game.StepEnd"},
		{game.StepPrecombatMain, "game.StepPrecombatMain"},
		{game.StepPostcombatMain, "game.StepPostcombatMain"},
	}
	for _, tc := range tests {
		got, err := renderStep(tc.step)
		if err != nil {
			t.Errorf("renderStep(%d): unexpected error: %v", tc.step, err)
			continue
		}
		if got != tc.want {
			t.Errorf("renderStep(%d) = %q, want %q", tc.step, got, tc.want)
		}
	}
}

func TestRenderStepHelperRejectsUnknownStep(t *testing.T) {
	t.Parallel()
	if _, err := renderStep(game.Step(99)); err == nil {
		t.Fatal("expected error for unknown step")
	}
}

func TestRenderTriggerPatternAllSteps(t *testing.T) {
	t.Parallel()
	steps := []struct {
		step game.Step
		want string
	}{
		{game.StepUpkeep, "game.StepUpkeep"},
		{game.StepDraw, "game.StepDraw"},
		{game.StepBeginningOfCombat, "game.StepBeginningOfCombat"},
		{game.StepEnd, "game.StepEnd"},
		{game.StepPrecombatMain, "game.StepPrecombatMain"},
		{game.StepPostcombatMain, "game.StepPostcombatMain"},
	}
	for _, tc := range steps {
		ctx := newRenderCtx()
		pattern := game.TriggerPattern{
			Event: game.EventBeginningOfStep,
			Step:  tc.step,
		}
		rendered, err := (Renderer{}).renderTriggerPattern(ctx, &pattern)
		if err != nil {
			t.Errorf("step %d: unexpected error: %v", tc.step, err)
			continue
		}
		if !strings.Contains(rendered, tc.want) {
			t.Errorf("step %d: rendered pattern missing %q:\n%s", tc.step, tc.want, rendered)
		}
		src := "package p\nvar _ = " + rendered
		if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
			t.Errorf("step %d: rendered pattern is not valid Go: %v\n%s", tc.step, err, rendered)
		}
	}
}

func TestRenderTriggerPatternUnknownStepErrors(t *testing.T) {
	t.Parallel()
	pattern := game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.Step(99),
	}
	if _, err := (Renderer{}).renderTriggerPattern(newRenderCtx(), &pattern); err == nil {
		t.Fatal("expected error for unknown step in trigger pattern")
	}
}

func TestRenderTriggerPatternRejectsMismatchedStepEvent(t *testing.T) {
	t.Parallel()
	tests := []game.TriggerPattern{
		{Event: game.EventPermanentEnteredBattlefield, Step: game.StepUpkeep},
		{Event: game.EventBeginningOfStep},
	}
	for _, pattern := range tests {
		if _, err := (Renderer{}).renderTriggerPattern(newRenderCtx(), &pattern); err == nil {
			t.Fatalf("expected mismatched pattern error for %+v", pattern)
		}
	}
}

func TestRenderTriggerPatternCastWithCardSelection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		pattern   game.TriggerPattern
		wantParts []string
	}{
		{
			name: "unrestricted spell",
			pattern: game.TriggerPattern{
				Event:      game.EventSpellCast,
				Controller: game.TriggerControllerYou,
			},
			wantParts: []string{"game.EventSpellCast", "Controller: game.TriggerControllerYou"},
		},
		{
			name: "creature spell",
			pattern: game.TriggerPattern{
				Event:         game.EventSpellCast,
				Controller:    game.TriggerControllerYou,
				CardSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
			wantParts: []string{
				"game.EventSpellCast",
				"Controller: game.TriggerControllerYou",
				"CardSelection:",
				"types.Creature",
			},
		},
		{
			name: "noncreature spell",
			pattern: game.TriggerPattern{
				Event:         game.EventSpellCast,
				Controller:    game.TriggerControllerAny,
				CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature}},
			},
			wantParts: []string{"ExcludedTypes:", "types.Creature"},
		},
		{
			name: "discard creature card",
			pattern: game.TriggerPattern{
				Event:         game.EventCardDiscarded,
				Player:        game.TriggerPlayerYou,
				CardSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
			wantParts: []string{
				"game.EventCardDiscarded",
				"CardSelection:",
				"RequiredTypes:",
				"types.Creature",
			},
		},
		{
			name: "discard noncreature nonland card",
			pattern: game.TriggerPattern{
				Event:         game.EventCardDiscarded,
				Player:        game.TriggerPlayerYou,
				CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature, types.Land}},
			},
			wantParts: []string{
				"game.EventCardDiscarded",
				"ExcludedTypes:",
				"types.Creature",
				"types.Land",
			},
		},
		{
			name: "instant or sorcery",
			pattern: game.TriggerPattern{
				Event:         game.EventSpellCast,
				Controller:    game.TriggerControllerOpponent,
				CardSelection: game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
			},
			wantParts: []string{"RequiredTypesAny:", "types.Instant", "types.Sorcery"},
		},
		{
			name: "blue spell",
			pattern: game.TriggerPattern{
				Event:         game.EventSpellCast,
				Controller:    game.TriggerControllerYou,
				CardSelection: game.Selection{ColorsAny: []color.Color{color.Blue}},
			},
			wantParts: []string{"CardSelection:", "ColorsAny:", "color.Blue"},
		},
		{
			name: "colorless spell",
			pattern: game.TriggerPattern{
				Event:         game.EventSpellCast,
				Controller:    game.TriggerControllerYou,
				CardSelection: game.Selection{Colorless: true},
			},
			wantParts: []string{"CardSelection:", "Colorless: true"},
		},
		{
			name: "multicolored spell",
			pattern: game.TriggerPattern{
				Event:         game.EventSpellCast,
				Controller:    game.TriggerControllerYou,
				CardSelection: game.Selection{Multicolored: true},
			},
			wantParts: []string{"CardSelection:", "Multicolored: true"},
		},
		{
			name: "mana value spell",
			pattern: game.TriggerPattern{
				Event:      game.EventSpellCast,
				Controller: game.TriggerControllerYou,
				CardSelection: game.Selection{
					ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 5}),
				},
			},
			wantParts: []string{"CardSelection:", "ManaValue:", "compare.GreaterOrEqual", "Value: 5"},
		},
		{
			name: "Spirit or Arcane spell",
			pattern: game.TriggerPattern{
				Event:      game.EventSpellCast,
				Controller: game.TriggerControllerYou,
				CardSelection: game.Selection{
					SubtypesAny: []types.Sub{types.Spirit, types.Arcane},
				},
			},
			wantParts: []string{"CardSelection:", "SubtypesAny:", `types.Sub("Spirit")`, `types.Sub("Arcane")`},
		},
		{
			name: "legendary spell",
			pattern: game.TriggerPattern{
				Event:      game.EventSpellCast,
				Controller: game.TriggerControllerYou,
				CardSelection: game.Selection{
					Supertypes: []types.Super{types.Legendary},
				},
			},
			wantParts: []string{"CardSelection:", "Supertypes:", "types.Legendary"},
		},
		{
			name: "kicked spell",
			pattern: game.TriggerPattern{
				Event:             game.EventSpellCast,
				Controller:        game.TriggerControllerYou,
				RequireKickerPaid: true,
			},
			wantParts: []string{"RequireKickerPaid: true"},
		},
		{
			name: "historic spell",
			pattern: game.TriggerPattern{
				Event:           game.EventSpellCast,
				Controller:      game.TriggerControllerYou,
				RequireHistoric: true,
			},
			wantParts: []string{"RequireHistoric: true"},
		},
		{
			name: "spell from graveyard",
			pattern: game.TriggerPattern{
				Event:         game.EventSpellCast,
				Controller:    game.TriggerControllerYou,
				MatchFromZone: true,
				FromZone:      zone.Graveyard,
			},
			wantParts: []string{"MatchFromZone: true", "FromZone: zone.Graveyard"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := newRenderCtx()
			rendered, err := (Renderer{}).renderTriggerPattern(ctx, &tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, want := range tc.wantParts {
				if !strings.Contains(rendered, want) {
					t.Errorf("rendered pattern missing %q:\n%s", want, rendered)
				}
			}
			src := "package p\nvar _ = " + rendered
			if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
				t.Fatalf("rendered pattern is not valid Go: %v\n%s", err, rendered)
			}
		})
	}
}

func TestRenderTriggerPatternCyclingEvents(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		pattern game.TriggerPattern
		want    []string
	}{
		{
			name: "cycled",
			pattern: game.TriggerPattern{
				Event:  game.EventCycled,
				Player: game.TriggerPlayerYou,
			},
			want: []string{"game.EventCycled", "Player: game.TriggerPlayerYou"},
		},
		{
			name: "cycle or discard",
			pattern: game.TriggerPattern{
				Event:       game.EventCardDiscarded,
				Player:      game.TriggerPlayerYou,
				ExcludeSelf: true,
			},
			want: []string{"game.EventCardDiscarded", "Player: game.TriggerPlayerYou", "ExcludeSelf: true"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rendered, err := (Renderer{}).renderTriggerPattern(newRenderCtx(), &tc.pattern)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range tc.want {
				if !strings.Contains(rendered, want) {
					t.Fatalf("rendered pattern missing %q:\n%s", want, rendered)
				}
			}
			src := "package p\nvar _ = " + rendered
			if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
				t.Fatalf("rendered pattern is not valid Go: %v\n%s", err, rendered)
			}
		})
	}
}

func TestRenderStaticAbilityHandCyclingGrant(t *testing.T) {
	t.Parallel()
	rendered, err := (Renderer{}).renderStaticAbility(newRenderCtx(), &game.StaticAbility{
		Text: "Each land card in your hand has cycling {R}.",
		RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectGrantHandCardAbility,
			AffectedPlayer: game.PlayerYou,
			CardSelection: game.Selection{
				RequiredTypes: []types.Card{types.Land},
			},
			GrantedAbility: game.CyclingActivatedAbility(cost.Mana{cost.R}),
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.RuleEffectGrantHandCardAbility",
		"AffectedPlayer: game.PlayerYou",
		"RequiredTypes: []types.Card{types.Land}",
		"game.CyclingActivatedAbility(cost.Mana{cost.R})",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered static ability missing %q:\n%s", want, rendered)
		}
	}
	src := "package p\nimport (\n\"github.com/natefinch/council4/mtg/game\"\n\"github.com/natefinch/council4/mtg/game/cost\"\n\"github.com/natefinch/council4/mtg/game/types\"\n)\nvar _ = " + rendered
	if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
		t.Fatalf("rendered static ability is not valid Go: %v\n%s", err, rendered)
	}
}

func TestRenderStaticAbilityCyclingCostModifier(t *testing.T) {
	t.Parallel()
	rendered, err := (Renderer{}).renderStaticAbility(newRenderCtx(), &game.StaticAbility{
		Text: "As long as you have seven or more cards in hand, you may pay {0} rather than pay cycling costs.",
		Condition: opt.Val(game.Condition{
			Text:       "As long as you have seven or more cards in hand",
			Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerHandSize, Op: compare.GreaterOrEqual, Value: 7}},
		}),
		RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCostModifier,
			AffectedPlayer: game.PlayerYou,
			CostModifier: game.CostModifier{
				Kind:           game.CostModifierAbility,
				AbilityKeyword: game.Cycling,
				SetManaCost:    opt.Val(cost.Mana{}),
			},
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Aggregate: game.AggregateControllerHandSize, Op: compare.GreaterOrEqual, Value: 7",
		"game.RuleEffectCostModifier",
		"AffectedPlayer: game.PlayerYou",
		"Kind: game.CostModifierAbility",
		"AbilityKeyword: game.Cycling",
		"SetManaCost: opt.Val(cost.Mana{})",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered static ability missing %q:\n%s", want, rendered)
		}
	}
	src := "package p\nimport (\n\"github.com/natefinch/council4/mtg/game\"\n\"github.com/natefinch/council4/mtg/game/cost\"\n\"github.com/natefinch/council4/opt\"\n)\nvar _ = " + rendered
	if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
		t.Fatalf("rendered static ability is not valid Go: %v\n%s", err, rendered)
	}
}

func TestRenderStaticAbilityColorSpellCostModifier(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := (Renderer{}).renderStaticAbility(ctx, &game.StaticAbility{
		RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCostModifier,
			AffectedPlayer: game.PlayerYou,
			CostModifier: game.CostModifier{
				Kind:             game.CostModifierSpell,
				CardSelection:    game.Selection{ColorsAny: []color.Color{color.Red}},
				GenericReduction: 1,
			},
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Kind: game.CostModifierSpell",
		"CardSelection: game.Selection{ColorsAny: []color.Color{color.Red}}",
		"GenericReduction: 1",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered static ability missing %q:\n%s", want, rendered)
		}
	}
	if _, ok := ctx.imports[importColor]; !ok {
		t.Fatal("rendering a colored cost modifier did not register the color import")
	}
	src := "package p\nimport (\n\"github.com/natefinch/council4/mtg/game\"\n\"github.com/natefinch/council4/mtg/game/color\"\n)\nvar _ = " + rendered
	if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
		t.Fatalf("rendered static ability is not valid Go: %v\n%s", err, rendered)
	}
}

func TestRenderStaticAbilityColorlessSpellCostModifier(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := (Renderer{}).renderStaticAbility(ctx, &game.StaticAbility{
		RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCostModifier,
			AffectedPlayer: game.PlayerYou,
			CostModifier: game.CostModifier{
				Kind:             game.CostModifierSpell,
				CardSelection:    game.Selection{Colorless: true},
				GenericReduction: 1,
			},
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "Colorless: true") {
		t.Fatalf("rendered static ability missing colorless match:\n%s", rendered)
	}
	if strings.Contains(rendered, "Color: color.") {
		t.Fatalf("colorless cost modifier should not render a Color field:\n%s", rendered)
	}
	if _, ok := ctx.imports[importColor]; ok {
		t.Fatal("rendering a colorless cost modifier should not register the color import")
	}
}

func TestRenderTriggerPatternRejectsCardSelectionOnNonCastEvent(t *testing.T) {
	t.Parallel()
	pattern := game.TriggerPattern{
		Event:         game.EventPermanentEnteredBattlefield,
		CardSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}
	if _, err := (Renderer{}).renderTriggerPattern(newRenderCtx(), &pattern); err == nil {
		t.Fatal("expected error: CardSelection only allowed on EventSpellCast patterns")
	}
}

func TestRenderTriggerPatternRejectsUnsupportedCardSelectionFields(t *testing.T) {
	t.Parallel()
	pattern := game.TriggerPattern{
		Event: game.EventSpellCast,
		CardSelection: game.Selection{
			Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 2}),
		},
	}
	if _, err := (Renderer{}).renderTriggerPattern(newRenderCtx(), &pattern); err == nil {
		t.Fatal("expected error: Power is unsupported in CardSelection")
	}
}

func TestRenderTriggerPatternAllowsSubtypeCardSelectionOnDiscardEvent(t *testing.T) {
	t.Parallel()
	pattern := game.TriggerPattern{
		Event:  game.EventCardDiscarded,
		Player: game.TriggerPlayerYou,
		CardSelection: game.Selection{
			SubtypesAny: []types.Sub{types.Island, types.Pirate, types.Vehicle},
		},
	}
	rendered, err := (Renderer{}).renderTriggerPattern(newRenderCtx(), &pattern)
	if err != nil {
		t.Fatalf("rendering subtype-union discard CardSelection: %v", err)
	}
	if !strings.Contains(rendered, "SubtypesAny:") {
		t.Fatalf("rendered discard trigger missing subtype filter:\n%s", rendered)
	}
}

func TestRenderTriggerPatternRejectsUnavailableCardSelectionOnDiscardEvent(t *testing.T) {
	t.Parallel()
	pattern := game.TriggerPattern{
		Event:         game.EventCardDiscarded,
		Player:        game.TriggerPlayerYou,
		CardSelection: game.Selection{Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 2})},
	}
	if _, err := (Renderer{}).renderTriggerPattern(newRenderCtx(), &pattern); err == nil {
		t.Fatal("expected error: discard CardSelection cannot filter on power")
	}
}

// TestRenderTriggerPatternSubjectSelection verifies that a SubjectSelection on
// an EventPermanentDied pattern renders correctly and produces valid Go syntax.
func TestRenderTriggerPatternSubjectSelection(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	pattern := game.TriggerPattern{
		Event:       game.EventPermanentDied,
		Controller:  game.TriggerControllerYou,
		ExcludeSelf: true,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}
	rendered, err := (Renderer{}).renderTriggerPattern(ctx, &pattern)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.EventPermanentDied",
		"Controller: game.TriggerControllerYou",
		"ExcludeSelf: true",
		"SubjectSelection:",
		"types.Creature",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered pattern missing %q:\n%s", want, rendered)
		}
	}
	src := "package p\nvar _ = " + rendered
	if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
		t.Fatalf("rendered pattern is not valid Go: %v\n%s", err, rendered)
	}
}

// TestRenderTriggerPatternSubjectSelectionNonToken verifies a NonToken
// SubjectSelection renders correctly.
func TestRenderTriggerPatternSubjectSelectionNonToken(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	pattern := game.TriggerPattern{
		Event: game.EventPermanentDied,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			NonToken:      true,
		},
	}
	rendered, err := (Renderer{}).renderTriggerPattern(ctx, &pattern)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "NonToken: true") {
		t.Fatalf("rendered pattern missing NonToken:\n%s", rendered)
	}
	src := "package p\nvar _ = " + rendered
	if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
		t.Fatalf("rendered pattern is not valid Go: %v\n%s", err, rendered)
	}
}

func TestRenderTriggerPatternSubjectSelectionSupportsEnterEvent(t *testing.T) {
	t.Parallel()
	pattern := game.TriggerPattern{
		Event: game.EventPermanentEnteredBattlefield,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}
	rendered, err := (Renderer{}).renderTriggerPattern(newRenderCtx(), &pattern)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "SubjectSelection: game.Selection{") {
		t.Fatalf("rendered pattern missing SubjectSelection:\n%s", rendered)
	}
}
