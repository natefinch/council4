package game

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CardLayout identifies a card layout that affects how printed card faces are
// represented.
type CardLayout string

// Card layout values model the printed layout metadata from card databases.
const (
	LayoutNormal           CardLayout = ""
	LayoutTransform        CardLayout = "transform"
	LayoutModalDFC         CardLayout = "modal_dfc"
	LayoutMeld             CardLayout = "meld"
	LayoutDoubleFacedToken CardLayout = "double_faced_token"
	LayoutReversibleCard   CardLayout = "reversible_card"
)

// FaceIndex identifies one printed face of a card. The zero value is the front
// face so existing single-face cards and actions keep their historical meaning.
type FaceIndex int

// Face index values identify the front and back faces of a card.
const (
	FaceFront FaceIndex = iota
	FaceBack
)

// PT represents a creature's power or toughness. It can be a numeric
// value or a star (*) indicating a characteristic-defining ability (CR 208.2).
type PT struct {
	// Value is the numeric power or toughness. Ignored if IsStar is true.
	Value int

	// IsStar is true if this is a * value determined by a CDA.
	IsStar bool
}

// CardDef is the immutable definition of a Magic card — the "printed" card
// data from the card database. Multiple CardInstances in a game may reference
// the same CardDef.
type CardDef struct {
	// CardFace is the printed front-face characteristics.
	CardFace

	// Layout records the card layout when it changes face behavior.
	// Empty means a normal single-faced card.
	Layout CardLayout

	// ColorIdentity is the card's color identity for Commander deck
	// construction (CR 903.4). Includes colors from mana cost, color
	// indicator, and mana symbols in rules text.
	ColorIdentity color.Identity

	// Back holds the printed back-face characteristics for double-faced cards.
	// The CardDef root fields are the printed front-face characteristics.
	Back opt.V[CardFace]
}

// CardFace is one printed face of a card. It mirrors the printed
// characteristics from CardDef that can differ between faces.
type CardFace struct {
	Name             string
	ManaCost         opt.V[cost.Mana]
	Colors           []color.Color
	Supertypes       []types.Super
	Types            []types.Card
	Subtypes         []types.Sub
	Power            opt.V[PT]
	Toughness        opt.V[PT]
	DynamicPower     opt.V[DynamicValue]
	DynamicToughness opt.V[DynamicValue]
	Loyalty          opt.V[int]
	Defense          opt.V[int]

	SpellAbility         opt.V[SpellAbilityBody]
	ActivatedAbilities   []ActivatedAbilityBody
	ManaAbilities        []ManaAbilityBody
	LoyaltyAbilities     []LoyaltyAbilityBody
	TriggeredAbilities   []TriggeredAbilityBody
	ReplacementAbilities []ReplacementAbilityDef
	StaticAbilities      []StaticAbilityBody

	Abilities        []AbilityDef
	ImplementationID string
	OracleText       string
}

// IsLegendary reports whether this card has the types.Legendary supertype.
func (c *CardDef) IsLegendary() bool {
	return c.HasSupertype(types.Legendary)
}

// HasSupertype reports whether this card has the given supertype.
func (c *CardDef) HasSupertype(supertype types.Super) bool {
	face := c.DefaultFace()
	return face.HasSupertype(supertype)
}

// HasType reports whether this card has the given card type.
func (c *CardDef) HasType(t types.Card) bool {
	face := c.DefaultFace()
	return face.HasType(t)
}

// HasSubtype reports whether this card has the given subtype.
func (c *CardDef) HasSubtype(sub types.Sub) bool {
	face := c.DefaultFace()
	return face.HasSubtype(sub)
}

// HasAnySubtype reports whether this card has any of the given subtypes.
func (c *CardDef) HasAnySubtype(subtypes ...types.Sub) bool {
	face := c.DefaultFace()
	return face.HasAnySubtype(subtypes...)
}

// HasKeyword reports whether any of this card's abilities grants the
// given keyword.
func (c *CardDef) HasKeyword(kw Keyword) bool {
	face := c.DefaultFace()
	return face.HasKeyword(kw)
}

// ManaValue returns the card's mana value from its printed mana cost (CR 202.3).
// Cards with no mana cost, such as lands and source-card-derived no-cost tokens,
// have mana value 0.
func (c *CardDef) ManaValue() int {
	face := c.DefaultFace()
	return face.ManaValue()
}

