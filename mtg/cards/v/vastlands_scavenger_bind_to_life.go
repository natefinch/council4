package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VastlandsScavenger is the card definition for Vastlands Scavenger // Bind to Life.
//
// Type: Creature — Bear Druid // Instant
// Cost: {1}{G}{G} // {4}{G}
// Face: Bind to Life — Instant ({4}{G})
//
// Oracle text:
//
//	Deathtouch
//	This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
var VastlandsScavenger = newVastlandsScavenger()

func newVastlandsScavenger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Vastlands Scavenger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:         []color.Color{color.Green},
			EntersPrepared: true,
			Types:          []types.Card{types.Creature},
			Subtypes:       []types.Sub{types.Bear, types.Druid},
			Power:          opt.Val(game.PT{Value: 4}),
			Toughness:      opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
			},
			OracleText: `
			Deathtouch
			This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
		`,
		},
		Layout: game.LayoutPrepare,
		Alternate: opt.Val(game.CardFace{
			Name: "Bind to Life",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Mill{
							Amount:        game.Fixed(7),
							Player:        game.ControllerReference(),
							PublishLinked: game.LinkedKey("milled-cards"),
						},
					},
					{
						Primitive: game.ChooseFromZone{
							Player:     game.ControllerReference(),
							SourceZone: zone.Graveyard,
							Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
							Quantity:   game.Fixed(1),
							Destination: game.ChooseDestination{
								Zone: zone.Battlefield,
							},
							Riders: game.ChooseRiders{
								FromLinked: game.LinkedKey("milled-cards"),
							},
							Prompt: "Choose a card to return to the battlefield",
						},
					},
				},
			}.Ability()),
			OracleText: `
			Mill seven cards. Then put a creature card from among them onto the battlefield.
		`,
		}),
	}
}
