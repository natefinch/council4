package game

import (
	"fmt"

	"github.com/natefinch/council4/opt"
)

// GroupReferenceDomain identifies the candidate set a GroupReference draws
// permanents from before its Selection narrows them.
type GroupReferenceDomain int

// Group reference domain values identify supported candidate domains. The zero
// value groupDomainNone is invalid so a zero-value GroupReference never names a
// real group.
const (
	groupDomainNone GroupReferenceDomain = iota

	// GroupDomainBattlefield draws from every permanent on the battlefield.
	GroupDomainBattlefield

	// GroupDomainAttachedObject draws the single permanent that the anchor
	// object is attached to, such as the creature an Equipment equips.
	GroupDomainAttachedObject

	// GroupDomainObjectControlled draws from battlefield permanents controlled
	// by the controller of the anchor object, such as the creatures the
	// defending player controls.
	GroupDomainObjectControlled

	// GroupDomainPlayerControlled draws from battlefield permanents controlled
	// by the player named by the anchor player reference, such as every
	// creature a targeted player controls.
	GroupDomainPlayerControlled

	// GroupDomainSameName draws from every battlefield permanent whose name
	// equals the anchor object's name and satisfies the Selection, such as a
	// targeted permanent and all other permanents with the same name as it
	// (Maelstrom Pulse, the Echoing cycle). The anchor itself is included
	// because it shares its own name, so destroying the whole group covers the
	// "<target> and all other <group> with the same name" wording in one move.
	GroupDomainSameName

	// GroupDomainTriggeringAttackers draws from the creatures that were declared
	// as attackers in the attack that triggered the resolving ability — the
	// "they"/"those creatures" back-reference in "Whenever one or more creatures
	// you control attack, they gain <keyword> until end of turn." (Angelic
	// Guardian). Membership is the set of permanents named by the
	// EventAttackerDeclared events sharing the resolving ability's trigger
	// event's simultaneous batch, narrowed by the Selection (controller and
	// type). Because it binds the declared attackers rather than re-querying the
	// board, a creature that started attacking after the declaration is excluded
	// and a declared attacker that left combat before resolution is still
	// included. It draws no anchor.
	GroupDomainTriggeringAttackers

	// GroupDomainCapturedObjects draws from the permanents a delayed trigger
	// captured at schedule time under its CapturedObjectGroup reference — the
	// "the tokens" back-reference of "... create a token that's a copy of this
	// creature ... Exile the tokens at end of combat." (the myriad keyword,
	// CR 702.116). Membership is the still-on-battlefield permanents named by the
	// resolving stack object's CapturedObjectIDs. It draws no anchor and sets no
	// Selection. Appended last so existing domain ordinals are unchanged.
	GroupDomainCapturedObjects

	// GroupDomainLinkedObjects draws from every permanent a prior instruction in
	// the same resolution remembered under the group's linked key via
	// PublishLinked — the "the tokens" back-reference of "each opponent creates a
	// token that's a copy of it. The tokens are goaded for the rest of the game."
	// (Life of the Party). Membership is the still-on-battlefield permanents named
	// by the source-and-link-scoped linked objects, so it binds exactly the tokens
	// created now rather than re-querying the board. It draws no anchor and sets no
	// Selection.
	GroupDomainLinkedObjects

	// GroupDomainAttackedThisTurn draws from permanents whose current object
	// identity appears in an attacker-declared event during the current turn.
	// Permanents that left the battlefield are absent, and a card that returned as
	// a new object does not inherit its former attack history.
	GroupDomainAttackedThisTurn
)

