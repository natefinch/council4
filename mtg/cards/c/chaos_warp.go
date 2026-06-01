package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Chaos Warp
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield.
//
// Missing primitives:
//   - No EffectType for "shuffle a permanent into its owner's library" (not Bounce/Exile/Destroy).
//   - No EffectReveal for "reveals the top card of their library."
//   - The conditional "if it's a permanent card, put it onto the battlefield" requires checking
//     the card type of the newly revealed top card, which EffectCondition/EffectResultCondition
//     cannot express declaratively. ImplementationID "chaos-warp" must handle all three steps.
var ChaosWarp = &game.CardDef{
	Name: "Chaos Warp",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(2),
		mana.ColoredMana(mana.Red),
	}),
	ManaValue:        3,
	Colors:           []mana.Color{mana.Red},
	ColorIdentity:    mana.NewColorIdentity(mana.Red),
	Types:            []game.CardType{game.TypeInstant},
	OracleText:       "The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield.",
	ImplementationID: "chaos-warp",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield.",
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "permanent",
					Allow:      game.TargetAllowPermanent,
				},
			},
			// No declarative Effects: shuffle-into-library, reveal-top-card, and
			// conditional put-onto-battlefield have no EffectType primitives.
			// ImplementationID "chaos-warp" handles all three steps.
		},
	},
}
