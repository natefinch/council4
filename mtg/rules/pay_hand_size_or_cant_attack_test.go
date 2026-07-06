package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// payHandSizeOrCantAttackInstruction builds the single PlayerMayPayGenericOrRule
// instruction that Champions of Minas Tirith lowers to: the triggering opponent
// may pay generic mana equal to their hand size, and on non-payment their
// creatures can't attack the source's controller for the rest of that combat.
func payHandSizeOrCantAttackInstruction() game.Instruction {
	handSize := game.EventPlayerReference()
	handSelection := game.Selection{}
	return game.Instruction{
		Primitive: game.PlayerMayPayGenericOrRule{
			Player: game.EventPlayerReference(),
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:      game.DynamicAmountCountCardsInZone,
				Player:    &handSize,
				CardZone:  zone.Hand,
				Selection: &handSelection,
			}),
			RuleEffects: []game.RuleEffect{{
				Kind:                      game.RuleEffectCantAttack,
				AffectedPlayerRef:         game.EventPlayerReference(),
				DefendingPlayer:           game.PlayerYou,
				DefendingPlayerDirectOnly: true,
			}},
			Duration: game.DurationUntilEndOfCombat,
		},
	}
}

// addChampionsPermanent adds a Champions of Minas Tirith-style permanent under
// controller: "At the beginning of combat on each opponent's turn, if you're the
// monarch, that opponent may pay {X}, where X is the number of cards in their
// hand. If they don't, they can't attack you this combat."
func addChampionsPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	pt := game.PT{Value: 4}
	toughness := game.PT{Value: 6}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Champions of Minas Tirith",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(toughness),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerAt,
				Pattern: game.TriggerPattern{
					Event:      game.EventBeginningOfStep,
					Controller: game.TriggerControllerOpponent,
					Step:       game.StepBeginningOfCombat,
				},
				InterveningIf:        "if you're the monarch",
				InterveningCondition: opt.Val(game.Condition{ControllerIsMonarch: true}),
			},
			Content: game.Mode{
				Sequence: []game.Instruction{payHandSizeOrCantAttackInstruction()},
			}.Ability(),
		}},
	}})
}

// resolveChampionsCombatTrigger emits the beginning-of-combat event on the
// active player's turn and resolves the resulting trigger, if any. It reports
// whether a trigger was put on the stack.
func resolveChampionsCombatTrigger(engine *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) bool {
	emitBeginningOfStepEvent(g, game.StepBeginningOfCombat)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		return false
	}
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, &log)
	return true
}

func countCantAttackRuleEffects(g *game.Game) int {
	count := 0
	for i := range g.RuleEffects {
		if g.RuleEffects[i].Kind == game.RuleEffectCantAttack {
			count++
		}
	}
	return count
}

// TestPayHandSizeOrCantAttackPaymentAvoidsRestriction proves the triggering
// opponent may pay {X} equal to their hand size; paying installs no restriction
// and taps exactly hand-size lands.
func TestPayHandSizeOrCantAttackPaymentAvoidsRestriction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addChampionsPermanent(g, game.Player1)
	g.Players[game.Player1].IsMonarch = true

	// Player2 (the active opponent) holds two cards, so X = 2.
	addCardToHand(g, game.Player2, greenInstant())
	addCardToHand(g, game.Player2, greenInstant())
	lands := []*game.Permanent{
		addBasicLandPermanent(g, game.Player2, types.Plains),
		addBasicLandPermanent(g, game.Player2, types.Plains),
		addBasicLandPermanent(g, game.Player2, types.Plains),
	}
	attacker := addCombatPermanent(g, game.Player2, greenCreature())

	g.Turn.ActivePlayer = game.Player2
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}}}
	if !resolveChampionsCombatTrigger(engine, g, agents) {
		t.Fatal("beginning-of-combat trigger did not fire while controller was the monarch")
	}

	if got := countCantAttackRuleEffects(g); got != 0 {
		t.Fatalf("can't-attack rule effects = %d, want 0 after payment", got)
	}
	tapped := 0
	for _, land := range lands {
		if land.Tapped {
			tapped++
		}
	}
	if tapped != 2 {
		t.Fatalf("tapped lands = %d, want 2 (hand size)", tapped)
	}
	if !canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player1}) {
		t.Fatal("opponent creature can't attack the monarch after paying")
	}
}

// TestPayHandSizeOrCantAttackDeclineRestrictsAttackingController proves that when
// the triggering opponent declines to pay, their creatures can't attack the
// source's controller ("you") but may still attack other players.
func TestPayHandSizeOrCantAttackDeclineRestrictsAttackingController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addChampionsPermanent(g, game.Player1)
	g.Players[game.Player1].IsMonarch = true

	addCardToHand(g, game.Player2, greenInstant())
	addBasicLandPermanent(g, game.Player2, types.Plains)
	attacker := addCombatPermanent(g, game.Player2, greenCreature())

	g.Turn.ActivePlayer = game.Player2
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}}}
	if !resolveChampionsCombatTrigger(engine, g, agents) {
		t.Fatal("beginning-of-combat trigger did not fire while controller was the monarch")
	}

	if got := countCantAttackRuleEffects(g); got != 1 {
		t.Fatalf("can't-attack rule effects = %d, want 1 after declining", got)
	}
	if canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player1}) {
		t.Fatal("opponent creature may attack the controller after declining to pay")
	}
	if !canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player3}) {
		t.Fatal("opponent creature can't attack a different player after declining to pay")
	}
	if !canAttackWith(g, attacker, game.Player2) {
		t.Fatal("opponent creature can't attack at all after declining to pay")
	}
}

