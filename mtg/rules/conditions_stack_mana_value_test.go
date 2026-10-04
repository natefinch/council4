package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

func stackSelectionCondition(targetIndex int, selection game.Selection) game.EffectCondition {
	ref := game.TargetStackObjectReference(targetIndex)
	return game.EffectCondition{
		Object: ref,
		Condition: opt.Val(game.Condition{
			Object:        opt.Val(ref),
			ObjectMatches: opt.Val(selection),
		}),
	}
}

func resolveStackSelectionCounter(t *testing.T, g *game.Game, target *game.StackObject,
	targetIndex int, selection game.Selection, wantCountered bool,
) {
	t.Helper()
	targets := []game.Target{game.StackObjectTarget(target.ID)}
	var decoy *game.StackObject
	if targetIndex == 1 {
		decoy = spellWithManaValue(g, game.Player2, cost.Mana{cost.O(9)})
		targets = append([]game.Target{game.StackObjectTarget(decoy.ID)}, targets...)
	}
	condition := stackSelectionCondition(targetIndex, selection)
	gate := game.Instruction{
		Primitive: game.CounterObject{Object: game.TargetStackObjectReference(targetIndex)},
		Condition: opt.Val(condition),
	}
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{gate}, targets)
	resolving, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("gated spell missing from stack")
	}
	ctx := conditionContext{controller: game.Player1, obj: resolving}
	if got := conditionObjectMatches(g, ctx, &condition.Condition.Val); got != wantCountered {
		t.Errorf("conditionObjectMatches = %v, want %v", got, wantCountered)
	}

	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})

	_, remains := stackObjectByID(g, target.ID)
	if countered := !remains; countered != wantCountered {
		t.Errorf("countered = %v, want %v", countered, wantCountered)
	}
	if wantCountered && target.Kind == game.StackSpell && !target.Copy && !g.Players[game.Player2].Graveyard.Contains(target.SourceID) {
		t.Error("countered physical spell did not move to graveyard")
	}
	if decoy != nil {
		if _, remains := stackObjectByID(g, decoy.ID); !remains {
			t.Error("target-zero decoy was countered")
		}
	}
}

func TestStackConditionManaValueResolution(t *testing.T) {
	tests := []struct {
		name        string
		manaValue   int
		comparison  compare.Int
		colors      []color.Color
		targetIndex int
		want        bool
	}{
		{name: "below threshold", manaValue: 2, comparison: compare.Int{Op: compare.GreaterOrEqual, Value: 3}},
		{name: "at threshold", manaValue: 3, comparison: compare.Int{Op: compare.GreaterOrEqual, Value: 3}, want: true},
		{name: "above threshold", manaValue: 4, comparison: compare.Int{Op: compare.GreaterOrEqual, Value: 3}, want: true},
		{name: "known zero equality", comparison: compare.Int{Op: compare.Equal, Value: 0}, want: true},
		{name: "zero below positive threshold", comparison: compare.Int{Op: compare.GreaterOrEqual, Value: 1}},
		{name: "upper bound includes boundary", manaValue: 3, comparison: compare.Int{Op: compare.LessOrEqual, Value: 3}, want: true},
		{name: "upper bound rejects higher", manaValue: 4, comparison: compare.Int{Op: compare.LessOrEqual, Value: 3}},
		{name: "numeric and color both match", manaValue: 3, comparison: compare.Int{Op: compare.GreaterOrEqual, Value: 3}, colors: []color.Color{color.Blue}, want: true},
		{name: "color cannot replace numeric match", manaValue: 2, comparison: compare.Int{Op: compare.GreaterOrEqual, Value: 3}, colors: []color.Color{color.Blue}},
		{name: "numeric cannot replace color match", manaValue: 4, comparison: compare.Int{Op: compare.GreaterOrEqual, Value: 3}, colors: []color.Color{color.Red}},
		{name: "nonzero target slot", manaValue: 3, comparison: compare.Int{Op: compare.GreaterOrEqual, Value: 3}, targetIndex: 1, want: true},
		{name: "nonzero slot rejects despite qualifying decoy", manaValue: 2, comparison: compare.Int{Op: compare.GreaterOrEqual, Value: 3}, targetIndex: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(tt.manaValue)})
			g.CardInstances[target.SourceID].Def.Colors = tt.colors
			selection := game.Selection{ManaValue: opt.Val(tt.comparison)}
			if len(tt.colors) != 0 {
				selection.ColorsAny = []color.Color{color.Blue}
			}
			resolveStackSelectionCounter(t, g, target, tt.targetIndex, selection, tt.want)
		})
	}
}

