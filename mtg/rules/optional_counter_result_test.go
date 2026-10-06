package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

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
			if tt.creature {
				addCreaturePermanent(g, game.Player1)
			}
			resolving := pushCounterTaxSpell(g, def, target)
			agent := &scopedMayAgent{accept: []bool{tt.accept}}
			agents := [game.NumPlayers]PlayerAgent{}
			agents[game.Player1] = agent
			NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			wantLife := 41
			if tt.succeeds {
				wantLife += 2
			}
			result, available := resolving.ResolutionResults["if-you-do"]
			if g.Players[game.Player1].Life != wantLife || agent.next != tt.choices ||
				available != tt.creature || available && (result.Accepted != tt.accept || result.Succeeded != tt.succeeds) {
				t.Fatalf("life=%d choices=%d result=%#v available=%v", g.Players[game.Player1].Life, agent.next, result, available)
			}
			if _, remains := stackObjectByID(g, target.ID); remains == tt.succeeds {
				t.Fatal("optional answer was mistaken for actual counter success")
			}
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
			if tt.creature {
				addCreaturePermanent(g, game.Player1)
			}
			g.Players[game.Player2].ManaPool.Add(mana.C, tt.mana)
			resolving := pushCounterTaxSpell(g, def, target)
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
			result, available := resolving.ResolutionResults["if-you-do"]
			if g.Players[game.Player1].Life != wantLife || g.Players[game.Player1].Hand.Size() != 0 ||
				agent.next != 2 || g.Players[game.Player2].ManaPool.Total() != tt.mana-tt.spent ||
				available != tt.resultAvailable || available && (!result.Accepted || result.Succeeded != tt.succeeded) {
				t.Fatalf("life=%d choices=%d mana=%d result=%#v available=%v; want life=%d",
					g.Players[game.Player1].Life, agent.next, g.Players[game.Player2].ManaPool.Total(), result, available, wantLife)
			}
			if _, remains := stackObjectByID(g, target.ID); remains == tt.succeeded {
				t.Fatal("independent optional group redirected the actual counter outcome")
			}
		})
	}
}
