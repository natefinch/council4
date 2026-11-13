package game

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrototypeProfile is the inset frame of a prototype card (CR 702.160a,
// 719.1): the prototype keyword ability and the card's secondary mana cost,
// power and toughness. The profile is printed on the card's front face and is
// part of its copiable values (CR 719.2a). It never changes the card's normal
// characteristics; only a spell cast prototyped, and the permanent that spell
// becomes, use it (CR 719.3b, 719.4).
type PrototypeProfile struct {
	ManaCost  cost.Mana
	Power     PT
	Toughness PT
}

// Colors returns the colors of the profile's mana symbols (CR 105.2, 719.3b),
// in WUBRG order.
func (p PrototypeProfile) Colors() []color.Color {
	symbols := p.ManaCost.Colors()
	var colors []color.Color
	for _, candidate := range color.AllColors() {
		if slices.Contains(symbols, mana.Color(candidate.Abbreviation())) {
			colors = append(colors, candidate)
		}
	}
	return colors
}

// Problem describes why the profile is not a printable prototype inset frame
// (CR 719.1), or returns "" for a valid profile. A valid profile has a
// nonempty mana cost of printable mana symbols (CR 107.4) and fixed,
// nonnegative power and toughness. The prototyped cast fails closed on any
// problem rather than offering a spell with a missing or malformed cost.
func (p PrototypeProfile) Problem() string {
	if len(p.ManaCost) == 0 {
		return "prototype mana cost is empty"
	}
	for _, symbol := range p.ManaCost {
		if !printableManaSymbol(symbol) {
			return "prototype mana cost has an invalid symbol " + symbol.String()
		}
	}
	if p.Power.IsStar || p.Toughness.IsStar {
		return "prototype requires fixed power and toughness"
	}
	if p.Power.Value < 0 || p.Toughness.Value < 0 {
		return "prototype power and toughness must be nonnegative"
	}
	return ""
}

// printableManaSymbol reports whether symbol is a mana symbol that can appear
// in a printed mana cost (CR 107.4). Phyrexian generic symbols exist only for
// rules-generated commander tax and are not printable.
func printableManaSymbol(symbol cost.Symbol) bool {
	switch symbol.Kind {
	case cost.GenericSymbol:
		return symbol.Generic >= 0
	case cost.ColoredSymbol, cost.TwobridSymbol, cost.PhyrexianSymbol:
		return isPrintedManaColor(symbol.Color)
	case cost.HybridSymbol:
		return isPrintedManaColor(symbol.Color) && isPrintedManaColor(symbol.AltColor) && symbol.Color != symbol.AltColor
	case cost.ColorlessSymbol, cost.VariableSymbol, cost.SnowSymbol:
		return true
	default:
		return false
	}
}

func isPrintedManaColor(c mana.Color) bool {
	switch c {
	case mana.W, mana.U, mana.B, mana.R, mana.G:
		return true
	default:
		return false
	}
}

// CanCastPrototyped reports whether this card offers the FacePrototype cast
// choice (CR 719.3): a single-faced creature card whose front face has a
// valid prototype profile. An invalid profile is never offered.
func (c *CardDef) CanCastPrototyped() bool {
	return c != nil && c.Layout == LayoutNormal && !c.Back.Exists && !c.Alternate.Exists &&
		c.Prototype.Exists && c.HasType(types.Creature) && !c.DynamicPower.Exists && !c.DynamicToughness.Exists &&
		c.Prototype.Val.Problem() == ""
}

// DeclaresPrototype reports whether any printed face of the card carries a
// prototype profile, valid or not. Catalog admission uses it to validate the
// profile before a deck is loaded.
func (c *CardDef) DeclaresPrototype() bool {
	if c == nil {
		return false
	}
	for _, index := range c.FaceIndexes() {
		if face, ok := c.FaceView(index); ok && face.Prototype.Exists {
			return true
		}
	}
	return false
}

// PhysicalFace returns the printed face of the card that a cast or permanent
// face choice shows. A prototyped spell or permanent is the card's front face
// with its prototype characteristics (CR 719.3b), so permissions and states
// that belong to the physical front face, such as a plotted card's free cast
// from exile, also authorize the prototyped cast (CR 719.3; prototype isn't an
// alternative cost).
func PhysicalFace(index FaceIndex) FaceIndex {
	if index == FacePrototype {
		return FaceFront
	}
	return index
}

// prototypeFace returns the characteristics of a prototyped spell and of the
// permanent it becomes (CR 719.3b, 719.5): the front face with only its mana
// cost, power and toughness replaced by the profile, and its colors set to the
// profile mana cost's colors. Name, types, abilities and every other
// characteristic, including the prototype profile itself (CR 719.2a), are
// unchanged. The returned face shares the immutable front face's slices.
func (c *CardDef) prototypeFace() (CardFace, bool) {
	if !c.CanCastPrototyped() {
		return CardFace{}, false
	}
	face := c.CardFace
	profile := c.Prototype.Val
	face.ManaCost = opt.Val(profile.ManaCost)
	face.Colors = profile.Colors()
	face.Power = opt.Val(profile.Power)
	face.Toughness = opt.Val(profile.Toughness)
	return face, true
}
