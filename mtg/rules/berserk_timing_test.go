package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func beforeCombatDamageInstant() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Types: []types.Card{types.Instant},
		StaticAbilities: []game.StaticAbility{{
			CastOnlyBeforeCombatDamageStep: true,
		}},
		SpellAbility: opt.Val(game.Mode{}.Ability()),
	}}
}

func TestBeforeCombatDamageCastRestrictionAcrossTurnStructure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		phase       game.Phase
		step        game.Step
		combats     int
		extraPhases []game.Phase
		want        bool
	}{
		{"beginning", game.PhaseBeginning, game.StepUpkeep, 0, nil, true},
		{"precombat main", game.PhasePrecombatMain, game.StepNone, 0, nil, true},
		{"beginning of combat", game.PhaseCombat, game.StepBeginningOfCombat, 1, nil, true},
		{"declare attackers", game.PhaseCombat, game.StepDeclareAttackers, 1, nil, true},
		{"declare blockers", game.PhaseCombat, game.StepDeclareBlockers, 1, nil, true},
		{"first strike damage", game.PhaseCombat, game.StepFirstStrikeDamage, 1, nil, false},
		{"first strike despite extra combat", game.PhaseCombat, game.StepFirstStrikeDamage, 1, []game.Phase{game.PhaseCombat}, false},
		{"normal damage", game.PhaseCombat, game.StepCombatDamage, 1, nil, false},
		{"end combat no extra", game.PhaseCombat, game.StepEndOfCombat, 1, nil, false},
		{"end combat before extra", game.PhaseCombat, game.StepEndOfCombat, 1, []game.Phase{game.PhaseCombat}, true},
		{"postcombat no extra", game.PhasePostcombatMain, game.StepNone, 1, nil, false},
		{"postcombat turn with no combat", game.PhasePostcombatMain, game.StepNone, 0, nil, false},
		{"postcombat before extra", game.PhasePostcombatMain, game.StepNone, 1, []game.Phase{game.PhaseCombat}, true},
		{"postcombat before later queued combat", game.PhasePostcombatMain, game.StepNone, 1, []game.Phase{game.PhasePostcombatMain, game.PhaseCombat}, true},
		{"ending", game.PhaseEnding, game.StepEnd, 1, []game.Phase{game.PhaseCombat}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Turn.Phase = test.phase
			g.Turn.Step = test.step
			g.Turn.CombatPhasesThisTurn = test.combats
			g.Turn.ExtraPhases = append([]game.Phase(nil), test.extraPhases...)
			if got := canCastAtCurrentTiming(g, game.Player1, beforeCombatDamageInstant()); got != test.want {
				t.Fatalf("canCastAtCurrentTiming = %v, want %v", got, test.want)
			}
		})
	}
}

func TestCastRestrictionsCompose(t *testing.T) {
	t.Parallel()
	card := beforeCombatDamageInstant()
	card.StaticAbilities = append(card.StaticAbilities, game.StaticAbility{
		CastOnlyAfterAttackedThisStep: true,
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{PlayersAttacked: map[game.PlayerID]bool{game.Player1: true}}
	if !canCastAtCurrentTiming(g, game.Player1, card) {
		t.Fatal("composed restrictions rejected a timing satisfying both")
	}

	g.Turn.Step = game.StepDeclareBlockers
	if canCastAtCurrentTiming(g, game.Player1, card) {
		t.Fatal("composed restrictions ignored the declare-attackers restriction")
	}
}

func TestFreeCastStillObeysBeforeCombatDamageRestriction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   beforeCombatDamageInstant(),
		Owner: game.Player1,
	}

	g.Players[game.Player1].Exile.Add(cardID)
	g.Turn.Phase = game.PhasePostcombatMain
	g.Turn.CombatPhasesThisTurn = 1

	if engine.castFreeSpellFromSource(
		g, g.Players[game.Player1], game.Player1, cardID, zone.Exile, false,
		[game.NumPlayers]PlayerAgent{}, &TurnLog{},
	) {
		t.Fatal("free cast bypassed the before-combat-damage restriction")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("rejected free cast removed the card from exile")
	}

	g.Turn.Phase = game.PhasePrecombatMain
	if !engine.castFreeSpellFromSource(
		g, g.Players[game.Player1], game.Player1, cardID, zone.Exile, false,
		[game.NumPlayers]PlayerAgent{}, &TurnLog{},
	) {
		t.Fatal("free cast was rejected while before combat damage")
	}
}

func TestFreeCopyCastStillObeysBeforeCombatDamageRestriction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   beforeCombatDamageInstant(),
		Owner: game.Player1,
	}
	g.Turn.Phase = game.PhasePostcombatMain
	g.Turn.CombatPhasesThisTurn = 1

	if engine.castFreeCopyOfCard(
		g, game.Player1, cardID, [game.NumPlayers]PlayerAgent{}, &TurnLog{},
	) {
		t.Fatal("free copy cast bypassed the before-combat-damage restriction")
	}

	g.Turn.Phase = game.PhasePrecombatMain
	if !engine.castFreeCopyOfCard(
		g, game.Player1, cardID, [game.NumPlayers]PlayerAgent{}, &TurnLog{},
	) {
		t.Fatal("free copy cast was rejected while before combat damage")
	}
}

func TestPaidResolutionCastObeysBeforeCombatDamageRestriction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   beforeCombatDamageInstant(),
		Owner: game.Player1,
	}
	g.Players[game.Player1].Graveyard.Add(cardID)
	g.Turn.Phase = game.PhasePostcombatMain
	g.Turn.CombatPhasesThisTurn = 1

	if engine.castPaidSpellFromSource(
		g, g.Players[game.Player1], game.Player1, cardID, zone.Graveyard,
		[game.NumPlayers]PlayerAgent{}, &TurnLog{},
	) {
		t.Fatal("paid resolution cast bypassed the before-combat-damage restriction")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("rejected paid cast removed the card from the graveyard")
	}

	g.Turn.Phase = game.PhasePrecombatMain
	if !engine.castPaidSpellFromSource(
		g, g.Players[game.Player1], game.Player1, cardID, zone.Graveyard,
		[game.NumPlayers]PlayerAgent{}, &TurnLog{},
	) {
		t.Fatal("paid resolution cast was rejected while before combat damage")
	}
}
