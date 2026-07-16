package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func inklingAttackRestrictionSource(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Subtype Ward",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantAttack,
				AffectedController: game.ControllerAny,
				PermanentTypes:     []types.Card{types.Creature},
				AffectedSelection:  game.Selection{SubtypesAny: []types.Sub{types.Inkling}},
				DefendingPlayer:    game.PlayerYou,
			}},
		}},
	}})
}

func TestSubtypeAttackRestrictionTracksLiveSourceController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := inklingAttackRestrictionSource(g, game.Player1)
	inkling := addSubtypedCreaturePermanent(g, game.Player2, types.Inkling)
	other := addSubtypedCreaturePermanent(g, game.Player2, types.Spirit)
	walker1 := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "First Walker", Types: []types.Card{types.Planeswalker}, Loyalty: opt.Val(3),
	}})
	walker3 := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name: "Third Walker", Types: []types.Card{types.Planeswalker}, Loyalty: opt.Val(3),
	}})

	protectedPlayer := game.AttackTarget{Player: game.Player1}
	protectedWalker := game.AttackTarget{Player: game.Player1, PlaneswalkerID: walker1.ObjectID}
	otherPlayer := game.AttackTarget{Player: game.Player3}
	otherWalker := game.AttackTarget{Player: game.Player3, PlaneswalkerID: walker3.ObjectID}
	if canAttackTarget(g, inkling, protectedPlayer) ||
		canAttackTarget(g, inkling, protectedWalker) {
		t.Fatal("Inkling could attack the source controller or their planeswalker")
	}
	if !canAttackTarget(g, inkling, otherPlayer) ||
		!canAttackTarget(g, inkling, otherWalker) ||
		!canAttackTarget(g, other, protectedPlayer) {
		t.Fatal("restriction affected an unprotected target or non-Inkling")
	}

	source.Controller = game.Player3
	if !canAttackTarget(g, inkling, protectedPlayer) ||
		!canAttackTarget(g, inkling, protectedWalker) {
		t.Fatal("restriction did not release the old source controller")
	}
	if canAttackTarget(g, inkling, otherPlayer) ||
		canAttackTarget(g, inkling, otherWalker) {
		t.Fatal("restriction did not follow the source's new controller")
	}
}

func TestSubtypeAttackRestrictionRejectsIllegalDeclaration(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	inklingAttackRestrictionSource(g, game.Player1)
	inkling := addSubtypedCreaturePermanent(g, game.Player2, types.Inkling)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}

	illegal, _ := action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: inkling.ObjectID,
		Target:   game.AttackTarget{Player: game.Player1},
	}}).DeclareAttackersPayload()
	if NewEngine(nil).applyDeclareAttackers(g, game.Player2, illegal) {
		t.Fatal("illegal Inkling attack declaration was accepted")
	}
	legal, _ := action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: inkling.ObjectID,
		Target:   game.AttackTarget{Player: game.Player3},
	}}).DeclareAttackersPayload()
	if !NewEngine(nil).applyDeclareAttackers(g, game.Player2, legal) {
		t.Fatal("legal Inkling attack against another opponent was rejected")
	}
}

func TestSubtypeAttackRestrictionRevalidatesAfterControlChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := inklingAttackRestrictionSource(g, game.Player1)
	inkling := addSubtypedCreaturePermanent(g, game.Player2, types.Inkling)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}

	declaration, _ := action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: inkling.ObjectID,
		Target:   game.AttackTarget{Player: game.Player3},
	}}).DeclareAttackersPayload()
	source.Controller = game.Player3
	if NewEngine(nil).applyDeclareAttackers(g, game.Player2, declaration) {
		t.Fatal("declaration stayed legal after the restriction source changed controllers")
	}
}

