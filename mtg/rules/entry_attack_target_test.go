package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine gap #3683. Official CR effective 2025-11-14, read
// 506.2a, 506.3a-e, 508.4, 508.4a, 508.4c, 802.2 and 802.2a. 508.4: a creature put onto the battlefield attacking attacks
// the defending player, planeswalker a defending player controls, or battle a
// defending player protects its controller chooses as it enters; it is
// "attacking" but never "attacked". 508.4c: attack requirements and
// restrictions don't apply. 802.2: with the attack multiple players option
// every opponent of the attacking player is a defending player, attacked or
// not. 506.3e: a creature that is also a battle never becomes attacking.
// Hero of Bladehold (oracle 454a8902-8120-4373-96ee-bbf352b04e8d, {2}{W}{W}
// 3/4) "Battle cry ... Whenever this creature attacks, create two 1/1 white
// Soldier creature tokens that are tapped and attacking." Rulings 2011-06-01:
// both abilities trigger in the order its controller chooses; you choose which
// opponent or planeswalker an opponent controls each token is attacking; the
// tokens were never declared as attackers. Jace Beleren (oracle
// cc2b6e58-2d6f-495e-95f3-9496105c4cba, loyalty 3). Invasion of Segovia (oracle
// 7b3c0d89-55ed-43d7-a00a-7b060857ec4e, Battle — Siege, defense 4). The
// non-token route is the PutOnBattlefield.EntryAttacking Thunderkin Awakener
// shape (oracle c0171343-9c0f-4001-8f4e-a5af9256c4d2; 2019-07-12 rulings: its
// controller chooses what the returned creature attacks, which need not match
// what Thunderkin attacks).

// entryAttackChooser answers each entry-attack choice with the next scripted
// destination, recording the destinations it was offered, and accepts the
// default for every other choice.
type entryAttackChooser struct {
	picks   []game.AttackTarget
	offered [][]game.AttackTarget
}

func (*entryAttackChooser) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (a *entryAttackChooser) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind != game.ChoiceCombatDeclaration || len(a.picks) == 0 {
		return request.DefaultSelection
	}
	var offered []game.AttackTarget
	answer := -1
	for _, option := range request.Options {
		offered = append(offered, option.Attack.Val.Target)
		if option.Attack.Val.Target == a.picks[0] {
			answer = option.Index
		}
	}
	a.offered = append(a.offered, offered)
	a.picks = a.picks[1:]
	return []int{answer}
}

// Hero of Bladehold attacks Player2. Each of its tapped-and-attacking Soldier
// tokens may attack any defending player, including the unattacked Player3 and
// Player4, a planeswalker Player3 controls, or a battle Player4 protects, but
// never Player1's own planeswalker. One token attacks Player3's Jace and the
// other Player2's Siege protected by Player4; both deal combat damage there.
func TestHeroOfBladeholdTokensAttackChosenPlaneswalkerAndBattle(t *testing.T) {
	g := bandCombatGame()
	engine := choicesProgramEngine(t)
	hero := registryCombatPermanent(t, g, game.Player1, "Hero of Bladehold")
	registryCombatPermanent(t, g, game.Player1, "Jace Beleren").Counters.Add(counter.Loyalty, 3)
	jace := registryCombatPermanent(t, g, game.Player3, "Jace Beleren")
	jace.Counters.Add(counter.Loyalty, 3)
	siege := registryCombatPermanent(t, g, game.Player2, "Invasion of Segovia")
	siege.Counters.Add(counter.Defense, 4)
	siege.BattleProtector = opt.Val(game.Player4)

	if !engine.applyDeclareAttackers(g, game.Player1, action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		{Attacker: hero.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}) {
		t.Fatal("Hero of Bladehold could not attack")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Hero of Bladehold's attack created no triggers")
	}
	atJace := game.AttackTarget{Player: game.Player3, PlaneswalkerID: jace.ObjectID}
	atSiege := game.AttackTarget{Player: game.Player4, BattleID: siege.ObjectID}
	chooser := &entryAttackChooser{picks: []game.AttackTarget{atJace, atSiege}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: chooser}
	log := &TurnLog{}
	for !g.Stack.IsEmpty() {
		engine.resolveTopOfStackWithChoices(g, agents, log)
	}

	want := []game.AttackTarget{
		{Player: game.Player2}, {Player: game.Player3}, {Player: game.Player4}, atJace, atSiege,
	}
	if len(chooser.offered) != 2 || !slices.Equal(chooser.offered[0], want) || !slices.Equal(chooser.offered[1], want) {
		t.Fatalf("offered destinations = %+v, want %+v for each token", chooser.offered, want)
	}
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			tokens = append(tokens, permanent)
		}
	}
	if len(tokens) != 2 {
		t.Fatalf("tokens = %d, want two Soldiers", len(tokens))
	}
	for i, target := range []game.AttackTarget{atJace, atSiege} {
		attack, attacking := attackerDeclarationFor(g, tokens[i].ObjectID)
		if !attacking || attack.Target != target || !tokens[i].Tapped {
			t.Fatalf("token %d attack = %+v (attacking %v, tapped %v), want tapped attacking %+v", i, attack, attacking, tokens[i].Tapped, target)
		}
	}
	if !g.Combat.PlayersAttacked[game.Player3] || !g.Combat.PlayersAttacked[game.Player4] {
		t.Fatalf("players attacked = %v, want Player3 and Player4 recorded", g.Combat.PlayersAttacked)
	}
	if g.Combat.DeclaredPlayerAttacks[game.Player1][game.Player3] || g.Combat.DeclaredPlayerAttacks[game.Player1][game.Player4] {
		t.Fatal("tokens entering attacking counted as Player1 having attacked those players")
	}
	if controller, ok := g.Combat.CombatantControllers[siege.ObjectID]; !ok || controller != game.Player2 {
		t.Fatalf("Siege combat controller = %v/%v, want Player2", controller, ok)
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("tokens entering attacking triggered an attack ability")
	}

	jacePower, siegePower := effectivePower(g, tokens[0]), effectivePower(g, tokens[1])
	engine.resolveCombatDamage(g, log)
	if got := jace.Counters.Get(counter.Loyalty); got != 3-jacePower {
		t.Fatalf("Jace loyalty = %d, want %d after the token's damage", got, 3-jacePower)
	}
	if got := siege.Counters.Get(counter.Defense); got != 4-siegePower {
		t.Fatalf("Siege defense = %d, want %d after the token's damage", got, 4-siegePower)
	}
}

