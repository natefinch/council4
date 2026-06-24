package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestRenderConditionForETBReplacementRejectsNegativeThresholds(t *testing.T) {
	tests := map[string]game.Condition{
		"controller life": {ControllerLifeAtLeast: -1},
		"any player life": {AnyPlayerLifeAtMost: -1},
		"opponent count":  {OpponentCountAtLeast: -1},
	}

	for name, condition := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := (Renderer{}).renderConditionForETBReplacement(&renderCtx{}, &condition); err == nil {
				t.Fatal("expected negative threshold error")
			}
		})
	}
}

func TestRenderApplyContinuousTemporaryEffects(t *testing.T) {
	t.Parallel()
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.ApplyContinuous{
		ContinuousEffects: []game.ContinuousEffect{
			{
				Layer: game.LayerPowerToughnessModify,
				Group: game.BattlefieldGroup(game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Controller:    game.ControllerYou,
				}),
				PowerDelta:     2,
				ToughnessDelta: 2,
			},
			{
				Layer:       game.LayerAbility,
				Group:       game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
				AddKeywords: []game.Keyword{game.Hexproof, game.Indestructible},
			},
		},
		Duration: game.DurationUntilEndOfTurn,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.ApplyContinuous",
		"game.BattlefieldGroup",
		"Controller: game.ControllerYou",
		"game.LayerPowerToughnessModify",
		"PowerDelta: 2",
		"ToughnessDelta: 2",
		"game.LayerAbility",
		"game.Hexproof",
		"game.Indestructible",
		"game.DurationUntilEndOfTurn",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered temporary effect missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderApplyContinuousChooseFromGroup(t *testing.T) {
	t.Parallel()
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.ApplyContinuous{
		ChooseFrom: game.ObjectControlledGroup(
			game.SourcePermanentReference(),
			game.Selection{RequiredTypes: []types.Card{types.Land}},
		),
		ChooseUpTo: game.Dynamic(game.DynamicAmount{
			Kind:      game.DynamicAmountChosenNumber,
			ResultKey: game.ResultKey("primal-pay"),
		}),
		Prompt: "Choose lands to animate",
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:    game.LayerType,
			AddTypes: []types.Card{types.Creature},
		}},
		Duration: game.DurationPermanent,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.ApplyContinuous",
		"ChooseFrom:",
		"game.ObjectControlledGroup",
		"ChooseUpTo:",
		"DynamicAmountChosenNumber",
		`Prompt: "Choose lands to animate"`,
		"game.DurationPermanent",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered choose-from group missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderReplacementAbilityGroupEntersTapped(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		ability game.ReplacementAbility
		want    string
	}{
		{
			name: "opponent creatures",
			ability: game.EntersTappedGroupReplacement(
				"Creatures your opponents control enter tapped.",
				game.TriggerControllerOpponent,
				types.Creature,
			),
			want: `game.EntersTappedGroupReplacement("Creatures your opponents control enter tapped.", game.TriggerControllerOpponent, types.Creature)`,
		},
		{
			name: "all permanents",
			ability: game.EntersTappedGroupReplacement(
				"Permanents enter the battlefield tapped.",
				game.TriggerControllerAny,
			),
			want: `game.EntersTappedGroupReplacement("Permanents enter the battlefield tapped.", game.TriggerControllerAny)`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			rendered, err := (Renderer{}).renderReplacementAbility(newRenderCtx(), &test.ability)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(rendered, test.want) {
				t.Fatalf("rendered group enters-tapped missing %q:\n%s", test.want, rendered)
			}
		})
	}
}

