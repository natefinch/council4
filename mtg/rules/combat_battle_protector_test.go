package rules

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3506. Official CR effective 2026-09-25 (verified local
// copy), read 310.9-310.12a, 506.4, 506.4c, 508.1b, 704.5 and 800.4a/e. 310.9:
// each battle has a protector, chosen as it enters; 310.9a: a battle without
// battle types is protected by its controller; 310.9b: a battle can't be
// attacked by its protector; 310.9c: only the protector's creatures can block
// its attackers; 310.9d: a battle's attackers attack in the direction of its
// protector; 310.11 (SBA): a battle that isn't being attacked with no protector,
// or whose protector can't protect it, gets a new protector chosen by its
// controller, or is put into its owner's graveyard if none can be chosen;
// 310.12a: a Siege's controller chooses an opponent to protect it. 506.4: a
// permanent is removed from combat if its controller or protector changes;
// 506.4c: a creature attacking a removed battle keeps attacking nothing and,
// unblocked, deals no combat damage. 800.4e: combat damage isn't assigned to a
// player who has left the game. Cards (pinned corpus): Invasion of Segovia
// (oracle 7b3c0d89-55ed-43d7-a00a-7b060857ec4e, {2}{U} Battle — Siege, defense
// 4) "(As a Siege enters, choose an opponent to protect it. You and others can
// attack it. When it's defeated, exile it, then cast it transformed.) When this
// Siege enters, create two 1/1 blue Kraken creature tokens with trample."
// Rulings 2023-04-14: the controller chooses an opponent as protector as the
// Siege enters and can't be its protector; if the protector gains control, a
// new protector is chosen as a state-based action; any player other than the
// protector, including the controller, can attack it; only the protector's
// creatures can block its attackers; if the protector leaves the game while it
// isn't being attacked a new protector is chosen, and while it is attacked only
// once no creature attacks it; if the controller changes during combat it is
// removed from combat. Word of Seizing (oracle
// b80d2841-03b0-42d3-ac9b-51e793b1dc24, {3}{R}{R} instant) "Split second ...
// Untap target permanent and gain control of it until end of turn. It gains
// haste until end of turn." Grizzly Bears (oracle
// 14c8f55d-d177-4c25-a931-ebeb9e6062a0, {1}{G} 2/2).

// protectorChooser answers battle-protector prompts with player and records
// the options it was offered.
func protectorChooser(player game.PlayerID, offered *[][]string) *controlScript {
	return &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if !strings.HasPrefix(request.Prompt, "Choose the protector") {
			return fallbackChoice(request)
		}
		var labels []string
		selected := -1
		for i, option := range request.Options {
			labels = append(labels, option.Label)
			if option.Label == playerLabel(player) {
				selected = i
			}
		}
		*offered = append(*offered, labels)
		return []int{selected}
	}}
}

func playerLabel(player game.PlayerID) string {
	return fmt.Sprintf("Player %d", player+1)
}

// siegeUnderAttack puts Invasion of Segovia under controller's control
// protected by protector and has Player1's Grizzly Bears attack it.
func siegeUnderAttack(t *testing.T, controller, protector game.PlayerID) (*game.Game, *Engine, *game.Permanent, *game.Permanent) {
	t.Helper()
	g := bandCombatGame()
	engine := choicesProgramEngine(t)
	attacker := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	siege := registryCombatPermanent(t, g, controller, "Invasion of Segovia")
	siege.Counters.Add(counter.Defense, 4)
	siege.BattleProtector = opt.Val(protector)
	target := game.AttackTarget{Player: protector, BattleID: siege.ObjectID}
	if !slices.Contains(legalAttackTargets(g, game.Player1), target) {
		t.Fatalf("Player1 attack targets %+v lack the Siege toward its protector", legalAttackTargets(g, game.Player1))
	}
	if !engine.applyDeclareAttackers(g, game.Player1, action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{attackAt(attacker, target)}}) {
		t.Fatal("attack on the Siege was rejected")
	}
	return g, engine, attacker, siege
}

