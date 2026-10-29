package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func characteristicTestTargets() []TargetSpec {
	return []TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "target player", Allow: TargetAllowPlayer},
		{MinTargets: 1, MaxTargets: 1, Constraint: "target creature", Allow: TargetAllowPermanent,
			Selection: opt.Val(Selection{RequiredTypes: []types.Card{types.Creature}})},
		{MinTargets: 1, MaxTargets: 1, Constraint: "any target", Allow: TargetAllowPermanent | TargetAllowPlayer},
	}
}

func objectPowerOf(ref ObjectReference) DynamicAmount {
	return DynamicAmount{Kind: DynamicAmountObjectPower, Multiplier: 1, Object: ref}
}

func TestCharacteristicQuantityRejectsPlayerOnlyTargetSlot(t *testing.T) {
	targets := characteristicTestTargets()
	refs := map[string]ObjectReference{
		"target permanent":          TargetPermanentReference(0),
		"target object":             TargetObjectReference(0),
		"target stack object":       TargetStackObjectReference(0),
		"target attached permanent": TargetAttachedPermanentReference(0),
		"target card":               TargetCardReference(0),
	}
	for name, ref := range refs {
		for _, kind := range []DynamicAmountKind{
			DynamicAmountObjectPower, DynamicAmountObjectToughness, DynamicAmountObjectManaValue,
			DynamicAmountTargetPower, DynamicAmountTargetToughness, DynamicAmountTargetManaValue,
		} {
			err := validateQuantity(Dynamic(DynamicAmount{Kind: kind, Multiplier: 1, Object: ref}), targets, true)
			if err == nil || !strings.Contains(err.Error(), "cannot hold the referenced object") {
				t.Fatalf("%s kind %v on player-only slot: error = %v, want object-slot rejection", name, kind, err)
			}
		}
		counters := DynamicAmount{Kind: DynamicAmountObjectCounters, Multiplier: 1, CounterKind: counter.PlusOnePlusOne, Object: ref}
		if err := validateQuantity(Dynamic(counters), targets, true); err == nil {
			t.Fatalf("%s counter count on player-only slot validated", name)
		}
	}
	for _, slot := range []int{1, 2} {
		for _, ref := range []ObjectReference{TargetPermanentReference(slot), TargetObjectReference(slot)} {
			if err := validateQuantity(Dynamic(objectPowerOf(ref)), targets, true); err != nil {
				t.Fatalf("power of permanent-capable slot %d via %v rejected: %v", slot, ref.Kind(), err)
			}
		}
	}
}

func TestCharacteristicQuantityAcceptsMatchingObjectSlots(t *testing.T) {
	targets := []TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "target spell", Allow: TargetAllowStackObject},
		{MinTargets: 1, MaxTargets: 1, Constraint: "target spell or nonland permanent", Allow: TargetAllowStackObject | TargetAllowPermanent},
		{MinTargets: 1, MaxTargets: 1, Constraint: "target creature card in a graveyard", Allow: TargetAllowCard},
	}
	valid := []ObjectReference{
		TargetStackObjectReference(0), TargetObjectReference(0),
		TargetObjectReference(1), TargetPermanentReference(1), TargetStackObjectReference(1),
		TargetCardReference(2),
	}
	for _, ref := range valid {
		amount := DynamicAmount{Kind: DynamicAmountObjectManaValue, Multiplier: 1, Object: ref}
		if err := validateQuantity(Dynamic(amount), targets, true); err != nil {
			t.Fatalf("mana value via %v of slot %d rejected: %v", ref.Kind(), ref.TargetIndex(), err)
		}
	}
	invalid := []ObjectReference{TargetPermanentReference(0), TargetCardReference(0), TargetStackObjectReference(2), TargetObjectReference(2)}
	for _, ref := range invalid {
		amount := DynamicAmount{Kind: DynamicAmountObjectManaValue, Multiplier: 1, Object: ref}
		if err := validateQuantity(Dynamic(amount), targets, true); err == nil {
			t.Fatalf("mana value via %v of mismatched slot %d validated", ref.Kind(), ref.TargetIndex())
		}
	}
}

func TestCharacteristicQuantityRejectsPlayerSlotInsideCombinator(t *testing.T) {
	targets := characteristicTestTargets()
	sum := DynamicAmount{Kind: DynamicAmountSumOf, Operands: []DynamicAmount{
		objectPowerOf(TargetPermanentReference(1)),
		{Kind: DynamicAmountObjectToughness, Multiplier: 1, Object: TargetPermanentReference(0)},
	}}
	if err := validateQuantity(Dynamic(sum), targets, true); err == nil {
		t.Fatal("sum-of operand reading a player-only slot validated")
	}
}

func TestSumAndDifferenceQuantityOperandArity(t *testing.T) {
	targets := characteristicTestTargets()
	power := objectPowerOf(TargetPermanentReference(1))
	toughness := DynamicAmount{Kind: DynamicAmountObjectToughness, Multiplier: 1, Object: TargetPermanentReference(1)}
	tests := []struct {
		name    string
		amount  DynamicAmount
		wantErr string
	}{
		{"sum one operand", DynamicAmount{Kind: DynamicAmountSumOf, Operands: []DynamicAmount{power}}, "at least two operands"},
		{"sum two operands", DynamicAmount{Kind: DynamicAmountSumOf, Operands: []DynamicAmount{power, toughness}}, ""},
		{"sum three operands", DynamicAmount{Kind: DynamicAmountSumOf, Operands: []DynamicAmount{power, toughness, power}}, ""},
		{"difference one operand", DynamicAmount{Kind: DynamicAmountDifferenceOf, Operands: []DynamicAmount{power}}, "exactly two operands"},
		{"difference two operands", DynamicAmount{Kind: DynamicAmountDifferenceOf, Operands: []DynamicAmount{power, toughness}}, ""},
		{"difference three operands", DynamicAmount{Kind: DynamicAmountDifferenceOf, Operands: []DynamicAmount{power, toughness, power}}, "exactly two operands"},
		{"invalid nested operand", DynamicAmount{Kind: DynamicAmountSumOf, Operands: []DynamicAmount{power, {Kind: DynamicAmountObjectPower}}}, "object"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateQuantity(Dynamic(test.amount), targets, true)
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("validateQuantity() error = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("validateQuantity() error = %v, want containing %q", err, test.wantErr)
			}
		})
	}
}

