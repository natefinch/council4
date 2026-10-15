package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CreepingRenaissance is the card definition for Creeping Renaissance.
//
// Type: Sorcery
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Choose a permanent type. Return all cards of the chosen type from your graveyard to your hand.
//	Flashback {5}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var CreepingRenaissance = newCreepingRenaissance

func newCreepingRenaissance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Creeping Renaissance",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {5}{G}{G}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(5),
								cost.G,
								cost.G,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Choose{
							Choice: game.ResolutionChoice{
								Kind:      game.ResolutionChoiceCardType,
								Prompt:    "Choose a permanent type",
								CardTypes: []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker},
							},
							PublishChoice: game.ChoiceKey("chosen-permanent-type"),
						},
					},
					{
						Primitive: game.MassReturnFromGraveyard{
							Player:      game.ControllerReference(),
							Selection:   game.Selection{ChosenCardTypeFrom: game.ChoiceKey("chosen-permanent-type")},
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Choose a permanent type. Return all cards of the chosen type from your graveyard to your hand.
			Flashback {5}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
