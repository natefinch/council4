package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UnexplainedAbsence is the card definition for Unexplained Absence.
//
// Type: Instant
// Cost: {3}{W}
//
// Oracle text:
//
//	For each player, exile up to one target nonland permanent that player controls. For each permanent exiled this way, its controller cloaks the top card of their library. (To cloak a card, put it onto the battlefield face down as a 2/2 creature with ward {2}. Turn it face up any time for its mana cost if it's a creature card.)
var UnexplainedAbsence = newUnexplainedAbsence

func newUnexplainedAbsence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Unexplained Absence",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ForEachPlayer{
							Scope:     game.AllPlayersReference(),
							Chooser:   game.ControllerReference(),
							Selection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
							Removal:   game.DistributiveRemovalExile,
							LinkedKey: game.LinkedKey("exiled-for-each-player-manifest"),
						},
					},
					{
						Primitive: game.ManifestForEachLinked{
							Cloak:     true,
							LinkedKey: game.LinkedKey("exiled-for-each-player-manifest"),
						},
					},
				},
			}.Ability()),
			OracleText: `
			For each player, exile up to one target nonland permanent that player controls. For each permanent exiled this way, its controller cloaks the top card of their library. (To cloak a card, put it onto the battlefield face down as a 2/2 creature with ward {2}. Turn it face up any time for its mana cost if it's a creature card.)
		`,
		},
	}
}
