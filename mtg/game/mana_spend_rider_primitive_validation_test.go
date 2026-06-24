package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// validScryRider is the Path of Ancestry spend rider used as a valid baseline in
// these validation tests: scry 1 when the tagged mana is spent to cast a
// creature spell sharing a creature type with the commander.
func validScryRider() ManaSpendRider {
	return ManaSpendRider{
		Condition: ManaSpendCastCommanderCreatureType,
		Effect: Mode{Sequence: []Instruction{
			{Primitive: Scry{Amount: Fixed(1), Player: ControllerReference()}},
		}},
	}
}

func addManaWithRider(rider ManaSpendRider) AddMana {
	return AddMana{Amount: Fixed(1), ManaColor: mana.G, SpendRider: opt.Val(rider)}
}

// TestAddManaSpendRiderValidationAcceptsModeledRider confirms a fully modeled
// rider (recognized condition, non-empty untargeted effect) validates.
func TestAddManaSpendRiderValidationAcceptsModeledRider(t *testing.T) {
	t.Parallel()
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(validScryRider())}}); err != nil {
		t.Fatalf("ValidateInstructionSequence() = %v, want nil", err)
	}
}

// TestAddManaSpendRiderValidationRejectsUnknownCondition confirms the unknown
// condition value is rejected rather than treated as a no-op rider.
func TestAddManaSpendRiderValidationRejectsUnknownCondition(t *testing.T) {
	t.Parallel()
	rider := validScryRider()
	rider.Condition = ManaSpendConditionUnknown
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(rider)}}); err == nil {
		t.Fatal("ValidateInstructionSequence() = nil, want error for unknown condition")
	}
}

// TestAddManaSpendRiderValidationRejectsOutOfRangeCondition confirms the
// exhaustive enum switch rejects any value outside the modeled conditions, not
// just the zero unknown value.
func TestAddManaSpendRiderValidationRejectsOutOfRangeCondition(t *testing.T) {
	t.Parallel()
	rider := validScryRider()
	rider.Condition = ManaSpendCastLegendarySpell + 1
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(rider)}}); err == nil {
		t.Fatal("ValidateInstructionSequence() = nil, want error for out-of-range condition")
	}
}

func TestAddManaSpendRiderValidationChosenTypeBoundaries(t *testing.T) {
	t.Parallel()
	valid := ManaSpendRider{
		Condition:         ManaSpendCastChosenCreatureType,
		Restriction:       ManaSpendRestrictedToCondition,
		SpellRuleEffect:   RuleEffectCantBeCountered,
		ChosenSubtypeFrom: EntryTypeChoiceKey,
	}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(valid)}}); err != nil {
		t.Fatalf("valid chosen-type rider: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*ManaSpendRider)
	}{
		{"unknown restriction", func(r *ManaSpendRider) {
			r.Restriction = ManaSpendRestrictedToCondition + 1
		}},
		{"missing chosen subtype source", func(r *ManaSpendRider) {
			r.ChosenSubtypeFrom = ""
		}},
		{"unsupported chosen subtype source", func(r *ManaSpendRider) {
			r.ChosenSubtypeFrom = ChoiceKey("other-choice")
		}},
		{"unknown spell rule effect", func(r *ManaSpendRider) {
			r.SpellRuleEffect = RuleEffectCantBeBlocked
		}},
		{"triggered and spell effects together", func(r *ManaSpendRider) {
			r.Effect = validScryRider().Effect
		}},
		{"declared targets", func(r *ManaSpendRider) {
			r.Effect.Targets = []TargetSpec{{MinTargets: 1, MaxTargets: 1}}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			rider := valid
			test.mutate(&rider)
			if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(rider)}}); err == nil {
				t.Fatal("ValidateInstructionSequence() = nil, want error")
			}
		})
	}
}

func TestAddManaSpendRiderValidationLegendaryBoundaries(t *testing.T) {
	t.Parallel()
	valid := ManaSpendRider{
		Condition:       ManaSpendCastLegendarySpell,
		Restriction:     ManaSpendRestrictedToCondition,
		SpellRuleEffect: RuleEffectCantBeCountered,
	}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(valid)}}); err != nil {
		t.Fatalf("valid legendary rider: %v", err)
	}
	bare := valid
	bare.SpellRuleEffect = RuleEffectNone
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(bare)}}); err != nil {
		t.Fatalf("valid bare legendary rider: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*ManaSpendRider)
	}{
		{"unrestricted", func(r *ManaSpendRider) {
			r.Restriction = ManaSpendUnrestricted
		}},
		{"unexpected chosen subtype source", func(r *ManaSpendRider) {
			r.ChosenSubtypeFrom = EntryTypeChoiceKey
		}},
		{"unknown spell rule effect", func(r *ManaSpendRider) {
			r.SpellRuleEffect = RuleEffectCantBeBlocked
		}},
		{"unexpected effect sequence", func(r *ManaSpendRider) {
			r.Effect = validScryRider().Effect
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			rider := valid
			test.mutate(&rider)
			if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(rider)}}); err == nil {
				t.Fatal("ValidateInstructionSequence() = nil, want error")
			}
		})
	}
}

