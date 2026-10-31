package rules

import (
	"testing"

	cardregistry "github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func unearthSource() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Unearth Source",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		ActivatedAbilities: []game.ActivatedAbility{
			game.UnearthActivatedAbility(cost.Mana{cost.O(0)}),
		},
	}}
}

func TestUnearthReturnsSourceToBattlefieldWithHaste(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	cardID := addCardToGraveyard(g, game.Player1, unearthSource())

	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	act := action.ActivateAbility(cardID, 0, nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("unearth ability was not legal from the graveyard")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("unearth activation failed")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("unearth source remained in graveyard after resolution")
	}
	var permanent *game.Permanent
	for _, p := range g.Battlefield {
		if p.CardInstanceID == cardID {
			permanent = p
			break
		}
	}
	if permanent == nil {
		t.Fatal("unearth source did not return to the battlefield")
	}
	if permanent.Controller != game.Player1 {
		t.Fatalf("unearthed permanent controller = %v; want %v", permanent.Controller, game.Player1)
	}
	if !hasKeyword(g, permanent, game.Haste) {
		t.Fatal("unearthed permanent did not gain haste")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d; want 1 end-step exile", len(g.DelayedTriggers))
	}
	if g.DelayedTriggers[0].Timing != game.DelayedAtBeginningOfNextEndStep {
		t.Fatalf("delayed trigger timing = %v; want next end step", g.DelayedTriggers[0].Timing)
	}
}

// Source review: Unearth (CR 702.84a) and the Whip of Erebos clause share
// "Exile it at the beginning of the next end step. If it would leave the
// battlefield, exile it instead of putting it anywhere else." Rules read from
// the Comprehensive Rules effective September 25, 2026
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt):
// 702.84a, 614.1a, 614.2-614.7, 616.1, 400.7, 603.6c, 603.10a, 611.2a and
// 903.9a/b. Scrapwork Cohort rulings (2022-10-14,
// https://api.scryfall.com/cards/91b0f17e-603a-4840-b2c3-381ec6f102f3/rulings):
// unearth does nothing if the card left the graveyard before resolution; the
// end-step trigger can be countered but the replacement still exiles it later;
// neither exile ability is granted to the permanent, so losing abilities does
// not matter; leaving for any reason exiles it instead, but an effect that
// exiles it succeeds; once it leaves, a later return is a new object that
// unearth no longer affects.

func activateAndResolveUnearth(t *testing.T, engine *Engine, g *game.Game, cardID id.ID) *game.Permanent {
	t.Helper()
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cardID, 0, nil, 0)) {
		t.Fatal("unearth activation failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	return unearthedPermanent(g, cardID)
}

func unearthedPermanent(g *game.Game, cardID id.ID) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return permanent
		}
	}
	return nil
}

func assertExiledNotElsewhere(t *testing.T, g *game.Game, cardID id.ID) {
	t.Helper()
	player := g.Players[game.Player1]
	if !player.Exile.Contains(cardID) {
		t.Fatal("unearthed card was not exiled")
	}
	if player.Graveyard.Contains(cardID) || player.Hand.Contains(cardID) || player.Library.Contains(cardID) {
		t.Fatal("unearthed card reached a zone other than exile")
	}
	if unearthedPermanent(g, cardID) != nil {
		t.Fatal("unearthed card is still on the battlefield")
	}
}

func fireUnearthEndStepTrigger(t *testing.T, engine *Engine, g *game.Game) {
	t.Helper()
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want 1 unearth end-step trigger", len(g.DelayedTriggers))
	}
	pending := drainReadyDelayedTriggers(g, []game.Event{{Kind: game.EventBeginningOfStep, Step: game.StepEnd}})
	if len(pending) != 1 {
		t.Fatalf("ready delayed triggers = %d, want 1", len(pending))
	}
	trigger := pending[0]
	obj := &game.StackObject{
		Controller:       trigger.controller,
		SourceID:         trigger.sourceID,
		SourceCardID:     trigger.sourceCardID,
		CapturedObjectID: trigger.capturedObjectID,
	}
	engine.resolveAbilityContentWithChoices(g, obj, trigger.inline.Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}

