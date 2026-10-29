package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: issue #3435. Default-registry cards checked against the
// pinned corpus: War Behemoth 652109b9-d607-42b6-945d-0c0dd5bba89c ({5}{W}
// Beast, Morph {4}{W}; 2014-09-20 rulings: the face-down spell has no mana
// cost, mana value 0, is colorless, and {3} is an alternative cost), Alley
// Assailant edf238c9-61de-4f3a-b82f-05af46e5e81b (Disguise {4}{B}{B}),
// Lodestone Golem 683a5707-cddb-494d-9b41-51b4584ded69 ("Nonartifact spells
// cost {1} more to cast."; 2010-03-01 rulings: affects every nonartifact
// spell's total cost, increases before reductions), Heartless Summoning
// c50fdb09-e06c-463c-87dd-3c46f4db889a ("Creature spells you cast cost {2}
// less to cast."; 2025-01-24 ruling: reduces only generic mana), Pearl
// Medallion 1225264f-b298-4b56-b208-c5da87bd6867 ("White spells you cast cost
// {1} less to cast."; 2023-07-28 rulings: reductions apply to alternative
// costs), Krosan Drover e92a7141-119f-4bf8-a82d-eb7c0c37185c ("Creature spells
// you cast with mana value 6 or greater cost {2} less to cast.") and Thalia,
// Guardian of Thraben c9f8b8fb-1cd8-450e-a1fe-892e7a323479 ("Noncreature
// spells cost {1} more to cast."). CR 2026-09-25: 702.37c and 702.168b turn
// the card into a 2/2 face-down creature with no name, subtypes or mana cost,
// apply effects that would apply to casting a card with those characteristics
// (not the face-up card's), and pay {3} following the alternative-cost rules;
// 601.2a moves the card to the stack first; 601.2f adds increases and
// subtracts reductions to lock in the total cost; 601.2h pays it.

func faceDownCostGame(t *testing.T, hand string, color mana.Color, amount int, battlefield ...string) (*game.Game, *Engine, id.ID) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, productionChoiceCard(t, hand))
	for _, name := range battlefield {
		addCombatPermanent(g, game.Player2, productionChoiceCard(t, name))
	}
	g.Players[game.Player1].ManaPool.Add(color, amount)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	return g, engine, cardID
}

// castFaceDownPaying casts cardID face down from hand and checks that exactly
// spent mana was paid and recorded on the spell and its cast event.
func castFaceDownPaying(t *testing.T, engine *Engine, g *game.Game, cardID id.ID, kind game.FaceDownKind, spent int) *game.StackObject {
	t.Helper()
	before := g.Players[game.Player1].ManaPool.Total()
	if !faceDownCastOffered(engine, g, cardID, kind) {
		t.Fatal("face-down cast not offered")
	}
	if !engine.applyAction(g, game.Player1, actionBuild.castFaceDown(cardID, game.FaceFront, kind)) {
		t.Fatal("face-down cast failed")
	}
	if got := before - g.Players[game.Player1].ManaPool.Total(); got != spent {
		t.Fatalf("face-down cast spent %d mana, want %d", got, spent)
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.FaceDown || obj.SourceID != cardID || obj.FaceDownKind != kind || obj.SourceZone != zone.Hand {
		t.Fatalf("stack top = %+v, want face-down spell cast from hand", obj)
	}
	if obj.ManaSpentToCast != spent {
		t.Fatalf("stack object ManaSpentToCast = %d, want %d", obj.ManaSpentToCast, spent)
	}
	assertEvent(t, g.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.CardID == cardID && event.FromZone == zone.Hand && event.ToZone == zone.Stack &&
			event.ManaSpentToCast.Exists && event.ManaSpentToCast.Val == spent &&
			event.ManaValue.Exists && event.ManaValue.Val == 0 && len(event.Colors) == 0 &&
			len(event.CardTypes) == 1 && event.CardTypes[0] == types.Creature
	})
	return obj
}

// TestLodestoneGolemTaxesFaceDownCastFromHand proves the {3} face-down cost
// is a nonartifact spell's alternative cost that Lodestone Golem increases: {3}
// is not enough and {4} is spent.
func TestLodestoneGolemTaxesFaceDownCastFromHand(t *testing.T) {
	g, engine, behemoth := faceDownCostGame(t, "War Behemoth", mana.W, 3, "Lodestone Golem")
	assertFaceDownCastRejected(t, engine, g, behemoth, game.FaceDownMorph)
	if got := g.Players[game.Player1].ManaPool.Total(); got != 3 {
		t.Fatalf("rejected face-down cast changed mana pool to %d, want 3", got)
	}
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast || event.Kind == game.EventZoneChanged && event.CardID == behemoth {
			t.Fatalf("rejected face-down cast emitted %+v", event)
		}
	}

	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	castFaceDownPaying(t, engine, g, behemoth, game.FaceDownMorph, 4)
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentByCardID(g, behemoth)
	if permanent == nil || !permanent.FaceDown || !faceDownIsTwoTwo(g, permanent) {
		t.Fatalf("taxed face-down spell resolved as %+v, want face-down 2/2", permanent)
	}
}

// TestHeartlessSummoningReducesFaceDownDisguiseCast proves a creature-spell
// reduction applies to the face-down alternative cost of a disguise card.
func TestHeartlessSummoningReducesFaceDownDisguiseCast(t *testing.T) {
	g, engine, assailant := faceDownCostGame(t, "Alley Assailant", mana.B, 0)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Heartless Summoning"))
	assertFaceDownCastRejected(t, engine, g, assailant, game.FaceDownDisguise)

	g.Players[game.Player1].ManaPool.Add(mana.B, 2)
	castFaceDownPaying(t, engine, g, assailant, game.FaceDownDisguise, 1)
	if got := g.Players[game.Player1].ManaPool.Total(); got != 1 {
		t.Fatalf("mana left after reduced face-down cast = %d, want 1", got)
	}
}

// TestFaceDownCostIgnoresFaceUpCharacteristicFilters proves that modifiers
// keyed to the hidden card's color, mana value or noncreature type do not
// apply to the face-down spell (CR 702.37c), while they still apply to the
// same card cast face up.
func TestFaceDownCostIgnoresFaceUpCharacteristicFilters(t *testing.T) {
	g, engine, behemoth := faceDownCostGame(t, "War Behemoth", mana.W, 2)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Pearl Medallion"))
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Krosan Drover"))
	addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Thalia, Guardian of Thraben"))

	assertFaceDownCastRejected(t, engine, g, behemoth, game.FaceDownMorph)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	if !engine.canCastSpell(g, game.Player1, behemoth, nil, 0, nil) {
		t.Fatal("face-up War Behemoth not castable for {2}{W} with white and mana value reductions")
	}
	castFaceDownPaying(t, engine, g, behemoth, game.FaceDownMorph, 3)
}
