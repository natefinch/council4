package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MerfolkSecretkeeper is the card definition for Merfolk Secretkeeper // Venture Deeper.
//
// Type: Creature — Merfolk Wizard // Sorcery — Adventure
// Cost: {U} // {U}
// Face: Venture Deeper — Sorcery — Adventure ({U})
//
// Oracle text:
//
//	Venture Deeper
//	Target player mills four cards. (Then exile this card. You may cast the creature later from exile.)
var MerfolkSecretkeeper = newMerfolkSecretkeeper

func newMerfolkSecretkeeper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Merfolk Secretkeeper",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Wizard},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 4}),
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Venture Deeper",
			ManaCost: opt.Val(cost.Mana{
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
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(4),
							Player:      game.TargetPlayerReference(0),
							Destination: zone.Graveyard,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player mills four cards. (Then exile this card. You may cast the creature later from exile.)
		`,
		}),
	}
}
