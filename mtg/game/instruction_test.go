package game

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type instructionKeyTypeCheck struct {
	Result ResultKey
	Choice ChoiceKey
	Linked LinkedKey
}

func TestInstructionKeyTypesAreDistinct(t *testing.T) {
	var _ instructionKeyTypeCheck
	if reflect.TypeFor[ResultKey]() == reflect.TypeFor[ChoiceKey]() {
		t.Fatal("ResultKey and ChoiceKey must be distinct types")
	}
	if reflect.TypeFor[ResultKey]() == reflect.TypeFor[LinkedKey]() {
		t.Fatal("ResultKey and LinkedKey must be distinct types")
	}
	if reflect.TypeFor[ChoiceKey]() == reflect.TypeFor[LinkedKey]() {
		t.Fatal("ChoiceKey and LinkedKey must be distinct types")
	}
}

func TestValidateInstructionSequenceAcceptsLinkedBattlefieldSource(t *testing.T) {
	seq := []Instruction{
		{
			Primitive: Reveal{
				Amount:        Fixed(1),
				Player:        ControllerReference(),
				PublishLinked: LinkedKey("revealed-card"),
			},
		},
		{
			Primitive: PutOnBattlefield{
				Source: LinkedBattlefieldSource(LinkedKey("revealed-card")),
			},
			CardCondition: opt.Val(CardCondition{
				Card:                 CardReference{Kind: CardReferenceLinked, LinkID: "revealed-card"},
				RequirePermanentCard: true,
			}),
		},
	}

	if err := ValidateInstructionSequence(seq); err != nil {
		t.Fatalf("ValidateInstructionSequence() error = %v, want nil", err)
	}
}

func TestValidateInstructionSequenceAcceptsPublishedReanimatedPermanent(t *testing.T) {
	linkedKey := LinkedKey("reanimated-card")
	resultKey := ResultKey("reanimation-move")
	seq := []Instruction{
		{
			Primitive: PutOnBattlefield{
				Source:        CardBattlefieldSource(CardReference{Kind: CardReferenceTarget}),
				PublishLinked: linkedKey,
			},
			PublishResult: resultKey,
		},
		{
			Primitive: LoseLife{
				Amount: Dynamic(DynamicAmount{
					Kind:   DynamicAmountObjectManaValue,
					Object: LinkedObjectReference(string(linkedKey)),
				}),
				Player: ControllerReference(),
			},
			ResultGate: opt.Val(InstructionResultGate{Key: resultKey, Succeeded: TriTrue}),
		},
	}

	targets := []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowCard,
		TargetZone: zone.Graveyard,
	}}
	if err := ValidateInstructionSequence(seq, targets); err != nil {
		t.Fatalf("ValidateInstructionSequence() error = %v, want nil", err)
	}
}

func TestValidateInstructionSequenceRejectsRepublishingLinkedBattlefieldSource(t *testing.T) {
	seq := []Instruction{
		{Primitive: Reveal{
			Amount:        Fixed(1),
			Player:        ControllerReference(),
			PublishLinked: "revealed-card",
		}},
		{Primitive: PutOnBattlefield{
			Source:        LinkedBattlefieldSource("revealed-card"),
			PublishLinked: "entered-card",
		}},
	}

	err := ValidateInstructionSequence(seq)
	if err == nil || !strings.Contains(err.Error(), "can publish only a referenced card") {
		t.Fatalf("ValidateInstructionSequence() error = %v", err)
	}
}

func TestValidateInstructionSequenceAcceptsLinkedSearchResult(t *testing.T) {
	key := LinkedKey("searched-land")
	seq := []Instruction{
		{Primitive: Search{
			Player: ControllerReference(),
			Spec: SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Battlefield,
				CardType:    opt.Val(types.Land),
			},
			Amount:        Fixed(1),
			PublishLinked: key,
		}},
		{Primitive: Untap{Object: LinkedObjectReference(string(key))}},
	}
	if err := ValidateInstructionSequence(seq); err != nil {
		t.Fatalf("ValidateInstructionSequence() error = %v, want nil", err)
	}
}

