package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// enchantPlayerAuraCard builds a minimal "Enchant player" (or, with
// PlayerOpponent, "Enchant opponent") Aura CardDef whose only ability is the
// Enchant keyword. It exercises the player-attachment engine without depending
// on any generated card body beyond the attachment itself.
func enchantPlayerAuraCard(name string, relation game.PlayerRelation) *game.CardDef {
	selection := opt.V[game.Selection]{}
	if relation != game.PlayerAny {
		selection = opt.Val(game.Selection{Player: relation})
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
				Allow:     game.TargetAllowPlayer,
				Selection: selection,
			}}},
		}},
	}}
}

// addEnchantPlayerAura places a resolved Enchant-player Aura permanent on the
// battlefield under controller, ready to be attached to a player.
func addEnchantPlayerAura(g *game.Game, controller game.PlayerID, relation game.PlayerRelation) *game.Permanent {
	return addCombatPermanent(g, controller, enchantPlayerAuraCard("Test Curse", relation))
}

// TestEnchantPlayerAuraResolvesAttachedToTargetPlayer proves an Aura spell that
// enchants a player resolves attached to the chosen player, records the player
// pointer (not a permanent one), and survives the state-based action check.
func TestEnchantPlayerAuraResolvesAttachedToTargetPlayer(t *testing.T) {
	g, engine := setupBestowMain(t)
	auraID := addCardToHand(g, game.Player1, enchantPlayerAuraCard("Test Curse", game.PlayerAny))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	targets := []game.Target{game.PlayerTarget(game.Player2)}
	if !engine.applyAction(g, game.Player1, action.CastSpell(auraID, targets, 0, nil)) {
		t.Fatal("cast of Enchant-player Aura failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	aura, ok := findPermanentByCardID(g, auraID)
	if !ok {
		t.Fatal("Aura did not enter the battlefield")
	}
	if !aura.AttachedToPlayer.Exists || aura.AttachedToPlayer.Val != game.Player2 {
		t.Fatalf("Aura AttachedToPlayer = %v, want Player2", aura.AttachedToPlayer)
	}
	if aura.AttachedTo.Exists {
		t.Fatalf("Aura AttachedTo (permanent) = %v, want none", aura.AttachedTo)
	}

	engine.applyStateBasedActions(g)
	if _, ok := findPermanentByCardID(g, auraID); !ok {
		t.Fatal("Aura left the battlefield after SBA, want it attached to the player")
	}
}

// TestEnchantPlayerAuraFizzlesWhenTargetPlayerEliminated proves an Aura spell
// whose only target is a player who has left the game does not enter the
// battlefield and is put into its owner's graveyard (CR 608.2b).
func TestEnchantPlayerAuraFizzlesWhenTargetPlayerEliminated(t *testing.T) {
	g, engine := setupBestowMain(t)
	auraID := addCardToHand(g, game.Player1, enchantPlayerAuraCard("Test Curse", game.PlayerAny))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	targets := []game.Target{game.PlayerTarget(game.Player2)}
	if !engine.applyAction(g, game.Player1, action.CastSpell(auraID, targets, 0, nil)) {
		t.Fatal("cast failed")
	}
	g.Players[game.Player2].Eliminated = true
	g.TurnOrder.Eliminate(game.Player2)

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := findPermanentByCardID(g, auraID); ok {
		t.Fatal("Aura entered the battlefield despite an illegal target, want graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(auraID) {
		t.Fatal("fizzled Aura not in its owner's graveyard")
	}
}

// TestEnchantPlayerAuraSBAGraveyardsWhenPlayerEliminated proves that when the
// enchanted player leaves the game the Aura becomes illegal and is put into its
// owner's graveyard by the state-based action check (CR 704.5m, CR 800.4a).
func TestEnchantPlayerAuraSBAGraveyardsWhenPlayerEliminated(t *testing.T) {
	g, engine := setupBestowMain(t)
	aura := addEnchantPlayerAura(g, game.Player1, game.PlayerAny)
	if !attachAuraToPlayer(g, aura, game.Player2) {
		t.Fatal("attachAuraToPlayer failed")
	}
	g.Players[game.Player2].Eliminated = true
	g.TurnOrder.Eliminate(game.Player2)

	engine.applyStateBasedActions(g)

	if _, ok := findPermanentByCardID(g, aura.CardInstanceID); ok {
		t.Fatal("Aura still on the battlefield after the enchanted player left, want graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(aura.CardInstanceID) {
		t.Fatal("illegal Aura not in its owner's graveyard")
	}
}

// TestEnchantOpponentAuraSBAGraveyardsAfterControlChange proves an "Enchant
// opponent" Aura becomes illegal when its controller becomes the enchanted
// player, because that player is no longer an opponent (CR 704.5m).
func TestEnchantOpponentAuraSBAGraveyardsAfterControlChange(t *testing.T) {
	g, engine := setupBestowMain(t)
	aura := addEnchantPlayerAura(g, game.Player1, game.PlayerOpponent)
	if !attachAuraToPlayer(g, aura, game.Player2) {
		t.Fatal("attachAuraToPlayer to opponent failed")
	}
	// Player2 gains control of the Aura; it is no longer attached to an opponent.
	aura.Controller = game.Player2

	engine.applyStateBasedActions(g)

	if _, ok := findPermanentByCardID(g, aura.CardInstanceID); ok {
		t.Fatal("Enchant-opponent Aura survived becoming controlled by the enchanted player")
	}
	if !g.Players[aura.Owner].Graveyard.Contains(aura.CardInstanceID) {
		t.Fatal("illegal Aura not in its owner's graveyard")
	}
}

// TestEnchantPlayerAuraReattachMovesBetweenPlayers proves attaching an Aura
// already on one player to another player moves it, clearing the prior pointer.
func TestEnchantPlayerAuraReattachMovesBetweenPlayers(t *testing.T) {
	g, _ := setupBestowMain(t)
	aura := addEnchantPlayerAura(g, game.Player1, game.PlayerAny)
	if !attachAuraToPlayer(g, aura, game.Player2) {
		t.Fatal("first attach failed")
	}
	if !attachAuraToPlayer(g, aura, game.Player3) {
		t.Fatal("reattach failed")
	}
	if !aura.AttachedToPlayer.Exists || aura.AttachedToPlayer.Val != game.Player3 {
		t.Fatalf("AttachedToPlayer = %v, want Player3", aura.AttachedToPlayer)
	}
}

// TestEnchantPlayerAuraDetachClearsPlayerPointer proves detaching an Aura from a
// player clears its player pointer.
func TestEnchantPlayerAuraDetachClearsPlayerPointer(t *testing.T) {
	g, _ := setupBestowMain(t)
	aura := addEnchantPlayerAura(g, game.Player1, game.PlayerAny)
	if !attachAuraToPlayer(g, aura, game.Player2) {
		t.Fatal("attach failed")
	}
	detachPermanent(g, aura)
	if aura.AttachedToPlayer.Exists {
		t.Fatal("AttachedToPlayer still set after detach")
	}
}

// TestEnchantPlayerAuraLeavingBattlefieldClearsPlayerPointer proves that when a
// player-attached Aura leaves the battlefield its player pointer is cleared.
func TestEnchantPlayerAuraLeavingBattlefieldClearsPlayerPointer(t *testing.T) {
	g, _ := setupBestowMain(t)
	aura := addEnchantPlayerAura(g, game.Player1, game.PlayerAny)
	if !attachAuraToPlayer(g, aura, game.Player2) {
		t.Fatal("attach failed")
	}
	movePermanentToZone(g, aura, zone.Graveyard)
	if aura.AttachedToPlayer.Exists {
		t.Fatal("AttachedToPlayer still set after the Aura left the battlefield")
	}
}

// TestEnchantedPlayerAttackedTriggerMatches proves the "Whenever enchanted
// player is attacked" trigger fires for a direct attack on the enchanted player,
// does not fire when a different player is attacked, and does not fire when the
// enchanted player's planeswalker or battle is attacked (an attack on a
// planeswalker or battle is not an attack on the player, CR 508.1).
func TestEnchantedPlayerAttackedTriggerMatches(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	curse := addEnchantPlayerAura(g, game.Player1, game.PlayerAny)
	if !attachAuraToPlayer(g, curse, game.Player2) {
		t.Fatal("attachAuraToPlayer failed")
	}
	pattern := &game.TriggerPattern{
		Event:                                 game.EventAttackerDeclared,
		AttackedPlayerIsSourceEnchantedPlayer: true,
		OneOrMore:                             true,
	}
	attacks := func(target game.AttackTarget) game.Event {
		return game.Event{
			Kind:         game.EventAttackerDeclared,
			Controller:   game.Player3,
			PermanentID:  g.IDGen.Next(),
			AttackTarget: target,
		}
	}

	if !triggerMatchesEvent(g, curse, pattern, attacks(game.AttackTarget{Player: game.Player2})) {
		t.Fatal("trigger did not fire when the enchanted player was attacked")
	}
	if triggerMatchesEvent(g, curse, pattern, attacks(game.AttackTarget{Player: game.Player3})) {
		t.Fatal("trigger wrongly fired when a different player was attacked")
	}
	if triggerMatchesEvent(g, curse, pattern, attacks(game.AttackTarget{Player: game.Player2, PlaneswalkerID: g.IDGen.Next()})) {
		t.Fatal("trigger wrongly fired when the enchanted player's planeswalker was attacked")
	}
	if triggerMatchesEvent(g, curse, pattern, attacks(game.AttackTarget{Player: game.Player2, BattleID: g.IDGen.Next()})) {
		t.Fatal("trigger wrongly fired when the enchanted player's battle was attacked")
	}
}

// TestEnchantedPlayerAttackedTriggerRequiresPlayerAttachment proves the trigger
// never fires from an Aura that is not attached to a player, so an ordinary
// permanent Aura carrying the pattern stays inert.
func TestEnchantedPlayerAttackedTriggerRequiresPlayerAttachment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	curse := addEnchantPlayerAura(g, game.Player1, game.PlayerAny)
	pattern := &game.TriggerPattern{
		Event:                                 game.EventAttackerDeclared,
		AttackedPlayerIsSourceEnchantedPlayer: true,
		OneOrMore:                             true,
	}
	event := game.Event{
		Kind:         game.EventAttackerDeclared,
		Controller:   game.Player3,
		PermanentID:  g.IDGen.Next(),
		AttackTarget: game.AttackTarget{Player: game.Player2},
	}
	if triggerMatchesEvent(g, curse, pattern, event) {
		t.Fatal("trigger fired from an Aura that is not attached to any player")
	}
}

// TestOpponentsAttackingTriggerPlayerGroupResolves proves the reusable
// "opponents attacking the trigger's player" group resolves to each opponent of
// the resolving controller who has a creature attacking the player the trigger's
// attack event named, once per opponent, excluding the controller, attackers on
// other players, and attackers on that player's planeswalker (Curse of Opulence
// "Each opponent attacking that player does the same").
func TestOpponentsAttackingTriggerPlayerGroupResolves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	// The enchanted, attacked player is Player2; Curse of Opulence is controlled
	// by Player1. Player3 attacks Player2 with two creatures (must appear once);
	// Player4 attacks Player2 directly (must appear) and also attacks Player2's
	// planeswalker (must not add Player4 twice). Player1 attacks Player3 (must be
	// excluded as the controller and as an attack on a different player).
	p3a := addCombatCreaturePermanent(g, game.Player3)
	p3b := addCombatCreaturePermanent(g, game.Player3)
	p4a := addCombatCreaturePermanent(g, game.Player4)
	p4pw := addCombatCreaturePermanent(g, game.Player4)
	p1a := addCombatCreaturePermanent(g, game.Player1)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: p3a.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: p3b.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: p4a.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: p4pw.ObjectID, Target: game.AttackTarget{Player: game.Player2, PlaneswalkerID: g.IDGen.Next()}},
		{Attacker: p1a.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
	}}

	obj := &game.StackObject{
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:         game.EventAttackerDeclared,
			AttackTarget: game.AttackTarget{Player: game.Player2},
		},
	}
	members := newReferenceResolver(g, obj).playerGroup(game.OpponentsAttackingTriggerPlayerReference())

	got := make([]bool, game.NumPlayers)
	for _, member := range members {
		got[member] = true
	}
	if len(members) != 2 || !got[game.Player3] || !got[game.Player4] {
		t.Fatalf("group members = %v, want exactly {Player3, Player4}", members)
	}
}

