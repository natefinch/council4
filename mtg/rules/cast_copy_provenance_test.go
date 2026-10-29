package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine gap #3388 (cast copies of permanent cards lose cast
// provenance on entry). Oracle text from the pinned corpus for Iridescent Tiger
// ("When this creature enters, if you cast it, add {W}{U}{B}{R}{G}."), Sphinx's
// Insight (Addendum — "If you cast this spell during your main phase, you gain
// 2 life."), Skyrider Elf (Converge enters-with-counters) and Sphere of
// Resistance ("Spells cost {1} more to cast."). Scryfall rulings: Iridescent
// Tiger 2025-04-04 (its ability triggers if you cast it from any zone, but not if
// it is put onto the battlefield without being cast); Sphinx's Insight
// 2024-01-12 (a copy of an Addendum spell was never cast, so it gets no bonus);
// Wondrous Crucible and Shiko, Paragon of the Way (you cast the copy while the
// ability resolves; a resolving permanent-spell copy becomes a token that isn't
// "created"). CR 2026-09-25 from the official TXT
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt):
// 707.10 (a copy put onto the stack isn't cast; the Dawnglow Infusion example:
// no mana was spent to cast a copy), 707.12/707.12a (casting a copy follows
// 601.2a-h, so that copy was cast), 111.13 (a resolving permanent spell copy
// becomes a token), 601.2, 603.4 (intervening "if" conditions) and 722.3c (a
// prepared permanent's spell copy is cast). A cast copy keeps its own cast
// controller, cast zone, timing and mana spent; copying a cast copy or storm
// copying it produces an uncast copy with none of those facts.

func castCopyProvenanceGame() *game.Game {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	return g
}

func castCopyOfExiledCard(t *testing.T, engine *Engine, g *game.Game, cardID id.ID) *game.StackObject {
	t.Helper()
	if !engine.castFreeCopyOfCard(g, game.Player1, cardID, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("casting a copy of the exiled card failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Copy || !obj.CastCopy || obj.SourceZone != zone.Exile || obj.Controller != game.Player1 {
		t.Fatalf("stack object = %+v, want a cast copy from exile controlled by Player1", obj)
	}
	return obj
}

func spellCastEvents(g *game.Game, stackObjectID id.ID) int {
	count := 0
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast && event.StackObjectID == stackObjectID {
			count++
		}
	}
	return count
}

func newestTokenNamed(g *game.Game, name string) *game.Permanent {
	var found *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == name {
			found = permanent
		}
	}
	return found
}

func enteredEventFor(g *game.Game, permanent *game.Permanent) (game.Event, bool) {
	for i := len(g.Events) - 1; i >= 0; i-- {
		event := g.Events[i]
		if event.Kind == game.EventPermanentEnteredBattlefield && event.PermanentID == permanent.ObjectID {
			return event, true
		}
	}
	return game.Event{}, false
}

func wubrgPool(g *game.Game) bool {
	pool := &g.Players[game.Player1].ManaPool
	for _, c := range []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G} {
		if pool.Amount(c) != 1 {
			return false
		}
	}
	return pool.Total() == 5
}

// resolveCastCopyTiger resolves the cast Iridescent Tiger copy on top of the
// stack and its enters trigger, returning the resulting token.
func resolveCastCopyTiger(t *testing.T, engine *Engine, g *game.Game, tigerCardID id.ID) *game.Permanent {
	t.Helper()
	engine.resolveTopOfStack(g, &TurnLog{})
	token := newestTokenNamed(g, "Iridescent Tiger")
	if token == nil {
		t.Fatal("the cast Iridescent Tiger copy did not become a token")
	}
	if !token.EnteredFromCast || !token.HasSpellAnnouncement {
		t.Fatalf("token cast state = cast:%v announced:%v, want a cast entry", token.EnteredFromCast, token.HasSpellAnnouncement)
	}
	event, ok := enteredEventFor(g, token)
	if !ok || !event.EnterWasCast || !event.EnterHasCastController || event.EnterCastController != game.Player1 ||
		event.EnterCastFromZone != zone.Exile {
		t.Fatalf("enter event = %+v, want cast by Player1 from exile", event)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Iridescent Tiger's \"if you cast it\" trigger did not trigger for a cast copy")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !wubrgPool(g) {
		t.Fatalf("mana pool = %+v, want {W}{U}{B}{R}{G}", g.Players[game.Player1].ManaPool)
	}
	if !g.Players[game.Player1].Exile.Contains(tigerCardID) {
		t.Fatal("casting a copy moved the exiled Iridescent Tiger card")
	}
	return token
}