func TestStackConditionManaValueSpellCharacteristics(t *testing.T) {
	tests := []struct {
		name       string
		manaCost   cost.Mana
		x          int
		alternate  bool
		token      bool
		copy       bool
		copyValues bool
		faceDown   bool
		want       int
	}{
		{name: "X is chosen stack value", manaCost: cost.Mana{cost.X, cost.U}, x: 3, want: 4},
		{name: "each X contributes", manaCost: cost.Mana{cost.X, cost.X, cost.U}, x: 3, want: 7},
		{name: "X can be zero", manaCost: cost.Mana{cost.X}, want: 0},
		{name: "absent mana cost is known zero", want: 0},
		{name: "selected alternate face", manaCost: cost.Mana{cost.X, cost.R, cost.R}, x: 3, alternate: true, want: 5},
		{name: "physical spell copy", manaCost: cost.Mana{cost.X, cost.U}, x: 3, copy: true, want: 4},
		{name: "copy values override printed cost", manaCost: cost.Mana{cost.U}, x: 3, copy: true, copyValues: true, want: 7},
		{name: "token backed spell", manaCost: cost.Mana{cost.O(4)}, token: true, want: 4},
		{name: "token backed alternate face", manaCost: cost.Mana{cost.X, cost.R, cost.R}, x: 3, alternate: true, token: true, want: 5},
		{name: "token backed copy", manaCost: cost.Mana{cost.X, cost.X, cost.U}, x: 3, token: true, copy: true, want: 7},
		{name: "face down hides printed cost and X", manaCost: cost.Mana{cost.X, cost.O(8)}, x: 3, faceDown: true, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, match := range []bool{true, false} {
				t.Run(map[bool]string{true: "equal", false: "near miss"}[match], func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					target := spellWithManaValue(g, game.Player2, tt.manaCost)
					def := g.CardInstances[target.SourceID].Def
					if tt.manaCost == nil {
						def.ManaCost = opt.V[cost.Mana]{}
					}
					if tt.alternate {
						def.Alternate = opt.Val(def.CardFace)
						def.ManaCost = opt.Val(cost.Mana{cost.O(1)})
						target.Face = game.FaceAlternate
					}
					if tt.token {
						target.SourceTokenDef = def
						delete(g.CardInstances, target.SourceID)
						target.SourceID = 0
						target.Copy = true
					}
					if tt.copy {
						target.Copy = true
					}
					if tt.copyValues {
						target.CopyValues = opt.Val(game.CopyableValues{
							ManaCost: opt.Val(cost.Mana{cost.X, cost.X, cost.U}),
						})
					}
					target.XValue = tt.x
					target.FaceDown = tt.faceDown
					value := tt.want
					if !match {
						value++
					}
					resolveStackSelectionCounter(t, g, target, 0, game.Selection{
						ManaValue: opt.Val(compare.Int{Op: compare.Equal, Value: value}),
					}, match)
				})
			}
		})
	}
}

