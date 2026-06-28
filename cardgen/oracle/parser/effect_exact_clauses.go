package parser

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
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
	if len(effect.Targets) != 1 {
		return false
	}
	if effect.Amount.VariableX {
		return exactVariableXGraveyardReturnEffectSyntax(effect, text)
	}
	if !exactGraveyardCardTargetSyntax(&effect.Targets[0]) {
		return false
	}
	return matchesGraveyardReturnDestination(text, "Return "+effect.Targets[0].Text)
}

// graveyardReturnDestinationSuffixes are the canonical destination clauses that
// follow a graveyard-card return's target noun phrase. An owner-relative hand
// destination ("to its owner's hand", the plural "to their owners' hands", or
// the opponent-graveyard "to their hand") returns a graveyard card the
// controller does not own (a card targeted "from a graveyard" or "from an
// opponent's graveyard"). The runtime MoveCard handler always routes a returned
// card to its own owner's hand, so these lower identically to the "to your
// hand." form.
var graveyardReturnDestinationSuffixes = []string{
	" to your hand.",
	" to its owner's hand.",
	" to their owners' hands.",
	" to their hand.",
	" to the battlefield.",
	" to the battlefield tapped.",
	" to the battlefield under your control.",
	" to the battlefield tapped under your control.",
	" on top of your library.",
	" on the top of your library.",
	" on bottom of your library.",
	" on the bottom of your library.",
}

// matchesGraveyardReturnDestination reports whether text is the return clause
// prefix joined with one canonical destination suffix.
func matchesGraveyardReturnDestination(text, prefix string) bool {
	for _, suffix := range graveyardReturnDestinationSuffixes {
		if strings.EqualFold(text, prefix+suffix) {
			return true
		}
	}
	return false
}

// exactVariableXGraveyardReturnEffectSyntax reports whether effect is the
// variable-count graveyard-card return "Return X target <noun> cards from
// <owner> graveyard to <destination>", whose chosen {X} fixes how many cards are
// returned. The compiler carries the X count on the effect amount and leaves the
// lone target a singular cardinality, so this path reconstructs the target's
// plural noun phrase ("X target creature cards from your graveyard") and the full
// clause byte-for-byte, accepting the same destinations and noun filters the
// fixed targeted return does. It fails closed for every qualifier those shared
// reconstructors reject, so an unrepresentable variable-count return keeps
// failing rather than lowering to a wrong predicate.
func exactVariableXGraveyardReturnEffectSyntax(effect *EffectSyntax, text string) bool {
	targetText, ok := exactVariableXGraveyardTargetText(&effect.Targets[0])
	if !ok {
		return false
	}
	return matchesGraveyardReturnDestination(text, "Return X "+targetText)
}

// exactVariableXGraveyardTargetText reconstructs the canonical plural target noun
// phrase of a variable-count graveyard-card return ("target creature cards from
// your graveyard") from the target's typed selection and reports whether it
// matches target.Text. It accepts the same graveyard noun filters as
// exactGraveyardCardTargetSyntax (a card-type, a union of card types, a permanent
// card, a single color, a colorless or multicolored card, a single subtype, a
// single supertype, or the plain "card" noun, with an optional power, toughness,
// or "with mana value N or less" qualifier) and fails closed for every other
// shape, including any multi-target or "up to N" cardinality, whose count words
// the variable-count "X target " prefix replaces.
func exactVariableXGraveyardTargetText(target *TargetSyntax) (string, bool) {
	sel := target.Selection
	if sel.Zone != zone.Graveyard || sel.Other {
		return "", false
	}
	if target.Cardinality != (TargetCardinalitySyntax{Min: 1, Max: 1}) || sel.Another {
		return "", false
	}
	if sel.All || sel.Attacking || sel.Blocking || sel.Tapped || sel.Untapped ||
		sel.Keyword != KeywordUnknown ||
		len(sel.SourceTypes) != 0 ||
		len(sel.ExcludedSupertypes) != 0 || len(sel.ExcludedColors) != 0 {
		return "", false
	}
	owner, ok := graveyardOwnerSuffix(sel.Controller)
	if !ok {
		return "", false
	}
	noun, ok := graveyardCardNoun(sel, true)
	if !ok {
		return "", false
	}
	noun += "s"
	qualifier, ok := graveyardNumericQualifier(sel)
	if !ok {
		return "", false
	}
	reconstructed := "target " + noun + qualifier + " " + owner
	if !strings.EqualFold(target.Text, reconstructed) {
		return "", false
	}
	return reconstructed, true
}

// exactChosenGraveyardReturnEffectSyntax recognizes the non-target "Return a
// <filter> card from your graveyard to your hand." recursion wording and the
// reanimation form "Return a <filter> card from your graveyard to the
// battlefield." (optionally "tapped" and/or "under your control"), where the
// returned card is chosen from the controller's own graveyard at resolution
// rather than targeted (Raise Dead targets; Takenuma's "return a creature or
// planeswalker card" does not). It reconstructs the canonical noun phrase from
// the effect's typed Selection the same way the targeted path does, accepting a
// single card-type, a union of card types, a permanent card, a single color, a
// colorless or multicolored card, a single subtype, a single supertype adjective,
// or the plain "card" noun, with an optional power, toughness, or "with mana value
// N or less" qualifier, and fails closed for every other selection shape so an
// unrepresentable filter keeps failing rather than lowering to a wrong predicate.
func exactChosenGraveyardReturnEffectSyntax(effect *EffectSyntax, text string) bool {
	if len(effect.References) != 0 {
		return false
	}
	if effect.ToZone != zone.Hand && effect.ToZone != zone.Battlefield {
		return false
	}
	sel := effect.Selection
	if sel.Zone != zone.Graveyard || sel.Controller != SelectionControllerYou {
		return false
	}
	// A battlefield entry-tapped rider ("... to the battlefield tapped.") leaves
	// the entry word inside the selector span, setting sel.Tapped; graveyard
	// cards are never tapped, so that filter is vacuous and is ignored when it
	// coincides with the entry-tapped destination. A genuine tapped filter
	// without entry-tapped still fails closed.
	entryTapped := effect.ToZone == zone.Battlefield && effect.EntersTapped
	if sel.All || sel.Another || sel.Other || sel.Attacking || sel.Blocking ||
		(sel.Tapped && !entryTapped) || sel.Untapped ||
		sel.Keyword != KeywordUnknown || sel.ExcludedKeyword != KeywordUnknown ||
		len(sel.SourceTypes) != 0 ||
		len(sel.ExcludedSupertypes) != 0 ||
		len(sel.ExcludedColors) != 0 || len(sel.Alternatives) != 0 {
		return false
	}
	noun, ok := graveyardCardNoun(sel, false)
	if !ok {
		return false
	}
	manaClause, ok := graveyardNumericQualifier(sel)
	if !ok {
		return false
	}
	article := indefiniteArticle(noun)
	prefix := "Return " + article + " " + noun + manaClause + " from your graveyard to "
	switch effect.ToZone {
	case zone.Hand:
		if effect.EntersTapped || effect.UnderYourControl {
			return false
		}
		return strings.EqualFold(text, prefix+"your hand.")
	case zone.Battlefield:
		destination := "the battlefield"
		if effect.EntersTapped {
			destination += " tapped"
		}
		if effect.UnderYourControl {
			destination += " under your control"
		}
		return strings.EqualFold(text, prefix+destination+".")
	default:
		return false
	}
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

// exactSourceSpellShuffleIntoLibrarySyntax recognizes the exact resolution tail
// "Shuffle <this card> into its owner's library." (Green Sun's Zenith, the
// Beacon cycle, Blue Sun's Zenith), where the shuffled object is the resolving
// spell itself named by "this card"/"this spell" or its own name. Lowering turns
// it into a single source-spell shuffle-into-library instruction. The parser
// owns this wording; any other shuffle clause leaves it non-exact so lowering
// fails closed.
func exactSourceSpellShuffleIntoLibrarySyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectShuffle ||
		effect.Negated ||
		effect.Duration != EffectDurationNone ||
		effect.FromZone != zone.None ||
		effect.ToZone != zone.Library ||
		len(effect.Targets) != 0 ||
		len(effect.References) != 2 {
		return false
	}
	object := effect.References[0]
	if object.Kind != ReferenceThisObject && object.Kind != ReferenceSelfName {
		return false
	}
	if object.Kind == ReferenceThisObject && object.Text != "this spell" && object.Text != "this card" {
		return false
	}
	owner := effect.References[1]
	if owner.Kind != ReferencePronoun || owner.Pronoun != PronounIts {
		return false
	}
	return effect.Text == "Shuffle "+object.Text+" into its owner's library."
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

// exactExileUntilSourceLeavesEffectSyntax recognizes the O-Ring exile clause
// "exile <target> until <this permanent> leaves the battlefield." (Banisher
// Priest, Banishing Light, Fairgrounds Warden) and its "you may exile ..."
// optional offer (Angel of Sanctions). The single target is the exiled
// permanent and the trailing "until <self> leaves the battlefield" names the
// source permanent as the duration anchor, not a second object. It marks the
// effect so lowering links the exile to the source and synthesizes the paired
// leaves-the-battlefield return trigger. The optional "you may" prefix is
// carried by effect.Optional and stripped by exactEffectClauseText before the
// clause-text comparison, so it surfaces later as the trigger's Optional flag.
// The parser owns this wording; any other exile shape leaves the clause
// non-exact so lowering fails closed.
func exactExileUntilSourceLeavesEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectExile || effect.Negated {
		return false
	}
	if effect.Duration != EffectDurationNone || effect.FromZone != zone.None || effect.ToZone != zone.None {
		return false
	}
	if len(effect.Targets) != 1 || !effect.Targets[0].Exact {
		return false
	}
	if len(effect.References) != 1 {
		return false
	}
	reference := effect.References[0]
	if reference.Kind != ReferenceThisObject && reference.Kind != ReferenceSelfName {
		return false
	}
	expected := "Exile " + effect.Targets[0].Text + " until " + reference.Text + " leaves the battlefield."
	if !strings.EqualFold(exactEffectClauseText(effect), expected) {
		return false
	}
	effect.ExileUntilSourceLeaves = true
	return true
}

// exactExileForEachPlayerUntilLeavesEffectSyntax recognizes the distributive
// Saga exile clause "For each player, exile up to one [other] target
// <permanent> that player controls until <this Saga> leaves the battlefield."
// (Vault 13: Dweller's Journey). The leading "For each player," distributes a
// single "up to one" target pool across every player; the controller chooses
// one eligible permanent per player at resolution and the exiled permanents are
// linked to the source so a paired chapter returns them. The "that player"
// reference is the distributive anchor and the trailing self-reference is the
// duration anchor, neither a second object.
//
// effectSubjectStart drops the "For each player," prefix from the reconstructed
// clause text, so the recognizer confirms that prefix on the raw effect text and
// rebuilds the remainder from the single target and source anchor. The trailing
// "until this Saga leaves the battlefield" contributes the source's own "Saga"
// subtype to the parsed selection; because the clause is matched by exact
// wording, that spurious subtype is removed so the candidate filter is the
// printed "[other] <permanent>" rather than "Saga". Any other exile shape leaves
// the clause non-exact so lowering fails closed.
func exactExileForEachPlayerUntilLeavesEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectExile || effect.Negated || effect.Optional {
		return false
	}
	if effect.Context != EffectContextController {
		return false
	}
	if effect.Duration != EffectDurationNone || effect.FromZone != zone.None || effect.ToZone != zone.None {
		return false
	}
	if len(effect.Targets) != 1 {
		return false
	}
	if effect.Targets[0].Cardinality.Min != 0 || effect.Targets[0].Cardinality.Max != 1 {
		return false
	}
	sourceRef, ok := exileForEachPlayerReferences(effect.References)
	if !ok {
		return false
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(effect.Text)), "for each player, ") {
		return false
	}
	expected := "Exile " + effect.Targets[0].Text + " until " + sourceRef.Text + " leaves the battlefield."
	if !strings.EqualFold(exactEffectClauseText(effect), expected) {
		return false
	}
	stripSourceSubtypeContamination(&effect.Selection, sourceRef)
	effect.ExileForEachPlayerUntilSourceLeaves = true
	return true
}

