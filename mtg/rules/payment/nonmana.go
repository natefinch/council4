package payment

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// maxNonManaConversions bounds the monocolored hybrid and generic Phyrexian
// symbols whose generic payment a nonmana payment mechanic tries; any further
// such symbol is left for the remaining mana payment.
const maxNonManaConversions = 8

// nonManaPaymentCost is a spell's total cost (CR 601.2f) being paid in part by
// a nonmana payment mechanic: convoke (CR 702.51a), improvise (CR 702.126a) or
// delve (CR 702.66a). Each mechanic pays for generic mana in the total cost and
// convoke also for its colored mana. Every other symbol, such as {C}, {S} or a
// Phyrexian symbol announced as 2 life, stays in the remaining cost with its
// announced payment (CR 601.2b) for the mana payment that follows.
type nonManaPaymentCost struct {
	// symbols are the cost's effective symbols (see effectiveManaSymbols), the
	// order in which the Phyrexian preferences apply.
	symbols []cost.Symbol
	// phyrexian holds the announced payment of each Phyrexian symbol.
	phyrexian []phyrexianAnnouncement
	// removed marks symbols no longer in the remaining cost: generic and X
	// symbols counted in generic, symbols paid by the mechanic and symbols
	// converted to generic mana.
	removed []bool
	// generic is the generic mana in the total cost and reduction the amount
	// of it the mechanic pays.
	generic   int
	reduction int
	prefs     *Preferences
}

type phyrexianAnnouncement struct {
	payLife, manaOnly bool
}

// newNonManaPaymentCost prepares manaCost, with X announced as xValue, to be
// paid in part by a nonmana mechanic. prefs holds the caster's announced
// Phyrexian payments, in effective symbol order; it is read, never advanced.
func newNonManaPaymentCost(s State, playerID game.PlayerID, manaCost *cost.Mana, xValue int, prefs *Preferences) (*nonManaPaymentCost, bool) {
	if xValue < 0 {
		return nil, false
	}
	var symbols []cost.Symbol
	if manaCost != nil {
		symbols = effectiveManaSymbols(s, playerID, *manaCost)
	}
	c := &nonManaPaymentCost{
		symbols:   symbols,
		phyrexian: make([]phyrexianAnnouncement, len(symbols)),
		removed:   make([]bool, len(symbols)),
		prefs:     prefs,
	}
	offset := 0
	for i, symbol := range symbols {
		switch symbol.Kind {
		case cost.GenericSymbol:
			c.generic += symbol.Generic
			c.removed[i] = true
		case cost.VariableSymbol:
			c.generic += xValue
			c.removed[i] = true
		case cost.PhyrexianSymbol, cost.PhyrexianGenericSymbol:
			payLife, manaOnly := prefs.phyrexianChoiceAt(offset)
			c.phyrexian[i] = phyrexianAnnouncement{payLife: payLife, manaOnly: manaOnly}
			offset++
		default:
		}
	}
	return c, true
}

// convertible returns the remaining symbols that may be paid as generic mana:
// a monocolored hybrid symbol's {2} half (CR 107.4e) and a generic Phyrexian
// symbol not announced as 2 life.
func (c *nonManaPaymentCost) convertible() []int {
	var indices []int
	for i, symbol := range c.symbols {
		if c.removed[i] || len(indices) == maxNonManaConversions {
			continue
		}
		switch symbol.Kind {
		case cost.TwobridSymbol:
			indices = append(indices, i)
		case cost.PhyrexianGenericSymbol:
			if !c.phyrexian[i].payLife {
				indices = append(indices, i)
			}
		default:
		}
	}
	return indices
}

// convert pays the convertible symbols selected by mask's bits as generic
// mana. It reports false when mask selects a symbol that isn't convertible, so
// callers enumerating masks skip it.
func (c *nonManaPaymentCost) convert(mask int) bool {
	indices := c.convertible()
	if mask >= 1<<len(indices) {
		return false
	}
	for bit, i := range indices {
		if mask&(1<<bit) == 0 {
			continue
		}
		c.removed[i] = true
		if c.symbols[i].Kind == cost.TwobridSymbol {
			c.generic += 2
		} else {
			c.generic += c.symbols[i].Generic
		}
	}
	return true
}

func (c *nonManaPaymentCost) genericDue() int {
	return max(c.generic-c.reduction, 0)
}

// remaining returns the cost left after the mechanic's payment, its unpaid
// generic mana first and every other symbol in order, and the preferences
// that pay it: the announced payments of the Phyrexian symbols it keeps, in
// the same order.
func (c *nonManaPaymentCost) remaining() (*cost.Mana, *Preferences) {
	var modified cost.Mana
	if due := c.genericDue(); due > 0 {
		modified = append(modified, cost.O(due))
	}
	var kept []int
	offset := 0
	for i, symbol := range c.symbols {
		phyrexian := symbol.Kind == cost.PhyrexianSymbol || symbol.Kind == cost.PhyrexianGenericSymbol
		if !c.removed[i] {
			modified = append(modified, symbol)
			if phyrexian {
				kept = append(kept, offset)
			}
		}
		if phyrexian {
			offset++
		}
	}
	return &modified, c.prefs.withPhyrexianChoices(kept)
}