func TestAttackBatchCreatesCorrelatedTokensForAttackingPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	tokenDef := &game.CardDef{CardFace: game.CardFace{
		Name:            "Inkling",
		Colors:          []color.Color{color.White, color.Black},
		Types:           []types.Card{types.Creature},
		Subtypes:        []types.Sub{types.Inkling},
		Power:           opt.Val(game.PT{Value: 2}),
		Toughness:       opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
	}}
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                    game.EventAttackerDeclared,
		Player:                   game.TriggerPlayerOpponent,
		AttackRecipient:          game.AttackRecipientPlayer,
		OneOrMore:                true,
		OneOrMorePerAttackTarget: true,
	}, []game.Instruction{{Primitive: game.CreateToken{
		Amount:                 game.Fixed(1),
		Source:                 game.TokenDef(tokenDef),
		Recipient:              opt.Val(game.EventPlayerReference()),
		EntryTapped:            true,
		EntryAttackingDefender: opt.Val(game.DefendingPlayerReference()),
	}}}, nil)

	attackers := []*game.Permanent{
		addCombatCreaturePermanent(g, game.Player2),
		addCombatCreaturePermanent(g, game.Player2),
		addCombatCreaturePermanent(g, game.Player2),
		addCombatCreaturePermanent(g, game.Player2),
		addCombatCreaturePermanent(g, game.Player2),
	}
	walker := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name: "Walker", Types: []types.Card{types.Planeswalker}, Loyalty: opt.Val(3),
	}})
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	declarations, _ := action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attackers[0].ObjectID, Target: game.AttackTarget{Player: game.Player3}},
		{Attacker: attackers[1].ObjectID, Target: game.AttackTarget{Player: game.Player3}},
		{Attacker: attackers[2].ObjectID, Target: game.AttackTarget{Player: game.Player4}},
		{Attacker: attackers[3].ObjectID, Target: game.AttackTarget{Player: game.Player3, PlaneswalkerID: walker.ObjectID}},
		{Attacker: attackers[4].ObjectID, Target: game.AttackTarget{Player: game.Player1}},
	}).DeclareAttackersPayload()
	if !engine.applyDeclareAttackers(g, game.Player2, declarations) {
		t.Fatal("attack declaration failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 2 {
		t.Fatalf("stack size = %d, want one trigger for each directly attacked opponent", g.Stack.Size())
	}
	for g.Stack.Size() > 0 {
		engine.resolveTopOfStack(g, &TurnLog{})
	}

	targets := map[game.PlayerID]int{}
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			tokens = append(tokens, permanent)
		}
	}
	if len(tokens) != 2 {
		t.Fatalf("tokens = %d, want 2", len(tokens))
	}
	for _, token := range tokens {
		if token.Controller != game.Player2 || token.Owner != game.Player2 || !token.Tapped {
			t.Fatalf("token state = %#v, want tapped token owned and controlled by attacking player", token)
		}
		def := token.TokenDef
		if def == nil ||
			def.Power.Val.Value != 2 ||
			def.Toughness.Val.Value != 1 ||
			len(def.Colors) != 2 ||
			len(def.Subtypes) != 1 ||
			def.Subtypes[0] != types.Inkling ||
			!hasKeyword(g, token, game.Flying) {
			t.Fatalf("token characteristics = %#v", def)
		}
		declaration, ok := attackDeclarationForAttacker(g, token.ObjectID)
		if !ok || !declaration.Target.IsPlayerAttack() {
			t.Fatalf("token attack declaration = %#v", declaration)
		}
		targets[declaration.Target.Player]++
	}
	if targets[game.Player3] != 1 || targets[game.Player4] != 1 || len(targets) != 2 {
		t.Fatalf("token attack targets = %#v, want Player3 and Player4 once each", targets)
	}
}

func TestAttackBatchOpponentScopeTracksSourceControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                    game.EventAttackerDeclared,
		Player:                   game.TriggerPlayerOpponent,
		AttackRecipient:          game.AttackRecipientPlayer,
		OneOrMore:                true,
		OneOrMorePerAttackTarget: true,
	}, nil, nil)
	cardDef, ok := permanentCardDef(g, source)
	if !ok {
		t.Fatal("trigger source definition missing")
	}
	p := &cardDef.TriggeredAbilities[0].Trigger.Pattern
	event := game.Event{
		Kind:         game.EventAttackerDeclared,
		Controller:   game.Player2,
		Player:       game.Player1,
		AttackTarget: game.AttackTarget{Player: game.Player1},
	}
	if triggerMatchesEvent(g, source, p, event) {
		t.Fatal("attack on current source controller matched opponent-defender scope")
	}
	source.Controller = game.Player3
	if !triggerMatchesEvent(g, source, p, event) {
		t.Fatal("attack on former source controller did not match after control change")
	}
	event.Player = game.Player3
	event.AttackTarget.Player = game.Player3
	if triggerMatchesEvent(g, source, p, event) {
		t.Fatal("attack on new source controller matched after control change")
	}
}
