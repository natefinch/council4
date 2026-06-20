package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCounterThenTokenUsesTargetControllerLKI(t *testing.T) {
	g, target := counterThenTokenGame(t, false)

	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, target.ID); ok {
		t.Fatal("target spell remained after being countered")
	}
	assertBirdToken(t, g, game.Player2)
}

func TestCounterThenTokenStillCreatesForUncounterableSpell(t *testing.T) {
	g, target := counterThenTokenGame(t, true)

	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, target.ID); !ok {
		t.Fatal("uncounterable target spell left the stack")
	}
	assertBirdToken(t, g, game.Player2)
}

func TestCounterThenTokenDoesNothingWhenOnlyTargetIsIllegal(t *testing.T) {
	g, target := counterThenTokenGame(t, false)
	if _, ok := g.Stack.RemoveByID(target.ID); !ok {
		t.Fatal("target spell missing before resolution")
	}

	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})

	for _, permanent := range g.Battlefield {
		if permanent.Token {
			t.Fatal("token created after the spell was countered by rules")
		}
	}
}

func counterThenTokenGame(t *testing.T, uncounterable bool) (*game.Game, *game.StackObject) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	targetDef := &game.CardDef{CardFace: game.CardFace{
		Name:  "Target Sorcery",
		Types: []types.Card{types.Sorcery},
	}}
	if uncounterable {
		targetDef.StaticAbilities = []game.StaticAbility{game.CantBeCounteredStaticBody}
	}
	targetID := addCardToHand(g, game.Player2, targetDef)
	g.Players[game.Player2].Hand.Remove(targetID)
	target := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   targetID,
		Controller: game.Player2,
	}
	g.Stack.Push(target)

	counterID := addCardToHand(g, game.Player1, counterThenTokenDef())
	g.Players[game.Player1].Hand.Remove(counterID)
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     counterID,
		Controller:   game.Player1,
		Targets:      []game.Target{game.StackObjectTarget(target.ID)},
		TargetCounts: []int{1},
	})
	return g, target
}

func counterThenTokenDef() *game.CardDef {
	bird := &game.CardDef{CardFace: game.CardFace{
		Name:            "Bird",
		Colors:          []color.Color{color.Blue},
		Types:           []types.Card{types.Creature},
		Subtypes:        []types.Sub{types.Bird},
		Power:           opt.Val(game.PT{Value: 2}),
		Toughness:       opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
	}}
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Then Token",
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      game.TargetAllowStackObject,
				Predicate: game.TargetPredicate{
					SpellCardTypesAny: []types.Card{types.Enchantment, types.Instant, types.Sorcery},
					StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
				},
			}},
			Sequence: []game.Instruction{
				{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}},
				{Primitive: game.CreateToken{
					Amount:    game.Fixed(1),
					Source:    game.TokenDef(bird),
					Recipient: opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
				}},
			},
		}.Ability()),
	}}
}

func assertBirdToken(t *testing.T, g *game.Game, controller game.PlayerID) {
	t.Helper()
	var token *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			if token != nil {
				t.Fatal("created more than one token")
			}
			token = permanent
		}
	}
	if token == nil {
		t.Fatal("Bird token not created")
	}
	if token.Controller != controller || token.Owner != controller {
		t.Fatalf("token controller/owner = %v/%v, want %v", token.Controller, token.Owner, controller)
	}
	if token.TokenDef == nil ||
		token.TokenDef.Name != "Bird" ||
		!slices.Equal(token.TokenDef.Colors, []color.Color{color.Blue}) ||
		!slices.Equal(token.TokenDef.Types, []types.Card{types.Creature}) ||
		!slices.Equal(token.TokenDef.Subtypes, []types.Sub{types.Bird}) ||
		!token.TokenDef.Power.Exists || token.TokenDef.Power.Val.Value != 2 ||
		!token.TokenDef.Toughness.Exists || token.TokenDef.Toughness.Val.Value != 2 ||
		len(token.TokenDef.StaticAbilities) != 1 ||
		!game.BodyHasKeyword(&token.TokenDef.StaticAbilities[0], game.Flying) {
		t.Fatalf("token = %#v, want 2/2 blue Bird with flying", token)
	}
}
