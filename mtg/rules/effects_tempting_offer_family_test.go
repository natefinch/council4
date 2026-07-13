package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// temptingOfferCounterInstruction models the Tempt with Glory idiom: the acting
// player puts a +1/+1 counter on each creature they control. The acting player is
// addressed through GroupOfferMemberReference() so the counters land on the
// controller's creatures (base and reward) or the accepting opponent's creatures.
func temptingOfferCounterInstruction() game.Instruction {
	return game.Instruction{
		Primitive: game.AddCounter{
			Amount:      game.Fixed(1),
			Group:       game.PlayerControlledGroup(game.GroupOfferMemberReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			CounterKind: counter.PlusOnePlusOne,
		},
		Optional:           true,
		OptionalActorGroup: opt.Val(game.OpponentsReference()),
		TemptingOffer:      true,
	}
}

// temptingOfferReturnInstruction models the Tempt with Immortality idiom: the
// acting player returns a creature card from their graveyard to the battlefield.
// The acting player is addressed through GroupOfferMemberReference() so the
// reanimated creature comes from and enters under that player's control.
func temptingOfferReturnInstruction() game.Instruction {
	return game.Instruction{
		Primitive: game.ReturnFromGraveyardChoice(
			game.GroupOfferMemberReference(),
			game.Selection{RequiredTypes: []types.Card{types.Creature}},
			game.Fixed(1),
			zone.Battlefield,
			false,
			opt.V[int]{},
			false,
			"",
		),
		Optional:           true,
		OptionalActorGroup: opt.Val(game.OpponentsReference()),
		TemptingOffer:      true,
	}
}

// temptingOfferCopyInstruction models the Tempt with Reflections idiom: the acting
// player creates a token that is a copy of the controller's single target
// creature. The copy enters under the acting player's control through
// GroupOfferMemberReference() while every resolution copies the same target.
func temptingOfferCopyInstruction() game.Instruction {
	return game.Instruction{
		Primitive: game.CreateToken{
			Amount: game.Fixed(1),
			Source: game.TokenCopyOf(game.TokenCopySpec{
				Source: game.TokenCopySourceObject,
				Object: game.TargetPermanentReference(0),
			}),
			Recipient: opt.Val(game.GroupOfferMemberReference()),
		},
		Optional:           true,
		OptionalActorGroup: opt.Val(game.OpponentsReference()),
		TemptingOffer:      true,
	}
}

func plusOneCounterCount(permanent *game.Permanent) int {
	return permanent.Counters.Get(counter.PlusOnePlusOne)
}

func addCreatureCardToGraveyard(g *game.Game, playerID game.PlayerID) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Graveyard Creature",
			Types: []types.Card{types.Creature},
		}},
		Owner: playerID,
	}
	g.Players[playerID].Graveyard.Add(cardID)
	return cardID
}

func creaturesByController(g *game.Game) map[game.PlayerID]int {
	counts := make(map[game.PlayerID]int)
	for _, permanent := range g.Battlefield {
		counts[permanent.Controller]++
	}
	return counts
}

// TestTemptingOfferCounterActorOwnershipAndRepeat proves the +1/+1 counter idiom
// (Tempt with Glory) puts counters on each acting player's own creatures: the
// controller's creature gains one counter for the base plus one per accepting
// opponent, each accepting opponent's creature gains one, and a declining
// opponent's creature gains none.
func TestTemptingOfferCounterActorOwnershipAndRepeat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatures := map[game.PlayerID]*game.Permanent{
		game.Player1: addCreaturePermanent(g, game.Player1),
		game.Player2: addCreaturePermanent(g, game.Player2),
		game.Player3: addCreaturePermanent(g, game.Player3),
		game.Player4: addCreaturePermanent(g, game.Player4),
	}
	addInstructionSpellToStack(g, []game.Instruction{temptingOfferCounterInstruction()})

	// Player2 and Player3 accept; Player4 declines.
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{1}}},
		game.Player4: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	// Controller: one base counter plus one reward per accepting opponent (2) = 3.
	if got := plusOneCounterCount(creatures[game.Player1]); got != 3 {
		t.Fatalf("controller creature has %d +1/+1 counters, want 3 (1 base + 2 rewards)", got)
	}
	if got := plusOneCounterCount(creatures[game.Player2]); got != 1 {
		t.Fatalf("Player2 creature has %d +1/+1 counters, want 1 (accepted)", got)
	}
	if got := plusOneCounterCount(creatures[game.Player3]); got != 1 {
		t.Fatalf("Player3 creature has %d +1/+1 counters, want 1 (accepted)", got)
	}
	if got := plusOneCounterCount(creatures[game.Player4]); got != 0 {
		t.Fatalf("Player4 creature has %d +1/+1 counters, want 0 (declined)", got)
	}
}

// TestTemptingOfferCounterAllDecline proves that when no opponent accepts, only
// the controller's creatures gain the single base counter and no reward is
// applied.
func TestTemptingOfferCounterAllDecline(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	controllerCreature := addCreaturePermanent(g, game.Player1)
	opponentCreatures := map[game.PlayerID]*game.Permanent{
		game.Player2: addCreaturePermanent(g, game.Player2),
		game.Player3: addCreaturePermanent(g, game.Player3),
		game.Player4: addCreaturePermanent(g, game.Player4),
	}
	addInstructionSpellToStack(g, []game.Instruction{temptingOfferCounterInstruction()})

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player4: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := plusOneCounterCount(controllerCreature); got != 1 {
		t.Fatalf("controller creature has %d +1/+1 counters, want 1 (base only)", got)
	}
	for pid, creature := range opponentCreatures {
		if got := plusOneCounterCount(creature); got != 0 {
			t.Fatalf("opponent %v creature has %d +1/+1 counters, want 0 (declined)", pid, got)
		}
	}
}

