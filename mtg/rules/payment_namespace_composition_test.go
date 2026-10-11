package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompiledPaymentNamespacesKeepPayerAndCounterOutcomesDistinct(t *testing.T) {
	for _, tax := range []int{0, 3} {
		body := fmt.Sprintf("Counter target spell unless its controller pays {%d}. If they do, you gain 2 life. If that spell is countered this way, you gain 3 life. You gain 1 life.", tax)
		def := compileCounterTaxCard(t, "Payment Actor Probe", body, "Instant")
		for _, tc := range []struct {
			name              string
			paid, protected   bool
			unpayable         bool
			wantPaid, counter bool
		}{
			{name: "paid", paid: true, wantPaid: true},
			{name: "declined", counter: true},
			{name: "protected declined", protected: true},
			{name: "protected paid", paid: true, protected: true, wantPaid: true},
			{name: "unpayable", unpayable: true, counter: true},
		} {
			if tax == 0 && tc.unpayable {
				continue
			}
			t.Run(fmt.Sprintf("%d/%s", tax, tc.name), func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				target := addStackSpell(g, game.Player2, "Target", []types.Card{types.Sorcery})
				g.CardInstances[target.SourceID].Owner = game.Player4
				if tc.protected {
					g.CardInstances[target.SourceID].Def.StaticAbilities = []game.StaticAbility{game.CantBeCounteredStaticBody}
				}
				g.Players[game.Player1].ManaPool.Add(mana.C, 8)
				if !tc.unpayable {
					g.Players[game.Player2].ManaPool.Add(mana.C, tax)
				}
				obj := pushCounterTaxSpell(g, def, target)
				outer := game.InstructionResolutionResult{Amount: 97, Succeeded: true}
				obj.ResolutionResults = map[string]game.InstructionResolutionResult{
					"payment-clause-1": outer, "result-clause-1": outer,
				}
				agents := [game.NumPlayers]PlayerAgent{}
				if !tc.paid {
					agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{{0}}}
				}
				NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				wantLife := 41
				if tc.wantPaid {
					wantLife += 2
				}
				if tc.counter {
					wantLife += 3
				}
				if g.Players[game.Player1].Life != wantLife {
					t.Fatalf("payer and counter outcomes conflated: life=%d want=%d", g.Players[game.Player1].Life, wantLife)
				}
				if _, remains := stackObjectByID(g, target.ID); remains == tc.counter {
					t.Fatal("counter result disagrees with actual disposition")
				}
				if len(obj.ResolutionResults) != 2 || obj.ResolutionResults["payment-clause-1"] != outer ||
					obj.ResolutionResults["result-clause-1"] != outer {
					t.Fatalf("local receipts leaked or overwrote outer frame: %+v", obj.ResolutionResults)
				}
				if g.Players[game.Player1].ManaPool.Total() != 8 || g.Players[game.Player4].Life != 40 ||
					g.Players[game.Player4].Graveyard.Contains(target.SourceID) != tc.counter {
					t.Fatal("payer/controller/physical-owner isolation changed")
				}
			})
		}
	}
}

func TestCompiledMultiplePaymentNamespacesRepeatAndModeIsolation(t *testing.T) {
	body := "Counter target spell unless its controller pays {3}. You gain 1 life. If they do, you gain 2 life. If that spell is countered this way, you gain 3 life. " +
		"Counter target spell unless its controller pays {2}. You gain 1 life. If they do, you gain 4 life. If that spell is countered this way, you gain 5 life."
	for _, modal := range []bool{false, true} {
		text := body
		if modal {
			text = "Choose one —\n• " + body + "\n• " + body
		}

		def := compileCounterTaxCard(t, "Multiple Payment Probe", text, "Instant")
		for _, firstPaid := range []bool{false, true} {
			for _, secondPaid := range []bool{false, true} {
				t.Run(fmt.Sprintf("modal=%v/paid=%v,%v", modal, firstPaid, secondPaid), func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					for invocation := range 2 {
						before := g.Players[game.Player1].Life
						first := addStackSpell(g, game.Player2, "First", []types.Card{types.Sorcery})
						second := addStackSpell(g, game.Player3, "Second", []types.Card{types.Sorcery})
						if firstPaid {
							g.Players[game.Player2].ManaPool.Add(mana.C, 3)
						}
						if secondPaid {
							g.Players[game.Player3].ManaPool.Add(mana.C, 2)
						}
						obj := pushCounterTaxSpell(g, def, first)
						obj.Targets = []game.Target{game.StackObjectTarget(first.ID), game.StackObjectTarget(second.ID)}
						obj.TargetCounts = []int{1, 1}
						if modal {
							obj.ChosenModes = []int{invocation}
						}
						NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
						want := 2 + 3 + 5
						if firstPaid {
							want--
						}
						if secondPaid {
							want--
						}
						if g.Players[game.Player1].Life != before+want {
							t.Fatalf("independent receipts reused: life=%d want=%d", g.Players[game.Player1].Life, before+want)
						}
						if len(obj.ResolutionResults) != 0 {
							t.Fatalf("receipt survived invocation/mode: %+v", obj.ResolutionResults)
						}
					}

				})
			}
		}
	}
}

