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

func TestRenderDynamicQuantityFieldsAndImports(t *testing.T) {
	renderer := Renderer{}
	ctx := newRenderCtx()
	rendered, err := renderer.renderQuantity(ctx, game.Dynamic(game.DynamicAmount{
		Kind:       game.DynamicAmountCountSelector,
		Multiplier: 2,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
		}),
	}))
	if err != nil {
		t.Fatal(err)
	}

	for _, wanted := range []string{
		"game.DynamicAmountCountSelector",
		"Multiplier: 2",
		"game.BattlefieldGroup",
		"types.Creature",
		"game.ControllerYou",
	} {
		if !strings.Contains(rendered, wanted) {
			t.Fatalf("quantity missing %q:\n%s", wanted, rendered)
		}
	}
	if _, ok := ctx.imports[importTypes]; !ok {
		t.Fatal("dynamic group did not request types import")
	}
	if _, ok := ctx.imports[importCounter]; ok {
		t.Fatal("dynamic group requested unused counter import")
	}

	rendered, err = renderer.renderQuantity(ctx, game.Dynamic(game.DynamicAmount{
		Kind: game.DynamicAmountControllerBasicLandTypeCount,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "game.DynamicAmountControllerBasicLandTypeCount") {
		t.Fatalf("domain quantity = %s", rendered)
	}

	ctx = newRenderCtx()
	rendered, err = renderer.renderQuantity(ctx, game.Dynamic(game.DynamicAmount{
		Kind:        game.DynamicAmountTargetCounters,
		CounterKind: counter.PlusOnePlusOne,
		Object:      game.TargetPermanentReference(0),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "counter.PlusOnePlusOne") {
		t.Fatalf("counter quantity = %s", rendered)
	}
	if _, ok := ctx.imports[importCounter]; !ok {
		t.Fatal("counter quantity did not request counter import")
	}
}

func TestRenderNamedCounterPrimitives(t *testing.T) {
	t.Parallel()
	renderer := Renderer{}
	tests := []struct {
		primitive game.Primitive
		wants     []string
	}{
		{
			primitive: game.AddCounter{
				Amount:      game.Fixed(2),
				Object:      game.TargetPermanentReference(0),
				CounterKind: counter.Stun,
			},
			wants: []string{"game.AddCounter", "game.Fixed(2)", "counter.Stun"},
		},
		{
			primitive: game.AddPlayerCounter{
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:   game.DynamicAmountObjectPower,
					Object: game.SourcePermanentReference(),
				}),
				Player:      game.TargetPlayerReference(1),
				CounterKind: counter.Poison,
			},
			wants: []string{
				"game.AddPlayerCounter",
				"game.DynamicAmountObjectPower",
				"game.SourcePermanentReference()",
				"game.TargetPlayerReference(1)",
				"counter.Poison",
			},
		},
	}
	for _, test := range tests {
		ctx := newRenderCtx()
		rendered, err := renderer.renderPrimitive(ctx, test.primitive)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range test.wants {
			if !strings.Contains(rendered, want) {
				t.Fatalf("rendered primitive missing %q:\n%s", want, rendered)
			}
		}
		if _, ok := ctx.imports[importCounter]; !ok {
			t.Fatal("counter primitive did not request counter import")
		}
	}
}

func TestRenderExplorePrimitive(t *testing.T) {
	t.Parallel()
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.Explore{
		Creature: game.SourcePermanentReference(),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"game.Explore", "Creature: game.SourcePermanentReference()"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered explore missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderCounterObjectPrimitive(t *testing.T) {
	t.Parallel()
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.CounterObject{
		Object: game.TargetStackObjectReference(0),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"game.CounterObject", "Object: game.TargetStackObjectReference(0)"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered counter missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderCreateDelayedTriggerPrimitive(t *testing.T) {
	t.Parallel()
	content := game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.Sacrifice{Object: game.SourceCardPermanentReference()},
		}},
	}.Ability()
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.CreateDelayedTrigger{
		Trigger: game.DelayedTriggerDef{
			Timing:  game.DelayedAtBeginningOfNextUpkeep,
			Content: content,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.CreateDelayedTrigger",
		"game.DelayedAtBeginningOfNextUpkeep",
		"game.Sacrifice",
		"game.SourceCardPermanentReference()",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered delayed trigger missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderTargetPredicateSpellCardTypes(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, ok, err := (Renderer{}).renderTargetPredicate(ctx, game.TargetPredicate{
		SpellCardTypes:         []types.Card{types.Instant},
		ExcludedSpellCardTypes: []types.Card{types.Creature},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("predicate rendered empty")
	}
	for _, want := range []string{
		"SpellCardTypes: []types.Card{types.Instant}",
		"ExcludedSpellCardTypes: []types.Card{types.Creature}",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered predicate missing %q:\n%s", want, rendered)
		}
	}
	if _, ok := ctx.imports[importTypes]; !ok {
		t.Fatal("spell card predicate did not request types import")
	}
}

func TestRenderManifestPrimitive(t *testing.T) {
	t.Parallel()
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.Manifest{})
	if err != nil {
		t.Fatal(err)
	}
	if rendered != "game.Manifest{}" {
		t.Fatalf("rendered manifest = %q, want game.Manifest{}", rendered)
	}

	rendered, err = (Renderer{}).renderPrimitive(newRenderCtx(), game.Manifest{Dread: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"game.Manifest", "Dread: true"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered manifest dread missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderReturnToHandAdditionalCost(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := renderAdditional(ctx, cost.Additional{
		Kind:               cost.AdditionalReturnToHand,
		Text:               "Return a tapped creature you control to its owner's hand",
		Amount:             1,
		MatchPermanentType: true,
		PermanentType:      types.Creature,
		RequireTapped:      true,
		RequireSupertype:   types.Snow,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"cost.AdditionalReturnToHand",
		"MatchPermanentType: true",
		"PermanentType: types.Creature",
		"RequireTapped: true",
		"RequireSupertype: types.Snow",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered additional missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderRevealAdditionalCostWithXAndColor(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := renderAdditional(ctx, cost.Additional{
		Kind:           cost.AdditionalReveal,
		Text:           "Reveal X blue cards from your hand",
		AmountFromX:    true,
		Source:         zone.Hand,
		MatchCardColor: true,
		CardColor:      color.Blue,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"cost.AdditionalReveal",
		"AmountFromX: true",
		"Source: zone.Hand",
		"MatchCardColor: true",
		"CardColor: color.Blue",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered additional missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderIssue210AdditionalCosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		additional cost.Additional
		wants      []string
		wantImport string
	}{
		{
			name:       "exert",
			additional: cost.Additional{Kind: cost.AdditionalExert, Text: "Exert this creature", Amount: 1},
			wants:      []string{"cost.AdditionalExert", `Text: "Exert this creature"`, "Amount: 1"},
		},
		{
			name:       "mill",
			additional: cost.Additional{Kind: cost.AdditionalMill, Text: "Mill four cards", Amount: 4},
			wants:      []string{"cost.AdditionalMill", `Text: "Mill four cards"`, "Amount: 4"},
		},
		{
			name: "put counter",
			additional: cost.Additional{
				Kind:        cost.AdditionalPutCounter,
				Text:        "Put a verse counter on Test Bard",
				Amount:      1,
				CounterKind: counter.Verse,
			},
			wants:      []string{"cost.AdditionalPutCounter", `Text: "Put a verse counter on Test Bard"`, "Amount: 1", "CounterKind: counter.Verse"},
			wantImport: importCounter,
		},
		{
			name:       "collect evidence",
			additional: cost.Additional{Kind: cost.AdditionalCollectEvidence, Text: "Collect evidence 4", Amount: 4, Source: zone.Graveyard},
			wants:      []string{"cost.AdditionalCollectEvidence", `Text: "Collect evidence 4"`, "Amount: 4", "Source: zone.Graveyard"},
			wantImport: importZone,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := newRenderCtx()
			rendered, err := renderAdditional(ctx, test.additional)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range test.wants {
				if !strings.Contains(rendered, want) {
					t.Fatalf("rendered additional missing %q:\n%s", want, rendered)
				}
			}
			if test.wantImport != "" {
				if _, ok := ctx.imports[test.wantImport]; !ok {
					t.Fatalf("rendered additional did not request import %q", test.wantImport)
				}
			}
		})
	}
}

func TestRenderEveryRecognizedCounterKind(t *testing.T) {
	t.Parallel()
	for kind := counter.PlusOnePlusOne; kind <= counter.Experience; kind++ {
		rendered, err := renderCounterKind(kind)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		if rendered == "" {
			t.Fatalf("%s rendered empty", kind)
		}
	}
}

func TestRenderReplacementAbilityRejectsMixedETBCounters(t *testing.T) {
	t.Parallel()
	ability := game.EntersTappedReplacement("This creature enters tapped with a +1/+1 counter on it.")
	ability.Replacement.EntersWithCounters = []game.CounterPlacement{{
		Kind:   counter.PlusOnePlusOne,
		Amount: 1,
	}}
	if _, err := (Renderer{}).renderReplacementAbility(newRenderCtx(), &ability); err == nil {
		t.Fatal("expected mixed ETB counter replacement to fail closed")
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

func TestRenderResolutionPaymentRejectsPromptWithoutCost(t *testing.T) {
	if _, err := (Renderer{}).renderResolutionPayment(&renderCtx{}, game.ResolutionPayment{
		Prompt: "Pay?",
	}); err == nil {
		t.Fatal("expected prompt-only payment error")
	}
}

func TestRenderResolutionPaymentPayer(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := (Renderer{}).renderResolutionPayment(ctx, game.ResolutionPayment{
		Prompt:   "Pay {2}?",
		Payer:    opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
		ManaCost: opt.Val(cost.Mana{cost.O(2)}),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Prompt: \"Pay {2}?\"",
		"Payer: opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0)))",
		"cost.O(2)",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered payment missing %q:\n%s", want, rendered)
		}
	}
	if _, ok := ctx.imports[importOpt]; !ok {
		t.Fatal("resolution payment payer did not request opt import")
	}
}

func TestRenderPayPrimitive(t *testing.T) {
	t.Parallel()
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.Pay{
		Payment: game.ResolutionPayment{
			ManaCost: opt.Val(cost.Mana{cost.U}),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"game.Pay", "Payment: game.ResolutionPayment", "cost.U"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered pay missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderInstructionEnvelope(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := (Renderer{}).renderInstruction(ctx, &game.Instruction{
		Primitive:     game.CounterObject{Object: game.TargetStackObjectReference(0)},
		PublishResult: "countered",
		ResultGate: opt.Val(game.InstructionResultGate{
			Key:       "unless-paid",
			Succeeded: game.TriFalse,
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.CounterObject",
		"PublishResult: game.ResultKey(\"countered\")",
		"ResultGate: opt.Val(game.InstructionResultGate",
		"Key: \"unless-paid\"",
		"Succeeded: game.TriFalse",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered instruction missing %q:\n%s", want, rendered)
		}
	}
	if _, ok := ctx.imports[importOpt]; !ok {
		t.Fatal("result gate did not request opt import")
	}
}

func TestRenderConditionForETBReplacementRejectsNegativePermanentCount(t *testing.T) {
	tests := map[string]game.Condition{
		"controller": {
			ControllerControls: game.PermanentFilter{MinCount: -1},
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

func TestRenderContinuousDynamicPT(t *testing.T) {
	domain := opt.Val(game.DynamicAmount{
		Kind:       game.DynamicAmountControllerBasicLandTypeCount,
		Multiplier: 1,
	})
	rendered, err := (Renderer{}).renderContinuousEffect(newRenderCtx(), &game.ContinuousEffect{
		Layer:                 game.LayerPowerToughnessModify,
		AffectedSource:        true,
		PowerDeltaDynamic:     domain,
		ToughnessDeltaDynamic: domain,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"PowerDeltaDynamic: opt.Val(game.DynamicAmount{",
		"ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{",
		"game.DynamicAmountControllerBasicLandTypeCount",
		"Multiplier: 1",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered continuous effect missing %q:\n%s", want, rendered)
		}
	}
}

// renderTestCards are representative cards exercising every lowered ability
// category through the full typed pipeline and deterministic renderer.
var renderTestCards = []*ScryfallCard{
	{
		Name:       "Render Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		ManaCost:   "{1}{G}",
		Colors:     []string{"G"},
		OracleText: "Flying\nVigilance",
		Power:      new("2"),
		Toughness:  new("2"),
	},
	{
		Name:       "Render Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {G}.",
	},
	{
		Name:       "Render Bolt",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{R}",
		Colors:     []string{"R"},
		OracleText: "Render Bolt deals 3 damage to any target.",
	},
	{
		Name:       "Render Ninja",
		Layout:     "normal",
		TypeLine:   "Creature — Human Ninja",
		ManaCost:   "{2}{U}",
		Colors:     []string{"U"},
		OracleText: "Ninjutsu {1}{U}",
		Power:      new("2"),
		Toughness:  new("2"),
	},
	{
		Name:       "Render Mutator",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		ManaCost:   "{3}{G}",
		Colors:     []string{"G"},
		OracleText: "Mutate {1}{G}\nWhenever this creature mutates, draw a card.",
		Power:      new("3"),
		Toughness:  new("3"),
	},
}

func generateExecutable(t *testing.T, card *ScryfallCard) string {
	t.Helper()
	source, diagnostics, err := GenerateExecutableCardSource(card, "cards")
	if err != nil {
		t.Fatalf("GenerateExecutableCardSource(%q): %v", card.Name, err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("GenerateExecutableCardSource(%q) diagnostics: %#v", card.Name, diagnostics)
	}
	return source
}

func TestRenderUsesEquipMechanicTemplate(t *testing.T) {
	card := &ScryfallCard{Name: "Test Equipment", Layout: "normal", TypeLine: "Artifact — Equipment"}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:               card.Name,
		Types:              []types.Card{types.Artifact},
		Subtypes:           []types.Sub{types.Equipment},
		ActivatedAbilities: []game.ActivatedAbility{game.EquipActivatedAbility(cost.Mana{cost.O(2)})},
	}}

	source, err := (Renderer{}).RenderCardSource(card, []*game.CardDef{def}, []faceRenderHints{{}}, "cards")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "game.EquipActivatedAbility(cost.Mana{cost.O(2)})") {
		t.Fatalf("source does not use Equip template:\n%s", source)
	}
}

func TestRenderUsesEnchantMechanicTemplate(t *testing.T) {
	target := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "creature",
		Allow:      game.TargetAllowPermanent,
		Predicate: game.TargetPredicate{
			PermanentTypes: []types.Card{types.Creature},
		},
	}
	card := &ScryfallCard{Name: "Test Aura", Layout: "normal", TypeLine: "Enchantment — Aura"}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:            card.Name,
		Types:           []types.Card{types.Enchantment},
		Subtypes:        []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{game.EnchantStaticAbility(&target)},
	}}

	source, err := (Renderer{}).RenderCardSource(card, []*game.CardDef{def}, []faceRenderHints{{}}, "cards")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "game.EnchantStaticAbility(&game.TargetSpec{") {
		t.Fatalf("source does not use Enchant template:\n%s", source)
	}
}

func TestRenderTargetPredicateQualifiers(t *testing.T) {
	ctx := newRenderCtx()
	lit, ok, err := (Renderer{}).renderTargetPredicate(ctx, game.TargetPredicate{
		PermanentTypes:  []types.Card{types.Creature},
		ExcludedTypes:   []types.Card{types.Artifact},
		Colors:          []color.Color{color.Green},
		ExcludedColors:  []color.Color{color.Blue},
		Controller:      game.ControllerYou,
		Tapped:          game.TriTrue,
		CombatState:     game.CombatStateAttacking,
		Keyword:         game.Flying,
		ExcludedKeyword: game.Deathtouch,
		ManaValue: opt.Val(compare.Int{
			Op:    compare.LessOrEqual,
			Value: 3,
		}),
		Power: opt.Val(compare.Int{
			Op:    compare.GreaterThan,
			Value: 1,
		}),
		Toughness: opt.Val(compare.Int{
			Op:    compare.Equal,
			Value: 2,
		}),
		Another: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("renderTargetPredicate() did not render qualified predicate")
	}
	if _, ok := ctx.imports[importOpt]; !ok {
		t.Fatal("renderTargetPredicate() did not request opt import")
	}
	for _, want := range []string{
		"ExcludedTypes: []types.Card{types.Artifact}",
		"Colors: []color.Color{color.Green}",
		"ExcludedColors: []color.Color{color.Blue}",
		"Controller: game.ControllerYou",
		"Tapped: game.TriTrue",
		"CombatState: game.CombatStateAttacking",
		"Keyword: game.Flying",
		"ExcludedKeyword: game.Deathtouch",
		"ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})",
		"Power: opt.Val(compare.Int{Op: compare.GreaterThan, Value: 1})",
		"Toughness: opt.Val(compare.Int{Op: compare.Equal, Value: 2})",
		"Another: true",
	} {
		if !strings.Contains(lit, want) {
			t.Fatalf("predicate literal %q does not contain %q", lit, want)
		}
	}
}

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

func TestRenderUsesProtectionMechanicTemplate(t *testing.T) {
	card := &ScryfallCard{Name: "Test Bear", Layout: "normal", TypeLine: "Creature — Bear"}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  card.Name,
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{
			game.ProtectionFromColorsStaticAbility(color.Black, color.Red),
		},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}

	source, err := (Renderer{}).RenderCardSource(card, []*game.CardDef{def}, []faceRenderHints{{}}, "cards")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "game.ProtectionFromColorsStaticAbility(color.Black, color.Red)") {
		t.Fatalf("source does not use Protection template:\n%s", source)
	}
}

func TestRenderDeterministic(t *testing.T) {
	t.Parallel()
	for _, card := range renderTestCards {
		t.Run(card.Name, func(t *testing.T) {
			t.Parallel()
			first := generateExecutable(t, card)
			for i := range 5 {
				again := generateExecutable(t, card)
				if again != first {
					t.Fatalf("render not deterministic on iteration %d", i)
				}
			}
		})
	}
}

func TestRenderParses(t *testing.T) {
	t.Parallel()
	for _, card := range renderTestCards {
		t.Run(card.Name, func(t *testing.T) {
			t.Parallel()
			source := generateExecutable(t, card)
			if _, err := parser.ParseFile(token.NewFileSet(), "card.go", source, parser.AllErrors); err != nil {
				t.Fatalf("generated source does not parse: %v\n%s", err, source)
			}
		})
	}
}

func TestRenderNoTODO(t *testing.T) {
	t.Parallel()
	for _, card := range renderTestCards {
		t.Run(card.Name, func(t *testing.T) {
			t.Parallel()
			source := generateExecutable(t, card)
			if strings.Contains(source, "TODO") {
				t.Fatalf("executable source unexpectedly contains TODO:\n%s", source)
			}
		})
	}
}

func TestRenderImportsDeterministicOrder(t *testing.T) {
	t.Parallel()
	source := generateExecutable(t, renderTestCards[1])
	start := strings.Index(source, "import (")
	if start < 0 {
		t.Fatalf("no import block found:\n%s", source)
	}
	end := strings.Index(source[start:], ")")
	block := source[start : start+end]
	var paths []string
	for line := range strings.SplitSeq(block, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, `"`) {
			paths = append(paths, line)
		}
	}
	for i := 1; i < len(paths); i++ {
		if paths[i-1] > paths[i] {
			t.Fatalf("imports not sorted: %q before %q", paths[i-1], paths[i])
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

func TestRenderUnsupportedAbilityLayerFieldsErrors(t *testing.T) {
	t.Parallel()
	tests := map[string]game.ContinuousEffect{
		"unsupported field": {
			Layer:          game.LayerAbility,
			Group:          game.BattlefieldGroup(game.Selection{}),
			RemoveKeywords: []game.Keyword{game.Flying},
		},
		"PT field in ability layer": {
			Layer:      game.LayerAbility,
			Group:      game.BattlefieldGroup(game.Selection{}),
			PowerDelta: 1,
		},
		"dynamic PT field in ability layer": {
			Layer:             game.LayerAbility,
			Group:             game.BattlefieldGroup(game.Selection{}),
			PowerDeltaDynamic: opt.Val(game.DynamicAmount{Kind: game.DynamicAmountOpponentCount}),
		},
		"keyword field in PT layer": {
			Layer:       game.LayerPowerToughnessModify,
			Group:       game.BattlefieldGroup(game.Selection{}),
			AddKeywords: []game.Keyword{game.Flying},
		},
		"source and group recipients": {
			Layer:          game.LayerAbility,
			AffectedSource: true,
			Group:          game.BattlefieldGroup(game.Selection{}),
			AddKeywords:    []game.Keyword{game.Flying},
		},
	}
	for name, effect := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			def := &game.CardDef{
				CardFace: game.CardFace{
					Name:  "Test",
					Types: []types.Card{types.Enchantment},
					StaticAbilities: []game.StaticAbility{{
						ContinuousEffects: []game.ContinuousEffect{effect},
					}},
				},
			}
			card := &ScryfallCard{Name: "Test", Layout: "normal", TypeLine: "Enchantment"}
			_, err := Renderer{}.RenderCardSource(card, []*game.CardDef{def}, []faceRenderHints{{}}, "cards")
			if err == nil {
				t.Fatal("expected error for incompatible continuous-effect fields")
			}
		})
	}
}

// TestRenderHintDivergenceErrors verifies the renderer refuses to use a
// static-ability VarName hint whose recorded body diverges from the validated
// CardDef value, returning a divergence error instead of emitting a wrong var.
func TestRenderHintDivergenceErrors(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name: "Test Bear", Layout: "normal", TypeLine: "Creature — Bear",
		OracleText: "Flying", Power: new("2"), Toughness: new("2"),
	}
	faceAbilities, diagnostics := lowerExecutableFaces(card)
	if len(diagnostics) != 0 {
		t.Fatalf("lowering diagnostics: %v", diagnostics)
	}
	defs, err := assembleCardDefs(card, faceAbilities)
	if err != nil {
		t.Fatalf("assembleCardDefs: %v", err)
	}
	hints := []faceRenderHints{{
		StaticVarNames: []staticVarHint{{
			VarName: "game.FlyingStaticBody",
			Body:    game.VigilanceStaticBody,
		}},
	}}
	_, err = Renderer{}.RenderCardSource(card, defs, hints, "cards")
	if err == nil {
		t.Fatal("expected error for hint body divergence, got nil")
	}
	if !strings.Contains(err.Error(), "divergence") {
		t.Fatalf("error should mention 'divergence', got: %v", err)
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
			Text:                      "As long as you have seven or more cards in hand",
			ControllerHandSizeAtLeast: 7,
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
		"ControllerHandSizeAtLeast: 7",
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
