package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// counterReceiptProbes observe the spell's "if-you-do" receipt while it resolves.
var counterReceiptProbes = []resultProbe{
	{key: "if-you-do", life: 100},
	{key: "if-you-do", accepted: game.TriTrue, life: 1000},
	{key: "if-you-do", succeeded: game.TriTrue, life: 10_000},
}

func counterReceiptVisibility(available, accepted, succeeded bool) int {
	if !available {
		return 0
	}
	visible := 100
	if accepted {
		visible += 1000
	}
	if succeeded {
		visible += 10_000
	}
	return visible
}

func probedCounterDef(def *game.CardDef) *game.CardDef {
	probed := *def
	probed.SpellAbility = opt.Val(withResultProbes(def.SpellAbility.Val, counterReceiptProbes))
	return &probed
}

// A repeated invocation whose producer is gated off must not see the previous
// invocation's receipt.
func assertCounterReceiptNotReused(t *testing.T, g *game.Game, resolving *game.StackObject, def *game.CardDef, creature *game.Permanent) {
	t.Helper()
	assertResultKeysCleaned(t, resolving, "if-you-do")
	if creature != nil && !sacrificePermanent(g, creature) {
		t.Fatal("could not remove the condition creature")
	}
	before := g.Players[game.Player1].Life
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &scopedMayAgent{}}
	NewEngine(nil).resolveAbilityContentWithChoices(g, resolving, probedCounterDef(def).SpellAbility.Val, agents, &TurnLog{})
	if got := g.Players[game.Player1].Life - before; got != 1 {
		t.Fatalf("repeated invocation gained %d, want only its unconditional 1 life", got)
	}
	assertResultKeysCleaned(t, resolving, "if-you-do")
}

func TestCompiledOptionalCounterActualResult(t *testing.T) {
	def := compileCounterTaxCard(t, "Optional Counter Boundary",
		"If you control a creature, you may counter target spell. If that spell is countered this way, you gain 2 life. You gain 1 life.", "Instant")
	for _, tt := range []struct {
		name                                  string
		accept, protected, creature, succeeds bool
		choices                               int
	}{
		{"accepted", true, false, true, true, 1},
		{"declined", false, false, true, false, 1},
		{"accepted ineffective", true, true, true, false, 1},
		{"declined uncounterable", false, true, true, false, 1},
		{"condition false", true, false, false, false, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Sorcery})
			if tt.protected {
				g.CardInstances[target.SourceID].Def.StaticAbilities = []game.StaticAbility{game.CantBeCounteredStaticBody}
			}
			var creature *game.Permanent
			if tt.creature {
				creature = addCreaturePermanent(g, game.Player1)
			}
			resolving := pushCounterTaxSpell(g, probedCounterDef(def), target)
			agent := &scopedMayAgent{accept: []bool{tt.accept}}
			agents := [game.NumPlayers]PlayerAgent{}
			agents[game.Player1] = agent
			NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			wantLife := 41
			if tt.succeeds {
				wantLife += 2
			}
			// The receipt is available exactly when the creature gate admitted the
			// optional action, distinguishing acceptance from actual success.
			visible := counterReceiptVisibility(tt.creature, tt.accept, tt.succeeds)
			if g.Players[game.Player1].Life != wantLife+visible || agent.next != tt.choices {
				t.Fatalf("life=%d choices=%d, want life %d plus in-resolution receipt %d",
					g.Players[game.Player1].Life, agent.next, wantLife, visible)
			}
			if _, remains := stackObjectByID(g, target.ID); remains == tt.succeeds {
				t.Fatal("optional answer was mistaken for actual counter success")
			}
			assertCounterReceiptNotReused(t, g, resolving, def, creature)
		})
	}
}

func TestCompiledCounterResultAndIndependentOptionalGroups(t *testing.T) {
	def := compileCounterTaxCard(t, "Independent Optional Counter Boundary",
		"You may discard a card and gain 3 life. If you control a creature, counter target spell unless its controller pays {3}. "+
			"If that spell is countered this way, you gain 2 life. You gain 1 life. You may discard a card and gain 5 life.", "Instant")
	for _, tt := range []struct {
		name                       string
		mana, spent                int
		decline, protected         bool
		creature, first, last      bool
		succeeded, resultAvailable bool
	}{
		{"paid first group accepted", 3, 3, false, false, true, true, false, false, false},
		{"declined first group declined", 3, 0, true, false, true, false, true, true, true},
		{"unable both groups accepted", 2, 0, false, false, true, true, true, true, true},
		{"uncounterable both groups accepted", 3, 0, true, true, true, true, true, false, true},
		{"false condition first group accepted", 3, 0, false, false, false, true, false, false, false},
		{"succeeded both groups declined", 3, 0, true, false, true, false, false, true, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Sorcery})
			if tt.protected {
				g.CardInstances[target.SourceID].Def.StaticAbilities = []game.StaticAbility{game.CantBeCounteredStaticBody}
			}
			var creature *game.Permanent
			if tt.creature {
				creature = addCreaturePermanent(g, game.Player1)
			}
			g.Players[game.Player2].ManaPool.Add(mana.C, tt.mana)
			resolving := pushCounterTaxSpell(g, probedCounterDef(def), target)
			agent := &scopedMayAgent{accept: []bool{tt.first, tt.last}}
			agents := [game.NumPlayers]PlayerAgent{}
			agents[game.Player1] = agent
			if tt.decline {
				agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{{0}}}
			}
			NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			wantLife := 41
			if tt.first {
				wantLife += 3
			}
			if tt.last {
				wantLife += 5
			}
			if tt.succeeded {
				wantLife += 2
			}
			// A mandatory counter receipt is always accepted when available; the
			// independent optional groups never publish or alter it.
			visible := counterReceiptVisibility(tt.resultAvailable, true, tt.succeeded)
			if g.Players[game.Player1].Life != wantLife+visible || g.Players[game.Player1].Hand.Size() != 0 ||
				agent.next != 2 || g.Players[game.Player2].ManaPool.Total() != tt.mana-tt.spent {
				t.Fatalf("life=%d choices=%d mana=%d; want life=%d plus in-resolution receipt %d",
					g.Players[game.Player1].Life, agent.next, g.Players[game.Player2].ManaPool.Total(), wantLife, visible)
			}
			if _, remains := stackObjectByID(g, target.ID); remains == tt.succeeded {
				t.Fatal("independent optional group redirected the actual counter outcome")
			}
			assertCounterReceiptNotReused(t, g, resolving, def, creature)
		})
	}
}
