package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IgnobleHierarch is the card definition for Ignoble Hierarch.
//
// Type: Creature — Goblin Shaman
// Cost: {G}
//
// Oracle text:
//
//	Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
//	{T}: Add {B}, {R}, or {G}.
var IgnobleHierarch = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green, color.Red),
		CardFace: game.CardFace{
			Name: "Ignoble Hierarch",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Shaman},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			OracleText: `
				Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
				{T}: Add {B}, {R}, or {G}.
			`,
		},
	}

	card.StaticAbilities = append(card.StaticAbilities,
		game.ExaltedStaticBody,
	)

	card.ManaAbilities = append(card.ManaAbilities,
		game.ManaAbility{
			Text: `
				{T}: Add {B}, {R}, or {G}.
			`,
			AdditionalCosts: cost.Tap,
			Content: game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Choose{
							Choice: game.ResolutionChoice{
								Kind:   game.ResolutionChoiceMana,
								Prompt: "Choose a color",
								Colors: []mana.Color{
									mana.B,
									mana.R,
									mana.G,
								},
							},
							PublishChoice: game.ChoiceKey("ignoble-hierarch-color"),
						},
					},
					{
						Primitive: game.AddMana{
							Amount:     game.Fixed(1),
							ChoiceFrom: game.ChoiceKey("ignoble-hierarch-color"),
						},
					},
				},
			}.Ability(),
		},
	)
	return card
}()
