package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MorbidOpportunist is the card definition for Morbid Opportunist.
//
// Type: Creature — Human Rogue
// Cost: {2}{B}
//
// Oracle text:
//
//	Whenever one or more other creatures die, draw a card. This ability triggers only once each turn.
var MorbidOpportunist = &game.CardDef{
	Name: "Morbid Opportunist",
	ManaCost: opt.Val(cost.Mana{
		cost.O(2),
		cost.B,
	}),
	Colors:        []color.Color{color.Black},
	ColorIdentity: mana.NewColorIdentity(color.Black),
	Types:         []types.Card{types.Creature},
	Subtypes:      []types.Sub{types.Human, types.Rogue},
	Power:         opt.Val(game.PT{Value: 1}),
	Toughness:     opt.Val(game.PT{Value: 3}),
	OracleText:    "Whenever one or more other creatures die, draw a card. This ability triggers only once each turn.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.TriggeredAbility,
			Text: "Whenever one or more other creatures die, draw a card. This ability triggers only once each turn.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                 game.EventPermanentDied,
					ExcludeSelf:           true,
					RequirePermanentTypes: []types.Card{types.Creature},
					OneOrMore:             true,
				},
			}),
			MaxTriggersPerTurn: 1,
			Effects: []game.Effect{
				{Type: game.EffectDraw, Amount: 1, TargetIndex: game.TargetIndexController},
			},
		},
	},
}