// IsPermanent reports whether this card becomes a permanent when it resolves
// (i.e., it has at least one permanent card type).
func (c *CardDef) IsPermanent() bool {
	face := c.DefaultFace()
	return face.IsPermanent()
}

// DefaultFace returns the card characteristics used outside the stack and
// battlefield. For double-faced cards, that is the front face.
func (c *CardDef) DefaultFace() CardFace {
	return c.clone()
}

// Face returns the requested printed face. For single-faced cards, FaceFront
// maps to the root card characteristics.
func (c *CardDef) Face(index FaceIndex) (CardFace, bool) {
	switch index {
	case FaceFront:
		return c.clone(), true
	case FaceBack:
		return c.Back.Val, c.Back.Exists
	default:
		return CardFace{}, false
	}
}

// FaceDef returns a CardDef-shaped copy of one face's characteristics. It is a
// bridge for rules helpers that still operate on CardDef values.
func (c *CardDef) FaceDef(index FaceIndex) (*CardDef, bool) {
	face, ok := c.Face(index)
	if !ok {
		return nil, false
	}
	return face.ToCardDef(c), true
}

// FaceIndexes returns the printed faces available on this card.
func (c *CardDef) FaceIndexes() []FaceIndex {
	if c.Back.Exists {
		return []FaceIndex{FaceFront, FaceBack}
	}
	return []FaceIndex{FaceFront}
}

// CanChooseCastFace reports whether this face can be chosen while casting the
// card as a spell. Modal DFCs may choose any non-land face; other layouts cast
// only their front face.
func (c *CardDef) CanChooseCastFace(index FaceIndex) bool {
	face, ok := c.Face(index)
	if !ok || face.HasType(types.Land) {
		return false
	}
	if c.IsModalDoubleFaced() {
		return true
	}
	return index == FaceFront
}

// CanChooseLandFace reports whether this face can be played as a land.
func (c *CardDef) CanChooseLandFace(index FaceIndex) bool {
	face, ok := c.Face(index)
	if !ok || !face.HasType(types.Land) {
		return false
	}
	if !c.Back.Exists {
		return index == FaceFront
	}
	if c.IsModalDoubleFaced() {
		return true
	}
	return index == FaceFront
}

// LegalCastFaces returns all faces that may be chosen while casting this card.
func (c *CardDef) LegalCastFaces() []FaceIndex {
	var faces []FaceIndex
	for _, face := range c.FaceIndexes() {
		if c.CanChooseCastFace(face) {
			faces = append(faces, face)
		}
	}
	return faces
}

// IsModalDoubleFaced reports whether this card is a modal double-faced card.
func (c *CardDef) IsModalDoubleFaced() bool {
	return c.Layout == LayoutModalDFC
}

// IsTransformingDoubleFaced reports whether this card can use transform-style
// face switching. Meld and reversible cards are intentionally excluded.
func (c *CardDef) IsTransformingDoubleFaced() bool {
	return c.Layout == LayoutTransform || c.Layout == LayoutDoubleFacedToken
}

// HasSupertype reports whether this face has the given supertype.
func (f *CardFace) HasSupertype(supertype types.Super) bool {
	return slices.Contains(f.Supertypes, supertype)
}

// HasType reports whether this face has the given card type.
func (f *CardFace) HasType(t types.Card) bool {
	return slices.Contains(f.Types, t)
}

// HasSubtype reports whether this face has the given subtype.
func (f *CardFace) HasSubtype(sub types.Sub) bool {
	return slices.Contains(f.Subtypes, sub)
}

// HasAnySubtype reports whether this face has any of the given subtypes.
func (f *CardFace) HasAnySubtype(subtypes ...types.Sub) bool {
	return slices.ContainsFunc(subtypes, f.HasSubtype)
}

// HasKeyword reports whether any ability on this face grants the given keyword.
func (f *CardFace) HasKeyword(kw Keyword) bool {
	abilities := f.AbilityDefs()
	for i := range abilities {
		if abilities[i].HasKeyword(kw) {
			return true
		}
	}
	return false
}

