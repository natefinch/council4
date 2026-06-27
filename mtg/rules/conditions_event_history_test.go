package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestEventHistoryConditionCurrentTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source1 := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})
	source2 := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Bear",
		Types: []types.Card{types.Creature},
	}})

	attackedCond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:      game.EventAttackerDeclared,
				Controller: game.TriggerControllerYou,
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	})

	ctx1 := conditionContext{controller: game.Player1, source: source1}
	ctx2 := conditionContext{controller: game.Player2, source: source2}
	if conditionSatisfied(g, ctx1, attackedCond) {
		t.Fatal("condition satisfied before any attacks")
	}

	// Player1 attacks — satisfied for Player1's source but not Player2's.
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1})

	if !conditionSatisfied(g, ctx1, attackedCond) {
		t.Fatal("condition not satisfied after Player1 attacked")
	}
	if conditionSatisfied(g, ctx2, attackedCond) {
		t.Fatal("condition satisfied for Player2 source when only Player1 attacked")
	}
}

func TestEventHistoryConditionAttackerCount(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})

	twoAttackersCond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:      game.EventAttackerDeclared,
				Controller: game.TriggerControllerYou,
			},
			Window:   game.EventHistoryCurrentTurn,
			MinCount: 2,
		}),
	})

	ctx := conditionContext{controller: game.Player1, source: source}

	// One attacker you control is not enough for "two or more creatures".
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1})
	if conditionSatisfied(g, ctx, twoAttackersCond) {
		t.Fatal("condition satisfied after only one attacker")
	}

	// An opponent's attacker does not count toward your controller-scoped total.
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player2})
	if conditionSatisfied(g, ctx, twoAttackersCond) {
		t.Fatal("condition satisfied counting an opponent's attacker")
	}

	// A second attacker you control reaches the required count.
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1})
	if !conditionSatisfied(g, ctx, twoAttackersCond) {
		t.Fatal("condition not satisfied after attacking with two creatures")
	}
}

func TestEventHistoryConditionPreviousTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})

	lifeLostCond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:  game.EventLifeLost,
				Player: game.TriggerPlayerOpponent,
			},
			Window: game.EventHistoryPreviousTurn,
		}),
	})

	ctx := conditionContext{controller: game.Player1, source: source}
	if conditionSatisfied(g, ctx, lifeLostCond) {
		t.Fatal("condition satisfied before any events")
	}

	// Emit life-loss on this turn then advance to the next turn so it becomes
	// "previous turn" for the upkeep check.
	emitEvent(g, game.Event{Kind: game.EventLifeLost, Player: game.Player2, Amount: 3})
	g.EventTurnStarts = append(g.EventTurnStarts, len(g.Events))
	g.Turn.TurnNumber++

	if !conditionSatisfied(g, ctx, lifeLostCond) {
		t.Fatal("condition not satisfied after opponent lost life last turn")
	}
}

func TestEventHistoryConditionNegatedNoSpells(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})

	noSpellsCond := opt.Val(game.Condition{
		Negate: true,
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{Event: game.EventSpellCast},
			Window:  game.EventHistoryPreviousTurn,
		}),
	})

	ctx := conditionContext{controller: game.Player1, source: source}

	// No previous turn yet — EventsPreviousTurn returns nil, no spells found →
	// condition (negated) is satisfied.
	if !conditionSatisfied(g, ctx, noSpellsCond) {
		t.Fatal("negated condition not satisfied when no previous-turn events exist")
	}

	// Emit a spell cast on "last turn" then advance.
	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player1})
	g.EventTurnStarts = append(g.EventTurnStarts, len(g.Events))
	g.Turn.TurnNumber++

	if conditionSatisfied(g, ctx, noSpellsCond) {
		t.Fatal("negated condition satisfied when a spell was cast last turn")
	}
}