func TestValidateInstructionSequenceAcceptsBoundedGroupUntap(t *testing.T) {
	seq := []Instruction{{Primitive: Untap{
		Group: BattlefieldGroup(Selection{
			RequiredTypes: []types.Card{types.Land},
		}),
		ChooseUpTo: true,
		Amount:     Fixed(3),
	}}}
	if err := ValidateInstructionSequence(seq); err != nil {
		t.Fatalf("ValidateInstructionSequence() error = %v, want nil", err)
	}
}

func TestValidateInstructionSequenceRejectsMalformedBoundedUntap(t *testing.T) {
	for _, primitive := range []Untap{
		{Object: TargetPermanentReference(0), ChooseUpTo: true, Amount: Fixed(3)},
		{Group: BattlefieldGroup(Selection{RequiredTypes: []types.Card{types.Land}}), ChooseUpTo: true},
		{Group: BattlefieldGroup(Selection{RequiredTypes: []types.Card{types.Land}}), Amount: Fixed(3)},
	} {
		if err := ValidateInstructionSequence([]Instruction{{Primitive: primitive}}); err == nil {
			t.Fatalf("ValidateInstructionSequence(%#v) error = nil, want failure", primitive)
		}
	}
}

func TestValidateInstructionSequenceRejectsUnknownOrForwardLinkedUntap(t *testing.T) {
	key := LinkedKey("searched-land")
	for _, seq := range [][]Instruction{
		{{Primitive: Untap{Object: LinkedObjectReference(string(key))}}},
		{
			{Primitive: Untap{Object: LinkedObjectReference(string(key))}},
			{Primitive: Search{
				Player: ControllerReference(),
				Spec: SearchSpec{
					SourceZone:  zone.Library,
					Destination: zone.Battlefield,
				},
				Amount:        Fixed(1),
				PublishLinked: key,
			}},
		},
	} {
		err := ValidateInstructionSequence(seq)
		if err == nil || !strings.Contains(err.Error(), `linked key "searched-land" not yet published`) {
			t.Errorf("ValidateInstructionSequence(%#v) error = %v, want linked-key validation failure", seq, err)
		}
	}
}

func TestValidateInstructionSequenceRejectsInvalidLinkedSearch(t *testing.T) {
	tests := []Search{
		{
			Player:        ControllerReference(),
			Spec:          SearchSpec{SourceZone: zone.Library, Destination: zone.Hand},
			Amount:        Fixed(1),
			PublishLinked: "searched-card",
		},
		{
			Player:        ControllerReference(),
			Spec:          SearchSpec{SourceZone: zone.Library, Destination: zone.Battlefield},
			Amount:        Fixed(2),
			PublishLinked: "searched-cards",
		},
	}
	for _, search := range tests {
		err := ValidateInstructionSequence([]Instruction{{Primitive: search}})
		if err == nil || !strings.Contains(err.Error(), "linked search requires exactly one card moved to the battlefield") {
			t.Errorf("Search %#v validation error = %v", search, err)
		}
	}
}

func TestValidateInstructionSequenceAcceptsDelayedLinkedBattlefieldSource(t *testing.T) {
	key := LinkedKey("delayed-blink")
	seq := []Instruction{
		{Primitive: Exile{Object: TargetPermanentReference(0), ExileLinkedKey: key}},
		{Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
			Timing: DelayedAtBeginningOfNextEndStep,
			Content: Mode{Sequence: []Instruction{{Primitive: PutOnBattlefield{
				Source: LinkedBattlefieldSource(key),
			}}}}.Ability(),
		}}},
	}
	if err := ValidateInstructionSequence(seq, []TargetSpec{{MinTargets: 1, MaxTargets: 1}}); err != nil {
		t.Fatalf("ValidateInstructionSequence() error = %v, want nil", err)
	}
}