// AbilityDefs returns all abilities on this face in the legacy AbilityDef view.
func (f *CardFace) AbilityDefs() []AbilityDef {
	if !f.hasCategorizedAbilities() {
		return f.Abilities
	}
	abilities := append([]AbilityDef(nil), f.Abilities...)
	if f.SpellAbility.Exists {
		abilities = append(abilities, spellAbilityDef(&f.SpellAbility.Val))
	}
	for i := range f.ActivatedAbilities {
		abilities = append(abilities, activatedAbilityDef(&f.ActivatedAbilities[i]))
	}
	for i := range f.ManaAbilities {
		abilities = append(abilities, manaAbilityDef(&f.ManaAbilities[i]))
	}
	for i := range f.LoyaltyAbilities {
		abilities = append(abilities, loyaltyAbilityDef(&f.LoyaltyAbilities[i]))
	}
	for i := range f.TriggeredAbilities {
		abilities = append(abilities, triggeredAbilityDef(&f.TriggeredAbilities[i]))
	}
	for i := range f.ReplacementAbilities {
		abilities = append(abilities, replacementAbilityDef(&f.ReplacementAbilities[i]))
	}
	for i := range f.StaticAbilities {
		abilities = append(abilities, staticAbilityDef(&f.StaticAbilities[i]))
	}
	return abilities
}

// ManaValue returns this face's mana value from its printed mana cost (CR 202.3).
// Faces with no mana cost have mana value 0.
func (f *CardFace) ManaValue() int {
	if !f.ManaCost.Exists {
		return 0
	}
	return f.ManaCost.Val.ManaValue()
}

// IsPermanent reports whether this face becomes a permanent when it resolves.
func (f *CardFace) IsPermanent() bool {
	for _, t := range f.Types {
		if t.IsPermanent() {
			return true
		}
	}
	return false
}

// ToCardDef converts a face into a CardDef-shaped value for existing rules
// helpers. ColorIdentity stays on the physical card and is copied from parent.
func (f *CardFace) ToCardDef(parent *CardDef) *CardDef {
	return &CardDef{
		CardFace:      f.clone(),
		ColorIdentity: parent.ColorIdentity,
	}
}

func (f *CardFace) hasCategorizedAbilities() bool {
	return f.SpellAbility.Exists ||
		len(f.ActivatedAbilities) != 0 ||
		len(f.ManaAbilities) != 0 ||
		len(f.LoyaltyAbilities) != 0 ||
		len(f.TriggeredAbilities) != 0 ||
		len(f.ReplacementAbilities) != 0 ||
		len(f.StaticAbilities) != 0
}

func (f *CardFace) clone() CardFace {
	return CardFace{
		Name:                 f.Name,
		ManaCost:             f.ManaCost,
		Colors:               append([]color.Color(nil), f.Colors...),
		Supertypes:           append([]types.Super(nil), f.Supertypes...),
		Types:                append([]types.Card(nil), f.Types...),
		Subtypes:             append([]types.Sub(nil), f.Subtypes...),
		Power:                f.Power,
		Toughness:            f.Toughness,
		DynamicPower:         f.DynamicPower,
		DynamicToughness:     f.DynamicToughness,
		Loyalty:              f.Loyalty,
		Defense:              f.Defense,
		SpellAbility:         f.SpellAbility,
		ActivatedAbilities:   append([]ActivatedAbilityBody(nil), f.ActivatedAbilities...),
		ManaAbilities:        append([]ManaAbilityBody(nil), f.ManaAbilities...),
		LoyaltyAbilities:     append([]LoyaltyAbilityBody(nil), f.LoyaltyAbilities...),
		TriggeredAbilities:   append([]TriggeredAbilityBody(nil), f.TriggeredAbilities...),
		ReplacementAbilities: append([]ReplacementAbilityDef(nil), f.ReplacementAbilities...),
		StaticAbilities:      append([]StaticAbilityBody(nil), f.StaticAbilities...),
		Abilities:            append([]AbilityDef(nil), f.Abilities...),
		ImplementationID:     f.ImplementationID,
		OracleText:           f.OracleText,
	}
}

func spellAbilityDef(body *SpellAbilityBody) AbilityDef {
	ability := AbilityDef{
		Kind:             SpellAbility,
		Text:             body.Text,
		Body:             *body,
		AdditionalCosts:  append([]AdditionalCost(nil), body.AdditionalCosts...),
		AlternativeCosts: append([]AlternativeCost(nil), body.AlternativeCosts...),
		KickerCost:       body.KickerCost,
		KickerEffects:    append([]Effect(nil), body.KickerEffects...),
	}
	applyAbilityContent(&ability, body.Content)
	return ability
}

