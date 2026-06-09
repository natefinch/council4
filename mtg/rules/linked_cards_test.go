package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestChaosWarpLikeEffectsUseTargetOwnerAndLinkedReveal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Warped Creature",
		Types: []types.Card{types.Creature}},
	})
	target.Controller = game.Player3
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Chaos-Like Spell"}}),
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	log := TurnLog{}

	for _, instr := range chaosWarpLikeInstructions() {
		engine.resolveInstructionWithChoices(g, obj, &instr, [game.NumPlayers]PlayerAgent{}, &log)
	}

	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("original permanent object remained on battlefield")
	}
	if g.Players[game.Player2].Library.Contains(target.CardInstanceID) {
		t.Fatal("linked permanent card remained in library")
	}
	if got := countCardPermanentsControlledBy(g, game.Player2, target.CardInstanceID); got != 1 {
		t.Fatalf("Player2 battlefield copies of warped card = %d, want 1", got)
	}
	if got := countCardPermanentsControlledBy(g, game.Player1, target.CardInstanceID); got != 0 {
		t.Fatalf("spell controller battlefield copies of warped card = %d, want 0", got)
	}
	if got := countCardPermanentsControlledBy(g, game.Player3, target.CardInstanceID); got != 0 {
		t.Fatalf("old controller battlefield copies of warped card = %d, want 0", got)
	}
	if !eventRevealedCard(g, target.CardInstanceID, obj.ID) {
		t.Fatal("reveal event for linked card was not emitted")
	}
}

func TestLinkedNonPermanentCardStaysInLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	instantID := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Not Permanent",
		Types: []types.Card{types.Instant}},
	})
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reveal-Like Spell"}}),
		Controller: game.Player1,
		Targets:    []game.Target{game.PlayerTarget(game.Player2)},
	}
	log := TurnLog{}

	revealInstr := game.Instruction{
		Primitive: game.Reveal{
			Amount:        game.Fixed(1),
			Player:        game.TargetPlayerReference(0),
			PublishLinked: "revealed",
		},
	}
	engine.resolveInstructionWithChoices(g, obj, &revealInstr, [game.NumPlayers]PlayerAgent{}, &log)
	putInstr := game.Instruction{
		Primitive: game.PutOnBattlefield{
			Source: game.LinkedBattlefieldSource("revealed"),
		},
		CardCondition: opt.Val(game.CardCondition{
			Card: game.CardReference{
				Kind:   game.CardReferenceLinked,
				LinkID: "revealed",
			},
			RequirePermanentCard: true,
		}),
	}
	engine.resolveInstructionWithChoices(g, obj, &putInstr, [game.NumPlayers]PlayerAgent{}, &log)

	if !g.Players[game.Player2].Library.Contains(instantID) {
		t.Fatal("nonpermanent linked card left library")
	}
	if got := countCardPermanentsControlledBy(g, game.Player1, instantID); got != 0 {
		t.Fatalf("nonpermanent card permanents = %d, want 0", got)
	}
	if !eventRevealedCard(g, instantID, obj.ID) {
		t.Fatal("reveal event for linked nonpermanent was not emitted")
	}
}

func chaosWarpLikeInstructions() []game.Instruction {
	ownerOfTarget := opt.Val(game.ObjectOwnerReference(game.TargetPermanentReference(0)))
	return []game.Instruction{
		{Primitive: game.ShufflePermanentIntoLibrary{Object: game.TargetPermanentReference(0)}},
		{
			Primitive: game.Reveal{
				Amount:        game.Fixed(1),
				Player:        ownerOfTarget.Val,
				PublishLinked: "revealed",
				Recipient:     ownerOfTarget,
			},
		},
		{
			Primitive: game.PutOnBattlefield{
				Source:    game.LinkedBattlefieldSource("revealed"),
				Recipient: ownerOfTarget,
			},
			CardCondition: opt.Val(game.CardCondition{
				Card:                 game.CardReference{Kind: game.CardReferenceLinked, LinkID: "revealed"},
				RequirePermanentCard: true,
			}),
		},
	}
}

func countCardPermanentsControlledBy(g *game.Game, controller game.PlayerID, cardID id.ID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID && permanent.Controller == controller {
			count++
		}
	}
	return count
}

func eventRevealedCard(g *game.Game, cardID, stackObjectID id.ID) bool {
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed && event.CardID == cardID && event.StackObjectID == stackObjectID {
			return true
		}
	}
	return false
}
