package game

import (
	"fmt"

	"github.com/natefinch/council4/opt"
)

// ResultKey is a key published by an Instruction after its effect resolves.
// Other instructions reference it in ResultGate or InstructionResultGate.
type ResultKey string

// ChoiceKey is a key published by a Choose primitive.
// It is consumed by AddMana.ChoiceFrom and similar choice-consuming fields.
type ChoiceKey string

// LinkedKey is a key published by a Reveal or exile-like primitive.
// It is consumed by PutOnBattlefield.Source when the source is a linked object,
// and by CardCondition references to linked objects.
type LinkedKey string

// InstructionResultGate gates an Instruction on a previously published ResultKey.
type InstructionResultGate struct {
	Key       ResultKey
	Accepted  TriState
	Succeeded TriState
}

// Instruction wraps one Primitive with sequencing envelope metadata.
// It represents a single step in a resolved ability or spell.
type Instruction struct {
	// Primitive is the data-only effect building block.
	Primitive Primitive

	// Condition is an additional condition evaluated against the resolving stack object.
	Condition opt.V[EffectCondition]

	// CardCondition gates the instruction on properties of a referenced card.
	CardCondition opt.V[CardCondition]

	// ResultGate gates this instruction on the recorded result of a prior instruction
	// identified by ResultGate.Key. Use this for "if you do" / "if you don't" branches.
	ResultGate opt.V[InstructionResultGate]

	// Optional causes the engine to ask the controller whether to apply this instruction.
	// The result is published via PublishResult if set.
	Optional bool

	// PublishResult publishes this instruction's result under the given key so that
	// downstream instructions can reference it via ResultGate.
	PublishResult ResultKey

	// Description is a short human-readable label for logs and diagnostics.
	Description string
}

// Quantity is a resolved numeric value — either a fixed literal or a dynamic
// formula evaluated when the instruction resolves (CR 608.2c).
// Exactly one of fixed or dynamic is meaningful; use Fixed() or Dynamic().
type Quantity struct {
	fixed   int
	dynamic opt.V[DynamicAmount]
}

// Fixed returns a Quantity with a constant value.
func Fixed(n int) Quantity { return Quantity{fixed: n} }

// Dynamic returns a Quantity computed at resolution time.
func Dynamic(d DynamicAmount) Quantity { return Quantity{dynamic: opt.Val(d)} }

// IsDynamic reports whether this Quantity is computed at resolution time.
func (q Quantity) IsDynamic() bool { return q.dynamic.Exists }

// Value returns the fixed value of the Quantity. If dynamic, returns 0.
func (q Quantity) Value() int {
	if q.dynamic.Exists {
		return 0
	}
	return q.fixed
}

// DynamicAmount returns the dynamic formula, if any.
func (q Quantity) DynamicAmount() opt.V[DynamicAmount] { return q.dynamic }

// ValidateInstructionSequence checks that a slice of Instructions has no structural
// errors: no duplicate published keys, no ResultGate referencing an unknown or
// forward ResultKey, and no primitive or card-condition linked references
// pointing at an unpublished key.
//
// It returns the first validation error found, or nil.
func ValidateInstructionSequence(seq []Instruction, targetSpecs ...[]TargetSpec) error {
	var targets []TargetSpec
	checkTargets := len(targetSpecs) != 0
	if checkTargets {
		targets = targetSpecs[0]
	}
	publishedResults := map[ResultKey]int{}
	publishedChoices := map[ChoiceKey]int{}
	publishedLinked := map[LinkedKey]int{}
	for i := range seq {
		instr := &seq[i]
		if instr.Primitive == nil {
			return fmt.Errorf("instruction[%d]: nil Primitive", i)
		}
		if err := instr.Primitive.validatePrimitive(targets, checkTargets); err != nil {
			return fmt.Errorf("instruction[%d]: %w", i, err)
		}
		if instr.ResultGate.Exists {
			key := instr.ResultGate.Val.Key
			if key != "" {
				if _, ok := publishedResults[key]; !ok {
					return fmt.Errorf("instruction[%d]: ResultGate references key %q not yet published", i, key)
				}
			}
		}
		if err := validateLinkedCardCondition(i, instr.CardCondition, publishedLinked); err != nil {
			return err
		}
		refs := instr.Primitive.instructionRefs()
		for _, key := range refs.consumesResults {
			if _, ok := publishedResults[key]; !ok {
				return fmt.Errorf("instruction[%d]: primitive references result key %q not yet published", i, key)
			}
		}
		for _, key := range refs.consumesChoices {
			if _, ok := publishedChoices[key]; !ok {
				return fmt.Errorf("instruction[%d]: primitive references choice key %q not yet published", i, key)
			}
		}
		for _, key := range refs.consumesLinked {
			if _, ok := publishedLinked[key]; !ok {
				return fmt.Errorf("instruction[%d]: primitive references linked key %q not yet published", i, key)
			}
		}
		if instr.PublishResult != "" {
			if prev, dup := publishedResults[instr.PublishResult]; dup {
				return fmt.Errorf("instruction[%d]: duplicate result key %q (first used at index %d)", i, instr.PublishResult, prev)
			}
			publishedResults[instr.PublishResult] = i
		}
		if refs.publishesChoice != "" {
			if prev, dup := publishedChoices[refs.publishesChoice]; dup {
				return fmt.Errorf("instruction[%d]: duplicate choice key %q (first used at index %d)", i, refs.publishesChoice, prev)
			}
			publishedChoices[refs.publishesChoice] = i
		}
		if refs.publishesLinked != "" {
			if prev, dup := publishedLinked[refs.publishesLinked]; dup {
				return fmt.Errorf("instruction[%d]: duplicate linked key %q (first used at index %d)", i, refs.publishesLinked, prev)
			}
			publishedLinked[refs.publishesLinked] = i
		}
	}
	return nil
}

func validateLinkedCardCondition(idx int, cond opt.V[CardCondition], published map[LinkedKey]int) error {
	if !cond.Exists || cond.Val.Card.Kind != CardReferenceLinked {
		return nil
	}
	key := LinkedKey(cond.Val.Card.LinkID)
	if key == "" {
		return nil
	}
	if _, ok := published[key]; ok {
		return nil
	}
	return fmt.Errorf("instruction[%d]: CardCondition references linked key %q not yet published", idx, key)
}
