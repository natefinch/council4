package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Bushwhack
//
// Type: Sorcery
// Cost: {G}
//
// Oracle text:
//
//	Choose one —
//	• Search your library for a basic land card, reveal it, put it into your hand, then shuffle.
//	• Target creature you control fights target creature you don't control. (Each deals damage equal to its power to the other.)
//
// Missing primitives:
//   - SearchSpec has no MatchSupertype field; "basic" cannot be enforced
//     declaratively — the search allows any land card.
var Bushwhack = &game.CardDef{
	Name: "Bushwhack",
	ManaCost: opt.Val(mana.Cost{
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     1,
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Types:         []game.CardType{game.TypeSorcery},
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
							// "basic" supertype not enforceable; searches any land.
							Type:        game.EffectSearch,
							TargetIndex: -1,
							Search: opt.Val(game.SearchSpec{
								SourceZone:    game.ZoneLibrary,
								Destination:   game.ZoneHand,
								MatchCardType: true,
								CardType:      game.TypeLand,
								Reveal:        true,
								Shuffle:       true,
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
								PermanentTypes: []game.CardType{game.TypeCreature},
								Controller:     game.ControllerYou,
							},
						},
						{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "creature you don't control",
							Allow:      game.TargetAllowPermanent,
							Predicate: game.TargetPredicate{
								PermanentTypes: []game.CardType{game.TypeCreature},
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
