package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3492. Wings of the Guard (CN2 26 printing
// 4084a009-f366-416b-b82d-a762189abf43, oracle
// 73c0a809-fb0e-4773-a385-a6e32feb03cd, {1}{W} 1/1 Bird) "Flying / Melee
// (Whenever this creature attacks, it gets +1/+1 until end of turn for each
// opponent you attacked this combat.)". Scryfall rulings (all 2016-08-23): the
// bonus is set as melee resolves, counting each opponent attacked with one or
// more creatures whether or not those creatures are still attacking or on the
// battlefield and whether or not that opponent is still in the game; only
// opponents, not planeswalkers, count; creatures entering the battlefield
// attacking were never declared and don't count. CR (effective 2025-09-19,
// official text): 702.121a melee counts "each opponent you attacked with a
// creature this combat"; 508.6 a player has "attacked [a player]" if they
// declared one or more creatures as attackers attacking that player; 508.4
// creatures put onto the battlefield attacking never "attacked"; 506.4 control
// changes and leaving the battlefield remove a creature from combat. Supporting
// cards (pinned corpus): Grizzly Bears (oracle
// 14c8f55d-d177-4c25-a931-ebeb9e6062a0, {1}{G} 2/2), Murder (oracle
// 938b4e2c-88d9-4637-bc00-e228920c9a78, {1}{B}{B} "Destroy target creature."),
// Word of Seizing (oracle b80d2841-03b0-42d3-ac9b-51e793b1dc24, {3}{R}{R}
// split second "Untap target permanent and gain control of it until end of
// turn. It gains haste until end of turn.").

// meleeAttack declares registered Wings of the Guard and a registered Grizzly
// Bears as Player1's attackers at the given targets and puts Wings' melee
// trigger on the stack.
func meleeAttack(t *testing.T, wingsTarget, bearsTarget func(g *game.Game) game.AttackTarget) (*game.Game, *Engine, *game.Permanent, *game.Permanent) {
	t.Helper()
	g := bandCombatGame()
	engine := choicesProgramEngine(t)
	wings := registryCombatPermanent(t, g, game.Player1, "Wings of the Guard")
	bears := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	if !engine.applyDeclareAttackers(g, game.Player1, action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		attackAt(wings, wingsTarget(g)),
		attackAt(bears, bearsTarget(g)),
	}}) {
		t.Fatal("attack declaration was rejected")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Wings of the Guard's melee trigger was not put on the stack")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || trigger.SourceID != wings.ObjectID || g.Stack.Size() != 1 {
		t.Fatal("the stack does not hold exactly Wings of the Guard's melee trigger")
	}
	return g, engine, wings, bears
}

func attackPlayer(player game.PlayerID) func(*game.Game) game.AttackTarget {
	return func(*game.Game) game.AttackTarget { return game.AttackTarget{Player: player} }
}