// exileForEachPlayerReferences confirms the distributive exile clause carries
// exactly the two anchors its wording requires: a "that player" reference (the
// per-player distribution anchor) and a self reference (the duration anchor). It
// returns the source reference so the caller can rebuild and compare the
// trailing "until <self> leaves the battlefield" phrase.
func exileForEachPlayerReferences(references []Reference) (Reference, bool) {
	if len(references) != 2 {
		return Reference{}, false
	}
	var thatPlayer, sourceRef *Reference
	for index := range references {
		switch references[index].Kind {
		case ReferenceThatPlayer:
			thatPlayer = &references[index]
		case ReferenceThisObject, ReferenceSelfName:
			sourceRef = &references[index]
		default:
		}
	}
	if thatPlayer == nil || sourceRef == nil {
		return Reference{}, false
	}
	return *sourceRef, true
}

// stripSourceSubtypeContamination removes the source permanent's own printed
// subtypes from a parsed selection's any-of subtype filter. The distributive
// exile clause ends with "until this Saga leaves the battlefield", which leaks
// the source's "Saga" subtype into the exile selection; once the clause is
// matched by exact wording the leaked subtype is dropped so the candidate filter
// reflects the printed "[other] <permanent>" wording rather than the duration
// phrase. Only subtypes named verbatim in the source reference text are removed.
func stripSourceSubtypeContamination(selection *SelectionSyntax, sourceRef Reference) {
	if len(selection.SubtypesAny) == 0 {
		return
	}
	sourceWords := strings.Fields(strings.ToLower(sourceRef.Text))
	selection.SubtypesAny = slices.DeleteFunc(selection.SubtypesAny, func(subtype types.Sub) bool {
		return slices.Contains(sourceWords, strings.ToLower(string(subtype)))
	})
}

// exactReturnExiledCardEffectSyntax recognizes the explicit O-Ring leaves-the-
// battlefield clause "return the exiled card to the battlefield under its
// owner's control." (Oblivion Ring, Journey to Nowhere, Fiend Hunter). The
// returned card is the one a sibling enters-the-battlefield exile removed,
// identified by the source link, so the effect carries no target. It marks the
// effect so lowering emits the linked battlefield return; any other return shape
// leaves the clause non-exact so lowering fails closed.
func exactReturnExiledCardEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectReturn || effect.Negated || effect.Optional {
		return false
	}
	if effect.Context != EffectContextController || !effect.UnderOwnersControl {
		return false
	}
	if effect.ToZone != zone.Battlefield || effect.FromZone != zone.None {
		return false
	}
	if len(effect.Targets) != 0 {
		return false
	}
	if !strings.EqualFold(exactEffectClauseText(effect),
		"Return the exiled card to the battlefield under its owner's control.") {
		return false
	}
	effect.ReturnExiledCard = true
	return true
}

// exactExileEntireHandEffectSyntax recognizes the involuntary whole-hand exile
// clause "Exile all cards from your hand." (Wormfang Behemoth). The whole hand
// moves to exile with no choice, and the exiled set is linked to the source so a
// paired leaves-the-battlefield "return the exiled cards" trigger returns it; the
// clause carries no target. It marks the effect so lowering emits the linked
// entire-hand exile; any other exile shape leaves the clause non-exact so
// lowering fails closed.
func exactExileEntireHandEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectExile || effect.Negated || effect.Optional {
		return false
	}
	if effect.Context != EffectContextController || effect.FromZone != zone.Hand {
		return false
	}
	if len(effect.Targets) != 0 {
		return false
	}
	if !strings.EqualFold(exactEffectClauseText(effect), "Exile all cards from your hand.") {
		return false
	}
	effect.ExileEntireHand = true
	return true
}

// exactReturnExiledCardsToHandEffectSyntax recognizes the leaves-the-battlefield
// clause "Return the exiled cards to their owner's hand." (Wormfang Behemoth).
// The returned cards are the set a sibling entire-hand exile removed, identified
// by the source link rather than a target, so the effect carries no target. It
// marks the effect so lowering emits the linked return to hand; any other return
// shape leaves the clause non-exact so lowering fails closed.
func exactReturnExiledCardsToHandEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectReturn || effect.Negated || effect.Optional {
		return false
	}
	if effect.Context != EffectContextController {
		return false
	}
	if effect.ToZone != zone.Hand || effect.FromZone != zone.None {
		return false
	}
	if len(effect.Targets) != 0 {
		return false
	}
	if !strings.EqualFold(exactEffectClauseText(effect),
		"Return the exiled cards to their owner's hand.") {
		return false
	}
	effect.ReturnExiledCardsToHand = true
	return true
}

// exactBottomLinkedExiledCardsEffectSyntax reports whether a put-into-library
// effect is the linked disposal clause "The owner of each card exiled with
// <this permanent> puts that card on the bottom of their library." (Trial of a
// Time Lord). The disposed cards are the ones a sibling exile-until-leaves
// clause removed, identified by the source link, so the effect carries no
// target. It marks the effect so lowering emits the linked library-bottom
// disposal; any other put shape leaves the clause non-exact so lowering fails
// closed.
func exactBottomLinkedExiledCardsEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectPut || effect.Negated || effect.Optional {
		return false
	}
	if effect.Context != EffectContextSource || effect.ToZone != zone.Library {
		return false
	}
	if len(effect.Targets) != 0 {
		return false
	}
	anchor, ok := exiledWithSelfAnchorText(effect)
	if !ok {
		return false
	}
	canonical := "The owner of each card exiled with " + anchor +
		" puts that card on the bottom of their library."
	if !strings.EqualFold(exactEffectClauseText(effect), canonical) {
		return false
	}
	effect.BottomLinkedExiledCards = true
	return true
}

// exactReturnLinkedExiledToBattlefieldPartialEffectSyntax recognizes the Saga
// chapter clause "Return <count> cards exiled with <this Saga> to the
// battlefield under their owners' control." (Vault 13: Dweller's Journey). The
// returned cards are a fixed-size subset, chosen at resolution, of the set a
// sibling distributive exile clause linked to the source, so the effect carries
// no target and reads its source through the link rather than a printed object.
// The spelled count is rebuilt from the parsed amount so only the exact printed
// number matches. It marks the effect so lowering emits the partial linked
// battlefield return; any other return shape leaves the clause non-exact so
// lowering fails closed.
func exactReturnLinkedExiledToBattlefieldPartialEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectReturn || effect.Negated || effect.Optional {
		return false
	}
	if effect.Context != EffectContextController || !effect.UnderOwnersControl {
		return false
	}
	if effect.ToZone != zone.Battlefield || effect.FromZone != zone.None {
		return false
	}
	if len(effect.Targets) != 0 {
		return false
	}
	if !effect.Amount.Known || effect.Amount.Value < 1 {
		return false
	}
	count, ok := cardinalWord(effect.Amount.Value)
	if !ok {
		return false
	}
	anchor, ok := linkedExiledSourceAnchorText(effect)
	if !ok {
		return false
	}
	canonical := "Return " + count + " cards exiled with " + anchor +
		" to the battlefield under their owners' control."
	if !strings.EqualFold(exactEffectClauseText(effect), canonical) {
		return false
	}
	effect.ReturnLinkedExiledToBattlefieldPartial = true
	return true
}

// linkedExiledSourceAnchorText returns the self-reference text ("this Saga")
// that names the source link of a linked-exile return or disposal clause,
// scanning both the effect's references and subject references because the
// anchor may appear in either set depending on the surrounding sentence.
func linkedExiledSourceAnchorText(effect *EffectSyntax) (string, bool) {
	for _, references := range [][]Reference{effect.References, effect.SubjectReferences} {
		for _, reference := range references {
			if reference.Kind != ReferenceThisObject && reference.Kind != ReferenceSelfName {
				continue
			}
			if text := strings.TrimSpace(reference.Text); text != "" {
				return text, true
			}
		}
	}
	return "", false
}

// exactPutLinkedExiledRestOnLibraryBottomEffectSyntax recognizes the Saga
// chapter disposal clause "put the rest on the bottom of their owners'
// libraries." (Vault 13: Dweller's Journey). The disposed cards are the linked
// exiled set a sibling partial-return clause did not bring back, identified by
// the source link, so the effect carries no target. It marks the effect so
// lowering routes the unreturned remainder to the bottom of their owners'
// libraries; any other put shape leaves the clause non-exact so lowering fails
// closed.
func exactPutLinkedExiledRestOnLibraryBottomEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectPut || effect.Negated || effect.Optional {
		return false
	}
	if effect.Context != EffectContextController {
		return false
	}
	if len(effect.Targets) != 0 {
		return false
	}
	if !strings.EqualFold(exactEffectClauseText(effect),
		"put the rest on the bottom of their owners' libraries.") {
		return false
	}
	effect.PutLinkedExiledRestOnLibraryBottom = true
	return true
}

// exactCounterExiledCardManaValueEffectSyntax recognizes the chapter II clause
// "Put a number of +1/+1 counters on target creature you control equal to the
// mana value of the exiled card." (The Aesir Escape Valhalla). The count is the
// mana value of the card a sibling chapter exiled under the source link, read
// through that link rather than a printed number, so the parser drops the amount
// span and the generic counter recognizer cannot match. The target is kept; the
// effect is marked so lowering scales the placement by the linked exiled card's
// mana value. Any other counter shape leaves the clause non-exact so lowering
// fails closed.
func exactCounterExiledCardManaValueEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectPut || effect.Negated || effect.Optional {
		return false
	}
	if !effect.CounterKnown {
		return false
	}
	if len(effect.Targets) != 1 || !effect.Targets[0].Exact {
		return false
	}
	target := effect.Targets[0]
	if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
		return false
	}
	canonical := "Put a number of " + effect.CounterKind.String() + " counters on " +
		target.Text + " equal to the mana value of the exiled card."
	if !strings.EqualFold(exactEffectClauseText(effect), canonical) {
		return false
	}
	effect.CounterExiledCardManaValue = true
	return true
}

// exactReturnSourceAndExiledCardToHandEffectSyntax recognizes the chapter III
// clause "Return this Saga and the exiled card to their owner's hand." (The
// Aesir Escape Valhalla). It returns both the source permanent and the card a
// sibling chapter exiled under the source link, identified by the link rather
// than a target, so the effect carries no target. It marks the effect so
// lowering emits a source bounce paired with a linked return to hand; any other
// return shape leaves the clause non-exact so lowering fails closed.
func exactReturnSourceAndExiledCardToHandEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectReturn || effect.Negated || effect.Optional {
		return false
	}
	if effect.Context != EffectContextController {
		return false
	}
	if effect.ToZone != zone.Hand || effect.FromZone != zone.None {
		return false
	}
	if len(effect.Targets) != 0 {
		return false
	}
	anchor, ok := selfReferenceAnchorText(effect)
	if !ok {
		return false
	}
	canonical := "Return " + anchor + " and the exiled card to their owner's hand."
	if !strings.EqualFold(exactEffectClauseText(effect), canonical) {
		return false
	}
	effect.ReturnSourceAndExiledCardToHand = true
	return true
}

// selfReferenceAnchorText returns the source-anchor wording ("this Saga") that a
// clause names as its own permanent, drawn from a this-object or self-name
// reference in either the subject or the effect references. It reports false
// when no such anchor is present.
func selfReferenceAnchorText(effect *EffectSyntax) (string, bool) {
	for _, group := range [][]Reference{effect.SubjectReferences, effect.References} {
		for _, reference := range group {
			if reference.Kind != ReferenceThisObject && reference.Kind != ReferenceSelfName {
				continue
			}
			if text := strings.TrimSpace(reference.Text); text != "" {
				return text, true
			}
		}
	}
	return "", false
}

