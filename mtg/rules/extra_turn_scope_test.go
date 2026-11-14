package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Alchemist's Gambit (oracle 271d8db6-8606-46d7-9812-75aa22fcfd8e,
// pinned corpus) reads "Cleave {4}{U}{U}{R} ... Take an extra turn after this
// one. During that turn, damage can't be prevented. [At the beginning of that
// turn's end step, you lose the game.] Exile Alchemist's Gambit." Scryfall WotC
// rulings (2021-11-19): if several extra turns are created, the most recently
// created is taken first; if the extra turn or its end step is skipped, the
// delayed triggered ability never triggers; cleave removes the bracketed text,
// so no delayed trigger is created. CR 500.7 (verified official copy) adds an
// extra turn directly after the current turn, and the most recently created
// extra turn is taken first. These scenarios exercise the shared operation,
// game.TakeExtraTurn, not a registered card.

const (
	extraTurnLoseText   = "At the beginning of that turn's end step, you lose the game."
	extraTurnProbeDealt = 1
)

var extraTurnGambitProgram = game.ProgramReference{ID: "test/extra-turn-gambit", Version: "1"}

// gambitExtraTurn is the "Take an extra turn after this one. During that turn,
// damage can't be prevented." grant; withTrigger adds the end-step loss that
// cleave removes.
func gambitExtraTurn(withTrigger bool) game.TakeExtraTurn {
	grant := game.TakeExtraTurn{
		Player:      game.ControllerReference(),
		RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectDamageCantBePrevented}},
	}
	if withTrigger {
		grant.EndStepTriggers = []game.DelayedTriggerDef{{
			Text: extraTurnLoseText,
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.PlayerLosesGame{Player: game.ControllerReference()},
			}}}.Ability(),
		}}
	}
	return grant
}

// extraTurnEngine returns an Engine whose authored test program resolves the
// gambit grant through cardprogram.Context.Execute, the authored-card route.
func extraTurnEngine(t *testing.T) *Engine {
	t.Helper()
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: extraTurnGambitProgram,
		Resolve: func(ctx cardprogram.Context) error {
			_, err := ctx.Execute(gambitExtraTurn(true))
			return err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	return NewEngineWithPrograms(nil, catalog)
}

// castExtraTurnSorcery has player really cast and resolve a free sorcery with
// the given spell ability during their main phase.
func castExtraTurnSorcery(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID, name string, ability game.AbilityContent) {
	t.Helper()
	card := addCardToHand(g, player, &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Sorcery},
		ManaCost:     opt.Val(cost.Mana{cost.O(0)}),
		SpellAbility: opt.Val(ability),
	}})
	setMainPhasePriority(g, player)
	if !engine.applyAction(g, player, action.CastSpell(card, nil, 0, nil)) {
		t.Fatalf("%s could not be cast", name)
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if !g.Players[player].Graveyard.Contains(card) {
		t.Fatalf("%s did not resolve", name)
	}
}

// extraTurnDamageDealt sets up "prevent all damage" and has a Player2 creature
// deal 1 damage to Player4, returning how much was dealt.
func extraTurnDamageDealt(engine *Engine, g *game.Game, source *game.Permanent) int {
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3}, game.PreventDamage{All: true, Global: true}, nil)
	return dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player2, game.Player4, extraTurnProbeDealt, false)
}

// extraTurnScopeCount counts the rule effects and delayed triggers bound to an
// extra turn.
func extraTurnScopeCount(g *game.Game) (effects, triggers int) {
	for i := range g.RuleEffects {
		if g.RuleEffects[i].ExtraTurnID != 0 {
			effects++
		}
	}
	for i := range g.DelayedTriggers {
		if g.DelayedTriggers[i].ExtraTurnID != 0 {
			triggers++
		}
	}
	return effects, triggers
}