// TestAddManaSpendRiderValidationCreatureSpellHasteBoundaries confirms the
// unrestricted creature-spell haste rider requires a non-empty SpellGainsKeywords
// slice, an unrestricted restriction, and no other rider fields.
func TestAddManaSpendRiderValidationCreatureSpellHasteBoundaries(t *testing.T) {
	t.Parallel()
	valid := ManaSpendRider{
		Condition:          ManaSpendCastCreatureSpell,
		SpellGainsKeywords: []Keyword{Haste},
	}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(valid)}}); err != nil {
		t.Fatalf("valid creature-spell haste rider: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*ManaSpendRider)
	}{
		{"missing granted keywords", func(r *ManaSpendRider) {
			r.SpellGainsKeywords = nil
		}},
		{"restricted", func(r *ManaSpendRider) {
			r.Restriction = ManaSpendRestrictedToCondition
		}},
		{"unexpected chosen subtype source", func(r *ManaSpendRider) {
			r.ChosenSubtypeFrom = EntryTypeChoiceKey
		}},
		{"unexpected spell rule effect", func(r *ManaSpendRider) {
			r.SpellRuleEffect = RuleEffectCantBeCountered
		}},
		{"unexpected effect sequence", func(r *ManaSpendRider) {
			r.Effect = validScryRider().Effect
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			rider := valid
			test.mutate(&rider)
			if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(rider)}}); err == nil {
				t.Fatal("ValidateInstructionSequence() = nil, want error")
			}
		})
	}
}

// TestAddManaSpendRiderValidationCreatureSpellRestrictedBoundaries covers the
// bare restricted creature-spell spend rider (Beastcaller Savant): the tagged
// mana may be spent only on creature spells, with no granted keywords, chosen
// subtype, rule effect, or effect sequence.
func TestAddManaSpendRiderValidationCreatureSpellRestrictedBoundaries(t *testing.T) {
	t.Parallel()
	valid := ManaSpendRider{
		Condition:   ManaSpendCastCreatureSpell,
		Restriction: ManaSpendRestrictedToCondition,
	}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(valid)}}); err != nil {
		t.Fatalf("valid restricted creature-spell rider: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*ManaSpendRider)
	}{
		{"unexpected granted keywords", func(r *ManaSpendRider) {
			r.SpellGainsKeywords = []Keyword{Haste}
		}},
		{"unexpected chosen subtype source", func(r *ManaSpendRider) {
			r.ChosenSubtypeFrom = EntryTypeChoiceKey
		}},
		{"unexpected spell rule effect", func(r *ManaSpendRider) {
			r.SpellRuleEffect = RuleEffectCantBeCountered
		}},
		{"unexpected effect sequence", func(r *ManaSpendRider) {
			r.Effect = validScryRider().Effect
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			rider := valid
			test.mutate(&rider)
			if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(rider)}}); err == nil {
				t.Fatal("ValidateInstructionSequence() = nil, want error")
			}
		})
	}
}

func TestCardDefChosenTypeManaRiderRequiresEntryChoice(t *testing.T) {
	t.Parallel()
	rider := ManaSpendRider{
		Condition:         ManaSpendCastChosenCreatureType,
		Restriction:       ManaSpendRestrictedToCondition,
		SpellRuleEffect:   RuleEffectCantBeCountered,
		ChosenSubtypeFrom: EntryTypeChoiceKey,
	}
	card := &CardDef{CardFace: CardFace{
		Name:          "Choice-Free Cavern",
		Types:         []types.Card{types.Land},
		ManaAbilities: []ManaAbility{TapManaChoiceWithSpendRiderAbility("", rider, mana.W, mana.U, mana.B, mana.R, mana.G)},
	}}
	if issues := ValidateCardDef(card); len(issues) == 0 {
		t.Fatal("ValidateCardDef() accepted a chosen-type mana rider without an entry-time type choice")
	}

	card.ReplacementAbilities = []ReplacementAbility{EntryTypeChoiceReplacement("As this land enters, choose a creature type.")}
	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("ValidateCardDef() with entry choice = %#v, want no issues", issues)
	}
}

// TestAddManaSpendRiderValidationRejectsEmptyEffect confirms a rider with no
// effect instructions is rejected.
func TestAddManaSpendRiderValidationRejectsEmptyEffect(t *testing.T) {
	t.Parallel()
	rider := validScryRider()
	rider.Effect = Mode{}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(rider)}}); err == nil {
		t.Fatal("ValidateInstructionSequence() = nil, want error for empty rider effect")
	}
}

// TestAddManaSpendRiderValidationRejectsDeclaredTargets confirms a rider that
// declares target specs is rejected: a fired rider is put on the stack with no
// targets of its own, so it could never choose a legal target.
func TestAddManaSpendRiderValidationRejectsDeclaredTargets(t *testing.T) {
	t.Parallel()
	rider := validScryRider()
	rider.Effect.Targets = []TargetSpec{{MinTargets: 1, MaxTargets: 1}}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(rider)}}); err == nil {
		t.Fatal("ValidateInstructionSequence() = nil, want error for declared rider targets")
	}
}

// TestAddManaSpendRiderValidationRejectsTargetedInstruction confirms a rider
// whose effect references a target is rejected even when it declares no target
// specs, because the sequence is validated against an empty target set.
func TestAddManaSpendRiderValidationRejectsTargetedInstruction(t *testing.T) {
	t.Parallel()
	rider := validScryRider()
	rider.Effect = Mode{Sequence: []Instruction{
		{Primitive: Destroy{Object: TargetPermanentReference(0)}},
	}}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(rider)}}); err == nil {
		t.Fatal("ValidateInstructionSequence() = nil, want error for targeted rider instruction")
	}
}