// exiledWithSelfAnchorText returns the source-anchor wording ("this Saga") that
// the linked disposal clause names as the exile source, drawn from a subject
// this-object or self-name reference. It reports false when no such anchor is
// present so the recognizer cannot match an unanchored clause.
func exiledWithSelfAnchorText(effect *EffectSyntax) (string, bool) {
	for _, reference := range effect.SubjectReferences {
		if reference.Kind != ReferenceThisObject && reference.Kind != ReferenceSelfName {
			continue
		}
		if text := strings.TrimSpace(reference.Text); text != "" {
			return text, true
		}
	}
	return "", false
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

// exactCounteredSpellDestinationSyntax recognizes the counter rider "If that
// spell is countered this way, put it [on top of its owner's library | into its
// owner's hand] instead of into that player's graveyard." (Memory Lapse, Lapse
// of Certainty, Remand) and marks it so a preceding counter effect lowers to a
// single CounterObject that redirects the countered spell to the named zone. The
// parser owns this wording; any other destination phrasing leaves the clause
// non-exact so lowering fails closed.
func exactCounteredSpellDestinationSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectPut || effect.Negated {
		return false
	}
	var destination string
	switch {
	case effect.ToZone == zone.Library && effect.Destination == EffectDestinationTop:
		destination = "put it on top of its owner's library"
	case effect.ToZone == zone.Hand && effect.Destination == EffectDestinationUnspecified:
		destination = "put it into its owner's hand"
	default:
		return false
	}
	expected := "If that spell is countered this way, " + destination + " instead of into that player's graveyard."
	if !strings.EqualFold(strings.TrimSpace(effect.Text), expected) {
		return false
	}
	effect.CounteredSpellDestinationReplacement = true
	return true
}

// exactGraveyardPutEffectSyntax reports whether a "Put <graveyard card>"
// destination clause reconstructs one of the supported destinations exactly.
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
		" on top of its owner's library.",
		" on the top of its owner's library.",
		" on bottom of its owner's library.",
		" on the bottom of its owner's library.",
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
// single subtype, or the plain "card" noun, with an optional power, toughness, or
// "with mana value N or less" qualifier and an optional multi-target or "up to N"
// count. It also accepts a single supertype adjective ("legendary creature card",
// "snow land card"). It fails closed for every qualifier the canonical phrasing
// cannot render (keyword, excluded supertype, excluded types or colors,
// combinations), so an unrepresentable target keeps failing rather than lowering
// to a wrong predicate.
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
		sel.Keyword != KeywordUnknown ||
		len(sel.SourceTypes) != 0 ||
		len(sel.ExcludedSupertypes) != 0 || len(sel.ExcludedColors) != 0 {
		return false
	}
	owner, ok := graveyardOwnerSuffix(sel.Controller)
	if !ok {
		return false
	}
	// "from a single graveyard" restricts every chosen card to one graveyard. It
	// only attaches to the any-graveyard owner relation; "your"/"an opponent's"
	// graveyard already names one graveyard, so a "single" flag there has no
	// canonical wording and fails closed.
	if sel.SingleGraveyard {
		if sel.Controller != SelectionControllerAny {
			return false
		}
		owner = "from a single graveyard"
	}
	prefix, plural, ok := graveyardCardCardinalityPrefix(target.Cardinality, sel.Another)
	if !ok {
		return false
	}
	noun, ok := graveyardCardNoun(sel, plural)
	if !ok {
		return false
	}
	if plural {
		noun += "s"
	}
	manaClause, ok := graveyardNumericQualifier(sel)
	if !ok {
		return false
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
// multiTargetCardinalityPrefix for "up to one ", "up to <N> ", and "<N> ". The
// enumerated lower-bounded ranges "one or two" and "one, two, or three" render
// their explicit count words. It fails closed for a self-exclusion combined with
// a multi-target count, which has no canonical phrasing.
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
	switch c {
	case TargetCardinalitySyntax{Min: 1, Max: 2}:
		return "one or two target ", true, true
	case TargetCardinalitySyntax{Min: 1, Max: 3}:
		return "one, two, or three target ", true, true
	}
	countPrefix, plural, ok := multiTargetCardinalityPrefix(c)
	if !ok {
		return "", false, false
	}
	return countPrefix + "target ", plural, true
}

// graveyardCardNoun reconstructs the singular graveyard-card noun ("creature
// card", "sorcery card", "creature or enchantment card", "permanent card",
// "green card", "red sorcery card", "multicolored creature card", "colorless
// card", "Zombie card", "legendary creature card", or the plain "card") from
// selection's typed fields. It accepts an optional single supertype adjective
// ("legendary", "snow") or single color qualifier (a single color, "colorless",
// or "multicolored") followed by at most one type/subtype/permanent core,
// rendered in canonical Oracle order, and fails closed for any combination it
// could not reconstruct, including a supertype combined with a color or "historic"
// qualifier (which has no single canonical word order). A plural multi-target
// return joins a card-type union with the "and/or" conjunction the Oracle
// templating uses for plural counts ("instant and/or sorcery cards") rather than
// the singular "or".
func graveyardCardNoun(sel SelectionSyntax, plural bool) (string, bool) {
	colorPrefix, hasColor, ok := graveyardColorPrefix(sel)
	if !ok {
		return "", false
	}
	excludedPrefix, hasExcluded, ok := graveyardExcludedTypePrefix(sel)
	if !ok {
		return "", false
	}

	hasTypes := len(sel.RequiredTypesAny) > 0
	hasSubtype := len(sel.SubtypesAny) > 0
	isPermanent := sel.Kind == SelectionPermanent
	// One or more subtype adjectives may qualify a card-type noun ("Zombie
	// creature card", "Angel or Human creature card", "Vampire or Wizard creature
	// cards"); they precede the type noun in canonical Oracle order and count as
	// one combined core with that type rather than as a second independent core.
	// The permanent noun has no subtype-qualified Oracle phrasing, so a subtype
	// never combines with it.
	subtypeQualifiesType := hasTypes && hasSubtype && !isPermanent
	cores := 0
	for _, present := range []bool{hasTypes, hasSubtype, isPermanent} {
		if present {
			cores++
		}
	}
	if cores > 1 && !subtypeQualifiesType {
		return "", false
	}

	var core string
	switch {
	case subtypeQualifiesType:
		typeNoun, ok := graveyardCardTypeNoun(sel, plural)
		if !ok {
			return "", false
		}
		core = graveyardSubtypeAdjective(sel, plural) + " " + typeNoun
	case hasTypes:
		core, ok = graveyardCardTypeNoun(sel, plural)
		if !ok {
			return "", false
		}
	case isPermanent:
		core = "permanent card"
	case hasSubtype:
		core = graveyardSubtypeAdjective(sel, plural) + " card"
	default:
		// The plain "card" noun requires the generic card kind, unless a color
		// qualifier or the "historic" qualifier already restricts it ("green
		// card", "colorless card", "historic card").
		if sel.Kind != SelectionCard && !hasColor && !sel.Historic {
			return "", false
		}
		core = "card"
	}
	// A "historic" qualifier ("historic card", "historic permanent card")
	// precedes the core noun, after any color qualifier, in canonical Oracle
	// order.
	historicPrefix := ""
	if sel.Historic {
		historicPrefix = "historic "
	}
	supertypePrefix, hasSupertype, ok := graveyardSupertypePrefix(sel)
	if !ok {
		return "", false
	}
	// A supertype adjective ("legendary creature card", "snow land card") has no
	// single canonical order relative to a color word in Oracle text (both
	// "legendary white" and "white legendary" occur), so the two never combine
	// here; the "historic" qualifier likewise renders no supertype phrasing.
	if hasSupertype && (hasColor || sel.Historic) {
		return "", false
	}
	// An excluded-card-type prefix ("nonland permanent card", "noncreature
	// artifact card", "noncreature, nonland card") leads the noun. No printed
	// card combines an excluded card type with a color, supertype, historic, or
	// subtype qualifier, so those have no canonical combined order and fail
	// closed rather than guessing one.
	if hasExcluded && (hasColor || hasSupertype || sel.Historic || hasSubtype) {
		return "", false
	}
	return excludedPrefix + supertypePrefix + colorPrefix + historicPrefix + core, true
}

// graveyardExcludedTypePrefix renders the optional leading excluded-card-type
// qualifier of a graveyard-card noun: a single excluded type ("nonland ",
// "noncreature ") or a comma-joined pair ("noncreature, nonland ") rendered in
// the selection's stored order, each "non"-prefixed and followed by a trailing
// space so the caller appends the core noun directly. The runtime Selection
// matches it through the order-independent ExcludedTypes filter. It reports
// whether any excluded card type was present and fails closed for more than two
// excluded types or an unknown card type, which have no canonical Oracle wording.
func graveyardExcludedTypePrefix(sel SelectionSyntax) (prefix string, hasExcluded, ok bool) {
	if len(sel.ExcludedTypes) == 0 {
		return "", false, true
	}
	if len(sel.ExcludedTypes) > 2 {
		return "", false, false
	}
	words := make([]string, 0, len(sel.ExcludedTypes))
	for _, cardType := range sel.ExcludedTypes {
		word, ok := cardTypeWord(cardType)
		if !ok {
			return "", false, false
		}
		words = append(words, "non"+word)
	}
	return strings.Join(words, ", ") + " ", true, true
}

// graveyardSupertypePrefix renders the optional leading supertype qualifier of a
// graveyard-card noun ("legendary ", "snow ", "basic ", "world "), followed by a
// trailing space so the caller appends the core noun directly. It reports whether
// a supertype qualifier was present and fails closed for more than one supertype
// or an unknown supertype word, which have no canonical single-adjective
// phrasing.
func graveyardSupertypePrefix(sel SelectionSyntax) (prefix string, hasSupertype, ok bool) {
	if len(sel.Supertypes) == 0 {
		return "", false, true
	}
	if len(sel.Supertypes) != 1 {
		return "", false, false
	}
	word, ok := supertypeWord(sel.Supertypes[0])
	if !ok {
		return "", false, false
	}
	return word + " ", true, true
}

// graveyardColorPrefix renders the optional leading color qualifier of a
// graveyard-card noun: a single color word ("green "), "colorless ", or
// "multicolored ", each followed by a trailing space so the caller appends the
// core noun directly. It reports whether any color qualifier was present and
// fails closed when more than one color signal is set (e.g. a color list, or a
// color combined with colorless/multicolored), which has no canonical phrasing.
func graveyardColorPrefix(sel SelectionSyntax) (prefix string, hasColor, ok bool) {
	hasColors := len(sel.ColorsAny) > 0
	signals := 0
	for _, present := range []bool{hasColors, sel.Colorless, sel.Multicolored} {
		if present {
			signals++
		}
	}
	if signals == 0 {
		return "", false, true
	}
	if signals > 1 {
		return "", false, false
	}
	switch {
	case hasColors:
		if len(sel.ColorsAny) != 1 {
			return "", false, false
		}
		word, ok := colorWord(sel.ColorsAny[0])
		if !ok {
			return "", false, false
		}
		return word + " ", true, true
	case sel.Colorless:
		return "colorless ", true, true
	default:
		return "multicolored ", true, true
	}
}

// graveyardCardTypeNoun reconstructs the card-type noun ("creature card",
// "sorcery card", "creature or enchantment card"). A permanent single type is
// carried by the selection Kind, so lowering's Kind-to-type mapping reproduces
// it. A single instant or sorcery type has no permanent selection Kind; it
// arrives as a generic SelectionCard whose RequiredTypesAny retains the exact
// type, which lowering reproduces as a type-restricted card target. A union of
// two or more types is carried explicitly by the compiler, so each member is
// rendered from its card-type word and joined with " or " for a singular target
// or " and/or " for a plural multi-target return.
func graveyardCardTypeNoun(sel SelectionSyntax, plural bool) (string, bool) {
	if len(sel.RequiredTypesAny) == 1 {
		word, ok := cardTypeWord(sel.RequiredTypesAny[0])
		if !ok {
			return "", false
		}
		noun, ok := permanentSelectionNoun(sel.Kind)
		if ok {
			if word != noun {
				return "", false
			}
			return noun + " card", true
		}
		// A non-permanent single card type (instant, sorcery) has no permanent
		// selection Kind and arrives as a generic SelectionCard. Its type is
		// retained in RequiredTypesAny, so lowering restricts the card target to
		// that type; render the noun from the card-type word directly.
		if sel.Kind == SelectionCard {
			return word + " card", true
		}
		return "", false
	}
	words := make([]string, 0, len(sel.RequiredTypesAny))
	for _, cardType := range sel.RequiredTypesAny {
		word, ok := cardTypeWord(cardType)
		if !ok {
			return "", false
		}
		words = append(words, word)
	}
	conjunction := "or"
	if plural {
		conjunction = "and/or"
	}
	return serialList(words, conjunction) + " card", true
}

