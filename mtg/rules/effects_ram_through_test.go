package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// ramThroughInstructions builds the two mutually exclusive Damage instructions
// Ram Through lowers to: the dealer (target 0) deals damage equal to its power
// to the bitten creature (target 1), and when the dealer has trample the excess
// beyond lethal is redirected to the bitten creature's controller. Exactly one
// branch fires per resolution, gated on the dealer having trample.
func ramThroughInstructions() []game.Instruction {
	dealerPower := game.Dynamic(game.DynamicAmount{
		Kind:       game.DynamicAmountObjectPower,
		Multiplier: 1,
		Object:     game.TargetPermanentReference(0),
	})
	trampleMatch := opt.Val(game.Selection{Keyword: game.Trample})
	return []game.Instruction{
		{
			Primitive: game.Damage{
				Amount:          dealerPower,
				Recipient:       game.AnyTargetDamageRecipient(1),
				DamageSource:    opt.Val(game.TargetPermanentReference(0)),
				ExcessRecipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(1))),
			},
			Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
				Object:        opt.Val(game.TargetPermanentReference(0)),
				ObjectMatches: trampleMatch,
			})}),
		},
		{
			Primitive: game.Damage{
				Amount:       dealerPower,
				Recipient:    game.AnyTargetDamageRecipient(1),
				DamageSource: opt.Val(game.TargetPermanentReference(0)),
			},
			Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
				Negate:        true,
				Object:        opt.Val(game.TargetPermanentReference(0)),
				ObjectMatches: trampleMatch,
			})}),
		},
	}
}

// TestRamThroughTrampleRedirectsExcessToController proves Ram Through's runtime
// behavior end to end over the two gated branches: a 5-power dealer with trample
// biting an indestructible 4/4 marks only 4 lethal damage and redirects the 1
// excess to the bitten creature's controller, while the same dealer without
// trample marks the full 5 and spares the controller.
func TestRamThroughTrampleRedirectsExcessToController(t *testing.T) {
	cases := []struct {
		name              string
		trample           bool
		wantMarked        int
		wantControllerHit int
	}{
		{"with trample redirects excess", true, 4, 1},
		{"without trample deals all to creature", false, 5, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			var dealer *game.Permanent
			if tc.trample {
				dealer = addCombatCreaturePermanentWithPower(g, game.Player1, 5, game.Trample)
			} else {
				dealer = addCombatCreaturePermanentWithPower(g, game.Player1, 5)
			}
			victim := addCombatCreaturePermanentWithPower(g, game.Player2, 4, game.Indestructible)
			beforeP2 := g.Players[game.Player2].Life

			addInstructionSpellToStackForController(g, game.Player1, ramThroughInstructions(), []game.Target{
				game.PermanentTarget(dealer.ObjectID),
				game.PermanentTarget(victim.ObjectID),
			})

			engine.resolveTopOfStack(g, &TurnLog{})

			if victim.MarkedDamage != tc.wantMarked {
				t.Fatalf("victim marked damage = %d, want %d", victim.MarkedDamage, tc.wantMarked)
			}
			if lost := beforeP2 - g.Players[game.Player2].Life; lost != tc.wantControllerHit {
				t.Fatalf("controller life lost = %d, want %d", lost, tc.wantControllerHit)
			}
			if dealer.MarkedDamage != 0 {
				t.Fatalf("dealer marked damage = %d, want 0", dealer.MarkedDamage)
			}
		})
	}
}
