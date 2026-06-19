package cardgen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

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

func TestRenderShuffleRevealLinkedPutSequence(t *testing.T) {
	t.Parallel()
	key := game.LinkedKey("revealed-card")
	owner := game.ObjectOwnerReference(game.TargetPermanentReference(0))
	sequence := []game.Instruction{
		{Primitive: game.ShufflePermanentIntoLibrary{Object: game.TargetPermanentReference(0)}},
		{Primitive: game.Reveal{
			Amount:        game.Fixed(1),
			Player:        owner,
			PublishLinked: key,
		}},
		{
			Primitive: game.PutOnBattlefield{
				Source:    game.LinkedBattlefieldSource(key),
				Recipient: opt.Val(owner),
			},
			CardCondition: opt.Val(game.CardCondition{
				Card:                 game.CardReference{Kind: game.CardReferenceLinked, LinkID: string(key)},
				RequirePermanentCard: true,
			}),
		},
	}
	renderer := Renderer{}
	ctx := newRenderCtx()
	var rendered []string
	for i := range sequence {
		value, err := renderer.renderInstruction(ctx, &sequence[i])
		if err != nil {
			t.Fatal(err)
		}
		rendered = append(rendered, value)
	}
	joined := strings.Join(rendered, "\n")
	for _, want := range []string{
		"game.ShufflePermanentIntoLibrary",
		"Object: game.TargetPermanentReference(0)",
		"game.Reveal",
		"Amount: game.Fixed(1)",
		"Player: game.ObjectOwnerReference(game.TargetPermanentReference(0))",
		`PublishLinked: game.LinkedKey("revealed-card")`,
		"game.LinkedBattlefieldSource(game.LinkedKey(\"revealed-card\"))",
		"CardCondition: opt.Val(game.CardCondition",
		"RequirePermanentCard: true",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("rendered sequence missing %q:\n%s", want, joined)
		}
	}
	if _, ok := ctx.imports[importOpt]; !ok {
		t.Fatal("linked put did not request opt import")
	}
}

func TestRenderSearchPrimitive(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := (Renderer{}).renderPrimitive(ctx, game.Search{
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:   zone.Library,
			Destination:  zone.Battlefield,
			CardType:     opt.Val(types.Land),
			Supertype:    opt.Val(types.Basic),
			SubtypesAny:  []types.Sub{types.Forest, types.Plains},
			Reveal:       true,
			EntersTapped: true,
		},
		Amount: game.Fixed(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.Search",
		"Player: game.ControllerReference()",
		"SourceZone: zone.Library",
		"Destination: zone.Battlefield",
		"CardType: opt.Val(types.Land)",
		"Supertype: opt.Val(types.Basic)",
		"SubtypesAny: []types.Sub{types.Forest, types.Plains}",
		"Reveal: true",
		"EntersTapped: true",
		"Amount: game.Fixed(1)",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered search missing %q:\n%s", want, rendered)
		}
	}
	for _, requiredImport := range []string{importZone, importTypes, importOpt} {
		if _, ok := ctx.imports[requiredImport]; !ok {
			t.Fatalf("search primitive did not request import %q", requiredImport)
		}
	}
}

func TestRenderSearchPrimitivePermanentManaValue(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := (Renderer{}).renderPrimitive(ctx, game.Search{
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:   zone.Library,
			Destination:  zone.Battlefield,
			Permanent:    true,
			SubtypesAny:  []types.Sub{types.Rebel},
			MaxManaValue: opt.Val(5),
		},
		Amount: game.Fixed(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Permanent: true",
		"SubtypesAny: []types.Sub{types.Rebel}",
		"MaxManaValue: opt.Val(5)",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered search missing %q:\n%s", want, rendered)
		}
	}
	if _, ok := ctx.imports[importOpt]; !ok {
		t.Fatal("search primitive with a mana-value bound did not request the opt import")
	}
}

