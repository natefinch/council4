package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCounterPlacementPrimitiveValidation(t *testing.T) {
	t.Parallel()
	permanentTargets := []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowPermanent,
	}}
	playerTargets := []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowPlayer,
	}}
	legacyPermanentTargets := []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "target creature",
	}}
	legacyPlayerTargets := []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "target opponent",
	}}
	valid := []struct {
		name      string
		primitive Primitive
		targets   []TargetSpec
	}{
		{
			"permanent",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Charge},
			permanentTargets,
		},
		{
			"player",
			AddPlayerCounter{Amount: Fixed(1), Player: TargetPlayerReference(0), CounterKind: counter.Poison},
			playerTargets,
		},
		{
			"dynamic source",
			AddPlayerCounter{
				Amount: Dynamic(DynamicAmount{
					Kind:   DynamicAmountObjectPower,
					Object: SourcePermanentReference(),
				}),
				Player:      TargetPlayerReference(0),
				CounterKind: counter.Energy,
			},
			playerTargets,
		},
		{
			"legacy permanent",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Charge},
			legacyPermanentTargets,
		},
		{
			"legacy player",
			AddPlayerCounter{Amount: Fixed(1), Player: TargetPlayerReference(0), CounterKind: counter.Poison},
			legacyPlayerTargets,
		},
		{
			"kind choice on permanent",
			AddCounter{
				Amount:      Fixed(1),
				Object:      TargetPermanentReference(0),
				KindChoices: []counter.Kind{counter.PlusOnePlusOne, counter.Loyalty},
			},
			permanentTargets,
		},
		{
			"kind choice on linked object",
			AddCounter{
				Amount:      Fixed(1),
				Object:      LinkedObjectReference("reanimated"),
				KindChoices: []counter.Kind{counter.PlusOnePlusOne, counter.Loyalty},
			},
			permanentTargets,
		},
	}
	for _, test := range valid {
		if err := ValidateInstructionSequence(
			[]Instruction{{Primitive: test.primitive}},
			test.targets,
		); err != nil {
			t.Fatalf("%s: ValidateInstructionSequence() = %v", test.name, err)
		}
	}

	invalid := []struct {
		name      string
		primitive Primitive
		targets   []TargetSpec
	}{
		{
			"player kind on permanent",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Energy},
			permanentTargets,
		},
		{
			"permanent kind on player",
			AddPlayerCounter{Amount: Fixed(1), Player: TargetPlayerReference(0), CounterKind: counter.Charge},
			playerTargets,
		},
		{
			"unknown kind",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Kind(1000)},
			permanentTargets,
		},
		{
			"zero",
			AddCounter{Amount: Fixed(0), Object: TargetPermanentReference(0), CounterKind: counter.Charge},
			permanentTargets,
		},
		{
			"negative",
			AddPlayerCounter{Amount: Fixed(-1), Player: TargetPlayerReference(0), CounterKind: counter.Poison},
			playerTargets,
		},
		{
			"target out of range",
			AddPlayerCounter{Amount: Fixed(1), Player: TargetPlayerReference(1), CounterKind: counter.Poison},
			playerTargets,
		},
		{
			"player reference to permanent target",
			AddPlayerCounter{Amount: Fixed(1), Player: TargetPlayerReference(0), CounterKind: counter.Poison},
			permanentTargets,
		},
		{
			"permanent reference to player target",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Charge},
			playerTargets,
		},
		{
			"permanent reference to legacy player target",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Charge},
			legacyPlayerTargets,
		},
		{
			"player reference to legacy permanent target",
			AddPlayerCounter{Amount: Fixed(1), Player: TargetPlayerReference(0), CounterKind: counter.Poison},
			legacyPermanentTargets,
		},
		{
			"permanent reference to mixed target",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Charge},
			[]TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowPermanent | TargetAllowPlayer,
			}},
		},
		{
			"player reference to mixed target",
			AddPlayerCounter{Amount: Fixed(1), Player: TargetPlayerReference(0), CounterKind: counter.Poison},
			[]TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowPermanent | TargetAllowPlayer,
			}},
		},
		{
			"permanent reference to unknown target domain",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Charge},
			[]TargetSpec{{MinTargets: 1, MaxTargets: 1}},
		},
		{
			"kind choice with a single kind",
			AddCounter{
				Amount:      Fixed(1),
				Object:      TargetPermanentReference(0),
				KindChoices: []counter.Kind{counter.PlusOnePlusOne},
			},
			permanentTargets,
		},
		{
			"kind choice with a player-only kind",
			AddCounter{
				Amount:      Fixed(1),
				Object:      TargetPermanentReference(0),
				KindChoices: []counter.Kind{counter.PlusOnePlusOne, counter.Poison},
			},
			permanentTargets,
		},
		{
			"kind choice with duplicate kinds",
			AddCounter{
				Amount:      Fixed(1),
				Object:      TargetPermanentReference(0),
				KindChoices: []counter.Kind{counter.PlusOnePlusOne, counter.PlusOnePlusOne},
			},
			permanentTargets,
		},
		{
			"kind choice on a group",
			AddCounter{
				Amount:      Fixed(1),
				Group:       BattlefieldGroup(Selection{RequiredTypes: []types.Card{types.Creature}, Controller: ControllerYou}),
				KindChoices: []counter.Kind{counter.PlusOnePlusOne, counter.Loyalty},
			},
			permanentTargets,
		},
	}
	for _, test := range invalid {
		if err := ValidateInstructionSequence(
			[]Instruction{{Primitive: test.primitive}},
			test.targets,
		); err == nil {
			t.Fatalf("%s: ValidateInstructionSequence() = nil", test.name)
		}
	}
}