func TestCastCopyOfPermanentCardEntersAsCast(t *testing.T) {
	g := castCopyProvenanceGame()
	engine := NewEngine(nil)
	tigerCardID := addCardToExile(g, game.Player1, registryCard(t, "Iridescent Tiger"))
	castCopyOfExiledCard(t, engine, g, tigerCardID)

	replay := g.Clone()
	if obj, ok := replay.Stack.Peek(); !ok || !obj.Copy || !obj.CastCopy {
		t.Fatalf("cloned stack object = %+v, want the cast copy", obj)
	}
	resolveCastCopyTiger(t, engine, g, tigerCardID)
	resolveCastCopyTiger(t, engine, replay, tigerCardID)
	if len(g.Events) != len(replay.Events) {
		t.Fatalf("clone replay diverged: %d events, want %d", len(replay.Events), len(g.Events))
	}
}

func TestUncastCopiesOfCastCopyEnterWithoutCastFacts(t *testing.T) {
	g := castCopyProvenanceGame()
	engine := NewEngine(nil)
	tigerDef := registryCard(t, "Iridescent Tiger")
	tigerCardID := addCardToExile(g, game.Player1, tigerDef)
	castCopy := castCopyOfExiledCard(t, engine, g, tigerCardID)

	if spellCastEvents(g, castCopy.ID) != 1 {
		t.Fatal("casting a copy did not emit exactly one spell-cast event")
	}
	resolveStormTriggerFor(engine, g, castCopy, 1)
	addEffectSpellToStack(g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(castCopy.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Stack.Size() != 3 {
		t.Fatalf("stack size = %d, want cast copy, storm copy and stack copy", g.Stack.Size())
	}
	for _, obj := range g.Stack.Objects() {
		if obj.ID != castCopy.ID && spellCastEvents(g, obj.ID) != 0 {
			t.Fatalf("uncast copy %+v emitted a spell-cast event", obj)
		}
	}

	for _, label := range []string{"copy of the cast copy", "storm copy"} {
		obj, _ := g.Stack.Peek()
		if !obj.Copy || obj.CastCopy {
			t.Fatalf("%s = %+v, want an uncast copy", label, obj)
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		token := newestTokenNamed(g, "Iridescent Tiger")
		if token == nil || token.EnteredFromCast || !token.HasSpellAnnouncement {
			t.Fatalf("%s token = %+v, want an uncast announced spell copy", label, token)
		}
		if event, ok := enteredEventFor(g, token); !ok || event.EnterWasCast || event.EnterHasCastController {
			t.Fatalf("%s enter event = %+v, want no cast facts", label, event)
		}
		if engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatalf("Iridescent Tiger triggered for the uncast %s", label)
		}
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("an uncast copy added mana")
	}

	if obj, _ := g.Stack.Peek(); obj.ID != castCopy.ID {
		t.Fatalf("remaining stack object = %+v, want the original cast copy", obj)
	}
	resolveCastCopyTiger(t, engine, g, tigerCardID)
}

func TestCastCopyAddendumUsesItsOwnCastTiming(t *testing.T) {
	for _, tc := range []struct {
		name       string
		active     game.PlayerID
		copyOnTop  bool
		wantGained int
	}{
		{name: "cast copy in controller's main phase", active: game.Player1, wantGained: 2},
		{name: "cast copy on another player's turn", active: game.Player2, wantGained: 0},
		{name: "copy of the cast copy", active: game.Player1, copyOnTop: true, wantGained: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := castCopyProvenanceGame()
			g.Turn.ActivePlayer = tc.active
			engine := NewEngine(nil)
			cardID := addCardToExile(g, game.Player1, registryCard(t, "Sphinx's Insight"))
			castCopy := castCopyOfExiledCard(t, engine, g, cardID)
			if castCopy.CastDuringControllerMainPhase != (tc.active == game.Player1) {
				t.Fatalf("cast timing = %v, want %v", castCopy.CastDuringControllerMainPhase, tc.active == game.Player1)
			}
			if tc.copyOnTop {
				addEffectSpellToStack(g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
					[]game.Target{game.StackObjectTarget(castCopy.ID)})
				engine.resolveTopOfStack(g, &TurnLog{})
			}
			life := g.Players[game.Player1].Life
			engine.resolveTopOfStack(g, &TurnLog{})
			if got := g.Players[game.Player1].Life - life; got != tc.wantGained {
				t.Fatalf("life gained = %d, want %d", got, tc.wantGained)
			}
			if !g.Players[game.Player1].Exile.Contains(cardID) {
				t.Fatal("resolving a copy moved the exiled card")
			}
		})
	}
}

func TestPreparedSpellCopyIsCastWithMainPhaseTiming(t *testing.T) {
	g := castCopyProvenanceGame()
	engine := NewEngine(nil)
	def := prepareCard()
	alternate := def.Alternate.Val
	alternate.Name = "Prepared Addendum"
	alternate.SpellAbility = opt.Val(game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()},
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{CastDuringControllerMainPhase: true}),
			}),
		}},
	}.Ability())
	def.Alternate = opt.Val(alternate)
	cardID := addCardToHand(g, game.Player1, def)
	for range 4 {
		addBasicLandPermanent(g, game.Player1, types.Plains)
	}
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("casting the prepare creature failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("casting the prepared spell copy failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Copy || !obj.CastCopy || !obj.CastDuringControllerMainPhase || obj.ManaSpentToCast != 1 {
		t.Fatalf("prepared stack object = %+v, want a cast copy cast in the main phase with {W} spent", obj)
	}
	life := g.Players[game.Player1].Life
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != life+2 {
		t.Fatalf("life = %d, want %d", g.Players[game.Player1].Life, life+2)
	}
}

