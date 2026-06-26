package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestShieldCounterPreventsDamageBeforeMutationAndEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	target.Counters.Add(counter.Shield, 1)

	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, target, 3, false)

	if dealt != 0 {
		t.Fatalf("dealt damage = %d, want 0", dealt)
	}
	if target.MarkedDamage != 0 {
		t.Fatalf("marked damage = %d, want 0", target.MarkedDamage)
	}
	if target.Counters.Get(counter.Shield) != 0 {
		t.Fatalf("shield counters = %d, want 0", target.Counters.Get(counter.Shield))
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.Event) bool {
		return event.SourceID == sourceID &&
			event.PermanentID == target.ObjectID &&
			event.Amount == 3 &&
			event.DamageRecipient == game.DamageRecipientPermanent
	})
	assertNoEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.PermanentID == target.ObjectID
	})
}

func TestShieldCounterReplacesDestroyBeforeZoneChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	target.Counters.Add(counter.Shield, 1)

	removed, ok := destroyPermanent(g, target.ObjectID)

	if ok || removed != nil {
		t.Fatalf("destroyPermanent() = %+v, %v, want nil, false for replaced destroy", removed, ok)
	}
	if _, ok := permanentByObjectID(g, target.ObjectID); !ok {
		t.Fatal("shield-replaced permanent left the battlefield")
	}
	if target.Counters.Get(counter.Shield) != 0 {
		t.Fatalf("shield counters = %d, want 0", target.Counters.Get(counter.Shield))
	}
	if g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("shield-replaced permanent moved to graveyard")
	}
	assertEvent(t, g.Events, game.EventDestroyReplaced, func(event game.Event) bool {
		return event.PermanentID == target.ObjectID &&
			event.FromZone == zone.Battlefield &&
			event.ToZone == zone.Graveyard
	})
	assertNoEvent(t, g.Events, game.EventPermanentDied, func(event game.Event) bool {
		return event.PermanentID == target.ObjectID
	})
}

func TestPreventedCombatDamageDoesNotGrantLifelinkOrMarkDeathtouch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Deathtouch, game.Lifelink)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	blocker.Counters.Add(counter.Shield, 1)
	g.Players[game.Player1].Life = 40
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
		},
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if g.Players[game.Player1].Life != 40 {
		t.Fatalf("lifelink controller life = %d, want 40", g.Players[game.Player1].Life)
	}
	if blocker.MarkedDamage != 0 || blocker.MarkedDeathtouchDamage {
		t.Fatalf("blocker damage = %d deathtouch = %v, want no marked damage", blocker.MarkedDamage, blocker.MarkedDeathtouchDamage)
	}
}

func TestDamageAddendReplacementIncreasesMatchingSourceDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	replacementSource := addReplacementPermanent(t, g, game.Player1, damageAddendReplacementCardDef())
	redSourceID := addColoredSourceCard(g, game.Player1, color.Red)
	blueSourceID := addColoredSourceCard(g, game.Player1, color.Blue)

	if dealt := dealPlayerDamage(g, redSourceID, 0, game.Player1, game.Player2, 2, false); dealt != 3 {
		t.Fatalf("red source damage = %d, want 3", dealt)
	}
	if dealt := dealPlayerDamage(g, blueSourceID, 0, game.Player1, game.Player2, 2, false); dealt != 2 {
		t.Fatalf("blue source damage = %d, want 2", dealt)
	}
	if dealt := dealPlayerDamage(g, redSourceID, 0, game.Player2, game.Player1, 2, false); dealt != 2 {
		t.Fatalf("opponent-controlled red source damage = %d, want 2", dealt)
	}
	if dealt := dealPlayerDamage(g, replacementSource.CardInstanceID, replacementSource.ObjectID, game.Player1, game.Player2, 2, false); dealt != 2 {
		t.Fatalf("replacement source's own red damage = %d, want 2", dealt)
	}
}

func TestDamageMultiplierReplacementDoublesPermanentDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
	sourceID := addColoredSourceCard(g, game.Player1, color.Green)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)

	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, target, 2, false)
	if dealt != 4 {
		t.Fatalf("damage dealt = %d, want 4", dealt)
	}
	if target.MarkedDamage != 4 {
		t.Fatalf("marked damage = %d, want 4", target.MarkedDamage)
	}
}

func TestDamageReplacementEffectsUseFallbackOrderingWhenNoChoiceIsAvailable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, damageAddendReplacementCardDef())
	addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)

	dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 2, false)
	if dealt != 6 {
		t.Fatalf("stacked damage replacements dealt = %d, want deterministic add-then-double result 6", dealt)
	}
	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %d, want 1", len(g.ReplacementDecisions))
	}
	if got := g.ReplacementDecisions[0].Player; got != game.Player2 {
		t.Fatalf("replacement decision player = %v, want damaged player", got)
	}
	decision := g.ReplacementDecisions[0]
	if !decision.UsedFallback || len(decision.Selected) != 1 || decision.Selected[0] != 0 {
		t.Fatalf("replacement decision = %+v, want deterministic fallback (first effect)", decision)
	}
}

