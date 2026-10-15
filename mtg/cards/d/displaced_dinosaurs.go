package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DisplacedDinosaurs is the card definition for Displaced Dinosaurs.
//
// Type: Creature — Dinosaur
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	As a historic permanent you control enters, it becomes a 7/7 Dinosaur creature in addition to its other types. (Artifacts, legendaries, and Sagas are historic.)
var DisplacedDinosaurs = newDisplacedDinosaurs

func newDisplacedDinosaurs() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Displaced Dinosaurs",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersBecomesGroupReplacement("As a historic permanent you control enters, it becomes a 7/7 Dinosaur creature in addition to its other types. (Artifacts, legendaries, and Sagas are historic.)", game.EntersBecomesGroupParams{Controller: game.TriggerControllerYou, Historic: true, AddTypes: []types.Card{types.Creature}, AddSubtypes: []types.Sub{types.Dinosaur}, BasePower: opt.Val(7), BaseToughness: opt.Val(7)}),
			},
			OracleText: `
			As a historic permanent you control enters, it becomes a 7/7 Dinosaur creature in addition to its other types. (Artifacts, legendaries, and Sagas are historic.)
		`,
		},
	}
}