func TestCastCopyCarriesManaSpentIntoEntryReplacement(t *testing.T) {
	g := castCopyProvenanceGame()
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player2, registryCard(t, "Sphere of Resistance"))
	cardID := addCardToExile(g, game.Player1, registryCard(t, "Skyrider Elf"))
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	castCopy := castCopyOfExiledCard(t, engine, g, cardID)
	if castCopy.XValue != 0 || castCopy.ManaSpentToCast != 1 || castCopy.ColorsOfManaSpentToCast != 1 {
		t.Fatalf("cast copy X=%d spent=%d colors=%d, want X=0 with {G} spent for Sphere of Resistance",
			castCopy.XValue, castCopy.ManaSpentToCast, castCopy.ColorsOfManaSpentToCast)
	}
	addEffectSpellToStack(g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(castCopy.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if obj, _ := g.Stack.Peek(); !obj.Copy || obj.CastCopy {
		t.Fatalf("copy of the cast copy = %+v, want an uncast copy", obj)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	uncast := newestTokenNamed(g, "Skyrider Elf")
	if uncast == nil || uncast.Counters.Get(counter.PlusOnePlusOne) != 0 {
		t.Fatalf("uncast copy token = %+v, want no Converge counters", uncast)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	cast := newestTokenNamed(g, "Skyrider Elf")
	if cast == nil || cast == uncast || cast.Counters.Get(counter.PlusOnePlusOne) != 1 {
		t.Fatalf("cast copy token = %+v, want one Converge counter for {G}", cast)
	}
}