func TestValidateInstructionSequenceRejectsDelayedUnknownLinkedBattlefieldSource(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{{Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
		Timing: DelayedAtBeginningOfNextEndStep,
		Content: Mode{Sequence: []Instruction{{Primitive: PutOnBattlefield{
			Source: LinkedBattlefieldSource("missing"),
		}}}}.Ability(),
	}}}})
	if err == nil || !strings.Contains(err.Error(), `linked key "missing" not yet published`) {
		t.Fatalf("error = %v, want linked-key validation failure", err)
	}
}

func TestValidateInstructionSequenceAcceptsDelayedLinkedObject(t *testing.T) {
	key := LinkedKey("delayed-target")
	err := ValidateInstructionSequence([]Instruction{
		{Primitive: ModifyPT{
			Object:         TargetPermanentReference(0),
			PowerDelta:     Fixed(2),
			ToughnessDelta: Fixed(2),
			Duration:       DurationUntilEndOfTurn,
			PublishLinked:  key,
		}},
		{Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
			Timing: DelayedAtBeginningOfNextEndStep,
			Content: Mode{Sequence: []Instruction{{Primitive: Bounce{
				Object: LinkedObjectReference(string(key)),
			}}}}.Ability(),
		}}},
	}, []TargetSpec{{MinTargets: 1, MaxTargets: 1}})
	if err != nil {
		t.Fatalf("ValidateInstructionSequence() error = %v", err)
	}
}

func TestValidateInstructionSequenceRejectsDelayedUnknownLinkedObject(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{{Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
		Timing: DelayedAtBeginningOfNextEndStep,
		Content: Mode{Sequence: []Instruction{{Primitive: Bounce{
			Object: LinkedObjectReference("missing"),
		}}}}.Ability(),
	}}}})
	if err == nil {
		t.Fatal("ValidateInstructionSequence() accepted unknown delayed linked object")
	}
}

func TestValidateInstructionSequenceAcceptsDelayedBoundedDrawChoice(t *testing.T) {
	key := ChoiceKey("draw-count")
	targetController := CapturedTargetControllerReference(0)
	err := ValidateInstructionSequence([]Instruction{{Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
		Timing: DelayedAtBeginningOfNextUpkeep,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:            ResolutionChoiceNumber,
						PlayerReference: &targetController,
						MinNumber:       0,
						MaxNumber:       2,
					},
					PublishChoice: key,
				},
			},
			{
				Primitive: Draw{
					Amount: Dynamic(DynamicAmount{
						Kind:      DynamicAmountChosenNumber,
						ResultKey: ResultKey(key),
					}),
					Player: CapturedTargetControllerReference(0),
				},
			},
		}}.Ability(),
	}}}}, []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowStackObject,
	}})
	if err != nil {
		t.Fatalf("ValidateInstructionSequence() error = %v", err)
	}
}

func TestValidateInstructionSequenceRejectsDelayedCapturedTargetControllerOutOfRange(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{{Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
		Timing: DelayedAtBeginningOfNextUpkeep,
		Content: Mode{Sequence: []Instruction{{Primitive: Draw{
			Amount: Fixed(1),
			Player: CapturedTargetControllerReference(1),
		}}}}.Ability(),
	}}}}, []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowStackObject,
	}})
	if err == nil || !strings.Contains(err.Error(), "target index 1 has no matching target specification") {
		t.Fatalf("error = %v, want captured target bounds failure", err)
	}
}

func TestValidateInstructionSequenceRejectsCapturedTargetControllerInOtherPlayerFields(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{{Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
		Timing: DelayedAtBeginningOfNextUpkeep,
		Content: Mode{Sequence: []Instruction{{Primitive: GainLife{
			Amount: Fixed(1),
			Player: CapturedTargetControllerReference(1),
		}}}}.Ability(),
	}}}}, []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowStackObject,
	}})
	if err == nil || !strings.Contains(err.Error(), "target index 1 has no matching target specification") {
		t.Fatalf("error = %v, want captured target bounds failure", err)
	}
}