func TestRenderReplacementAbilityEntersTappedWithCounters(t *testing.T) {
	t.Parallel()
	ability := game.EntersTappedWithCountersReplacement(
		"This land enters tapped with two charge counters on it.",
		game.CounterPlacement{Kind: counter.Charge, Amount: 2},
	)
	rendered, err := (Renderer{}).renderReplacementAbility(newRenderCtx(), &ability)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`game.EntersTappedWithCountersReplacement("This land enters tapped with two charge counters on it."`,
		"game.CounterPlacement{Kind: counter.Charge, Amount: 2}",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered combined replacement missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderZoneChangeTriggerExclusionAndFaceDownFilters(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := (Renderer{}).renderTriggerPattern(ctx, &game.TriggerPattern{
		Event:         game.EventZoneChanged,
		MatchFromZone: true,
		FromZone:      zone.Battlefield,
		ExcludeToZone: true,
		ToZone:        zone.Graveyard,
		MatchFaceDown: true,
		FaceDown:      true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{
		"game.EventZoneChanged",
		"ExcludeToZone: true",
		"ToZone: zone.Graveyard",
		"MatchFaceDown: true",
		"FaceDown: true",
	} {
		if !strings.Contains(rendered, wanted) {
			t.Fatalf("rendered trigger missing %q:\n%s", wanted, rendered)
		}
	}
}

func TestRenderZoneDestinationReplacement(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	ability := game.ReplacementAbility{
		Text: "If Darksteel Colossus would be put into a graveyard from anywhere, reveal Darksteel Colossus and shuffle it into its owner's library instead.",
		Replacement: game.ReplacementEffect{
			MatchEvent:         game.EventZoneChanged,
			MatchToZone:        true,
			ToZone:             zone.Graveyard,
			ReplaceToZone:      zone.Library,
			ShuffleIntoLibrary: true,
			RevealSource:       true,
			Duration:           game.DurationPermanent,
		},
	}
	rendered, err := (Renderer{}).renderReplacementAbility(ctx, &ability)
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{
		"game.ReplacementAbility",
		"game.EventZoneChanged",
		"ToZone: zone.Graveyard",
		"ReplaceToZone: zone.Library",
		"ShuffleIntoLibrary: true",
		"RevealSource: true",
	} {
		if !strings.Contains(rendered, wanted) {
			t.Fatalf("rendered replacement missing %q:\n%s", wanted, rendered)
		}
	}
	if _, ok := ctx.imports[importZone]; !ok {
		t.Fatal("zone-destination replacement did not request zone import")
	}
}

func TestRenderTokenCreationReplacement(t *testing.T) {
	t.Parallel()
	ability := game.TokenCreationReplacement(
		"If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.",
		2,
		game.TriggerControllerYou,
	)
	rendered, err := (Renderer{}).renderReplacementAbility(newRenderCtx(), &ability)
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{
		"game.TokenCreationReplacement",
		"2",
		"game.TriggerControllerYou",
	} {
		if !strings.Contains(rendered, wanted) {
			t.Fatalf("rendered replacement missing %q:\n%s", wanted, rendered)
		}
	}
}

func TestRenderDamageReplacement(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	ability := game.DamageReplacementExcludingSource(
		"If another red source you control would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.",
		0,
		1,
		[]color.Color{color.Red},
		game.TriggerControllerYou,
	)
	rendered, err := (Renderer{}).renderReplacementAbility(ctx, &ability)
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{
		"game.DamageReplacementExcludingSource",
		"0",
		"1",
		"color.Red",
		"game.TriggerControllerYou",
	} {
		if !strings.Contains(rendered, wanted) {
			t.Fatalf("rendered replacement missing %q:\n%s", wanted, rendered)
		}
	}
	if _, ok := ctx.imports[importColor]; !ok {
		t.Fatal("damage replacement did not request color import")
	}
}

func TestRenderCounterPlacementReplacement(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	ability := game.CounterPlacementReplacement(
		"If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.",
		2,
		0,
		counter.PlusOnePlusOne,
		game.TriggerControllerYou,
	)
	rendered, err := (Renderer{}).renderReplacementAbility(ctx, &ability)
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{
		"game.CounterPlacementReplacement",
		"2",
		"counter.PlusOnePlusOne",
		"game.TriggerControllerYou",
	} {
		if !strings.Contains(rendered, wanted) {
			t.Fatalf("rendered replacement missing %q:\n%s", wanted, rendered)
		}
	}
	if _, ok := ctx.imports[importCounter]; !ok {
		t.Fatal("counter-placement replacement did not request counter import")
	}
}

func TestRenderAnyCounterPlacementReplacement(t *testing.T) {
	t.Parallel()
	ability := game.AnyCounterPlacementReplacement(
		"If one or more counters would be put on a permanent or player, twice that many of each of those kinds of counters are put on that permanent or player instead.",
		2,
		0,
		game.TriggerControllerYou,
	)
	rendered, err := (Renderer{}).renderReplacementAbility(newRenderCtx(), &ability)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "game.AnyCounterPlacementReplacement") {
		t.Fatalf("rendered replacement missing any-counter constructor:\n%s", rendered)
	}
}

func TestRenderControlledPermanentCounterPlacementReplacement(t *testing.T) {
	t.Parallel()
	ability := game.ControlledPermanentCounterPlacementReplacement(
		"If an effect would put one or more counters on a permanent you control, it puts twice that many of those counters on that permanent instead.",
		2,
		0,
		game.TriggerControllerYou,
	)
	rendered, err := (Renderer{}).renderReplacementAbility(newRenderCtx(), &ability)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "game.ControlledPermanentCounterPlacementReplacement") {
		t.Fatalf("rendered replacement missing controlled-permanent constructor:\n%s", rendered)
	}
}

func TestRenderControlledPermanentCounterKindPlacementReplacement(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	ability := game.ControlledPermanentCounterKindPlacementReplacement(
		"If one or more +1/+1 counters would be put on a permanent you control, that many plus one +1/+1 counters are put on that permanent instead.",
		0,
		1,
		counter.PlusOnePlusOne,
		game.TriggerControllerYou,
	)
	rendered, err := (Renderer{}).renderReplacementAbility(ctx, &ability)
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{
		"game.ControlledPermanentCounterKindPlacementReplacement",
		"counter.PlusOnePlusOne",
	} {
		if !strings.Contains(rendered, wanted) {
			t.Fatalf("rendered replacement missing %q:\n%s", wanted, rendered)
		}
	}
	if _, ok := ctx.imports[importCounter]; !ok {
		t.Fatal("controlled-permanent kind replacement did not request counter import")
	}
}

func TestRenderConditionForETBReplacementRejectsNegativePermanentCount(t *testing.T) {
	tests := map[string]game.Condition{
		"controller": {
			ControlsMatching: opt.Val(game.SelectionCount{MinCount: -1}),
		},
		"one opponent": {
			AnyOpponentControls: opt.Val(game.SelectionCount{MinCount: -1}),
		},
		"all opponents": {
			OpponentsControl: opt.Val(game.SelectionCount{MinCount: -1}),
		},
	}
	for name, condition := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := (Renderer{}).renderConditionForETBReplacement(&renderCtx{}, &condition); err == nil {
				t.Fatal("expected negative permanent-count threshold error")
			}
		})
	}
}

