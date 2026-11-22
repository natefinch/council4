package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned-corpus BBD 45 printing 8750abff-ae70-4abe-a51a-b93f71b48ef1
// (Oracle ID 37c1de8d-30d6-4b08-9368-279cf66d3e85), Oracle text:
// "Assist (Another player can pay up to {5} of this spell's cost.)
// When this creature enters, target opponent loses 2 life and you gain 2 life."
// Its Scryfall rulings URI
// (https://api.scryfall.com/cards/8750abff-ae70-4abe-a51a-b93f71b48ef1/rulings)
// returned four WotC rulings dated 2018-06-08: a spell's targets are chosen
// before the helper is invited (this creature has no cast-time target; its
// opponent target is chosen when the entry trigger is put on the stack); Assist
// pays only generic mana, leaving colored costs to the caster even when mana
// can be spent as any color; and the payment limit follows the spell's actual
// generic total cost.
// Official Wizards Comprehensive Rules effective 2026-09-25, read from
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt:
// 115.1-2 (players as targets), 119.3 (life totals), 601.2c/g-h (target choice
// and payment sequence), 603.2-3/3d/6a (triggering, target choice and entry),
// 608.2b-c (target legality and instruction order), and 702.132a (Assist).
// Together they require choosing an opponent target when the trigger is put on
// the stack, placing the entry ability on the stack, resolving its instructions
// in order, and allowing the helper to pay generic mana before the caster pays.
func TestBloodbornScoundrelsAssistAndEnterLifeTrigger(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("37c1de8d-30d6-4b08-9368-279cf66d3e85")
	if err != nil {
		t.Fatal(err)
	}
	if def.Name != "Bloodborn Scoundrels" {
		t.Fatalf("Oracle lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player2].ManaPool.Add(mana.C, 5)
	cardID := addCardToHand(g, game.Player1, def)
	casterLife, helperLife, targetLife := g.Players[game.Player1].Life, g.Players[game.Player2].Life, g.Players[game.Player3].Life

	assistInvited, assistPaid, targetChosen := false, false, false
	agent := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
		switch {
		case request.Prompt == assistPlayerPrompt:
			for _, option := range request.Options {
				for _, target := range option.Targets {
					if target.Kind == game.TargetPlayer && target.PlayerID == game.Player2 {
						assistInvited = true
						return []int{option.Index}
					}
				}
			}
			t.Fatal("Assist did not offer Player 2 as a helper")
		case request.Prompt == assistContributionPrompt:
			if request.NumberRange.Val.Maximum.Val != 5 {
				t.Fatalf("Assist contribution maximum = %d, want 5", request.NumberRange.Val.Maximum.Val)
			}
			assistPaid = true
			return []int{5}
		case request.Kind == game.ChoiceTarget:
			for _, option := range request.Options {
				if len(option.Targets) == 1 && option.Targets[0].Kind == game.TargetPlayer &&
					option.Targets[0].PlayerID == game.Player3 {
					targetChosen = true
					return []int{option.Index}
				}
			}
			t.Fatal("Bloodborn Scoundrels did not offer Player 3 as an opponent target")
		}
		return request.DefaultSelection
	}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1], agents[game.Player2] = agent, agent

	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("Bloodborn Scoundrels cast with Assist failed")
	}
	if !assistInvited || !assistPaid || g.Players[game.Player1].ManaPool.Total() != 0 ||
		g.Players[game.Player2].ManaPool.Total() != 0 {
		t.Fatal("cast did not pay {5}{B} using Player 2's Assist contribution")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	entered := false
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID && permanent.Controller == game.Player1 {
			entered = true
			break
		}
	}
	if !entered {
		t.Fatal("Bloodborn Scoundrels did not enter the battlefield")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Bloodborn Scoundrels' entry ability did not trigger")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || len(trigger.Targets) != 1 || trigger.Targets[0].Kind != game.TargetPlayer ||
		trigger.Targets[0].PlayerID != game.Player3 || !targetChosen {
		t.Fatal("entry ability did not target the chosen opponent")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player3].Life != targetLife-2 || g.Players[game.Player1].Life != casterLife+2 ||
		g.Players[game.Player2].Life != helperLife {
		t.Fatalf("life totals after resolution: caster=%d target=%d helper=%d; want %d, %d, %d",
			g.Players[game.Player1].Life, g.Players[game.Player3].Life, g.Players[game.Player2].Life,
			casterLife+2, targetLife-2, helperLife)
	}
}
