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
	target := addCombatPermanent(g, game.Player2, &game.CardDef{
		Name:  "Warped Creature",
		Types: []types.Card{types.Creature},
	})
	target.Controller = game.Player3
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(g, game.Player1, &game.CardDef{Name: "Chaos-Like Spell"}),
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	log := TurnLog{}

	for _, effect := range chaosWarpLikeEffects() {
		engine.resolveEffect(g, obj, &effect, &log)
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
	instantID := addCardToLibrary(g, game.Player2, &game.CardDef{
		Name:  "Not Permanent",
		Types: []types.Card{types.Instant},
	})
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(g, game.Player1, &game.CardDef{Name: "Reveal-Like Spell"}),
		Controller: game.Player1,
		Targets:    []game.Target{game.PlayerTarget(game.Player2)},
	}
	log := TurnLog{}

	engine.resolveEffect(g, obj, &game.Effect{
		Type:        game.EffectReveal,
		Amount:      1,
		TargetIndex: 0,
		LinkID:      "revealed",
	}, &log)
	engine.resolveEffect(g, obj, &game.Effect{
		Type:   game.EffectPutOnBattlefield,
		LinkID: "revealed",
		CardCondition: opt.Val(game.CardCondition{
			Card: game.CardReference{
				Kind:   game.CardReferenceLinked,
				LinkID: "revealed",
			},
			RequirePermanentCard: true,
		}),
	}, &log)

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

func chaosWarpLikeEffects() []game.Effect {
	ownerOfTarget := opt.Val(game.PlayerReference{
		Kind: game.PlayerReferenceObjectOwner,
		Object: opt.Val(game.ObjectReference{
			Kind:        game.ObjectReferenceTargetPermanent,
			TargetIndex: 0,
		}),
	})
	return []game.Effect{
		{Type: game.EffectShufflePermanentIntoLibrary, TargetIndex: 0},
		{
			Type:        game.EffectReveal,
			Amount:      1,
			TargetIndex: 0,
			LinkID:      "revealed",
			Recipient:   ownerOfTarget,
		},
		{
			Type:        game.EffectPutOnBattlefield,
			TargetIndex: 0,
			LinkID:      "revealed",
			Recipient:   ownerOfTarget,
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
