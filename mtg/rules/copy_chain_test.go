package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// copyChainPaidTestKey wires the copy-chain resolution Pay to its gated copy in
// the runtime tests, mirroring the generated copy-chain instruction shape.
const copyChainPaidTestKey = game.ResultKey("copy-chain-paid")

// copyChainGatedInstructions models the resolved payment-gated copy-chain body
// (String of Disappearances, Chain Lightning, Chain Stasis): the affected
// target's controller may pay a mana cost, and only if they pay may they copy
// the spell, with the copy controlled by that affected controller so its own
// iterative offer chains off the copier's new target.
func copyChainGatedInstructions() []game.Instruction {
	return []game.Instruction{
		{
			Primitive: game.Pay{Payment: game.ResolutionPayment{
				Prompt:   "Pay {2}{U}?",
				Payer:    opt.Val(game.AffectedTargetControllerReference(0)),
				ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U}),
			}},
			PublishResult: copyChainPaidTestKey,
		},
		{
			Primitive: game.CopyStackObject{
				Object:  game.ResolvingStackObjectReference(),
				Chooser: opt.Val(game.AffectedTargetControllerReference(0)),
			},
			Optional:      true,
			OptionalActor: opt.Val(game.AffectedTargetControllerReference(0)),
			ResultGate:    opt.Val(game.InstructionResultGate{Key: copyChainPaidTestKey, Succeeded: game.TriTrue}),
		},
	}
}

// TestCopyStackObjectChooserControlsCopy verifies that a CopyStackObject whose
// Chooser is the affected target's controller puts the copy under that player's
// control (CR 707.10a), not the resolving spell's controller. The spell is
// controlled by Player1 but targets a permanent controlled by Player2, so the
// copy's controller must be Player2.
func TestCopyStackObjectChooserControlsCopy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCreaturePermanent(g, game.Player2)
	addInstructionSpellToStackForController(g, game.Player1,
		[]game.Instruction{{
			Primitive: game.CopyStackObject{
				Object:  game.ResolvingStackObjectReference(),
				Chooser: opt.Val(game.AffectedTargetControllerReference(0)),
			},
		}},
		[]game.Target{game.PermanentTarget(victim.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok || !top.Copy {
		t.Fatal("copy not on top of stack after resolving the chooser copy")
	}
	if top.Controller != game.Player2 {
		t.Fatalf("copy controller = %v, want Player2 (the affected target's controller)", top.Controller)
	}
}

// TestCopyChainChooserResolvesAfterBounce verifies that the affected target's
// controller is resolved via last-known-information: String of Disappearances
// returns the target creature to its owner's hand and only then offers the copy,
// so the permanent has already left the battlefield when the copy's Chooser
// resolves. The pre-bounce controller (Player2) must still control the copy.
func TestCopyChainChooserResolvesAfterBounce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCreaturePermanent(g, game.Player2)
	addInstructionSpellToStackForController(g, game.Player1,
		[]game.Instruction{
			{Primitive: game.Bounce{Object: game.TargetPermanentReference(0)}},
			{
				Primitive: game.CopyStackObject{
					Object:  game.ResolvingStackObjectReference(),
					Chooser: opt.Val(game.AffectedTargetControllerReference(0)),
				},
			},
		},
		[]game.Target{game.PermanentTarget(victim.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok || !top.Copy {
		t.Fatal("copy not on top of stack after resolving the chooser copy")
	}
	if top.Controller != game.Player2 {
		t.Fatalf("copy controller = %v, want Player2 (the bounced target's last-known controller)", top.Controller)
	}
}

// target's controller not paying the copy-chain resolution cost (here, having no
// mana) skips the result-gated copy: only the original leaves the stack.
func TestCopyChainPaymentGateBlocksCopyWhenUnpaid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCreaturePermanent(g, game.Player2)
	addInstructionSpellToStackForController(g, game.Player1,
		copyChainGatedInstructions(),
		[]game.Target{game.PermanentTarget(victim.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0 (original popped, no copy when the affected player cannot pay)", got)
	}
}

// TestCopyChainPaymentGateCopiesWhenPaid verifies that the affected target's
// controller paying the copy-chain resolution cost pushes the result-gated copy
// under that player's control. Player2 controls the targeted creature and enough
// Islands to pay {2}{U}, and accepts both the payment and the copy offer.
func TestCopyChainPaymentGateCopiesWhenPaid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCreaturePermanent(g, game.Player2)
	addBasicLandPermanent(g, game.Player2, types.Island)
	addBasicLandPermanent(g, game.Player2, types.Island)
	addBasicLandPermanent(g, game.Player2, types.Island)
	addInstructionSpellToStackForController(g, game.Player1,
		copyChainGatedInstructions(),
		[]game.Target{game.PermanentTarget(victim.ObjectID)})
	agents := [game.NumPlayers]PlayerAgent{game.Player2: optionalMayAgent{accept: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok || !top.Copy {
		t.Fatal("copy not on top of stack after the affected player paid")
	}
	if top.Controller != game.Player2 {
		t.Fatalf("copy controller = %v, want Player2 (the affected target's controller and payer)", top.Controller)
	}
}
