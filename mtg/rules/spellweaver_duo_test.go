package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Spellweaver Duo Oracle ID 1fd4204f-f627-4919-9d48-79b172f4560e, printing
// revision 87ba22f6-5299-46c3-bd01-9140634e60ea (Battlebond #42), mana cost
// {6}{U}, Creature — Human Wizard 4/4; exact Oracle text:
// "Assist (Another player can pay up to {6} of this spell's cost.)
// When this creature enters, you may return target tapped creature to its
// owner's hand."
// Its rulings URI
// https://api.scryfall.com/cards/87ba22f6-5299-46c3-bd01-9140634e60ea/rulings
// was retrieved 2026-11-21 and returned four WotC rulings dated 2018-06-08:
// targets are chosen before the assisting player; only generic mana can be
// paid by that player; colored costs stay with the caster even when mana can be
// spent as any color; and cost increases/reductions change the amount Assist
// permits. Raw response: .cardwork/source-review/spellweaver-duo-rulings.json.
// Read the run-pinned official Wizards Comprehensive Rules effective
// 2026-09-25: https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt,
// verified local copy .cardwork/source-review/MagicCompRules-20260925.txt
// (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca).
// Read 601.2c/f-h and 702.132a for target-before-assist ordering, helper mana
// abilities, and generic-only payment; 603.2-3 and 603.3d for the self-entry
// trigger and its target announcement; 608.2b and 608.2d for tapped-target
// legality and the optional resolving choice; and 108.3/400.7 for ownership
// and the returned creature becoming a new object in its owner's hand.
const spellweaverDuoTestOracleID = "1fd4204f-f627-4919-9d48-79b172f4560e"

type spellweaverDuoChoiceAgent struct {
	simPassPolicy
	requests []game.ChoiceRequest
}

func (a *spellweaverDuoChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if request.Kind == game.ChoiceMay {
		return []int{1}
	}
	return request.DefaultSelection
}

func TestSpellweaverDuoAssistAndEntersTrigger(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(spellweaverDuoTestOracleID)
	if err != nil {
		t.Fatalf("exact Spellweaver Duo registry lookup failed: %v", err)
	}
	if def.Name != "Spellweaver Duo" {
		t.Fatalf("Oracle ID lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	decoy := addCombatPermanent(g, game.Player2, vanillaCreature("untapped decoy", 2, 2))
	target := addCombatPermanent(g, game.Player2, vanillaCreature("tapped target", 2, 2))
	target.Tapped = true
	target.Owner = game.Player3
	g.CardInstances[target.CardInstanceID].Owner = game.Player3
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player2].ManaPool.Add(mana.C, 6)
	setMainPhasePriority(g, game.Player1)

	engine := NewEngineWithPrograms(nil, registry.Programs())
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, nil)) {
		t.Fatal("Spellweaver Duo failed to cast with its {U} paid by its controller and {6} paid with Assist")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 || g.Players[game.Player2].ManaPool.Total() != 0 {
		t.Fatal("Spellweaver Duo did not charge the assisting player for the generic cost and its controller for {U}")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Spellweaver Duo's enters ability did not trigger")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || trigger.Kind != game.StackTriggeredAbility || trigger.SourceCardID != spell ||
		len(trigger.Targets) != 1 || trigger.Targets[0].PermanentID != target.ObjectID {
		t.Fatalf("Spellweaver Duo did not put its targeted enters ability on the stack: %+v", trigger)
	}

	agent := &spellweaverDuoChoiceAgent{}
	agents := simPassPolicies()
	agents[game.Player1] = agent
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if len(agent.requests) != 1 || agent.requests[0].Kind != game.ChoiceMay ||
		agent.requests[0].Player != game.Player1 {
		t.Fatalf("Spellweaver Duo did not offer its controller the optional return: %+v", agent.requests)
	}
	if !g.Players[game.Player3].Hand.Contains(target.CardInstanceID) ||
		g.Players[game.Player2].Hand.Contains(target.CardInstanceID) {
		t.Fatal("Spellweaver Duo did not return the tapped creature to its owner's hand")
	}
	if _, ok := permanentByObjectID(g, decoy.ObjectID); !ok {
		t.Fatal("Spellweaver Duo returned an untapped creature")
	}
}
