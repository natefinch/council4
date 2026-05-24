package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

func TestPlayerTargetedSpellCreatesOneLegalActionPerAlivePlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, playerDamageSpell())
	if !engine.eliminatePlayer(g, game.Player4) {
		t.Fatal("eliminatePlayer() = false, want true")
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	var castTargets []game.Target
	for _, act := range legal {
		if act.Kind == action.ActionCastSpell && act.CastSpell.CardID == spellID {
			if len(act.CastSpell.Targets) != 1 {
				t.Fatalf("cast targets = %d, want 1", len(act.CastSpell.Targets))
			}
			castTargets = append(castTargets, act.CastSpell.Targets[0])
		}
	}
	wantTargets := []game.Target{
		game.PlayerTarget(game.Player1),
		game.PlayerTarget(game.Player2),
		game.PlayerTarget(game.Player3),
	}
	if len(castTargets) != len(wantTargets) {
		t.Fatalf("cast actions = %d, want %d", len(castTargets), len(wantTargets))
	}
	for i, want := range wantTargets {
		if castTargets[i] != want {
			t.Fatalf("target %d = %+v, want %+v", i, castTargets[i], want)
		}
	}
}

func TestTargetedEffectUsesSelectedTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, playerDamageSpell())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("applyAction() = false, want true")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Life != 40 {
		t.Fatalf("caster life = %d, want 40", g.Players[game.Player1].Life)
	}
	if g.Players[game.Player2].Life != 37 {
		t.Fatalf("target life = %d, want 37", g.Players[game.Player2].Life)
	}
}

func TestDeadPlayerTargetDoesNotApplyEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDamage,
		Amount:      3,
		TargetIndex: 0,
	}, []game.Target{game.PlayerTarget(game.Player2)})
	if !engine.eliminatePlayer(g, game.Player2) {
		t.Fatal("eliminatePlayer() = false, want true")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Life != 40 {
		t.Fatalf("dead target life = %d, want 40", g.Players[game.Player2].Life)
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceID) {
		t.Fatal("spell did not move to graveyard")
	}
}

func TestTargetThatDiesBeforeResolutionDoesNotApplyEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, playerDamageSpell())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("applyAction() = false, want true")
	}
	if !engine.eliminatePlayer(g, game.Player2) {
		t.Fatal("eliminatePlayer() = false, want true")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Life != 40 {
		t.Fatalf("dead target life = %d, want 40", g.Players[game.Player2].Life)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("spell did not move to graveyard")
	}
}

func playerDamageSpell() *game.CardDef {
	return &game.CardDef{
		Name:  "Needle Drop",
		Types: []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				Targets: []game.TargetSpec{
					{MinTargets: 1, MaxTargets: 1, Constraint: "player"},
				},
				Effects: []game.Effect{
					{Type: game.EffectDamage, Amount: 3, TargetIndex: 0},
				},
			},
		},
	}
}
