package rules

import (
	"testing"

	cardc "github.com/natefinch/council4/mtg/cards/c"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// conduitCardDef loads the registered Conduit of Worlds definition and asserts
// the two-ability shape its Oracle text lowers to: the shared
// play-lands-from-graveyard static and the tap-to-cast, sorcery-speed activated
// ability. Sourcing behavior from the real generated definition proves the
// curated card — not a hand-written stand-in — drives the runtime.
func conduitCardDef(t *testing.T) *game.CardDef {
	t.Helper()
	def := cardc.ConduitOfWorlds()
	if got := len(def.StaticAbilities); got != 1 {
		t.Fatalf("Conduit of Worlds has %d static abilities, want 1", got)
	}
	if got := len(def.ActivatedAbilities); got != 1 {
		t.Fatalf("Conduit of Worlds has %d activated abilities, want 1", got)
	}
	ability := def.ActivatedAbilities[0]
	if len(ability.AdditionalCosts) != 1 || ability.AdditionalCosts[0].Kind != cost.AdditionalTap {
		t.Fatalf("activation cost = %+v, want a single tap", ability.AdditionalCosts)
	}
	if ability.Timing != game.SorceryOnly {
		t.Fatalf("activation timing = %v, want SorceryOnly", ability.Timing)
	}
	return def
}

// addConduit puts the real Conduit of Worlds artifact onto the battlefield under
// the given controller so its registered static and activated abilities are live.
func addConduit(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, cardc.ConduitOfWorlds())
}

// graveyardBear is a plain {1} creature card used as the nonland permanent card
// Conduit targets and casts from a graveyard.
func graveyardBear() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Graveyard Bear",
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		Types:    []types.Card{types.Creature},
	}}
}

// setUpConduitMainPhase configures Player1's precombat main phase with priority,
// the sorcery-speed timing in which Conduit's ability is activatable.
func setUpConduitMainPhase(g *game.Game) {
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
}

// resolveConduitWith activates Conduit targeting cardID in Player1's graveyard,
// resolves the ability with the given optional-cast decision, and returns the
// activation success and the source permanent. Mana is added to Player1's pool so
// the {1} cast can be paid at resolution when accepted.
func activateConduitTargeting(t *testing.T, g *game.Game, engine *Engine, source *game.Permanent, cardID id.ID) bool {
	t.Helper()
	act := action.ActivateAbility(source.ObjectID, 0, []game.Target{currentCardTarget(t, g, cardID)}, 0)
	return engine.applyAction(g, game.Player1, act)
}

func TestConduitOfWorldsCardDefShape(t *testing.T) {
	t.Parallel()
	def := conduitCardDef(t)
	mode := def.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard {
		t.Fatalf("target allow=%v zone=%v, want a graveyard card", target.Allow, target.TargetZone)
	}
	if target.Selection.Val.Controller != game.ControllerYou {
		t.Fatalf("target controller = %v, want ControllerYou", target.Selection.Val.Controller)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("resolution sequence = %d instructions, want 2 (cast, lock)", len(mode.Sequence))
	}
}

// TestConduitPlaysLandsFromGraveyard proves the static ability lets Conduit's
// controller — and only its controller — play lands from their graveyard.
func TestConduitPlaysLandsFromGraveyard(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addConduit(g, game.Player1)
	landID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})
	setUpConduitMainPhase(g)

	if !canPlayLandFromZoneByRuleEffect(g, game.Player1, landID, zone.Graveyard) {
		t.Fatal("Conduit's controller cannot play a land from their graveyard")
	}
	if canPlayLandFromZoneByRuleEffect(g, game.Player2, landID, zone.Graveyard) {
		t.Fatal("the opponent may play a land from Player1's graveyard")
	}
	if !engine.applyAction(g, game.Player1, action.PlayLandFaceFromZone(landID, zone.Graveyard, game.FaceFront)) {
		t.Fatal("playing a land from the graveyard was rejected despite Conduit's static")
	}
	if g.Players[game.Player1].Graveyard.Contains(landID) {
		t.Fatal("land remained in the graveyard after being played")
	}
}

// TestConduitActivationIsSorcerySpeedAndTaps proves the activated ability is
// legal only at sorcery speed and taps Conduit as its cost.
func TestConduitActivationIsSorcerySpeedAndTaps(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addConduit(g, game.Player1)
	bearID := addCardToGraveyard(g, game.Player1, graveyardBear())
	setUpConduitMainPhase(g)

	act := action.ActivateAbility(source.ObjectID, 0, []game.Target{currentCardTarget(t, g, bearID)}, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Conduit's ability is not legal at sorcery speed with an empty stack")
	}

	// A non-empty stack means it is no longer the player's main phase with an
	// empty stack, so the sorcery-speed ability is not activatable.
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player2})
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Conduit's ability is legal while the stack is non-empty (should be sorcery speed)")
	}
	g.Stack.Pop()

	if !activateConduitTargeting(t, g, engine, source, bearID) {
		t.Fatal("activating Conduit failed")
	}
	if !source.Tapped {
		t.Fatal("Conduit was not tapped by its activation cost")
	}
}

