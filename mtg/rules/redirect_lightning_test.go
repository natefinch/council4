package rules

import (
	"testing"

	cardsr "github.com/natefinch/council4/mtg/cards/r"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// pushRedirectLightning casts the curated Redirect Lightning at a stack object,
// pushing it on top of the stack with the given target. It resolves the real
// lowered spell ability (a single ChooseNewTargets over the targeted stack
// object), so the test exercises the shipped card rather than a hand-built
// effect.
func pushRedirectLightning(g *game.Game, controller game.PlayerID, targetStackObjectID id.ID) {
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID:    sourceID,
		Def:   cardsr.RedirectLightning(),
		Owner: controller,
	}
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     sourceID,
		Controller:   controller,
		Targets:      []game.Target{game.StackObjectTarget(targetStackObjectID)},
		TargetCounts: []int{1},
	})
}

// TestRedirectLightningRetargetsSingleTargetSpell proves the curated Redirect
// Lightning resolves its change-target effect against a single-target spell on
// the stack: the victim spell is retargeted from creature A to creature B while
// it is still on the stack.
func TestRedirectLightningRetargetsSingleTargetSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	creatureA := addCreaturePermanent(g, game.Player1)
	creatureB := addCreaturePermanent(g, game.Player1)

	victimID := g.IDGen.Next()
	g.CardInstances[victimID] = &game.CardInstance{
		ID: victimID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Single Target Bolt",
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				}},
			}.Ability())},
		},
		Owner: game.Player2,
	}
	victimObj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     victimID,
		Controller:   game.Player2,
		Targets:      []game.Target{game.PermanentTarget(creatureA.ObjectID)},
		TargetCounts: []int{1},
	}
	g.Stack.Push(victimObj)

	pushRedirectLightning(g, game.Player1, victimObj.ID)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if len(victimObj.Targets) != 1 {
		t.Fatalf("victim has %d targets after retarget, want 1", len(victimObj.Targets))
	}
	if got := victimObj.Targets[0].PermanentID; got != creatureB.ObjectID {
		t.Fatalf("victim retargeted to %v, want creature B %v", got, creatureB.ObjectID)
	}
}

// TestRedirectLightningRetargetsSingleTargetAbility proves the same card also
// changes the target of an activated ability on the stack, the second stack
// object form its target predicate admits.
func TestRedirectLightningRetargetsSingleTargetAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	creatureA := addCreaturePermanent(g, game.Player1)
	creatureB := addCreaturePermanent(g, game.Player1)

	source := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Pinging Rig",
		Types: []types.Card{types.Artifact},
	}})
	pingBody := &game.ActivatedAbility{
		Content: game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      game.TargetAllowPermanent,
				Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			}},
		}.Ability(),
	}
	abilityObj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackActivatedAbility,
		SourceID:        source.ObjectID,
		SourceCardID:    source.CardInstanceID,
		Controller:      game.Player2,
		Targets:         []game.Target{game.PermanentTarget(creatureA.ObjectID)},
		TargetCounts:    []int{1},
		InlineActivated: pingBody,
	}
	g.Stack.Push(abilityObj)

	pushRedirectLightning(g, game.Player1, abilityObj.ID)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := abilityObj.Targets[0].PermanentID; got != creatureB.ObjectID {
		t.Fatalf("ability retargeted to %v, want creature B %v", got, creatureB.ObjectID)
	}
}
