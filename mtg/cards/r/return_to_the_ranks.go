package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ReturnToTheRanks is the card definition for Return to the Ranks.
//
// Type: Sorcery
// Cost: {X}{W}{W}
//
// Oracle text:
//
//	Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
//	Return X target creature cards with mana value 2 or less from your graveyard to the battlefield.
var ReturnToTheRanks = newReturnToTheRanks

func newReturnToTheRanks() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Return to the Ranks",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.ConvokeStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:   0,
						MaxTargets:   20,
						Constraint:   "target creature cards with mana value 2 or less from your graveyard",
						Allow:        game.TargetAllowCard,
						TargetZone:   zone.Graveyard,
						Selection:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})}),
						CountEqualsX: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 2}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 3}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 4}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 5}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 6}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 7}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 8}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 9}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 10}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 11}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 12}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 13}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 14}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 15}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 16}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 17}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 18}),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 19}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
			Return X target creature cards with mana value 2 or less from your graveyard to the battlefield.
		`,
		},
	}
}