// A creature card returned tapped and attacking by the Thunderkin shape may
// attack a planeswalker controlled by Player4, whom nothing has attacked: every
// opponent is a defending player (CR 802.2).
func TestPutOnBattlefieldEntryAttackingChoosesUnattackedPlayersPlaneswalker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, thunderkinShapeDef())
	returnedCard := addCardToGraveyard(g, game.Player1, toughnessTestCreature("Returned Elemental", types.Elemental, 2, 1))
	planeswalker := addCombatPermanent(g, game.Player4, &game.CardDef{CardFace: game.CardFace{
		Name:    "Test Planeswalker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(5),
	}})
	planeswalker.Counters.Add(counter.Loyalty, 5)

	declareThunderkinAttack(t, g, engine, source)
	atPlaneswalker := game.AttackTarget{Player: game.Player4, PlaneswalkerID: planeswalker.ObjectID}
	chooser := &entryAttackChooser{picks: []game.AttackTarget{atPlaneswalker}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: chooser}, &TurnLog{})

	returned, ok := reanimatedPermanent(g, returnedCard)
	if !ok {
		t.Fatal("target was not returned")
	}
	if attack, attacking := attackerDeclarationFor(g, returned.ObjectID); !attacking || attack.Target != atPlaneswalker {
		t.Fatalf("returned attack = %+v (attacking %v), want attacking %+v", attack, attacking, atPlaneswalker)
	}
	if !g.Combat.PlayersAttacked[game.Player4] {
		t.Fatal("Player4 was not recorded as attacked")
	}
	engine.resolveCombatDamage(g, &TurnLog{})
	if got := planeswalker.Counters.Get(counter.Loyalty); got != 3 {
		t.Fatalf("planeswalker loyalty = %d, want 3 after 2 combat damage", got)
	}
}

// Without an agent, a token entering attacking joins the player already being
// attacked rather than the first opponent in seat order, and a creature that is
// also a battle enters without attacking (CR 506.3e).
func TestCreateTokenEntryAttackingDefaultsToAttackedPlayerAndSkipsCreatureBattles(t *testing.T) {
	g := bandCombatGame()
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanent(g, game.Player1)
	g.Combat.Attackers = []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player3}}}

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:         game.Fixed(1),
		Source:         game.TokenDef(attackingTokenDef()),
		EntryAttacking: true,
	}, &TurnLog{})
	token := newlyCreatedToken(g)
	if attack, attacking := attackerDeclarationFor(g, token.ObjectID); !attacking || attack.Target != (game.AttackTarget{Player: game.Player3}) {
		t.Fatalf("token attack = %+v (attacking %v), want the default to join the attack on Player3", attack, attacking)
	}

	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenDef(&game.CardDef{CardFace: game.CardFace{
			Name:      "Creature Battle",
			Types:     []types.Card{types.Creature, types.Battle},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			Defense:   opt.Val(3),
		}}),
		EntryAttacking: true,
	}, &TurnLog{})
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.ObjectID != token.ObjectID {
			if _, attacking := attackerDeclarationFor(g, permanent.ObjectID); attacking {
				t.Fatal("a creature that is also a battle entered attacking")
			}
			return
		}
	}
	t.Fatal("creature battle token did not enter the battlefield")
}
