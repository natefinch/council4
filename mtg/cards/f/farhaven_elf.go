package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FarhavenElf is the card definition for Farhaven Elf.
//
// Type: Creature — Elf Druid
// Cost: {2}{G}
//
// Oracle text:
//
//	When this creature enters, you may search your library for a basic land card, put it onto the battlefield tapped, then shuffle.
var FarhavenElf = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name: "Farhaven Elf",
		ManaCost: opt.Val(cost.Mana{
			cost.O(2),
			cost.G,
		}),
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elf, types.Druid},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		OracleText: `
			When this creature enters, you may search your library for a basic land card, put it onto the battlefield tapped, then shuffle.
		`,
		TriggeredAbilities: []game.TriggeredAbilityBody{
			{
				Text: `
					When this creature enters, you may search your library for a basic land card, put it onto the battlefield tapped, then shuffle.
				`,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:  game.EventPermanentEnteredBattlefield,
						Source: game.TriggerSourceSelf,
					},
				},
				Optional: true,
				Content: game.PlainAbilityContent{
					Sequence: []game.Instruction{
						{
							Primitive: game.Search{
								TargetIndex: game.TargetIndexController,
								Spec: game.SearchSpec{
									SourceZone:   zone.Library,
									Destination:  zone.Battlefield,
									CardType:     opt.Val(types.Land),
									Supertype:    opt.Val(types.Basic),
									EntersTapped: true,
									Shuffle:      true,
								},
							},
						},
					},
				},
			},
		},
	},
}