// nonManaConversionCount is the number of convertible symbols in manaCost; a
// mechanic tries each subset of them paid as generic mana.
func nonManaConversionCount(s State, playerID game.PlayerID, manaCost *cost.Mana, xValue int, prefs *Preferences) int {
	c, ok := newNonManaPaymentCost(s, playerID, manaCost, xValue, prefs)
	if !ok {
		return 0
	}
	return len(c.convertible())
}

// convokePayment taps creatures for c's colored mana (CR 702.51a), converting
// the convertible symbols mask selects to generic mana, and returns those taps
// with the creatures left to pay for its generic mana. Each tapped creature
// pays for one generic mana or one colored mana of its color: a colored
// symbol, either half of a hybrid symbol, the colored half of a monocolored
// hybrid symbol or a Phyrexian symbol announced as mana. A Phyrexian symbol
// announced as 2 life keeps that payment. One whose payment wasn't announced is
// paid by a creature only when enough other creatures remain for the generic
// mana, so it can still be paid with life. It reports false when mask selects
// a symbol that isn't convertible.
func convokePayment(s State, playerID game.PlayerID, c *nonManaPaymentCost, exclude map[id.ID]bool, mask int) (taps, rest []*game.Permanent, ok bool) {
	candidates := convokeCandidates(s, playerID, exclude)
	used := make(map[id.ID]bool)
	tapForColor := func(i int) {
		for _, color := range c.symbols[i].Colors() {
			permanent, ok := chooseConvokeColoredCreature(s, candidates, used, color)
			if !ok {
				continue
			}
			taps = append(taps, permanent)
			used[permanent.ObjectID] = true
			c.removed[i] = true
			return
		}
	}
	for _, pays := range []func(i int, symbol cost.Symbol) bool{
		func(_ int, symbol cost.Symbol) bool { return symbol.Kind == cost.ColoredSymbol },
		func(_ int, symbol cost.Symbol) bool { return symbol.Kind == cost.HybridSymbol },
		func(i int, symbol cost.Symbol) bool {
			return symbol.Kind == cost.PhyrexianSymbol && c.phyrexian[i].manaOnly && !c.phyrexian[i].payLife
		},
		func(_ int, symbol cost.Symbol) bool { return symbol.Kind == cost.TwobridSymbol },
	} {
		for i, symbol := range c.symbols {
			if !c.removed[i] && pays(i, symbol) {
				tapForColor(i)
			}
		}
	}
	if !c.convert(mask) {
		return nil, nil, false
	}
	unused := func() int {
		count := 0
		for _, permanent := range candidates {
			if !used[permanent.ObjectID] {
				count++
			}
		}
		return count
	}
	for i, symbol := range c.symbols {
		if c.removed[i] || symbol.Kind != cost.PhyrexianSymbol || c.phyrexian[i].payLife || c.phyrexian[i].manaOnly {
			continue
		}
		if unused() <= c.genericDue() {
			break
		}
		tapForColor(i)
	}
	for _, permanent := range candidates {
		if !used[permanent.ObjectID] {
			rest = append(rest, permanent)
		}
	}
	return taps, rest, true
}

func chooseConvokeColoredCreature(s State, candidates []*game.Permanent, used map[id.ID]bool, m mana.Color) (*game.Permanent, bool) {
	if m == mana.C {
		// can't pay for colorless pips via convoke.
		return nil, false
	}
	for _, permanent := range candidates {
		if used[permanent.ObjectID] {
			continue
		}
		for _, c := range s.PermanentEffectiveColors(permanent) {
			if cost.ManaForColor(c) == m {
				return permanent, true
			}
		}
	}
	return nil, false
}

// payGenericByTapping pays c's generic mana by tapping candidates, in order,
// for a nonmana mechanic alongside the permanents already tapped for it, then
// pays the rest of c with payRemaining. A permanent tapped for the mechanic
// can't also be tapped for mana, so tapping one may leave the rest unpayable.
// It tries tapping as many candidates as the generic mana allows, then fewer;
// for each number it tries the first candidates and then the first ones the
// remaining payment doesn't use. Tapping fewer only leaves more generic mana to
// pay, so it stops once the remaining payment fails with every candidate
// untapped. The mechanic must tap at least one permanent.
func payGenericByTapping(c *nonManaPaymentCost, tapped, candidates []*game.Permanent, payRemaining func(*nonManaPaymentCost, []*game.Permanent) (paymentPlan, bool)) ([]*game.Permanent, paymentPlan, bool) {
	tapped = slices.Clip(tapped)
	for count := min(len(candidates), c.genericDue()); count > 0 || (count == 0 && len(tapped) > 0); count-- {
		c.reduction = count
		taps := append(tapped, candidates[:count]...)
		if plan, ok := payRemaining(c, taps); ok {
			return taps, plan, true
		}
		if count == 0 {
			break
		}
		plan, ok := payRemaining(c, tapped)
		if !ok {
			break
		}
		needed := make(map[id.ID]bool, len(plan.manaTaps))
		for _, tap := range plan.manaTaps {
			needed[tap.permanent.ObjectID] = true
		}
		taps = tapped
		for _, permanent := range candidates {
			if len(taps) < len(tapped)+count && !needed[permanent.ObjectID] {
				taps = append(taps, permanent)
			}
		}
		if len(taps) == len(tapped)+count {
			return taps, plan, true
		}
	}
	return nil, paymentPlan{}, false
}

