package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestChooseNewTargetsEffectRetargetsStackSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	creatureA := addCreaturePermanent(g, game.Player1)
	creatureB := addCreaturePermanent(g, game.Player1)

	victimID := g.IDGen.Next()
	g.CardInstances[victimID] = &game.CardInstance{
		ID: victimID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Targeted Spell",
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

	addEffectSpellToStack(g, game.Player1, game.ChooseNewTargets{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(victimObj.ID)})

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if len(victimObj.Targets) != 1 {
		t.Fatalf("victim spell has %d targets, want 1", len(victimObj.Targets))
	}
	if got := victimObj.Targets[0].PermanentID; got != creatureB.ObjectID {
		t.Fatalf("victim spell retargeted to %v, want creature B %v", got, creatureB.ObjectID)
	}
}

// TestChooseNewTargetsThenLoseLifeReadsLiveTargetManaValue proves the Imp's
// Mischief sequence retargets the victim spell and then makes the controller
// lose life equal to that spell's live mana value, read through
// DynamicAmountObjectManaValue over the still-on-stack target spell.
func TestChooseNewTargetsThenLoseLifeReadsLiveTargetManaValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].Life = 20

	creatureA := addCreaturePermanent(g, game.Player1)
	creatureB := addCreaturePermanent(g, game.Player1)

	victimID := g.IDGen.Next()
	g.CardInstances[victimID] = &game.CardInstance{
		ID: victimID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Targeted Spell",
			Types:    []types.Card{types.Instant},
			ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B, cost.B}),
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

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.ChooseNewTargets{Object: game.TargetStackObjectReference(0)}},
		{Primitive: game.LoseLife{
			Player: game.ControllerReference(),
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:   game.DynamicAmountObjectManaValue,
				Object: game.TargetStackObjectReference(0),
			}),
		}},
	}, []game.Target{game.StackObjectTarget(victimObj.ID)})

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := victimObj.Targets[0].PermanentID; got != creatureB.ObjectID {
		t.Fatalf("victim spell retargeted to %v, want creature B %v", got, creatureB.ObjectID)
	}
	if got := g.Players[game.Player1].Life; got != 16 {
		t.Fatalf("controller life = %d, want 16 (lost 4 = victim spell mana value)", got)
	}
}