// graveyardSubtypeAdjective renders the one-or-more subtype adjectives a
// graveyard-card noun may carry ("Zombie", "Aura or Equipment", "Bat, Lizard,
// Rat, or Squirrel"). A single subtype renders its bare name; a disjunction of
// two or more joins each name with the canonical Oracle conjunction (" or " for
// a singular target, " and/or " for a plural multi-target count) and the
// serial comma for three or more. The matched card carries any one of the named
// subtypes, mirroring the runtime Selection.SubtypesAny union the lowering
// builds, so the disjunction reads "or" rather than the conjunctive "and".
func graveyardSubtypeAdjective(sel SelectionSyntax, plural bool) string {
	words := make([]string, 0, len(sel.SubtypesAny))
	for _, subtype := range sel.SubtypesAny {
		words = append(words, string(subtype))
	}
	conjunction := "or"
	if plural {
		conjunction = "and/or"
	}
	return serialList(words, conjunction)
}

// words join as "A <conj> B", and three or more use the serial-comma form
// "A, B, <conj> C". The conjunction is "or" for a singular target or "and/or"
// for a plural multi-target count.
func serialList(words []string, conjunction string) string {
	switch len(words) {
	case 0:
		return ""
	case 1:
		return words[0]
	case 2:
		return words[0] + " " + conjunction + " " + words[1]
	default:
		head := strings.Join(words[:len(words)-1], ", ")
		return head + ", " + conjunction + " " + words[len(words)-1]
	}
}

// graveyardNumericQualifier renders the single optional numeric qualifier a
// graveyard-card selection may carry after its noun: the fixed " with mana value
// N or less", the dynamic life-total mana-value bound (Betor, Ancestor's Voice),
// or a fixed " with power/toughness N or less/greater" characteristic bound
// (Dusk // Dawn's "creature cards with power 2 or less", Reveillark). Printed
// cards carry at most one such qualifier, so it fails closed when more than one
// is set rather than guessing their order, and fails closed for any comparison
// operator the canonical Oracle wording does not use. The empty string is
// returned (ok=true) when the selection carries no numeric qualifier.
func graveyardNumericQualifier(sel SelectionSyntax) (string, bool) {
	clauses := 0
	qualifier := ""
	if sel.MatchManaValue {
		clause, ok := graveyardManaValueClause(sel.ManaValue)
		if !ok {
			return "", false
		}
		qualifier = clause
		clauses++
	}
	if sel.ManaValueDynamic != "" {
		clause, ok := graveyardManaValueDynamicClause(sel.ManaValueDynamic)
		if !ok {
			return "", false
		}
		qualifier = clause
		clauses++
	}
	if sel.MatchPower {
		clause, ok := graveyardCharacteristicClause("power", sel.Power)
		if !ok {
			return "", false
		}
		qualifier = clause
		clauses++
	}
	if sel.MatchToughness {
		clause, ok := graveyardCharacteristicClause("toughness", sel.Toughness)
		if !ok {
			return "", false
		}
		qualifier = clause
		clauses++
	}
	if clauses > 1 {
		return "", false
	}
	return qualifier, true
}