func TestCompiledMultiplePaymentDestinationOwnership(t *testing.T) {
	def := compileCounterTaxCard(t, "Multiple Payment Destination",
		"Counter target spell unless its controller pays {3}. Counter target spell unless its controller pays {4}. "+counterOutcomeExile, "Instant")
	for _, tc := range []struct {
		name, firstDestination, secondDestination string
		sameTarget, firstPaid, secondPaid         bool
	}{
		{name: "independent declines", firstDestination: "graveyard", secondDestination: "exile"},
		{name: "first pays", firstPaid: true, firstDestination: "stack", secondDestination: "exile"},
		{name: "second pays", secondPaid: true, firstDestination: "graveyard", secondDestination: "stack"},
		{name: "same target gone", sameTarget: true, firstDestination: "graveyard", secondDestination: "graveyard"},
		{name: "same target first pays", sameTarget: true, firstPaid: true, firstDestination: "exile", secondDestination: "exile"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			first := addStackSpell(g, game.Player2, "First", []types.Card{types.Sorcery})
			second := first
			if !tc.sameTarget {
				second = addStackSpell(g, game.Player3, "Second", []types.Card{types.Sorcery})
			}
			g.CardInstances[first.SourceID].Owner = game.Player4
			g.CardInstances[second.SourceID].Owner = game.Player4
			if tc.firstPaid {
				g.Players[first.Controller].ManaPool.Add(mana.C, 3)
			}
			if tc.secondPaid {
				g.Players[second.Controller].ManaPool.Add(mana.C, 4)
			}
			obj := pushCounterTaxSpell(g, def, first)
			obj.Targets = []game.Target{game.StackObjectTarget(first.ID), game.StackObjectTarget(second.ID)}
			obj.TargetCounts = []int{1, 1}
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			for _, target := range []struct {
				object      *game.StackObject
				destination string
			}{{first, tc.firstDestination}, {second, tc.secondDestination}} {
				_, onStack := stackObjectByID(g, target.object.ID)
				if onStack != (target.destination == "stack") ||
					g.Players[game.Player4].Graveyard.Contains(target.object.SourceID) != (target.destination == "graveyard") ||
					g.Players[game.Player4].Exile.Contains(target.object.SourceID) != (target.destination == "exile") {
					t.Fatalf("counter destination stole another payment's target: expected %s", target.destination)
				}
			}
			if len(obj.ResolutionResults) != 0 {
				t.Fatal("payment results survived resolution")
			}
		})
	}
}

func TestCompiledPaymentSelectedModesDoNotSharePayerSuccess(t *testing.T) {
	body := "Counter target spell unless its controller pays {3}. You gain 1 life. If they do, you gain 2 life."
	def := compileCounterTaxCard(t, "Payment Selected Modes", "Choose two —\n• "+body+"\n• "+body, "Instant")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addStackSpell(g, game.Player2, "Shared Target", []types.Card{types.Sorcery})
	g.Players[game.Player2].ManaPool.Add(mana.C, 3)
	obj := pushCounterTaxSpell(g, def, target)
	obj.ChosenModes = []int{0, 1}
	obj.Targets = []game.Target{game.StackObjectTarget(target.ID), game.StackObjectTarget(target.ID)}
	obj.TargetCounts = []int{1, 1}
	outer := game.InstructionResolutionResult{Accepted: true, Succeeded: true, Amount: 99}
	obj.ResolutionResults = map[string]game.InstructionResolutionResult{"payment-clause-1": outer}
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != 44 {
		t.Fatalf("selected mode reused a different payer success: life=%d want=44", g.Players[game.Player1].Life)
	}
	if _, remains := stackObjectByID(g, target.ID); remains {
		t.Fatal("second mode did not counter after its own unpayable tax")
	}
	if len(obj.ResolutionResults) != 1 || obj.ResolutionResults["payment-clause-1"] != outer {
		t.Fatal("selected modes leaked their local payment results")
	}
}
