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
	LayoutAdventure        CardLayout = "adventure"
	LayoutSplit            CardLayout = "split"
	LayoutPrepare          CardLayout = "prepare"
)

// FaceIndex identifies one printed face of a card. The zero value is the front
// face so existing single-face cards and actions keep their historical meaning.
type FaceIndex int

// Face index values identify the front, back, and alternate faces of a card.
const (
	FaceFront FaceIndex = iota
	FaceBack
	FaceAlternate
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

	// Alternate holds an alternate spell on the front side, used for split cards,
	// adventures, prepared spells, etc.
	Alternate opt.V[CardFace]
}

// CardFace is one printed face of a card. It mirrors the printed
// characteristics from CardDef that can differ between faces.
type CardFace struct {
	Name     string
	ManaCost opt.V[cost.Mana]

	// AdditionalCosts are non-mana costs paid in addition to this face's
	// ManaCost when it is cast as a spell.
	AdditionalCosts []cost.Additional

	// AlternativeCosts replace this face's ManaCost when one is selected.
	AlternativeCosts []cost.Alternative

	// Overload replaces this face's normal spell targets and instructions when
	// the spell is cast for its overload cost.
	Overload opt.V[OverloadAbility]

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
	EntersPrepared   bool

	// SpellAbility is the resolving content of this face when cast as a spell.
	// Its rules text is OracleText.
	SpellAbility         opt.V[AbilityContent]
	ActivatedAbilities   []ActivatedAbility
	ManaAbilities        []ManaAbility
	LoyaltyAbilities     []LoyaltyAbility
	TriggeredAbilities   []TriggeredAbility
	ChapterAbilities     []ChapterAbility
	ReplacementAbilities []ReplacementAbility
	StaticAbilities      []StaticAbility

	ImplementationID string
	OracleText       string
}

// OverloadAbility is the alternate cost and resolving content produced by
// replacing a spell's target wording with the corresponding qualifying group.
type OverloadAbility struct {
	Cost         cost.Mana
	SpellAbility AbilityContent
}

// IsLegendary reports whether this card has the types.Legendary supertype.
func (c *CardDef) IsLegendary() bool {
	return c.HasSupertype(types.Legendary)
}

// HasSupertype reports whether this card has the given supertype.
func (c *CardDef) HasSupertype(supertype types.Super) bool {
	return c.CardFace.HasSupertype(supertype)
}

// HasType reports whether this card has the given card type.
func (c *CardDef) HasType(t types.Card) bool {
	return c.CardFace.HasType(t)
}

// HasSubtype reports whether this card has the given subtype.
func (c *CardDef) HasSubtype(sub types.Sub) bool {
	return c.CardFace.HasSubtype(sub)
}

// HasAnySubtype reports whether this card has any of the given subtypes.
func (c *CardDef) HasAnySubtype(subtypes ...types.Sub) bool {
	return c.CardFace.HasAnySubtype(subtypes...)
}

// HasKeyword reports whether any of this card's abilities grants the
// given keyword.
func (c *CardDef) HasKeyword(kw Keyword) bool {
	return c.CardFace.HasKeyword(kw)
}

// ManaValue returns the card's mana value from its printed mana cost (CR 202.3).
// Cards with no mana cost, such as lands and source-card-derived no-cost tokens,
// have mana value 0.
func (c *CardDef) ManaValue() int {
	return c.CardFace.ManaValue()
}

