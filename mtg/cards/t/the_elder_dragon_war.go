package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheElderDragonWar is the card definition for The Elder Dragon War.
//
// Type: Enchantment — Saga
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger. Sacrifice after III.)
//	I — This Saga deals 2 damage to each creature and each opponent.
//	II — Discard any number of cards, then draw that many cards.
//	III — Create a 4/4 red Dragon creature token with flying.
var TheElderDragonWar = newTheElderDragonWar()

func newTheElderDragonWar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "The Elder Dragon War",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			StaticAbilities: []game.StaticAbility{
				game.ReadAheadStaticBody,
			},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I — This Saga deals 2 damage to each creature and each opponent.",
					Chapters: []int{1},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(2),
									Recipient:    game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(2),
									Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
				game.ChapterAbility{
					Text:     "II — Discard any number of cards, then draw that many cards.",
					Chapters: []int{2},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.DiscardThenDraw{
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.ChapterAbility{
					Text:     "III — Create a 4/4 red Dragon creature token with flying.",
					Chapters: []int{3},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(theElderDragonWarToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger. Sacrifice after III.)
			I — This Saga deals 2 damage to each creature and each opponent.
			II — Discard any number of cards, then draw that many cards.
			III — Create a 4/4 red Dragon creature token with flying.
		`,
		},
	}
}

var theElderDragonWarToken = newTheElderDragonWarToken()

func newTheElderDragonWarToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Dragon",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