// GroupReference is pure rules data describing WHERE a mass effect finds a group
// of permanents: a candidate domain, a Selection that narrows it, an optional
// anchor object that the domain is defined relative to, and an optional excluded
// object. Selection still describes WHAT matches; GroupReference describes the
// binding. The zero value is invalid.
type GroupReference struct {
	domain       GroupReferenceDomain
	selection    Selection
	anchor       opt.V[ObjectReference]
	playerAnchor opt.V[PlayerReference]
	exclude      opt.V[ObjectReference]
	// attackedDefenderFilter narrows a GroupDomainTriggeringAttackers group to the
	// declared attackers whose defending player relates to the resolving ability's
	// controller as given ("one or more creatures attack one of your opponents",
	// Frontier Warmonger → TriggerControllerOpponent). TriggerControllerAny (the
	// zero value) imposes no defender restriction, so a plain triggering-attackers
	// group keeps its prior behavior.
	attackedDefenderFilter TriggerControllerFilter
	// linkedKey names the PublishLinked key a GroupDomainLinkedObjects group reads
	// its members from. It is empty for every other domain.
	linkedKey LinkedKey
}

// BattlefieldGroup matches every battlefield permanent satisfying selection.
func BattlefieldGroup(selection Selection) GroupReference {
	return GroupReference{domain: GroupDomainBattlefield, selection: selection}
}

// BattlefieldGroupExcluding matches every battlefield permanent satisfying
// selection except the permanent identified by exclude.
func BattlefieldGroupExcluding(selection Selection, exclude ObjectReference) GroupReference {
	return GroupReference{domain: GroupDomainBattlefield, selection: selection, exclude: opt.Val(exclude)}
}

// GroupRef returns a pointer to a copy of group for specs that hold a group
// reference by pointer to stay within the by-value size budget, mirroring how
// GroupDamageRecipient stores its group.
func GroupRef(group GroupReference) *GroupReference {
	g := group
	return &g
}

// AttachedObjectGroup matches the single permanent that anchor is attached to.
func AttachedObjectGroup(anchor ObjectReference) GroupReference {
	return GroupReference{domain: GroupDomainAttachedObject, anchor: opt.Val(anchor)}
}

// ObjectControlledGroup matches every battlefield permanent controlled by the
// controller of anchor and satisfying selection.
func ObjectControlledGroup(anchor ObjectReference, selection Selection) GroupReference {
	return GroupReference{domain: GroupDomainObjectControlled, selection: selection, anchor: opt.Val(anchor)}
}

// ObjectControlledGroupExcluding matches every battlefield permanent controlled
// by the controller of anchor and satisfying selection, except the permanent
// identified by exclude.
func ObjectControlledGroupExcluding(anchor ObjectReference, selection Selection, exclude ObjectReference) GroupReference {
	return GroupReference{domain: GroupDomainObjectControlled, selection: selection, anchor: opt.Val(anchor), exclude: opt.Val(exclude)}
}

// PlayerControlledGroup matches every battlefield permanent controlled by the
// player named by player and satisfying selection.
func PlayerControlledGroup(player PlayerReference, selection Selection) GroupReference {
	return GroupReference{domain: GroupDomainPlayerControlled, selection: selection, playerAnchor: opt.Val(player)}
}

// PlayerControlledGroupExcluding matches every battlefield permanent controlled
// by the player named by player and satisfying selection, except the permanent
// identified by exclude.
func PlayerControlledGroupExcluding(player PlayerReference, selection Selection, exclude ObjectReference) GroupReference {
	return GroupReference{domain: GroupDomainPlayerControlled, selection: selection, playerAnchor: opt.Val(player), exclude: opt.Val(exclude)}
}

// SameNamePermanentGroup matches every battlefield permanent whose name equals
// anchor's name and satisfying selection. The anchor is the permanent the group
// is defined relative to (a targeted permanent), and it is included in the group
// because it shares its own name.
func SameNamePermanentGroup(anchor ObjectReference, selection Selection) GroupReference {
	return GroupReference{domain: GroupDomainSameName, selection: selection, anchor: opt.Val(anchor)}
}

// TriggeringAttackersGroup matches the creatures declared as attackers in the
// attack that triggered the resolving ability, narrowed by selection. It binds
// the "they"/"those creatures" back-reference of a one-or-more-attackers trigger
// to the specific declared attackers rather than re-querying the board.
func TriggeringAttackersGroup(selection Selection) GroupReference {
	return GroupReference{domain: GroupDomainTriggeringAttackers, selection: selection}
}

