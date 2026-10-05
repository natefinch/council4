package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestValidateCounteredSpellConditionInformation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		object    opt.V[ObjectReference]
		selection opt.V[Selection]
		types     []types.Card
		valid     bool
	}{
		{"numeric target stack", opt.Val(TargetStackObjectReference(1)), opt.Val(Selection{ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}), nil, true},
		{"flag only", opt.V[ObjectReference]{}, opt.V[Selection]{}, nil, false},
		{"permanent", opt.Val(TargetPermanentReference(0)), opt.Val(Selection{ManaValue: opt.Val(compare.Int{Op: compare.Equal})}), nil, false},
		{"event", opt.Val(EventStackObjectReference()), opt.Val(Selection{ManaValue: opt.Val(compare.Int{Op: compare.Equal})}), nil, false},
		{"power", opt.Val(TargetStackObjectReference(0)), opt.Val(Selection{Power: opt.Val(compare.Int{Op: compare.Equal})}), nil, false},
		{"unavailable color", opt.Val(TargetStackObjectReference(0)), opt.Val(Selection{ManaValue: opt.Val(compare.Int{Op: compare.Equal}), Colorless: true}), nil, false},
		{"unavailable type", opt.Val(TargetStackObjectReference(0)), opt.Val(Selection{ManaValue: opt.Val(compare.Int{Op: compare.Equal})}), []types.Card{types.Instant}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			condition := Condition{Object: tt.object, ObjectMatches: tt.selection, Types: tt.types, UseCounteredSpellManaValue: true}
			if condition.Empty() {
				t.Fatal("explicit information policy was discarded as empty")
			}
			card := &CardDef{CardFace: CardFace{Name: "Information", SpellAbility: opt.Val(Mode{
				Targets: []TargetSpec{{MinTargets: 2, MaxTargets: 2, Allow: TargetAllowStackObject,
					Predicate: TargetPredicate{StackObjectKinds: []StackObjectKind{StackSpell}}}},
				Sequence: []Instruction{{
					Primitive: GainLife{Player: ControllerReference(), Amount: Fixed(1)},
					Condition: opt.Val(EffectCondition{Condition: opt.Val(condition)}),
				}}}.Ability())}}
			issues := ValidateCardDef(card)
			if tt.valid && len(issues) != 0 {
				t.Fatalf("valid fixture produced issues: %+v", issues)
			}
			invalid := hasCardDefIssue(issues, CardDefIssueInvalidCondition)
			if invalid == tt.valid {
				t.Fatalf("invalid=%t, want valid=%t", invalid, tt.valid)
			}
		})
	}
}
