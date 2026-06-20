package parser

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
)

func exactGraveyardReturnEffectSyntax(effect *EffectSyntax) bool {
	text := exactEffectClauseText(effect)
	if len(effect.Targets) == 0 {
		switch {
		case strings.EqualFold(text, "Return this card from your graveyard to your hand."),
			strings.EqualFold(text, "Return this card from your graveyard to the battlefield."),
			strings.EqualFold(text, "Return this card from your graveyard to the battlefield tapped."):
			return true
		case strings.HasPrefix(strings.ToLower(text), "return this card from your graveyard to the battlefield with "),
			strings.HasPrefix(strings.ToLower(text), "return this card from your graveyard to the battlefield tapped with "):
			return effect.CounterKnown && effect.CounterKind == counter.PlusOnePlusOne
		default:
			return exactChosenGraveyardReturnEffectSyntax(effect, text)
		}
	}
	if len(effect.Targets) != 1 || !exactGraveyardCardTargetSyntax(&effect.Targets[0]) {
		return false
	}
	prefix := "Return " + effect.Targets[0].Text
	for _, suffix := range []string{
		" to your hand.",
		" to the battlefield.",
		" to the battlefield tapped.",
		" to the battlefield under your control.",
		" to the battlefield tapped under your control.",
		" on top of your library.",
		" on the top of your library.",
		" on bottom of your library.",
		" on the bottom of your library.",
	} {
		if strings.EqualFold(text, prefix+suffix) {
			return true
		}
	}
	return false
}

// exactChosenGraveyardReturnEffectSyntax recognizes the non-target "Return a
// <filter> card from your graveyard to your hand." recursion wording, where the
// returned card is chosen from the controller's own graveyard at resolution
// rather than targeted (Raise Dead targets; Takenuma's "return a creature or
// planeswalker card" does not). It reconstructs the canonical noun phrase from
// the effect's typed Selection the same way the targeted path does, accepting a
// single card-type, a union of card types, a permanent card, a single color, a
// colorless or multicolored card, a single subtype, or the plain "card" noun,
// with an optional "with mana value N or less" qualifier, and fails closed for
// every other selection shape so an unrepresentable filter keeps failing rather
// than lowering to a wrong predicate.
func exactChosenGraveyardReturnEffectSyntax(effect *EffectSyntax, text string) bool {
	if len(effect.References) != 0 || effect.ToZone != zone.Hand {
		return false
	}
	sel := effect.Selection
	if sel.Zone != zone.Graveyard || sel.Controller != SelectionControllerYou {
		return false
	}
	if sel.All || sel.Another || sel.Other || sel.Attacking || sel.Blocking ||
		sel.Tapped || sel.Untapped || sel.MatchPower || sel.MatchToughness ||
		sel.Keyword != KeywordUnknown || sel.ExcludedKeyword != KeywordUnknown ||
		len(sel.ExcludedTypes) != 0 || len(sel.SourceTypes) != 0 ||
		len(sel.Supertypes) != 0 || len(sel.ExcludedSupertypes) != 0 ||
		len(sel.ExcludedColors) != 0 || len(sel.Alternatives) != 0 {
		return false
	}
	noun, ok := graveyardCardNoun(sel)
	if !ok {
		return false
	}
	manaClause := ""
	if sel.MatchManaValue {
		clause, ok := graveyardManaValueClause(sel.ManaValue)
		if !ok {
			return false
		}
		manaClause = clause
	}
	article := indefiniteArticle(noun)
	return strings.EqualFold(text, "Return "+article+" "+noun+manaClause+" from your graveyard to your hand.")
}

func exactChosenCardsBattlefieldReturnEffectSyntax(effect *EffectSyntax) bool {
	return len(effect.Targets) == 0 &&
		len(effect.References) == 1 &&
		effect.References[0].Kind == ReferenceChosenCards &&
		strings.EqualFold(
			exactEffectClauseText(effect),
			"Return the chosen cards to the battlefield tapped.",
		)
}

// exactGraveyardExileEffectSyntax recognizes "Exile <target> from <owner>
// graveyard." for a graveyard-card target the executable backend lowers to a
// MoveCard from the graveyard to exile. It reuses the shared graveyard-card
// target reconstruction, so it accepts the same single card type, type union,
// permanent, subtype, color, and plain "card" nouns, owner suffixes, optional
// "with mana value N or less" qualifier, and "up to N" counts that the graveyard
// return and put paths accept, failing closed for every other wording.
func exactGraveyardExileEffectSyntax(effect *EffectSyntax) bool {
	if len(effect.Targets) != 1 || !exactGraveyardCardTargetSyntax(&effect.Targets[0]) {
		return false
	}
	return strings.EqualFold(exactEffectClauseText(effect), "Exile "+effect.Targets[0].Text+".")
}

// parseGraveyardZoneExile recognizes the whole-graveyard exile "Exile target
// player's graveyard." (and its "target opponent's" variant), returning whose
// graveyard the effect wipes. Unlike single-card graveyard exile, the graveyard
// noun is the exiled object rather than a card-target zone, so the target is a
// player and the effect carries no FromZone card target. It fails closed for
// every other wording — single-card exile, "that player's graveyard", and the
// unmodeled "all graveyards" / multi-graveyard forms — by anchoring on the exact
// one-player-target shape; exactPlayerGraveyardExileEffectSyntax then verifies
// the reconstruction owns the wording byte-for-byte.
func exactSourceSpellExileSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectExile ||
		effect.Negated ||
		effect.Duration != EffectDurationNone ||
		effect.FromZone != zone.None ||
		effect.ToZone != zone.None ||
		len(effect.Targets) != 0 ||
		len(effect.References) != 1 {
		return false
	}
	reference := effect.References[0]
	if reference.Kind != ReferenceThisObject && reference.Kind != ReferenceSelfName {
		return false
	}
	if reference.Kind == ReferenceThisObject && reference.Text != "this spell" {
		return false
	}
	return effect.Text == "Exile "+reference.Text+"."
}

// exactCounteredSpellExileSyntax recognizes the exact counter rider "If that
// spell is countered this way, exile it instead of putting it into its owner's
// graveyard." and marks it so a preceding counter effect lowers to a single
// counter-and-exile primitive. The parser owns this wording; any other exile
// rider leaves the clause non-exact so lowering fails closed.
func exactCounteredSpellExileSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectExile {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(effect.Text),
		"If that spell is countered this way, exile it instead of putting it into its owner's graveyard.") {
		effect.CounteredSpellExileReplacement = true
		return true
	}
	return false
}