func TestAddPlayerCounterInstructionReferences(t *testing.T) {
	t.Parallel()
	err := ValidateInstructionSequence([]Instruction{{
		Primitive: AddPlayerCounter{
			Amount: Dynamic(DynamicAmount{
				Kind:      DynamicAmountPreviousEffectResult,
				ResultKey: ResultKey("missing"),
			}),
			Player:      ControllerReference(),
			CounterKind: counter.Energy,
		},
	}}, nil)
	if err == nil {
		t.Fatal("missing previous result reference was accepted")
	}
}

func TestCounterObjectPrimitiveValidation(t *testing.T) {
	t.Parallel()
	counterPrimitive := CounterObject{Object: TargetStackObjectReference(0)}
	stackTarget := []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowStackObject,
		Predicate:  TargetPredicate{StackObjectKinds: []StackObjectKind{StackSpell}},
	}}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: counterPrimitive}}, stackTarget); err != nil {
		t.Fatalf("stack target: ValidateInstructionSequence() = %v", err)
	}

	for _, targets := range [][]TargetSpec{
		{{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowPermanent}},
		{{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowPermanent | TargetAllowStackObject}},
	} {
		if err := ValidateInstructionSequence([]Instruction{{Primitive: counterPrimitive}}, targets); err == nil {
			t.Fatalf("incompatible target %+v: ValidateInstructionSequence() = nil", targets[0])
		}
	}
}

func TestSlimeAndBountyCounterPlacementValidation(t *testing.T) {
	t.Parallel()
	permanentTargets := []TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowPermanent}}
	playerTargets := []TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowPlayer}}
	for _, kind := range []counter.Kind{counter.Slime, counter.Bounty} {
		for _, primitive := range []Primitive{
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: kind},
			RemoveCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: kind},
		} {
			if err := ValidateInstructionSequence([]Instruction{{Primitive: primitive}}, permanentTargets); err != nil {
				t.Fatalf("%s %T: ValidateInstructionSequence() = %v", kind, primitive, err)
			}
		}
		if err := ValidateInstructionSequence([]Instruction{{Primitive: AddPlayerCounter{
			Amount: Fixed(1), Player: TargetPlayerReference(0), CounterKind: kind,
		}}}, playerTargets); err == nil {
			t.Fatalf("%s accepted as a player counter; want permanent-only marker", kind)
		}
		selection := Selection{MatchCounter: true, RequiredCounter: kind}
		if problems := selection.Validate(); len(problems) != 0 {
			t.Fatalf("%s counter selection Validate() = %v", kind, problems)
		}
	}
	for name, selection := range map[string]Selection{
		"required":     {MatchCounter: true, RequiredCounter: counter.Bounty + 1},
		"count":        {RequiredCounter: counter.Bounty + 1, RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 1})},
		"excluded":     {MatchExcludedCounter: true, ExcludedCounter: counter.Bounty + 1},
		"nested AnyOf": {AnyOf: []Selection{{MatchCounter: true, RequiredCounter: counter.Slime}, {MatchCounter: true, RequiredCounter: -1}}},
	} {
		if problems := selection.Validate(); len(problems) == 0 {
			t.Fatalf("%s invalid counter kind accepted by Selection.Validate", name)
		}
	}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: AddCounter{
		Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Bounty + 1,
	}}}, permanentTargets); err == nil {
		t.Fatal("kind after Bounty accepted; want unrecognized counter kind rejected")
	}
}

func TestDamagedSubjectMarkerCounterTriggerValidation(t *testing.T) {
	t.Parallel()
	markerOnDamaged := Selection{
		RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker},
		MatchCounter:     true,
		RequiredCounter:  counter.Bounty,
	}
	if residue := triggerSubjectSelectionResidue(markerOnDamaged, EventDamageDealt); !residue.Empty() {
		t.Fatalf("damaged-permanent bounty filter residue = %+v; want supported", residue)
	}
	if residue := triggerSubjectSelectionResidue(markerOnDamaged, EventSpellCast); residue.Empty() {
		t.Fatal("spell-cast subject counter filter accepted; want unavailable counters to fail closed")
	}
	extra := Selection{MatchCounter: true, RequiredCounter: counter.Slime, MatchCommander: true, MatchGoaded: true}
	residue := triggerSubjectSelectionResidue(extra, EventDamageDealt)
	if residue.MatchCounter || !residue.MatchCommander || !residue.MatchGoaded {
		t.Fatalf("damage subject residue = %+v; want only the counter predicate supported", residue)
	}
}