// Destroying, sacrificing, bouncing or tucking the unearthed permanent exiles
// it instead (CR 702.84a, CR 614.1a), with no dies event.
func TestUnearthExilesInsteadOfLeavingBattlefieldElsewhere(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		leave func(g *game.Game, permanent *game.Permanent) bool
	}{
		{"destroy", func(g *game.Game, permanent *game.Permanent) bool {
			_, ok := destroyPermanent(g, permanent.ObjectID)
			return ok
		}},
		{"sacrifice", sacrificePermanent},
		{"bounce to hand", func(g *game.Game, permanent *game.Permanent) bool {
			return movePermanentToZone(g, permanent, zone.Hand)
		}},
		{"put into library", func(g *game.Game, permanent *game.Permanent) bool {
			return movePermanentToZone(g, permanent, zone.Library)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := addCardToGraveyard(g, game.Player1, unearthSource())
			permanent := activateAndResolveUnearth(t, engine, g, cardID)
			if permanent == nil {
				t.Fatal("unearth did not return the card")
			}
			if len(g.ReplacementEffects) != 1 || g.ReplacementEffects[0].AffectedObjectID != permanent.ObjectID {
				t.Fatalf("replacements = %+v, want one bound to the unearthed object", g.ReplacementEffects)
			}
			eventsBefore := len(g.Events)

			test.leave(g, permanent)

			assertExiledNotElsewhere(t, g, cardID)
			for _, event := range g.Events[eventsBefore:] {
				if event.Kind == game.EventZoneChanged && event.CardID == cardID && event.ToZone != zone.Exile {
					t.Fatalf("zone change to %v, want only exile", event.ToZone)
				}
			}
			expireCleanupDurations(g)
			if len(g.ReplacementEffects) != 0 {
				t.Fatalf("replacements = %d after the object left and cleanup, want 0", len(g.ReplacementEffects))
			}
		})
	}
}

// The replacement is not tied to the end-step trigger or to the current turn:
// it survives cleanup, so the permanent is exiled if it later leaves even when
// the end-step trigger was countered (Scrapwork Cohort ruling).
func TestUnearthReplacementOutlivesCounteredEndStepTrigger(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, unearthSource())
	permanent := activateAndResolveUnearth(t, engine, g, cardID)

	// The end-step trigger is countered: it leaves without resolving.
	g.DelayedTriggers = nil
	expireCleanupDurations(g)
	g.Turn.TurnNumber++
	expireCleanupDurations(g)

	if len(g.ReplacementEffects) != 1 {
		t.Fatalf("replacements = %d after cleanup, want 1", len(g.ReplacementEffects))
	}
	if _, ok := destroyPermanent(g, permanent.ObjectID); !ok {
		t.Fatal("destroy failed")
	}
	assertExiledNotElsewhere(t, g, cardID)
}

// The end-step trigger exiles exactly the returned object; the replacement does
// not interfere with an effect that exiles it, and is discarded at cleanup once
// its object is gone.
func TestUnearthEndStepTriggerExilesReturnedObject(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, unearthSource())
	permanent := activateAndResolveUnearth(t, engine, g, cardID)
	if g.DelayedTriggers[0].CapturedObjectID != permanent.ObjectID {
		t.Fatalf("end-step trigger captured %v, want %v", g.DelayedTriggers[0].CapturedObjectID, permanent.ObjectID)
	}

	fireUnearthEndStepTrigger(t, engine, g)

	assertExiledNotElsewhere(t, g, cardID)
	expireCleanupDurations(g)
	if len(g.ReplacementEffects) != 0 {
		t.Fatalf("replacements = %d after exile and cleanup, want 0", len(g.ReplacementEffects))
	}
}

// Losing all abilities does not stop either exile clause: neither is granted to
// the permanent (Scrapwork Cohort ruling).
func TestUnearthReplacementAppliesAfterPermanentLosesAbilities(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, unearthSource())
	permanent := activateAndResolveUnearth(t, engine, g, cardID)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Duration: game.DurationPermanent, Layer: game.LayerAbility,
		AffectedObjectID: permanent.ObjectID, RemoveAllAbilities: true,
	})

	if !sacrificePermanent(g, permanent) {
		t.Fatal("sacrifice failed")
	}
	assertExiledNotElsewhere(t, g, cardID)
}

// Once the unearthed object leaves, a later return of the same card is a new
// object that neither exile clause affects (CR 400.7).
func TestUnearthDoesNotFollowCardToNewObject(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, unearthSource())
	permanent := activateAndResolveUnearth(t, engine, g, cardID)
	if !movePermanentToZone(g, permanent, zone.Exile) {
		t.Fatal("exile failed")
	}
	assertExiledNotElsewhere(t, g, cardID)

	g.Players[game.Player1].Exile.Remove(cardID)
	g.Players[game.Player1].Graveyard.Add(cardID)
	returned := addCardToBattlefieldFromGraveyard(t, engine, g, cardID)
	if _, ok := destroyPermanent(g, returned.ObjectID); !ok {
		t.Fatal("destroy failed")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("new object should be put into the graveyard normally")
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("new object was exiled by a stale unearth replacement")
	}
}