func parseGraveyardZoneExile(effect *EffectSyntax) GraveyardZoneExileKind {
	if effect.Kind != EffectExile || effect.Negated {
		return GraveyardZoneExileNone
	}
	if effect.FromZone != zone.None {
		return GraveyardZoneExileNone
	}
	if len(effect.Targets) == 0 {
		switch {
		case strings.EqualFold(strings.TrimSpace(effect.Text), "Exile all graveyards."),
			strings.EqualFold(strings.TrimSpace(effect.Text), "Exile each player's graveyard."):
			return GraveyardZoneExileAll
		}
		return GraveyardZoneExileNone
	}
	if len(effect.Targets) != 1 {
		return GraveyardZoneExileNone
	}
	target := effect.Targets[0]
	if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
		return GraveyardZoneExileNone
	}
	switch {
	case strings.EqualFold(target.Text, "target player's graveyard") &&
		target.Selection.Controller == SelectionControllerAny:
		return GraveyardZoneExileTargetPlayer
	case strings.EqualFold(target.Text, "target opponent's graveyard") &&
		target.Selection.Controller == SelectionControllerOpponent:
		return GraveyardZoneExileTargetOpponent
	}
	return GraveyardZoneExileNone
}

// exactPlayerGraveyardExileEffectSyntax reports whether a recognized
// whole-graveyard exile reconstructs its clause text byte-for-byte, so the
// typed GraveyardZoneExile owner relation owns the wording. Any other exile
// wording fails closed here.
func exactPlayerGraveyardExileEffectSyntax(effect *EffectSyntax) bool {
	var canonical string
	switch effect.GraveyardZoneExile {
	case GraveyardZoneExileTargetPlayer:
		canonical = "Exile target player's graveyard."
	case GraveyardZoneExileTargetOpponent:
		canonical = "Exile target opponent's graveyard."
	case GraveyardZoneExileAll:
		text := exactEffectClauseText(effect)
		return strings.EqualFold(text, "Exile all graveyards.") ||
			strings.EqualFold(text, "Exile each player's graveyard.")
	default:
		return false
	}
	return strings.EqualFold(exactEffectClauseText(effect), canonical)
}

func exactGraveyardPutEffectSyntax(effect *EffectSyntax) bool {
	if len(effect.Targets) != 1 || !exactGraveyardCardTargetSyntax(&effect.Targets[0]) {
		return false
	}
	text := exactEffectClauseText(effect)
	prefix := "Put " + effect.Targets[0].Text
	for _, suffix := range []string{
		" onto the battlefield.",
		" onto the battlefield under your control.",
		" on top of your library.",
		" on the top of your library.",
		" on bottom of your library.",
		" on the bottom of your library.",
	} {
		if strings.EqualFold(text, prefix+suffix) {
			return true
		}
	}
	return false
}

// exactGraveyardCardTargetSyntax reports whether target is a graveyard-card
// target the executable backend can lower exactly. It reconstructs the canonical
// Oracle noun phrase ("Return target <noun> from <owner> graveyard") byte-for-
// byte from the selection's typed fields and compares it to target.Text,
// accepting a single card-type, a union of card types ("creature or enchantment
// card"), a permanent card, a single color, a colorless or multicolored card, a
// single subtype, or the plain "card" noun, with an optional "with mana value N
// or less" qualifier and an optional multi-target or "up to N" count. It fails
// closed for every qualifier the canonical phrasing cannot render (power,
// toughness, keyword, supertype, excluded types or colors, combinations), so an
// unrepresentable target keeps failing rather than lowering to a wrong predicate.
func exactGraveyardCardTargetSyntax(target *TargetSyntax) bool {
	sel := target.Selection
	if sel.Zone != zone.Graveyard || sel.Other {
		return false
	}
	if target.Cardinality == (TargetCardinalitySyntax{Min: 0, Max: 2}) &&
		strings.EqualFold(target.Text, "up to two target cards with cycling from your graveyard") {
		return true
	}
	if sel.All || sel.Attacking || sel.Blocking || sel.Tapped || sel.Untapped ||
		sel.Keyword != KeywordUnknown || sel.MatchPower || sel.MatchToughness ||
		len(sel.ExcludedTypes) != 0 || len(sel.SourceTypes) != 0 ||
		len(sel.Supertypes) != 0 || len(sel.ExcludedColors) != 0 {
		return false
	}
	owner, ok := graveyardOwnerSuffix(sel.Controller)
	if !ok {
		return false
	}
	prefix, plural, ok := graveyardCardCardinalityPrefix(target.Cardinality, sel.Another)
	if !ok {
		return false
	}
	noun, ok := graveyardCardNoun(sel)
	if !ok {
		return false
	}
	if plural {
		noun += "s"
	}
	manaClause := ""
	if sel.MatchManaValue {
		// The canonical mana-value qualifier follows the singular noun; the
		// multi-target plural noun never carries it in printed Oracle wording.
		if plural {
			return false
		}
		clause, ok := graveyardManaValueClause(sel.ManaValue)
		if !ok {
			return false
		}
		manaClause = clause
	}
	return strings.EqualFold(target.Text, prefix+noun+manaClause+" "+owner)
}

// graveyardOwnerSuffix renders the canonical "from <owner> graveyard" clause for
// a graveyard-card target's controller relation: "your" for the controller,
// "a" for any graveyard, and "an opponent's" for an opponent. It fails closed for
// the "you don't control" relation, which has no graveyard-owner phrasing.
func graveyardOwnerSuffix(controller SelectionController) (string, bool) {
	switch controller {
	case SelectionControllerYou:
		return "from your graveyard", true
	case SelectionControllerAny:
		return "from a graveyard", true
	case SelectionControllerOpponent:
		return "from an opponent's graveyard", true
	default:
		return "", false
	}
}

// graveyardCardCardinalityPrefix returns the canonical count words preceding the
// graveyard-card noun, whether that noun is plural, and whether the cardinality
// is one the round-trip represents. Single targets render "target " (or
// "another target " for a self-exclusion); optional and multi-target counts reuse
// multiTargetCardinalityPrefix for "up to one ", "up to <N> ", and "<N> ". It
// fails closed for a self-exclusion combined with a multi-target count, which has
// no canonical phrasing.
func graveyardCardCardinalityPrefix(c TargetCardinalitySyntax, another bool) (prefix string, plural, ok bool) {
	if c == (TargetCardinalitySyntax{Min: 1, Max: 1}) {
		if another {
			return "another target ", false, true
		}
		return "target ", false, true
	}
	if another {
		return "", false, false
	}
	countPrefix, plural, ok := multiTargetCardinalityPrefix(c)
	if !ok {
		return "", false, false
	}
	return countPrefix + "target ", plural, true
}