func TestSiegeEntersWithChosenOpponentProtector(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	g.Turn.PriorityPlayer = game.Player1
	engine := choicesProgramEngine(t)
	card := addCardToHand(g, game.Player1, registryCardDef(t, "Invasion of Segovia"))
	g.Players[game.Player1].ManaPool.Add(mana.U, 3)
	var offered [][]string
	agents := [game.NumPlayers]PlayerAgent{game.Player1: protectorChooser(game.Player3, &offered)}
	log := &TurnLog{}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(card, nil, 0, nil), agents, log) {
		t.Fatal("casting Invasion of Segovia failed")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)

	siege, ok := findPermanentByCardID(g, card)
	if !ok {
		t.Fatal("Invasion of Segovia did not enter the battlefield")
	}
	if protector, ok := battleProtector(siege); !ok || protector != game.Player3 {
		t.Fatalf("protector = %v/%v, want Player3", protector, ok)
	}
	if want := [][]string{{"Player 2", "Player 3", "Player 4"}}; !slices.EqualFunc(offered, want, slices.Equal) {
		t.Fatalf("protector options = %v, want %v", offered, want)
	}
	target := game.AttackTarget{Player: game.Player3, BattleID: siege.ObjectID}
	for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player4} {
		if !slices.Contains(legalAttackTargets(g, player), target) {
			t.Fatalf("player %d can't attack the Siege toward its protector", player+1)
		}
	}
	for _, candidate := range legalAttackTargets(g, game.Player3) {
		if candidate.BattleID == siege.ObjectID {
			t.Fatalf("the protector can attack its own Siege: %+v", candidate)
		}
	}
	if isLegalAttackTarget(g, game.Player2, game.AttackTarget{Player: game.Player1, BattleID: siege.ObjectID}) {
		t.Fatal("the Siege can be attacked toward its controller instead of its protector")
	}
}

func TestStolenSiegeStopsBeingAttacked(t *testing.T) {
	g, engine, attacker, siege := siegeUnderAttack(t, game.Player2, game.Player3)
	g.Turn.Step = game.StepDeclareBlockers
	if engine.applyDeclareBlockers(g, game.Player2, action.DeclareBlockersAction{}) {
		t.Fatal("the Siege's controller declared blockers for an attack toward its protector")
	}
	if !engine.applyDeclareBlockers(g, game.Player3, action.DeclareBlockersAction{}) {
		t.Fatal("the protector's empty block declaration was rejected")
	}

	castAndResolveWordOfSeizing(t, g, engine, game.Player1, siege)
	want := []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{NoTarget: true, Player: game.Player3}}}
	if !slices.Equal(g.Combat.Attackers, want) {
		t.Fatalf("attackers = %+v, want %+v", g.Combat.Attackers, want)
	}
	assertNoCombatDamage(t, g, engine, attacker)
	if defense := siege.Counters.Get(counter.Defense); defense != 4 {
		t.Fatalf("Siege defense = %d, want 4", defense)
	}
	engine.applyStateBasedActions(g)
	if protector, ok := battleProtector(siege); !ok || protector != game.Player3 {
		t.Fatalf("protector = %v/%v, want Player3 to keep protecting the stolen Siege", protector, ok)
	}
}

func TestSiegeTakenByItsProtectorGetsNewProtector(t *testing.T) {
	g, engine, attacker, siege := siegeUnderAttack(t, game.Player1, game.Player3)
	castAndResolveWordOfSeizing(t, g, engine, game.Player3, siege)
	want := []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{NoTarget: true, Player: game.Player3}}}
	if !slices.Equal(g.Combat.Attackers, want) {
		t.Fatalf("attackers = %+v, want %+v", g.Combat.Attackers, want)
	}

	var offered [][]string
	engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player3: protectorChooser(game.Player4, &offered)}, &TurnLog{})
	defer g.ClearChoiceContext()
	engine.applyStateBasedActions(g)
	if protector, ok := battleProtector(siege); !ok || protector != game.Player4 {
		t.Fatalf("protector = %v/%v, want Player4 chosen by the new controller", protector, ok)
	}
	if want := [][]string{{"Player 1", "Player 2", "Player 4"}}; !slices.EqualFunc(offered, want, slices.Equal) {
		t.Fatalf("protector options = %v, want %v", offered, want)
	}
	assertNoCombatDamage(t, g, engine, attacker)
	if defense := siege.Counters.Get(counter.Defense); defense != 4 {
		t.Fatalf("Siege defense = %d, want 4", defense)
	}
}

func TestBattleWithChangedProtectorStopsBeingAttacked(t *testing.T) {
	g, _, attacker, siege := siegeUnderAttack(t, game.Player2, game.Player3)
	siege.BattleProtector = opt.Val(game.Player4)
	if !settleControlChanges(g) {
		t.Fatal("a protector change was not settled")
	}
	want := []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{NoTarget: true, Player: game.Player3}}}
	if !slices.Equal(g.Combat.Attackers, want) {
		t.Fatalf("attackers = %+v, want %+v", g.Combat.Attackers, want)
	}
}

