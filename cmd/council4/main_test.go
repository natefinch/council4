package main

import (
	"bytes"
	"math/rand/v2"
	"strings"
	"sync"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules"
)

func TestSpellModeRunsDeterministicallyWithCastsAndResolves(t *testing.T) {
	result := runSpellMode(t, 1)

	if result.TurnCount != 52 {
		t.Fatalf("spell mode produced %d turns, want 52", result.TurnCount)
	}
	if !result.HasWinner || result.Winner != game.Player4 {
		t.Fatalf("spell mode winner = (%v,%v), want (true,%v)", result.HasWinner, result.Winner, game.Player4)
	}
	casts, resolves := countCastsAndResolves(result)
	if casts == 0 {
		t.Fatal("spell mode produced no casts")
	}
	if resolves == 0 {
		t.Fatal("spell mode produced no resolves")
	}
}

func TestCombatModeRunsDeterministicallyWithAttacksAndCombatDamage(t *testing.T) {
	_, result := runCombatModeSeed1(t)

	if result.TurnCount != 73 {
		t.Fatalf("combat mode produced %d turns, want 73", result.TurnCount)
	}
	if !result.HasWinner || result.Winner != game.Player4 {
		t.Fatalf("combat mode winner = (%v,%v), want (true,%v)", result.HasWinner, result.Winner, game.Player4)
	}
	attacks, damage := countAttacksAndDamage(result)
	if attacks == 0 {
		t.Fatal("combat mode produced no attacks")
	}
	if damage == 0 {
		t.Fatal("combat mode produced no combat damage")
	}
}