// graveyardCardNoun reconstructs the singular graveyard-card noun ("creature
// card", "creature or enchantment card", "permanent card", "green card",
// "multicolored card", "colorless card", "Zombie card", or the plain "card")
// from selection's typed fields. It accepts exactly one restriction category so
// combinations it could not render in canonical order fail closed.
func graveyardCardNoun(sel SelectionSyntax) (string, bool) {
	hasTypes := len(sel.RequiredTypesAny) > 0
	hasColors := len(sel.ColorsAny) > 0
	hasColorQualifier := sel.Colorless || sel.Multicolored
	hasSubtype := len(sel.SubtypesAny) > 0
	isPermanent := sel.Kind == SelectionPermanent

	categories := 0
	for _, present := range []bool{hasTypes, hasColors, hasColorQualifier, hasSubtype, isPermanent} {
		if present {
			categories++
		}
	}
	if categories > 1 {
		return "", false
	}
	switch {
	case hasTypes:
		return graveyardCardTypeNoun(sel)
	case isPermanent:
		return "permanent card", true
	case hasColors:
		if len(sel.ColorsAny) != 1 {
			return "", false
		}
		word, ok := colorWord(sel.ColorsAny[0])
		if !ok {
			return "", false
		}
		return word + " card", true
	case hasColorQualifier:
		if sel.Colorless && sel.Multicolored {
			return "", false
		}
		if sel.Colorless {
			return "colorless card", true
		}
		return "multicolored card", true
	case hasSubtype:
		if len(sel.SubtypesAny) != 1 {
			return "", false
		}
		return string(sel.SubtypesAny[0]) + " card", true
	default:
		if sel.Kind != SelectionCard {
			return "", false
		}
		return "card", true
	}
}

// graveyardCardTypeNoun reconstructs the card-type noun ("creature card",
// "creature or enchantment card"). A single type must be carried by the selection
// Kind so lowering's Kind-to-type mapping reproduces it (this excludes the
// instant and sorcery types, whose single-type form the compiler does not retain
// and which would otherwise lower to an unrestricted card). A union of two or
// more types is carried explicitly by the compiler, so each member is rendered
// from its card-type word and joined with " or ".
func graveyardCardTypeNoun(sel SelectionSyntax) (string, bool) {
	if len(sel.RequiredTypesAny) == 1 {
		noun, ok := permanentSelectionNoun(sel.Kind)
		if !ok {
			return "", false
		}
		word, ok := cardTypeWord(sel.RequiredTypesAny[0])
		if !ok || word != noun {
			return "", false
		}
		return noun + " card", true
	}
	words := make([]string, 0, len(sel.RequiredTypesAny))
	for _, cardType := range sel.RequiredTypesAny {
		word, ok := cardTypeWord(cardType)
		if !ok {
			return "", false
		}
		words = append(words, word)
	}
	return strings.Join(words, " or ") + " card", true
}

// graveyardManaValueClause renders the canonical " with mana value N or less"
// qualifier from a mana-value comparison. It accepts only the "or less" bound the
// printed Oracle wording uses, failing closed for any other comparison operator.
func graveyardManaValueClause(manaValue compare.Int) (string, bool) {
	if manaValue.Op != compare.LessOrEqual {
		return "", false
	}
	return " with mana value " + strconv.Itoa(manaValue.Value) + " or less", true
}

func titleFirstEffectText(text string) string {
	if text == "" {
		return ""
	}
	return strings.ToUpper(text[:1]) + text[1:]
}

func signedEffectAmountText(amount SignedAmountSyntax) string {
	if amount.Negative {
		return fmt.Sprintf("-%d", amount.Value)
	}
	return fmt.Sprintf("+%d", amount.Value)
}

// signedPTSideText renders one power/toughness delta side for exact
// reconstruction: a variable "X" side renders as "+X"/"-X", a fixed side as its
// signed integer ("+2"/"-1").
func signedPTSideText(amount SignedAmountSyntax) string {
	if amount.VariableX {
		if amount.Negative {
			return "-X"
		}
		return "+X"
	}
	return signedEffectAmountText(amount)
}

func exactCounterEffectSyntax(effect *EffectSyntax) bool {
	if exactDirectTargetEffectSyntax(effect, "Counter") {
		return true
	}
	return len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		effect.Payment.Payer == EffectPaymentPayerTargetController &&
		len(effect.Payment.ManaCost) > 0 &&
		strings.EqualFold(
			exactEffectClauseText(effect),
			"Counter "+effect.Targets[0].Text+" unless its controller pays "+effect.Payment.ManaCost.String()+".",
		)
}

func exactDirectTargetEffectSyntax(effect *EffectSyntax, verb string) bool {
	return len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		strings.EqualFold(exactEffectClauseText(effect), verb+" "+effect.Targets[0].Text+".")
}

// exactChooseNewTargetsEffectSyntax recognizes the retarget effect "[You may]
// choose new targets for <target spell or ability>." The optional "You may"
// wrapper is carried by effect.Optional (exactEffectClauseText drops it), and
// the single stack-object target is the spell or ability whose targets are
// re-chosen. Any trailing rider ("choose new targets for the copy", "Then copy
// that spell") leaves a non-stack target or extra clause text and fails closed.
func exactChooseNewTargetsEffectSyntax(effect *EffectSyntax) bool {
	return len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		len(effect.References) == 0 &&
		strings.EqualFold(exactEffectClauseText(effect), "Choose new targets for "+effect.Targets[0].Text+".")
}

func exactNegatedNextUntapStepSyntax(effect *EffectSyntax) bool {
	if !effect.Negated || effect.Context != EffectContextUnknown ||
		len(effect.Targets) != 0 || len(effect.References) != 0 {
		return false
	}
	words := normalizedWords(effect.Tokens)
	verb := slices.Index(words, "untap")
	return verb == 4 &&
		slices.Equal(words[:verb], []string{"lands", "you", "control", "don't"}) &&
		slices.Equal(words[verb+1:], []string{"during", "your", "next", "untap", "step"})
}

// exactPriorSubjectNextUntapStepSyntax recognizes the prior-subject "doesn't
// untap during its controller's next untap step" clause that follows a tap
// effect (e.g. "Tap target creature. It doesn't untap during its controller's
// next untap step."). The subject is the just-tapped permanent, referenced by
// the pronoun "It"/"They" or a "That <permanent>"/"Those <permanents>" demonstrative,
// and the possessive controller pronoun ("its"/"their") agrees in number with
// the subject. Only the single "next untap step" duration is exact; every other
// wording — "next two untap steps", "for as long as you control ...", or "during
// your next untap step" on a prior subject — leaves the clause non-exact so
// lowering fails closed.
func exactPriorSubjectNextUntapStepSyntax(effect *EffectSyntax) bool {
	if !effect.Negated || effect.Optional ||
		len(effect.Targets) != 0 || len(effect.References) != 2 ||
		effect.Duration != EffectDurationNone || effect.DelayedTiming != DelayedTimingNone {
		return false
	}
	subject := effect.References[0]
	possessive := effect.References[1]
	plural := false
	switch subject.Kind {
	case ReferenceThatObject:
	case ReferencePronoun:
		switch subject.Pronoun {
		case PronounIt:
		case PronounThose, PronounThey:
			plural = true
		default:
			return false
		}
	default:
		return false
	}
	if possessive.Kind != ReferencePronoun {
		return false
	}
	words := normalizedWords(effect.Tokens)
	verb := slices.Index(words, "untap")
	if verb < 1 {
		return false
	}
	var negation string
	var tail []string
	if plural {
		negation, tail = "don't", []string{"during", "their", "controller's", "next", "untap", "step"}
		if possessive.Pronoun != PronounTheir {
			return false
		}
	} else {
		negation, tail = "doesn't", []string{"during", "its", "controller's", "next", "untap", "step"}
		if possessive.Pronoun != PronounIts {
			return false
		}
	}
	return words[verb-1] == negation && slices.Equal(words[verb+1:], tail)
}