func TestCardDefValidationRejectsPlayerTargetPowerLifeAmount(t *testing.T) {
	spell := func(allow TargetAllow, kind DynamicAmountKind, ref ObjectReference) *CardDef {
		return &CardDef{CardFace: CardFace{
			Name:  "Characteristic Life Probe",
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(Mode{
				Targets: []TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "target", Allow: allow}},
				Sequence: []Instruction{{Primitive: LoseLife{
					Player: ControllerReference(),
					Amount: Dynamic(DynamicAmount{Kind: kind, Multiplier: 1, Object: ref}),
				}}},
			}.Ability()),
		}}
	}
	for _, kind := range []DynamicAmountKind{DynamicAmountObjectPower, DynamicAmountObjectToughness, DynamicAmountObjectManaValue} {
		for _, ref := range []ObjectReference{TargetPermanentReference(0), TargetObjectReference(0)} {
			if issues := ValidateCardDef(spell(TargetAllowPlayer, kind, ref)); len(issues) == 0 {
				t.Fatalf("card reading %v via %v from a player-only target validated", kind, ref.Kind())
			}
			if issues := ValidateCardDef(spell(TargetAllowPermanent, kind, ref)); len(issues) != 0 {
				t.Fatalf("card reading %v via %v from a permanent target rejected: %v", kind, ref.Kind(), issues)
			}
		}
	}
}

func TestTargetSpecMatchableKindsInfersLegacyConstraints(t *testing.T) {
	tests := []struct {
		constraint string
		allow      TargetAllow
		want       TargetAllow
	}{
		{"target player", TargetAllowUnspecified, TargetAllowPlayer},
		{"target opponent", TargetAllowUnspecified, TargetAllowPlayer},
		{"any target", TargetAllowUnspecified, TargetAllowPermanent | TargetAllowPlayer},
		{"target creature", TargetAllowUnspecified, TargetAllowPermanent},
		{"target nonland permanent", TargetAllowUnspecified, TargetAllowPermanent},
		{"target creature card in a graveyard", TargetAllowUnspecified, TargetAllowPermanent | TargetAllowCard},
		{"target card in your hand", TargetAllowUnspecified, TargetAllowCard},
		{"target instant card from your library", TargetAllowUnspecified, TargetAllowCard},
		{"target spell", TargetAllowUnspecified, TargetAllowUnspecified},
		{"target spell", TargetAllowStackObject, TargetAllowStackObject},
		{"target creature card in a graveyard", TargetAllowCard, TargetAllowCard},
	}
	for _, test := range tests {
		spec := TargetSpec{Constraint: test.constraint, Allow: test.allow}
		if got := spec.MatchableKinds(); got != test.want {
			t.Fatalf("MatchableKinds(%q, allow %v) = %v, want %v", test.constraint, test.allow, got, test.want)
		}
	}
}

func TestCharacteristicQuantityInfersUnspecifiedTargetKinds(t *testing.T) {
	targets := []TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "target player"},
		{MinTargets: 1, MaxTargets: 1, Constraint: "target creature"},
		{MinTargets: 1, MaxTargets: 1, Constraint: "any target"},
		{MinTargets: 1, MaxTargets: 1, Constraint: "target creature card in a graveyard"},
		{MinTargets: 1, MaxTargets: 1, Constraint: "target card in your hand"},
	}
	valid := []ObjectReference{
		TargetPermanentReference(1), TargetObjectReference(1), TargetAttachedPermanentReference(1),
		TargetPermanentReference(2), TargetObjectReference(2),
		TargetCardReference(3), TargetPermanentReference(3),
		TargetCardReference(4),
	}
	for _, ref := range valid {
		for _, kind := range []DynamicAmountKind{DynamicAmountObjectPower, DynamicAmountObjectToughness, DynamicAmountObjectManaValue} {
			amount := DynamicAmount{Kind: kind, Multiplier: 1, Object: ref}
			if err := validateQuantity(Dynamic(amount), targets, true); err != nil {
				t.Fatalf("%v via %v of inferred slot %d rejected: %v", kind, ref.Kind(), ref.TargetIndex(), err)
			}
		}
	}
	invalid := []ObjectReference{
		TargetPermanentReference(0), TargetObjectReference(0), TargetCardReference(0), TargetStackObjectReference(0),
		TargetCardReference(1), TargetStackObjectReference(1), TargetCardReference(2),
		TargetStackObjectReference(3), TargetPermanentReference(4), TargetObjectReference(4),
	}
	for _, ref := range invalid {
		amount := DynamicAmount{Kind: DynamicAmountObjectManaValue, Multiplier: 1, Object: ref}
		if err := validateQuantity(Dynamic(amount), targets, true); err == nil {
			t.Fatalf("mana value via %v of inferred mismatched slot %d validated", ref.Kind(), ref.TargetIndex())
		}
	}
}