func TestValidateInstructionSequenceRejectsCapturedControllerFromNonStackTarget(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{{Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
		Timing: DelayedAtBeginningOfNextUpkeep,
		Content: Mode{Sequence: []Instruction{{Primitive: Draw{
			Amount: Fixed(1),
			Player: CapturedTargetControllerReference(0),
		}}}}.Ability(),
	}}}}, []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowPlayer,
	}})
	if err == nil || !strings.Contains(err.Error(), "incompatible target kind") {
		t.Fatalf("error = %v, want captured stack-target validation failure", err)
	}
}

func TestValidateInstructionSequenceDoesNotUseEnclosingTargetsForDelayedTargetReferences(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{{Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
		Timing: DelayedAtBeginningOfNextUpkeep,
		Content: Mode{Sequence: []Instruction{{Primitive: Draw{
			Amount: Fixed(1),
			Player: TargetPlayerReference(0),
		}}}}.Ability(),
	}}}}, []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowPlayer,
	}})
	if err == nil || !strings.Contains(err.Error(), "target index 0 has no matching target specification") {
		t.Fatalf("error = %v, want delayed local target bounds failure", err)
	}
}

func TestValidateInstructionSequenceRejectsInvalidNumberChoice(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{{
		Primitive: Choose{Choice: ResolutionChoice{
			Kind:      ResolutionChoiceNumber,
			MinNumber: 2,
			MaxNumber: 1,
		}},
	}})
	if err == nil || !strings.Contains(err.Error(), "nonnegative inclusive range") {
		t.Fatalf("error = %v, want bounded-number validation failure", err)
	}
}

func TestValidateInstructionSequenceAcceptsLinkedCardConsumers(t *testing.T) {
	for _, primitive := range []Primitive{
		MoveCard{
			Card:        CardReference{Kind: CardReferenceLinked, LinkID: "revealed-card"},
			FromZone:    zone.Graveyard,
			Destination: zone.Hand,
		},
		GrantCastPermission{
			Card:     CardReference{Kind: CardReferenceLinked, LinkID: "revealed-card"},
			FromZone: zone.Graveyard,
			Face:     FaceAlternate,
			Duration: DurationUntilEndOfYourNextTurn,
		},
	} {
		seq := []Instruction{
			{Primitive: Reveal{
				Amount:        Fixed(1),
				Player:        ControllerReference(),
				PublishLinked: LinkedKey("revealed-card"),
			}},
			{Primitive: primitive},
		}
		if err := ValidateInstructionSequence(seq); err != nil {
			t.Errorf("%T linked consumer: %v", primitive, err)
		}
	}
}

func TestValidateInstructionSequenceRejectsUnknownOrForwardLinkedCardConsumers(t *testing.T) {
	for _, primitive := range []Primitive{
		MoveCard{
			Card:        CardReference{Kind: CardReferenceLinked, LinkID: "later"},
			FromZone:    zone.Graveyard,
			Destination: zone.Hand,
		},
		GrantCastPermission{
			Card:     CardReference{Kind: CardReferenceLinked, LinkID: "later"},
			FromZone: zone.Graveyard,
			Face:     FaceAlternate,
			Duration: DurationUntilEndOfYourNextTurn,
		},
	} {
		seq := []Instruction{
			{Primitive: primitive},
			{Primitive: Reveal{
				Amount:        Fixed(1),
				Player:        ControllerReference(),
				PublishLinked: LinkedKey("later"),
			}},
		}
		err := ValidateInstructionSequence(seq)
		if err == nil || !strings.Contains(err.Error(), `linked key "later" not yet published`) {
			t.Errorf("%T error = %v, want linked-key validation failure", primitive, err)
		}
	}
}