// TestEnchantPlayerAuraCoexistsWithPermanentAttachments proves a player-attached
// Aura coexists with a permanent Aura and an Equipment on a creature: the
// state-based action check leaves all three attached rather than treating the
// player Aura's absent permanent host as an illegal attachment.
func TestEnchantPlayerAuraCoexistsWithPermanentAttachments(t *testing.T) {
	g, engine := setupBestowMain(t)
	creature := addCombatCreaturePermanent(g, game.Player1)
	permAura := addAuraPermanent(g, game.Player1)
	if !attachPermanent(g, permAura, creature) {
		t.Fatal("permanent Aura attach failed")
	}
	equipment := addEquipmentPermanent(g, game.Player1)
	if !attachPermanent(g, equipment, creature) {
		t.Fatal("equipment attach failed")
	}
	playerAura := addEnchantPlayerAura(g, game.Player1, game.PlayerAny)
	if !attachAuraToPlayer(g, playerAura, game.Player2) {
		t.Fatal("player Aura attach failed")
	}

	engine.applyStateBasedActions(g)

	if p, ok := findPermanentByCardID(g, permAura.CardInstanceID); !ok || !p.AttachedTo.Exists {
		t.Fatal("permanent Aura wrongly detached or graveyarded")
	}
	if e, ok := findPermanentByCardID(g, equipment.CardInstanceID); !ok || !e.AttachedTo.Exists {
		t.Fatal("Equipment wrongly detached or graveyarded")
	}
	if pa, ok := findPermanentByCardID(g, playerAura.CardInstanceID); !ok || !pa.AttachedToPlayer.Exists {
		t.Fatal("player Aura wrongly detached or graveyarded")
	}
}
