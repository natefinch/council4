package game

import (
	"reflect"
	"strings"
	"testing"

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
