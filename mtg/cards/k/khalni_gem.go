package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KhalniGem is the card definition for Khalni Gem.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	When this artifact enters, return two lands you control to their owner's hand.
//	{T}: Add two mana of any one color.
var KhalniGem = newKhalniGem

func newKhalniGem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Khalni Gem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceCountAbility("{T}: Add two mana of any one color.", 2, mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Bounce{
									ControlledChoice: true,
									Amount:           game.Fixed(2),
									Group:            game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this artifact enters, return two lands you control to their owner's hand.
			{T}: Add two mana of any one color.
		`,
		},
	}
}