func TestEventHistoryConditionCreatureDiedCurrentTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})

	creatureDiedCond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:            game.EventPermanentDied,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	})

	ctx := conditionContext{controller: game.Player1, source: source}
	if conditionSatisfied(g, ctx, creatureDiedCond) {
		t.Fatal("condition satisfied before any creature deaths")
	}

	// A non-creature permanent died — should not satisfy.
	addAndEmitArtifactDied(g)
	if conditionSatisfied(g, ctx, creatureDiedCond) {
		t.Fatal("condition satisfied after non-creature died")
	}

	// A creature dies — now satisfied.
	emitCreatureDiedEvent(g)
	if !conditionSatisfied(g, ctx, creatureDiedCond) {
		t.Fatal("condition not satisfied after creature died")
	}
}

func TestEventHistoryConditionFailsClosedWithNilSource(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1})

	cond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:      game.EventAttackerDeclared,
				Controller: game.TriggerControllerYou,
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	})
	if conditionSatisfied(g, conditionContext{controller: game.Player1, source: nil}, cond) {
		t.Fatal("condition satisfied with nil source; should fail closed")
	}
}

// addAndEmitArtifactDied registers and emits an EventPermanentDied for a
// non-creature artifact so tests can verify creature-type filtering.
func addAndEmitArtifactDied(g *game.Game) {
	perm := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Dead Relic",
		Types: []types.Card{types.Artifact},
	}})
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentDied,
		PermanentID: perm.ObjectID,
		CardTypes:   []types.Card{types.Artifact},
	})
}

// emitCreatureDiedEvent emits an EventPermanentDied that looks like a creature
// died by recording creature card types on the event directly.
func emitCreatureDiedEvent(g *game.Game) {
	perm := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Dead Bear",
		Types: []types.Card{types.Creature},
	}})
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentDied,
		PermanentID: perm.ObjectID,
		CardTypes:   []types.Card{types.Creature},
	})
}

func TestEventHistoryConditionRevoltLeftBattlefield(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source1 := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})
	source2 := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Bear",
		Types: []types.Card{types.Creature},
	}})

	// The pattern produced by the Revolt event-history recognizer: any permanent
	// you control leaving the battlefield this turn.
	revoltCond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:         game.EventZoneChanged,
				Controller:    game.TriggerControllerYou,
				MatchFromZone: true,
				FromZone:      zone.Battlefield,
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	})

	ctx1 := conditionContext{controller: game.Player1, source: source1}
	ctx2 := conditionContext{controller: game.Player2, source: source2}
	if conditionSatisfied(g, ctx1, revoltCond) {
		t.Fatal("condition satisfied before any permanent left the battlefield")
	}

	// A non-battlefield departure (entering from the hand) must not satisfy revolt.
	emitEvent(g, game.Event{Kind: game.EventZoneChanged, Controller: game.Player1, PermanentID: source1.ObjectID, FromZone: zone.Hand, ToZone: zone.Battlefield})
	if conditionSatisfied(g, ctx1, revoltCond) {
		t.Fatal("condition satisfied after a non-battlefield zone change")
	}

	// An opponent's permanent leaving the battlefield satisfies the opponent's
	// source but not the controller-scoped Player1 source.
	emitEvent(g, game.Event{Kind: game.EventZoneChanged, Controller: game.Player2, PermanentID: source2.ObjectID, FromZone: zone.Battlefield, ToZone: zone.Exile})
	if conditionSatisfied(g, ctx1, revoltCond) {
		t.Fatal("condition satisfied for Player1 when only Player2's permanent left the battlefield")
	}
	if !conditionSatisfied(g, ctx2, revoltCond) {
		t.Fatal("condition not satisfied for Player2 after their permanent left the battlefield")
	}

	// A permanent the controller owns leaving the battlefield to any zone
	// (here exile rather than the graveyard) satisfies revolt.
	emitEvent(g, game.Event{Kind: game.EventZoneChanged, Controller: game.Player1, PermanentID: source1.ObjectID, FromZone: zone.Battlefield, ToZone: zone.Exile})
	if !conditionSatisfied(g, ctx1, revoltCond) {
		t.Fatal("condition not satisfied after the controller's permanent left the battlefield")
	}
}

