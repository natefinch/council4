package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// JidoorAristocraticCapital is the card definition for Jidoor, Aristocratic Capital // Overture.
//
// Type: Land — Town // Sorcery — Adventure
// Cost: {4}{U}{U}
// Face: Overture — Sorcery — Adventure ({4}{U}{U})
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {U}.
var JidoorAristocraticCapital = newJidoorAristocraticCapital

func newJidoorAristocraticCapital() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:     "Jidoor, Aristocratic Capital",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Town},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {U}.
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Overture",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target opponent",
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Mill{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:      game.DynamicAmountCountCardsInZone,
								Divisor:   2,
								Player:    func() *game.PlayerReference { ref := game.TargetPlayerReference(0); return &ref }(),
								CardZone:  zone.Library,
								Selection: &game.Selection{},
							}),
							Player: game.TargetPlayerReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target opponent mills half their library, rounded down. (Then exile this card. You may play the land later from exile.)
		`,
		}),
	}
}