// TestPayHandSizeOrCantAttackDeclineStillAllowsPlaneswalkerAttacks proves the
// official ruling: an opponent who declines to pay can still attack planeswalkers
// the controller controls (and battles they protect), because "can't attack you"
// restricts only direct attacks on the player (CR 508.1).
func TestPayHandSizeOrCantAttackDeclineStillAllowsPlaneswalkerAttacks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addChampionsPermanent(g, game.Player1)
	g.Players[game.Player1].IsMonarch = true
	planeswalker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Controller Planeswalker",
		Types: []types.Card{types.Planeswalker},
	}})

	addCardToHand(g, game.Player2, greenInstant())
	addBasicLandPermanent(g, game.Player2, types.Plains)
	attacker := addCombatPermanent(g, game.Player2, greenCreature())

	g.Turn.ActivePlayer = game.Player2
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}}}
	if !resolveChampionsCombatTrigger(engine, g, agents) {
		t.Fatal("beginning-of-combat trigger did not fire while controller was the monarch")
	}

	if canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player1}) {
		t.Fatal("opponent creature may attack the controller directly after declining to pay")
	}
	if !canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player1, PlaneswalkerID: planeswalker.ObjectID}) {
		t.Fatal("opponent creature can't attack the controller's planeswalker after declining, but the ruling allows it")
	}
}

// TestPayHandSizeOrCantAttackRestrictionAffectsOnlyTriggeringOpponent proves the
// restriction scopes to the triggering opponent's creatures, leaving another
// opponent's creatures free to attack the controller.
func TestPayHandSizeOrCantAttackRestrictionAffectsOnlyTriggeringOpponent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addChampionsPermanent(g, game.Player1)
	g.Players[game.Player1].IsMonarch = true

	addCardToHand(g, game.Player2, greenInstant())
	addBasicLandPermanent(g, game.Player2, types.Plains)
	triggeringAttacker := addCombatPermanent(g, game.Player2, greenCreature())
	otherAttacker := addCombatPermanent(g, game.Player3, greenCreature())

	g.Turn.ActivePlayer = game.Player2
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}}}
	if !resolveChampionsCombatTrigger(engine, g, agents) {
		t.Fatal("beginning-of-combat trigger did not fire while controller was the monarch")
	}

	if canAttackTarget(g, triggeringAttacker, game.AttackTarget{Player: game.Player1}) {
		t.Fatal("the triggering opponent's creature may still attack the controller")
	}
	if !canAttackTarget(g, otherAttacker, game.AttackTarget{Player: game.Player1}) {
		t.Fatal("a non-triggering opponent's creature was wrongly restricted")
	}
}

// TestPayHandSizeOrCantAttackDoesNotFireWhenNotMonarch proves the intervening
// "if you're the monarch" gate stops the trigger when the controller is not the
// monarch, so no restriction is installed.
func TestPayHandSizeOrCantAttackDoesNotFireWhenNotMonarch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addChampionsPermanent(g, game.Player1)
	// Player1 is not the monarch.
	addCardToHand(g, game.Player2, greenInstant())
	attacker := addCombatPermanent(g, game.Player2, greenCreature())

	g.Turn.ActivePlayer = game.Player2
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}}}
	if resolveChampionsCombatTrigger(engine, g, agents) {
		t.Fatal("trigger fired while the controller was not the monarch")
	}

	if got := countCantAttackRuleEffects(g); got != 0 {
		t.Fatalf("can't-attack rule effects = %d, want 0 when the trigger does not fire", got)
	}
	if !canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player1}) {
		t.Fatal("opponent creature was restricted even though the trigger did not fire")
	}
}

// TestPayHandSizeOrCantAttackEmptyHandNeverRestricts proves an empty-handed
// opponent pays {0} for free, so no restriction is installed and no choice is
// offered.
func TestPayHandSizeOrCantAttackEmptyHandNeverRestricts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addChampionsPermanent(g, game.Player1)
	g.Players[game.Player1].IsMonarch = true
	attacker := addCombatPermanent(g, game.Player2, greenCreature())

	g.Turn.ActivePlayer = game.Player2
	// A declining agent would install the restriction if a choice were offered;
	// with an empty hand no choice is offered and no restriction is installed.
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}}}
	if !resolveChampionsCombatTrigger(engine, g, agents) {
		t.Fatal("beginning-of-combat trigger did not fire while controller was the monarch")
	}

	if got := countCantAttackRuleEffects(g); got != 0 {
		t.Fatalf("can't-attack rule effects = %d, want 0 for an empty hand", got)
	}
	if !canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player1}) {
		t.Fatal("empty-handed opponent's creature was wrongly restricted")
	}
}

// TestPayHandSizeOrCantAttackRestrictionExpiresAtEndOfCombat proves the
// "this combat" restriction is removed when combat ends.
func TestPayHandSizeOrCantAttackRestrictionExpiresAtEndOfCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addChampionsPermanent(g, game.Player1)
	g.Players[game.Player1].IsMonarch = true

	addCardToHand(g, game.Player2, greenInstant())
	attacker := addCombatPermanent(g, game.Player2, greenCreature())

	g.Turn.ActivePlayer = game.Player2
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}}}
	if !resolveChampionsCombatTrigger(engine, g, agents) {
		t.Fatal("beginning-of-combat trigger did not fire while controller was the monarch")
	}
	if canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player1}) {
		t.Fatal("restriction was not installed after declining to pay")
	}

	expireEndOfCombatRuleEffects(g)

	if got := countCantAttackRuleEffects(g); got != 0 {
		t.Fatalf("can't-attack rule effects = %d, want 0 after combat ends", got)
	}
	if !canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player1}) {
		t.Fatal("restriction persisted past the end of combat")
	}
}