func TestRenderConditionRejectsTextWithoutPredicate(t *testing.T) {
	condition := game.Condition{Text: "some condition", Negate: true}
	renderer := Renderer{}
	ctx := &renderCtx{}

	if _, err := renderer.renderConditionForETBReplacement(ctx, &condition); err == nil {
		t.Fatal("expected ETB replacement condition without predicate to fail")
	}

	if _, err := renderer.renderStaticAbilityCondition(ctx, &condition); err == nil {
		t.Fatal("expected static ability condition without predicate to fail")
	}
}

func TestRenderLiveStateCondition(t *testing.T) {
	condition := game.Condition{
		Text:                                    "if ability-word conditions are met",
		ControllerHandEmpty:                     true,
		ControllerGraveyardCardCountAtLeast:     7,
		ControllerGraveyardCardTypeCountAtLeast: 4,
		ControllerBasicLandTypeCountAtLeast:     5,
		ControllerCreaturePowerDiversityAtLeast: 3,
		ControlsMatching: opt.Val(game.SelectionCount{
			Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
			MinCount:  3,
		}),
	}
	rendered, err := (Renderer{}).renderStaticAbilityCondition(newRenderCtx(), &condition)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"ControllerHandEmpty: true",
		"ControllerGraveyardCardCountAtLeast: 7",
		"ControllerGraveyardCardTypeCountAtLeast: 4",
		"ControllerBasicLandTypeCountAtLeast: 5",
		"ControllerCreaturePowerDiversityAtLeast: 3",
		"ControlsMatching: opt.Val",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered condition missing %q:\n%s", want, rendered)
		}
	}
}

// TestRenderUnsupportedReplacementErrors verifies the renderer returns an error
// (rather than silently omitting a field) when a CardDef contains a typed value
// the renderer cannot spell, here a non-EntersTapped replacement ability.
func TestRenderUnsupportedReplacementErrors(t *testing.T) {
	t.Parallel()
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Test",
			Types: []types.Card{types.Creature},
			ReplacementAbilities: []game.ReplacementAbility{
				{
					Text: "unsupported",
					Replacement: game.ReplacementEffect{
						EntersTapped: false,
						Condition:    opt.Val(game.Condition{Text: "some condition"}),
					},
				},
			},
		},
	}
	card := &ScryfallCard{Name: "Test", Layout: "normal", TypeLine: "Creature"}
	_, err := Renderer{}.RenderCardSource(card, []*game.CardDef{def}, []faceRenderHints{{}}, "cards")
	if err == nil {
		t.Fatal("expected error for unsupported replacement ability, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("error should mention 'unsupported', got: %v", err)
	}
}
