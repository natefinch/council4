package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// exciseSequence is the ordered instruction pair the cardgen backend emits for
// "Exile target nonland permanent. Its controller incubates X, where X is its
// mana value." (Excise the Imperfect): exile the target, then have the exiled
// permanent's last-known controller incubate its last-known mana value.
func exciseSequence() []game.Instruction {
	return []game.Instruction{
		{Primitive: game.Exile{Object: game.TargetPermanentReference(0)}},
		{Primitive: game.Incubate{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:       game.DynamicAmountObjectManaValue,
				Multiplier: 1,
				Object:     game.TargetPermanentReference(0),
			}),
			Recipient: opt.Val(game.ObjectControllerReference(game.TargetPermanentReference(0))),
		}},
	}
}

// findIncubatorToken returns the single Incubator token on the battlefield, or
// nil when none exists.
func findIncubatorToken(g *game.Game) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == "Incubator" {
			return permanent
		}
	}
	return nil
}

func addManaValuePermanent(g *game.Game, owner, controller game.PlayerID, manaValue int) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Owner: owner,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:     "MV Permanent",
			Types:    []types.Card{types.Artifact},
			ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
		}},
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          owner,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

// TestIncubateExileUsesStolenControllerNotOwner proves the LKI resolution the
// incubate sequence relies on: when the exiled permanent was controlled by a
// player other than its owner, the incubate recipient is the last-known
// controller (who stole it), not the owner, and the counter count is the exiled
// permanent's last-known mana value.
func TestIncubateExileUsesStolenControllerNotOwner(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Player2 owns the permanent, but Player1 controls it (stole it) with mana value 4.
	stolen := addManaValuePermanent(g, game.Player2, game.Player1, 4)
	addInstructionSpellToStackForController(g, game.Player2, exciseSequence(), []game.Target{
		game.PermanentTarget(stolen.ObjectID),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, stolen.ObjectID); ok {
		t.Fatal("stolen permanent remained on the battlefield after exile")
	}
	if !g.Players[game.Player2].Exile.Contains(stolen.CardInstanceID) {
		t.Fatal("exiled permanent did not move to its owner's exile zone")
	}
	incubator := findIncubatorToken(g)
	if incubator == nil {
		t.Fatal("no Incubator token was created")
	}
	if incubator.Controller != game.Player1 {
		t.Fatalf("Incubator controller = %d, want Player1 (last-known controller, not owner)", incubator.Controller)
	}
	if got := incubator.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("Incubator +1/+1 counters = %d, want 4 (exiled permanent's mana value)", got)
	}
}

// TestIncubateTokenTargetManaValueZeroCreatesEmptyIncubator proves incubate 0
// still creates an Incubator token (CR 701.55a places no minimum): exiling a
// token, whose mana value is 0, incubates 0 and mints an Incubator with no
// counters for the token's controller.
func TestIncubateTokenTargetManaValueZeroCreatesEmptyIncubator(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := addTokenCreaturePermanent(g, game.Player1, "Bear")
	addInstructionSpellToStackForController(g, game.Player1, exciseSequence(), []game.Target{
		game.PermanentTarget(token.ObjectID),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, token.ObjectID); ok {
		t.Fatal("targeted token remained on the battlefield after exile")
	}
	incubator := findIncubatorToken(g)
	if incubator == nil {
		t.Fatal("no Incubator token was created for incubate 0")
	}
	if incubator.Controller != game.Player1 {
		t.Fatalf("Incubator controller = %d, want Player1", incubator.Controller)
	}
	if got := incubator.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("Incubator +1/+1 counters = %d, want 0 (token mana value is 0)", got)
	}
}

// TestIncubateFizzlesWhenTargetLeavesBeforeResolution proves the single-target
// spell is countered on resolution when its only target has left the
// battlefield, so nothing is exiled and no Incubator token is created.
func TestIncubateFizzlesWhenTargetLeavesBeforeResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)

	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID:    sourceID,
		Owner: game.Player1,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Excise the Imperfect",
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
				Sequence: exciseSequence(),
			}.Ability()),
		}},
	}
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   sourceID,
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	})

	// Target leaves the battlefield in response, before the spell resolves.
	movePermanentToZone(g, target, zone.Graveyard)
	engine.resolveTopOfStack(g, &TurnLog{})

	if findIncubatorToken(g) != nil {
		t.Fatal("Incubator token was created for a fizzled spell")
	}
}

// TestIncubateCountersCarryThroughTransform proves the Incubator token created
// by incubate transforms into a 0/0 colorless Phyrexian artifact creature whose
// +1/+1 counters (placed by incubate) survive the transform, so the creature
// side is a 4/4.
func TestIncubateCountersCarryThroughTransform(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addManaValuePermanent(g, game.Player1, game.Player1, 4)
	addInstructionSpellToStackForController(g, game.Player1, exciseSequence(), []game.Target{
		game.PermanentTarget(source.ObjectID),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	incubator := findIncubatorToken(g)
	if incubator == nil {
		t.Fatal("no Incubator token was created")
	}
	if got := incubator.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("Incubator +1/+1 counters = %d, want 4", got)
	}
	if !transformPermanent(g, incubator) {
		t.Fatal("Incubator token did not transform")
	}
	power := effectivePower(g, incubator)
	toughness, okT := effectiveToughness(g, incubator)
	if !okT {
		t.Fatal("transformed Incubator has undefined toughness")
	}
	if power != 4 || toughness != 4 {
		t.Fatalf("transformed Incubator P/T = %d/%d, want 4/4 (0/0 base plus four +1/+1 counters)", power, toughness)
	}
}
