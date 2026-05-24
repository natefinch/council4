package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// CardImplementation is the rules-side escape hatch for cards that cannot be
// expressed with declarative effect primitives yet.
type CardImplementation interface {
	ResolveSpell(ctx *CardContext, obj *game.StackObject, card *game.CardInstance)
}

// CardContext exposes rules mutation helpers to hand-written card
// implementations without exposing the whole game state.
type CardContext struct {
	engine *Engine
	g      *game.Game
	log    *TurnLog
}

// RegisterCardImplementation registers a hand-written card implementation.
func (e *Engine) RegisterCardImplementation(implementationID string, impl CardImplementation) {
	if implementationID == "" {
		panic("rules: card implementation ID must not be empty")
	}
	if impl == nil {
		panic(fmt.Sprintf("rules: card implementation %q is nil", implementationID))
	}
	if e.cardImplementations == nil {
		e.cardImplementations = map[string]CardImplementation{}
	}
	if _, exists := e.cardImplementations[implementationID]; exists {
		panic(fmt.Sprintf("rules: card implementation %q already registered", implementationID))
	}
	e.cardImplementations[implementationID] = impl
}

// DrawCards draws amount cards for player and records the same draw logs and
// game events as declarative draw effects.
func (c *CardContext) DrawCards(player game.PlayerID, amount int) {
	if c == nil || c.engine == nil {
		return
	}
	c.engine.drawCards(c.g, player, amount, c.log)
}

// TargetPlayer returns the stack object's chosen player target at index.
func (c *CardContext) TargetPlayer(obj *game.StackObject, index int) (game.PlayerID, bool) {
	if c == nil || obj == nil || index < 0 || index >= len(obj.Targets) {
		return 0, false
	}
	target := obj.Targets[index]
	if target.Kind != game.TargetPlayer || !isPlayerAlive(c.g, target.PlayerID) {
		return 0, false
	}
	return target.PlayerID, true
}

// TargetPermanentID returns the stack object's chosen permanent target at index.
func (c *CardContext) TargetPermanentID(obj *game.StackObject, index int) (id.ID, bool) {
	if c == nil || obj == nil || index < 0 || index >= len(obj.Targets) {
		return 0, false
	}
	target := obj.Targets[index]
	if target.Kind != game.TargetPermanent || permanentByObjectID(c.g, target.PermanentID) == nil {
		return 0, false
	}
	return target.PermanentID, true
}

// DealPlayerDamageFromStack deals noncombat damage from obj to player.
func (c *CardContext) DealPlayerDamageFromStack(obj *game.StackObject, player game.PlayerID, amount int) int {
	if c == nil || obj == nil {
		return 0
	}
	sourceID, sourceObjectID := damageSourceIDs(c.g, obj)
	return dealPlayerDamage(c.g, sourceID, sourceObjectID, obj.Controller, player, amount, false)
}

// DealPermanentDamageFromStack deals noncombat damage from obj to permanentID.
func (c *CardContext) DealPermanentDamageFromStack(obj *game.StackObject, permanentID id.ID, amount int) int {
	if c == nil || obj == nil {
		return 0
	}
	permanent := permanentByObjectID(c.g, permanentID)
	if permanent == nil {
		return 0
	}
	sourceID, sourceObjectID := damageSourceIDs(c.g, obj)
	return dealPermanentDamage(c.g, sourceID, sourceObjectID, obj.Controller, permanent, amount, false)
}

func (e *Engine) resolveCardImplementationSpell(g *game.Game, obj *game.StackObject, card *game.CardInstance, log *TurnLog) bool {
	if card == nil || card.Def == nil || card.Def.ImplementationID == "" {
		return false
	}
	impl, ok := e.cardImplementations[card.Def.ImplementationID]
	if !ok {
		panic(fmt.Sprintf("rules: card implementation %q is not registered", card.Def.ImplementationID))
	}
	impl.ResolveSpell(&CardContext{engine: e, g: g, log: log}, obj, card)
	return true
}