// TestConduitTargetsNonlandPermanentCardInOwnGraveyardOnly proves the target
// restriction: a nonland permanent card in the controller's own graveyard is a
// legal target, while a land card, a card in an opponent's graveyard, and a
// noncreature nonpermanent (instant) card are all illegal.
func TestConduitTargetsNonlandPermanentCardInOwnGraveyardOnly(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addConduit(g, game.Player1)
	setUpConduitMainPhase(g)

	legalBear := addCardToGraveyard(g, game.Player1, graveyardBear())
	ownLand := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Own Forest", Types: []types.Card{types.Land}}})
	ownInstant := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Own Bolt", Types: []types.Card{types.Instant}}})
	opponentBear := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Opponent Bear", Types: []types.Card{types.Creature}}})

	legalTarget := func(cardID id.ID) bool {
		act := action.ActivateAbility(source.ObjectID, 0, []game.Target{currentCardTarget(t, g, cardID)}, 0)
		return containsAction(engine.legalActions(g, game.Player1), act)
	}

	if !legalTarget(legalBear) {
		t.Fatal("a nonland permanent card in the controller's graveyard is not a legal target")
	}
	if legalTarget(ownLand) {
		t.Fatal("a land card is a legal target (should be nonland only)")
	}
	if legalTarget(ownInstant) {
		t.Fatal("an instant card is a legal target (should be a permanent card only)")
	}
	if legalTarget(opponentBear) {
		t.Fatal("a card in the opponent's graveyard is a legal target (should be your graveyard only)")
	}
}

// TestConduitSuccessfulPaidCastAndLock proves the successful path end to end:
// activating and accepting the optional cast pays for the targeted graveyard
// creature, puts it on the stack under the controller's control, and — because a
// spell was cast — forbids the controller (but not the opponent) from casting
// further spells this turn.
func TestConduitSuccessfulPaidCastAndLock(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addConduit(g, game.Player1)
	bearID := addCardToGraveyard(g, game.Player1, graveyardBear())
	setUpConduitMainPhase(g)

	if !activateConduitTargeting(t, g, engine, source, bearID) {
		t.Fatal("activating Conduit failed")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(bearID) {
		t.Fatal("targeted creature still in graveyard after an accepted paid cast")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != bearID {
		t.Fatalf("stack top = %#v, want the cast graveyard creature %v", top, bearID)
	}
	if top.Controller != game.Player1 {
		t.Fatalf("cast spell controller = %v, want Player1", top.Controller)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana pool = %d after paying {1}, want 0", got)
	}
	if !spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("controller can still cast spells; the can't-cast-additional lock was not applied")
	}
	if spellCastProhibited(g, game.Player2, vanillaCreatureDef()) {
		t.Fatal("the self lock wrongly restricts the opponent")
	}
}

// TestConduitDeclineCastsNothingNoLock proves declining the optional cast casts
// nothing and, because no spell was cast, applies no restriction.
func TestConduitDeclineCastsNothingNoLock(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addConduit(g, game.Player1)
	bearID := addCardToGraveyard(g, game.Player1, graveyardBear())
	setUpConduitMainPhase(g)

	if !activateConduitTargeting(t, g, engine, source, bearID) {
		t.Fatal("activating Conduit failed")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(bearID) {
		t.Fatal("targeted creature left the graveyard despite declining the cast")
	}
	if spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("the can't-cast lock was applied without a cast")
	}
}

// TestConduitPriorSpellThisTurnSkipsCast proves the resolution-time condition: if
// the controller already cast a spell this turn, activation is still legal but
// the ability's paid cast is skipped and no restriction is applied.
func TestConduitPriorSpellThisTurnSkipsCast(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addConduit(g, game.Player1)
	bearID := addCardToGraveyard(g, game.Player1, graveyardBear())
	setUpConduitMainPhase(g)
	g.AppendEvent(game.Event{Kind: game.EventSpellCast, Controller: game.Player1})

	if !activateConduitTargeting(t, g, engine, source, bearID) {
		t.Fatal("activation was rejected after a prior spell; the condition is a resolution gate, not an activation restriction")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(bearID) {
		t.Fatal("targeted creature was cast despite a prior spell this turn")
	}
	if spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("the can't-cast lock was applied despite the condition failing")
	}
}

// TestConduitOpponentSpellThisTurnStillCasts proves the resolution-time
// condition is scoped to the controller's own casts: an opponent having cast a
// spell this turn does not satisfy "you haven't cast a spell this turn", so the
// paid cast still proceeds and the lock is applied. This distinguishes the
// per-player history gate from a global "any spell cast this turn" reading.
func TestConduitOpponentSpellThisTurnStillCasts(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addConduit(g, game.Player1)
	bearID := addCardToGraveyard(g, game.Player1, graveyardBear())
	setUpConduitMainPhase(g)
	g.AppendEvent(game.Event{Kind: game.EventSpellCast, Controller: game.Player2})

	if !activateConduitTargeting(t, g, engine, source, bearID) {
		t.Fatal("activating Conduit failed")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(bearID) {
		t.Fatal("an opponent's spell this turn wrongly blocked the controller's cast")
	}
	if !spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("the can't-cast lock was not applied after a successful cast")
	}
}
