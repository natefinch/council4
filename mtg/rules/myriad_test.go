package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// myriadCreatureDef is a plain creature that carries the reusable myriad
// triggered body, so a copy of it is itself a myriad creature.
func myriadCreatureDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:               "Myriad Beast",
		Types:              []types.Card{types.Creature},
		Subtypes:           []types.Sub{types.Beast},
		Power:              opt.Val(game.PT{Value: 3}),
		Toughness:          opt.Val(game.PT{Value: 3}),
		TriggeredAbilities: []game.TriggeredAbility{game.MyriadTriggeredBody},
	}}
}

// Resolving the myriad triggered ability while its creature attacks one opponent
// creates, for each other opponent, a tapped token copy that is put onto the
// battlefield attacking that specific opponent (CR 702.116). Those tokens are
// then exiled when the paired end-of-combat delayed trigger fires.
func TestMyriadCreatesTappedAttackingTokenPerOtherOpponentAndExiles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatPermanent(g, game.Player1, myriadCreatureDef())
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
		AttackersDeclared: true,
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat

	obj := &game.StackObject{
		Controller:      game.Player1,
		SourceID:        attacker.ObjectID,
		SourceCardID:    attacker.CardInstanceID,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:           game.EventAttackerDeclared,
			SourceObjectID: attacker.ObjectID,
			Player:         game.Player2,
		},
	}
	// Two "you may" prompts, one per non-defending opponent (Player3, Player4);
	// accept both.
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {1}}},
	}
	log := TurnLog{}

	engine.resolveAbilityContentWithChoices(g, obj, game.MyriadTriggeredBody.Content, agents, &log)

	tokens := myriadTokens(g)
	if len(tokens) != 2 {
		t.Fatalf("created %d token copies, want 2 (one per non-defending opponent)", len(tokens))
	}
	attacked := map[game.PlayerID]bool{}
	for _, token := range tokens {
		if !token.Token {
			t.Fatalf("token %d is not a token", token.ObjectID)
		}
		if !token.Tapped {
			t.Fatalf("token %d entered untapped, want tapped", token.ObjectID)
		}
		def, ok := permanentFaceDef(g, token)
		if !ok || def.Name != "Myriad Beast" {
			t.Fatalf("token %d is not a copy of the source creature: %+v", token.ObjectID, def)
		}
		declaration, ok := attackerDeclarationFor(g, token.ObjectID)
		if !ok {
			t.Fatalf("token %d was not declared attacking", token.ObjectID)
		}
		if declaration.Target.Player == game.Player2 {
			t.Fatalf("token %d attacks the defending player, want another opponent", token.ObjectID)
		}
		attacked[declaration.Target.Player] = true
	}
	if !attacked[game.Player3] || !attacked[game.Player4] {
		t.Fatalf("tokens attack %v, want one attacking Player3 and one attacking Player4", attacked)
	}

	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("scheduled %d delayed triggers, want 1 end-of-combat exile", len(g.DelayedTriggers))
	}
	delayed := g.DelayedTriggers[0]
	if delayed.Timing != game.DelayedAtEndOfCombat {
		t.Fatalf("delayed trigger timing = %v, want DelayedAtEndOfCombat", delayed.Timing)
	}
	if len(delayed.CapturedObjectIDs) != 2 {
		t.Fatalf("captured %d objects, want the 2 created tokens", len(delayed.CapturedObjectIDs))
	}

	// Fire the end-of-combat delayed trigger and resolve its content, exiling the
	// captured tokens.
	exileObj := &game.StackObject{
		Controller:        delayed.Controller,
		SourceID:          delayed.SourceObjectID,
		SourceCardID:      delayed.SourceID,
		CapturedObjectIDs: append([]id.ID(nil), delayed.CapturedObjectIDs...),
	}
	engine.resolveAbilityContentWithChoices(g, exileObj, delayed.Ability.Content, agents, &log)

	for _, token := range tokens {
		if _, ok := permanentByObjectID(g, token.ObjectID); ok {
			t.Fatalf("token %d remained on the battlefield, want exiled at end of combat", token.ObjectID)
		}
	}
}

// TestMyriadTriggerFiresWhenAttackerDeclared proves an intrinsic myriad
// creature's attacks trigger matches on EventAttackerDeclared for itself,
// regardless of which player it attacks (CR 702.116). The pattern is shared with
// the Dethrone and Training keywords, so a match here confirms myriad is wired
// into the same attack-trigger detection path.
func TestMyriadTriggerFiresWhenAttackerDeclared(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, myriadCreatureDef())
	pattern := &game.MyriadTriggeredBody.Trigger.Pattern

	for _, defender := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		event := game.Event{
			Kind:         game.EventAttackerDeclared,
			Controller:   game.Player1,
			PermanentID:  attacker.ObjectID,
			AttackTarget: game.AttackTarget{Player: defender},
		}
		if !triggerMatchesEvent(g, attacker, pattern, event) {
			t.Fatalf("myriad did not trigger when attacking Player%d", defender)
		}
	}

	// It must not trigger for a different creature's attack.
	other := addCombatPermanent(g, game.Player1, myriadCreatureDef())
	event := game.Event{
		Kind:         game.EventAttackerDeclared,
		Controller:   game.Player1,
		PermanentID:  other.ObjectID,
		AttackTarget: game.AttackTarget{Player: game.Player2},
	}
	if triggerMatchesEvent(g, attacker, pattern, event) {
		t.Fatal("myriad wrongly triggered on another creature's attack (TriggerSourceSelf)")
	}
}

// TestGrantedMyriadFiresWhenEquippedCreatureAttacks proves the "equipped creature
// has myriad" grant (Blade of Selves) delivers a functioning myriad: a creature
// that gains game.MyriadTriggeredBody through a continuous AddAbilities effect
// (the exact body the Blade of Selves grant appends) carries the granted trigger
// and fires it when it is declared attacking.
func TestGrantedMyriadFiresWhenEquippedCreatureAttacks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	if got := countGrantedTriggeredAbilities(g, creature); got != 0 {
		t.Fatalf("granted triggered abilities before the grant = %d, want 0", got)
	}

	myriad := game.MyriadTriggeredBody
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.ApplyContinuous{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:        game.LayerAbility,
				Group:        game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
				AddAbilities: []game.Ability{&myriad},
			}},
		}},
	}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := countGrantedTriggeredAbilities(g, creature); got != 1 {
		t.Fatalf("granted triggered abilities after the grant = %d, want 1 (myriad)", got)
	}

	pending := detectTriggeredAbilitiesFromPermanent(g, creature, game.Event{
		Kind:         game.EventAttackerDeclared,
		Controller:   game.Player1,
		PermanentID:  creature.ObjectID,
		AttackTarget: game.AttackTarget{Player: game.Player2},
	})
	if len(pending) != 1 {
		t.Fatalf("granted myriad produced %d pending triggers on attack, want 1", len(pending))
	}
}

func myriadTokens(g *game.Game) []*game.Permanent {
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			tokens = append(tokens, permanent)
		}
	}
	return tokens
}

func attackerDeclarationFor(g *game.Game, objectID id.ID) (game.AttackDeclaration, bool) {
	if g.Combat == nil {
		return game.AttackDeclaration{}, false
	}
	for _, declaration := range g.Combat.Attackers {
		if declaration.Attacker == objectID {
			return declaration, true
		}
	}
	return game.AttackDeclaration{}, false
}
