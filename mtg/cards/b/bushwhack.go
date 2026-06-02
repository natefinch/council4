package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Bushwhack is the card definition for Bushwhack.
//
// Type: Sorcery
// Cost: {G}
//
// Oracle text:
//
//	Choose one —
//	• Search your library for a basic land card, reveal it, put it into your hand, then shuffle.
//	• Target creature you control fights target creature you don't control. (Each deals damage equal to its power to the other.)
var Bushwhack = &game.CardDef{
	Name: "Bushwhack",
	ManaCost: opt.Val(mana.Cost{
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     1,
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Types:         []types.Card{types.Sorcery},
	OracleText:    "Choose one —\n• Search your library for a basic land card, reveal it, put it into your hand, then shuffle.\n• Target creature you control fights target creature you don't control. (Each deals damage equal to its power to the other.)",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "Choose one —\n• Search your library for a basic land card, reveal it, put it into your hand, then shuffle.\n• Target creature you control fights target creature you don't control.",
			Modes: []game.Mode{
				{
					Text: "Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
					Effects: []game.Effect{
						{
							Type:        game.EffectSearch,
							TargetIndex: game.TargetIndexController,
							Search: opt.Val(game.SearchSpec{
								SourceZone:  game.ZoneLibrary,
								Destination: game.ZoneHand,
								CardType:    opt.Val(types.Land),
								Supertype:   opt.Val(types.Basic),
								Reveal:      true,
								Shuffle:     true,
							}),
						},
					},
				},
				{
					Text: "Target creature you control fights target creature you don't control.",
					Targets: []game.TargetSpec{
						{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "creature you control",
							Allow:      game.TargetAllowPermanent,
							Predicate: game.TargetPredicate{
								PermanentTypes: []types.Card{types.Creature},
								Controller:     game.ControllerYou,
							},
						},
						{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "creature you don't control",
							Allow:      game.TargetAllowPermanent,
							Predicate: game.TargetPredicate{
								PermanentTypes: []types.Card{types.Creature},
								Controller:     game.ControllerNotYou,
							},
						},
					},
					Effects: []game.Effect{
						{Type: game.EffectFight},
					},
				},
			},
		},
	},
}