func TestStackConditionManaValueUnavailableFailsClosed(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*game.Game, *game.StackObject, *game.StackObject)
	}{
		{name: "missing definition", setup: func(g *game.Game, target, _ *game.StackObject) {
			delete(g.CardInstances, target.SourceID)
		}},
		{name: "missing selected face", setup: func(_ *game.Game, target, _ *game.StackObject) {
			target.Face = game.FaceAlternate
		}},
		{name: "activated ability", setup: func(_ *game.Game, target, _ *game.StackObject) {
			target.Kind = game.StackActivatedAbility
			target.SourceCardID = target.SourceID
		}},
		{name: "triggered ability", setup: func(_ *game.Game, target, _ *game.StackObject) {
			target.Kind = game.StackTriggeredAbility
			target.SourceCardID = target.SourceID
		}},
		{name: "unknown stack ID", setup: func(g *game.Game, _ *game.StackObject, resolving *game.StackObject) {
			resolving.Targets[0] = game.StackObjectTarget(g.IDGen.Next())
		}},
		{name: "zero stack ID", setup: func(_ *game.Game, _ *game.StackObject, resolving *game.StackObject) {
			resolving.Targets[0] = game.StackObjectTarget(0)
		}},
		{name: "missing target slot", setup: func(_ *game.Game, _ *game.StackObject, resolving *game.StackObject) {
			resolving.Targets = nil
		}},
		{name: "removed target", setup: func(g *game.Game, target, _ *game.StackObject) {
			g.Stack.RemoveByID(target.ID)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, value := range []int{0, 3} {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				target := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(3)})
				condition := stackSelectionCondition(0, game.Selection{
					ManaValue: opt.Val(compare.Int{Op: compare.Equal, Value: value}),
				})
				gate := game.Instruction{
					Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
					Condition: opt.Val(condition),
				}
				addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{gate},
					[]game.Target{game.StackObjectTarget(target.ID)})
				resolving, ok := g.Stack.Peek()
				if !ok {
					t.Fatal("gated spell missing from stack")
				}
				tt.setup(g, target, resolving)
				if conditionObjectMatches(g, conditionContext{controller: game.Player1, obj: resolving}, &condition.Condition.Val) {
					t.Errorf("unknown/nonspell target matched mana value %d", value)
				}
				before := g.Players[game.Player1].Life
				NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
				if got := g.Players[game.Player1].Life; got != before {
					t.Errorf("unavailable mana value %d allowed gated life gain: got %d, want %d", value, got, before)
				}
			}
		})
	}
}

func TestStackConditionManaValueDoesNotUseCounteredTargetLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(3)})
	condition := stackSelectionCondition(0, game.Selection{
		ManaValue: opt.Val(compare.Int{Op: compare.Equal, Value: 3}),
	})
	instructions := []game.Instruction{
		{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}},
		{
			Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			Condition: opt.Val(condition),
		},
	}
	addInstructionSpellToStackForController(g, game.Player1, instructions,
		[]game.Target{game.StackObjectTarget(target.ID)})
	resolving, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("gated spell missing from stack")
	}
	before := g.Players[game.Player1].Life
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if _, remains := stackObjectByID(g, target.ID); remains {
		t.Fatal("target was not countered")
	}
	if resolving.TargetManaValueLKI[0] != 3 {
		t.Fatal("target mana value LKI was not captured")
	}
	if conditionObjectMatches(g, conditionContext{controller: game.Player1, obj: resolving}, &condition.Condition.Val) {
		t.Error("numeric condition matched removed target via LKI")
	}
	if got := g.Players[game.Player1].Life; got != before {
		t.Errorf("removed target allowed gated life gain: got %d, want %d", got, before)
	}
	dynamic := game.DynamicAmount{
		Kind:   game.DynamicAmountObjectManaValue,
		Object: game.TargetStackObjectReference(0),
	}
	if got := dynamicObjectManaValue(g, resolving, &dynamic); got != 3 {
		t.Errorf("dynamic amount lost countered-target LKI: got %d, want 3", got)
	}
}

func TestStackConditionColorOnlyWithoutSpellManaValue(t *testing.T) {
	for _, kind := range []game.StackObjectKind{game.StackActivatedAbility, game.StackTriggeredAbility} {
		for _, blue := range []bool{true, false} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(3)})
			target.Kind = kind
			target.SourceCardID = target.SourceID
			if blue {
				g.CardInstances[target.SourceID].Def.Colors = []color.Color{color.Blue}
			}
			resolveStackSelectionCounter(t, g, target, 0, game.Selection{
				ColorsAny: []color.Color{color.Blue},
			}, blue)
		}
	}
}