// exactControlledBounceEffectSyntax recognizes the controlled-choice battlefield
// bounce "Return a/an/another <permanent> you control to its owner's hand." that
// lowers to a Bounce whose resolving controller chooses one permanent they
// control. It carries no target (the choice is made at resolution, not by
// targeting) and the singular "its owner's hand" destination; every other
// wording fails closed so the targeted and mass bounce paths are untouched.
func exactControlledBounceEffectSyntax(effect *EffectSyntax) bool {
	if effect.ToZone != zone.Hand || len(effect.Targets) != 0 ||
		effect.Context != EffectContextController {
		return false
	}
	phrase, ok := exactControlledBounceSelectionText(effect.Selection)
	if !ok {
		return false
	}
	return strings.EqualFold(exactEffectClauseText(effect), "Return "+phrase+" to its owner's hand.")
}

func exactBounceEffectSyntax(effect *EffectSyntax) bool {
	return len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		strings.EqualFold(exactEffectClauseText(effect), "Return "+effect.Targets[0].Text+" to its owner's hand.")
}

// exactDualBounceEffectSyntax recognizes the dual-target battlefield bounce
// "Return target <A> and target <B> to their owners' hands." (e.g. Aether
// Tradewinds, Peel from Reality, Churning Eddy) that the executable backend
// lowers to two single-target specs, one Bounce per target. It accepts only two
// exact single (cardinality-one) permanent targets joined by " and " and the
// exact plural possessive destination, failing closed for every other wording so
// the single-target and multi-slot bounce paths are untouched.
func exactDualBounceEffectSyntax(effect *EffectSyntax) bool {
	if len(effect.Targets) != 2 {
		return false
	}
	for i := range effect.Targets {
		target := &effect.Targets[i]
		if !target.Exact ||
			target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
			return false
		}
	}
	reconstruction := "Return " + effect.Targets[0].Text + " and " + effect.Targets[1].Text + " to their owners' hands."
	return strings.EqualFold(exactEffectClauseText(effect), reconstruction)
}

// exactMultiBounceEffectSyntax recognizes the plural battlefield bounce
// "Return <N target permanents> to their owners' hands." (and the optional "up
// to N" form) that the executable backend lowers to one multi-target spec with
// one Bounce per slot. It accepts only the exact plural possessive destination
// for a multi-target permanent, failing closed for every other wording so the
// single-target "to its owner's hand" path is untouched.
func exactMultiBounceEffectSyntax(effect *EffectSyntax) bool {
	return len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		effect.Targets[0].Cardinality.Max >= 2 &&
		strings.EqualFold(exactEffectClauseText(effect), "Return "+effect.Targets[0].Text+" to their owners' hands.")
}

// exactSelfBounceEffectSyntax recognizes "Return <subject> to its owner's hand."
// where the subject is the source permanent itself, named either as "this
// <object>" (ReferenceThisObject, e.g. Etherium-Horn Sorcerer's "Return this
// creature to its owner's hand.") or by the card's own name (ReferenceSelfName,
// e.g. Selenia, Dark Angel's "Return Selenia to its owner's hand."). The subject
// is reconstructed byte-exactly from the recognized self-reference's tokens, so
// any other wording fails the round-trip and stays unsupported.
func exactSelfBounceEffectSyntax(effect *EffectSyntax) bool {
	if len(effect.Targets) != 0 || len(effect.References) == 0 {
		return false
	}
	switch effect.References[0].Kind {
	case ReferenceThisObject, ReferenceSelfName:
	default:
		return false
	}
	subject := joinedEffectText(effect.References[0].Tokens)
	return strings.EqualFold(exactEffectClauseText(effect), "Return "+subject+" to its owner's hand.")
}

func exactDirectPronounEffectSyntax(effect *EffectSyntax, exact string) bool {
	return len(effect.Targets) == 0 &&
		effect.Duration == EffectDurationNone &&
		strings.EqualFold(exactEffectClauseText(effect), exact)
}

func exactDirectReferenceEffectSyntax(effect *EffectSyntax, verb string) bool {
	if len(effect.Targets) != 0 || effect.Optional || effect.Duration != EffectDurationNone {
		return false
	}
	object, ok := exactObjectReferenceText(effect.References)
	return ok && strings.EqualFold(exactEffectClauseText(effect), verb+" "+object+".")
}

func exactObjectReferenceText(references []Reference) (string, bool) {
	if len(references) != 1 {
		return "", false
	}
	switch references[0].Kind {
	case ReferenceThatObject, ReferenceThisObject, ReferenceSelfName:
	case ReferencePronoun:
		if references[0].Pronoun != PronounIt {
			return "", false
		}
	default:
		return "", false
	}
	return joinedEffectText(references[0].Tokens), true
}

// exactSelfSubjectReferenceText returns the rendered text of a single source
// self-reference, either "this <object>" (ReferenceThisObject) or the card's own
// name (ReferenceSelfName), used to recognize effects whose subject is the source
// permanent itself (e.g. "This creature gains flying until end of turn.").
func exactSelfSubjectReferenceText(references []Reference) (string, bool) {
	if len(references) != 1 {
		return "", false
	}
	switch references[0].Kind {
	case ReferenceThisObject, ReferenceSelfName:
		return joinedEffectText(references[0].Tokens), true
	}
	return "", false
}

// exactThoseSubjectReference reports whether the effect's subject is the single
// demonstrative back-reference "those" ("Those creatures gain …"), which names a
// group introduced by a preceding clause. The referenced group is reconstructed
// downstream from that clause, so only the demonstrative itself is matched here.
func exactThoseSubjectReference(references []Reference) bool {
	return len(references) == 1 &&
		references[0].Kind == ReferencePronoun &&
		references[0].Pronoun == PronounThose
}

// modifyPTSubjectReferences returns the effect's references with the dynamic
// power referent removed when the amount is "where X is its power" (or "this
// creature's power"/"<name>'s power"). That referent is a second reference that
// names the power source, not the pumped subject, so a self-pump such as "This
// creature gets +X/+X until end of turn, where X is its power." carries two
// references (the source subject and the "its" power referent). Dropping the
// referent whose span matches the amount's referent span lets the subject
// reconstruction see the single subject reference and recognize the clause.
func modifyPTSubjectReferences(effect *EffectSyntax) []Reference {
	if effect.Amount.DynamicKind != EffectDynamicAmountSourcePower {
		return effect.References
	}
	subjects := make([]Reference, 0, len(effect.References))
	for _, reference := range effect.References {
		if reference.Span == effect.Amount.ReferenceSpan {
			continue
		}
		subjects = append(subjects, reference)
	}
	return subjects
}

func exactFightEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context == EffectContextPriorSubject &&
		len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		strings.EqualFold(exactEffectClauseText(effect), "fights "+effect.Targets[0].Text+".") {
		return true
	}
	return len(effect.Targets) == 2 &&
		effect.Targets[0].Exact &&
		effect.Targets[1].Exact &&
		strings.EqualFold(exactEffectClauseText(effect), effect.Targets[0].Text+" fights "+effect.Targets[1].Text+".")
}

func exactMassEffectSyntax(effect *EffectSyntax, prefix string) bool {
	text := exactEffectClauseText(effect)
	if !strings.HasPrefix(strings.ToLower(text), strings.ToLower(prefix)) || !strings.HasSuffix(text, ".") {
		return false
	}
	phrase := text[len(prefix) : len(text)-1]
	return exactMassGroupPhrase(phrase) || exactMassSubtypePhrase(&effect.Selection, phrase)
}

// exactMassSubtypePhrase reconstructs the canonical mass phrase for a subtype
// group ("all Islands", "all Goblins", "all Dragon creatures") from the parsed
// selection and compares it byte-exactly to the source phrase. A bare subtype
// noun ("Islands") selects any permanent carrying that subtype; a subtype before
// a permanent card-type noun ("Dragon creatures") also restricts to that card
// type. It accepts exactly one subtype with an optional single permanent
// card-type noun and no controller clause, failing closed for every other
// qualifier so unsupported mass wordings keep failing the round-trip.
func exactMassSubtypePhrase(selection *SelectionSyntax, phrase string) bool {
	if len(selection.SubtypesAny) != 1 ||
		selection.Controller != SelectionControllerAny ||
		selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking ||
		selection.Tapped || selection.Untapped ||
		selection.Keyword != KeywordUnknown || selection.ExcludedKeyword != KeywordUnknown ||
		selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.Colorless || selection.Multicolored ||
		!selectionRedundantRequiredNoun(*selection) || len(selection.ExcludedTypes) != 0 ||
		len(selection.Supertypes) != 0 ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedColors) != 0 {
		return false
	}
	subtype := strings.ToLower(string(selection.SubtypesAny[0]))
	if noun, ok := permanentSelectionNoun(selection.Kind); ok {
		// A subtype qualifying an explicit card type stays singular while the
		// card-type noun pluralizes ("dragon creatures").
		return strings.EqualFold(phrase, subtype+" "+noun+"s")
	}
	if selection.Kind != SelectionUnknown {
		return false
	}
	// A bare subtype noun pluralizes on its own. "Plains" is already plural; the
	// other recorded subtypes add a trailing "s". Mismatched pluralizations fall
	// through and fail closed without producing a false positive.
	return strings.EqualFold(phrase, subtype) || strings.EqualFold(phrase, subtype+"s")
}

// exactMassBounceEffectSyntax recognizes the mass battlefield return
// "Return all <group> to their owners' hands." (and the "you control" variant
// "Return all <group> you control to their owner's hand.") that lowers to a
// single group Bounce, mirroring the mass destroy/exile group syntax. The return
// wording differs from destroy/exile only by its "to their owners' hands"
// destination suffix; that possessive is reconstructed canonically here so the
// group phrase between "Return all " and the suffix can be validated by the
// shared exactMassGroupPhrase. It fails closed for every other return wording so
// the single- and multi-target bounce paths are untouched.
func exactMassBounceEffectSyntax(effect *EffectSyntax) bool {
	if effect.ToZone != zone.Hand {
		return false
	}
	const prefix = "Return all "
	text := exactEffectClauseText(effect)
	if !strings.HasPrefix(strings.ToLower(text), strings.ToLower(prefix)) {
		return false
	}
	for _, suffix := range []string{" to their owners' hands.", " to their owner's hand."} {
		if remainder, ok := strings.CutSuffix(text, suffix); ok {
			return exactMassGroupPhrase(remainder[len(prefix):])
		}
	}
	return false
}

func exactMassGroupPhrase(phrase string) bool {
	if phrase == "" || strings.TrimSpace(phrase) != phrase {
		return false
	}
	phrase = strings.ToLower(phrase)
	hadControllerSuffix := false
	for _, suffix := range []string{" you don't control", " your opponents control", " you control"} {
		if remainder, ok := strings.CutSuffix(phrase, suffix); ok {
			phrase = remainder
			hadControllerSuffix = true
			break
		}
	}
	if exactMassNumericPhrase(phrase) {
		return true
	}
	if !hadControllerSuffix {
		if keyword, ok := strings.CutPrefix(phrase, "creatures with "); ok {
			return keyword != "" &&
				!strings.Contains(keyword, " ") &&
				exactTemporaryKeywordList(keyword)
		}
	}
	if exactMassBaseNoun(phrase) {
		return true
	}
	for _, prefix := range []string{
		"other ", "tapped ", "untapped ", "nonland ", "nonartifact ", "noncreature ", "nonenchantment ",
		"white ", "blue ", "black ", "red ", "green ", "nonwhite ", "nonblue ", "nonblack ", "nonred ", "nongreen ",
		"attacking ", "blocking ", "attacking or blocking ",
	} {
		if remainder, ok := strings.CutPrefix(phrase, prefix); ok {
			return exactMassBaseNoun(remainder)
		}
	}
	// "nonbasic" is a supertype exclusion meaningful only for lands ("Destroy all
	// nonbasic lands."); every other base noun fails closed.
	if remainder, ok := strings.CutPrefix(phrase, "nonbasic "); ok {
		return remainder == "lands"
	}
	return false
}

func exactMassBaseNoun(phrase string) bool {
	switch phrase {
	case "creatures", "artifacts", "enchantments", "lands", "planeswalkers", "permanents",
		"creatures and lands", "creatures and planeswalkers", "artifacts and enchantments",
		"artifacts and creatures", "artifacts, creatures, and enchantments",
		"artifacts, creatures, and lands":
		return true
	default:
		return false
	}
}

// exactMassNumericPhrase recognizes a mass group restricted by a numeric
// "with mana value"/"with power"/"with toughness" comparison, optionally behind
// a single excluded-type prefix ("nonland permanents with mana value 1 or
// less"). Power and toughness exist only on creatures, so they are accepted
// solely on the plain "creatures" noun; mana value applies to every permanent
// and so is accepted on any base noun. It fails closed for comparison shapes
// without a canonical Oracle phrasing the round-trip can reproduce.
func exactMassNumericPhrase(phrase string) bool {
	for _, exPrefix := range []string{"", "nonland ", "nonartifact ", "noncreature ", "nonenchantment "} {
		rest, ok := strings.CutPrefix(phrase, exPrefix)
		if !ok {
			continue
		}
		for _, noun := range []string{"creatures", "artifacts", "enchantments", "lands", "planeswalkers", "permanents"} {
			comparison, ok := strings.CutPrefix(rest, noun+" with ")
			if !ok {
				continue
			}
			qualifiers := []string{"mana value"}
			if exPrefix == "" && noun == "creatures" {
				qualifiers = []string{"mana value", "power", "toughness"}
			}
			if exactMassComparisonClause(comparison, qualifiers) {
				return true
			}
		}
	}
	return false
}