// descendCond is the descend intervening-if condition "if you descended this
// turn": a current-turn zone change moving a nontoken permanent card into the
// controller-owned graveyard from anywhere (CR 701.51).
func descendCond() opt.V[game.Condition] {
	return opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:       game.EventZoneChanged,
				Player:      game.TriggerPlayerYou,
				MatchToZone: true,
				ToZone:      zone.Graveyard,
				SubjectSelection: game.Selection{
					RequiredTypesAny: []types.Card{
						types.Artifact,
						types.Battle,
						types.Creature,
						types.Enchantment,
						types.Land,
						types.Planeswalker,
					},
					NonToken: true,
				},
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	})
}

func TestEventHistoryDescendPermanentCard(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Source",
		Types: []types.Card{types.Creature},
	}})
	source2 := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Source",
		Types: []types.Card{types.Creature},
	}})
	ctx1 := conditionContext{controller: game.Player1, source: source}
	ctx2 := conditionContext{controller: game.Player2, source: source2}
	cond := descendCond()

	if conditionSatisfied(g, ctx1, cond) {
		t.Fatal("descend satisfied before any permanent card entered the graveyard")
	}

	// A permanent card Player1 owns put into their graveyard from the library
	// (a mill) satisfies descend for Player1 but not Player2.
	creatureID := g.IDGen.Next()
	g.CardInstances[creatureID] = &game.CardInstance{
		ID:    creatureID,
		Owner: game.Player1,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Milled Bear",
			Types: []types.Card{types.Creature},
		}},
	}
	emitEvent(g, game.Event{
		Kind:     game.EventZoneChanged,
		Player:   game.Player1,
		CardID:   creatureID,
		FromZone: zone.Library,
		ToZone:   zone.Graveyard,
	})
	if !conditionSatisfied(g, ctx1, cond) {
		t.Fatal("descend not satisfied after a permanent card entered Player1's graveyard")
	}
	if conditionSatisfied(g, ctx2, cond) {
		t.Fatal("descend satisfied for Player2 when only Player1's graveyard received a permanent card")
	}
}

func TestEventHistoryDescendIgnoresNonpermanentCard(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Source",
		Types: []types.Card{types.Creature},
	}})
	ctx := conditionContext{controller: game.Player1, source: source}
	cond := descendCond()

	// An instant card going to the graveyard is not a permanent card, so it
	// must not satisfy descend.
	instantID := g.IDGen.Next()
	g.CardInstances[instantID] = &game.CardInstance{
		ID:    instantID,
		Owner: game.Player1,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Discarded Bolt",
			Types: []types.Card{types.Instant},
		}},
	}
	emitEvent(g, game.Event{
		Kind:     game.EventZoneChanged,
		Player:   game.Player1,
		CardID:   instantID,
		FromZone: zone.Hand,
		ToZone:   zone.Graveyard,
	})
	if conditionSatisfied(g, ctx, cond) {
		t.Fatal("descend satisfied after a nonpermanent card entered the graveyard")
	}

	// A permanent card moving somewhere other than the graveyard also does not
	// satisfy descend.
	landID := g.IDGen.Next()
	g.CardInstances[landID] = &game.CardInstance{
		ID:    landID,
		Owner: game.Player1,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Exiled Island",
			Types: []types.Card{types.Land},
		}},
	}
	emitEvent(g, game.Event{
		Kind:     game.EventZoneChanged,
		Player:   game.Player1,
		CardID:   landID,
		FromZone: zone.Hand,
		ToZone:   zone.Exile,
	})
	if conditionSatisfied(g, ctx, cond) {
		t.Fatal("descend satisfied after a permanent card moved to a zone other than the graveyard")
	}
}
