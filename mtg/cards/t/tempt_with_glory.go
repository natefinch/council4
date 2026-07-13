package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TemptWithGlory is the card definition for Tempt with Glory.
//
// Type: Sorcery
// Cost: {5}{W}
//
// Oracle text:
//
//	Tempting offer — Put a +1/+1 counter on each creature you control. Each opponent may put a +1/+1 counter on each creature they control. For each opponent who does, put a +1/+1 counter on each creature you control.
var TemptWithGlory = newTemptWithGlory

func newTemptWithGlory() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Tempt with Glory",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(1),
							Group:       game.PlayerControlledGroup(game.GroupOfferMemberReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							CounterKind: counter.PlusOnePlusOne,
						},
						Optional:           true,
						OptionalActorGroup: opt.Val(game.OpponentsReference()),
						TemptingOffer:      true,
					},
				},
			}.Ability()),
			OracleText: `
			Tempting offer — Put a +1/+1 counter on each creature you control. Each opponent may put a +1/+1 counter on each creature they control. For each opponent who does, put a +1/+1 counter on each creature you control.
		`,
		},
	}
}
