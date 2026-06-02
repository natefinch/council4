package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
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
	Name: "Farhaven Elf",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(2),
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     3,
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Types:         []types.Card{types.Creature},
	Subtypes:      []types.Sub{types.Sub("Elf"), types.Druid},
	Power:         opt.Val(game.PT{Value: 1}),
	Toughness:     opt.Val(game.PT{Value: 1}),
	OracleText:    "When this creature enters, you may search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.TriggeredAbility,
			Text: "When this creature enters, you may search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhen,
				Pattern: game.TriggerPattern{
					Event:  game.EventPermanentEnteredBattlefield,
					Source: game.TriggerSourceSelf,
				},
			}),
			Optional: true,
			Effects: []game.Effect{
				{
					Type:        game.EffectSearch,
					TargetIndex: game.TargetIndexController,
					Search: opt.Val(game.SearchSpec{
						SourceZone:   game.ZoneLibrary,
						Destination:  game.ZoneBattlefield,
						CardType:     opt.Val(types.Land),
						Supertype:    opt.Val(types.Basic),
						EntersTapped: true,
						Shuffle:      true,
					}),
				},
			},
		},
	},
}
