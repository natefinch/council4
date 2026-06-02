// Package action defines player decisions that can be chosen by agents and
// applied by the rules engine.
package action

import (
	"errors"
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// ActionKind identifies which kind of game action an Action represents.
//
//nolint:revive // ActionKind is the established exported API name.
type ActionKind int

// Action kind values identify the payload carried by an Action.
const (
	ActionPass ActionKind = iota
	ActionPlayLand
	ActionCastSpell
	ActionActivateAbility
	ActionSuspendCard
	ActionCastFaceDown
	ActionTurnFaceUp
	ActionDeclareAttackers
	ActionDeclareBlockers
)

// Action is a tagged struct representing a single player decision.
type Action struct {
	Kind ActionKind

	playLand         PlayLandAction
	castSpell        CastSpellAction
	activateAbility  ActivateAbilityAction
	suspendCard      SuspendCardAction
	castFaceDown     CastFaceDownAction
	turnFaceUp       TurnFaceUpAction
	declareAttackers DeclareAttackersAction
	declareBlockers  DeclareBlockersAction
}

// PlayLandAction is the payload for playing a land from hand.
type PlayLandAction struct {
	CardID id.ID
	Face   game.FaceIndex
}

// CastSpellAction is the payload for casting a spell.
type CastSpellAction struct {
	CardID      id.ID
	SourceZone  game.ZoneType
	Face        game.FaceIndex
	Targets     []game.Target
	XValue      int
	ChosenModes []int
	KickerPaid  bool
}

// ActivateAbilityAction is the payload for activating an ability.
type ActivateAbilityAction struct {
	SourceID     id.ID
	AbilityIndex int
	Targets      []game.Target
	XValue       int
}

// SuspendCardAction is the payload for suspending a card from hand.
type SuspendCardAction struct {
	CardID id.ID
}

// CastFaceDownAction is the payload for casting a card face-down via Morph or
// Disguise.
type CastFaceDownAction struct {
	CardID       id.ID
	Face         game.FaceIndex
	FaceDownKind game.FaceDownKind
}

// TurnFaceUpAction is the payload for turning a face-down permanent face up.
type TurnFaceUpAction struct {
	PermanentID id.ID
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
	return PlayLandFace(cardID, game.FaceFront)
}

// PlayLandFace creates an action to play a specific land face from hand.
func PlayLandFace(cardID id.ID, face game.FaceIndex) Action {
	return Action{
		Kind: ActionPlayLand,
		playLand: PlayLandAction{
			CardID: cardID,
			Face:   face,
		},
	}
}

// CastSpell creates an action to cast a spell.
func CastSpell(cardID id.ID, targets []game.Target, xValue int, chosenModes []int) Action {
	return CastSpellFromZone(cardID, game.ZoneHand, targets, xValue, chosenModes)
}

// CastSpellFromZone creates an action to cast a spell from a specific zone.
func CastSpellFromZone(cardID id.ID, sourceZone game.ZoneType, targets []game.Target, xValue int, chosenModes []int) Action {
	return CastSpellFaceFromZone(cardID, sourceZone, game.FaceFront, targets, xValue, chosenModes)
}

// CastSpellFace creates an action to cast a specific printed face from hand.
func CastSpellFace(cardID id.ID, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int) Action {
	return CastSpellFaceFromZone(cardID, game.ZoneHand, face, targets, xValue, chosenModes)
}

// CastSpellFaceFromZone creates an action to cast a specific printed face from
// a specific source zone.
func CastSpellFaceFromZone(cardID id.ID, sourceZone game.ZoneType, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int) Action {
	return Action{
		Kind: ActionCastSpell,
		castSpell: CastSpellAction{
			CardID:      cardID,
			SourceZone:  sourceZone,
			Face:        face,
			Targets:     copyTargets(targets),
			XValue:      xValue,
			ChosenModes: copyInts(chosenModes),
		},
	}
}

// CastKickedSpell creates an action to cast a spell with kicker paid.
func CastKickedSpell(cardID id.ID, targets []game.Target, xValue int, chosenModes []int) Action {
	return CastKickedSpellFromZone(cardID, game.ZoneHand, targets, xValue, chosenModes)
}

// CastKickedSpellFromZone creates an action to cast a spell from a specific zone
// with kicker paid.
func CastKickedSpellFromZone(cardID id.ID, sourceZone game.ZoneType, targets []game.Target, xValue int, chosenModes []int) Action {
	return CastKickedSpellFaceFromZone(cardID, sourceZone, game.FaceFront, targets, xValue, chosenModes)
}

// CastKickedSpellFaceFromZone creates an action to cast a specific printed face
// from a specific source zone with kicker paid.
func CastKickedSpellFaceFromZone(cardID id.ID, sourceZone game.ZoneType, face game.FaceIndex, targets []game.Target, xValue int, chosenModes []int) Action {
	action := CastSpellFaceFromZone(cardID, sourceZone, face, targets, xValue, chosenModes)
	action.castSpell.KickerPaid = true
	return action
}

// CastCommanderSpell creates an action to cast a commander from the command zone.
func CastCommanderSpell(cardID id.ID, targets []game.Target, xValue int, chosenModes []int) Action {
	return CastSpellFromZone(cardID, game.ZoneCommand, targets, xValue, chosenModes)
}

// ActivateAbility creates an action to activate an ability.
func ActivateAbility(sourceID id.ID, abilityIndex int, targets []game.Target, xValue int) Action {
	return Action{
		Kind: ActionActivateAbility,
		activateAbility: ActivateAbilityAction{
			SourceID:     sourceID,
			AbilityIndex: abilityIndex,
			Targets:      copyTargets(targets),
			XValue:       xValue,
		},
	}
}

// SuspendCard creates an action to exile a card from hand with time counters.
func SuspendCard(cardID id.ID) Action {
	return Action{
		Kind:        ActionSuspendCard,
		suspendCard: SuspendCardAction{CardID: cardID},
	}
}

// CastFaceDown creates an action to cast a card face-down via Morph or Disguise.
func CastFaceDown(cardID id.ID, face game.FaceIndex, kind game.FaceDownKind) Action {
	return Action{
		Kind: ActionCastFaceDown,
		castFaceDown: CastFaceDownAction{
			CardID:       cardID,
			Face:         face,
			FaceDownKind: kind,
		},
	}
}

// TurnFaceUp creates an action to turn a face-down permanent face up.
func TurnFaceUp(permanentID id.ID) Action {
	return Action{
		Kind:       ActionTurnFaceUp,
		turnFaceUp: TurnFaceUpAction{PermanentID: permanentID},
	}
}

// DeclareAttackers creates an action to declare attackers.
func DeclareAttackers(attackers []game.AttackDeclaration) Action {
	return Action{
		Kind: ActionDeclareAttackers,
		declareAttackers: DeclareAttackersAction{
			Attackers: copyAttackers(attackers),
		},
	}
}

// DeclareBlockers creates an action to declare blockers.
func DeclareBlockers(blockers []game.BlockDeclaration) Action {
	return Action{
		Kind: ActionDeclareBlockers,
		declareBlockers: DeclareBlockersAction{
			Blockers: copyBlockers(blockers),
		},
	}
}

// PlayLandPayload returns the play-land payload when this action plays a land.
func (a Action) PlayLandPayload() (PlayLandAction, bool) {
	if a.Kind != ActionPlayLand {
		return PlayLandAction{}, false
	}
	return a.playLand, true
}

// CastSpellPayload returns the cast-spell payload when this action casts a spell.
func (a Action) CastSpellPayload() (CastSpellAction, bool) {
	if a.Kind != ActionCastSpell {
		return CastSpellAction{}, false
	}
	payload := a.castSpell
	payload.Targets = copyTargets(payload.Targets)
	payload.ChosenModes = copyInts(payload.ChosenModes)
	return payload, true
}

// ActivateAbilityPayload returns the activate-ability payload when this action
// activates an ability.
func (a Action) ActivateAbilityPayload() (ActivateAbilityAction, bool) {
	if a.Kind != ActionActivateAbility {
		return ActivateAbilityAction{}, false
	}
	payload := a.activateAbility
	payload.Targets = copyTargets(payload.Targets)
	return payload, true
}

// SuspendCardPayload returns the suspend-card payload when this action suspends
// a card.
func (a Action) SuspendCardPayload() (SuspendCardAction, bool) {
	if a.Kind != ActionSuspendCard {
		return SuspendCardAction{}, false
	}
	return a.suspendCard, true
}

// CastFaceDownPayload returns the face-down cast payload when this action casts
// a card face-down.
func (a Action) CastFaceDownPayload() (CastFaceDownAction, bool) {
	if a.Kind != ActionCastFaceDown {
		return CastFaceDownAction{}, false
	}
	return a.castFaceDown, true
}

// TurnFaceUpPayload returns the turn-face-up payload when this action turns a
// face-down permanent face up.
func (a Action) TurnFaceUpPayload() (TurnFaceUpAction, bool) {
	if a.Kind != ActionTurnFaceUp {
		return TurnFaceUpAction{}, false
	}
	return a.turnFaceUp, true
}

// DeclareAttackersPayload returns the declare-attackers payload when this action
// declares attackers.
func (a Action) DeclareAttackersPayload() (DeclareAttackersAction, bool) {
	if a.Kind != ActionDeclareAttackers {
		return DeclareAttackersAction{}, false
	}
	payload := a.declareAttackers
	payload.Attackers = copyAttackers(payload.Attackers)
	return payload, true
}

// DeclareBlockersPayload returns the declare-blockers payload when this action
// declares blockers.
func (a Action) DeclareBlockersPayload() (DeclareBlockersAction, bool) {
	if a.Kind != ActionDeclareBlockers {
		return DeclareBlockersAction{}, false
	}
	payload := a.declareBlockers
	payload.Blockers = copyBlockers(payload.Blockers)
	return payload, true
}

// Validate reports whether the action has exactly the payload expected by Kind
// and the required payload fields are present.
func (a Action) Validate() error {
	if err := a.validatePayloadIsolation(); err != nil {
		return err
	}
	switch a.Kind {
	case ActionPass:
		return nil
	case ActionPlayLand:
		if a.playLand.CardID == 0 {
			return errors.New("play land action missing card ID")
		}
	case ActionCastSpell:
		if a.castSpell.CardID == 0 {
			return errors.New("cast spell action missing card ID")
		}
		if a.castSpell.SourceZone == game.ZoneNone {
			return errors.New("cast spell action missing source zone")
		}
		if a.castSpell.XValue < 0 {
			return errors.New("cast spell action has negative X value")
		}
	case ActionActivateAbility:
		if a.activateAbility.SourceID == 0 {
			return errors.New("activate ability action missing source ID")
		}
		if a.activateAbility.AbilityIndex < 0 {
			return errors.New("activate ability action has negative ability index")
		}
		if a.activateAbility.XValue < 0 {
			return errors.New("activate ability action has negative X value")
		}
	case ActionSuspendCard:
		if a.suspendCard.CardID == 0 {
			return errors.New("suspend card action missing card ID")
		}
	case ActionCastFaceDown:
		if a.castFaceDown.CardID == 0 {
			return errors.New("cast face-down action missing card ID")
		}
		if a.castFaceDown.FaceDownKind == game.FaceDownNone {
			return errors.New("cast face-down action missing face-down kind")
		}
	case ActionTurnFaceUp:
		if a.turnFaceUp.PermanentID == 0 {
			return errors.New("turn face-up action missing permanent ID")
		}
	case ActionDeclareAttackers:
		for _, attacker := range a.declareAttackers.Attackers {
			if attacker.Attacker == 0 {
				return errors.New("declare attackers action has attacker with zero object ID")
			}
		}
	case ActionDeclareBlockers:
		for _, blocker := range a.declareBlockers.Blockers {
			if blocker.Blocker == 0 || blocker.Blocking == 0 {
				return errors.New("declare blockers action has blocker or blocking object with zero ID")
			}
		}
	default:
		return fmt.Errorf("unknown action kind %d", a.Kind)
	}
	return nil
}

func (a Action) validatePayloadIsolation() error {
	if a.Kind != ActionPlayLand && !playLandActionEmpty(a.playLand) {
		return fmt.Errorf("action kind %d includes play land payload", a.Kind)
	}
	if a.Kind != ActionCastSpell && !castSpellActionEmpty(a.castSpell) {
		return fmt.Errorf("action kind %d includes cast spell payload", a.Kind)
	}
	if a.Kind != ActionActivateAbility && !activateAbilityActionEmpty(a.activateAbility) {
		return fmt.Errorf("action kind %d includes activate ability payload", a.Kind)
	}
	if a.Kind != ActionSuspendCard && !suspendCardActionEmpty(a.suspendCard) {
		return fmt.Errorf("action kind %d includes suspend card payload", a.Kind)
	}
	if a.Kind != ActionCastFaceDown && !castFaceDownActionEmpty(a.castFaceDown) {
		return fmt.Errorf("action kind %d includes cast face-down payload", a.Kind)
	}
	if a.Kind != ActionTurnFaceUp && !turnFaceUpActionEmpty(a.turnFaceUp) {
		return fmt.Errorf("action kind %d includes turn face-up payload", a.Kind)
	}
	if a.Kind != ActionDeclareAttackers && !declareAttackersActionEmpty(a.declareAttackers) {
		return fmt.Errorf("action kind %d includes declare attackers payload", a.Kind)
	}
	if a.Kind != ActionDeclareBlockers && !declareBlockersActionEmpty(a.declareBlockers) {
		return fmt.Errorf("action kind %d includes declare blockers payload", a.Kind)
	}
	return nil
}

func playLandActionEmpty(a PlayLandAction) bool {
	return a.CardID == 0 && a.Face == 0
}

func castSpellActionEmpty(a CastSpellAction) bool {
	return a.CardID == 0 &&
		a.SourceZone == game.ZoneNone &&
		a.Face == 0 &&
		len(a.Targets) == 0 &&
		a.XValue == 0 &&
		len(a.ChosenModes) == 0 &&
		!a.KickerPaid
}

func activateAbilityActionEmpty(a ActivateAbilityAction) bool {
	return a.SourceID == 0 &&
		a.AbilityIndex == 0 &&
		len(a.Targets) == 0 &&
		a.XValue == 0
}

func suspendCardActionEmpty(a SuspendCardAction) bool {
	return a.CardID == 0
}

func castFaceDownActionEmpty(a CastFaceDownAction) bool {
	return a.CardID == 0 && a.Face == 0 && a.FaceDownKind == game.FaceDownNone
}

func turnFaceUpActionEmpty(a TurnFaceUpAction) bool {
	return a.PermanentID == 0
}

func declareAttackersActionEmpty(a DeclareAttackersAction) bool {
	return len(a.Attackers) == 0
}

func declareBlockersActionEmpty(a DeclareBlockersAction) bool {
	return len(a.Blockers) == 0
}

func copyTargets(targets []game.Target) []game.Target {
	return append([]game.Target(nil), targets...)
}

func copyInts(values []int) []int {
	return append([]int(nil), values...)
}

func copyAttackers(attackers []game.AttackDeclaration) []game.AttackDeclaration {
	return append([]game.AttackDeclaration(nil), attackers...)
}

func copyBlockers(blockers []game.BlockDeclaration) []game.BlockDeclaration {
	return append([]game.BlockDeclaration(nil), blockers...)
}
