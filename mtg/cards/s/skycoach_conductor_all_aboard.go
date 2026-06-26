package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SkycoachConductor is the card definition for Skycoach Conductor // All Aboard.
//
// Type: Creature — Bird Pilot // Instant
// Cost: {2}{U} // {U}
// Face: All Aboard — Instant ({U})
//
// Oracle text:
//
//	Flash
//	Flying, vigilance
//	This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
var SkycoachConductor = newSkycoachConductor()

func newSkycoachConductor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Skycoach Conductor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:         []color.Color{color.Blue},
			EntersPrepared: true,
			Types:          []types.Card{types.Creature},
			Subtypes:       []types.Sub{types.Bird, types.Pilot},
			Power:          opt.Val(game.PT{Value: 2}),
			Toughness:      opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			OracleText: `
			Flash
			Flying, vigilance
			This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
		`,
		},
		Layout: game.LayoutPrepare,
		Alternate: opt.Val(game.CardFace{
			Name: "All Aboard",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target non-Pilot creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Pilot"), Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Exile{
							Object:         game.TargetPermanentReference(0),
							ExileLinkedKey: game.LinkedKey("blink-1"),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.LinkedBattlefieldSource(game.LinkedKey("blink-1")),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile target non-Pilot creature you control, then return that card to the battlefield under its owner's control.
		`,
		}),
	}
}