// TestRenderSearchPrimitiveSplitDestination verifies the SplitDestination slot
// of a split-destination land tutor renders as an opt-wrapped game.SearchDestination
// literal carrying its secondary zone and tapped flag.
func TestRenderSearchPrimitiveSplitDestination(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := (Renderer{}).renderPrimitive(ctx, game.Search{
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:       zone.Library,
			Destination:      zone.Battlefield,
			CardType:         opt.Val(types.Land),
			Supertype:        opt.Val(types.Basic),
			Reveal:           true,
			EntersTapped:     true,
			SplitDestination: opt.Val(game.SearchDestination{Zone: zone.Hand}),
		},
		Amount: game.Fixed(2),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Destination: zone.Battlefield",
		"EntersTapped: true",
		"SplitDestination: opt.Val(game.SearchDestination{",
		"Zone: zone.Hand",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered split search missing %q:\n%s", want, rendered)
		}
	}
	for _, requiredImport := range []string{importZone, importOpt} {
		if _, ok := ctx.imports[requiredImport]; !ok {
			t.Fatalf("split search primitive did not request import %q", requiredImport)
		}
	}
}

// TestRenderSearchPrimitiveSplitDestinationTapped verifies a tapped secondary
// battlefield slot renders its EntersTapped flag inside the SearchDestination
// literal.
func TestRenderSearchPrimitiveSplitDestinationTapped(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := (Renderer{}).renderPrimitive(ctx, game.Search{
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:       zone.Library,
			Destination:      zone.Hand,
			SplitDestination: opt.Val(game.SearchDestination{Zone: zone.Battlefield, EntersTapped: true}),
		},
		Amount: game.Fixed(2),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"SplitDestination: opt.Val(game.SearchDestination{",
		"Zone: zone.Battlefield",
		"EntersTapped: true",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered split search missing %q:\n%s", want, rendered)
		}
	}
}