// exactMassComparisonClause reports whether comparison is a canonical
// "<qualifier> N", "<qualifier> N or less", "<qualifier> N or greater", or
// "<qualifier> equal to N" clause for one of the allowed qualifiers.
func exactMassComparisonClause(comparison string, qualifiers []string) bool {
	for _, qualifier := range qualifiers {
		rest, ok := strings.CutPrefix(comparison, qualifier+" ")
		if !ok {
			continue
		}
		parts := strings.Fields(rest)
		switch {
		case len(parts) == 1:
			_, err := strconv.Atoi(parts[0])
			return err == nil
		case len(parts) == 3 && parts[0] == "equal" && parts[1] == "to":
			_, err := strconv.Atoi(parts[2])
			return err == nil
		case len(parts) == 3 && parts[1] == "or" && (parts[2] == "less" || parts[2] == "greater"):
			_, err := strconv.Atoi(parts[0])
			return err == nil
		}
	}
	return false
}

func exactEffectClauseText(effect *EffectSyntax) string {
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb < 0 {
		return ""
	}
	start := effectSubjectStart(effect.Tokens, verb)
	if effect.Optional && effectWordsAt(effect.Tokens, start, "you", "may") && start+2 == verb {
		start = verb
	}
	text := joinedEffectText(effect.Tokens[start:])
	if len(effect.Tokens) > 0 && effect.Tokens[len(effect.Tokens)-1].Kind != shared.Period {
		text += "."
	}
	if effect.DelayedTiming != DelayedTimingNone {
		for _, suffix := range []string{
			" at the beginning of the next end step.",
			" at the beginning of the next turn's upkeep.",
		} {
			if prefix, ok := strings.CutSuffix(text, suffix); ok {
				return prefix + "."
			}
		}
	}
	return text
}

func exactDamageEffectSyntax(effect *EffectSyntax) bool {
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb < 0 {
		return false
	}
	subjectStart := effectSubjectStart(effect.Tokens, verb)
	subjectTokens := effect.Tokens[subjectStart:verb]
	subject := ""
	if len(subjectTokens) == 0 {
		if effect.Context != EffectContextPriorSubject {
			return false
		}
	} else {
		subject = joinedEffectText(subjectTokens)
		subjectSpan := shared.SpanOf(subjectTokens)
		exactSubject := false
		for _, reference := range effect.SubjectReferences {
			if !spanCovers(subjectSpan, reference.Span) {
				continue
			}
			exactSubject = reference.Kind == ReferenceSelfName ||
				reference.Kind == ReferenceThisObject ||
				reference.Kind == ReferencePronoun && reference.Pronoun == PronounIt
		}
		if !exactSubject {
			return false
		}
	}
	verbText := effect.Tokens[verb].Text
	if !equalWord(effect.Tokens[verb], "deal") && !equalWord(effect.Tokens[verb], "deals") {
		return false
	}
	prefix := verbText
	if subject != "" {
		prefix = subject + " " + verbText
	}
	text := joinedEffectText(effect.Tokens[subjectStart:])
	if len(effect.Tokens) > 0 && effect.Tokens[len(effect.Tokens)-1].Kind != shared.Period {
		text += "."
	}
	if effect.Divided {
		return exactDividedDamageText(effect, prefix, text)
	}
	if effect.DamageRecipientReference != DamageRecipientReferenceNone {
		if len(effect.Targets) != 0 ||
			effect.Amount.DynamicForm != EffectDynamicAmountFormNone {
			return false
		}
		amount := "X"
		if effect.Amount.Known {
			amount = strconv.Itoa(effect.Amount.Value)
		} else if !effect.Amount.VariableX {
			return false
		}
		recipient, ok := damageRecipientTokens(effect.Tokens)
		if !ok {
			return false
		}
		return text == fmt.Sprintf("%s %s damage to %s.", prefix, amount, joinedEffectText(recipient))
	}
	if len(effect.Targets) == 0 {
		// A "where X is the number of ..." or "equal to ..." dynamic amount on
		// a single-recipient group clause is reconstructed from the captured
		// amount phrase, mirroring the single-target dynamic forms.
		if effect.Amount.DynamicForm == EffectDynamicAmountFormWhereX ||
			effect.Amount.DynamicForm == EffectDynamicAmountFormEqual {
			return exactGroupDynamicDamageText(effect, prefix, text)
		}
		amount, ok := exactGroupDamageAmountText(effect.Amount)
		if !ok {
			return false
		}
		if len(effect.DamageRecipientPair) == 2 {
			first, ok := exactGroupDamageRecipientText(effect.DamageRecipientPair[0])
			if !ok {
				return false
			}
			second, ok := exactGroupDamageRecipientText(effect.DamageRecipientPair[1])
			if !ok {
				return false
			}
			return text == fmt.Sprintf("%s %s damage to %s and %s.", prefix, amount, first, second)
		}
		recipient, ok := exactGroupDamageRecipientText(effect.Selection)
		if !ok {
			return false
		}
		return text == fmt.Sprintf("%s %s damage to %s.", prefix, amount, recipient)
	}
	// "<prefix> A damage to <target0> and B damage to <target1>." deals to two
	// independently chosen single targets, reconstructed by a dedicated helper to
	// keep this dispatcher's branch count bounded.
	if effect.HasSecondTargetDamageRider {
		return exactSecondTargetDamageEffectSyntax(effect, prefix, text)
	}
	if len(effect.Targets) != 1 || !effect.Targets[0].Exact {
		return false
	}
	target := effect.Targets[0].Text
	switch effect.Amount.DynamicForm {
	case EffectDynamicAmountFormNone:
		amount := "X"
		if effect.Amount.Known {
			amount = strconv.Itoa(effect.Amount.Value)
		} else if !effect.Amount.VariableX {
			return false
		}
		// "<prefix> <amount> damage to each of <target>." deals the full amount
		// to each of several chosen targets, a genuine multi-target cardinality
		// (Max >= 2). The singular and "up to one" forms (Max <= 1) keep the
		// bare target phrase.
		recipient := target
		if effect.Targets[0].Cardinality.Max >= 2 {
			recipient = "each of " + target
		}
		// A "... and N damage to you" rider follows a single-target (Max <= 1)
		// fixed-amount clause; it is reconstructed only for that bounded shape.
		if effect.HasSelfDamageRider {
			if !effect.Amount.Known || effect.Targets[0].Cardinality.Max >= 2 {
				return false
			}
			return text == fmt.Sprintf("%s %s damage to %s and %d damage to you.",
				prefix, amount, recipient, effect.SelfDamageRiderValue)
		}
		// A "... and N damage to that creature's controller/owner" rider follows
		// a single-target (Max <= 1) fixed-amount clause; the rider recipient is
		// reconstructed from its captured tokens so the round-trip stays exact.
		if effect.TargetControllerDamageRiderRecipient != DamageRecipientReferenceNone {
			if !effect.Amount.Known || effect.Targets[0].Cardinality.Max >= 2 {
				return false
			}
			riderValue, _, riderRecipient := targetControllerDamageRiderTokens(effect)
			if riderRecipient == nil {
				return false
			}
			return text == fmt.Sprintf("%s %s damage to %s and %d damage to %s.",
				prefix, amount, recipient, riderValue, joinedEffectText(riderRecipient))
		}
		return text == fmt.Sprintf("%s %s damage to %s.", prefix, amount, recipient)
	case EffectDynamicAmountFormEqual:
		return text == fmt.Sprintf("%s damage %s to %s.", prefix, effect.Amount.Text, target) ||
			text == fmt.Sprintf("%s damage to %s %s.", prefix, target, effect.Amount.Text)
	case EffectDynamicAmountFormForEach:
		return text == fmt.Sprintf("%s %d damage %s to %s.", prefix, effect.Amount.Multiplier, effect.Amount.Text, target)
	case EffectDynamicAmountFormWhereX:
		return text == fmt.Sprintf("%s X damage to %s, %s.", prefix, target, effect.Amount.Text)
	default:
		return false
	}
}