// graveyardCharacteristicClause renders the canonical " with power N or less",
// " with power N or greater", or the toughness equivalents from a fixed
// power/toughness comparison. It accepts only the "or less" and "or greater"
// bounds the printed Oracle wording uses, failing closed for any other operator.
func graveyardCharacteristicClause(characteristic string, bound compare.Int) (string, bool) {
	switch bound.Op {
	case compare.LessOrEqual:
		return " with " + characteristic + " " + strconv.Itoa(bound.Value) + " or less", true
	case compare.GreaterOrEqual:
		return " with " + characteristic + " " + strconv.Itoa(bound.Value) + " or greater", true
	default:
		return "", false
	}
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

// graveyardManaValueDynamicClause renders the canonical " with mana value less
// than or equal to the amount of life you (lost|gained) this turn" qualifier from
// a dynamic mana-value bound (Betor, Ancestor's Voice). It fails closed for any
// other dynamic amount.
func graveyardManaValueDynamicClause(kind EffectDynamicAmountKind) (string, bool) {
	switch kind {
	case EffectDynamicAmountLifeLostThisTurn:
		return " with mana value less than or equal to the amount of life you lost this turn", true
	case EffectDynamicAmountLifeGainedThisTurn:
		return " with mana value less than or equal to the amount of life you gained this turn", true
	default:
		return "", false
	}
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

// exactBecomeMonarchEffectSyntax recognizes the monarch-designation effect
// (CR 720) in its controller form "You become the monarch." and its single
// player-target form "Target player becomes the monarch." / "Target opponent
// becomes the monarch.". Any other subject leaves the clause non-exact so
// lowering fails closed.
func exactBecomeMonarchEffectSyntax(effect *EffectSyntax) bool {
	text := exactEffectClauseText(effect)
	if len(effect.Targets) == 0 {
		return strings.EqualFold(text, "You become the monarch.")
	}
	return len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		strings.EqualFold(text, effect.Targets[0].Text+" becomes the monarch.")
}

// exactMultiDistinctTargetEffectSyntax recognizes a verb applied to two or more
// distinct single targets, each named by its own "target <noun>" clause:
// "Destroy target artifact, target creature, target enchantment, and target
// land." (Decimate), "Destroy target artifact and target creature." Every target
// is exact with single cardinality, and the canonical list joins the target
// texts in an Oracle serial series — "A and B" for two, "A, B, and C" for three
// or more with the serial comma. Any other shape (a plural or optional target, a
// non-exact target, or trailing clause text) leaves the reconstruction
// mismatched and the clause non-exact, so lowering fails closed.
func exactMultiDistinctTargetEffectSyntax(effect *EffectSyntax, verb string) bool {
	if len(effect.Targets) < 2 {
		return false
	}
	single := TargetCardinalitySyntax{Min: 1, Max: 1}
	texts := make([]string, 0, len(effect.Targets))
	for i := range effect.Targets {
		target := effect.Targets[i]
		if !target.Exact || target.Cardinality != single {
			return false
		}
		texts = append(texts, target.Text)
	}
	return strings.EqualFold(exactEffectClauseText(effect), verb+" "+joinSerialTargetTexts(texts)+".")
}

// joinSerialTargetTexts joins distinct target clause texts the way Oracle text
// lists a verb's several targets: a single text as-is, two joined by "and", and
// three or more in a serial-comma series ("target artifact, target creature, and
// target land").
func joinSerialTargetTexts(texts []string) string {
	switch len(texts) {
	case 0:
		return ""
	case 1:
		return texts[0]
	case 2:
		return texts[0] + " and " + texts[1]
	default:
		return strings.Join(texts[:len(texts)-1], ", ") + ", and " + texts[len(texts)-1]
	}
}

// exactRemoveFromCombatEffectSyntax recognizes the resolving effect "Remove
// <target> from combat." (Reconnaissance, "Remove target attacking creature you
// control from combat."), whose single permanent target is the creature taken
// out of combat. The "from combat" clause is the verb's fixed suffix; any other
// wording leaves the clause non-exact so lowering fails closed.
func exactRemoveFromCombatEffectSyntax(effect *EffectSyntax) bool {
	return len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		strings.EqualFold(exactEffectClauseText(effect), "Remove "+effect.Targets[0].Text+" from combat.")
}

// exactRegenerateSelfEffectSyntax recognizes the self-regeneration form
// "Regenerate this creature." (and the "this permanent"/"this token" object
// nouns) or "Regenerate <CardName>." where the regenerated permanent is the
// ability's own source. The single reference is the source self-reference
// ("this <object>" or the card's own name), reconstructed exactly with no
// target; any other shape leaves the clause non-exact so lowering fails closed.
func exactRegenerateSelfEffectSyntax(effect *EffectSyntax) bool {
	if len(effect.Targets) != 0 || len(effect.References) != 1 {
		return false
	}
	reference := effect.References[0]
	if reference.Kind != ReferenceThisObject && reference.Kind != ReferenceSelfName {
		return false
	}
	return strings.EqualFold(exactEffectClauseText(effect), "Regenerate "+reference.Text+".")
}

// exactRegenerateAttachedEffectSyntax recognizes the attached-recipient form
// "Regenerate enchanted creature." (Aura) or "Regenerate equipped creature."
// (Equipment), where the regenerated permanent is the one the source is attached
// to. There is no target or reference; lowering routes it to the runtime's
// source attached-permanent reference. Any other wording leaves the clause
// non-exact so lowering fails closed.
func exactRegenerateAttachedEffectSyntax(effect *EffectSyntax) bool {
	if len(effect.Targets) != 0 || len(effect.References) != 0 {
		return false
	}
	text := exactEffectClauseText(effect)
	return strings.EqualFold(text, "Regenerate enchanted creature.") ||
		strings.EqualFold(text, "Regenerate equipped creature.")
}

// exactExileAttachedEffectSyntax recognizes the attached-recipient form "Exile
// enchanted creature." (Aura) or "Exile equipped creature." (Equipment), where
// the exiled permanent is the one the source is attached to. There is no target
// or reference; lowering routes it to the runtime's source attached-permanent
// reference. Any other wording leaves the clause non-exact so lowering fails
// closed.
func exactExileAttachedEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectExile || effect.Negated {
		return false
	}
	if len(effect.Targets) != 0 || len(effect.References) != 0 {
		return false
	}
	if effect.Duration != EffectDurationNone || effect.FromZone != zone.None || effect.ToZone != zone.None {
		return false
	}
	text := exactEffectClauseText(effect)
	return strings.EqualFold(text, "Exile enchanted creature.") ||
		strings.EqualFold(text, "Exile equipped creature.")
}

// exactCopyStackObjectEffectSyntax recognizes the resolving effect "Copy <target
// activated or triggered ability you control>." The single stack-object target
// is the ability to copy. The optional "You may choose new targets for the
// copy[ies]." rider is a separate folded sentence (CopyMayChooseNewTargets), so
// the effect's own clause text never carries it; any other trailing clause
// leaves the text non-exact and fails closed.
func exactCopyStackObjectEffectSyntax(effect *EffectSyntax) bool {
	return exactDirectTargetEffectSyntax(effect, "Copy") ||
		exactCopyReferencedSpellEffectSyntax(effect)
}

// exactCopyReferencedSpellEffectSyntax recognizes the resolving effect "Copy
// that spell." / "Copy it." / "Copy this spell.", whose copy source is a
// back-reference to the triggering spell ("Whenever you cast a spell ..., copy
// that spell.", Reflections of Littjara) or to the resolving spell itself
// ("Copy this spell.", Sevinne's Reclamation). It requires no targets and a
// single "that spell"/"it"/"this spell" reference; the compiler binds that
// reference to the triggering or resolving spell and lowering copies it. The
// optional "You may choose new targets for the copy." rider folds separately
// once this clause is exact.
func exactCopyReferencedSpellEffectSyntax(effect *EffectSyntax) bool {
	if len(effect.Targets) != 0 {
		return false
	}
	references := effectClauseReferences(effect)
	if len(references) != 1 {
		return false
	}
	reference := references[0]
	switch reference.Kind {
	case ReferenceThatObject, ReferenceThisObject:
	case ReferencePronoun:
		if reference.Pronoun != PronounIt {
			return false
		}
	default:
		return false
	}
	return strings.EqualFold(exactEffectClauseText(effect), "Copy "+reference.Text+".")
}

// effectClauseReferences returns the effect's references that fall at or after
// its verb, dropping references that belong to a leading condition clause ("If
// this spell was cast from a graveyard, ...") which the parser records on the
// gated effect. Exact recognizers that constrain reference count read the
// clause's own references through this helper so a condition's back-reference
// does not defeat the match.
func effectClauseReferences(effect *EffectSyntax) []Reference {
	var clause []Reference
	for _, reference := range effect.References {
		if reference.Span.Start.Offset >= effect.VerbSpan.Start.Offset {
			clause = append(clause, reference)
		}
	}
	return clause
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

// exactTargetNextUntapStepSyntax recognizes the standalone targeted stun spell or
// ability "Target <permanent> doesn't untap during its controller's next untap
// step." (Sleeper Dart, House Guildmage, Skyline Cascade), where the stunned
// permanent is the effect's own single target rather than a just-tapped prior
// subject. The clause carries one target and one possessive "its" reference (the
// "its controller's" owner of the next untap step) and no duration; only the
// single "next untap step" window is exact, so every plural, mass, or multi-step
// wording leaves the clause non-exact and lowering fails closed.
func exactTargetNextUntapStepSyntax(effect *EffectSyntax) bool {
	if !effect.Negated || effect.Optional ||
		effect.Context != EffectContextTarget ||
		len(effect.Targets) != 1 || len(effect.References) != 1 ||
		effect.Duration != EffectDurationNone || effect.DelayedTiming != DelayedTimingNone {
		return false
	}
	possessive := effect.References[0]
	if possessive.Kind != ReferencePronoun || possessive.Pronoun != PronounIts {
		return false
	}
	words := normalizedWords(effect.Tokens)
	verb := slices.Index(words, "untap")
	if verb < 1 || len(words) == 0 || words[0] != "target" || words[verb-1] != "doesn't" {
		return false
	}
	return slices.Equal(words[verb+1:], []string{"during", "its", "controller's", "next", "untap", "step"})
}

// exactSourceNextUntapStepSyntax recognizes the standalone self-source stun
// clause "This <permanent> doesn't untap during your next untap step." in which
// the stunned permanent is the source itself (the dual lands Mogg Hollows /
// Rootwater Depths and Arbalest Elite append it to a mana or damage ability so
// the source skips its own next untap). The clause carries a single
// "This <permanent>" self reference and no target, possessive controller
// reference, or duration; only the single "your next untap step" window is
// exact, so every plural, mass, or multi-step wording leaves the clause
// non-exact and lowering fails closed.
func exactSourceNextUntapStepSyntax(effect *EffectSyntax) bool {
	if !effect.Negated || effect.Optional ||
		effect.Context != EffectContextSource ||
		len(effect.Targets) != 0 || len(effect.References) != 1 ||
		effect.Duration != EffectDurationNone || effect.DelayedTiming != DelayedTimingNone {
		return false
	}
	if effect.References[0].Kind != ReferenceThisObject {
		return false
	}
	words := normalizedWords(effect.Tokens)
	verb := slices.Index(words, "untap")
	if verb < 1 || words[verb-1] != "doesn't" {
		return false
	}
	return slices.Equal(words[verb+1:], []string{"during", "your", "next", "untap", "step"})
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

// The bounce-to-hand destination possessive is the only part of a return clause
// shared verbatim across every battlefield bounce scope. Naming the three
// possessive renderings once lets the single-, multi-, dual-, controlled-choice,
// self-, and mass-group exactness branches reconstruct the same destination
// instead of each spelling out the literal, keeping the typed destination in one
// place.
const (
	bounceHandDestSingular   = "to its owner's hand."
	bounceHandDestPlural     = "to their owners' hands."
	bounceHandDestTheirOwner = "to their owner's hand."
)

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
	return strings.EqualFold(exactEffectClauseText(effect), "Return "+phrase+" "+bounceHandDestSingular)
}

func exactBounceEffectSyntax(effect *EffectSyntax) bool {
	return len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		strings.EqualFold(exactEffectClauseText(effect), "Return "+effect.Targets[0].Text+" "+bounceHandDestSingular)
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
	reconstruction := "Return " + effect.Targets[0].Text + " and " + effect.Targets[1].Text + " " + bounceHandDestPlural
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
		strings.EqualFold(exactEffectClauseText(effect), "Return "+effect.Targets[0].Text+" "+bounceHandDestPlural)
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
	return strings.EqualFold(exactEffectClauseText(effect), "Return "+subject+" "+bounceHandDestSingular)
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

// exactBackReferenceEffectSyntax recognizes a removal verb acting on a
// demonstrative back-reference object — "Exile that creature.", "Destroy it." —
// where the object was introduced by a preceding clause or the triggering event.
// Unlike exactDirectReferenceEffectSyntax it rejects self-references ("this
// creature" / the card's own name), which name the source permanent itself and
// are handled by the dedicated source paths; admitting them here would mislabel
// nonsensical spell bodies such as "Exile this creature." as exact.
func exactBackReferenceEffectSyntax(effect *EffectSyntax, verb string) bool {
	if len(effect.Targets) != 0 || effect.Optional || effect.Duration != EffectDurationNone {
		return false
	}
	object, ok := exactBackReferenceObjectText(effect.References)
	return ok && strings.EqualFold(exactEffectClauseText(effect), verb+" "+object+".")
}

func exactBackReferenceObjectText(references []Reference) (string, bool) {
	if len(references) != 1 {
		return "", false
	}
	switch references[0].Kind {
	case ReferenceThatObject:
	case ReferencePronoun:
		if references[0].Pronoun != PronounIt {
			return "", false
		}
	default:
		return "", false
	}
	return joinedEffectText(references[0].Tokens), true
}

// referencesOutsideSpan returns the references whose source span falls outside
// the given span. A zero span (no dynamic amount) excludes nothing, so the full
// reference list is returned unchanged.
func referencesOutsideSpan(references []Reference, span shared.Span) []Reference {
	if span == (shared.Span{}) {
		return references
	}
	var result []Reference
	for _, reference := range references {
		if spanCovers(span, reference.Span) {
			continue
		}
		result = append(result, reference)
	}
	return result
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
	// A mass effect may carry an explicit "You" controller actor when it is one
	// clause of a sequence ("..., and you untap all lands you control"). The
	// canonical mass phrase has no actor, so strip a leading "You " before the
	// prefix check when the effect's subject is its controller.
	if effect.Context == EffectContextController {
		if len(text) > 4 && strings.EqualFold(text[:4], "you ") {
			text = text[4:]
		}
	}
	if !strings.HasPrefix(strings.ToLower(text), strings.ToLower(prefix)) || !strings.HasSuffix(text, ".") {
		return false
	}
	phrase := text[len(prefix) : len(text)-1]
	return exactMassPluralGroupPhrase(&effect.Selection, phrase)
}

// exactMassPluralGroupPhrase validates a plural "all <group>" mass phrase against
// the typed selection, accepting every group shape the mass-effect machinery
// models: the bare group noun, a single qualifying or excluded creature subtype,
// and the chosen-type or counter qualifiers stripped to a base group first. It is
// the shared validator for every plural mass form (destroy, exile, regenerate,
// tap, untap, and the mass return below) so each verb recognizes the same set of
// group wordings rather than maintaining its own narrower subset.
func exactMassPluralGroupPhrase(selection *SelectionSyntax, phrase string) bool {
	if base, ok := massChosenTypeBasePhrase(selection, phrase); ok {
		return exactMassGroupPhrase(selection, base)
	}
	if base, ok := massCounterBasePhrase(selection, phrase); ok {
		return exactMassGroupPhrase(selection, base) || exactMassSubtypePhrase(selection, base) || exactMassExcludedSubtypePhrase(selection, base)
	}
	return exactMassGroupPhrase(selection, phrase) || exactMassSubtypePhrase(selection, phrase) || exactMassExcludedSubtypePhrase(selection, phrase)
}

// massChosenTypeBasePhrase strips a trailing chosen-type qualifier ("of the
// chosen type" / "that aren't of the chosen type") from a mass group phrase when
// the selection records the matching chosen-type field, returning the base group
// phrase to validate and true. The base ("creatures") is then checked by the
// shared exactMassGroupPhrase, so "Destroy all creatures that aren't of the
// chosen type." (Kindred Dominance) round-trips through the same machinery as the
// bare mass group. It fails closed when neither chosen-type field is set or the
// phrase lacks the expected suffix.
func massChosenTypeBasePhrase(selection *SelectionSyntax, phrase string) (string, bool) {
	switch {
	case selection.SubtypeFromChosenTypeExcluded:
		if base, ok := strings.CutSuffix(phrase, " that aren't of the chosen type"); ok {
			return base, true
		}
	case selection.SubtypeFromChosenType:
		if base, ok := strings.CutSuffix(phrase, " of the chosen type"); ok {
			return base, true
		}
	default:
	}
	return "", false
}

// massCounterBasePhrase strips a trailing counter qualifier ("with a +1/+1
// counter on it" / "with a -1/-1 counter on them", or the negated "with no
// counters on them") from a mass group phrase when the selection records the
// matching counter requirement, returning the base group phrase to validate and
// true. The base ("creatures") is then checked by the shared mass group/subtype
// validators, so "Destroy all creatures with a +1/+1 counter on them." and
// "Destroy all creatures with no counters on them." round-trip through the same
// machinery as the bare mass group. Because stripping is driven by the modeled
// CounterKind/CounterAbsent (not by text), an unmodeled named counter leaves
// CounterRequired false and the phrase fails closed. The kind-agnostic "any
// counter" form is intentionally not accepted: the runtime cannot honor it for a
// mass group (it would require the zero-value counter kind in addition to any
// counter), so it stays fail closed.
func massCounterBasePhrase(selection *SelectionSyntax, phrase string) (string, bool) {
	if selection.CounterKindAbsent {
		kind := selection.CounterKind.String()
		for _, article := range []string{"a", "an"} {
			for _, pronoun := range []string{"it", "them"} {
				suffix := " without " + article + " " + kind + " counter on " + pronoun
				if base, ok := strings.CutSuffix(phrase, suffix); ok {
					return base, true
				}
			}
		}
		return "", false
	}
	if selection.CounterAbsent {
		for _, suffix := range []string{
			" with no counters on it", " with no counters on them",
			" with no counter on it", " with no counter on them",
		} {
			if base, ok := strings.CutSuffix(phrase, suffix); ok {
				return base, true
			}
		}
		return "", false
	}
	if !selection.CounterRequired || selection.CounterAny {
		return "", false
	}
	for _, suffix := range massCounterQualifierSuffixes(selection) {
		if base, ok := strings.CutSuffix(phrase, suffix); ok {
			return base, true
		}
	}
	return "", false
}

// massCounterQualifierSuffixes reconstructs the recognized counter-qualifier
// suffixes for a named-counter selection from its modeled counter kind, covering
// both the singular ("on it") and plural ("on them") pronoun and both articles
// ("a"/"an") so the reconstructed text matches the source.
func massCounterQualifierSuffixes(selection *SelectionSyntax) []string {
	pronouns := []string{"it", "them"}
	kind := selection.CounterKind.String()
	suffixes := make([]string, 0, 2*len(pronouns))
	for _, article := range []string{"a", "an"} {
		for _, pronoun := range pronouns {
			suffixes = append(suffixes, " with "+article+" "+kind+" counter on "+pronoun)
		}
	}
	return suffixes
}

// massEachGroupVerbEffectSyntax reports whether the effect is a recognized
// "<verb> each <group>" mass form for one of the group verbs that share the
// battlefield-group machinery (destroy, exile, tap, untap, regenerate). The
// singular "each" wording selects the whole matching group exactly like the
// plural "all" form, so its caller flags the selection All to lower to a group
// effect. Every other effect kind fails closed so per-player "each" distributive
// effects and "each creature" damage recipients keep their own handling.
func massEachGroupVerbEffectSyntax(effect *EffectSyntax) bool {
	switch effect.Kind {
	case EffectDestroy:
		return exactMassEachEffectSyntax(effect, "Destroy each ")
	case EffectExile:
		return exactMassEachEffectSyntax(effect, "Exile each ")
	case EffectTap:
		return exactMassEachEffectSyntax(effect, "Tap each ")
	case EffectUntap:
		return exactMassEachEffectSyntax(effect, "Untap each ")
	case EffectRegenerate:
		return exactMassEachEffectSyntax(effect, "Regenerate each ")
	default:
		return false
	}
}

// exactMassEachEffectSyntax recognizes the singular "each" mass form
// ("Destroy each nonland permanent with mana value 2 or less.") that selects
// every matching permanent just like the plural "all" form ("Destroy all
// nonland permanents ..."). The "each" wording names a single permanent type,
// so its group phrase is validated in the singular by exactMassEachGroupPhrase
// while reusing the shared numeric-comparison clause. It fails closed for every
// other wording so single-target and player-distributive forms are untouched.
func exactMassEachEffectSyntax(effect *EffectSyntax, prefix string) bool {
	text := exactEffectClauseText(effect)
	if effect.Context == EffectContextController {
		if len(text) > 4 && strings.EqualFold(text[:4], "you ") {
			text = text[4:]
		}
	}
	if !strings.HasPrefix(strings.ToLower(text), strings.ToLower(prefix)) || !strings.HasSuffix(text, ".") {
		return false
	}
	phrase := text[len(prefix) : len(text)-1]
	if base, ok := massCounterBasePhrase(&effect.Selection, phrase); ok {
		return exactMassEachGroupPhrase(&effect.Selection, base)
	}
	return exactMassEachGroupPhrase(&effect.Selection, phrase)
}

// exactMassEachGroupPhrase validates the singular "each" mass group phrase both
// as text shape and against the typed selection: massEachGroupPhraseTextShape
// recognizes the canonical wording, and selectionPhraseVerifiesMassGroup proves
// the typed SelectionSyntax renders to that same singular phrase (closing the
// soundness gap where text shape alone could accept a divergent selection).
func exactMassEachGroupPhrase(selection *SelectionSyntax, phrase string) bool {
	if !massEachGroupPhraseTextShape(phrase) {
		return false
	}
	return selectionPhraseVerifiesMassGroup(selection, phrase, numberSingular)
}

// selectionPhraseVerifiesMassGroup confirms the typed selection renders to the
// validated mass group phrase through the canonical selectionPhrase renderer.
// When selectionPhrase reports it cannot represent the selection's noun-phrase
// shape (ok=false) — for the subtype-noun and keyword group forms still owned by
// their dedicated validators — it returns true so callers fall back to the
// text-shape recognizer; otherwise the rendered phrase must match the source.
func selectionPhraseVerifiesMassGroup(selection *SelectionSyntax, phrase string, number grammaticalNumber) bool {
	rendered, ok := selectionPhrase(*selection, selectionPhraseOptions{Number: number})
	if !ok {
		return true
	}
	return strings.EqualFold(rendered, phrase)
}

// massEachGroupPhraseTextShape validates the singular group phrase that follows
// "Destroy each ". It mirrors massGroupPhraseTextShape's excluded-type/color
// prefixes, base nouns, and numeric comparison clauses, but in the singular
// ("nonland permanent with mana value 2 or less" rather than the plural
// "nonland permanents ..."), so an "each" mass clause round-trips to the same
// group selection the plural form lowers.
func massEachGroupPhraseTextShape(phrase string) bool {
	if phrase == "" || strings.TrimSpace(phrase) != phrase {
		return false
	}
	phrase = strings.ToLower(phrase)
	for _, suffix := range []string{" you don't control", " your opponents control", " you control"} {
		if remainder, ok := strings.CutSuffix(phrase, suffix); ok {
			phrase = remainder
			break
		}
	}
	if exactMassEachNumericPhrase(phrase) {
		return true
	}
	if exactMassEachBaseNoun(phrase) {
		return true
	}
	for _, prefix := range []string{
		"other ", "tapped ", "untapped ", "nonland ", "nonartifact ", "noncreature ", "nonenchantment ",
		"white ", "blue ", "black ", "red ", "green ", "nonwhite ", "nonblue ", "nonblack ", "nonred ", "nongreen ",
		"attacking ", "blocking ", "attacking or blocking ",
	} {
		if remainder, ok := strings.CutPrefix(phrase, prefix); ok {
			return exactMassEachBaseNoun(remainder)
		}
	}
	if remainder, ok := strings.CutPrefix(phrase, "nonbasic "); ok {
		return remainder == "land"
	}
	return false
}

func exactMassEachBaseNoun(phrase string) bool {
	switch phrase {
	case "creature", "artifact", "enchantment", "land", "planeswalker", "permanent":
		return true
	default:
		return false
	}
}

// exactMassEachNumericPhrase recognizes a singular "each" mass group restricted
// by a numeric "with mana value"/"with power"/"with toughness" comparison,
// optionally behind a single excluded-type prefix. It is the singular sibling of
// exactMassNumericPhrase.
func exactMassEachNumericPhrase(phrase string) bool {
	for _, exPrefix := range []string{"", "nonland ", "nonartifact ", "noncreature ", "nonenchantment "} {
		rest, ok := strings.CutPrefix(phrase, exPrefix)
		if !ok {
			continue
		}
		for _, noun := range []string{"creature", "artifact", "enchantment", "land", "planeswalker", "permanent"} {
			comparison, ok := strings.CutPrefix(rest, noun+" with ")
			if !ok {
				continue
			}
			qualifiers := []string{"mana value"}
			if exPrefix == "" && noun == "creature" {
				qualifiers = []string{"mana value", "power", "toughness"}
			}
			if exactMassComparisonClause(comparison, qualifiers) {
				return true
			}
		}
	}
	return false
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

// exactMassExcludedSubtypePhrase reconstructs the canonical mass phrase for an
// excluded-subtype group ("non-Dragon creatures", "non-Gorgon creatures") from
// the parsed selection and compares it byte-exactly to the source phrase. It
// mirrors exactMassSubtypePhrase but negates a single creature subtype: the
// group restricts to a permanent card type ("creatures") and excludes one
// subtype, rendered as "non-<subtype> <noun>s". It accepts exactly one excluded
// subtype with a required permanent card-type noun and no other qualifier,
// failing closed for every other wording so unsupported mass forms keep failing
// the round-trip.
func exactMassExcludedSubtypePhrase(selection *SelectionSyntax, phrase string) bool {
	if len(selection.ExcludedSubtypes) != 1 ||
		len(selection.SubtypesAny) != 0 ||
		selection.Controller != SelectionControllerAny ||
		selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking ||
		selection.Tapped || selection.Untapped ||
		selection.NonToken || selection.TokenOnly ||
		selection.Keyword != KeywordUnknown || selection.ExcludedKeyword != KeywordUnknown ||
		selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.Colorless || selection.Multicolored ||
		!selectionRedundantRequiredNoun(*selection) || len(selection.ExcludedTypes) != 0 ||
		len(selection.Supertypes) != 0 || len(selection.ExcludedSupertypes) != 0 ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedColors) != 0 {
		return false
	}
	noun, ok := permanentSelectionNoun(selection.Kind)
	if !ok {
		return false
	}
	excluded := strings.ToLower(string(selection.ExcludedSubtypes[0]))
	return strings.EqualFold(phrase, "non-"+excluded+" "+noun+"s")
}

// exactMassBounceEffectSyntax recognizes the mass battlefield return
// "Return all <group> to their owners' hands." (and the "you control" variant
// "Return all <group> you control to their owner's hand.") that lowers to a
// single group Bounce, mirroring the mass destroy/exile group syntax. The return
// wording differs from destroy/exile only by its "to their owners' hands"
// destination suffix; that possessive is reconstructed canonically here so the
// group phrase between "Return all " and the suffix can be validated by the
// shared exactMassPluralGroupPhrase, which recognizes the same group, subtype,
// excluded-subtype, counter, and chosen-type wordings as the mass destroy/exile
// forms. It fails closed for every other return wording so the single- and
// multi-target bounce paths are untouched.
func exactMassBounceEffectSyntax(effect *EffectSyntax) bool {
	if effect.ToZone != zone.Hand {
		return false
	}
	const prefix = "Return all "
	text := exactEffectClauseText(effect)
	if !strings.HasPrefix(strings.ToLower(text), strings.ToLower(prefix)) {
		return false
	}
	for _, suffix := range []string{" " + bounceHandDestPlural, " " + bounceHandDestTheirOwner} {
		if remainder, ok := strings.CutSuffix(text, suffix); ok {
			return exactMassPluralGroupPhrase(&effect.Selection, remainder[len(prefix):])
		}
	}
	return false
}

// exactMassEachBounceEffectSyntax recognizes the singular "each" mass return
// "Return each <group> to its owner's hand." (Wave Goodbye's "Return each
// creature without a +1/+1 counter on it to its owner's hand."). It is the
// "each" sibling of exactMassBounceEffectSyntax: the singular wording selects
// every matching permanent just like the plural "all" form, so it validates the
// group phrase in the singular through exactMassEachGroupPhrase while stripping a
// recognized counter qualifier first. It fails closed for every other return
// wording so the single- and multi-target bounce paths are untouched.
func exactMassEachBounceEffectSyntax(effect *EffectSyntax) bool {
	if effect.ToZone != zone.Hand {
		return false
	}
	const prefix = "Return each "
	text := exactEffectClauseText(effect)
	if !strings.HasPrefix(strings.ToLower(text), strings.ToLower(prefix)) {
		return false
	}
	for _, suffix := range []string{" " + bounceHandDestSingular, " " + bounceHandDestTheirOwner, " " + bounceHandDestPlural} {
		remainder, ok := strings.CutSuffix(text, suffix)
		if !ok {
			continue
		}
		phrase := remainder[len(prefix):]
		if base, ok := massCounterBasePhrase(&effect.Selection, phrase); ok {
			return exactMassEachGroupPhrase(&effect.Selection, base)
		}
		return exactMassEachGroupPhrase(&effect.Selection, phrase)
	}
	return false
}

// exactMassGroupPhrase validates the plural "all" mass group phrase both as text
// shape and against the typed selection: massGroupPhraseTextShape recognizes the
// canonical wording, and selectionPhraseVerifiesMassGroup proves the typed
// SelectionSyntax renders to that same plural phrase (closing the soundness gap
// where text shape alone could accept a divergent selection).
func exactMassGroupPhrase(selection *SelectionSyntax, phrase string) bool {
	if !massGroupPhraseTextShape(phrase) {
		return false
	}
	return selectionPhraseVerifiesMassGroup(selection, phrase, numberPlural)
}

func massGroupPhraseTextShape(phrase string) bool {
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
	if remainder, ok := massGroupStripAdjectivePrefixes(phrase); ok {
		if exactMassBaseNoun(remainder) {
			return true
		}
	}
	// "nonbasic" is a supertype exclusion meaningful only for lands ("Destroy all
	// nonbasic lands."); every other base noun fails closed.
	if remainder, ok := strings.CutPrefix(phrase, "nonbasic "); ok {
		return remainder == "lands"
	}
	return false
}

// massGroupStripAdjectivePrefixes strips the canonical adjective prefixes that
// precede a mass group base noun, in the order selectionPhrase renders them: a
// leading "other", a combat/tapped state, a single color or excluded color, and
// one or more comma-joined excluded card types ("noncreature, nonland"). It
// strips each recognized prefix once (or, for excluded types, repeatedly) so
// compound wordings such as "other tapped creatures", "other nonland permanents",
// and "noncreature, nonland permanents" reduce to their bare base noun. The typed
// selection verification in exactMassGroupPhrase confirms the stripped phrase
// still describes the source selection, so widening the recognized text shape
// here never accepts a phrase the typed selection contradicts. It reports the
// remaining base-noun phrase and whether any prefix was stripped.
func massGroupStripAdjectivePrefixes(phrase string) (string, bool) {
	stripped := false
	if remainder, ok := strings.CutPrefix(phrase, "other "); ok {
		phrase = remainder
		stripped = true
	}
	for _, prefix := range []string{
		"tapped ", "untapped ", "attacking or blocking ", "attacking ", "blocking ",
	} {
		if remainder, ok := strings.CutPrefix(phrase, prefix); ok {
			phrase = remainder
			stripped = true
			break
		}
	}
	for _, prefix := range []string{
		"white ", "blue ", "black ", "red ", "green ",
		"nonwhite ", "nonblue ", "nonblack ", "nonred ", "nongreen ",
	} {
		if remainder, ok := strings.CutPrefix(phrase, prefix); ok {
			phrase = remainder
			stripped = true
			break
		}
	}
	for {
		remainder, ok := massGroupStripExcludedTypePrefix(phrase)
		if !ok {
			break
		}
		phrase = remainder
		stripped = true
	}
	return phrase, stripped
}

// massGroupStripExcludedTypePrefix strips one leading "non<type>" excluded card
// type prefix from a mass group phrase, consuming a comma separator when more
// excluded types follow ("noncreature, nonland ...") and a plain space when the
// excluded type is the last prefix before the base noun ("nonland permanents").
// It reports the remaining phrase and whether a prefix was stripped.
func massGroupStripExcludedTypePrefix(phrase string) (string, bool) {
	for _, excluded := range []string{"nonland", "nonartifact", "noncreature", "nonenchantment"} {
		if remainder, ok := strings.CutPrefix(phrase, excluded+", "); ok {
			return remainder, true
		}
		if remainder, ok := strings.CutPrefix(phrase, excluded+" "); ok {
			return remainder, true
		}
	}
	return phrase, false
}

// exactSacrificeMassEffectSyntax recognizes the mass sacrifice wording
// "<player> sacrifices all <group> [they control] that are one or more colors."
// (All Is Dust), the only sacrifice form that removes every matching permanent a
// player controls rather than a chosen amount. It requires the typed selection's
// All and Colored flags so it fails closed for every bounded "sacrifices N
// <group>" wording, which exactSacrificeChoiceEffectSyntax continues to own. The
// "one or more colors" suffix is reconstructed canonically and is mandatory so a
// Colored selection without that exact text fails closed; the optional "they
// control" suffix is stripped because the per-player sacrifice already scopes
// each player to permanents they control. The remaining bare mass group phrase
// is validated by the shared exactMassGroupPhrase.
func exactSacrificeMassEffectSyntax(effect *EffectSyntax) bool {
	if !effect.Selection.All || !effect.Selection.Colored || len(effect.Targets) != 0 {
		return false
	}
	subject, ok := massSacrificeSubject(effect.Context)
	if !ok {
		return false
	}
	phrase := strings.ToLower(exactEffectClauseText(effect))
	phrase, ok = strings.CutPrefix(phrase, strings.ToLower(subject)+" sacrifices all ")
	if !ok {
		return false
	}
	phrase, ok = strings.CutSuffix(phrase, ".")
	if !ok {
		return false
	}
	phrase, ok = strings.CutSuffix(phrase, " that are one or more colors")
	if !ok {
		return false
	}
	if rest, ok := strings.CutSuffix(phrase, " they control"); ok {
		phrase = rest
	}
	return exactMassGroupPhrase(&effect.Selection, phrase)
}

// massSacrificeSubject maps the per-player sacrifice contexts to their printed
// subject phrase, failing closed for the controller and target contexts the mass
// sacrifice form does not model.
func massSacrificeSubject(context EffectContextKind) (string, bool) {
	switch context {
	case EffectContextEachPlayer:
		return "Each player", true
	case EffectContextEachOpponent:
		return "Each opponent", true
	case EffectContextEachOtherPlayer:
		return "Each other player", true
	case EffectContextReferencedPlayer:
		return "That player", true
	default:
		return "", false
	}
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

// effectSelfNameSpans returns the source spans of the card's own name within an
// effect clause, drawn from its self-name references. Subject-boundary detection
// uses them so an internal comma in a legendary name ("Syr Konrad, the Grim")
// does not truncate the subject.
func effectSelfNameSpans(effect *EffectSyntax) []shared.Span {
	var spans []shared.Span
	for _, reference := range effect.References {
		if reference.Kind == ReferenceSelfName {
			spans = append(spans, reference.Span)
		}
	}
	for _, reference := range effect.SubjectReferences {
		if reference.Kind == ReferenceSelfName {
			spans = append(spans, reference.Span)
		}
	}
	return spans
}

func exactEffectClauseText(effect *EffectSyntax) string {
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb < 0 {
		return ""
	}
	start := effectSubjectStart(effect.Tokens, verb, effectSelfNameSpans(effect))
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
	subjectStart := effectSubjectStart(effect.Tokens, verb, effectSelfNameSpans(effect))
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
				reference.Kind == ReferenceThatObject ||
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
	if effect.DamageRecipient.Reference != DamageRecipientReferenceNone {
		if len(effect.Targets) != 0 {
			return false
		}
		recipient, ok := damageRecipientTokens(effect.Tokens)
		if !ok {
			recipient, ok = damageRecipientTokensAfterAmount(effect.Tokens, effect.Amount)
			if !ok {
				return false
			}
		}
		recipientText := joinedEffectText(recipient)
		switch effect.Amount.DynamicForm {
		case EffectDynamicAmountFormNone:
			amount := "X"
			if effect.Amount.Known {
				amount = strconv.Itoa(effect.Amount.Value)
			} else if !effect.Amount.VariableX {
				return false
			}
			return text == fmt.Sprintf("%s %s damage to %s.", prefix, amount, recipientText)
		case EffectDynamicAmountFormEqual:
			// "<prefix> damage equal to <referent> to <recipient>." reconstructs
			// the dynamic amount phrase verbatim, so the recipient that follows
			// the amount span round-trips exactly and cannot bleed into it.
			return text == fmt.Sprintf("%s damage %s to %s.", prefix, effect.Amount.Text, recipientText)
		default:
			return false
		}
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
		if pair, ok := effect.DamageRecipient.GroupPair(); ok {
			first, ok := exactGroupDamageRecipientText(pair[0])
			if !ok {
				return false
			}
			second, ok := exactGroupDamageRecipientText(pair[1])
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
	if _, ok := SecondTargetDamageRider(effect.DamageRiders); ok {
		return exactSecondTargetDamageEffectSyntax(effect, prefix, text)
	}
	if len(effect.Targets) != 1 || !effect.Targets[0].Exact {
		return false
	}
	target := effect.Targets[0].Text
	switch effect.Amount.DynamicForm {
	case EffectDynamicAmountFormNone:
		amount := "X"
		switch {
		case effect.Amount.Known:
			amount = strconv.Itoa(effect.Amount.Value)
		case effect.Amount.VariableX:
		case effect.Amount.DynamicKind == EffectDynamicAmountTriggeringCounterCount:
			// "deals that much damage" reads the triggering counter count; the
			// amount word reconstructs as the literal "that much" (Shalai and
			// Hallar).
			amount = "that much"
		default:
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
		if selfRider, ok := SelfDamageRider(effect.DamageRiders); ok {
			if !effect.Amount.Known || effect.Targets[0].Cardinality.Max >= 2 {
				return false
			}
			return text == fmt.Sprintf("%s %s damage to %s and %d damage to you.",
				prefix, amount, recipient, selfRider.Value)
		}
		// A "... and N damage to that creature's controller/owner" rider follows
		// a single-target (Max <= 1) fixed-amount clause; the rider recipient is
		// reconstructed from its captured tokens so the round-trip stays exact.
		if _, ok := TargetControllerDamageRider(effect.DamageRiders); ok {
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
// exactly from their own captured phrases. The fixed form requires a known
// primary amount and fixed rider value; the dynamic form ("<prefix> X damage to
// <target0> and X damage to <target1>, where X is ...", The Brothers' War chapter
// III) shares one "where X" dynamic amount across both targets. Either keeps the
// round-trip exact for its bounded shape.
func exactSecondTargetDamageEffectSyntax(effect *EffectSyntax, prefix, text string) bool {
	if len(effect.Targets) != 2 ||
		!effect.Targets[0].Exact || !effect.Targets[1].Exact ||
		effect.Targets[0].Cardinality.Max != 1 {
		return false
	}
	rider, ok := SecondTargetDamageRider(effect.DamageRiders)
	if !ok {
		return false
	}
	if rider.Dynamic {
		if effect.Amount.DynamicForm != EffectDynamicAmountFormWhereX {
			return false
		}
		return text == fmt.Sprintf("%s X damage to %s and X damage to %s, %s.",
			prefix, effect.Targets[0].Text, effect.Targets[1].Text, effect.Amount.Text)
	}
	if !effect.Amount.Known ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormNone {
		return false
	}
	return text == fmt.Sprintf("%s %d damage to %s and %d damage to %s.",
		prefix, effect.Amount.Value, effect.Targets[0].Text,
		rider.Value, effect.Targets[1].Text)
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
	if pair, ok := effect.DamageRecipient.GroupPair(); ok {
		if len(effect.Targets) != 1 || !effect.Targets[0].Exact {
			return false
		}
		first, ok := exactGroupDamageRecipientText(pair[0])
		if !ok {
			return false
		}
		second, ok := exactGroupDamageRecipientText(pair[1])
		if !ok {
			return false
		}
		return text == fmt.Sprintf("%s deals damage %s to %s and %s.",
			effect.Targets[0].Text, effect.Amount.Text, first, second)
	}
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
// the source. It supports a fixed total or the spell's variable X and the
// cardinality and target nouns the executable backend can represent exactly,
// failing closed otherwise.
func exactDividedDamageText(effect *EffectSyntax, prefix, text string) bool {
	if effect.Negated || len(effect.Targets) != 1 {
		return false
	}
	amountText, ok := dividedDamageAmountText(effect.Amount)
	if !ok {
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
	expected := fmt.Sprintf("%s %s damage divided as you choose among %s %s.",
		prefix, amountText, cardinality, noun)
	return text == expected
}

// dividedDamageAmountText reconstructs the canonical amount token for a divided
// damage clause: the literal integer for a fixed total of at least one, or "X"
// for the spell's bare variable X. It fails closed for a non-positive fixed
// total, for any dynamic amount form ("equal to ...", "where X is ..."), and for
// the "X plus N" rider, none of which the divided path can represent, so those
// wordings keep failing the round-trip.
func dividedDamageAmountText(amount EffectAmountSyntax) (string, bool) {
	if amount.DynamicForm != EffectDynamicAmountFormNone ||
		amount.DynamicKind != EffectDynamicAmountNone ||
		amount.Addend != 0 || amount.Multiplier != 0 {
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
// supports "targets" (any target) and the creature-noun forms the divided
// lowering can represent: a plain "target creatures", a "creatures and/or
// planeswalkers" card-type union, an attacking/blocking combat state, and a
// single "with"/"without" keyword qualifier. It fails closed for every other
// selector, leaving the byte-exact round-trip to reject the wording.
func dividedTargetNoun(selection SelectionSyntax) (string, bool) {
	switch selection.Kind {
	case SelectionAny:
		return "targets", true
	case SelectionCreature:
		words, ok := dividedCreatureNounWords(selection)
		if !ok {
			return "", false
		}
		return strings.Join(append([]string{"target"}, words...), " "), true
	default:
		return "", false
	}
}

// dividedCreatureNounWords reconstructs the words that follow the "target"
// determiner of a divided-damage creature noun: an optional attacking/blocking
// combat prefix, the plural creature noun or "creature and/or planeswalker"
// union, and a single "with"/"without" keyword clause. It fails closed for every
// controller, color, subtype, supertype, tapped, numeric, or determiner
// qualifier the divided lowering does not yet represent.
func dividedCreatureNounWords(selection SelectionSyntax) ([]string, bool) {
	if selection.All || selection.Another || selection.Other ||
		selection.Tapped || selection.Untapped ||
		selection.Colorless || selection.Multicolored ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.PowerLessThanSource || selection.PowerGreaterThanSource ||
		selection.TokenOnly || selection.NonToken ||
		selection.Controller != SelectionControllerAny ||
		selection.Zone != zone.None ||
		len(selection.ExcludedTypes) != 0 ||
		len(selection.ExcludedColors) != 0 ||
		len(selection.ColorsAny) != 0 ||
		len(selection.SubtypesAny) != 0 ||
		len(selection.Supertypes) != 0 ||
		len(selection.ExcludedSupertypes) != 0 ||
		len(selection.ExcludedSubtypes) != 0 {
		return nil, false
	}
	var words []string
	switch {
	case selection.Attacking && selection.Blocking:
		words = append(words, "attacking", "or", "blocking")
	case selection.Attacking:
		words = append(words, "attacking")
	case selection.Blocking:
		words = append(words, "blocking")
	default:
	}
	nouns, ok := dividedCreatureNouns(selection.RequiredTypesAny)
	if !ok {
		return nil, false
	}
	words = append(words, nouns...)
	keywordWords, ok := permanentKeywordQualifierWords(selection)
	if !ok {
		return nil, false
	}
	words = append(words, keywordWords...)
	return words, true
}

// dividedCreatureNouns reconstructs the plural card-type noun(s) of a divided
// creature target. A bare or single creature type renders as "creatures"; a
// card-type union renders each member pluralized and joined with "and/or"
// ("creatures and/or planeswalkers"). The first listed type must be creature so
// the noun matches the SelectionCreature kind, and every member must name a
// permanent card type the round-trip can spell.
func dividedCreatureNouns(required []CardType) ([]string, bool) {
	if len(required) == 0 {
		return []string{"creatures"}, true
	}
	if required[0] != CardTypeCreature {
		return nil, false
	}
	nouns := make([]string, 0, len(required)*2)
	for i, cardType := range required {
		noun, ok := permanentCardTypeNoun(cardType)
		if !ok {
			return nil, false
		}
		if i > 0 {
			nouns = append(nouns, "and/or")
		}
		nouns = append(nouns, noun+"s")
	}
	return nouns, true
}

// exactGroupDamageAmountText reconstructs the canonical amount token for a group
// damage clause: the literal integer for a fixed amount of at least one, or "X"
// for the spell's variable X. It fails closed for a non-positive fixed amount
// and for any dynamic amount form ("equal to ...", "where X is ..."), which the
// group damage path reconstructs separately or not at all, so those wordings
// keep failing the round-trip.
func exactGroupDamageAmountText(amount EffectAmountSyntax) (string, bool) {
	if amount.DynamicForm != EffectDynamicAmountFormNone {
		return "", false
	}
	if amount.DynamicKind == EffectDynamicAmountTriggeringCounterCount {
		// "<source> deals that much damage to <group>." reads the triggering
		// event's quantity; the amount word reconstructs as the literal "that
		// much", mirroring the single-target branch (Magmakin Artillerist).
		return "that much", true
	}
	if amount.DynamicKind != EffectDynamicAmountNone {
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
	if len(effect.DamageRecipient.Groups) != 0 {
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
		return text == fmt.Sprintf("%s damage to %s %s.", prefix, recipient, effect.Amount.Text) ||
			text == fmt.Sprintf("%s damage %s to %s.", prefix, effect.Amount.Text, recipient)
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
// single-excluded-type, single-excluded-subtype, nontoken/token, keyword, and
// "other" qualifiers the executable backend can represent exactly, and fails
// closed for every other qualifier.
func exactGroupDamagePermanentRecipientText(selection SelectionSyntax) (string, bool) {
	if selection.All || selection.Another || selection.Zone != zone.None ||
		selection.Colorless || selection.Multicolored ||
		len(selection.Supertypes) > 1 ||
		len(selection.ExcludedColors) != 0 ||
		(len(selection.RequiredTypesAny) > 1 && !selection.ConjunctiveTypes) ||
		len(selection.ColorsAny) > 1 ||
		len(selection.ExcludedTypes) > 1 ||
		len(selection.ExcludedSubtypes) > 1 ||
		(selection.NonToken && selection.TokenOnly) {
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
	// A conjunctive two-type group ("each artifact creature you control") renders
	// the combined noun "artifact creature" the parser captured rather than the
	// single Kind noun. The two card types live in RequiredTypesAny with the
	// ConjunctiveTypes marker, so render the joined noun and skip the
	// single-type redundancy check below.
	if selection.ConjunctiveTypes {
		conjunctiveNoun, ok := conjunctiveCreatureTargetNoun(selection)
		if !ok {
			return "", false
		}
		noun, hasNoun = conjunctiveNoun, true
	} else if len(selection.RequiredTypesAny) == 1 {
		// The parser records a permanent noun both as the selection Kind and as a
		// redundant single-element RequiredTypesAny. Accept only that redundant
		// form (a union or a type inconsistent with the noun is not
		// representable here).
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
	if selection.NonToken {
		words = append(words, "nontoken")
	} else if selection.TokenOnly {
		words = append(words, "token")
	}
	if len(selection.Supertypes) == 1 {
		supertypeText, ok := supertypeWord(selection.Supertypes[0])
		if !ok {
			return "", false
		}
		words = append(words, supertypeText)
	}
	if len(selection.ColorsAny) == 1 {
		colorText, ok := colorWord(selection.ColorsAny[0])
		if !ok {
			return "", false
		}
		words = append(words, colorText)
	}
	if len(selection.SubtypesAny) > 1 {
		// A multi-subtype union ("each Pest, Bat, Insect, Snake, and Spider you
		// control") carries no card-type noun: the subtype list is the recipient
		// noun. The single-subtype redundancy ("each Goblin") is handled below.
		if hasNoun {
			return "", false
		}
	}
	words = append(words, groupSubtypeListWords(selection.SubtypesAny)...)
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
	if len(selection.ExcludedSubtypes) == 1 {
		if !hasNoun {
			return "", false
		}
		words = append(words, "non-"+string(selection.ExcludedSubtypes[0]))
	}
	if hasNoun {
		words = append(words, noun)
	} else if len(selection.SubtypesAny) == 0 {
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
		if selection.OpponentEach {
			words = append(words, "each", "opponent", "controls")
		} else {
			words = append(words, "your", "opponents", "control")
		}
	case SelectionControllerNotYou:
		words = append(words, "you", "don't", "control")
	default:
		return "", false
	}
	// A "named <Name>" filter ("each other creature you control named Charmed
	// Stray") follows the controller clause in canonical Oracle wording. The
	// verbatim name carries its own internal spacing, so it joins as one word.
	if selection.RequiredName != "" {
		words = append(words, "named", selection.RequiredName)
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
	counterWords, ok := groupSelectorCounterWords(selection)
	if !ok {
		return "", false
	}
	words = append(words, counterWords...)
	if selection.EnteredThisTurn {
		words = append(words, "that", "entered", "this", "turn")
	}
	recipient := strings.Join(words, " ")
	rider, ok := groupSelectorNumericRider(selection)
	if !ok {
		return "", false
	}
	if !selectionPhraseVerifiesGroupRecipient(selection, recipient, rider, counterWords) {
		return "", false
	}
	return recipient + rider, true
}

// selectionPhraseVerifiesGroupRecipient soft-gates the group damage recipient
// reconstruction through the canonical selectionPhrase renderer. When the
// renderer can model the selection's noun phrase, the bespoke recipient must
// match its "each"-determiner rendering exactly, so the two reconstructions
// cannot silently drift; when the renderer cannot model the selection
// (ok=false), the bespoke reconstruction stands alone. The cross-check is
// skipped when the recipient carries a trailing numeric rider or a counter
// clause, because selectionPhrase orders the numeric qualifier ahead of the
// controller clause and does not render counter qualifiers, so its rendering
// is intentionally not comparable to the recipient in those forms.
func selectionPhraseVerifiesGroupRecipient(selection SelectionSyntax, recipient, rider string, counterWords []string) bool {
	if rider != "" || len(counterWords) != 0 {
		return true
	}
	rendered, ok := selectionPhrase(selection, selectionPhraseOptions{
		Number:     numberSingular,
		Determiner: determinerEach,
	})
	if !ok {
		return true
	}
	return strings.EqualFold(rendered, recipient)
}

// exactSingularChosenPermanentRecipientText reconstructs the recipient text for a
// non-target single-choice counter recipient ("a creature you control",
// "another creature you control", Ajani Fells the Godsire chapter II). It reuses
// the group recipient assembly and swaps the "each" determiner for the singular
// article, so every filter the group form supports (types, colors, controller,
// keyword qualifiers) is supported here too. A genuine group ("each …", All set)
// is left to exactGroupDamagePermanentRecipientText.
func exactSingularChosenPermanentRecipientText(selection SelectionSyntax) (string, bool) {
	if selection.All {
		return "", false
	}
	singular := selection.Another || selection.Other
	probe := selection
	probe.Another = false
	probe.Other = false
	group, ok := exactGroupDamagePermanentRecipientText(probe)
	if !ok {
		return "", false
	}
	rest, ok := strings.CutPrefix(group, "each ")
	if !ok {
		return "", false
	}
	if singular {
		return "another " + rest, true
	}
	return singularArticleFor(rest) + " " + rest, true
}

func singularArticleFor(phrase string) string {
	if phrase == "" {
		return "a"
	}
	switch phrase[0] {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return "an"
	default:
		return "a"
	}
}

// subtype is its own noun ("each Goblin you control"). Two subtypes join with
// "and/or" ("each Merfolk and/or Knight you control"); three or more form a
// comma-separated list closed by "and" ("each Pest, Bat, Insect, Snake, and
// Spider you control"). The reconstruction is compared byte-for-byte against the
// source, so a recipient written with a different joiner fails closed instead of
// matching an approximation.
func groupSubtypeListWords(subtypes []types.Sub) []string {
	switch len(subtypes) {
	case 0:
		return nil
	case 1:
		return []string{string(subtypes[0])}
	case 2:
		return []string{string(subtypes[0]), "and/or", string(subtypes[1])}
	default:
		words := make([]string, 0, len(subtypes)+1)
		for i, sub := range subtypes {
			if i == len(subtypes)-1 {
				words = append(words, "and", string(sub))
			} else {
				words = append(words, string(sub)+",")
			}
		}
		return words
	}
}

// groupSelectorCounterWords reconstructs the canonical "with a <kind> counter on
// it" / "with a counter on it" qualifier a group recipient selector may carry,
// mirroring the runtime Selection.MatchCounter / MatchAnyCounter predicate. The
// kind-specific form names the counter ("with a +1/+1 counter on it"); the
// kind-agnostic form omits it ("with a counter on it"). A selector with neither
// flag yields no words. The two flags are mutually exclusive, so a selector that
// sets both fails closed rather than rendering an ambiguous qualifier.
func groupSelectorCounterWords(selection SelectionSyntax) ([]string, bool) {
	if selection.CounterRequired && selection.CounterAny {
		return nil, false
	}
	switch {
	case selection.CounterRequired:
		return []string{"with", "a", selection.CounterKind.String(), "counter", "on", "it"}, true
	case selection.CounterAny:
		return []string{"with", "a", "counter", "on", "it"}, true
	default:
		return nil, true
	}
}

// groupSelectorNumericRider reconstructs the canonical " with mana value N or
// less", " with power N or greater", or " with toughness N or less" rider a
// group recipient selector may carry, mirroring the runtime Selection numeric
// bound. It renders at most one numeric comparison (the runtime carries a single
// mana-value/power/toughness bound per selection), returning ok=false when more
// than one is active or the comparison has no canonical Oracle phrasing so the
// reconstruction fails closed instead of approximating the filter. An inactive
// numeric filter yields the empty rider and ok=true.
func groupSelectorNumericRider(selection SelectionSyntax) (string, bool) {
	active := 0
	rider := ""
	if selection.MatchManaValue {
		active++
		if selection.ManaValueX {
			rider = " with mana value X or less"
		} else {
			clause, ok := numericComparisonClause(selection.ManaValue)
			if !ok {
				return "", false
			}
			rider = " with mana value " + clause
		}
	}
	if selection.MatchPower {
		active++
		clause, ok := numericComparisonClause(selection.Power)
		if !ok {
			return "", false
		}
		rider = " with power " + clause
	}
	if selection.MatchToughness {
		active++
		clause, ok := numericComparisonClause(selection.Toughness)
		if !ok {
			return "", false
		}
		rider = " with toughness " + clause
	}
	if active > 1 {
		return "", false
	}
	return rider, true
}

// numericComparisonClause renders the canonical Oracle phrasing for a fixed
// integer comparison ("N or less", "N or greater", "N"), mirroring the forms
// exactMassComparisonClause accepts. It fails closed for the strict and
// unbounded comparisons no canonical permanent-filter wording uses.
func numericComparisonClause(bound compare.Int) (string, bool) {
	switch bound.Op {
	case compare.LessOrEqual:
		return strconv.Itoa(bound.Value) + " or less", true
	case compare.GreaterOrEqual:
		return strconv.Itoa(bound.Value) + " or greater", true
	case compare.Equal:
		return strconv.Itoa(bound.Value), true
	default:
		return "", false
	}
}
