package game

type damageRecipientKind int

const (
	damageRecipientObject damageRecipientKind = iota
	damageRecipientPlayer
	damageRecipientAnyTarget
	damageRecipientGroup
	damageRecipientPlayerGroup
	damageRecipientAttackedDefender
)

// DamageRecipient is a typed union identifying who receives damage.
// Use ObjectDamageRecipient, PlayerDamageRecipient, AnyTargetDamageRecipient,
// GroupDamageRecipient, or PlayerGroupDamageRecipient to construct.
type DamageRecipient struct {
	set         bool
	kind        damageRecipientKind
	object      ObjectReference
	player      PlayerReference
	group       *GroupReference
	playerGroup PlayerGroupReference
}

// ObjectDamageRecipient creates a recipient for a single permanent.
func ObjectDamageRecipient(object ObjectReference) DamageRecipient {
	return DamageRecipient{set: true, kind: damageRecipientObject, object: object}
}

// PlayerDamageRecipient creates a recipient for a single player.
func PlayerDamageRecipient(player PlayerReference) DamageRecipient {
	return DamageRecipient{set: true, kind: damageRecipientPlayer, player: player}
}

// AnyTargetDamageRecipient creates a recipient for a target slot that may name a
// player or permanent.
func AnyTargetDamageRecipient(targetIndex int) DamageRecipient {
	return DamageRecipient{
		set:    true,
		kind:   damageRecipientAnyTarget,
		object: TargetPermanentReference(targetIndex),
		player: TargetPlayerReference(targetIndex),
	}
}

// GroupDamageRecipient creates a recipient for a group of permanents.
func GroupDamageRecipient(group GroupReference) DamageRecipient {
	g := group
	return DamageRecipient{set: true, kind: damageRecipientGroup, group: &g}
}

// PlayerGroupDamageRecipient creates a recipient for a group of players.
func PlayerGroupDamageRecipient(group PlayerGroupReference) DamageRecipient {
	return DamageRecipient{set: true, kind: damageRecipientPlayerGroup, playerGroup: group}
}

// AttackedDefenderDamageRecipient creates a recipient that resolves to the
// player, planeswalker, or battle that the resolving ability's triggering
// attacker is attacking ("the player or planeswalker it's attacking"). Unlike
// PlayerDamageRecipient(DefendingPlayerReference()), which always addresses the
// defending player, this recipient routes damage to the attacked planeswalker
// or battle when the attack was declared against one, reading the trigger
// event's captured attack target with a live-combat fallback.
func AttackedDefenderDamageRecipient() DamageRecipient {
	return DamageRecipient{set: true, kind: damageRecipientAttackedDefender}
}

// Valid reports whether the recipient identifies a supported target set.
func (r DamageRecipient) Valid() bool {
	if !r.set {
		return false
	}
	switch r.kind {
	case damageRecipientObject:
		return r.object.Kind() != ObjectReferenceNone && len(r.object.Validate()) == 0
	case damageRecipientPlayer:
		return r.player.Kind() != PlayerReferenceNone && len(r.player.Validate()) == 0
	case damageRecipientAnyTarget:
		return r.object.Kind() == ObjectReferenceTargetPermanent &&
			r.player.Kind() == PlayerReferenceTargetPlayer &&
			len(r.object.Validate()) == 0 &&
			len(r.player.Validate()) == 0
	case damageRecipientGroup:
		return r.group != nil && r.group.Valid()
	case damageRecipientPlayerGroup:
		return len(r.playerGroup.Validate()) == 0
	case damageRecipientAttackedDefender:
		return true
	default:
		return false
	}
}

// ObjectReference returns the permanent reference when this recipient addresses one permanent.
func (r DamageRecipient) ObjectReference() (ObjectReference, bool) {
	if !r.Valid() || r.kind != damageRecipientObject {
		return ObjectReference{}, false
	}
	return r.object, true
}

// PlayerReference returns the player reference when this recipient addresses one player.
func (r DamageRecipient) PlayerReference() (PlayerReference, bool) {
	if !r.Valid() || r.kind != damageRecipientPlayer {
		return PlayerReference{}, false
	}
	return r.player, true
}

// GroupReference returns the group reference when this recipient addresses a permanent group.
func (r DamageRecipient) GroupReference() (GroupReference, bool) {
	if !r.Valid() || r.kind != damageRecipientGroup {
		return GroupReference{}, false
	}
	return *r.group, true
}

// PlayerGroupReference returns the player-group reference when this recipient addresses a player group.
func (r DamageRecipient) PlayerGroupReference() (PlayerGroupReference, bool) {
	if !r.Valid() || r.kind != damageRecipientPlayerGroup {
		return PlayerGroupReference{}, false
	}
	return r.playerGroup, true
}

