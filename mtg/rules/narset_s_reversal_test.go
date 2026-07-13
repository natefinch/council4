package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// narsetReversalInstructions is the copy-then-return stack-spell sequence that
// Narset's Reversal lowers to: copy the targeted instant or sorcery spell (with
// the "you may choose new targets for the copy" rider), then return the same
// targeted spell to its owner's hand. The copy clause addresses the stack object
// it duplicates; the return clause addresses the target object it removes.
func narsetReversalInstructions() []game.Instruction {
	return []game.Instruction{
		{Primitive: game.CopyStackObject{Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true}},
		{Primitive: game.Bounce{Object: game.TargetObjectReference(0)}},
	}
}

// addVictimSpellOnStack pushes an instant or sorcery spell onto the stack with a
// distinct owner (whose hand the return clause targets) and controller (who cast
// it). It returns the stack object so tests can target it and inspect the copy.
func addVictimSpellOnStack(g *game.Game, owner, controller game.PlayerID, def *game.CardDef) *game.StackObject {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: owner}
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   cardID,
		Controller: controller,
	}
	g.Stack.Push(obj)
	return obj
}

func simpleSorceryDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Sorcery}}}
}

// TestNarsetsReversalCopiesSpellAndReturnsOriginalToHand covers the core
// behavior: resolving Narset's Reversal against an instant or sorcery spell puts
// an independent copy under the caster's control on the stack and returns the
// original spell to its owner's hand (not the graveyard, and not as a counter).
func TestNarsetsReversalCopiesSpellAndReturnsOriginalToHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	victim := addVictimSpellOnStack(g, game.Player2, game.Player2, simpleSorceryDef("Victim Sorcery"))
	victimCardID := victim.SourceID

	depthBefore := g.Stack.Size()
	addInstructionSpellToStackForController(g, game.Player1, narsetReversalInstructions(),
		[]game.Target{game.StackObjectTarget(victim.ID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, victim.ID); ok {
		t.Fatal("original spell remained on the stack after return")
	}
	// The Narset spell was consumed and the copy pushed, so the stack keeps the
	// same depth it had with the victim plus Narset (copy replaces Narset's slot,
	// original removed).
	if got := g.Stack.Size(); got != depthBefore {
		t.Fatalf("stack size = %d, want %d", got, depthBefore)
	}
	top, ok := g.Stack.Peek()
	if !ok || !top.Copy {
		t.Fatal("copy of the spell not on top of stack")
	}
	if top.SourceID != victim.SourceID {
		t.Fatalf("copy SourceID = %v, want %v", top.SourceID, victim.SourceID)
	}
	if top.Controller != game.Player1 {
		t.Fatalf("copy controller = %v, want Player1 (the caster)", top.Controller)
	}
	if !g.Players[game.Player2].Hand.Contains(victimCardID) {
		t.Fatal("returned spell did not move to its owner's hand")
	}
	if g.Players[game.Player2].Graveyard.Contains(victimCardID) {
		t.Fatal("returned spell moved to graveyard, want hand")
	}
}

// TestNarsetsReversalReturnsToOwnerNotController proves the return clause sends
// the card to its owner's hand, not the controller's, while the copy is created
// under the Narset caster's control. Owner (Player3), controller (Player2), and
// caster (Player1) are all distinct.
func TestNarsetsReversalReturnsToOwnerNotController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	victim := addVictimSpellOnStack(g, game.Player3, game.Player2, simpleSorceryDef("Borrowed Spell"))
	victimCardID := victim.SourceID

	addInstructionSpellToStackForController(g, game.Player1, narsetReversalInstructions(),
		[]game.Target{game.StackObjectTarget(victim.ID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player3].Hand.Contains(victimCardID) {
		t.Fatal("returned spell did not go to its owner's (Player3) hand")
	}
	if g.Players[game.Player2].Hand.Contains(victimCardID) {
		t.Fatal("returned spell went to the controller's (Player2) hand, want owner's")
	}
	top, ok := g.Stack.Peek()
	if !ok || !top.Copy {
		t.Fatal("copy of the spell not on top of stack")
	}
	if top.Controller != game.Player1 {
		t.Fatalf("copy controller = %v, want Player1 (the caster)", top.Controller)
	}
}

// TestNarsetsReversalCopyPreservesModesXAndTargets proves the copy carries over
// the original spell's chosen modes, X value, and full target list (CR 707.10a
// / 707.2). The victim declares no retargetable target specs, so the "may choose
// new targets" rider is a no-op and the copy keeps every original target.
func TestNarsetsReversalCopyPreservesModesXAndTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	creatureA := addCreaturePermanent(g, game.Player2)
	creatureB := addCreaturePermanent(g, game.Player2)

	victim := addVictimSpellOnStack(g, game.Player2, game.Player2, simpleSorceryDef("Modal X Spell"))
	victim.ChosenModes = []int{0, 2}
	victim.ResolvedAmounts = map[string]int{"X": 3}
	victim.Targets = []game.Target{
		game.PermanentTarget(creatureA.ObjectID),
		game.PermanentTarget(creatureB.ObjectID),
	}
	victim.TargetCounts = []int{1, 1}

	addInstructionSpellToStackForController(g, game.Player1, narsetReversalInstructions(),
		[]game.Target{game.StackObjectTarget(victim.ID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok || !top.Copy {
		t.Fatal("copy of the spell not on top of stack")
	}
	if len(top.ChosenModes) != 2 || top.ChosenModes[0] != 0 || top.ChosenModes[1] != 2 {
		t.Fatalf("copy modes = %v, want [0 2]", top.ChosenModes)
	}
	if top.ResolvedAmounts["X"] != 3 {
		t.Fatalf("copy X = %d, want 3", top.ResolvedAmounts["X"])
	}
	if len(top.Targets) != 2 ||
		top.Targets[0].PermanentID != creatureA.ObjectID ||
		top.Targets[1].PermanentID != creatureB.ObjectID {
		t.Fatalf("copy targets = %+v, want both original creatures", top.Targets)
	}
}

// TestNarsetsReversalReturningACopyMakesItCeaseWithoutCard proves that when the
// targeted spell is itself a copy, the return clause removes it from the stack
// and it simply ceases to exist (CR 707.10c) — no card moves to any hand — while
// a fresh copy is still created by the copy clause.
func TestNarsetsReversalReturningACopyMakesItCeaseWithoutCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	victim := addVictimSpellOnStack(g, game.Player2, game.Player2, simpleSorceryDef("Copied Spell"))
	victim.Copy = true
	victimCardID := victim.SourceID

	addInstructionSpellToStackForController(g, game.Player1, narsetReversalInstructions(),
		[]game.Target{game.StackObjectTarget(victim.ID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, victim.ID); ok {
		t.Fatal("returned copy remained on the stack, want ceased to exist")
	}
	if g.Players[game.Player2].Hand.Contains(victimCardID) {
		t.Fatal("returning a copy put a card in hand, want no card moved")
	}
	top, ok := g.Stack.Peek()
	if !ok || !top.Copy {
		t.Fatal("new copy of the spell not on top of stack")
	}
	if top.ID == victim.ID {
		t.Fatal("new copy shares the returned copy's ID, want a distinct object")
	}
}

// TestNarsetsReversalBouncesUncounterableSpell proves the return is a bounce, not
// a counter: a spell that can't be countered is still returned to its owner's
// hand, because bounceStackSpellToHand never consults counterability.
func TestNarsetsReversalBouncesUncounterableSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	victim := addVictimSpellOnStack(g, game.Player2, game.Player2, simpleSorceryDef("Unstoppable Spell"))
	victim.RuleEffects = []game.RuleEffect{{Kind: game.RuleEffectCantBeCountered}}
	victimCardID := victim.SourceID

	addInstructionSpellToStackForController(g, game.Player1, narsetReversalInstructions(),
		[]game.Target{game.StackObjectTarget(victim.ID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, victim.ID); ok {
		t.Fatal("uncounterable spell remained on the stack after return")
	}
	if !g.Players[game.Player2].Hand.Contains(victimCardID) {
		t.Fatal("uncounterable spell was not returned to its owner's hand")
	}
}

// TestNarsetsReversalFizzlesWhenTargetLeavesStack proves that if the targeted
// spell is no longer on the stack at resolution, neither the copy nor the return
// happens (CR 608.2b): no copy is created and nothing moves to any hand.
func TestNarsetsReversalFizzlesWhenTargetLeavesStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	victim := addVictimSpellOnStack(g, game.Player2, game.Player2, simpleSorceryDef("Fleeting Spell"))
	victimCardID := victim.SourceID

	addInstructionSpellToStackForController(g, game.Player1, narsetReversalInstructions(),
		[]game.Target{game.StackObjectTarget(victim.ID)})

	// The target leaves the stack (resolves/countered) before Narset resolves.
	g.Stack.RemoveByID(victim.ID)

	depthBefore := g.Stack.Size()
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Stack.Size(); got != depthBefore-1 {
		t.Fatalf("stack size = %d, want %d (only Narset removed, no copy pushed)", got, depthBefore-1)
	}
	if top, ok := g.Stack.Peek(); ok && top.Copy {
		t.Fatal("a copy was created even though the target was gone")
	}
	if g.Players[game.Player2].Hand.Contains(victimCardID) {
		t.Fatal("a card was returned even though the target was gone")
	}
}

// TestNarsetsReversalMayChooseNewTargetsForCopy proves the "you may choose new
// targets for the copy" rider retargets the copy only. The victim targets
// creature A; the caster chooses creature B for the copy; the original is
// returned to hand while the copy targets creature B.
func TestNarsetsReversalMayChooseNewTargetsForCopy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	creatureA := addCreaturePermanent(g, game.Player2)
	creatureB := addCreaturePermanent(g, game.Player2)

	spellDef := &game.CardDef{CardFace: game.CardFace{
		Name:  "Shock",
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      game.TargetAllowPermanent,
				Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			}},
			Sequence: []game.Instruction{{
				Primitive: game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0)},
			}},
		}.Ability()),
	}}
	victim := addVictimSpellOnStack(g, game.Player2, game.Player2, spellDef)
	victim.Targets = []game.Target{game.PermanentTarget(creatureA.ObjectID)}
	victim.TargetCounts = []int{1}
	victimCardID := victim.SourceID

	addInstructionSpellToStackForController(g, game.Player1, narsetReversalInstructions(),
		[]game.Target{game.StackObjectTarget(victim.ID)})

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok || !top.Copy {
		t.Fatal("copy of the spell not on top of stack")
	}
	if len(top.Targets) != 1 || top.Targets[0].PermanentID != creatureB.ObjectID {
		t.Fatalf("copy targets = %+v, want retargeted to creature B %v", top.Targets, creatureB.ObjectID)
	}
	if !g.Players[game.Player2].Hand.Contains(victimCardID) {
		t.Fatal("original spell was not returned to its owner's hand")
	}
}