func TestSelectedReplacementEffectUsesRecordedOrder(t *testing.T) {
	first := game.ReplacementEffect{ID: 1, Description: "first"}
	second := game.ReplacementEffect{ID: 2, Description: "second"}

	selected := selectedReplacementEffect([]game.ReplacementEffect{first, second}, game.ReplacementDecision{Selected: []int{1, 0}})

	if selected.ID != second.ID {
		t.Fatalf("selected replacement = %+v, want second replacement", selected)
	}
}

func TestDamageReplacementExpiresWhenSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	damageSourceID := addColoredSourceCard(g, game.Player1, color.Red)

	dealt := dealPlayerDamage(g, damageSourceID, 0, game.Player1, game.Player2, 2, false)
	if dealt != 2 {
		t.Fatalf("damage after replacement source leaves = %d, want 2", dealt)
	}
}

func TestDamageReplacementAppliesAfterPrevention(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	g.PreventionShields = append(g.PreventionShields, game.PreventionShield{
		ID:         g.IDGen.Next(),
		Controller: game.Player2,
		Player:     game.Player2,
		Amount:     1,
		Duration:   game.DurationUntilEndOfTurn,
	})

	dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 3, false)
	if dealt != 4 {
		t.Fatalf("damage after prevention then replacement = %d, want 4", dealt)
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.Event) bool {
		return event.Player == game.Player2 && event.Amount == 1
	})
}

func TestPreventionShieldPreventsTrackedAmountAndExpires(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	obj := &game.StackObject{
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}

	resolveInstruction(engine, g, obj, game.PreventDamage{Amount: game.Fixed(2), Object: game.TargetPermanentReference(0)}, nil)
	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, target, 5, false)

	if dealt != 3 {
		t.Fatalf("dealt damage = %d, want 3 after prevention shield", dealt)
	}
	if target.MarkedDamage != 3 {
		t.Fatalf("marked damage = %d, want 3", target.MarkedDamage)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields = %+v, want consumed", g.PreventionShields)
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.Event) bool {
		return event.PermanentID == target.ObjectID && event.Amount == 2
	})

	resolveInstruction(engine, g, obj, game.PreventDamage{Amount: game.Fixed(1), Object: game.TargetPermanentReference(0)}, nil)
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields after cleanup = %+v, want expired", g.PreventionShields)
	}
}

func TestMultiplePreventionShieldsRecordDeterministicReplacementOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	obj := &game.StackObject{
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	resolveInstruction(engine, g, obj, game.PreventDamage{Amount: game.Fixed(1), Object: game.TargetPermanentReference(0)}, nil)
	resolveInstruction(engine, g, obj, game.PreventDamage{Amount: game.Fixed(1), Object: game.TargetPermanentReference(0)}, nil)

	dealPermanentDamage(g, sourceID, 0, game.Player1, target, 3, false)

	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %+v, want one deterministic prevention order", g.ReplacementDecisions)
	}
	decision := g.ReplacementDecisions[0]
	if decision.Player != game.Player2 || !decision.UsedFallback || len(decision.Selected) != 2 || decision.Selected[0] != 0 || decision.Selected[1] != 1 {
		t.Fatalf("replacement decision = %+v, want Player2 fallback order [0 1]", decision)
	}
}

func TestRegenerationReplacesDestroyAndRemovesFromCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	blocker.MarkedDamage = 2
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}},
		BlockerOrder: map[id.ID][]id.ID{
			attacker.ObjectID: []id.ID{blocker.ObjectID},
		},
	}

	resolveInstruction(engine, g, &game.StackObject{
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(blocker.ObjectID)},
	}, game.Regenerate{Object: game.TargetPermanentReference(0)}, nil)
	removed, ok := destroyPermanent(g, blocker.ObjectID)

	if ok || removed != nil {
		t.Fatalf("destroyPermanent() = %+v, %v, want regenerated replacement", removed, ok)
	}
	if _, ok := permanentByObjectID(g, blocker.ObjectID); !ok {
		t.Fatal("regenerated blocker left battlefield")
	}
	if !blocker.Tapped || blocker.MarkedDamage != 0 || blocker.RegenerationShields != 0 {
		t.Fatalf("regenerated blocker tapped=%v damage=%d shields=%d, want tapped, no damage, no shields", blocker.Tapped, blocker.MarkedDamage, blocker.RegenerationShields)
	}
	if len(g.Combat.Blockers) != 0 || len(g.Combat.BlockerOrder[attacker.ObjectID]) != 0 {
		t.Fatalf("combat after regeneration blockers=%+v order=%+v, want blocker removed", g.Combat.Blockers, g.Combat.BlockerOrder)
	}
	assertEvent(t, g.Events, game.EventDestroyReplaced, func(event game.Event) bool {
		return event.PermanentID == blocker.ObjectID
	})
}