// TestTemptingOfferReanimationActorOwnership proves the reanimation idiom (Tempt
// with Immortality) reanimates from each acting player's own graveyard under that
// player's control: the controller reanimates one creature for the base plus one
// per accepting opponent, and each accepting opponent reanimates one from their
// own graveyard.
func TestTemptingOfferReanimationActorOwnership(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// The controller resolves the reanimation once for the base plus once per
	// accepting opponent (two here), so its graveyard needs three creature cards.
	for range 3 {
		addCreatureCardToGraveyard(g, game.Player1)
	}
	addCreatureCardToGraveyard(g, game.Player2)
	addCreatureCardToGraveyard(g, game.Player3)
	addCreatureCardToGraveyard(g, game.Player4)
	addInstructionSpellToStack(g, []game.Instruction{temptingOfferReturnInstruction()})

	// Player2 and Player3 accept and each choose their sole graveyard creature;
	// Player4 declines. The controller chooses a creature for the base and each
	// reward. choiceOnlyAgent pops choices in order, so a leading accept choice is
	// followed by graveyard-card choices as they arise.
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}, {0}, {0}}},
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}, {0}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{1}, {0}}},
		game.Player4: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	counts := creaturesByController(g)
	// Controller reanimates the base creature plus one per accepting opponent (2).
	if counts[game.Player1] != 3 {
		t.Fatalf("controller controls %d battlefield creatures, want 3 (1 base + 2 rewards)", counts[game.Player1])
	}
	if counts[game.Player2] != 1 {
		t.Fatalf("Player2 controls %d battlefield creatures, want 1 (accepted)", counts[game.Player2])
	}
	if counts[game.Player3] != 1 {
		t.Fatalf("Player3 controls %d battlefield creatures, want 1 (accepted)", counts[game.Player3])
	}
	if counts[game.Player4] != 0 {
		t.Fatalf("Player4 controls %d battlefield creatures, want 0 (declined)", counts[game.Player4])
	}
	// The reanimated creatures leave the graveyards; a decliner keeps theirs.
	if g.Players[game.Player4].Graveyard.Size() != 1 {
		t.Fatalf("Player4 graveyard has %d cards, want 1 (declined, unchanged)", g.Players[game.Player4].Graveyard.Size())
	}
}

// TestTemptingOfferCopyActorControlAndTarget proves the copy idiom (Tempt with
// Reflections) copies the controller's single target creature for every
// resolution and enters each copy under the acting player's control: the
// controller makes the base copy plus one per accepting opponent, and each
// accepting opponent makes one copy of its own.
func TestTemptingOfferCopyActorControlAndTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player1)
	addInstructionSpellToStackForController(
		g,
		game.Player1,
		[]game.Instruction{temptingOfferCopyInstruction()},
		[]game.Target{game.PermanentTarget(target.ObjectID)},
	)

	// Player2 accepts; Player3 and Player4 decline.
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player4: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	counts := tokensByController(g)
	// Controller makes the base copy plus one reward for the sole accepter (2).
	if counts[game.Player1] != 2 {
		t.Fatalf("controller controls %d copy tokens, want 2 (1 base + 1 reward)", counts[game.Player1])
	}
	if counts[game.Player2] != 1 {
		t.Fatalf("Player2 controls %d copy tokens, want 1 (accepted)", counts[game.Player2])
	}
	if counts[game.Player3] != 0 || counts[game.Player4] != 0 {
		t.Fatalf("declining opponents control copy tokens: Player3=%d Player4=%d, want 0", counts[game.Player3], counts[game.Player4])
	}
}

// TestTemptingOfferPublishesAcceptedActors proves the Tempting offer publishes the
// exact set of accepting opponents (count and identity), the generic
// accepted-member result a future per-accepter consequence reads.
func TestTemptingOfferPublishesAcceptedActors(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for _, pid := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		addCreaturePermanent(g, pid)
	}
	const resultKey = game.ResultKey("tempting-accepters")
	instr := temptingOfferCounterInstruction()
	instr.PublishResult = resultKey
	obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}

	// Player2 and Player4 accept; Player3 declines.
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player4: &choiceOnlyAgent{choices: [][]int{{1}}},
	}

	engine.resolveInstructionWithChoices(g, obj, &instr, agents, &TurnLog{})

	result, ok := obj.ResolutionResults[string(resultKey)]
	if !ok {
		t.Fatal("no result published under the accepted-actors key")
	}
	if !result.Accepted {
		t.Fatal("result Accepted = false, want true (two opponents accepted)")
	}
	if result.Amount != 2 {
		t.Fatalf("result Amount = %d, want 2 accepters", result.Amount)
	}
	if got := result.AcceptedActors.Count(); got != 2 {
		t.Fatalf("AcceptedActors count = %d, want 2", got)
	}
	if !result.AcceptedActors.Contains(game.Player2) || !result.AcceptedActors.Contains(game.Player4) {
		t.Fatalf("AcceptedActors = %v, want Player2 and Player4", result.AcceptedActors.Members())
	}
	if result.AcceptedActors.Contains(game.Player1) || result.AcceptedActors.Contains(game.Player3) {
		t.Fatalf("AcceptedActors = %v, want to exclude the controller and the decliner", result.AcceptedActors.Members())
	}
}
