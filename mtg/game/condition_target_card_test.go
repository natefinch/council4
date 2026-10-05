package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestValidateTargetCardConditionProducer(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		producer Primitive
		object   ObjectReference
		key      ResultKey
		forward  bool
		valid    bool
	}{
		{"exact second slot", MoveCard{Card: CardReference{Kind: CardReferenceTarget, TargetIndex: 1}, FromZone: zone.Graveyard, Destination: zone.Exile}, TargetCardReference(1), "moved", false, true},
		{"unpublished", MoveCard{Card: CardReference{Kind: CardReferenceTarget}, FromZone: zone.Graveyard, Destination: zone.Exile}, TargetCardReference(0), "missing", false, false},
		{"forward", MoveCard{Card: CardReference{Kind: CardReferenceTarget}, FromZone: zone.Graveyard, Destination: zone.Exile}, TargetCardReference(0), "moved", true, false},
		{"wrong primitive", GainLife{Player: ControllerReference(), Amount: Fixed(1)}, TargetCardReference(0), "moved", false, false},
		{"wrong domain", MoveCard{Card: CardReference{Kind: CardReferenceTarget}, FromZone: zone.Graveyard, Destination: zone.Exile}, TargetPermanentReference(0), "moved", false, false},
		{"wrong slot", MoveCard{Card: CardReference{Kind: CardReferenceTarget}, FromZone: zone.Graveyard, Destination: zone.Exile}, TargetCardReference(1), "moved", false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			producer := Instruction{Primitive: tt.producer, PublishResult: "moved"}
			consumer := Instruction{
				Primitive: GainLife{Player: ControllerReference(), Amount: Fixed(1)},
				Condition: opt.Val(EffectCondition{Condition: opt.Val(Condition{
					Object: opt.Val(tt.object), TargetCardResultKey: tt.key,
					ObjectMatches: opt.Val(Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				})}),
			}
			seq := []Instruction{producer, consumer}
			if tt.forward {
				seq = []Instruction{consumer, producer}
			}
			err := ValidateInstructionSequence(seq, []TargetSpec{{
				MinTargets: 2, MaxTargets: 2, Allow: TargetAllowCard, TargetZone: zone.Graveyard,
			}})
			if (err == nil) != tt.valid {
				t.Fatalf("validation error=%v, want valid=%t", err, tt.valid)
			}
		})
	}
}
