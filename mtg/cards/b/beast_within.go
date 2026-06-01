package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Beast Within
//
// Type: Instant
// Cost: {2}{G}
//
// Oracle text:
//
//	Destroy target permanent. Its controller creates a 3/3 green Beast creature token.
//
// Missing primitives:
//   - EffectCreateToken always creates the token for the spell's controller
//     (r.obj.Controller). There is no TargetIndex or "controlled-by-target"
//     recipient field on EffectCreateToken, so the token cannot be assigned to
//     the destroyed permanent's controller declaratively.
//     ImplementationID "beast-within" is set so a hand-written handler can
//     issue the token to the correct player.
var BeastWithin = &game.CardDef{
	Name: "Beast Within",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(2),
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:        3,
	Colors:           []mana.Color{mana.Green},
	ColorIdentity:    mana.NewColorIdentity(mana.Green),
	Types:            []game.CardType{game.TypeInstant},
	OracleText:       "Destroy target permanent. Its controller creates a 3/3 green Beast creature token.",
	ImplementationID: "beast-within",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "Destroy target permanent. Its controller creates a 3/3 green Beast creature token.",
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "permanent",
					Allow:      game.TargetAllowPermanent,
				},
			},
			Effects: []game.Effect{
				{Type: game.EffectDestroy, TargetIndex: 0},
				// Token owner is wrong here (spell controller instead of
				// target's controller); the ImplementationID handler corrects this.
				{Type: game.EffectCreateToken, Amount: 1, Token: opt.Val(beastWithinToken)},
			},
		},
	},
}

var beastWithinToken = &game.CardDef{
	Name:      "Beast",
	Colors:    []mana.Color{mana.Green},
	Types:     []game.CardType{game.TypeCreature},
	Subtypes:  []string{"Beast"},
	Power:     opt.Val(game.PT{Value: 3}),
	Toughness: opt.Val(game.PT{Value: 3}),
}