// resolveMelee resolves Wings' melee trigger and checks its resulting size.
func resolveMelee(t *testing.T, g *game.Game, engine *Engine, wings *game.Permanent, want int) {
	t.Helper()
	trigger, ok := g.Stack.Peek()
	if !ok || trigger.SourceID != wings.ObjectID {
		t.Fatal("Wings of the Guard's melee trigger is not on top of the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Stack.Size() != 0 {
		t.Fatal("the melee trigger did not resolve")
	}
	toughness, _ := effectiveToughness(g, wings)
	if power := effectivePower(g, wings); power != want || toughness != want {
		t.Fatalf("Wings of the Guard = %d/%d, want %d/%d", power, toughness, want, want)
	}
}

// TestMeleeCountsAttackerThatLeftBattlefield is the #3492 scenario: Wings of
// the Guard attacks Player2 and Grizzly Bears attacks Player3, then the Bears
// are destroyed before melee resolves. Player1 still attacked both opponents
// (CR 508.6), so Wings gets +2/+2.
func TestMeleeCountsAttackerThatLeftBattlefield(t *testing.T) {
	g, engine, wings, bears := meleeAttack(t, attackPlayer(game.Player2), attackPlayer(game.Player3))
	castAndResolveRegistered(t, g, engine, game.Player3, "Murder", mana.B, 3, bears)
	if _, ok := permanentByObjectID(g, bears.ObjectID); ok {
		t.Fatal("Murder did not destroy Grizzly Bears")
	}
	if len(g.Combat.Attackers) != 1 {
		t.Fatalf("attackers = %+v, want only Wings of the Guard", g.Combat.Attackers)
	}
	resolveMelee(t, g, engine, wings, 3)
}

// TestMeleeCountsAttackerStolenOutOfCombat removes the Bears from combat by a
// control change instead (CR 506.4). The Bears are now Player3's creature and
// no longer attacking, but Player1 still attacked Player3 this combat.
func TestMeleeCountsAttackerStolenOutOfCombat(t *testing.T) {
	g, engine, wings, bears := meleeAttack(t, attackPlayer(game.Player2), attackPlayer(game.Player3))
	castAndResolveRegistered(t, g, engine, game.Player3, "Word of Seizing", mana.R, 5, bears)
	if effectiveController(g, bears) != game.Player3 || len(g.Combat.Attackers) != 1 {
		t.Fatal("Word of Seizing did not take the Bears out of Player1's attack")
	}
	resolveMelee(t, g, engine, wings, 3)
}

// TestMeleeCountsTriggerControllersAttacksAfterSourceStolen has Player2 steal
// Wings of the Guard itself before melee resolves. The trigger is still
// Player1's, and "you" is that controller: Player1 attacked two opponents, so
// the now-removed Wings still gets +2/+2.
func TestMeleeCountsTriggerControllersAttacksAfterSourceStolen(t *testing.T) {
	g, engine, wings, _ := meleeAttack(t, attackPlayer(game.Player2), attackPlayer(game.Player3))
	castAndResolveRegistered(t, g, engine, game.Player2, "Word of Seizing", mana.R, 5, wings)
	if effectiveController(g, wings) != game.Player2 || len(g.Combat.Attackers) != 1 {
		t.Fatal("Word of Seizing did not take Wings of the Guard out of combat")
	}
	if trigger, ok := g.Stack.Peek(); !ok || trigger.Controller != game.Player1 {
		t.Fatal("the melee trigger is no longer controlled by the attacking player")
	}
	resolveMelee(t, g, engine, wings, 3)
}

// TestMeleeCountsOpponentWhoLeftTheGame eliminates the opponent the Bears
// attacked before melee resolves. Per the ruling, it doesn't matter whether
// the attacked opponent is still in the game.
func TestMeleeCountsOpponentWhoLeftTheGame(t *testing.T) {
	g, engine, wings, _ := meleeAttack(t, attackPlayer(game.Player2), attackPlayer(game.Player3))
	g.Players[game.Player3].Life = 0
	engine.applyStateBasedActions(g)
	if !g.Players[game.Player3].Eliminated {
		t.Fatal("Player3 did not lose the game")
	}
	resolveMelee(t, g, engine, wings, 3)
}

// TestMeleeTargetDistinctions covers which attacks count: two creatures on one
// opponent count once, and an attack on a planeswalker does not count its
// controller as attacked.
func TestMeleeTargetDistinctions(t *testing.T) {
	attackPlaneswalker := func(g *game.Game) game.AttackTarget {
		planeswalker := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
			Name:    "Test Planeswalker",
			Types:   []types.Card{types.Planeswalker},
			Loyalty: opt.Val(3),
		}})
		planeswalker.Counters.Add(counter.Loyalty, 3)
		return game.AttackTarget{Player: game.Player3, PlaneswalkerID: planeswalker.ObjectID}
	}
	for _, test := range []struct {
		name  string
		bears func(*game.Game) game.AttackTarget
		want  int
	}{
		{name: "same opponent counted once", bears: attackPlayer(game.Player2), want: 2},
		{name: "planeswalker controller not counted", bears: attackPlaneswalker, want: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			g, engine, wings, _ := meleeAttack(t, attackPlayer(game.Player2), test.bears)
			resolveMelee(t, g, engine, wings, test.want)
		})
	}
}

// TestMeleeIgnoresCreaturesPutOntoBattlefieldAttacking puts a creature onto the
// battlefield attacking Player3 before melee resolves. It was never declared as
// an attacker, so Player1 hasn't attacked Player3 (CR 508.4, 508.6).
func TestMeleeIgnoresCreaturesPutOntoBattlefieldAttacking(t *testing.T) {
	g, engine, wings, _ := meleeAttack(t, attackPlayer(game.Player2), attackPlayer(game.Player2))
	token := addCombatPermanent(g, game.Player1, vanillaCreature("Soldier", 1, 1))
	declareTokenAttackingDefender(g, game.Player1, token, game.Player3)
	if _, attacking := attackDeclarationForAttacker(g, token.ObjectID); !attacking {
		t.Fatal("the creature did not enter attacking Player3")
	}
	resolveMelee(t, g, engine, wings, 2)
}