func TestValidateInstructionSequenceDoesNotTreatNonLinkedCardReferencesAsLinked(t *testing.T) {
	for _, reference := range []CardReference{
		{Kind: CardReferenceSource},
		{Kind: CardReferenceEvent},
	} {
		for _, primitive := range []Primitive{
			MoveCard{
				Card:        reference,
				FromZone:    zone.Graveyard,
				Destination: zone.Hand,
			},
			GrantCastPermission{
				Card:     reference,
				FromZone: zone.Graveyard,
				Face:     FaceAlternate,
				Duration: DurationUntilEndOfYourNextTurn,
			},
		} {
			if err := ValidateInstructionSequence([]Instruction{{Primitive: primitive}}); err != nil {
				t.Errorf("%T with card reference %v: %v", primitive, reference.Kind, err)
			}
		}
	}
}

func TestValidateInstructionSequenceRejectsSameZoneMoveCard(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{{Primitive: MoveCard{
		Card:        CardReference{Kind: CardReferenceEvent},
		FromZone:    zone.Graveyard,
		Destination: zone.Graveyard,
	}}})
	if err == nil || !strings.Contains(err.Error(), "different source and destination zones") {
		t.Fatalf("error = %v, want same-zone move validation failure", err)
	}
}

func TestValidateInstructionSequenceAcceptsChoiceConsumer(t *testing.T) {
	seq := []Instruction{
		{
			Primitive: Choose{
				Choice:        ResolutionChoice{Kind: ResolutionChoiceMana},
				PublishChoice: ChoiceKey("chosen-color"),
			},
		},
		{
			Primitive: AddMana{
				Amount:     Fixed(1),
				ChoiceFrom: ChoiceKey("chosen-color"),
			},
		},
	}

	if err := ValidateInstructionSequence(seq); err != nil {
		t.Fatalf("ValidateInstructionSequence() error = %v, want nil", err)
	}
}

func TestValidateInstructionSequenceRejectsChoiceKeyUsedAsLinkedKey(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{
		{
			Primitive: Choose{
				Choice:        ResolutionChoice{Kind: ResolutionChoiceMana},
				PublishChoice: ChoiceKey("chosen-color"),
			},
		},
		{
			Primitive: PutOnBattlefield{
				Source: LinkedBattlefieldSource(LinkedKey("chosen-color")),
			},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "linked key") {
		t.Fatalf("error = %v, want linked-key validation failure", err)
	}
}

func TestValidateInstructionSequenceRejectsDuplicateResultKey(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{
		{Primitive: Draw{Amount: Fixed(1), Player: ControllerReference()}, PublishResult: ResultKey("dup")},
		{Primitive: Draw{Amount: Fixed(1), Player: ControllerReference()}, PublishResult: ResultKey("dup")},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate result key") {
		t.Fatalf("error = %v, want duplicate result key", err)
	}
}

func TestValidateInstructionSequenceRejectsForwardResultGate(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{
		{
			Primitive:   Draw{Amount: Fixed(1), Player: ControllerReference()},
			ResultGate:  opt.Val(InstructionResultGate{Key: ResultKey("later")}),
			Description: "forward result gate",
		},
		{Primitive: Draw{Amount: Fixed(1), Player: ControllerReference()}, PublishResult: ResultKey("later")},
	})
	if err == nil || !strings.Contains(err.Error(), "not yet published") {
		t.Fatalf("error = %v, want forward-reference failure", err)
	}
}

func TestValidateInstructionSequenceAcceptsEventPlayerPaymentGate(t *testing.T) {
	t.Parallel()
	err := ValidateInstructionSequence([]Instruction{
		{
			Primitive: Pay{Payment: ResolutionPayment{
				Payer:    opt.Val(EventPlayerReference()),
				ManaCost: opt.Val(cost.Mana{cost.O(1)}),
			}},
			PublishResult: "unless-paid",
		},
		{
			Primitive: Draw{Player: ControllerReference(), Amount: Fixed(1)},
			Optional:  true,
			ResultGate: opt.Val(InstructionResultGate{
				Key:       "unless-paid",
				Succeeded: TriFalse,
			}),
		},
	})
	if err != nil {
		t.Fatalf("ValidateInstructionSequence() error = %v", err)
	}
}

