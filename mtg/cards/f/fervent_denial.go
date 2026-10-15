package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FerventDenial is the card definition for Fervent Denial.
//
// Type: Instant
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Counter target spell.
//	Flashback {5}{U}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var FerventDenial = newFerventDenial

func newFerventDenial() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Fervent Denial",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {5}{U}{U}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(5),
								cost.U,
								cost.U,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Counter target spell.
			Flashback {5}{U}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
