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
var BeastWithin = &game.CardDef{
	Name: "Beast Within",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(2),
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     3,
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Types:         []game.CardType{game.TypeInstant},
	OracleText:    "Destroy target permanent. Its controller creates a 3/3 green Beast creature token.",
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
				{
					Type:   game.EffectCreateToken,
					Amount: 1,
					Token:  opt.Val(beastWithinToken),
					Recipient: opt.Val(game.PlayerReference{
						Kind: game.PlayerReferenceObjectController,
						Object: opt.Val(game.ObjectReference{
							Kind:        game.ObjectReferenceTargetPermanent,
							TargetIndex: 0,
						}),
					}),
				},
			},
		},
	},
}

var beastWithinToken = &game.CardDef{
	Name:      "Beast",
	Colors:    []mana.Color{mana.Green},
	Types:     []game.CardType{game.TypeCreature},
	Subtypes:  []string{game.CreatureSubtypeBeast},
	Power:     opt.Val(game.PT{Value: 3}),
	Toughness: opt.Val(game.PT{Value: 3}),
}