// exactSecondTargetDamageEffectSyntax reconstructs the canonical two-target
// damage clause "<prefix> A damage to <target0> and B damage to <target1>." in
// which each target is chosen independently. Both targets must reconstruct
// exactly from their own captured phrases and the primary amount must be a fixed
// value, keeping the round-trip exact for this bounded shape.
func exactSecondTargetDamageEffectSyntax(effect *EffectSyntax, prefix, text string) bool {
	if len(effect.Targets) != 2 ||
		!effect.Targets[0].Exact || !effect.Targets[1].Exact ||
		!effect.Amount.Known ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormNone ||
		effect.Targets[0].Cardinality.Max != 1 {
		return false
	}
	return text == fmt.Sprintf("%s %d damage to %s and %d damage to %s.",
		prefix, effect.Amount.Value, effect.Targets[0].Text,
		effect.SecondTargetDamageRiderValue, effect.Targets[1].Text)
}

// exactSourcePowerDamageEffectSyntax reconstructs the canonical one-sided
// source-power damage clauses in which a target creature deals damage equal to
// its own power, either to itself ("Target creature deals damage to itself
// equal to its power.") or to a second target ("Target creature you control
// deals damage equal to its power to target creature you don't control."). The
// dealing creature is the clause subject (itself a target), the amount is that
// creature's power, and the recipient is the same creature or the second
// target. exactDamageEffectSyntax already rejects this shape because its subject
// is a target rather than the spell's own source, so this sibling reconstruction
// handles it. It fails closed for every other wording, including the static-
// group "Each creature deals damage to itself..." form (a non-target subject),
// so unsupported source-power damage keeps failing the round-trip.
func exactSourcePowerDamageEffectSyntax(effect *EffectSyntax) bool {
	if effect.Negated || effect.Divided ||
		effect.Context != EffectContextTarget ||
		effect.Amount.DynamicKind != EffectDynamicAmountSourcePower ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormEqual ||
		effect.Amount.Multiplier != 1 {
		return false
	}
	text := exactEffectClauseText(effect)
	switch len(effect.Targets) {
	case 1:
		if !effect.Targets[0].Exact {
			return false
		}
		return text == fmt.Sprintf("%s deals damage to itself %s.",
			effect.Targets[0].Text, effect.Amount.Text)
	case 2:
		if !effect.Targets[0].Exact || !effect.Targets[1].Exact {
			return false
		}
		return text == fmt.Sprintf("%s deals damage %s to %s.",
			effect.Targets[0].Text, effect.Amount.Text, effect.Targets[1].Text)
	default:
		return false
	}
}

// dividedDamageEffect reports whether a damage effect uses the "divided as you
// choose among <targets>" wording, where a fixed total is split among the
// chosen targets (CR 601.2d).
func dividedDamageEffect(effect *EffectSyntax) bool {
	if effect.Kind != EffectDealDamage {
		return false
	}
	return effectContainsWords(normalizedWords(effect.Tokens), "divided", "as", "you", "choose", "among")
}

// exactDividedDamageText reconstructs the canonical "deals N damage divided as
// you choose among <cardinality> <noun>" clause and compares it byte-for-byte to
// the source. It supports only a fixed total and the cardinality and target
// nouns the executable backend can represent exactly, failing closed otherwise.
func exactDividedDamageText(effect *EffectSyntax, prefix, text string) bool {
	if !effect.Amount.Known || effect.Amount.Value < 1 ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormNone ||
		effect.Amount.DynamicKind != EffectDynamicAmountNone ||
		effect.Negated ||
		len(effect.Targets) != 1 {
		return false
	}
	cardinality, ok := dividedCardinalityPhrase(effect.Targets[0].Cardinality)
	if !ok {
		return false
	}
	noun, ok := dividedTargetNoun(effect.Targets[0].Selection)
	if !ok {
		return false
	}
	expected := fmt.Sprintf("%s %d damage divided as you choose among %s %s.",
		prefix, effect.Amount.Value, cardinality, noun)
	return text == expected
}

// dividedCardinalityPhrase reconstructs the cardinal phrase that introduces the
// divided targets. It supports the enumerated "one or two" and "one, two, or
// three" ranges and the unbounded "any number of" form.
func dividedCardinalityPhrase(cardinality TargetCardinalitySyntax) (string, bool) {
	switch cardinality {
	case TargetCardinalitySyntax{Min: 1, Max: 2}:
		return "one or two", true
	case TargetCardinalitySyntax{Min: 1, Max: 3}:
		return "one, two, or three", true
	case TargetCardinalitySyntax{Min: 0, Max: 99}:
		return "any number of", true
	default:
		return "", false
	}
}

// dividedTargetNoun reconstructs the target noun phrase for divided damage. It
// supports "targets" (any target) and "target creatures" with no further
// qualifiers, failing closed for every other selector.
func dividedTargetNoun(selection SelectionSyntax) (string, bool) {
	switch selection.Kind {
	case SelectionAny:
		return "targets", true
	case SelectionCreature:
		if dividedPlainCreatureSelection(selection) {
			return "target creatures", true
		}
	default:
	}
	return "", false
}

// dividedPlainCreatureSelection reports that a creature selection carries no
// qualifier beyond its card type, so it reconstructs as a bare "target
// creatures" phrase.
func dividedPlainCreatureSelection(selection SelectionSyntax) bool {
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking ||
		selection.Tapped || selection.Untapped ||
		selection.Colorless || selection.Multicolored ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.Keyword != KeywordUnknown ||
		selection.Zone != zone.None ||
		selection.Controller != SelectionControllerAny ||
		len(selection.ExcludedTypes) != 0 ||
		len(selection.ExcludedColors) != 0 ||
		len(selection.ColorsAny) != 0 ||
		len(selection.SubtypesAny) != 0 ||
		len(selection.Supertypes) != 0 {
		return false
	}
	switch len(selection.RequiredTypesAny) {
	case 0:
		return true
	case 1:
		return selection.RequiredTypesAny[0] == CardTypeCreature
	default:
		return false
	}
}

