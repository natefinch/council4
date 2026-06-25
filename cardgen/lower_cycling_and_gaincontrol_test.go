package cardgen

import (
	goparser "go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerCyclingTriggers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oracle      string
		wantEvent   game.EventKind
		excludeSelf bool
	}{
		{
			name:      "cycle a card",
			oracle:    "Whenever you cycle a card, draw a card.",
			wantEvent: game.EventCycled,
		},
		{
			name:        "cycle another card",
			oracle:      "Whenever you cycle another card, draw a card.",
			wantEvent:   game.EventCycled,
			excludeSelf: true,
		},
		{
			name:      "cycle or discard",
			oracle:    "Whenever you cycle or discard a card, draw a card.",
			wantEvent: game.EventCardDiscarded,
		},
		{
			name:        "cycle or discard another",
			oracle:      "Whenever you cycle or discard another card, draw a card.",
			wantEvent:   game.EventCardDiscarded,
			excludeSelf: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: tc.oracle,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			pattern := face.TriggeredAbilities[0].Trigger.Pattern
			if pattern.Event != tc.wantEvent {
				t.Errorf("event = %v, want %v", pattern.Event, tc.wantEvent)
			}
			if pattern.Player != game.TriggerPlayerYou {
				t.Errorf("player = %v, want TriggerPlayerYou", pattern.Player)
			}
			if pattern.ExcludeSelf != tc.excludeSelf {
				t.Errorf("ExcludeSelf = %v, want %v", pattern.ExcludeSelf, tc.excludeSelf)
			}
		})
	}
}

func TestLowerCycleThisCardSelfTrigger(t *testing.T) {
	t.Parallel()
	// Magmakin Artillerist shape: a self-source "When you cycle this card"
	// trigger whose body uses the pronoun "it" for the cycled card.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Artillerist",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental",
		OracleText: "Cycling {1}{R} ({1}{R}, Discard this card: Draw a card.)\nWhen you cycle this card, it deals 1 damage to each opponent.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.Type != game.TriggerWhen {
		t.Errorf("trigger type = %v, want TriggerWhen", trigger.Type)
	}
	if trigger.Pattern.Event != game.EventCycled {
		t.Errorf("event = %v, want EventCycled", trigger.Pattern.Event)
	}
	if trigger.Pattern.Player != game.TriggerPlayerYou {
		t.Errorf("player = %v, want TriggerPlayerYou", trigger.Pattern.Player)
	}
	if trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Errorf("source = %v, want TriggerSourceSelf", trigger.Pattern.Source)
	}
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1 (Cycling)", len(face.ActivatedAbilities))
	}
}

func TestGenerateExecutableCardSourceMagmakinArtillerist(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Magmakin Artillerist",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental Pirate",
		ManaCost:   "{2}{R}",
		OracleText: "Whenever you discard one or more cards, this creature deals that much damage to each opponent.\nCycling {1}{R} ({1}{R}, Discard this card: Draw a card.)\nWhen you cycle this card, it deals 1 damage to each opponent.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "magmakin.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"game.EventCycled",
		"game.TriggerSourceSelf",
		"game.CyclingActivatedAbility",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestLowerHandCyclingGrants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		oracle    string
		wantTypes []types.Card
		wantCost  cost.Mana
	}{
		{
			name:      "land cards",
			oracle:    "Each land card in your hand has cycling {R}.",
			wantTypes: []types.Card{types.Land},
			wantCost:  cost.Mana{cost.R},
		},
		{
			name:      "creature cards",
			oracle:    "Each creature card in your hand has cycling {1}{U}. ({1}{U}, Discard that card: Draw a card.)",
			wantTypes: []types.Card{types.Creature},
			wantCost:  cost.Mana{cost.O(1), cost.U},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Grant",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: tc.oracle,
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("got %d static abilities, want 1", len(face.StaticAbilities))
			}
			body := face.StaticAbilities[0].Body
			if len(body.RuleEffects) != 1 {
				t.Fatalf("rule effects = %+v, want one", body.RuleEffects)
			}
			effect := body.RuleEffects[0]
			if effect.Kind != game.RuleEffectGrantHandCardAbility {
				t.Fatalf("rule effect kind = %v, want RuleEffectGrantHandCardAbility", effect.Kind)
			}
			if effect.AffectedPlayer != game.PlayerYou {
				t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
			}
			if !slices.Equal(effect.CardSelection.RequiredTypes, tc.wantTypes) {
				t.Fatalf("required types = %v, want %v", effect.CardSelection.RequiredTypes, tc.wantTypes)
			}
			gotCost, ok := game.ActivatedBodyCyclingCost(&effect.GrantedAbility)
			if !ok || !slices.Equal(gotCost, tc.wantCost) {
				t.Fatalf("cycling cost = %v, %v; want %v", gotCost, ok, tc.wantCost)
			}
		})
	}
}