// TriggeringAttackersAgainstDefenderGroup is TriggeringAttackersGroup further
// narrowed to the declared attackers whose defending player relates to the
// resolving ability's controller by defenderFilter ("one or more creatures
// attack one of your opponents or a planeswalker they control, those creatures
// gain menace", Frontier Warmonger → TriggerControllerOpponent).
func TriggeringAttackersAgainstDefenderGroup(selection Selection, defenderFilter TriggerControllerFilter) GroupReference {
	return GroupReference{
		domain:                 GroupDomainTriggeringAttackers,
		selection:              selection,
		attackedDefenderFilter: defenderFilter,
	}
}

// CapturedObjectsGroup matches the permanents a delayed trigger captured at
// schedule time under its CapturedObjectGroup reference (the "the tokens"
// back-reference of the myriad keyword's "Exile the tokens at end of combat.").
// It draws no anchor and sets no Selection; membership comes from the resolving
// stack object's captured object IDs.
func CapturedObjectsGroup() GroupReference {
	return GroupReference{domain: GroupDomainCapturedObjects}
}

// LinkedObjectsGroup matches every permanent a prior instruction in the same
// resolution remembered under key via PublishLinked (the "the tokens"
// back-reference of "each opponent creates a token that's a copy of it. The
// tokens are goaded for the rest of the game.", Life of the Party). It draws no
// anchor and sets no Selection; membership comes from the source-and-link-scoped
// linked objects.
func LinkedObjectsGroup(key LinkedKey) GroupReference {
	return GroupReference{domain: GroupDomainLinkedObjects, linkedKey: key}
}

// AttackedThisTurnGroup matches surviving battlefield permanents whose current
// object identity was declared as an attacker earlier this turn.
func AttackedThisTurnGroup(selection Selection) GroupReference {
	return GroupReference{domain: GroupDomainAttackedThisTurn, selection: selection}
}

// Domain reports the candidate domain the group draws from.
func (r GroupReference) Domain() GroupReferenceDomain { return r.domain }

// Empty reports whether this is the omitted zero-value group.
func (r GroupReference) Empty() bool {
	return r.domain == groupDomainNone &&
		r.selection.Empty() &&
		!r.anchor.Exists &&
		!r.playerAnchor.Exists &&
		!r.exclude.Exists &&
		r.attackedDefenderFilter == TriggerControllerAny &&
		r.linkedKey == ""
}

// Selection returns the characteristic predicate that narrows the domain.
func (r GroupReference) Selection() Selection { return r.selection }

// AttackedDefenderFilter returns the defending-player restriction applied to a
// triggering-attackers group, or TriggerControllerAny when unrestricted.
func (r GroupReference) AttackedDefenderFilter() TriggerControllerFilter {
	return r.attackedDefenderFilter
}

// LinkedKey returns the PublishLinked key a GroupDomainLinkedObjects group reads
// its members from, and whether this group draws from linked objects.
func (r GroupReference) LinkedKey() (LinkedKey, bool) {
	return r.linkedKey, r.domain == GroupDomainLinkedObjects
}

// Anchor returns the object the domain is defined relative to, if any.
func (r GroupReference) Anchor() (ObjectReference, bool) {
	return r.anchor.Val, r.anchor.Exists
}

// PlayerAnchor returns the player whose controlled permanents the domain draws
// from, if any.
func (r GroupReference) PlayerAnchor() (PlayerReference, bool) {
	return r.playerAnchor.Val, r.playerAnchor.Exists
}

// Exclusion returns the object dropped from the group, if any.
func (r GroupReference) Exclusion() (ObjectReference, bool) {
	return r.exclude.Val, r.exclude.Exists
}

// Valid reports whether the GroupReference names a structurally sound group.
func (r GroupReference) Valid() bool {
	return len(r.Validate()) == 0
}

