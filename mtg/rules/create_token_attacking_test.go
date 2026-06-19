package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func attackingTokenDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Soldier",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Soldier},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

func newlyCreatedToken(g *game.Game) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			return permanent
		}
	}
	return nil
}

// During the active player's combat, a token created with EntryAttacking enters
// the battlefield already attacking a defending player (CR 508.4) and is tapped
// when EntryTapped is also set.
func TestCreateTokenEntryAttackingDuringCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanent(g, game.Player1)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:         game.Fixed(1),
		Source:         game.TokenDef(attackingTokenDef()),
		EntryTapped:    true,
		EntryAttacking: true,
	}, &TurnLog{})

	token := newlyCreatedToken(g)
	if token == nil {
		t.Fatal("EntryAttacking token did not enter the battlefield")
	}
	if !token.Tapped {
		t.Fatal("EntryTapped token entered untapped")
	}
	if len(g.Combat.Attackers) != 2 {
		t.Fatalf("attackers = %+v, want the original attacker plus the new token", g.Combat.Attackers)
	}
	last := g.Combat.Attackers[len(g.Combat.Attackers)-1]
	if last.Attacker != token.ObjectID || last.Target.Player != game.Player2 {
		t.Fatalf("token attack declaration = %+v, want token attacking Player2", last)
	}
}

// Outside combat, the attacking entry is ignored: the token still enters but
// joins no combat.
func TestCreateTokenEntryAttackingOutsideCombatEntersNormally(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:         game.Fixed(1),
		Source:         game.TokenDef(attackingTokenDef()),
		EntryAttacking: true,
	}, &TurnLog{})

	token := newlyCreatedToken(g)
	if token == nil {
		t.Fatal("attacking token did not enter the battlefield outside combat")
	}
	if g.Combat != nil {
		t.Fatalf("combat unexpectedly created outside combat: %+v", g.Combat)
	}
}

// A token created for a player who is not the attacking player cannot enter
// attacking; it joins the battlefield but is not declared as an attacker.
func TestCreateTokenEntryAttackingForNonActivePlayerEntersNormally(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanent(g, game.Player1)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat

	obj := &game.StackObject{Controller: game.Player2}
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:         game.Fixed(1),
		Source:         game.TokenDef(attackingTokenDef()),
		Recipient:      opt.Val(game.ControllerReference()),
		EntryAttacking: true,
	}, &TurnLog{})

	token := newlyCreatedToken(g)
	if token == nil {
		t.Fatal("token did not enter the battlefield")
	}
	if len(g.Combat.Attackers) != 1 {
		t.Fatalf("attackers = %+v, want only the original attacker", g.Combat.Attackers)
	}
}