// TestTakeExtraTurnScopesEffectsAndEndStepTriggerToThatTurn casts the gambit
// grant through an authored program and plays the turns through the Engine:
// damage stays preventable for the rest of the casting turn (including its end
// step, where the trigger must wait), can't be prevented during the extra turn,
// the controller loses at the extra turn's end step, and damage is preventable
// again afterward.
func TestTakeExtraTurnScopesEffectsAndEndStepTriggerToThatTurn(t *testing.T) {
	g := extraTurnGame()
	engine := extraTurnEngine(t)
	source := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	agents := [game.NumPlayers]PlayerAgent{}

	castExtraTurnSorcery(t, g, engine, game.Player1, "Authored extra-turn gambit",
		game.Mode{Program: extraTurnGambitProgram}.Ability())
	if len(g.Turn.ExtraTurns) != 1 || g.Turn.ExtraTurns[0].Player != game.Player1 || g.Turn.ExtraTurns[0].ID == 0 {
		t.Fatalf("extra turns = %+v, want one identified Player1 extra turn", g.Turn.ExtraTurns)
	}
	if effects, triggers := extraTurnScopeCount(g); effects != 1 || triggers != 1 {
		t.Fatalf("scoped effects/triggers = %d/%d, want 1/1", effects, triggers)
	}
	if text := g.DelayedTriggers[len(g.DelayedTriggers)-1].Ability.Text; text != extraTurnLoseText {
		t.Fatalf("delayed trigger text = %q, want %q", text, extraTurnLoseText)
	}
	if dealt := extraTurnDamageDealt(engine, g, source); dealt != 0 {
		t.Fatalf("damage during the casting turn = %d, want prevented", dealt)
	}

	castingTurn := g.Turn.TurnNumber
	engine.finishCurrentTurn(g, agents)
	if g.IsGameOver() || !isPlayerAlive(g, game.Player1) {
		t.Fatal("the casting turn's end step fired the extra turn's end-step trigger")
	}
	if g.Turn.ActivePlayer != game.Player1 || g.Turn.TurnNumber != castingTurn+1 || g.Turn.ExtraTurnID == 0 {
		t.Fatalf("next turn = player %v turn %d extra %v, want Player1's extra turn", g.Turn.ActivePlayer, g.Turn.TurnNumber, g.Turn.ExtraTurnID)
	}
	if dealt := extraTurnDamageDealt(engine, g, source); dealt != extraTurnProbeDealt {
		t.Fatalf("damage during the extra turn = %d, want %d unpreventable", dealt, extraTurnProbeDealt)
	}

	clone := g.Clone()
	engine.runTurn(g, agents)
	if isPlayerAlive(g, game.Player1) {
		t.Fatal("Player1 did not lose at the beginning of the extra turn's end step")
	}
	if g.Turn.ActivePlayer != game.Player2 || g.Turn.ExtraTurnID != 0 {
		t.Fatalf("turn after the extra turn = player %v extra %v, want Player2's regular turn", g.Turn.ActivePlayer, g.Turn.ExtraTurnID)
	}
	if effects, triggers := extraTurnScopeCount(g); effects != 0 || triggers != 0 {
		t.Fatalf("scoped effects/triggers after the extra turn = %d/%d, want none", effects, triggers)
	}
	if dealt := extraTurnDamageDealt(engine, g, source); dealt != 0 {
		t.Fatalf("damage after the extra turn = %d, want prevented", dealt)
	}

	if !isPlayerAlive(clone, game.Player1) || clone.Turn.ExtraTurnID == 0 {
		t.Fatal("playing the original's extra turn changed the clone")
	}
	if effects, triggers := extraTurnScopeCount(clone); effects != 1 || triggers != 1 ||
		clone.RuleEffects[len(clone.RuleEffects)-1].ExtraTurnID != clone.Turn.ExtraTurnID ||
		clone.DelayedTriggers[len(clone.DelayedTriggers)-1].ExtraTurnID != clone.Turn.ExtraTurnID {
		t.Fatalf("clone scoped effects/triggers = %d/%d, want 1/1 bound to its current extra turn", effects, triggers)
	}
}

