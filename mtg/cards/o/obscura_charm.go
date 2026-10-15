package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ObscuraCharm is the card definition for Obscura Charm.
//
// Type: Instant
// Cost: {W}{U}{B}
//
// Oracle text:
//
//	Choose one —
//	• Return target multicolored permanent card with mana value 3 or less from your graveyard to the battlefield tapped.
//	• Counter target instant or sorcery spell.
//	• Destroy target creature or planeswalker with mana value 3 or less.
var ObscuraCharm = newObscuraCharm

func newObscuraCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Obscura Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Return target multicolored permanent card with mana value 3 or less from your graveyard to the battlefield tapped.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target multicolored permanent card with mana value 3 or less from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, Multicolored: true, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source:      game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
									EntryTapped: true,
								},
							},
						},
					},
					game.Mode{
						Text: "Counter target instant or sorcery spell.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target instant or sorcery spell",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
									StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
								},
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CounterObject{
									Object: game.TargetStackObjectReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Destroy target creature or planeswalker with mana value 3 or less.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature or planeswalker with mana value 3 or less",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Return target multicolored permanent card with mana value 3 or less from your graveyard to the battlefield tapped.
			• Counter target instant or sorcery spell.
			• Destroy target creature or planeswalker with mana value 3 or less.
		`,
		},
	}
}
