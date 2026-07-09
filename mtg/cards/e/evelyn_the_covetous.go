package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EvelynTheCovetous is the card definition for Evelyn, the Covetous.
//
// Type: Legendary Creature — Vampire Rogue
// Cost: {2}{U/B}{B}{B/R}
//
// Oracle text:
//
//	Flash
//	Whenever Evelyn or another Vampire you control enters, exile the top card of each player's library with a collection counter on it.
//	Once each turn, you may play a card from exile with a collection counter on it if it was exiled by an ability you controlled, and you may spend mana as though it were mana of any color to cast it.
var EvelynTheCovetous = newEvelynTheCovetous

func newEvelynTheCovetous() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Evelyn, the Covetous",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.U, mana.B),
				cost.B,
				cost.HybridMana(mana.B, mana.R),
			}),
			Colors:     []color.Color{color.Black, color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Vampire, types.Rogue},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                           game.RuleEffectPlayLandsFromZone,
							AffectedPlayer:                 game.PlayerYou,
							CastFromZone:                   zone.Exile,
							PermanentTypes:                 []types.Card{types.Land},
							ExileCounterFilter:             opt.Val(counter.Collection),
							ExileCounterExiledByController: true,
							OncePerTurn:                    true,
						},
						game.RuleEffect{
							Kind:                           game.RuleEffectCastSpellsFromZone,
							AffectedPlayer:                 game.PlayerYou,
							CastFromZone:                   zone.Exile,
							ExileCounterFilter:             opt.Val(counter.Collection),
							ExileCounterExiledByController: true,
							OncePerTurn:                    true,
							SpendAnyMana:                   true,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventPermanentEnteredBattlefield,
							Controller:             game.TriggerControllerYou,
							SubjectSelectionOrSelf: true,
							SubjectSelection:       game.Selection{SubtypesAny: []types.Sub{types.Sub("Vampire")}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ExileTopOfLibrary{
									Amount:      game.Fixed(1),
									PlayerGroup: game.AllPlayersReference(),
									Counter:     opt.Val(counter.Collection),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flash
			Whenever Evelyn or another Vampire you control enters, exile the top card of each player's library with a collection counter on it.
			Once each turn, you may play a card from exile with a collection counter on it if it was exiled by an ability you controlled, and you may spend mana as though it were mana of any color to cast it.
		`,
		},
	}
}
