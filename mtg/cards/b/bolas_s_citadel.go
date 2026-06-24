package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BolasSCitadel is the card definition for Bolas's Citadel.
//
// Type: Legendary Artifact
// Cost: {3}{B}{B}{B}
//
// Oracle text:
//
//	You may look at the top card of your library any time.
//	You may play lands and cast spells from the top of your library. If you cast a spell this way, pay life equal to its mana value rather than pay its mana cost.
//	{T}, Sacrifice ten nonland permanents: Each opponent loses 10 life.
var BolasSCitadel = newBolasSCitadel()

func newBolasSCitadel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bolas's Citadel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.LookAtTopCardAnyTimeStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                    game.RuleEffectCastSpellsFromZone,
							AffectedPlayer:          game.PlayerYou,
							CastFromZone:            zone.Library,
							TopCardOnly:             true,
							PayLifeEqualToManaValue: true,
						},
						game.RuleEffect{
							Kind:           game.RuleEffectPlayLandsFromZone,
							AffectedPlayer: game.PlayerYou,
							CastFromZone:   zone.Library,
							TopCardOnly:    true,
							PermanentTypes: []types.Card{types.Land},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Sacrifice ten nonland permanents: Each opponent loses 10 life.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:                 cost.AdditionalSacrifice,
							Text:                 "Sacrifice ten nonland permanents",
							Amount:               10,
							ExcludePermanentType: types.Land,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount:      game.Fixed(10),
									PlayerGroup: game.OpponentsReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			You may look at the top card of your library any time.
			You may play lands and cast spells from the top of your library. If you cast a spell this way, pay life equal to its mana value rather than pay its mana cost.
			{T}, Sacrifice ten nonland permanents: Each opponent loses 10 life.
		`,
		},
	}
}