func TestValidateInstructionSequenceRejectsUnknownDynamicResultKey(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{{
		Primitive: Draw{
			Amount: Dynamic(DynamicAmount{
				Kind:      DynamicAmountPreviousEffectResult,
				ResultKey: ResultKey("missing"),
			}),
			Player: ControllerReference(),
		},
	}})
	if err == nil || !strings.Contains(err.Error(), "result key") {
		t.Fatalf("error = %v, want result-key validation failure", err)
	}
}

func TestValidateInstructionSequenceRejectsOutOfRangePrimitiveTarget(t *testing.T) {
	err := ValidateInstructionSequence(
		[]Instruction{{Primitive: Destroy{Object: TargetPermanentReference(1)}}},
		[]TargetSpec{{MinTargets: 1, MaxTargets: 1}},
	)
	if err == nil || !strings.Contains(err.Error(), "target index 1") {
		t.Fatalf("error = %v, want target-index validation failure", err)
	}
}

func TestValidateInstructionSequenceCardReferenceIndexesCardTargetsOnly(t *testing.T) {
	targets := []TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowPlayer},
		{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowCard, TargetZone: zone.Graveyard},
	}
	seq := []Instruction{{Primitive: MoveCard{
		Card:        CardReference{Kind: CardReferenceTarget, TargetIndex: 0},
		FromZone:    zone.Graveyard,
		Destination: zone.Hand,
	}}}

	if err := ValidateInstructionSequence(seq, targets); err != nil {
		t.Fatalf("first card target after player target: ValidateInstructionSequence() = %v, want nil", err)
	}

	seq[0].Primitive = MoveCard{
		Card:        CardReference{Kind: CardReferenceTarget, TargetIndex: 1},
		FromZone:    zone.Graveyard,
		Destination: zone.Hand,
	}
	if err := ValidateInstructionSequence(seq, targets); err == nil {
		t.Fatal("second card target with one card-target spec: ValidateInstructionSequence() = nil, want error")
	}
}

func TestValidateInstructionSequenceRejectsNilPrimitive(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{{}})
	if err == nil || !strings.Contains(err.Error(), "nil Primitive") {
		t.Fatalf("error = %v, want nil primitive failure", err)
	}
}

func TestModifyPTQuantitySupportsDynamicPowerAndToughness(t *testing.T) {
	power := DynamicAmount{Kind: DynamicAmountX}
	toughness := DynamicAmount{Kind: DynamicAmountOpponentCount}
	primitive := ModifyPT{
		PowerDelta:     Dynamic(power),
		ToughnessDelta: Dynamic(toughness),
		Duration:       DurationUntilEndOfTurn,
	}
	if !primitive.PowerDelta.DynamicAmount().Exists || !reflect.DeepEqual(primitive.PowerDelta.DynamicAmount().Val, power) {
		t.Fatalf("power dynamic = %+v, want %+v", primitive.PowerDelta.DynamicAmount(), power)
	}
	if !primitive.ToughnessDelta.DynamicAmount().Exists || !reflect.DeepEqual(primitive.ToughnessDelta.DynamicAmount().Val, toughness) {
		t.Fatalf("toughness dynamic = %+v, want %+v", primitive.ToughnessDelta.DynamicAmount(), toughness)
	}
	if primitive.Duration != DurationUntilEndOfTurn {
		t.Fatalf("duration = %v, want until end of turn", primitive.Duration)
	}
}