// TestTakeExtraTurnBindsToItsOwnQueuedTurn queues the gambit grant and then a
// plain extra turn for the same player. The plain turn, created last, is taken
// first (CR 500.7) and gets neither the unpreventable damage nor the end-step
// loss; both apply to the gambit's own turn after it.
func TestTakeExtraTurnBindsToItsOwnQueuedTurn(t *testing.T) {
	g := extraTurnGame()
	engine := extraTurnEngine(t)
	source := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	agents := [game.NumPlayers]PlayerAgent{}

	castExtraTurnSorcery(t, g, engine, game.Player1, "Authored extra-turn gambit",
		game.Mode{Program: extraTurnGambitProgram}.Ability())
	castExtraTurnSorcery(t, g, engine, game.Player1, "Plain extra turn",
		game.Mode{Sequence: []game.Instruction{{Primitive: game.TakeExtraTurn{Player: game.ControllerReference()}}}}.Ability())
	if len(g.Turn.ExtraTurns) != 2 || g.Turn.ExtraTurns[0].ID == g.Turn.ExtraTurns[1].ID {
		t.Fatalf("extra turns = %+v, want two distinct Player1 extra turns", g.Turn.ExtraTurns)
	}
	gambitTurn, plainTurn := g.Turn.ExtraTurns[0].ID, g.Turn.ExtraTurns[1].ID

	engine.finishCurrentTurn(g, agents)
	if g.Turn.ActivePlayer != game.Player1 || g.Turn.ExtraTurnID != plainTurn {
		t.Fatalf("first extra turn = player %v id %v, want the plain turn %v", g.Turn.ActivePlayer, g.Turn.ExtraTurnID, plainTurn)
	}
	if dealt := extraTurnDamageDealt(engine, g, source); dealt != 0 {
		t.Fatalf("damage during the plain extra turn = %d, want prevented", dealt)
	}
	engine.runTurn(g, agents)
	if !isPlayerAlive(g, game.Player1) {
		t.Fatal("the plain extra turn's end step fired the gambit's trigger")
	}

	if g.Turn.ActivePlayer != game.Player1 || g.Turn.ExtraTurnID != gambitTurn {
		t.Fatalf("second extra turn = player %v id %v, want the gambit turn %v", g.Turn.ActivePlayer, g.Turn.ExtraTurnID, gambitTurn)
	}
	if dealt := extraTurnDamageDealt(engine, g, source); dealt != extraTurnProbeDealt {
		t.Fatalf("damage during the gambit's extra turn = %d, want %d unpreventable", dealt, extraTurnProbeDealt)
	}
	engine.runTurn(g, agents)
	if isPlayerAlive(g, game.Player1) {
		t.Fatal("Player1 did not lose at the gambit's extra turn end step")
	}
}

// TestTakeExtraTurnScopeEndsWhenTurnSkipped proves that when the extra turn
// is never taken, the delayed trigger never triggers and the "during that turn" effect never applies to another
// turn (Alchemist's Gambit ruling).
func TestTakeExtraTurnScopeEndsWhenTurnSkipped(t *testing.T) {
	tests := []struct {
		name string
		skip func(*game.Game)
	}{
		{
			name: "turn skipped",
			skip: func(g *game.Game) { g.SkippedTurns = map[game.PlayerID]int{game.Player1: 1} },
		},
		{
			name: "player skips extra turns",
			skip: func(g *game.Game) {
				addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
					Name:  "Extra Turn Suppressor",
					Types: []types.Card{types.Enchantment},
					StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
						Kind:           game.RuleEffectSkipExtraTurns,
						AffectedPlayer: game.PlayerOpponent,
					}}}},
				}})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := extraTurnGame()
			engine := extraTurnEngine(t)
			source := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
			agents := [game.NumPlayers]PlayerAgent{}
			castExtraTurnSorcery(t, g, engine, game.Player1, "Authored extra-turn gambit",
				game.Mode{Program: extraTurnGambitProgram}.Ability())
			test.skip(g)
			engine.finishCurrentTurn(g, agents)
			if g.Turn.ActivePlayer != game.Player2 || g.Turn.ExtraTurnID != 0 {
				t.Fatalf("turn = player %v extra %v, want Player2's regular turn", g.Turn.ActivePlayer, g.Turn.ExtraTurnID)
			}
			if effects, triggers := extraTurnScopeCount(g); effects != 0 || triggers != 0 {
				t.Fatalf("scoped effects/triggers = %d/%d, want dropped with the turn", effects, triggers)
			}
			if dealt := extraTurnDamageDealt(engine, g, source); dealt != 0 {
				t.Fatalf("damage on the following turn = %d, want prevented", dealt)
			}
			engine.runTurn(g, agents)
			if !isPlayerAlive(g, game.Player1) {
				t.Fatal("Player1 lost although the extra turn was skipped")
			}
		})
	}
}

