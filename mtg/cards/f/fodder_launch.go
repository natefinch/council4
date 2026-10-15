package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FodderLaunch is the card definition for Fodder Launch.
//
// Type: Kindred Sorcery — Goblin
// Cost: {3}{B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a Goblin.
//	Target creature gets -5/-5 until end of turn. Fodder Launch deals 5 damage to that creature's controller.
var FodderLaunch = newFodderLaunch

func newFodderLaunch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Fodder Launch",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Kindred, types.Sorcery},
			Subtypes: []types.Sub{types.Goblin},
			AdditionalCosts: []cost.Additional{
				{
					Kind:        cost.AdditionalSacrifice,
					Text:        "sacrifice a Goblin",
					Amount:      1,
					SubtypesAny: cost.SubtypeSet{types.Goblin},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(-5),
							ToughnessDelta: game.Fixed(-5),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(5),
							Recipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0))),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice a Goblin.
			Target creature gets -5/-5 until end of turn. Fodder Launch deals 5 damage to that creature's controller.
		`,
		},
	}
}
