package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestPaidCostSelectionDomain(t *testing.T) {
	for _, test := range []struct {
		name      string
		selection Selection
		supported bool
	}{
		{"types and colors", Selection{RequiredTypesAny: []types.Card{types.Creature, types.Artifact}, ColorsAny: []color.Color{color.Red}}, true},
		{"subtype and legendary", Selection{SubtypesAny: []types.Sub{types.Human}, Supertypes: []types.Super{types.Legendary}}, true},
		{"known-zero threshold", Selection{Toughness: opt.Val(compare.Int{Op: compare.LessOrEqual})}, true},
		{"disjunction", Selection{AnyOf: []Selection{{Colorless: true}, {Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})}}}, true},
		{"counter state", Selection{MatchNoCounters: true}, false},
		{"tapped state", Selection{Tapped: TriFalse}, false},
		{"token state", Selection{NonToken: true}, false},
		{"combat state", Selection{CombatState: CombatStateAttacking}, false},
		{"later owner", Selection{Owner: OwnerYou}, false},
		{"attachment state", Selection{MatchEquipped: true}, false},
		{"unsupported nested alternative", Selection{AnyOf: []Selection{{Colorless: true}, {MatchModified: true}}}, false},
		{"malformed color filter", Selection{Colorless: true, Multicolored: true}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := PaidCostSelectionSupported(test.selection); got != test.supported {
				t.Fatalf("supported=%v, want %v", got, test.supported)
			}
			card := &CardDef{CardFace: CardFace{Name: "Cost Traits", SpellAbility: opt.Val(Mode{
				Sequence: []Instruction{{
					Primitive: GainLife{Player: ControllerReference(), Amount: Fixed(1)},
					Condition: opt.Val(EffectCondition{Condition: opt.Val(Condition{
						Object:        opt.Val(PaidCostReference("cost", PaidCostSacrifice, "")),
						ObjectMatches: opt.Val(test.selection),
					})}),
				}},
			}.Ability())}}
			issues := ValidateCardDef(card)
			if invalid := hasCardDefIssue(issues, CardDefIssueInvalidSelection); invalid == test.supported {
				t.Fatalf("CardDef domain validation disagrees with snapshot support: %v", issues)
			}
		})
	}
}
