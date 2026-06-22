package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerMillThenOptionalAmongOneOfEachToBattlefield lowers the ordered sequence
// "you mill that many cards. You may put a <type-A> card and/or a <type-B> card
// from among them onto the battlefield." (Eivor, Wolf-Kissed's combat-damage
// trigger). The mill is mandatory: its dynamic count is the triggering combat
// damage dealt, and it publishes the milled cards. The optional "and/or" put
// then offers one independent optional pick per named type, each returning one
// of exactly those milled cards from the graveyard onto the battlefield,
// restricted to that type. "You may put a Saga card and/or a land card" thus
// lets the controller put up to one Saga and up to one land — two independent
// optionals, mirroring the inclusive "and/or".
//
// It keys entirely on the typed effect shape — a mandatory controller mill whose
// amount is the triggering combat damage, followed by an optional controller
// "them" put onto the battlefield whose selector is the inclusive one-of-each
// union of named card types — so it stays text-blind and fails closed on any
// other sequence. The triggering-combat-damage amount further requires the
// enclosing trigger to be a combat-damage event (enforced by
// lowerEventCombatDamageAmount), so a non-combat context fails closed.
func lowerMillThenOptionalAmongOneOfEachToBattlefield(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	mill := ctx.content.Effects[0]
	put := ctx.content.Effects[1]
	if mill.Kind != compiler.EffectMill ||
		mill.Context != parser.EffectContextController ||
		!mill.Exact ||
		mill.Negated ||
		mill.Optional ||
		mill.DelayedTiming != 0 ||
		mill.Amount.DynamicKind != compiler.DynamicAmountTriggeringCombatDamage ||
		len(mill.References) != 0 ||
		len(mill.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	if put.Kind != compiler.EffectPut ||
		put.Context != parser.EffectContextController ||
		!put.Optional ||
		put.Negated ||
		put.DelayedTiming != 0 ||
		put.ToZone != zone.Battlefield ||
		put.EntersTapped ||
		put.UnderYourControl ||
		put.Payment.Form != parser.EffectPaymentFormUnknown ||
		!put.Amount.Known ||
		put.Amount.RangeKnown ||
		put.Amount.VariableX ||
		put.Amount.Value != 1 ||
		!put.Selector.InclusiveOneOfEach ||
		len(put.Targets) != 0 ||
		len(put.References) != 1 ||
		put.References[0].Pronoun != compiler.ReferencePronounThem {
		return game.AbilityContent{}, false
	}
	selections, ok := oneOfEachCardSelections(put.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	millAmount, ok := lowerEventCombatDamageAmount(ctx, mill.Amount)
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence := []game.Instruction{
		{Primitive: game.Mill{
			Amount:        game.Dynamic(millAmount),
			Player:        game.ControllerReference(),
			PublishLinked: milledCardsLinkKey,
		}},
	}
	for _, selection := range selections {
		sequence = append(sequence, game.Instruction{
			Primitive: game.ReturnFromGraveyard{
				Player:      game.ControllerReference(),
				Amount:      game.Fixed(1),
				Destination: zone.Battlefield,
				FromLinked:  milledCardsLinkKey,
				Selection:   selection,
			},
			Optional: true,
		})
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}

// lowerMillThenPutAmongToBattlefield lowers the broadened ordered sequence "mill
// N cards. [You may] put a <type> card from among them onto the battlefield
// [tapped]." and "[you may] mill that many cards. Put any number of <type> cards
// from among them onto the battlefield [tapped].". The mill is mandatory or
// optional ("you may mill") and publishes the cards it milled; its amount is
// either a fixed printed count (Mole Module's "mill four cards") or the
// triggering combat damage ("mill that many cards"). The follow-up put returns
// either exactly one ("a <type> card") or any number ("any number of <type>
// cards") of those milled cards from the graveyard onto the battlefield,
// restricted to the put clause's single named card type and optionally entering
// tapped. A leading "you may" makes the put instruction itself optional.
//
// It keys entirely on the typed effect shape — a controller mill that publishes
// the milled cards followed by a controller "them" put onto the battlefield
// whose selector is a single named card type (never the inclusive one-of-each
// union handled by lowerMillThenOptionalAmongOneOfEachToBattlefield) — so it
// stays text-blind and fails closed on any other sequence. A dynamic mill amount
// further requires the enclosing trigger to be a combat-damage event (enforced
// by lowerEventCombatDamageAmount), so a non-combat context fails closed.
func lowerMillThenPutAmongToBattlefield(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	mill := ctx.content.Effects[0]
	put := ctx.content.Effects[1]
	if mill.Kind != compiler.EffectMill ||
		mill.Context != parser.EffectContextController ||
		!mill.Exact ||
		mill.Negated ||
		mill.DelayedTiming != 0 ||
		len(mill.References) != 0 ||
		len(mill.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	if put.Kind != compiler.EffectPut ||
		put.Context != parser.EffectContextController ||
		put.Negated ||
		put.DelayedTiming != 0 ||
		put.ToZone != zone.Battlefield ||
		put.UnderYourControl ||
		put.Selector.InclusiveOneOfEach ||
		put.Payment.Form != parser.EffectPaymentFormUnknown ||
		put.CounterKind != 0 ||
		len(put.Targets) != 0 ||
		len(put.References) != 1 ||
		put.References[0].Pronoun != compiler.ReferencePronounThem {
		return game.AbilityContent{}, false
	}
	anyNumber, ok := putAmongCountForm(put.Amount)
	if !ok {
		return game.AbilityContent{}, false
	}
	selection, ok := cardSelectionForSelector(put.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	millAmount, ok := millAmongAmount(ctx, mill.Amount)
	if !ok {
		return game.AbilityContent{}, false
	}
	returnPrimitive := game.ReturnFromGraveyard{
		Player:      game.ControllerReference(),
		Destination: zone.Battlefield,
		FromLinked:  milledCardsLinkKey,
		Selection:   selection,
		EntryTapped: put.EntersTapped,
		AnyNumber:   anyNumber,
	}
	if !anyNumber {
		returnPrimitive.Amount = game.Fixed(1)
	}
	sequence := []game.Instruction{
		{
			Primitive: game.Mill{
				Amount:        millAmount,
				Player:        game.ControllerReference(),
				PublishLinked: milledCardsLinkKey,
			},
			Optional: mill.Optional,
		},
		{
			Primitive: returnPrimitive,
			Optional:  put.Optional,
		},
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}

// putAmongCountForm classifies a put-from-among count as either a single card
// ("a <type> card", reported as anyNumber false) or the unbounded "any number of
// <type> cards" form (reported as anyNumber true). It fails closed on any ranged
// ("up to two"), variable-X, or dynamic count so an unmodeled quantity never
// silently collapses to one of the two supported forms.
func putAmongCountForm(amount compiler.CompiledAmount) (anyNumber bool, ok bool) {
	if amount.RangeKnown || amount.VariableX || amount.DynamicKind != compiler.DynamicAmountNone {
		return false, false
	}
	if amount.Known {
		return false, amount.Value == 1
	}
	// "Any number of <type> cards" carries no printed count: the amount is wholly
	// unspecified (not known, not ranged, not variable, not dynamic).
	return true, true
}

// millAmongAmount lowers the mill amount that feeds a put-from-among sequence. A
// printed count ("mill four cards") becomes a fixed quantity; the dynamic "mill
// that many cards" becomes the triggering combat damage, which requires a
// combat-damage trigger context and so fails closed elsewhere.
func millAmongAmount(ctx contentCtx, amount compiler.CompiledAmount) (game.Quantity, bool) {
	if amount.Known {
		return cardCountQuantity(amount, false)
	}
	dynamic, ok := lowerEventCombatDamageAmount(ctx, amount)
	if !ok {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}

// card and/or a land card") into one game.Selection per named type, so each can
// drive an independent optional put. It accepts only a plain card selection that
// is a pure union of named card types and subtypes; any other qualifier (color,
// supertype, mana value, controller, counter, keyword, zone, or exclusion) fails
// closed so a richer filter never silently splits into wrong picks. Subtype
// picks precede card-type picks deterministically, and at least two named types
// are required for the one-of-each wording to be meaningful.
func oneOfEachCardSelections(selector compiler.CompiledSelector) ([]game.Selection, bool) {
	if selector.Kind != compiler.SelectorCard ||
		selector.Controller != compiler.ControllerAny ||
		selector.All || selector.Another || selector.Other ||
		selector.Attacking || selector.Blocking ||
		selector.Tapped || selector.Untapped ||
		selector.NonToken || selector.TokenOnly ||
		selector.Colorless || selector.Multicolored || selector.BasicLandType ||
		selector.ConjunctiveTypes ||
		selector.MatchManaValue || selector.MatchTotalManaValue || selector.ManaValueX ||
		selector.MatchPower || selector.MatchToughness ||
		selector.MatchCounter || selector.MatchAnyCounter ||
		selector.PlayerOrPlaneswalker ||
		selector.SubtypeFromEntryChoice || selector.SubtypeFromChosenType ||
		selector.SubtypeFromChosenTypeExcluded ||
		selector.EnteredThisTurn ||
		selector.PowerLessThanSource || selector.PowerGreaterThanSource ||
		selector.Keyword != parser.KeywordUnknown ||
		selector.ExcludedKeyword != parser.KeywordUnknown ||
		selector.Zone != zone.None ||
		selector.RequiredName != "" ||
		len(selector.ExcludedTypes()) != 0 ||
		len(selector.Supertypes()) != 0 ||
		len(selector.ExcludedSupertypes()) != 0 ||
		len(selector.ColorsAny()) != 0 ||
		len(selector.ExcludedColors()) != 0 ||
		len(selector.ExcludedSubtypes()) != 0 ||
		len(selector.SourceTypes()) != 0 ||
		len(selector.Alternatives) != 0 {
		return nil, false
	}
	var selections []game.Selection
	for _, subtype := range selector.SubtypesAny() {
		selections = append(selections, game.Selection{SubtypesAny: []types.Sub{subtype}})
	}
	for _, cardType := range selector.RequiredTypesAny() {
		selections = append(selections, game.Selection{RequiredTypes: []types.Card{cardType}})
	}
	if len(selections) < 2 {
		return nil, false
	}
	return selections, true
}
