package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestWardPlansTreasurePaymentOrLeavesSourcesAlone(t *testing.T) {
	for _, scenario := range []string{"pay", "decline", "unpayable"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			warded := addCombatPermanent(g, game.Player2, registryCardDef(t, "Tolarian Terror"))
			spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
			g.Turn.PriorityPlayer = game.Player1
			if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0, nil)) ||
				!engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("targeting spell or ward trigger failed")
			}
			count := 2
			if scenario == "unpayable" {
				count = 1
			}
			for range count {
				if _, ok := createTokenPermanent(g, game.Player1, recklessLackeyTreasure(t)); !ok {
					t.Fatal("Treasure creation failed")
				}
			}
			consent, windows := 0, 0
			agent := &costedObservationAgent{t: t, choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
				if req.Kind == game.ChoiceMay {
					consent++
					if req.Prompt != "Pay ward cost?" || countTokenPermanentsNamed(g, "Treasure") != count {
						t.Fatal("ward availability consumed a resource or lost its prompt")
					}
					if scenario == "decline" {
						return []int{0}
					}
				}
				if req.Prompt == "Activate a mana ability while paying?" {
					windows++
					for _, option := range req.Options {
						if !strings.Contains(option.Label, "Treasure") {
							t.Fatalf("ward offered an uncompletable mana choice: %+v", req.Options)
						}
					}
				}
				return req.DefaultSelection
			}}
			log := &TurnLog{}
			engine.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), log)
			if scenario == "pay" {
				if consent != 1 || windows != 2 || countTokenPermanentsNamed(g, "Treasure") != 0 ||
					g.Stack.Size() != 1 || g.Players[game.Player1].ManaPool.Total() != 0 {
					t.Fatalf("ward payment failed: consent=%d windows=%d stack=%d", consent, windows, g.Stack.Size())
				}
			} else if windows != 0 || countTokenPermanentsNamed(g, "Treasure") != count ||
				!g.Players[game.Player1].Graveyard.Contains(spellID) || consent != map[bool]int{true: 1}[scenario == "decline"] {
				t.Fatal("declined/unpayable ward activated a source, offered an impossible payment or spared the spell")
			}
			if scenario == "decline" && (len(log.Choices) != 1 || log.Choices[0].Selected[0] != 0) {
				t.Fatal("ward decline was not retained in the log")
			}
		})
	}
}

func TestWardCountersSpellWhenCostIsNotPaid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warded := addWardPermanent(g, game.Player2, cost.Mana{cost.O(1)})
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0, nil)) {
		t.Fatal("targeting spell cast failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ward trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want ward to counter targeting spell", g.Stack.Size())
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("countered spell did not move to graveyard")
	}
}

func TestWardPaidLeavesSpellOnStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warded := addWardPermanent(g, game.Player2, cost.Mana{cost.G})
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0, nil)) {
		t.Fatal("targeting spell cast failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ward trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !forest.Tapped {
		t.Fatal("ward cost did not tap mana source")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want targeting spell still on stack", g.Stack.Size())
	}
	if g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("spell moved to graveyard after ward was paid")
	}
}

func TestWardDoesNotTriggerForControllerTargetingOwnPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warded := addWardPermanent(g, game.Player2, cost.Mana{cost.O(1)})
	spellID := addCardToHand(g, game.Player2, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player2

	if !engine.applyAction(g, game.Player2, action.CastSpell(spellID, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0, nil)) {
		t.Fatal("self-targeting spell cast failed")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ward triggered for controller's own spell")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want only targeting spell", g.Stack.Size())
	}
}

func TestWardCountersActivatedAbilityWhenCostIsNotPaid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warded := addWardPermanent(g, game.Player2, cost.Mana{cost.O(1)})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Targeting Permanent",
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.Mode{
				Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
			}.Ability(),
		}}},
	})
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0)) {
		t.Fatal("targeting activated ability failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ward trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want ward to counter activated ability", g.Stack.Size())
	}
}

func TestFaceDownDisguiseWardCountersSpellWhenCostIsNotPaid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warded := addFaceDownPermanent(g, game.Player2, disguiseCreature(cost.Mana{cost.W}), game.FaceDownDisguise)
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0, nil)) {
		t.Fatal("targeting spell cast failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("face-down disguise ward trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want disguise ward to counter targeting spell", g.Stack.Size())
	}
}

func addWardPermanent(g *game.Game, controller game.PlayerID, manaCost cost.Mana) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Ward Creature",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{game.WardStaticAbility(manaCost)},
	}})
}