func TestPrintTurnLogIncludesCastAndResolve(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addTestCard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Test Spell"}})
	result := &rules.GameResult{
		Turns: []rules.TurnLog{
			{
				TurnNumber:   1,
				ActivePlayer: game.Player1,
				Actions: []rules.ActionLog{
					{Player: game.Player1, Action: action.CastSpell(cardID, nil, 0, nil)},
				},
				Resolves: []rules.ResolveLog{
					{SourceID: cardID, Controller: game.Player1, Kind: game.StackSpell, Result: "resolved"},
				},
			},
		},
	}
	var out bytes.Buffer

	printTurnLog(&out, g, result, logOptions{})

	got := out.String()
	for _, want := range []string{`Player 1: cast "Test Spell"`, `resolve spell "Test Spell"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}

func TestPrintTurnLogNoPassKeepsCombatEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addTestCard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Test Attacker"}})
	blockerCardID := addTestCard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Test Blocker"}})
	attackerID := g.IDGen.Next()
	blockerID := g.IDGen.Next()
	g.Battlefield = append(g.Battlefield,
		&game.Permanent{
			ObjectID:       attackerID,
			CardInstanceID: cardID,
			Owner:          game.Player1,
			Controller:     game.Player1,
		},
		&game.Permanent{
			ObjectID:       blockerID,
			CardInstanceID: blockerCardID,
			Owner:          game.Player2,
			Controller:     game.Player2,
		},
	)
	result := &rules.GameResult{
		Turns: []rules.TurnLog{
			{
				TurnNumber:   1,
				ActivePlayer: game.Player1,
				Actions: []rules.ActionLog{
					{Player: game.Player1, Action: action.Pass()},
					{Player: game.Player1, Action: action.DeclareAttackers([]game.AttackDeclaration{
						{Attacker: attackerID, Target: game.AttackTarget{Player: game.Player2}},
					})},
					{Player: game.Player1, Action: action.DeclareAttackers(nil)},
					{Player: game.Player2, Action: action.DeclareBlockers([]game.BlockDeclaration{
						{Blocker: blockerID, Blocking: attackerID},
					})},
					{Player: game.Player2, Action: action.DeclareBlockers(nil)},
				},
				CombatDamage: []rules.CombatDamageLog{
					{Attacker: attackerID, SourceID: cardID, Controller: game.Player1, DefendingPlayer: game.Player2, Damage: 2},
				},
				CreatureDamage: []rules.CreatureDamageLog{
					{SourcePermanent: blockerID, SourceID: blockerCardID, Controller: game.Player2, DamagedPermanent: attackerID, DamagedSourceID: cardID, DamagedController: game.Player1, Damage: 2},
				},
				Deaths: []rules.PermanentDeathLog{
					{Permanent: attackerID, SourceID: cardID, Owner: game.Player1, Controller: game.Player1, Reason: rules.PermanentDeathReasonLethalDamage},
				},
				Losses: []rules.LossLog{
					{Player: game.Player2, Reason: rules.LossReasonZeroLife},
				},
			},
		},
	}
	var out bytes.Buffer

	printTurnLog(&out, g, result, logOptions{OmitPasses: true})

	got := out.String()
	if strings.Contains(got, "Player 1: pass") {
		t.Fatalf("output included pass action:\n%s", got)
	}
	for _, want := range []string{
		`declare attackers: "Test Attacker" at Player 2`,
		`declare no attackers`,
		`declare blockers: "Test Blocker" blocks "Test Attacker"`,
		`declare no blockers`,
		`"Test Attacker" deals 2 combat damage to Player 2`,
		`"Test Blocker" deals 2 combat damage to "Test Attacker"`,
		`"Test Attacker" dies (lethal damage)`,
		`Player 2: loses (0 life)`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}

func TestPrintTurnLogNoPassKeepsOtherEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addTestCard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Test Spell"}})
	drawID := addTestCard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	result := &rules.GameResult{
		Turns: []rules.TurnLog{
			{
				TurnNumber:   1,
				ActivePlayer: game.Player1,
				Draws: []rules.DrawLog{
					{Player: game.Player1, CardID: drawID},
				},
				Losses: []rules.LossLog{
					{Player: game.Player2, Reason: rules.LossReasonZeroLife},
				},
				Actions: []rules.ActionLog{
					{Player: game.Player1, Action: action.CastSpell(cardID, nil, 0, nil)},
					{Player: game.Player2, Action: action.Pass()},
				},
				Resolves: []rules.ResolveLog{
					{SourceID: cardID, Controller: game.Player1, Kind: game.StackSpell, Result: "graveyard"},
				},
			},
		},
	}
	var out bytes.Buffer

	printTurnLog(&out, g, result, logOptions{OmitPasses: true})

	got := out.String()
	if strings.Contains(got, "Player 2: pass") {
		t.Fatalf("output included pass action:\n%s", got)
	}
	for _, want := range []string{`draw "Drawn Card"`, `loses (0 life)`, `cast "Test Spell"`, `resolve spell "Test Spell" (graveyard)`} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}

func TestCombatVerboseLogPrintsResolveBeforeResolvedPermanentAttacks(t *testing.T) {
	g, result := runCombatModeSeed1(t)
	var out bytes.Buffer

	printTurnLog(&out, g, result, logOptions{OmitPasses: true})

	got := out.String()
	castIndex := strings.Index(got, `Player 1: cast "Hasty Wolf"`)
	if castIndex < 0 {
		t.Fatalf("output missing Hasty Wolf cast:\n%s", got)
	}
	afterCast := got[castIndex:]
	resolveIndex := strings.Index(afterCast, `resolve spell "Hasty Wolf" (battlefield)`)
	if resolveIndex < 0 {
		t.Fatalf("output missing Hasty Wolf resolve after cast:\n%s", got)
	}
	attackIndex := strings.Index(afterCast, `Player 1: declare attackers:`)
	if attackIndex < 0 {
		t.Fatalf("output missing Player 1 attack after cast:\n%s", got)
	}
	if !strings.Contains(afterCast, `Player 1: declare attackers: "Hasty Wolf" at Player 2`) {
		t.Fatalf("output did not preserve Hasty Wolf attacker name after later zone changes:\n%s", got)
	}
	if resolveIndex > attackIndex {
		t.Fatalf("Hasty Wolf resolve printed after attack:\n%s", got[castIndex:castIndex+attackIndex+len(`Player 1: declare attackers:`)])
	}
}

type modeRun struct {
	game   *game.Game
	result *rules.GameResult
}

var combatModeSeed1 struct {
	once sync.Once
	run  modeRun
	err  error
}

func runCombatModeSeed1(t testing.TB) (*game.Game, *rules.GameResult) {
	t.Helper()
	combatModeSeed1.once.Do(func() {
		combatModeSeed1.run, combatModeSeed1.err = runMode("combat", 1)
	})
	if combatModeSeed1.err != nil {
		t.Fatal(combatModeSeed1.err)
	}
	return combatModeSeed1.run.game, combatModeSeed1.run.result
}

func addTestCard(g *game.Game, owner game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   def,
		Owner: owner,
	}
	return cardID
}

func runSpellMode(t testing.TB, seed uint64) *rules.GameResult {
	t.Helper()
	run, err := runMode("spells", seed)
	if err != nil {
		t.Fatal(err)
	}
	return run.result
}

func runMode(mode string, seed uint64) (modeRun, error) {
	configs, agents, err := gameModeConfig(mode, 8, false)
	if err != nil {
		return modeRun{}, err
	}
	engine := rules.NewEngine(rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)))
	g := engine.NewGame(configs)
	return modeRun{game: g, result: engine.RunGame(g, agents)}, nil
}

func countCastsAndResolves(result *rules.GameResult) (casts, resolves int) {
	for i := range result.Turns {
		turn := &result.Turns[i]
		for j := range turn.Actions {
			logged := &turn.Actions[j]
			if logged.Action.Kind == action.ActionCastSpell {
				casts++
			}
		}
		resolves += len(turn.Resolves)
	}
	return casts, resolves
}

func countAttacksAndDamage(result *rules.GameResult) (attacks, damage int) {
	for i := range result.Turns {
		turn := &result.Turns[i]
		for j := range turn.Actions {
			logged := &turn.Actions[j]
			attackers, ok := logged.Action.DeclareAttackersPayload()
			if ok && len(attackers.Attackers) > 0 {
				attacks++
			}
		}
		damage += len(turn.CombatDamage)
	}
	return attacks, damage
}
