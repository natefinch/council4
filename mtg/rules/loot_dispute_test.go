package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func lootDisputeDragonToken() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:            "Dragon",
		Colors:          []color.Color{color.Red},
		Types:           []types.Card{types.Creature},
		Subtypes:        []types.Sub{types.Dragon},
		Power:           opt.Val(game.PT{Value: 5}),
		Toughness:       opt.Val(game.PT{Value: 5}),
		StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
	}}
}

func addDungeonCompletionDragonTrigger(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Dungeon Celebration",
		Types: []types.Card{types.Enchantment},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event:  game.EventCompletedDungeon,
				Player: game.TriggerPlayerYou,
			}},
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.CreateToken{
				Amount: game.Fixed(1),
				Source: game.TokenDef(lootDisputeDragonToken()),
			}}}}.Ability(),
		}},
	}})
}

func dragonTokensControlledBy(g *game.Game, controller game.PlayerID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == controller && permanent.TokenDef != nil &&
			permanent.TokenDef.Name == "Dragon" {
			count++
		}
	}
	return count
}

func TestAttackInitiativeHolderUsesEventTimeDirectDefenderAndBatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player4, &game.TriggerPattern{
		Event:                    game.EventAttackerDeclared,
		Controller:               game.TriggerControllerYou,
		Player:                   game.TriggerPlayerInitiative,
		AttackRecipient:          game.AttackRecipientPlayer,
		OneOrMore:                true,
		OneOrMorePerAttackTarget: true,
	}, nil, nil)
	// The ability follows the source's live controller when the attack event
	// happens, rather than its owner or original controller.
	source.Controller = game.Player1
	g.Players[game.Player3].HasInitiative = true

	attackers := []*game.Permanent{
		addCombatCreaturePermanent(g, game.Player1),
		addCombatCreaturePermanent(g, game.Player1),
		addCombatCreaturePermanent(g, game.Player1),
		addCombatCreaturePermanent(g, game.Player1),
	}
	walker := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name: "Walker", Types: []types.Card{types.Planeswalker}, Loyalty: opt.Val(3),
	}})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	declarations, _ := action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attackers[0].ObjectID, Target: game.AttackTarget{Player: game.Player3}},
		{Attacker: attackers[1].ObjectID, Target: game.AttackTarget{Player: game.Player3}},
		{Attacker: attackers[2].ObjectID, Target: game.AttackTarget{Player: game.Player3, PlaneswalkerID: walker.ObjectID}},
		{Attacker: attackers[3].ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}).DeclareAttackersPayload()
	if !engine.applyDeclareAttackers(g, game.Player1, declarations) {
		t.Fatal("attack declaration failed")
	}

	// Matching is captured at event time: later initiative and source-control
	// transfers do not rewrite who controlled the already-triggered ability.
	g.Players[game.Player3].HasInitiative = false
	g.Players[game.Player2].HasInitiative = true
	source.Controller = game.Player4
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want one attack-batch trigger", g.Stack.Size())
	}
	obj, _ := g.Stack.Peek()
	if obj.Controller != game.Player1 {
		t.Fatalf("trigger controller = Player%d, want event-time Player1", obj.Controller)
	}

	pattern := &game.TriggerPattern{
		Event:           game.EventAttackerDeclared,
		Controller:      game.TriggerControllerYou,
		Player:          game.TriggerPlayerInitiative,
		AttackRecipient: game.AttackRecipientPlayer,
	}
	source.Controller = game.Player1
	battle := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name: "Battle", Types: []types.Card{types.Battle}, Defense: opt.Val(3),
	}})
	for _, target := range []game.AttackTarget{
		{Player: game.Player2, PlaneswalkerID: walker.ObjectID},
		{Player: game.Player2, BattleID: battle.ObjectID},
	} {
		if triggerMatchesEvent(g, source, pattern, game.Event{
			Kind: game.EventAttackerDeclared, Controller: game.Player1,
			Player: game.Player2, AttackTarget: target,
		}) {
			t.Fatalf("nonplayer attack target %#v matched initiative-holder player trigger", target)
		}
	}
}

func TestCompletedDungeonTriggerIsControllerScopedAndCreatesDragon(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 12)
	addDungeonCompletionDragonTrigger(g, game.Player1)
	addDungeonCompletionDragonTrigger(g, game.Player2)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &ventureChoiceAgent{
		prefer: []string{"Lost Mine of Phandelver", "Goblin Lair", "Dark Pool", "Temple of Dumathoin"},
	}}
	for range 4 {
		if !engine.ventureIntoDungeon(g, game.Player1, agents, &TurnLog{}) {
			t.Fatal("venture failed")
		}
		drainDungeonStack(engine, g, agents)
	}
	if got := dragonTokensControlledBy(g, game.Player1); got != 1 {
		t.Fatalf("Player1 Dragon tokens = %d, want 1", got)
	}
	if got := dragonTokensControlledBy(g, game.Player2); got != 0 {
		t.Fatalf("Player2 Dragon tokens = %d, want 0", got)
	}
}
