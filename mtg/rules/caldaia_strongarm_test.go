package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: Oracle ID e0e6938b-7abf-4f69-a3d6-82a58912bff4, pinned
// printing 58b431d2-33c1-49d8-b45d-690c174bb456, from corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4:
// {4}{G}, 2/3 Human Warrior; "When this creature enters, put two +1/+1
// counters on target creature. Blitz {3}{G} ..." Scryfall card and rulings URI
// https://api.scryfall.com/cards/58b431d2-33c1-49d8-b45d-690c174bb456,
// retrieved 2026-11-19; five WotC rulings dated 2022-04-29 confirm ordinary
// cast timing/countering, optional attack, draw on any death, sacrifice only
// while the original permanent remains, and no Blitz consequences on copies.
// Official CR effective 2026-09-25,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
// (verified local copy SHA-256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca);
// read 118.9, 601.2b/f-h, 122.1a, 302.6, 603.2-3/7a, 513.1, 702.152a:
// Blitz is an announced alternative cost; the entry trigger targets when put
// on the stack and adds two counters, while Blitz grants haste and a death draw
// and schedules sacrifice at the beginning of the next end step.
func TestCaldaiaStrongarmBlitzCountersTargetCreature(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(authored.CaldaiaStrongarm.OracleID)
	if err != nil || def.Name != "Caldaia Strongarm" {
		t.Fatalf("checked Caldaia Strongarm identity lookup = %v, %v", def, err)
	}

	engine := NewEngineWithPrograms(nil, registry.Programs())
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	target := addCombatPermanent(g, game.Player1, vanillaCreature("Counter target", 2, 2))
	card := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: caldaiaStrongarmAgent{target: game.PermanentTarget(target.ObjectID)},
	}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(card, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("could not cast Caldaia Strongarm for its Blitz cost")
	}
	spell, ok := g.Stack.Peek()
	if !ok || !spell.Blitzed || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("Caldaia Strongarm did not pay {3}{G} for Blitz: spell=%+v mana=%d", spell, g.Players[game.Player1].ManaPool.Total())
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if permanentForCard(g, card) == nil {
		t.Fatal("Caldaia Strongarm did not enter the battlefield")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Caldaia Strongarm's entry trigger was not put on the stack")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || trigger.Kind != game.StackTriggeredAbility || len(trigger.Targets) != 1 ||
		trigger.Targets[0] != game.PermanentTarget(target.ObjectID) {
		t.Fatalf("Caldaia Strongarm's entry trigger targets = %+v, want the chosen creature", trigger)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("Caldaia Strongarm put %d +1/+1 counters on target, want 2", got)
	}
}

type caldaiaStrongarmAgent struct {
	target game.Target
}

func (caldaiaStrongarmAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a caldaiaStrongarmAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoicePayment {
		for _, option := range request.Options {
			if option.Label == "Blitz" {
				return []int{option.Index}
			}
		}
	}
	if request.Kind == game.ChoiceTarget {
		for _, option := range request.Options {
			for _, target := range option.Targets {
				if target == a.target {
					return []int{option.Index}
				}
			}
		}
	}
	return request.DefaultSelection
}
