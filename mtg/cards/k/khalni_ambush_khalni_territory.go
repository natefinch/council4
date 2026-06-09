package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KhalniAmbush is the card definition for Khalni Ambush // Khalni Territory.
//
// Type: Instant // Land
// Face: Khalni Territory — Land
//
// Oracle text:
//
//	Target creature you control fights target creature you don't control. (Each deals damage equal to its power to the other.)
var KhalniAmbush = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Khalni Ambush // Khalni Territory",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			OracleText: `
				Target creature you control fights target creature you don't control. (Each deals damage equal to its power to the other.)
			`,
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "creature you control",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{
								types.Creature,
							},
							Controller: game.ControllerYou,
						},
					},
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "creature you don't control",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{
								types.Creature,
							},
							Controller: game.ControllerNotYou,
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Fight{
							Object:        game.TargetPermanentReference(0),
							RelatedObject: game.TargetPermanentReference(1),
						},
					},
				},
			}.Ability()),
		},
		Layout: game.LayoutModalDFC,
	}

	back := game.CardFace{
		Name:  "Khalni Territory",
		Types: []types.Card{types.Land},
		OracleText: `
			This land enters tapped.
			{T}: Add {G}.
		`,
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedReplacement("This land enters tapped."),
		},
	}

	back.ManaAbilities = append(back.ManaAbilities,
		game.ManaAbility{
			Text: `
				{T}: Add {G}.
			`,
			AdditionalCosts: cost.Tap,
			Content: game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.G,
						},
					},
				},
			}.Ability(),
		},
	)

	card.Back = opt.Val(back)
	return card
}()