// If the card left the graveyard before unearth resolved, unearth does nothing:
// it does not bind to a different object of the same card (CR 400.7, Scrapwork
// Cohort ruling).
func TestUnearthDoesNothingWhenCardLeftGraveyard(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, unearthSource())
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cardID, 0, nil, 0)) {
		t.Fatal("unearth activation failed")
	}
	// In response, the card is reanimated by another effect.
	other := addCardToBattlefieldFromGraveyard(t, engine, g, cardID)

	engine.resolveTopOfStack(g, &TurnLog{})

	if len(g.ReplacementEffects) != 0 {
		t.Fatalf("replacements = %d, want 0", len(g.ReplacementEffects))
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers = %d, want 0 when unearth returned nothing", len(g.DelayedTriggers))
	}
	if hasKeyword(g, other, game.Haste) {
		t.Fatal("unrelated object gained haste")
	}
	if _, ok := destroyPermanent(g, other.ObjectID); !ok {
		t.Fatal("destroy failed")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("unrelated object should be put into the graveyard normally")
	}
}

// The replacement and its binding are game-owned state that a cloned game
// carries independently.
func TestUnearthReplacementIsCloneSafe(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, unearthSource())
	permanent := activateAndResolveUnearth(t, engine, g, cardID)

	clone := g.Clone()
	if _, ok := destroyPermanent(clone, permanent.ObjectID); !ok {
		t.Fatal("destroy in clone failed")
	}
	assertExiledNotElsewhere(t, clone, cardID)
	if unearthedPermanent(g, cardID) == nil || len(g.ReplacementEffects) != 1 {
		t.Fatal("original game changed when the clone's permanent left")
	}
	if _, ok := destroyPermanent(g, permanent.ObjectID); !ok {
		t.Fatal("destroy in original failed")
	}
	assertExiledNotElsewhere(t, g, cardID)
}

// A replacement created by a resolving ability lasts after its source leaves
// (CR 611.2a, CR 614.3): Whip of Erebos's reanimated creature is still exiled
// after the Whip itself is gone.
func TestCreatedLeaveBattlefieldReplacementOutlivesSourcePermanent(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	whip := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	affected := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	resolveInstruction(engine, g, &game.StackObject{
		Kind:         game.StackActivatedAbility,
		Controller:   game.Player1,
		SourceID:     whip.ObjectID,
		SourceCardID: whip.CardInstanceID,
		Targets:      []game.Target{game.PermanentTarget(affected.ObjectID)},
	}, game.CreateReplacement{
		Object: game.TargetPermanentReference(0),
		Replacement: &game.ReplacementEffect{
			Duration:      game.DurationPermanent,
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Battlefield,
			ReplaceToZone: zone.Exile,
		},
		Duration: game.DurationPermanent,
	}, nil)

	if !movePermanentToZone(g, whip, zone.Graveyard) {
		t.Fatal("moving the source failed")
	}
	expireCleanupDurations(g)
	if len(g.ReplacementEffects) != 1 {
		t.Fatalf("replacements = %d after the source left, want 1", len(g.ReplacementEffects))
	}
	if !movePermanentToZone(g, affected, zone.Graveyard) {
		t.Fatal("moving the affected creature failed")
	}
	if !g.Players[game.Player1].Exile.Contains(affected.CardInstanceID) {
		t.Fatal("affected creature should be exiled after the source left")
	}
}

// addCardToBattlefieldFromGraveyard reanimates cardID by an unrelated effect.
func addCardToBattlefieldFromGraveyard(t *testing.T, engine *Engine, g *game.Game, cardID id.ID) *game.Permanent {
	t.Helper()
	resolveInstruction(engine, g, &game.StackObject{
		Kind: game.StackActivatedAbility, Controller: game.Player1, SourceID: cardID, SourceCardID: cardID,
	}, game.PutOnBattlefield{Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource})}, nil)
	permanent := unearthedPermanent(g, cardID)
	if permanent == nil {
		t.Fatal("card was not returned to the battlefield")
	}
	return permanent
}

// Source review: Scrapwork Cohort (oracle 149694e6-5cf0-4983-9546-a4bd579b2ade)
// "Unearth {2}{W}". CR 702.84a (September 25, 2026) defines unearth as "Return
// this card from your graveyard to the battlefield. It gains haste. Exile it at
// the beginning of the next end step. If it would leave the battlefield, exile
// it instead ..."; CR 603.7a: a delayed triggered ability exists only once it
// is created during resolution. Ruling 2022-10-14: "If you activate a card's
// unearth ability but that card is removed from your graveyard before the
// ability resolves, that unearth ability will do nothing as it resolves."

