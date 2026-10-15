package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OlórinSSearingLight is the card definition for Olórin's Searing Light.
//
// Type: Instant
// Cost: {2}{R}{W}
//
// Oracle text:
//
//	Each opponent exiles a creature with the greatest power among creatures that player controls.
//	Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, Olórin's Searing Light deals damage to each opponent equal to the power of the creature they exiled.
var OlórinSSearingLight = newOlórinSSearingLight

func newOlórinSSearingLight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Olórin's Searing Light",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.W,
			}),
			Colors: []color.Color{color.Red, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ForEachPlayer{
							Scope:        game.OpponentsReference(),
							Chooser:      game.GroupOfferMemberReference(),
							Selection:    game.Selection{RequiredTypes: []types.Card{types.Creature}},
							Removal:      game.DistributiveRemovalExile,
							LinkedKey:    game.LinkedKey("correlated-opponent-exile"),
							ReplaceLink:  true,
							Required:     true,
							Extremum:     game.PermanentChoiceGreatestPower,
							Simultaneous: true,
						},
					},
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:   game.DynamicAmountObjectPower,
								Object: game.LinkedObjectReference("correlated-opponent-exile"),
								Player: func() *game.PlayerReference { ref := game.GroupOfferMemberReference(); return &ref }(),
							}),
							Recipient: game.PlayerDamageRecipient(game.GroupOfferMemberReference()),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControllerGraveyardInstantOrSorceryCountAtLeast: 2,
							}),
						}),
						ForEachPlayerGroup: opt.Val(game.OpponentsReference()),
					},
				},
			}.Ability()),
			OracleText: `
			Each opponent exiles a creature with the greatest power among creatures that player controls.
			Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, Olórin's Searing Light deals damage to each opponent equal to the power of the creature they exiled.
		`,
		},
	}
}