// TestRenderSearchPrimitiveSharedSubtype verifies the SharedSubtype correlation
// flag of a "that share a land type" search renders as a plain boolean field on
// the SearchSpec literal (Myriad Landscape).
func TestRenderSearchPrimitiveSharedSubtype(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := (Renderer{}).renderPrimitive(ctx, game.Search{
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:    zone.Library,
			Destination:   zone.Battlefield,
			CardType:      opt.Val(types.Land),
			Supertype:     opt.Val(types.Basic),
			EntersTapped:  true,
			SharedSubtype: true,
		},
		Amount: game.Fixed(2),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Destination: zone.Battlefield",
		"EntersTapped: true",
		"SharedSubtype: true",
		"Amount: game.Fixed(2)",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered shared-subtype search missing %q:\n%s", want, rendered)
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

func TestRenderDelayedBoundedDrawPrimitive(t *testing.T) {
	t.Parallel()
	key := game.ChoiceKey("draw-count")
	targetController := game.CapturedTargetControllerReference(0)
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.CreateDelayedTrigger{
		Trigger: game.DelayedTriggerDef{
			Timing: game.DelayedAtBeginningOfNextUpkeep,
			Content: game.Mode{Sequence: []game.Instruction{
				{
					Primitive: game.Choose{
						Choice: game.ResolutionChoice{
							Kind:            game.ResolutionChoiceNumber,
							PlayerReference: &targetController,
							MinNumber:       0,
							MaxNumber:       2,
						},
						PublishChoice: key,
					},
				},
				{
					Primitive: game.Draw{
						Amount: game.Dynamic(game.DynamicAmount{
							Kind:      game.DynamicAmountChosenNumber,
							ResultKey: game.ResultKey(key),
						}),
						Player: game.CapturedTargetControllerReference(0),
					},
				},
			}}.Ability(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"PlayerReference: func() *game.PlayerReference { ref := game.CapturedTargetControllerReference(0); return &ref }()",
		"game.ResolutionChoiceNumber",
		"MinNumber: 0",
		"MaxNumber: 2",
		`PublishChoice: game.ChoiceKey("draw-count")`,
		"game.DynamicAmountChosenNumber",
		`ResultKey: game.ResultKey("draw-count")`,
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered delayed bounded draw missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderLinkedExilePrimitive(t *testing.T) {
	t.Parallel()
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.Exile{
		Object:         game.TargetPermanentReference(1),
		ExileLinkedKey: game.LinkedKey("delayed-blink-1"),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.Exile",
		"game.TargetPermanentReference(1)",
		`ExileLinkedKey: game.LinkedKey("delayed-blink-1")`,
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered linked exile missing %q:\n%s", want, rendered)
		}
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

func TestRenderResolutionPaymentEventPlayer(t *testing.T) {
	t.Parallel()
	rendered, err := (Renderer{}).renderResolutionPayment(newRenderCtx(), game.ResolutionPayment{
		Prompt:   "Pay {1}?",
		Payer:    opt.Val(game.EventPlayerReference()),
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "Payer: opt.Val(game.EventPlayerReference())") {
		t.Fatalf("rendered payment missing event player:\n%s", rendered)
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

func TestRenderSacrificePermanentsPrimitive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		primitive  game.SacrificePermanents
		wantSubstr string
	}{
		{
			name: "target player creature",
			primitive: game.SacrificePermanents{
				Player:    game.TargetPlayerReference(0),
				Amount:    game.Fixed(1),
				Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
			wantSubstr: "game.SacrificePermanents",
		},
		{
			name: "opponents reference any permanent",
			primitive: game.SacrificePermanents{
				PlayerGroup: game.OpponentsReference(),
				Amount:      game.Fixed(1),
			},
			wantSubstr: "game.OpponentsReference()",
		},
		{
			name: "all players land",
			primitive: game.SacrificePermanents{
				PlayerGroup: game.AllPlayersReference(),
				Amount:      game.Fixed(1),
				Selection:   game.Selection{RequiredTypes: []types.Card{types.Land}},
			},
			wantSubstr: "game.AllPlayersReference()",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := newRenderCtx()
			rendered, err := (Renderer{}).renderPrimitive(ctx, tt.primitive)
			if err != nil {
				t.Fatalf("renderPrimitive() error = %v", err)
			}
			if !strings.Contains(rendered, tt.wantSubstr) {
				t.Fatalf("rendered = %q, want substring %q", rendered, tt.wantSubstr)
			}
			// Verify rendered Go is syntactically valid.
			src := "package p\nvar _ = " + rendered
			if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
				t.Fatalf("rendered output is not valid Go: %v\n%s", err, rendered)
			}
		})
	}
}

func TestRenderApplyRulePrimitiveCantBeBlockedThisTurn(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := (Renderer{}).renderPrimitive(ctx, game.ApplyRule{
		Object: opt.Val(game.TargetPermanentReference(0)),
		RuleEffects: []game.RuleEffect{
			{Kind: game.RuleEffectCantBeBlocked},
		},
		Duration: game.DurationThisTurn,
	})
	if err != nil {
		t.Fatalf("renderPrimitive() error = %v", err)
	}
	for _, want := range []string{
		"game.ApplyRule{",
		"Object: opt.Val(game.TargetPermanentReference(0)),",
		"RuleEffects: []game.RuleEffect{",
		"Kind: game.RuleEffectCantBeBlocked,",
		"Duration: game.DurationThisTurn,",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered ApplyRule missing %q:\n%s", want, rendered)
		}
	}
	if _, ok := ctx.imports[importOpt]; !ok {
		t.Fatal("ApplyRule object did not request opt import")
	}
	src := "package p\nvar _ = " + rendered
	if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
		t.Fatalf("rendered output is not valid Go: %v\n%s", err, rendered)
	}
}

// TestRenderMoveCardPlayerZoneGroup renders the player-zone group form of
// MoveCard ("Exile target player's graveyard.") with its target-player
// reference rather than a card reference.
func TestRenderMoveCardPlayerZoneGroup(t *testing.T) {
	t.Parallel()
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.MoveCard{
		Player:      game.TargetPlayerReference(0),
		FromZone:    zone.Graveyard,
		Destination: zone.Exile,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.MoveCard",
		"Player: game.TargetPlayerReference(0),",
		"FromZone: zone.Graveyard,",
		"Destination: zone.Exile,",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered move card missing %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "Card:") {
		t.Fatalf("player-zone move must not render a Card field:\n%s", rendered)
	}
}

// TestRenderMoveCardSingleCard keeps the single-card form rendering its card
// reference, guarding against the player-zone branch leaking into it.
func TestRenderMoveCardSingleCard(t *testing.T) {
	t.Parallel()
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.MoveCard{
		Card:        game.CardReference{Kind: game.CardReferenceTarget},
		FromZone:    zone.Graveyard,
		Destination: zone.Exile,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.MoveCard",
		"Card: game.CardReference{Kind: game.CardReferenceTarget}",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered move card missing %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "Player:") {
		t.Fatalf("single-card move must not render a Player field:\n%s", rendered)
	}
}