func TestLowerCyclingCostModifiers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		oracle              string
		wantReduction       int
		wantSetCost         bool
		wantHandSize        int
		wantFirstCycleLimit bool
	}{
		{
			name:          "Fluctuator",
			oracle:        "Cycling abilities you activate cost up to {2} less to activate.",
			wantReduction: 2,
		},
		{
			name:         "New Perspectives",
			oracle:       "As long as you have seven or more cards in hand, you may pay {0} rather than pay cycling costs.",
			wantSetCost:  true,
			wantHandSize: 7,
		},
		{
			name:                "Gavi Nest Warden",
			oracle:              "You may pay {0} rather than pay the cycling cost of the first card you cycle each turn.",
			wantSetCost:         true,
			wantFirstCycleLimit: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       tc.name,
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: tc.oracle,
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("got %d static abilities, want 1", len(face.StaticAbilities))
			}
			body := face.StaticAbilities[0].Body
			if len(body.RuleEffects) != 1 {
				t.Fatalf("rule effects = %+v, want one", body.RuleEffects)
			}
			if body.Condition.Exists != (tc.wantHandSize > 0) {
				t.Fatalf("condition exists = %v, want %v", body.Condition.Exists, tc.wantHandSize > 0)
			}
			if tc.wantHandSize > 0 && body.Condition.Val.ControllerHandSizeAtLeast != tc.wantHandSize {
				t.Fatalf("hand-size condition = %d, want %d", body.Condition.Val.ControllerHandSizeAtLeast, tc.wantHandSize)
			}
			effect := body.RuleEffects[0]
			if effect.Kind != game.RuleEffectCostModifier {
				t.Fatalf("rule effect kind = %v, want RuleEffectCostModifier", effect.Kind)
			}
			if effect.AffectedPlayer != game.PlayerYou {
				t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
			}
			modifier := effect.CostModifier
			if modifier.Kind != game.CostModifierAbility || modifier.AbilityKeyword != game.Cycling {
				t.Fatalf("modifier = %+v, want Cycling ability modifier", modifier)
			}
			if modifier.GenericReduction != tc.wantReduction {
				t.Fatalf("generic reduction = %d, want %d", modifier.GenericReduction, tc.wantReduction)
			}
			if modifier.SetManaCost.Exists != tc.wantSetCost {
				t.Fatalf("set mana cost exists = %v, want %v", modifier.SetManaCost.Exists, tc.wantSetCost)
			}
			if tc.wantSetCost && len(modifier.SetManaCost.Val) != 0 {
				t.Fatalf("set mana cost = %v, want zero", modifier.SetManaCost.Val)
			}
			if modifier.FirstCycleEachTurn != tc.wantFirstCycleLimit {
				t.Fatalf("first-cycle limit = %v, want %v", modifier.FirstCycleEachTurn, tc.wantFirstCycleLimit)
			}
		})
	}
}

func TestLowerHandCyclingGrantRejectsHistoric(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Jo Grant",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human",
		OracleText: "Each historic card in your hand has cycling {2}{W}. ({2}{W}, Discard that card: Draw a card.)",
		Power:      new("3"),
		Toughness:  new("4"),
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostic for unsupported historic hand cycling grant")
	}
	if !strings.Contains(diagnostics[0].Detail, "historic card predicates are not supported") {
		t.Fatalf("diagnostic = %#v, want historic predicate detail", diagnostics[0])
	}
}

func TestLowerGainControlUntapHasteSequence(t *testing.T) {
	t.Parallel()
	// Act of Treason pattern.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Act",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Gain control of target creature until end of turn. Untap that creature. It gains haste until end of turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	if mode.Targets[0].Selection.Val.RequiredTypesAny[0] != types.Creature {
		t.Fatalf("target type = %v, want Creature", mode.Targets[0].Selection.Val.RequiredTypesAny)
	}
	if len(mode.Sequence) != 3 {
		t.Fatalf("sequence len = %d, want 3", len(mode.Sequence))
	}
	checkGainControlPrimitive(t, mode, 0, game.DurationUntilEndOfTurn)
	checkUntapPrimitive(t, mode, 1)
	checkKeywordGrantPrimitive(t, mode, 2, game.Haste)
}

func TestLowerGainControlUntapHasteScrySequence(t *testing.T) {
	t.Parallel()
	// Portent of Betrayal pattern: Gain control + Untap + Haste + Scry.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Portent",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Gain control of target creature until end of turn. Untap that creature. It gains haste until end of turn. Scry 1.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 4 {
		t.Fatalf("sequence len = %d, want 4", len(mode.Sequence))
	}
	checkGainControlPrimitive(t, mode, 0, game.DurationUntilEndOfTurn)
	checkUntapPrimitive(t, mode, 1)
	checkKeywordGrantPrimitive(t, mode, 2, game.Haste)
	scry, ok := mode.Sequence[3].Primitive.(game.Scry)
	if !ok {
		t.Fatalf("sequence[3] = %T, want game.Scry", mode.Sequence[3].Primitive)
	}
	if scry.Amount.Value() != 1 {
		t.Fatalf("Scry.Amount = %v, want 1", scry.Amount)
	}
}