// delveCandidates returns up to limit cards in playerID's graveyard that delve
// may exile (CR 702.66a), excluding the spell itself and reserved cards.
func delveCandidates(s State, playerID game.PlayerID, limit int, sourceCardID id.ID, sourceZone zone.Type, reserved map[id.ID]bool) []id.ID {
	player, ok := s.Player(playerID)
	if !ok || limit <= 0 {
		return nil
	}
	var exiles []id.ID
	for _, cardID := range player.Graveyard.All() {
		if len(exiles) == limit {
			break
		}
		if (sourceZone == zone.Graveyard && cardID == sourceCardID) || reserved[cardID] {
			continue
		}
		exiles = append(exiles, cardID)
	}
	return exiles
}

// buildNonManaSpellPlan pays option's total cost with the nonmana payment
// mechanics the spell has, each tried after mana alone can't pay it. prefs
// holds the caster's announced Phyrexian payments for the whole total cost;
// each attempt pays its remaining cost with a fresh copy of those it keeps.
func buildNonManaSpellPlan(s State, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, option spellCostOption, xValue int, restrictions manaSourceRestrictions, prefs *Preferences) (paymentPlan, bool) {
	conversions := 1 << nonManaConversionCount(s, playerID, option.manaCost, xValue, prefs)
	payRemaining := func(c *nonManaPaymentCost, excluded []*game.Permanent) (paymentPlan, bool) {
		remainingRestrictions := restrictions
		if len(excluded) > 0 {
			remainingRestrictions = cloneManaSourceRestrictions(restrictions)
			for _, permanent := range excluded {
				remainingRestrictions.excluded[permanent.ObjectID] = true
			}
		}
		remainingCost, remainingPrefs := c.remaining()
		return buildPaymentPlanWithPreferences(s, playerID, remainingCost, xValue, remainingRestrictions, option.spendContext(), remainingPrefs)
	}
	if spellHasCostKeyword(s, playerID, option.card, cardID, sourceZone, game.Convoke) {
		for mask := range conversions {
			c, ok := newNonManaPaymentCost(s, playerID, option.manaCost, xValue, prefs)
			if !ok {
				return paymentPlan{}, false
			}
			colored, rest, ok := convokePayment(s, playerID, c, restrictions.excluded, mask)
			if !ok {
				continue
			}
			if taps, manaPlan, ok := payGenericByTapping(c, colored, rest, payRemaining); ok {
				manaPlan.convokeTaps = taps
				return manaPlan, true
			}
		}
	}
	if spellHasCostKeyword(s, playerID, option.card, cardID, sourceZone, game.Improvise) {
		for mask := range conversions {
			c, ok := newNonManaPaymentCost(s, playerID, option.manaCost, xValue, prefs)
			if !ok {
				return paymentPlan{}, false
			}
			// Improvise pays only generic mana (CR 702.126a).
			if !c.convert(mask) || c.genericDue() == 0 {
				continue
			}
			candidates := improviseCandidates(s, playerID, restrictions.excluded)
			if taps, manaPlan, ok := payGenericByTapping(c, nil, candidates, payRemaining); ok {
				manaPlan.improviseTaps = taps
				return manaPlan, true
			}
		}
	}
	if spellHasCostKeyword(s, playerID, option.card, cardID, sourceZone, game.Delve) {
		most, ok := newNonManaPaymentCost(s, playerID, option.manaCost, xValue, prefs)
		if !ok || !most.convert(conversions-1) {
			return paymentPlan{}, false
		}
		exiles := delveCandidates(s, playerID, most.genericDue(), cardID, sourceZone, restrictions.reservedGraveyardCardIDs)
		for exiled := 1; exiled <= len(exiles); exiled++ {
			for mask := range conversions {
				c, _ := newNonManaPaymentCost(s, playerID, option.manaCost, xValue, prefs)
				if !c.convert(mask) || c.genericDue() < exiled {
					continue
				}
				c.reduction = exiled
				if manaPlan, ok := payRemaining(c, nil); ok {
					manaPlan.delveExiles = append([]id.ID(nil), exiles[:exiled]...)
					return manaPlan, true
				}
			}
		}
	}
	return paymentPlan{}, false
}