func TestRegenerationShieldExpiresDuringCleanup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.RegenerationShields = 1

	NewEngine(nil).runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if creature.RegenerationShields != 0 {
		t.Fatalf("regeneration shields = %d, want cleanup expiry", creature.RegenerationShields)
	}
}

// replacementChoosingAgent picks the replacement/prevention effect whose label
// contains a preferred substring when asked a CR 616.1 selection, and otherwise
// defers. It passes on all priority actions.
type replacementChoosingAgent struct {
	prefer string
}

func (replacementChoosingAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a replacementChoosingAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceReplacement {
		for _, option := range request.Options {
			if strings.Contains(strings.ToLower(option.Label), a.prefer) {
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

// TestDamageReplacementSelectionHonorsPlayerChoice covers CR 616.1: when two
// replacement effects apply to the same damage event, the affected player chooses
// which to apply first, and that choice changes the result. The fallback order
// (add 1, then double) yields (2+1)*2 = 6; choosing to double first yields
// (2*2)+1 = 5.
func TestDamageReplacementSelectionHonorsPlayerChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, damageAddendReplacementCardDef())
	addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)

	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{
		// The damaged player (CR 616.1 chooser) elects to double the damage first.
		game.Player2: replacementChoosingAgent{prefer: "double"},
	}
	engine.setReplacementChoiceContext(g, agents, &TurnLog{})
	defer g.ClearChoiceContext()

	dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 2, false)

	if dealt != 5 {
		t.Fatalf("damage dealt = %d, want 5 (double-then-add chosen by the affected player)", dealt)
	}
	if len(g.ReplacementDecisions) != 1 || g.ReplacementDecisions[0].UsedFallback {
		t.Fatalf("replacement decisions = %+v, want one agent-made decision", g.ReplacementDecisions)
	}
}

// TestDamageReplacementFallbackRecordedWhenChooserHasNoChoiceAgent covers the
// UsedFallback metadata: when a choice context is present but the chooser's agent
// can't answer a CR 616.1 selection (it doesn't implement ChoiceAgent), the engine
// falls back to the first effect and records the decision as a fallback.
func TestDamageReplacementFallbackRecordedWhenChooserHasNoChoiceAgent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, damageAddendReplacementCardDef())
	addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)

	engine := NewEngine(nil)
	// firstLegalAgent implements ChooseAction but not ChoiceAgent, so the CR 616.1
	// selection falls back to the first match.
	agents := [game.NumPlayers]PlayerAgent{game.Player2: firstLegalAgent{}}
	engine.setReplacementChoiceContext(g, agents, &TurnLog{})
	defer g.ClearChoiceContext()

	dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 2, false)

	if dealt != 6 {
		t.Fatalf("damage dealt = %d, want fallback add-then-double 6", dealt)
	}
	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %d, want 1", len(g.ReplacementDecisions))
	}
	decision := g.ReplacementDecisions[0]
	if !decision.UsedFallback || len(decision.Selected) != 1 || decision.Selected[0] != 0 {
		t.Fatalf("replacement decision = %+v, want fallback first-effect", decision)
	}
}

func TestShieldAndRegenerationReplacementOrderIsRecorded(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Counters.Add(counter.Shield, 1)
	creature.RegenerationShields = 1

	destroyPermanent(g, creature.ObjectID)

	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %+v, want one shield/regeneration order", g.ReplacementDecisions)
	}
	decision := g.ReplacementDecisions[0]
	if decision.Player != game.Player1 || !decision.UsedFallback || len(decision.Selected) != 1 {
		t.Fatalf("replacement decision = %+v, want Player1 fallback (first effect)", decision)
	}
	if creature.Counters.Get(counter.Shield) != 0 || creature.RegenerationShields != 1 {
		t.Fatalf("shield counters=%d regeneration=%d, want shield used before regeneration", creature.Counters.Get(counter.Shield), creature.RegenerationShields)
	}
}

func TestRegenerationReplacesLethalDamageSBA(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.MarkedDamage = 2
	creature.RegenerationShields = 1

	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if len(deaths) != 0 {
		t.Fatalf("deaths = %+v, want regeneration to replace lethal-damage destruction", deaths)
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok || !creature.Tapped || creature.MarkedDamage != 0 {
		t.Fatalf("creature after regeneration = %+v, want tapped on battlefield with no damage", creature)
	}
}

func damageAddendReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:   "Embermaw Hellion",
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Red},
		ReplacementAbilities: []game.ReplacementAbility{
			game.DamageReplacementExcludingSource(
				"If another red source you control would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.",
				0,
				1,
				[]color.Color{color.Red},
				game.TriggerControllerYou,
			),
		},
	}}
}

func damageMultiplierReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Angrath's Marauders",
		Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.DamageReplacement(
				"If a source you control would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.",
				2,
				0,
				nil,
				game.TriggerControllerYou,
			),
		},
	}}
}
