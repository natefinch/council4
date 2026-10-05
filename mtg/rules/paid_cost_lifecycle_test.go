package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestPaidCostCounteredStackEntryDoesNotResolveOrRefund(t *testing.T) {
	for _, spell := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		paidID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Paid", Types: []types.Card{types.Creature}, Colors: []color.Color{color.Blue},
		}})
		text := "Discard a card: You gain 1 life. If the discarded card was blue, you gain 2 life."
		typeLine := "Artifact"
		if spell {
			text = "As an additional cost to cast this spell, discard a card.\nYou gain 1 life. If the discarded card was blue, you gain 2 life."
			typeLine = "Sorcery"
		}
		def := compiledPaidSubjectCard(t, text, typeLine)
		setSorcerySpeedTurn(g, game.Player1)
		life := g.Players[game.Player1].Life
		var activate action.Action
		if spell {
			sourceID := addCardToHand(g, game.Player1, def)
			activate = action.CastSpell(sourceID, nil, 0, nil)
		} else {
			source := addCombatPermanent(g, game.Player1, def)
			activate = action.ActivateAbility(source.ObjectID, 0, nil, 0)
		}
		if !engine.applyAction(g, game.Player1, activate) {
			t.Fatal("actual payment failed")
		}
		obj, ok := g.Stack.Peek()
		if !ok || len(obj.PaidCostSubjects) != 1 || obj.PaidCostSubjects[0].Snapshot.CardID != paidID {
			t.Fatal("actual payment was not attached to its stack entry")
		}
		if !counterStackObject(g, obj.ID) {
			t.Fatal("countering paid stack entry failed")
		}
		if _, remains := g.Stack.Peek(); remains || g.Players[game.Player1].Life != life ||
			!g.Players[game.Player1].Graveyard.Contains(paidID) {
			t.Fatal("countering refunded the cost or resolved its body")
		}
	}
}

func TestPaidCostClonedGameUsesIndependentFrozenPayment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, compiledPaidSubjectCard(t,
		"Sacrifice a creature: You gain 1 life. If the sacrificed creature was red, you gain 2 life.", "Artifact"))
	paid := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Paid", Types: []types.Card{types.Creature}, Colors: []color.Color{color.Red},
	}})
	setSorcerySpeedTurn(g, game.Player1)
	life := g.Players[game.Player1].Life
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("activation failed")
	}
	cloned := g.Clone()
	original, _ := g.Stack.Peek()
	original.PaidCostSubjects[0].Snapshot.Colors[0] = color.Blue
	clonedSource, ok := permanentByObjectID(cloned, source.ObjectID)
	if !ok || !movePermanentToZone(cloned, clonedSource, zone.Graveyard) {
		t.Fatal("removing the cloned ability source failed")
	}
	delete(cloned.LastKnownInformation, paid.ObjectID)
	engine.resolveTopOfStack(cloned, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if cloned.Players[game.Player1].Life != life+3 || g.Players[game.Player1].Life != life+1 {
		t.Fatal("cloned stack entries shared mutable cost facts or required the later source")
	}
}

func TestPaidCostInvalidConsumerTargetCannotPayOrPublish(t *testing.T) {
	for _, spell := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		paidID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Paid", Types: []types.Card{types.Land},
		}})
		text := "Discard a card: Tap target creature. If the discarded card was a land card, you gain 2 life."
		typeLine := "Artifact"
		if spell {
			text = "As an additional cost to cast this spell, discard a card.\nTap target creature. If the discarded card was a land card, you gain 2 life."
			typeLine = "Instant"
		}
		def := compiledPaidSubjectCard(t, text, typeLine)
		setSorcerySpeedTurn(g, game.Player1)
		invalidTarget := []game.Target{game.PermanentTarget(g.IDGen.Next())}
		var activate action.Action
		if spell {
			sourceID := addCardToHand(g, game.Player1, def)
			activate = action.CastSpell(sourceID, invalidTarget, 0, nil)
		} else {
			source := addCombatPermanent(g, game.Player1, def)
			activate = action.ActivateAbility(source.ObjectID, 0, invalidTarget, 0)
		}
		life := g.Players[game.Player1].Life
		if engine.applyAction(g, game.Player1, activate) {
			t.Fatal("invalid consumer target bypassed activation/cast validation")
		}
		if _, published := g.Stack.Peek(); published || !g.Players[game.Player1].Hand.Contains(paidID) ||
			g.Players[game.Player1].Graveyard.Contains(paidID) || g.Players[game.Player1].Life != life {
			t.Fatal("failed target validation paid a cost, published a subject, or resolved a body")
		}
	}
}