// TestTakeExtraTurnWithoutEndStepTrigger is the cleaved grant: damage can't be
// prevented during the extra turn, but no delayed trigger is created.
func TestTakeExtraTurnWithoutEndStepTrigger(t *testing.T) {
	g := extraTurnGame()
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	agents := [game.NumPlayers]PlayerAgent{}
	triggers := len(g.DelayedTriggers)
	castExtraTurnSorcery(t, g, engine, game.Player1, "Cleaved extra-turn gambit",
		game.Mode{Sequence: []game.Instruction{{Primitive: gambitExtraTurn(false)}}}.Ability())
	if len(g.DelayedTriggers) != triggers {
		t.Fatalf("delayed triggers = %+v, want none created", g.DelayedTriggers)
	}
	engine.finishCurrentTurn(g, agents)
	if dealt := extraTurnDamageDealt(engine, g, source); dealt != extraTurnProbeDealt {
		t.Fatalf("damage during the extra turn = %d, want %d unpreventable", dealt, extraTurnProbeDealt)
	}
	engine.runTurn(g, agents)
	if !isPlayerAlive(g, game.Player1) || g.Turn.ActivePlayer != game.Player2 {
		t.Fatal("cleaved extra turn did not end normally")
	}
	if dealt := extraTurnDamageDealt(engine, g, source); dealt != 0 {
		t.Fatalf("damage after the extra turn = %d, want prevented", dealt)
	}
}

// TestTakeExtraTurnValidation rejects declarations whose scope would not be the
// granted turn.
func TestTakeExtraTurnValidation(t *testing.T) {
	valid := gambitExtraTurn(true)
	if err := game.ValidatePrimitive(valid, nil); err != nil {
		t.Fatalf("gambit grant rejected: %v", err)
	}
	withDuration := gambitExtraTurn(false)
	withDuration.RuleEffects[0].Duration = game.DurationUntilEndOfTurn
	withTiming := gambitExtraTurn(true)
	withTiming.EndStepTriggers[0].Timing = game.DelayedAtBeginningOfNextUpkeep
	withEvent := gambitExtraTurn(true)
	withEvent.EndStepTriggers[0].Window = game.DelayedWindowThisTurn
	noContent := gambitExtraTurn(true)
	noContent.EndStepTriggers[0].Content = game.AbilityContent{}
	for name, invalid := range map[string]game.Primitive{
		"no player":             game.TakeExtraTurn{},
		"rule effect duration":  withDuration,
		"trigger timing":        withTiming,
		"trigger window":        withEvent,
		"trigger content":       noContent,
		"apply rule extra turn": game.ApplyRule{RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectDamageCantBePrevented}}, Duration: game.DurationDuringExtraTurn},
	} {
		if err := game.ValidatePrimitive(invalid, nil); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// extraTurnGame returns a game whose players can draw through several turns,
// so only the scoped trigger can make a player lose.
func extraTurnGame() *game.Game {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for player := range game.NumPlayers {
		fillLibrary(g, game.PlayerID(player), 5)
	}
	return g
}