func TestCreateTokenSourceStates(t *testing.T) {
	token := &CardDef{CardFace: CardFace{Name: "Beast"}}
	defSource := TokenDef(token)
	if got, ok := defSource.TokenDefRef(); !ok || got != token {
		t.Fatalf("token source = %+v, want Token only", defSource)
	}
	if _, ok := defSource.TokenCopy(); ok {
		t.Fatalf("token source = %+v, unexpectedly contains TokenCopy", defSource)
	}

	spec := TokenCopySpec{Source: TokenCopySourceSourceCard, NoManaCost: true}
	copySource := TokenCopyOf(spec)
	if _, ok := copySource.TokenDefRef(); ok {
		t.Fatalf("token copy source = %+v, unexpectedly contains Token", copySource)
	}
	if got, ok := copySource.TokenCopy(); !ok || !reflect.DeepEqual(got, spec) {
		t.Fatalf("token copy source = %+v, want TokenCopy", copySource)
	}
}

func TestPutOnBattlefieldSourceStates(t *testing.T) {
	cardRef := CardReference{Kind: CardReferenceSource}
	cardSource := CardBattlefieldSource(cardRef)
	if got, ok := cardSource.CardRef(); !ok || got != cardRef {
		t.Fatalf("card source = %+v, want card reference", cardSource)
	}
	if _, ok := cardSource.LinkedKey(); ok {
		t.Fatalf("card source = %+v, unexpectedly contains linked key", cardSource)
	}

	linkedSource := LinkedBattlefieldSource(LinkedKey("revealed-card"))
	if _, ok := linkedSource.CardRef(); ok {
		t.Fatalf("linked source = %+v, unexpectedly contains card reference", linkedSource)
	}
	if got, ok := linkedSource.LinkedKey(); !ok || got != LinkedKey("revealed-card") {
		t.Fatalf("linked source = %+v, want linked key", linkedSource)
	}
}

func TestQuantityValueSemantics(t *testing.T) {
	// Fixed: zero value is 0, Fixed(n) returns n, IsDynamic is false.
	zero := Quantity{}
	if zero.IsDynamic() {
		t.Fatal("zero Quantity.IsDynamic() = true, want false")
	}
	if zero.Value() != 0 {
		t.Fatalf("zero Quantity.Value() = %d, want 0", zero.Value())
	}
	if zero.DynamicAmount().Exists {
		t.Fatal("zero Quantity.DynamicAmount().Exists = true, want false")
	}

	fixed := Fixed(7)
	if fixed.IsDynamic() {
		t.Fatal("fixed Quantity.IsDynamic() = true, want false")
	}
	if fixed.Value() != 7 {
		t.Fatalf("fixed Quantity.Value() = %d, want 7", fixed.Value())
	}
	if fixed.DynamicAmount().Exists {
		t.Fatal("fixed Quantity.DynamicAmount().Exists = true, want false")
	}

	// Dynamic: IsDynamic is true, Value returns 0, DynamicAmount returns the formula.
	d := DynamicAmount{Kind: DynamicAmountX}
	dyn := Dynamic(d)
	if !dyn.IsDynamic() {
		t.Fatal("dynamic Quantity.IsDynamic() = false, want true")
	}
	if dyn.Value() != 0 {
		t.Fatalf("dynamic Quantity.Value() = %d, want 0", dyn.Value())
	}
	da := dyn.DynamicAmount()
	if !da.Exists {
		t.Fatal("dynamic Quantity.DynamicAmount().Exists = false, want true")
	}
	if da.Val.Kind != DynamicAmountX {
		t.Fatalf("dynamic Quantity.DynamicAmount().Val.Kind = %v, want DynamicAmountX", da.Val.Kind)
	}

	// Copy independence: copying a dynamic Quantity and reading via accessor returns independent copy.
	copied := dyn
	dacopy := copied.DynamicAmount()
	if !dacopy.Exists || dacopy.Val.Kind != DynamicAmountX {
		t.Fatalf("copy Quantity.DynamicAmount() = %+v, want DynamicAmountX", dacopy)
	}
}
