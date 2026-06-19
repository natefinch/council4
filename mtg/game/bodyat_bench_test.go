package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// benchAbilityFace builds a face with several abilities across the categories
// BodyAt walks, approximating an ability-dense permanent (e.g. a commander).
func benchAbilityFace() *CardFace {
	return &CardFace{
		Name:               "Bench",
		Types:              []types.Card{types.Creature},
		Power:              opt.Val(PT{Value: 1}),
		Toughness:          opt.Val(PT{Value: 1}),
		SpellAbility:       opt.Val(AbilityContent{Modes: []Mode{{}}, MinModes: 1, MaxModes: 1}),
		ActivatedAbilities: []ActivatedAbility{{}},
		ManaAbilities:      []ManaAbility{{}},
		TriggeredAbilities: []TriggeredAbility{{}},
		StaticAbilities:    []StaticAbility{ReachStaticBody},
	}
}

// BenchmarkBodyAt measures the allocation cost of reading every ability of a
// face via BodyAt. With pointer receivers BodyAt returns the element address and
// must allocate nothing; with value receivers it boxed a copy per call.
func BenchmarkBodyAt(b *testing.B) {
	face := benchAbilityFace()
	count := face.AbilityCount()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		for i := range count {
			_ = face.BodyAt(i)
		}
	}
}