// IsPermanent reports whether this card becomes a permanent when it resolves
// (i.e., it has at least one permanent card type).
func (c *CardDef) IsPermanent() bool {
	return c.CardFace.IsPermanent()
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
	case FaceAlternate:
		return c.Alternate.Val, c.Alternate.Exists
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

// AlternateFace returns the alternate face for adventure, split, and prepare layouts.
func (c *CardDef) AlternateFace() (CardFace, bool) {
	return c.Alternate.Val, c.Alternate.Exists
}

// FaceIndexes returns the printed faces available on this card.
func (c *CardDef) FaceIndexes() []FaceIndex {
	faces := []FaceIndex{FaceFront}
	if c.Back.Exists {
		faces = append(faces, FaceBack)
	}
	if c.Alternate.Exists {
		faces = append(faces, FaceAlternate)
	}
	return faces
}

// CanChooseCastFace reports whether this face can be chosen while casting the
// card as a spell. Modal DFCs may choose any non-land printed face they expose;
// adventure and split cards may choose their alternate spell face. Prepare
// cards cast copies of their spell face from prepared battlefield permanents.
// Other layouts cast only their front face.
func (c *CardDef) CanChooseCastFace(index FaceIndex) bool {
	face, ok := c.Face(index)
	if !ok || face.HasType(types.Land) {
		return false
	}
	if index == FaceAlternate {
		if !c.Alternate.Exists {
			return false
		}
		switch c.Layout {
		case LayoutAdventure, LayoutSplit:
			return true
		default:
			return false
		}
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
	if f.SpellAbility.Exists && BodyHasKeyword(&f.SpellAbility.Val, kw) {
		return true
	}
	for i := range f.ActivatedAbilities {
		if BodyHasKeyword(&f.ActivatedAbilities[i], kw) {
			return true
		}
	}
	for i := range f.ManaAbilities {
		if BodyHasKeyword(&f.ManaAbilities[i], kw) {
			return true
		}
	}
	for i := range f.LoyaltyAbilities {
		if BodyHasKeyword(&f.LoyaltyAbilities[i], kw) {
			return true
		}
	}
	for i := range f.TriggeredAbilities {
		if BodyHasKeyword(&f.TriggeredAbilities[i], kw) {
			return true
		}
	}
	for i := range f.StaticAbilities {
		if BodyHasKeyword(&f.StaticAbilities[i], kw) {
			return true
		}
	}
	return false
}

// AbilityCount returns the number of abilities on this face in the canonical
// index order (Spell, Activated, Mana, Loyalty, Triggered, Chapter,
// Replacement, Static).
func (f *CardFace) AbilityCount() int {
	n := 0
	if f.SpellAbility.Exists {
		n++
	}
	return n + len(f.ActivatedAbilities) + len(f.ManaAbilities) + len(f.LoyaltyAbilities) + len(f.TriggeredAbilities) + len(f.ChapterAbilities) + len(f.ReplacementAbilities) + len(f.StaticAbilities)
}

// ActivatedAbilityIndex returns the canonical index of an activated ability.
func (f *CardFace) ActivatedAbilityIndex(index int) int {
	if f.SpellAbility.Exists {
		return index + 1
	}
	return index
}

// ManaAbilityIndex returns the canonical index of a mana ability.
func (f *CardFace) ManaAbilityIndex(index int) int {
	return f.ActivatedAbilityIndex(len(f.ActivatedAbilities)) + index
}

// LoyaltyAbilityIndex returns the canonical index of a loyalty ability.
func (f *CardFace) LoyaltyAbilityIndex(index int) int {
	return f.ManaAbilityIndex(len(f.ManaAbilities)) + index
}

// TriggeredAbilityIndex returns the canonical index of a triggered ability.
func (f *CardFace) TriggeredAbilityIndex(index int) int {
	return f.LoyaltyAbilityIndex(len(f.LoyaltyAbilities)) + index
}

// BodyAt returns the ability body at the given canonical index. The canonical
// order is: Spell (if present), Activated, Mana, Loyalty, Triggered, Chapter,
// Replacement, Static. Returns nil for out-of-range indexes.
//
// The returned Ability wraps a POINTER to the addressed slice element (not a
// copy), so it allocates nothing. The pointer aliases into this face's ability
// slices and must be treated as read-only; see the isAbility receivers in
// ability_body.go for the rationale.
func (f *CardFace) BodyAt(index int) Ability {
	if index < 0 {
		return nil
	}
	i := index
	if f.SpellAbility.Exists {
		if i == 0 {
			return &f.SpellAbility.Val
		}
		i--
	}
	if i < len(f.ActivatedAbilities) {
		return &f.ActivatedAbilities[i]
	}
	i -= len(f.ActivatedAbilities)
	if i < len(f.ManaAbilities) {
		return &f.ManaAbilities[i]
	}
	i -= len(f.ManaAbilities)
	if i < len(f.LoyaltyAbilities) {
		return &f.LoyaltyAbilities[i]
	}
	i -= len(f.LoyaltyAbilities)
	if i < len(f.TriggeredAbilities) {
		return &f.TriggeredAbilities[i]
	}
	i -= len(f.TriggeredAbilities)
	if i < len(f.ChapterAbilities) {
		return &f.ChapterAbilities[i]
	}
	i -= len(f.ChapterAbilities)
	if i < len(f.ReplacementAbilities) {
		return &f.ReplacementAbilities[i]
	}
	i -= len(f.ReplacementAbilities)
	if i < len(f.StaticAbilities) {
		return &f.StaticAbilities[i]
	}
	return nil
}

// KickerKeyword returns the first kicker keyword on any ability of this face.
func (f *CardFace) KickerKeyword() (KickerKeyword, bool) {
	for i := range f.ActivatedAbilities {
		if kicker, ok := ActivatedBodyKicker(&f.ActivatedAbilities[i]); ok {
			return kicker, true
		}
	}
	for i := range f.StaticAbilities {
		if ka, ok := BodyKeywordAbility(&f.StaticAbilities[i], Kicker); ok {
			if kicker, ok := ka.(KickerKeyword); ok {
				return kicker, true
			}
		}
	}
	return KickerKeyword{}, false
}

// MutateCost returns the first Mutate cost on this face.
func (f *CardFace) MutateCost() (cost.Mana, bool) {
	for i := range f.StaticAbilities {
		if mutateCost, ok := StaticBodyMutateCost(&f.StaticAbilities[i]); ok {
			return mutateCost, true
		}
	}
	return nil, false
}

// WardKeywords returns all WardKeyword variants on this face.
func (f *CardFace) WardKeywords() []WardKeyword {
	var wards []WardKeyword
	for i := range f.StaticAbilities {
		if ka, ok := BodyKeywordAbility(&f.StaticAbilities[i], Ward); ok {
			if ward, ok := ka.(WardKeyword); ok {
				wards = append(wards, ward)
			}
		}
	}
	return wards
}

// MadnessCost returns the Madness cost if this face has a madness alternative cost.
func (f *CardFace) MadnessCost() (cost.Mana, bool) {
	for i := range f.StaticAbilities {
		if ka, ok := BodyKeywordAbility(&f.StaticAbilities[i], Madness); ok {
			if madness, ok := ka.(MadnessKeyword); ok {
				return madness.Cost, true
			}
		}
	}
	return nil, false
}

// FlashbackCost returns the Flashback alternative cost on this face.
func (f *CardFace) FlashbackCost() (cost.Mana, bool) {
	for i := range f.StaticAbilities {
		if ka, ok := BodyKeywordAbility(&f.StaticAbilities[i], Flashback); ok {
			if flashback, ok := ka.(FlashbackKeyword); ok {
				return flashback.Cost, true
			}
		}
	}
	return nil, false
}

// ClearAbilities removes every categorized ability from this face.
func (f *CardFace) ClearAbilities() {
	f.SpellAbility = opt.V[AbilityContent]{}
	f.ActivatedAbilities = nil
	f.ManaAbilities = nil
	f.LoyaltyAbilities = nil
	f.TriggeredAbilities = nil
	f.ChapterAbilities = nil
	f.ReplacementAbilities = nil
	f.StaticAbilities = nil
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

func (f *CardFace) clone() CardFace {
	return CardFace{
		Name:                 f.Name,
		ManaCost:             f.ManaCost,
		AdditionalCosts:      append([]cost.Additional(nil), f.AdditionalCosts...),
		AlternativeCosts:     cloneAlternativeCosts(f.AlternativeCosts),
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
		EntersPrepared:       f.EntersPrepared,
		SpellAbility:         cloneOptionalAbilityContent(f.SpellAbility),
		Overload:             cloneOverload(f.Overload),
		ActivatedAbilities:   append([]ActivatedAbility(nil), f.ActivatedAbilities...),
		ManaAbilities:        append([]ManaAbility(nil), f.ManaAbilities...),
		LoyaltyAbilities:     append([]LoyaltyAbility(nil), f.LoyaltyAbilities...),
		TriggeredAbilities:   cloneTriggeredAbilities(f.TriggeredAbilities),
		ChapterAbilities:     append([]ChapterAbility(nil), f.ChapterAbilities...),
		ReplacementAbilities: append([]ReplacementAbility(nil), f.ReplacementAbilities...),
		StaticAbilities:      append([]StaticAbility(nil), f.StaticAbilities...),
		ImplementationID:     f.ImplementationID,
		OracleText:           f.OracleText,
	}
}

func cloneOverload(overload opt.V[OverloadAbility]) opt.V[OverloadAbility] {
	if !overload.Exists {
		return opt.V[OverloadAbility]{}
	}
	return opt.Val(OverloadAbility{
		Cost:         slices.Clone(overload.Val.Cost),
		SpellAbility: cloneAbilityContent(overload.Val.SpellAbility),
	})
}

func cloneOptionalAbilityContent(content opt.V[AbilityContent]) opt.V[AbilityContent] {
	if !content.Exists {
		return opt.V[AbilityContent]{}
	}
	return opt.Val(cloneAbilityContent(content.Val))
}

func cloneAbilityContent(content AbilityContent) AbilityContent {
	cloned := content
	cloned.SharedTargets = cloneTargetSpecs(content.SharedTargets)
	cloned.Modes = make([]Mode, len(content.Modes))
	for i := range content.Modes {
		cloned.Modes[i] = content.Modes[i]
		cloned.Modes[i].Targets = cloneTargetSpecs(content.Modes[i].Targets)
		cloned.Modes[i].Sequence = make([]Instruction, len(content.Modes[i].Sequence))
		for j := range content.Modes[i].Sequence {
			cloned.Modes[i].Sequence[j] = content.Modes[i].Sequence[j]
			cloned.Modes[i].Sequence[j].Primitive = clonePrimitive(content.Modes[i].Sequence[j].Primitive)
		}
	}
	return cloned
}

func cloneTriggeredAbilities(abilities []TriggeredAbility) []TriggeredAbility {
	cloned := make([]TriggeredAbility, len(abilities))
	for i := range abilities {
		cloned[i] = abilities[i]
		cloned[i].KeywordAbilities = cloneKeywordAbilities(abilities[i].KeywordAbilities)
		cloned[i].Content = cloneAbilityContent(abilities[i].Content)
	}
	return cloned
}

func cloneKeywordAbilities(abilities []KeywordAbility) []KeywordAbility {
	cloned := make([]KeywordAbility, len(abilities))
	for i, ability := range abilities {
		if ability != nil {
			cloned[i] = ability.cloneKeywordAbility()
		}
	}
	return cloned
}

func cloneTargetSpecs(specs []TargetSpec) []TargetSpec {
	cloned := make([]TargetSpec, len(specs))
	for i := range specs {
		cloned[i] = specs[i]
		cloned[i].Predicate = cloneTargetPredicate(specs[i].Predicate)
		if specs[i].Selection.Exists {
			cloned[i].Selection.Val = cloneSelection(specs[i].Selection.Val)
		}
	}
	return cloned
}

func cloneTargetPredicate(predicate TargetPredicate) TargetPredicate {
	cloned := predicate
	cloned.PermanentTypes = slices.Clone(predicate.PermanentTypes)
	cloned.ExcludedTypes = slices.Clone(predicate.ExcludedTypes)
	cloned.Supertypes = slices.Clone(predicate.Supertypes)
	cloned.Subtypes = slices.Clone(predicate.Subtypes)
	cloned.SpellCardTypes = slices.Clone(predicate.SpellCardTypes)
	cloned.SpellCardTypesAny = slices.Clone(predicate.SpellCardTypesAny)
	cloned.ExcludedSpellCardTypes = slices.Clone(predicate.ExcludedSpellCardTypes)
	cloned.StackObjectKinds = slices.Clone(predicate.StackObjectKinds)
	cloned.SpellSupertypes = slices.Clone(predicate.SpellSupertypes)
	cloned.SpellColors = slices.Clone(predicate.SpellColors)
	cloned.SpellExcludedColors = slices.Clone(predicate.SpellExcludedColors)
	cloned.StackObjectSourceTypes = slices.Clone(predicate.StackObjectSourceTypes)
	cloned.Colors = slices.Clone(predicate.Colors)
	cloned.ExcludedColors = slices.Clone(predicate.ExcludedColors)
	return cloned
}

func cloneSelection(selection Selection) Selection {
	cloned := selection
	cloned.AnyOf = make([]Selection, len(selection.AnyOf))
	for i := range selection.AnyOf {
		cloned.AnyOf[i] = cloneSelection(selection.AnyOf[i])
	}
	cloned.RequiredTypes = slices.Clone(selection.RequiredTypes)
	cloned.RequiredTypesAny = slices.Clone(selection.RequiredTypesAny)
	cloned.ExcludedTypes = slices.Clone(selection.ExcludedTypes)
	cloned.Supertypes = slices.Clone(selection.Supertypes)
	cloned.SubtypesAny = slices.Clone(selection.SubtypesAny)
	cloned.ColorsAny = slices.Clone(selection.ColorsAny)
	cloned.ExcludedColors = slices.Clone(selection.ExcludedColors)
	return cloned
}

func clonePrimitive(primitive Primitive) Primitive {
	switch value := primitive.(type) {
	case Destroy:
		value.Group = cloneGroupReference(value.Group)
		return value
	case Tap:
		value.Group = cloneGroupReference(value.Group)
		return value
	case Untap:
		value.Group = cloneGroupReference(value.Group)
		return value
	case Bounce:
		value.Group = cloneGroupReference(value.Group)
		return value
	case Pay:
		value.Payment = cloneResolutionPayment(value.Payment)
		return value
	case PutOnBattlefield:
		value.Sources = slices.Clone(value.Sources)
		value.ContinuousEffects = slices.Clone(value.ContinuousEffects)
		value.EntryCounters = slices.Clone(value.EntryCounters)
		return value
	case SacrificePermanents:
		value.Selection = cloneSelection(value.Selection)
		return value
	default:
		return primitive
	}
}

func cloneResolutionPayment(payment ResolutionPayment) ResolutionPayment {
	cloned := payment
	if payment.ManaCost.Exists {
		cloned.ManaCost.Val = append(cost.Mana(nil), payment.ManaCost.Val...)
	}
	cloned.AdditionalCosts = append([]cost.Additional(nil), payment.AdditionalCosts...)
	if payment.DynamicGenericManaCost.Exists && payment.DynamicGenericManaCost.Val != nil {
		dynamic := cloneDynamicAmount(payment.DynamicGenericManaCost.Val)
		cloned.DynamicGenericManaCost.Val = &dynamic
	}
	if payment.ManaCostMultiplier.Exists && payment.ManaCostMultiplier.Val != nil {
		dynamic := cloneDynamicAmount(payment.ManaCostMultiplier.Val)
		cloned.ManaCostMultiplier.Val = &dynamic
	}
	return cloned
}

func cloneDynamicAmount(dynamic *DynamicAmount) DynamicAmount {
	cloned := *dynamic
	cloned.Group = cloneGroupReference(dynamic.Group)
	if dynamic.Selection != nil {
		selection := cloneSelection(*dynamic.Selection)
		cloned.Selection = &selection
	}
	if dynamic.Player != nil {
		player := *dynamic.Player
		cloned.Player = &player
	}
	return cloned
}

func cloneGroupReference(group GroupReference) GroupReference {
	group.selection = cloneSelection(group.selection)
	return group
}

func cloneAlternativeCosts(costs []cost.Alternative) []cost.Alternative {
	cloned := make([]cost.Alternative, len(costs))
	for i := range costs {
		cloned[i] = costs[i]
		if costs[i].ManaCost.Exists {
			cloned[i].ManaCost.Val = append(cost.Mana(nil), costs[i].ManaCost.Val...)
		}
		cloned[i].AdditionalCosts = append([]cost.Additional(nil), costs[i].AdditionalCosts...)
	}
	return cloned
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

	// ZoneVersion increments whenever this card changes zones. Event-card
	// references use it to avoid following a card to a new object.
	ZoneVersion uint64
}
