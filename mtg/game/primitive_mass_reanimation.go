package game

import "errors"

// ChooseCardFromEachGraveyard walks every player in Players (all players or all
// opponents) in APNAP order and, for each, has Chooser pick a card in that
// player's graveyard matching Selection — exactly one when able, or up to one
// when Optional — remembering each chosen card, card-scoped, under LinkedKey. It
// models the mass reanimation base "For each player, choose a creature [or
// planeswalker] card in that player's graveyard." (Breach the Multiverse). The
// choice is a non-targeted choose made as the spell resolves; targeted variants
// ("... choose up to one target creature card ...") are a different mechanism the
// parser must not route here. Each player's graveyard is an independent candidate
// pool, so at most one card per player is chosen, and a player with no eligible
// card contributes none. Nothing moves here: the chosen cards stay in their
// owners' graveyards until a paired ReanimateLinkedCards puts exactly those cards
// onto the battlefield. LinkedKey must be set; the chosen cards are otherwise
// unrecoverable for the reanimation.
type ChooseCardFromEachGraveyard struct {
	Chooser   PlayerReference
	Players   PlayerGroupReference
	Selection Selection
	Optional  bool
	LinkedKey LinkedKey
}

// Kind implements Primitive for ChooseCardFromEachGraveyard.
func (ChooseCardFromEachGraveyard) Kind() PrimitiveKind { return PrimitiveChooseCardFromEachGraveyard }

func (p ChooseCardFromEachGraveyard) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.LinkedKey == "" {
		return errors.New("choose card from each graveyard requires a linked key")
	}
	if err := firstProblem(p.Selection.Validate()); err != nil {
		return err
	}
	if err := validatePlayerGroupReference(p.Players); err != nil {
		return err
	}
	return validatePlayerReference(p.Chooser, targets, checkTargets)
}

func (p ChooseCardFromEachGraveyard) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesLinked: p.LinkedKey}
}

func (ChooseCardFromEachGraveyard) isPrimitive() {}

// ReanimateLinkedCards puts every card a sibling ChooseCardFromEachGraveyard
// remembered under LinkedKey from its owner's graveyard onto the battlefield at
// once under Controller's control, so the returned cards enter as one
// simultaneous batch: their enter-the-battlefield events and replacements resolve
// together and a shared simultaneous-entry id counts them as one group (CR
// 603.6b, 614). It consumes and clears LinkedKey after resolving. It models "Put
// those cards onto the battlefield under your control." LinkedKey must be set.
type ReanimateLinkedCards struct {
	Controller PlayerReference
	LinkedKey  LinkedKey
}

// Kind implements Primitive for ReanimateLinkedCards.
func (ReanimateLinkedCards) Kind() PrimitiveKind { return PrimitiveReanimateLinkedCards }

func (p ReanimateLinkedCards) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.LinkedKey == "" {
		return errors.New("reanimate linked cards requires a linked key")
	}
	return validatePlayerReference(p.Controller, targets, checkTargets)
}

func (p ReanimateLinkedCards) instructionRefs() primitiveRefs {
	return primitiveRefs{consumesLinked: []LinkedKey{p.LinkedKey}}
}

func (ReanimateLinkedCards) isPrimitive() {}
