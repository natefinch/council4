package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DauthiVoidwalker is the card definition for Dauthi Voidwalker.
//
// Type: Creature — Dauthi Rogue
// Cost: {B}{B}
//
// Oracle text:
//
//	Shadow (This creature can block or be blocked by only creatures with shadow.)
//	If a card would be put into an opponent's graveyard from anywhere, instead exile it with a void counter on it.
//	{T}, Sacrifice this creature: Choose an exiled card an opponent owns with a void counter on it. You may play it this turn without paying its mana cost.
var DauthiVoidwalker = newDauthiVoidwalker

func newDauthiVoidwalker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dauthi Voidwalker",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dauthi, types.Rogue},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ShadowStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Sacrifice this creature: Choose an exiled card an opponent owns with a void counter on it. You may play it this turn without paying its mana cost.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this creature",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PlayChosenExiledCard{
									Player:                game.ControllerReference(),
									Zone:                  zone.Exile,
									OwnerScope:            game.PlayerOpponent,
									Counter:               opt.Val(counter.Void),
									Duration:              game.DurationThisTurn,
									WithoutPayingManaCost: true,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.GraveyardRedirectExileWithCounterReplacement("If a card would be put into an opponent's graveyard from anywhere, instead exile it with a void counter on it.", game.TriggerControllerOpponent, game.TriggerControllerAny, false, counter.Void),
			},
			OracleText: `
			Shadow (This creature can block or be blocked by only creatures with shadow.)
			If a card would be put into an opponent's graveyard from anywhere, instead exile it with a void counter on it.
			{T}, Sacrifice this creature: Choose an exiled card an opponent owns with a void counter on it. You may play it this turn without paying its mana cost.
		`,
		},
	}
}
