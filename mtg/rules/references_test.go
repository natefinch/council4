package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestRecipientReferenceUsesDestroyedTargetControllerLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{
		Name:  "Borrowed Permanent",
		Types: []types.Card{types.Artifact},
	})
	target.Controller = game.Player3
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	token := &game.CardDef{
		Name:      "Beast",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	}
	log := TurnLog{}

	engine.resolveEffect(g, obj, game.Effect{Type: game.EffectDestroy, TargetIndex: 0}, &log)
	engine.resolveEffect(g, obj, game.Effect{
		Type:   game.EffectCreateToken,
		Amount: 1,
		Token:  opt.Val(token),
		Recipient: opt.Val(game.PlayerReference{
			Kind: game.PlayerReferenceObjectController,
			Object: opt.Val(game.ObjectReference{
				Kind:        game.ObjectReferenceTargetPermanent,
				TargetIndex: 0,
			}),
		}),
	}, &log)

	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("target permanent remained on battlefield")
	}
	if got := countControlledTokensNamed(g, game.Player3, types.Beast); got != 1 {
		t.Fatalf("Player3 Beast tokens = %d, want 1", got)
	}
	if got := countControlledTokensNamed(g, game.Player1, types.Beast); got != 0 {
		t.Fatalf("spell controller Beast tokens = %d, want 0", got)
	}
	if got := countControlledTokensNamed(g, game.Player2, types.Beast); got != 0 {
		t.Fatalf("target owner Beast tokens = %d, want 0", got)
	}
}

func TestDamageSourceReferenceAppliesCreatureDamageKeywords(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:      "Venomous Healer",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		Abilities: []game.AbilityDef{{
			Kind:     game.StaticAbility,
			Keywords: []game.Keyword{game.Deathtouch, game.Lifelink},
		}},
	})
	target := addCombatPermanent(g, game.Player2, &game.CardDef{
		Name:      "Large Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5}),
	})
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets: []game.Target{
			game.PermanentTarget(source.ObjectID),
			game.PermanentTarget(target.ObjectID),
		},
	}
	log := TurnLog{}

	engine.resolveEffect(g, obj, game.Effect{
		Type:        game.EffectDamage,
		TargetIndex: 1,
		DamageSource: opt.Val(game.ObjectReference{
			Kind:        game.ObjectReferenceTargetPermanent,
			TargetIndex: 0,
		}),
		DynamicAmount: opt.Val(game.DynamicAmount{
			Kind:        game.DynamicAmountTargetPower,
			TargetIndex: 0,
		}),
	}, &log)

	if got := target.MarkedDamage; got != 2 {
		t.Fatalf("marked damage = %d, want 2", got)
	}
	if !target.MarkedDeathtouchDamage {
		t.Fatal("target was not marked with deathtouch damage")
	}
	if got := g.Players[game.Player1].Life; got != 42 {
		t.Fatalf("Player1 life = %d, want 42 from lifelink", got)
	}
}

func TestLegacyTokenCreationStillUsesSpellController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := &game.CardDef{Name: "Legacy Token", Types: []types.Card{types.Creature}}
	obj := &game.StackObject{Controller: game.Player1}
	log := TurnLog{}

	engine.resolveEffect(g, obj, game.Effect{
		Type:   game.EffectCreateToken,
		Amount: 1,
		Token:  opt.Val(token),
	}, &log)

	if got := countControlledTokensNamed(g, game.Player1, "Legacy Token"); got != 1 {
		t.Fatalf("Player1 legacy tokens = %d, want 1", got)
	}
}

func countControlledTokensNamed(g *game.Game, controller game.PlayerID, name types.Sub) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if !permanent.Token || permanent.Controller != controller || permanent.TokenDef == nil || permanent.TokenDef.Name != string(name) {
			continue
		}
		count++
	}
	return count
}