func TestLowerGainControlCounterUntapHasteSequence(t *testing.T) {
	t.Parallel()
	// Mark of Mutiny's actual oracle text pattern: counter and untap are
	// in the same sentence ("Put a +1/+1 counter on it and untap it."),
	// followed by a haste grant.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mark",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Gain control of target creature until end of turn. Put a +1/+1 counter on it and untap it. That creature gains haste until end of turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 4 {
		t.Fatalf("sequence len = %d, want 4", len(mode.Sequence))
	}
	checkGainControlPrimitive(t, mode, 0, game.DurationUntilEndOfTurn)
	addCtr, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.AddCounter", mode.Sequence[1].Primitive)
	}
	if addCtr.Object != game.TargetPermanentReference(0) || addCtr.Amount.Value() != 1 {
		t.Fatalf("AddCounter = %+v", addCtr)
	}
	checkUntapPrimitive(t, mode, 2)
	checkKeywordGrantPrimitive(t, mode, 3, game.Haste)
}

func TestLowerGainControlActivatedAbility(t *testing.T) {
	t.Parallel()
	// Captivating Crew pattern: activated ability body.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Crew",
		Layout:     "normal",
		TypeLine:   "Creature — Human Pirate",
		OracleText: "{3}{R}: Gain control of target creature an opponent controls until end of turn. Untap that creature.",
		Power:      new("3"),
		Toughness:  new("3"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	if mode.Targets[0].Selection.Val.Controller != game.ControllerOpponent {
		t.Fatalf("target controller predicate = %v, want Opponent", mode.Targets[0].Selection.Val.Controller)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence len = %d, want 2", len(mode.Sequence))
	}
	checkGainControlPrimitive(t, mode, 0, game.DurationUntilEndOfTurn)
	checkUntapPrimitive(t, mode, 1)
}

func TestLowerGainControlPermanentDuration(t *testing.T) {
	t.Parallel()
	// Nicol Bolas style: single gain-control with permanent duration.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bolas",
		Layout:     "normal",
		TypeLine:   "Planeswalker — Bolas",
		OracleText: "−2: Gain control of target creature.",
	})
	if len(face.LoyaltyAbilities) != 1 {
		t.Fatalf("loyalty abilities = %d, want 1", len(face.LoyaltyAbilities))
	}
	mode := face.LoyaltyAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
	}
	checkGainControlPrimitive(t, mode, 0, game.DurationPermanent)
}

func TestLowerGainControlUntapReversedOrder(t *testing.T) {
	t.Parallel()
	// Threaten pattern: Untap first, gain control second (same sentence).
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Threaten",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Untap target creature and gain control of it until end of turn. That creature gains haste until end of turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 3 {
		t.Fatalf("mode targets=%d seq=%d, want 1 target 3 instructions", len(mode.Targets), len(mode.Sequence))
	}
	// Sequence order follows oracle text: Untap, then GainControl, then Haste.
	checkUntapPrimitive(t, mode, 0)
	checkGainControlPrimitive(t, mode, 1, game.DurationUntilEndOfTurn)
	checkKeywordGrantPrimitive(t, mode, 2, game.Haste)
}

func TestLowerGainControlRejectsControllerYouTarget(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Self-Control",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Gain control of target creature you control.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostic for gaining control of your own permanent")
	}
}

func TestLowerGainControlRejectsMultipleEffectsWithoutBackRef(t *testing.T) {
	t.Parallel()
	// A sequence where the second Untap has a new target (not a back-ref) should
	// fall through to the general ordered-sequence lowerer, not the gain-control
	// path.  We just verify it doesn't produce a bogus zero-diagnostic result.
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Weird",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Gain control of target creature until end of turn. Untap target land.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostic for unsupported multi-target gain-control spell")
	}
}

func TestLowerGainControlRejectsSourceBoundFollowOn(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Relic",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Gain control of target creature until end of turn. Untap this artifact.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostic for source-bound gain-control follow-on")
	}
}

func TestGenerateExecutableCardSourceGainControlRendersApplyContinuous(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Treason",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Gain control of target creature until end of turn. Untap that creature. It gains haste until end of turn.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "test_treason.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"game.ApplyContinuous",
		"game.LayerControl",
		"NewController: opt.Val(game.Player1)",
		"game.DurationUntilEndOfTurn",
		"game.Untap",
		"game.TargetPermanentReference(0)",
		"game.LayerAbility",
		"game.Haste",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourceGainControlPermanentDurationRenders(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Bolas",
		Layout:     "normal",
		TypeLine:   "Planeswalker — Bolas",
		OracleText: "−2: Gain control of target creature.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "test_bolas.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"game.ApplyContinuous",
		"game.LayerControl",
		"NewController: opt.Val(game.Player1)",
		"game.DurationPermanent",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