func activatedAbilityDef(body *ActivatedAbilityBody) AbilityDef {
	ability := AbilityDef{
		Kind:                ActivatedAbility,
		Text:                body.Text,
		Body:                *body,
		ManaCost:            body.ManaCost,
		AdditionalCosts:     append([]AdditionalCost(nil), body.AdditionalCosts...),
		AlternativeCosts:    append([]AlternativeCost(nil), body.AlternativeCosts...),
		ZoneOfFunction:      body.ZoneOfFunction,
		Timing:              body.Timing,
		ActivationCondition: body.ActivationCondition,
	}
	applyAbilityContent(&ability, body.Content)
	return ability
}

func manaAbilityDef(body *ManaAbilityBody) AbilityDef {
	return AbilityDef{
		Kind:                ActivatedAbility,
		Text:                body.Text,
		Body:                *body,
		ManaCost:            body.ManaCost,
		AdditionalCosts:     append([]AdditionalCost(nil), body.AdditionalCosts...),
		ZoneOfFunction:      body.ZoneOfFunction,
		Timing:              body.Timing,
		ActivationCondition: body.ActivationCondition,
		IsManaAbility:       true,
		Effects:             append([]Effect(nil), body.Sequence...),
	}
}

func loyaltyAbilityDef(body *LoyaltyAbilityBody) AbilityDef {
	ability := AbilityDef{
		Kind:                ActivatedAbility,
		Text:                body.Text,
		Body:                *body,
		ActivationCondition: body.ActivationCondition,
		IsLoyaltyAbility:    true,
		LoyaltyCost:         body.LoyaltyCost,
	}
	applyAbilityContent(&ability, body.Content)
	return ability
}

func triggeredAbilityDef(body *TriggeredAbilityBody) AbilityDef {
	ability := AbilityDef{
		Kind:               TriggeredAbility,
		Text:               body.Text,
		Body:               *body,
		Trigger:            opt.Val(body.Trigger),
		Optional:           body.Optional,
		MaxTriggersPerTurn: body.MaxTriggersPerTurn,
	}
	applyAbilityContent(&ability, body.Content)
	return ability
}

func replacementAbilityDef(body *ReplacementAbilityDef) AbilityDef {
	effects := append([]Effect(nil), body.Effects...)
	if body.Replacement.MatchEvent != EventUnknown ||
		body.Replacement.EntersTapped ||
		len(body.Replacement.EntersWithCounters) != 0 ||
		body.Replacement.ReplaceToZone != ZoneNone {
		effects = append(effects, Effect{
			Type:        EffectReplace,
			TargetIndex: TargetIndexController,
			Replacement: opt.Val(body.Replacement),
		})
	}
	return AbilityDef{
		Kind:    StaticAbility,
		Text:    body.Text,
		Effects: effects,
	}
}

func staticAbilityDef(body *StaticAbilityBody) AbilityDef {
	return AbilityDef{
		Kind:             StaticAbility,
		Text:             body.Text,
		Body:             *body,
		Condition:        body.Condition,
		ZoneOfFunction:   body.ZoneOfFunction,
		KeywordAbilities: append([]KeywordAbility(nil), body.KeywordAbilities...),
		Effects:          append([]Effect(nil), body.Effects...),
	}
}

func applyAbilityContent(ability *AbilityDef, content AbilityContent) {
	switch c := content.(type) {
	case PlainAbilityContent:
		ability.Targets = append([]TargetSpec(nil), c.Targets...)
		ability.Effects = append([]Effect(nil), c.Sequence...)
	case ModalAbilityContent:
		ability.Targets = append([]TargetSpec(nil), c.SharedTargets...)
		ability.Modes = append([]Mode(nil), c.Modes...)
		ability.MinModes = c.MinModes
		ability.MaxModes = c.MaxModes
		ability.AllowDuplicateModes = c.AllowDuplicateModes
	case nil:
	default:
		panic("game: unsupported AbilityContent")
	}
}

// CardInstance represents a specific card in a game — one of the 100 cards
// in a player's deck, or a card created during play. Each CardInstance has a
// unique ID and references an immutable CardDef.
type CardInstance struct {
	// ID is the unique identifier for this card instance in the game.
	ID id.ID

	// Def is the static card definition this instance is based on.
	Def *CardDef

	// Owner is the player who owns this card (the player whose deck it
	// started in). Owner never changes during a game (CR 108.3).
	Owner PlayerID
}
