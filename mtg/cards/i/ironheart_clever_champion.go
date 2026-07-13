package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IronheartCleverChampion is the card definition for Ironheart, Clever Champion.
//
// Type: Legendary Artifact Creature — Human Hero
// Cost: {4}{U}
//
// Oracle text:
//
//	Improvise (Your artifacts can help cast this spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)
//	Flying
//	Noncreature spells you cast have improvise.
var IronheartCleverChampion = newIronheartCleverChampion

func newIronheartCleverChampion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Ironheart, Clever Champion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Hero},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ImproviseStaticBody,
				game.FlyingStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectGrantSpellKeyword,
							AffectedController: game.ControllerYou,
							CardSelection:      game.Selection{ExcludedTypes: []types.Card{types.Creature}},
							GrantedKeyword:     game.Improvise,
						},
					},
				},
			},
			OracleText: `
			Improvise (Your artifacts can help cast this spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)
			Flying
			Noncreature spells you cast have improvise.
		`,
		},
	}
}
