package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Scryfall rulings for Carnivorous Cultivator // Enroot
// (79dd5c54-5ea5-47b5-8f9b-50ed57a5ea45, 2026-08-21) say the prepare spell copy
// is created in exile and its controller casts it from there, and that the cast
// isn't an alternative cost. Drannith Magistrate (98b0a4a8-9319-451b-9b79-
// b0bca7a41e91) rulings say its restriction overrides permissions to cast from
// other zones. Sage of the Beyond (226a4c5d-bf49-40d8-a7a4-341c18029634)
// reduces spells cast from anywhere other than hand. Rule of Law, Grand
// Abolisher and Blazing Firesinger // Seething Song (3ba971e7-0b7a-4750-896f-
// 7cf063e66b2a) Oracle text is from the pinned corpus. CR 2026-09-25 from the
// verified official copy: 722.3c (prepared copy created in exile and cast from
// there) and 601.3 (a player can begin casting only if no rule or effect
// prohibits it).

func preparedExileCastGame(t *testing.T, name string, color mana.Color, amount int) (*game.Game, *Engine, id.ID, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, registryCard(t, name))
	g.Players[game.Player1].ManaPool.Add(color, amount)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("casting %s failed", name)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentForCard(g, cardID)
	if permanent == nil || !permanent.Prepared {
		t.Fatalf("%s permanent = %+v, want prepared", name, permanent)
	}
	return g, engine, cardID, permanent
}

func preparedCopyCastLegal(g *game.Game, engine *Engine, playerID game.PlayerID, cast action.Action) bool {
	return slices.ContainsFunc(engine.legalActions(g, playerID), func(got action.Action) bool {
		return actionsEqual(got, cast)
	})
}

func TestPreparedCopyIsCastFromExile(t *testing.T) {
	g, engine, cardID, permanent := preparedExileCastGame(t, "Carnivorous Cultivator", mana.G, 3)
	cast := action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("casting Enroot's prepared copy failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Copy || obj.SourceZone != zone.Exile || obj.SourceCardID != cardID {
		t.Fatalf("stack object = %+v, want prepared copy cast from exile linked to card %v", obj, cardID)
	}
	if permanent.Prepared || permanentForCard(g, cardID) != permanent {
		t.Fatalf("permanent = %+v, want the same permanent on the battlefield unprepared", permanent)
	}
	var castEvents []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast && event.StackObjectID == obj.ID {
			castEvents = append(castEvents, event)
		}
	}
	if len(castEvents) != 1 || castEvents[0].FromZone != zone.Exile || castEvents[0].PermanentID != permanent.ObjectID {
		t.Fatalf("cast events = %+v, want one cast from exile linked to the prepared permanent", castEvents)
	}
}

func TestOpponentsDrannithMagistrateStopsPreparedCopyCast(t *testing.T) {
	g, engine, cardID, permanent := preparedExileCastGame(t, "Carnivorous Cultivator", mana.G, 3)
	cast := action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)

	addCombatPermanent(g, game.Player1, registryCard(t, "Drannith Magistrate"))
	if !preparedCopyCastLegal(g, engine, game.Player1, cast) {
		t.Fatal("your own Drannith Magistrate stopped your prepared copy cast")
	}

	addCombatPermanent(g, game.Player2, registryCard(t, "Drannith Magistrate"))
	if preparedCopyCastLegal(g, engine, game.Player1, cast) {
		t.Fatal("prepared copy cast is legal under an opponent's Drannith Magistrate")
	}
	if engine.applyAction(g, game.Player1, cast) {
		t.Fatal("prepared copy was cast from exile under an opponent's Drannith Magistrate")
	}
	if !permanent.Prepared || g.Stack.Size() != 0 {
		t.Fatalf("prepared = %v, stack = %d; want the permanent still prepared and nothing cast", permanent.Prepared, g.Stack.Size())
	}
}

func TestRuleOfLawCountsPreparedCopyCast(t *testing.T) {
	g, engine, cardID, permanent := preparedExileCastGame(t, "Carnivorous Cultivator", mana.G, 3)
	cast := action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
	if !preparedCopyCastLegal(g, engine, game.Player1, cast) {
		t.Fatal("prepared copy cast is not legal without Rule of Law")
	}

	addCombatPermanent(g, game.Player2, registryCard(t, "Rule of Law"))
	if preparedCopyCastLegal(g, engine, game.Player1, cast) {
		t.Fatal("prepared copy cast is legal after casting a spell this turn under Rule of Law")
	}
	if engine.applyAction(g, game.Player1, cast) || !permanent.Prepared {
		t.Fatal("Rule of Law did not stop the second spell of the turn")
	}
}

func TestGrandAbolisherStopsPreparedCopyCastOnItsControllersTurn(t *testing.T) {
	g, engine, cardID, permanent := preparedExileCastGame(t, "Blazing Firesinger", mana.R, 3)
	cast := action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
	g.Turn.ActivePlayer = game.Player2
	g.Players[game.Player1].ManaPool.Add(mana.R, 3)
	if !preparedCopyCastLegal(g, engine, game.Player1, cast) {
		t.Fatal("Seething Song's prepared copy cannot be cast on an opponent's turn")
	}

	addCombatPermanent(g, game.Player2, registryCard(t, "Grand Abolisher"))
	if preparedCopyCastLegal(g, engine, game.Player1, cast) || engine.applyAction(g, game.Player1, cast) {
		t.Fatal("prepared copy was castable during Grand Abolisher's controller's turn")
	}
	if !permanent.Prepared {
		t.Fatal("blocked prepared copy cast unprepared the permanent")
	}
}

func TestSageOfTheBeyondReducesPreparedCopyCastFromExile(t *testing.T) {
	g, engine, cardID, permanent := preparedExileCastGame(t, "Blazing Firesinger", mana.R, 3)
	addCombatPermanent(g, game.Player1, registryCard(t, "Sage of the Beyond"))
	pool := &g.Players[game.Player1].ManaPool
	if pool.Total() != 0 {
		t.Fatalf("mana pool total = %d after casting Blazing Firesinger, want 0", pool.Total())
	}
	pool.Add(mana.R, 1)
	cast := action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
	if !preparedCopyCastLegal(g, engine, game.Player1, cast) {
		t.Fatal("Sage of the Beyond did not make Seething Song's prepared copy castable for {R}")
	}
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("casting Seething Song's prepared copy for {R} failed")
	}
	if permanent.Prepared || pool.Total() != 0 {
		t.Fatalf("prepared = %v, pool total = %d; want unprepared after paying {R}", permanent.Prepared, pool.Total())
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := pool.Amount(mana.R); got != 5 {
		t.Fatalf("red mana after Seething Song = %d, want 5", got)
	}
}