func TestAttackedSiegeKeepsDepartedProtectorUntilUnattacked(t *testing.T) {
	g, engine, attacker, siege := siegeUnderAttack(t, game.Player2, game.Player3)
	g.Players[game.Player3].Life = 0
	engine.applyStateBasedActions(g)
	if !g.Players[game.Player3].Eliminated {
		t.Fatal("Player3 did not lose the game")
	}
	want := []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player3, BattleID: siege.ObjectID}}}
	if !slices.Equal(g.Combat.Attackers, want) {
		t.Fatalf("attackers = %+v, want the Siege still attacked: %+v", g.Combat.Attackers, want)
	}
	if protector, ok := battleProtector(siege); !ok || protector != game.Player3 {
		t.Fatalf("protector = %v/%v, want no new protector while attacked", protector, ok)
	}

	g.Turn.Step = game.StepCombatDamage
	engine.resolveCombatDamage(g, &TurnLog{})
	if defense := siege.Counters.Get(counter.Defense); defense != 2 {
		t.Fatalf("Siege defense = %d, want 2 after the unblocked Bears' damage", defense)
	}

	g.Combat = nil
	engine.applyStateBasedActions(g)
	if protector, ok := battleProtector(siege); !ok || protector != game.Player1 {
		t.Fatalf("protector = %v/%v, want Player1 once the Siege is no longer attacked", protector, ok)
	}
}

func TestSiegeWithoutPossibleProtectorGoesToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	siege := registryCombatPermanent(t, g, game.Player1, "Invasion of Segovia")
	siege.Counters.Add(counter.Defense, 4)
	siege.BattleProtector = opt.Val(game.Player2)
	for _, player := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		g.Players[player].Eliminated = true
	}
	_, deaths := NewEngine(nil).applyStateBasedActionsWithDeaths(g)
	if len(deaths) != 1 || deaths[0].Permanent != siege.ObjectID || deaths[0].Reason != PermanentDeathReasonNoProtector {
		t.Fatalf("deaths = %+v, want the Siege to die without a protector", deaths)
	}
	if !g.Players[game.Player1].Graveyard.Contains(siege.CardInstanceID) {
		t.Fatal("the Siege was not put into its owner's graveyard")
	}
}

func TestUnprotectedSiegeGetsProtectorChosenByController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	siege := registryCombatPermanent(t, g, game.Player2, "Invasion of Segovia")
	siege.Counters.Add(counter.Defense, 4)
	engine := NewEngine(nil)
	var offered [][]string
	engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player2: protectorChooser(game.Player4, &offered)}, &TurnLog{})
	defer g.ClearChoiceContext()
	engine.applyStateBasedActions(g)
	if protector, ok := battleProtector(siege); !ok || protector != game.Player4 {
		t.Fatalf("protector = %v/%v, want Player4", protector, ok)
	}
	if want := [][]string{{"Player 1", "Player 3", "Player 4"}}; !slices.EqualFunc(offered, want, slices.Equal) {
		t.Fatalf("protector options = %v, want %v", offered, want)
	}
}

func TestBattleWithoutBattleTypeEntersProtectedByController(t *testing.T) {
	g := mainPhaseGame(game.Player2)
	g.Turn.PriorityPlayer = game.Player2
	engine := NewEngine(nil)
	card := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Battle",
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		Types:    []types.Card{types.Battle},
		Defense:  opt.Val(3),
	}})
	g.Players[game.Player2].ManaPool.Add(mana.C, 1)
	if !engine.applyAction(g, game.Player2, action.CastSpell(card, nil, 0, nil)) {
		t.Fatal("casting the battle failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	battle, ok := findPermanentByCardID(g, card)
	if !ok {
		t.Fatal("the battle did not enter the battlefield")
	}
	if protector, ok := battleProtector(battle); !ok || protector != game.Player2 {
		t.Fatalf("protector = %v/%v, want its controller", protector, ok)
	}
	target := game.AttackTarget{Player: game.Player2, BattleID: battle.ObjectID}
	if !slices.Contains(legalAttackTargets(g, game.Player1), target) || slices.Contains(legalAttackTargets(g, game.Player2), target) {
		t.Fatal("only opponents of its controller-protector should be able to attack the battle")
	}
}
