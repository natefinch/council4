// Package action defines player decisions that can be chosen by agents and
// applied by the rules engine.
package action

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// ActionKind identifies which kind of game action an Action represents.
type ActionKind int

const (
	ActionPass ActionKind = iota
	ActionPlayLand
	ActionCastSpell
	ActionActivateAbility
	ActionDeclareAttackers
	ActionDeclareBlockers
)

// Action is a tagged struct representing a single player decision.
type Action struct {
	Kind ActionKind

	PlayLand         PlayLandAction
	CastSpell        CastSpellAction
	ActivateAbility  ActivateAbilityAction
	DeclareAttackers DeclareAttackersAction
	DeclareBlockers  DeclareBlockersAction
}

// PlayLandAction is the payload for playing a land from hand.
type PlayLandAction struct {
	CardID id.ID
}

// CastSpellAction is the payload for casting a spell.
type CastSpellAction struct {
	CardID      id.ID
	Targets     []game.Target
	XValue      int
	ChosenModes []int
}

// ActivateAbilityAction is the payload for activating an ability.
type ActivateAbilityAction struct {
	SourceID     id.ID
	AbilityIndex int
	Targets      []game.Target
	XValue       int
}

// DeclareAttackersAction is the payload for declaring attackers.
type DeclareAttackersAction struct {
	Attackers []game.AttackDeclaration
}

// DeclareBlockersAction is the payload for declaring blockers.
type DeclareBlockersAction struct {
	Blockers []game.BlockDeclaration
}

// Pass creates a pass-priority action.
func Pass() Action {
	return Action{Kind: ActionPass}
}

// PlayLand creates an action to play a land from hand.
func PlayLand(cardID id.ID) Action {
	return Action{
		Kind: ActionPlayLand,
		PlayLand: PlayLandAction{
			CardID: cardID,
		},
	}
}

// CastSpell creates an action to cast a spell.
func CastSpell(cardID id.ID, targets []game.Target, xValue int, chosenModes []int) Action {
	return Action{
		Kind: ActionCastSpell,
		CastSpell: CastSpellAction{
			CardID:      cardID,
			Targets:     targets,
			XValue:      xValue,
			ChosenModes: chosenModes,
		},
	}
}

// ActivateAbility creates an action to activate an ability.
func ActivateAbility(sourceID id.ID, abilityIndex int, targets []game.Target, xValue int) Action {
	return Action{
		Kind: ActionActivateAbility,
		ActivateAbility: ActivateAbilityAction{
			SourceID:     sourceID,
			AbilityIndex: abilityIndex,
			Targets:      targets,
			XValue:       xValue,
		},
	}
}

// DeclareAttackers creates an action to declare attackers.
func DeclareAttackers(attackers []game.AttackDeclaration) Action {
	return Action{
		Kind: ActionDeclareAttackers,
		DeclareAttackers: DeclareAttackersAction{
			Attackers: attackers,
		},
	}
}

// DeclareBlockers creates an action to declare blockers.
func DeclareBlockers(blockers []game.BlockDeclaration) Action {
	return Action{
		Kind: ActionDeclareBlockers,
		DeclareBlockers: DeclareBlockersAction{
			Blockers: blockers,
		},
	}
}
