package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VoiceOfVictory is the card definition for Voice of Victory.
//
// Type: Creature — Human Bard
// Cost: {1}{W}
//
// Oracle text:
//
//	Mobilize 2 (Whenever this creature attacks, create two tapped and attacking 1/1 red Warrior creature tokens. Sacrifice them at the beginning of the next end step.)
//	Your opponents can't cast spells during your turn.
var VoiceOfVictory = newVoiceOfVictory

func newVoiceOfVictory() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Voice of Victory",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Bard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                           game.RuleEffectCantCastSpells,
							AffectedPlayer:                 game.PlayerOpponent,
							RestrictedDuringControllerTurn: true,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.MobilizeTriggeredBody(game.MobilizeAmount{Fixed: 2}),
			},
			OracleText: `
			Mobilize 2 (Whenever this creature attacks, create two tapped and attacking 1/1 red Warrior creature tokens. Sacrifice them at the beginning of the next end step.)
			Your opponents can't cast spells during your turn.
		`,
		},
	}
}