// Validate reports structural problems with a GroupReference that represent
// card-definition bugs rather than board-state outcomes.
func (r GroupReference) Validate() []string {
	var problems []string
	switch r.domain {
	case GroupDomainBattlefield:
		if r.anchor.Exists {
			problems = append(problems, "battlefield group must not set an anchor object")
		}
	case GroupDomainAttachedObject:
		if !r.anchor.Exists {
			problems = append(problems, "attached-object group requires an anchor object")
		}
		if !r.selection.Empty() {
			problems = append(problems, "attached-object group must not set a Selection")
		}
		if r.exclude.Exists {
			problems = append(problems, "attached-object group must not set an exclusion")
		}
	case GroupDomainObjectControlled:
		if !r.anchor.Exists {
			problems = append(problems, "object-controlled group requires an anchor object")
		}
	case GroupDomainPlayerControlled:
		if !r.playerAnchor.Exists {
			problems = append(problems, "player-controlled group requires an anchor player")
		}
		if r.anchor.Exists {
			problems = append(problems, "player-controlled group must not set an anchor object")
		}
	case GroupDomainSameName:
		if !r.anchor.Exists {
			problems = append(problems, "same-name group requires an anchor object")
		}
		if r.exclude.Exists {
			problems = append(problems, "same-name group must not set an exclusion")
		}
	case GroupDomainTriggeringAttackers:
		if r.anchor.Exists {
			problems = append(problems, "triggering-attackers group must not set an anchor object")
		}
		if r.exclude.Exists {
			problems = append(problems, "triggering-attackers group must not set an exclusion")
		}
	case GroupDomainCapturedObjects:
		if r.anchor.Exists {
			problems = append(problems, "captured-objects group must not set an anchor object")
		}
		if r.exclude.Exists {
			problems = append(problems, "captured-objects group must not set an exclusion")
		}
		if !r.selection.Empty() {
			problems = append(problems, "captured-objects group must not set a Selection")
		}
	case GroupDomainLinkedObjects:
		if r.linkedKey == "" {
			problems = append(problems, "linked-objects group requires a linked key")
		}
		if r.anchor.Exists {
			problems = append(problems, "linked-objects group must not set an anchor object")
		}
		if r.exclude.Exists {
			problems = append(problems, "linked-objects group must not set an exclusion")
		}
		if !r.selection.Empty() {
			problems = append(problems, "linked-objects group must not set a Selection")
		}
	case GroupDomainAttackedThisTurn:
		if r.anchor.Exists {
			problems = append(problems, "attacked-this-turn group must not set an anchor object")
		}
		if r.playerAnchor.Exists {
			problems = append(problems, "attacked-this-turn group must not set an anchor player")
		}
		if r.exclude.Exists {
			problems = append(problems, "attacked-this-turn group must not set an exclusion")
		}
	case groupDomainNone:
		problems = append(problems, "group reference has no domain")
	default:
		problems = append(problems, fmt.Sprintf("unknown group domain %d", r.domain))
	}
	if r.anchor.Exists {
		problems = appendPrefixed(problems, "anchor", r.anchor.Val.Validate())
	}
	if r.playerAnchor.Exists {
		if r.domain != GroupDomainPlayerControlled {
			problems = append(problems, "only a player-controlled group may set an anchor player")
		}
		problems = appendPrefixed(problems, "player anchor", r.playerAnchor.Val.Validate())
	}
	if r.exclude.Exists {
		problems = appendPrefixed(problems, "exclude", r.exclude.Val.Validate())
	}
	if r.attackedDefenderFilter != TriggerControllerAny && r.domain != GroupDomainTriggeringAttackers {
		problems = append(problems, "only a triggering-attackers group may set a defender filter")
	}
	if r.linkedKey != "" && r.domain != GroupDomainLinkedObjects {
		problems = append(problems, "only a linked-objects group may set a linked key")
	}
	problems = append(problems, r.selection.Validate()...)
	return problems
}

// appendPrefixed appends each problem from src to dst with a "prefix: " label so
// nested reference problems retain their location.
func appendPrefixed(dst []string, prefix string, src []string) []string {
	for _, problem := range src {
		dst = append(dst, prefix+": "+problem)
	}
	return dst
}