// IsAttackedDefender reports whether this recipient addresses the player,
// planeswalker, or battle the triggering attacker is attacking.
func (r DamageRecipient) IsAttackedDefender() bool {
	return r.Valid() && r.kind == damageRecipientAttackedDefender
}

// AnyTargetObjectReference returns the permanent reference when this recipient addresses any target.
func (r DamageRecipient) AnyTargetObjectReference() (ObjectReference, bool) {
	if !r.Valid() || r.kind != damageRecipientAnyTarget {
		return ObjectReference{}, false
	}
	return r.object, true
}

// AnyTargetPlayerReference returns the player reference when this recipient addresses any target.
func (r DamageRecipient) AnyTargetPlayerReference() (PlayerReference, bool) {
	if !r.Valid() || r.kind != damageRecipientAnyTarget {
		return PlayerReference{}, false
	}
	return r.player, true
}

type tokenSourceKind int

const (
	tokenSourceDef tokenSourceKind = iota
	tokenSourceCopy
)

// TokenSource is a mutually-exclusive union for the token definition.
// Use TokenDef or TokenCopyOf to construct.
type TokenSource struct {
	set  bool
	kind tokenSourceKind
	def  *CardDef
	copy TokenCopySpec
}

// TokenDef creates a TokenSource using an explicit CardDef.
func TokenDef(def *CardDef) TokenSource {
	return TokenSource{set: true, kind: tokenSourceDef, def: def}
}

// TokenCopyOf creates a TokenSource using a TokenCopySpec (copy-of-something).
func TokenCopyOf(spec TokenCopySpec) TokenSource {
	return TokenSource{set: true, kind: tokenSourceCopy, copy: spec}
}

// Valid reports whether the source identifies a concrete token definition.
func (s TokenSource) Valid() bool {
	if !s.set {
		return false
	}
	switch s.kind {
	case tokenSourceDef:
		return s.def != nil
	case tokenSourceCopy:
		return s.copy.Source != TokenCopySourceNone
	default:
		return false
	}
}

// TokenDefRef returns the token CardDef when this source uses an explicit definition.
func (s TokenSource) TokenDefRef() (*CardDef, bool) {
	if !s.Valid() || s.kind != tokenSourceDef {
		return nil, false
	}
	return s.def, true
}

// TokenCopy returns the TokenCopySpec when this source copies another object/card.
func (s TokenSource) TokenCopy() (TokenCopySpec, bool) {
	if !s.Valid() || s.kind != tokenSourceCopy {
		return TokenCopySpec{}, false
	}
	return s.copy, true
}

type battlefieldSourceKind int

const (
	battlefieldSourceCard battlefieldSourceKind = iota
	battlefieldSourceLinked
)

// BattlefieldSource identifies what card or object to put on the battlefield.
// Use CardBattlefieldSource or LinkedBattlefieldSource to construct.
type BattlefieldSource struct {
	set    bool
	kind   battlefieldSourceKind
	card   CardReference
	linked LinkedKey
}

// CardBattlefieldSource creates a source referencing a specific card.
func CardBattlefieldSource(ref CardReference) BattlefieldSource {
	return BattlefieldSource{set: true, kind: battlefieldSourceCard, card: ref}
}

// LinkedBattlefieldSource creates a source referencing an object linked by key.
func LinkedBattlefieldSource(key LinkedKey) BattlefieldSource {
	return BattlefieldSource{set: true, kind: battlefieldSourceLinked, linked: key}
}

// Valid reports whether the source identifies a concrete card or linked object.
func (s BattlefieldSource) Valid() bool {
	if !s.set {
		return false
	}
	switch s.kind {
	case battlefieldSourceCard:
		return s.card.Kind != CardReferenceNone
	case battlefieldSourceLinked:
		return s.linked != ""
	default:
		return false
	}
}

// CardRef returns the direct card reference when this source names a specific card.
func (s BattlefieldSource) CardRef() (CardReference, bool) {
	if !s.Valid() || s.kind != battlefieldSourceCard {
		return CardReference{}, false
	}
	return s.card, true
}

// LinkedKey returns the linked-object key when this source uses linked objects.
func (s BattlefieldSource) LinkedKey() (LinkedKey, bool) {
	if !s.Valid() || s.kind != battlefieldSourceLinked {
		return "", false
	}
	return s.linked, true
}

// sourceLinkedKey returns the LinkedKey if this is a linked source; otherwise empty.
func (s BattlefieldSource) sourceLinkedKey() LinkedKey {
	key, ok := s.LinkedKey()
	if !ok {
		return ""
	}
	return key
}
