package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Loyal Fire Sage (TLE 242, Oracle ID 0c052420-2d8b-4bbf-8197-b8a36e7208c9,
// printing 1126b572-1cac-44aa-8fec-e2c72a9fea47) has Firebending 1 and
// "{5}: Create a 1/1 white Ally creature token." Pinned and live Oracle match.
// https://api.scryfall.com/cards/1126b572-1cac-44aa-8fec-e2c72a9fea47/rulings
// (2025-10-02): the trigger uses the stack; only its mana survives combat step
// boundaries and may be spent after combat damage, but not after leaving combat.
// Verified CR 2026-09-25 (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx)
// and 2026-11-06 (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt):
// 702.189a, 508.3a, 603.1–3 and 605.1b require a normal attack trigger; 106.4,
// 500.5/500.5a and 511.1–3 preserve its mana through end-of-combat priority,
// then empty it. CR 602.2 and 111.3–4 govern the card's separate activation
// and token. These fixtures test the shared mana operation, not the card.
// The retained test on main reproduced one red left at combat exit using the
// only preexisting retention option, PersistUntilEndOfTurn.
func TestCombatRetainedManaThroughCombatAndPayment(t *testing.T) {
	for _, route := range []string{"live expiry", "simulator expiry", "generic activation", "colored spell"} {
		t.Run(route, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			log := &TurnLog{}
			g.Turn.ActivePlayer = game.Player1
			g.Turn.MainPhasesThisTurn = 1
			expires := route == "live expiry" || route == "simulator expiry"
			add := game.AddMana{ManaColor: mana.R, Amount: game.Fixed(1), PersistUntilEndOfCombat: true}
			sequence := []game.Instruction{{Primitive: add}}
			if expires {
				// Identical color and provenance, different lifetimes.
				sequence = append(sequence,
					game.Instruction{Primitive: game.AddMana{ManaColor: mana.R, Amount: game.Fixed(1)}},
					game.Instruction{Primitive: game.AddMana{ManaColor: mana.R, Amount: game.Fixed(1), PersistUntilEndOfTurn: true}},
				)
			}
			def := vanillaCreature("Combat mana test attacker", 2, 2)
			def.TriggeredAbilities = []game.TriggeredAbility{{
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever, Pattern: game.TriggerPattern{Event: game.EventAttackerDeclared, Source: game.TriggerSourceSelf},
				},
				Content: game.Mode{Sequence: sequence}.Ability(),
			}}
			gainLife := game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1), Player: game.ControllerReference(),
			}}}}.Ability()

			var paymentAction func(action.Action) bool
			paymentStep := game.StepEndOfCombat
			switch route {
			case "generic activation":
				paymentStep = game.StepCombatDamage
				def.ActivatedAbilities = []game.ActivatedAbility{{ManaCost: opt.Val(cost.Mana{cost.O(5)}), Content: gainLife}}
				for range 4 {
					addBasicLandPermanent(g, game.Player1, types.Mountain)
				}
			case "colored spell":
				spell := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: "Combat mana test instant", Types: []types.Card{types.Instant},
					ManaCost: opt.Val(cost.Mana{cost.R}), SpellAbility: opt.Val(gainLife),
				}})
				paymentAction = func(a action.Action) bool {
					c, ok := a.CastSpellPayload()
					return ok && c.CardID == spell
				}
			}
			attacker := addCombatPermanent(g, game.Player1, def)
			if route == "generic activation" {
				paymentAction = func(a action.Action) bool {
					activation, ok := a.ActivateAbilityPayload()
					return ok && activation.SourceID == attacker.ObjectID
				}
			}
			attack := attackPlayer2(attacker)
			probe := &combatManaProbe{
				attack: action.DeclareAttackers(attack.Attackers), payment: paymentAction,
				paymentStep: paymentStep, red: make(map[game.Step]int),
			}
			policies := simPassPolicies()
			policies[game.Player1] = probe

			initialLife := g.Players[game.Player1].Life
			if route == "simulator expiry" {
				g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
				g.Combat = &game.CombatState{}
				after, ok := engine.Simulator().ResolveCombatWithAttackers(g, game.Player1, attack, policies)
				if !ok || g.Players[game.Player1].ManaPool.Total() != 0 {
					t.Fatal("Simulator failed to run an isolated combat")
				}
				g = after
			} else {
				engine.runCombatPhase(g, policies, log)
			}
			engine.runMainPhase(g, policies, game.PhasePostcombatMain, log)

			if !probe.triggerOnStack {
				t.Fatal("attack did not put the trigger on the stack before producing mana")
			}
			want := 1
			if expires {
				want = 2
			}
			for _, step := range []game.Step{game.StepDeclareBlockers, game.StepCombatDamage, game.StepEndOfCombat} {
				wantAtStep := want
				if route == "generic activation" && step == game.StepEndOfCombat {
					wantAtStep = 0 // The paid mana must not reappear at this boundary.
				}
				if got, seen := probe.red[step]; !seen || got != wantAtStep {
					t.Fatalf("red at first %v priority = %d (seen %v), want %d", step, got, seen, wantAtStep)
				}
			}
			if got, seen := probe.red[game.StepNone]; !seen || got != want-1 {
				t.Fatalf("red at postcombat main priority = %d (seen %v), want %d", got, seen, want-1)
			}
			if !expires && (!probe.paid || g.Players[game.Player1].Life != initialLife+1) {
				t.Fatal("combat mana did not fund and resolve the action")
			}
			wantSpent := map[string]int{"generic activation": 5, "colored spell": 1}[route]
			if got := g.Players[game.Player1].ManaPool.Spent(); got != wantSpent {
				t.Fatalf("mana spent = %d, want %d; expiry is not spending", got, wantSpent)
			}
		})
	}
}

type combatManaProbe struct {
	attack         action.Action
	payment        func(action.Action) bool
	paymentStep    game.Step
	red            map[game.Step]int
	triggerOnStack bool
	paid           bool
}

func (p *combatManaProbe) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if containsAction(legal, p.attack) {
		return p.attack
	}
	step := obs.Turn.Step
	if obs.Turn.Phase != game.PhaseCombat && obs.Turn.Phase != game.PhasePostcombatMain {
		return action.Pass()
	}
	red := obs.g.Players[obs.Player].ManaPool.Amount(mana.R)
	if _, seen := p.red[step]; !seen {
		p.red[step] = red
	}
	if step == game.StepDeclareAttackers && red == 0 {
		if stack := obs.Stack(); len(stack) == 1 && stack[0].Kind == game.StackTriggeredAbility {
			p.triggerOnStack = true
		}
	}
	if step == p.paymentStep && !p.paid && p.payment != nil {
		for _, candidate := range legal {
			if p.payment(candidate) {
				p.paid = true
				return candidate
			}
		}
	}
	return action.Pass()
}