func scrapworkCohortUnearthGame(t *testing.T) (*game.Game, *Engine, id.ID) {
	t.Helper()
	registry := cardregistry.NewDefaultRegistry()
	def, err := registry.LookupChecked("Scrapwork Cohort")
	if err != nil || def == nil {
		t.Fatalf("Scrapwork Cohort lookup = %v, %v", def, err)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	cardID := addCardToGraveyard(g, game.Player1, def)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cardID, 0, nil, 0)) {
		t.Fatal("Scrapwork Cohort unearth activation failed")
	}
	return g, engine, cardID
}

// resolveUnearthAndTriggers resolves the unearth ability and any abilities it
// triggers, returning how many triggered abilities were put on the stack.
func resolveUnearthAndTriggers(engine *Engine, g *game.Game) int {
	agents := [game.NumPlayers]PlayerAgent{}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	return drainUnearthTriggers(engine, g)
}

func drainUnearthTriggers(engine *Engine, g *game.Game) int {
	agents := [game.NumPlayers]PlayerAgent{}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	triggered := g.Stack.Size()
	for g.Stack.Size() > 0 {
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
		before := g.Stack.Size()
		engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
		triggered += g.Stack.Size() - before
	}
	return triggered
}

func runUnearthEndStep(engine *Engine, g *game.Game) int {
	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	return drainUnearthTriggers(engine, g)
}

// Exiling Scrapwork Cohort from the graveyard in response makes unearth do
// nothing: no delayed end-step trigger is created and none goes on the stack at
// the next end step (issue #3426).
func TestScrapworkCohortUnearthDoesNothingAfterCardLeftGraveyard(t *testing.T) {
	t.Parallel()
	g, engine, cardID := scrapworkCohortUnearthGame(t)
	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Exile) {
		t.Fatal("exiling Scrapwork Cohort from the graveyard failed")
	}

	if triggered := resolveUnearthAndTriggers(engine, g); triggered != 0 {
		t.Fatalf("unearth resolution triggered %d abilities, want 0", triggered)
	}
	if permanentForCard(g, cardID) != nil {
		t.Fatal("Scrapwork Cohort returned from exile")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers = %d, want 0", len(g.DelayedTriggers))
	}
	if len(g.ReplacementEffects) != 0 {
		t.Fatalf("replacements = %d, want 0", len(g.ReplacementEffects))
	}
	if triggered := runUnearthEndStep(engine, g); triggered != 0 {
		t.Fatalf("end step put %d abilities on the stack, want 0", triggered)
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("Scrapwork Cohort left exile")
	}
}

// A successful Scrapwork Cohort unearth still schedules the end-step exile of
// exactly the returned object, which the next end step exiles.
func TestScrapworkCohortUnearthExilesReturnedObjectAtEndStep(t *testing.T) {
	t.Parallel()
	g, engine, cardID := scrapworkCohortUnearthGame(t)

	if triggered := resolveUnearthAndTriggers(engine, g); triggered != 1 {
		t.Fatalf("unearth resolution triggered %d abilities, want the enters trigger", triggered)
	}

	permanent := permanentForCard(g, cardID)
	if permanent == nil {
		t.Fatal("Scrapwork Cohort did not return to the battlefield")
	}
	if len(g.Battlefield) != 2 {
		t.Fatalf("battlefield size = %d, want Scrapwork Cohort and its Soldier token", len(g.Battlefield))
	}
	if !hasKeyword(g, permanent, game.Haste) {
		t.Fatal("unearthed Scrapwork Cohort lacks haste")
	}
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].CapturedObjectID != permanent.ObjectID {
		t.Fatalf("delayed triggers = %+v, want one capturing %v", g.DelayedTriggers, permanent.ObjectID)
	}
	if len(g.ReplacementEffects) != 1 || g.ReplacementEffects[0].AffectedObjectID != permanent.ObjectID {
		t.Fatalf("replacements = %+v, want one bound to %v", g.ReplacementEffects, permanent.ObjectID)
	}
	if triggered := runUnearthEndStep(engine, g); triggered != 1 {
		t.Fatalf("end step put %d abilities on the stack, want the unearth exile trigger", triggered)
	}
	assertExiledNotElsewhere(t, g, cardID)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers = %d after the end step, want 0", len(g.DelayedTriggers))
	}
}
