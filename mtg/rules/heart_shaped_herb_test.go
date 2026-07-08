package rules

import (
	"testing"

	cardh "github.com/natefinch/council4/mtg/cards/h"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// herbChoiceAgent scripts the two decisions Heart-Shaped Herb's activated
// ability asks for while resolving: the optional "You may sacrifice a creature."
// (a ChoiceMay) and, when accepted, which creature to sacrifice (a ChoicePayment
// selection). It answers every other request with the request's default.
type herbChoiceAgent struct {
	sacrifice bool   // answer to the optional sacrifice
	sacName   string // name of the creature to sacrifice when several are legal
}

func (*herbChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *herbChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		if a.sacrifice {
			return []int{1}
		}
		return []int{0}
	}
	if a.sacName != "" {
		for _, option := range request.Options {
			if option.Card.Exists && option.Card.Val.Name == a.sacName {
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

// addHerbCreature puts a vanilla 2/2 creature owned and controlled by owner onto
// the battlefield so it can be sacrificed and returned by name.
func addHerbCreature(g *game.Game, owner game.PlayerID, name string) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, owner, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
}

// newHerbGame stages the real Heart-Shaped Herb on Player1's battlefield with
// two colorless lands to pay {2}, and returns the herb permanent so a test can
// drive its activated ability through the real activation path.
func newHerbGame(t *testing.T) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	herb := addCombatPermanent(g, game.Player1, cardh.HeartShapedHerb)
	addBasicLandPermanent(g, game.Player1, types.Island)
	addBasicLandPermanent(g, game.Player1, types.Island)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.ActivePlayer = game.Player1
	return g, engine, herb
}

// assertHerbActivationEnumerated proves the ability is reachable through the real
// driver-facing path: it must appear in legalActions before the test drives it,
// so the end-to-end coverage can never be masked by a hand-built action that the
// engine would never actually offer.
func assertHerbActivationEnumerated(t *testing.T, engine *Engine, g *game.Game, herb *game.Permanent) {
	t.Helper()
	for _, a := range engine.legalActions(g, game.Player1) {
		if a.Kind != action.ActionActivateAbility {
			continue
		}
		if p, ok := a.ActivateAbilityPayload(); ok && p.SourceID == herb.ObjectID && p.AbilityIndex == 0 {
			return
		}
	}
	t.Fatal("Heart-Shaped Herb's activated ability was not enumerated by legalActions")
}

// TestHeartShapedHerbSacrificeReturnsCreatureWithCountersAndMonarch drives the
// full engine path: activate {2},{T},Sacrifice this artifact, accept the
// optional sacrifice of a specific creature, and confirm that creature returns
// to the battlefield under its owner's control with three +1/+1 counters while
// its controller becomes the monarch. A second creature that is not chosen must
// be left untouched.
func TestHeartShapedHerbSacrificeReturnsCreatureWithCountersAndMonarch(t *testing.T) {
	g, engine, herb := newHerbGame(t)
	victim := addHerbCreature(g, game.Player1, "Chosen Sacrifice")
	victimCardID := victim.CardInstanceID
	bystander := addHerbCreature(g, game.Player1, "Untouched Creature")
	bystanderCardID := bystander.CardInstanceID

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &herbChoiceAgent{sacrifice: true, sacName: "Chosen Sacrifice"},
	}
	assertHerbActivationEnumerated(t, engine, g, herb)
	act := action.ActivateAbility(herb.ObjectID, 0, nil, 0)
	if !engine.applyActionWithChoices(g, game.Player1, act, agents, &TurnLog{}) {
		t.Fatal("applyActionWithChoices(activate Heart-Shaped Herb) = false, want true")
	}
	if permanentByCardID(g, herb.CardInstanceID) != nil {
		t.Fatal("Heart-Shaped Herb was not sacrificed as an activation cost")
	}
	if !g.Players[game.Player1].Graveyard.Contains(herb.CardInstanceID) {
		t.Fatal("sacrificed Heart-Shaped Herb did not reach its owner's graveyard")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	returned := permanentByCardID(g, victimCardID)
	if returned == nil {
		t.Fatal("the sacrificed creature did not return to the battlefield")
	}
	if returned.Owner != game.Player1 || returned.Controller != game.Player1 {
		t.Fatalf("returned creature owner/controller = %v/%v, want Player1/Player1", returned.Owner, returned.Controller)
	}
	if got := returned.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("returned creature +1/+1 counters = %d, want 3", got)
	}
	monarch := currentMonarch(g)
	if !monarch.Exists || monarch.Val != game.Player1 {
		t.Fatalf("monarch = %+v, want Player1", monarch)
	}

	if bystander := permanentByCardID(g, bystanderCardID); bystander == nil {
		t.Fatal("the unchosen creature was sacrificed")
	} else if bystander.Counters.Get(counter.PlusOnePlusOne) != 0 {
		t.Fatal("the unchosen creature gained +1/+1 counters")
	}
}

// TestHeartShapedHerbDeclineSacrificeDoesNothingAfterCost proves the "If you do"
// gate: declining the optional sacrifice leaves every creature in place, returns
// nothing, adds no counters, and does not make the controller the monarch — even
// though the artifact is still sacrificed to pay the cost.
func TestHeartShapedHerbDeclineSacrificeDoesNothingAfterCost(t *testing.T) {
	g, engine, herb := newHerbGame(t)
	creature := addHerbCreature(g, game.Player1, "Spared Creature")
	creatureCardID := creature.CardInstanceID

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &herbChoiceAgent{sacrifice: false},
	}
	act := action.ActivateAbility(herb.ObjectID, 0, nil, 0)
	if !engine.applyActionWithChoices(g, game.Player1, act, agents, &TurnLog{}) {
		t.Fatal("applyActionWithChoices(activate Heart-Shaped Herb) = false, want true")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	spared := permanentByCardID(g, creatureCardID)
	if spared == nil {
		t.Fatal("the creature was sacrificed even though the optional sacrifice was declined")
	}
	if got := spared.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("spared creature +1/+1 counters = %d, want 0", got)
	}
	if monarch := currentMonarch(g); monarch.Exists {
		t.Fatalf("monarch = %+v, want none (the become-monarch clause is gated on the sacrifice)", monarch)
	}
	if !g.Players[game.Player1].Graveyard.Contains(herb.CardInstanceID) {
		t.Fatal("Heart-Shaped Herb should still be sacrificed as an activation cost")
	}
}

// TestHeartShapedHerbPreventsOneDamageFromOpponentSource exercises the card's
// prevention replacement through the real damage pipeline: it prevents 1 damage
// from an opponent-controlled source to its controller, never over-prevents, and
// leaves the controller's own sources unaffected.
func TestHeartShapedHerbPreventsOneDamageFromOpponentSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, cardh.HeartShapedHerb)
	opponentSource := addColoredSourceCard(g, game.Player2, color.Red)
	ownSource := addColoredSourceCard(g, game.Player1, color.Red)

	if dealt := dealPlayerDamage(g, opponentSource, 0, game.Player2, game.Player1, 3, false); dealt != 2 {
		t.Fatalf("opponent-source damage to controller = %d, want 2 (1 prevented)", dealt)
	}
	if dealt := dealPlayerDamage(g, opponentSource, 0, game.Player2, game.Player1, 1, false); dealt != 0 {
		t.Fatalf("small opponent-source damage = %d, want 0 (fully prevented)", dealt)
	}
	if dealt := dealPlayerDamage(g, ownSource, 0, game.Player1, game.Player1, 3, false); dealt != 3 {
		t.Fatalf("own-source damage = %d, want 3 (not opponent-controlled)", dealt)
	}
}