// exactGroupDamageAmountText reconstructs the canonical amount token for a group
// damage clause: the literal integer for a fixed amount of at least one, or "X"
// for the spell's variable X. It fails closed for a non-positive fixed amount
// and for any dynamic amount form ("equal to ...", "where X is ..."), which the
// group damage path reconstructs separately or not at all, so those wordings
// keep failing the round-trip.
func exactGroupDamageAmountText(amount EffectAmountSyntax) (string, bool) {
	if amount.DynamicForm != EffectDynamicAmountFormNone ||
		amount.DynamicKind != EffectDynamicAmountNone {
		return "", false
	}
	switch {
	case amount.Known:
		if amount.Value < 1 {
			return "", false
		}
		return strconv.Itoa(amount.Value), true
	case amount.VariableX:
		return "X", true
	default:
		return "", false
	}
}

// exactGroupDynamicDamageText reconstructs the canonical single-recipient group
// damage clause whose amount is a trailing "where X is the number of ..." count
// phrase or a "equal to ..." dynamic phrase. The amount phrase is reproduced
// verbatim from the captured source so the round-trip stays byte-exact, exactly
// as the single-target dynamic-amount branches do:
//
//	"Chain Reaction deals X damage to each creature, where X is the number of creatures on the battlefield."
//	"Gates Ablaze deals X damage to each creature, where X is the number of Gates you control."
//	"Fanatic of Mogis deals damage to each opponent equal to your devotion to red."
//
// The recipient must be a single filtered group; it fails closed for the
// two-recipient pair form and for any amount form other than WhereX or Equal,
// keeping the dual-recipient and fixed paths unchanged and unsupported wordings
// rejected.
func exactGroupDynamicDamageText(effect *EffectSyntax, prefix, text string) bool {
	if len(effect.DamageRecipientPair) != 0 {
		return false
	}
	recipient, ok := exactGroupDamageRecipientText(effect.Selection)
	if !ok {
		return false
	}
	switch effect.Amount.DynamicForm {
	case EffectDynamicAmountFormWhereX:
		return text == fmt.Sprintf("%s X damage to %s, %s.", prefix, recipient, effect.Amount.Text)
	case EffectDynamicAmountFormEqual:
		return text == fmt.Sprintf("%s damage to %s %s.", prefix, recipient, effect.Amount.Text)
	default:
		return false
	}
}

// exactGroupDamageRecipientText reconstructs the canonical Oracle recipient
// phrase for a fixed-amount group damage spell ("each opponent", "each
// creature your opponents control", "each attacking creature", "each Goblin",
// "each nonartifact creature"). The caller compares the reconstruction against
// the literal source text, so any selector qualifier this renderer cannot
// represent makes the comparison fail and keeps the wording unsupported rather
// than silently dropping or approximating the filter.
func exactGroupDamageRecipientText(selection SelectionSyntax) (string, bool) {
	switch {
	case selection.Kind == SelectionOpponent && !selection.Other:
		return "each opponent", true
	case selection.Kind == SelectionPlayer && !selection.Other:
		return "each player", true
	}
	return exactGroupDamagePermanentRecipientText(selection)
}

// exactGroupDamagePermanentRecipientText reconstructs the recipient phrase for a
// group damage spell whose recipients form a single filtered permanent group. It
// renders only the controller, combat, tapped, single-color, single-subtype,
// single-excluded-type, keyword, and "other" qualifiers the executable backend
// can represent exactly, and fails closed for every other qualifier.
func exactGroupDamagePermanentRecipientText(selection SelectionSyntax) (string, bool) {
	if selection.All || selection.Another || selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.Colorless || selection.Multicolored ||
		len(selection.Supertypes) != 0 ||
		len(selection.ExcludedColors) != 0 ||
		len(selection.RequiredTypesAny) > 1 ||
		len(selection.ColorsAny) > 1 ||
		len(selection.SubtypesAny) > 1 ||
		len(selection.ExcludedTypes) > 1 {
		return "", false
	}
	if (selection.Attacking && selection.Blocking) ||
		(selection.Tapped && selection.Untapped) ||
		((selection.Tapped || selection.Untapped) && (selection.Attacking || selection.Blocking)) {
		return "", false
	}
	noun, hasNoun := permanentSelectionNoun(selection.Kind)
	if !hasNoun && selection.Kind != SelectionUnknown {
		return "", false
	}
	// The parser records a permanent noun both as the selection Kind and as a
	// redundant single-element RequiredTypesAny. Accept only that redundant form
	// (a union or a type inconsistent with the noun is not representable here).
	if len(selection.RequiredTypesAny) == 1 {
		requiredNoun, ok := permanentCardTypeNoun(selection.RequiredTypesAny[0])
		if !ok || !hasNoun || requiredNoun != noun {
			return "", false
		}
	}
	words := []string{"each"}
	if selection.Other {
		words = append(words, "other")
	}
	switch {
	case selection.Attacking:
		words = append(words, "attacking")
	case selection.Blocking:
		words = append(words, "blocking")
	case selection.Tapped:
		words = append(words, "tapped")
	case selection.Untapped:
		words = append(words, "untapped")
	default:
	}
	if len(selection.ColorsAny) == 1 {
		colorText, ok := colorWord(selection.ColorsAny[0])
		if !ok {
			return "", false
		}
		words = append(words, colorText)
	}
	if len(selection.SubtypesAny) == 1 {
		words = append(words, string(selection.SubtypesAny[0]))
	}
	if len(selection.ExcludedTypes) == 1 {
		if !hasNoun {
			return "", false
		}
		excludedNoun, ok := permanentCardTypeNoun(selection.ExcludedTypes[0])
		if !ok {
			return "", false
		}
		words = append(words, "non"+excludedNoun)
	}
	if hasNoun {
		words = append(words, noun)
	} else if len(selection.SubtypesAny) != 1 {
		return "", false
	}
	// The canonical Oracle ordering places the controller clause immediately
	// after the noun and before any "with"/"without" keyword qualifier, e.g.
	// "each creature you control with flying". Rendering the controller clause
	// here, ahead of the keyword clause, keeps those combined group recipients
	// byte-exact.
	switch selection.Controller {
	case SelectionControllerAny:
	case SelectionControllerYou:
		words = append(words, "you", "control")
	case SelectionControllerOpponent:
		words = append(words, "your", "opponents", "control")
	case SelectionControllerNotYou:
		words = append(words, "you", "don't", "control")
	default:
		return "", false
	}
	if selection.Keyword != KeywordUnknown {
		keywordWord, ok := selection.Keyword.OracleWord()
		if !ok {
			return "", false
		}
		words = append(words, "with", keywordWord)
	}
	if selection.ExcludedKeyword != KeywordUnknown {
		if selection.Keyword != KeywordUnknown {
			return "", false
		}
		keywordWord, ok := selection.ExcludedKeyword.OracleWord()
		if !ok {
			return "", false
		}
		words = append(words, "without", keywordWord)
	}
	return strings.Join(words, " "), true
}
