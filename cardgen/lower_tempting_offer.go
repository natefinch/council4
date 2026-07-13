package cardgen

import (
	"reflect"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// temptingOfferAbilityWord is the ability word that flags the Tempt cycle's
// shared "Tempting offer" idiom. It is a rules-free flavor prefix, but the
// executable backend does not strip it like other rules-free ability words:
// leaving it in place routes every "Tempting offer" ability through
// lowerTemptingOfferAbility, which either lowers the idiom or fails closed, so a
// Tempt card the backend cannot yet model can never fall through to generic
// lowering and silently drop the each-opponent offer or the reward repeat.
//
// The cycle prints the word inconsistently — Tempt with Bunnies capitalizes it
// as "Tempting Offer" while the rest print "Tempting offer" — so the gate matches
// it case-insensitively (see isTemptingOfferAbilityWord).
const temptingOfferAbilityWord = "Tempting offer"

// isTemptingOfferAbilityWord reports whether an ability word is the "Tempting
// offer" idiom, matched case-insensitively so the "Tempting Offer" capitalization
// on Tempt with Bunnies is recognized alongside the lowercase printing on the
// rest of the cycle.
func isTemptingOfferAbilityWord(word string) bool {
	return strings.EqualFold(word, temptingOfferAbilityWord)
}

// lowerTemptingOfferAbility recognizes the "Tempting offer" ability-word idiom
// shared by the Tempt cycle (Tempt with Vengeance and its siblings):
//
//	Tempting offer — Create X 1/1 red Elemental creature tokens with haste. Each
//	opponent may create X 1/1 red Elemental creature tokens with haste. For each
//	opponent who does, create X 1/1 red Elemental creature tokens with haste.
//	(Tempt with Vengeance)
//
// The idiom is a same-kind triple of effects: the controller performs an effect
// (effect 0), each opponent may perform the identical effect for themselves
// (effect 1), and for each opponent who does the controller performs the effect
// once more (effect 2). It lowers to a single Optional + OptionalActorGroup
// instruction flagged TemptingOffer, whose primitive addresses the acting player
// through GroupOfferMemberReference(); the runtime binds that reference to the
// controller for the base and reward resolutions and to each accepting opponent
// for that opponent's own resolution (see resolveTemptingOffer).
//
// It is gated on ability.AbilityWord == "Tempting offer", so it never touches any
// other corpus card. Once that gate matches, the ability is owned here: a
// structure or effect the backend cannot model fails closed with a diagnostic
// rather than returning handled=false, which would let the ability strip the
// word and mis-lower through the generic path.
func lowerTemptingOfferAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, bool, *shared.Diagnostic) {
	if ability.Kind != compiler.AbilitySpell || !isTemptingOfferAbilityWord(ability.AbilityWord) {
		return abilityLowering{}, false, nil
	}
	content, diagnostic := lowerTemptingOfferContent(ability, syntax)
	if diagnostic != nil {
		return abilityLowering{}, true, diagnostic
	}
	return abilityLowering{
		spellAbility: opt.Val(content),
		consumed: semanticConsumption{
			targets:    len(ability.Content.Targets),
			conditions: len(ability.Content.Conditions),
			effects:    len(ability.Content.Effects),
			keywords:   len(ability.Content.Keywords),
			references: len(ability.Content.References),
		},
		// The whole ability, including the "Tempting offer" ability word, is
		// consumed by this lowering, so credit its full span for coverage.
		sourceSpans: []shared.Span{ability.Span},
	}, true, nil
}

// lowerTemptingOfferContent verifies the Tempting-offer structure and lowers it,
// or fails closed with a diagnostic. It routes each Tempt shape to its
// recognizer: the search-and-ramp idiom (Tempt with Discovery) whenever the body
// contains a library search, the classic exact-three-same-kind idiom (Tempt with
// Vengeance, Glory, Immortality, Reflections) when the body is exactly three
// effects, and the generic compound-body idiom (Tempt with Bunnies) for a shared
// multi-effect body. No modes or keywords are allowed. Every other shape fails
// closed pending follow-up work under the Tempt-cycle generalization.
func lowerTemptingOfferContent(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	c := ability.Content
	if len(c.Modes) != 0 || len(c.Keywords) != 0 || len(c.Conditions) != 0 {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "modes, conditions, or keywords are unsupported")
	}
	if temptingOfferHasSearch(c) {
		return lowerTemptingOfferSearch(ability, syntax)
	}
	if len(c.Effects) == 3 {
		return lowerTemptingOfferThreeEffect(ability, syntax)
	}
	return lowerTemptingOfferCompoundBody(ability, syntax)
}

// temptingOfferHasSearch reports whether any effect of the shared body is a
// library search, marking the search-and-ramp idiom (Tempt with Discovery) so it
// routes to its dedicated recognizer rather than the same-kind or compound paths.
func temptingOfferHasSearch(content compiler.AbilityContent) bool {
	for i := range content.Effects {
		if content.Effects[i].Kind == compiler.EffectSearch {
			return true
		}
	}
	return false
}

// lowerTemptingOfferThreeEffect lowers the classic Tempting-offer shape: exactly
// three effects of the same kind with the you / each-opponent-may / you contexts,
// dispatched to a kind-specific lowering. The modeled shared effects are token
// creation (Tempt with Vengeance), a copy of a single targeted creature (Tempt
// with Reflections), a +1/+1 counter placement on each creature the acting player
// controls (Tempt with Glory), and a creature reanimation from the acting
// player's graveyard (Tempt with Immortality). Every other kind fails closed.
func lowerTemptingOfferThreeEffect(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	c := ability.Content
	base, offer, reward := c.Effects[0], c.Effects[1], c.Effects[2]
	if base.Kind != offer.Kind || base.Kind != reward.Kind {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the three effects are not the same kind")
	}
	if base.Context != parser.EffectContextController || base.Optional ||
		offer.Context != parser.EffectContextEachOpponent || !offer.Optional ||
		reward.Context != parser.EffectContextController || reward.Optional {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the three effects are not the you / each-opponent-may / you idiom")
	}
	switch base.Kind {
	case compiler.EffectCreate:
		return lowerTemptingOfferCreate(ability, syntax)
	case compiler.EffectPut:
		return lowerTemptingOfferCounter(ability, syntax)
	case compiler.EffectReturn:
		return lowerTemptingOfferReturn(ability, syntax)
	default:
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the shared effect is not a modeled Tempting offer effect")
	}
}

// temptingOfferAbilityContent wraps a single acting-player-addressed primitive as
// the lone instruction of a Tempting-offer ability: it is optional, offered to
// the opponents as a group, and flagged TemptingOffer so the runtime performs the
// base, per-opponent, and reward resolutions through resolveTemptingOffer. The
// primitive addresses the acting player through GroupOfferMemberReference(), and
// targets, when present, are shared by every resolution (the controller chooses
// them once at cast time).
func temptingOfferAbilityContent(primitive game.Primitive, targets []game.TargetSpec) game.AbilityContent {
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive:          primitive,
			Optional:           true,
			OptionalActorGroup: opt.Val(game.OpponentsReference()),
			TemptingOffer:      true,
		}},
	}.Ability()
}

// temptingOfferBenignReferences reports whether every ability-level reference is a
// bare controller pronoun ("they"/"their" in "each creature they control", "their
// graveyard"). Those pronouns only name the acting player, whose identity the
// primitive already carries through GroupOfferMemberReference(), so they are
// consumed wholesale without affecting the lowered primitive. Any other reference
// shape fails closed rather than being silently dropped.
func temptingOfferBenignReferences(references []compiler.CompiledReference) bool {
	for i := range references {
		if references[i].Kind != compiler.ReferencePronoun {
			return false
		}
	}
	return true
}

// lowerTemptingOfferCreate lowers a Tempting-offer whose shared effect creates
// tokens. A synthesized-token creation with no target or reference (Tempt with
// Vengeance) lowers each clause as a plain controller-recipient token and
// requires all three to be identical. A copy of a single targeted creature (Tempt
// with Reflections) lowers to a copy token addressing the target the controller
// chooses at cast time. Every other token shape fails closed.
func lowerTemptingOfferCreate(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	if len(ability.Content.Targets) == 0 && len(ability.Content.References) == 0 {
		return lowerTemptingOfferSynthesizedToken(ability, syntax)
	}
	return lowerTemptingOfferCopyToken(ability, syntax)
}

// lowerTemptingOfferSynthesizedToken lowers a Tempting-offer whose shared effect
// creates synthesized creature tokens (Tempt with Vengeance). Each of the three
// effects is lowered independently as a plain controller-recipient token creation
// and the resulting CreateToken primitives are required to be identical, proving
// all three create the same tokens; the offer then addresses the acting player
// through GroupOfferMemberReference() so the runtime enters the tokens under the
// controller (base and reward) or the accepting opponent (their own copy).
func lowerTemptingOfferSynthesizedToken(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	parent := contentCtx{
		text:          syntax.Text,
		span:          ability.Content.Span,
		content:       ability.Content,
		enclosingKind: compiler.AbilitySpell,
	}
	var tokens [3]game.CreateToken
	for i := range ability.Content.Effects {
		token, diagnostic := lowerTemptingOfferToken(parent, ability, ability.Content.Effects[i])
		if diagnostic != nil {
			return game.AbilityContent{}, diagnostic
		}
		tokens[i] = token
	}
	if !reflect.DeepEqual(tokens[0], tokens[1]) || !reflect.DeepEqual(tokens[0], tokens[2]) {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the three token creations are not identical")
	}
	primitive := tokens[0]
	primitive.Recipient = opt.Val(game.GroupOfferMemberReference())
	return temptingOfferAbilityContent(primitive, nil), nil
}

// lowerTemptingOfferCopyToken lowers a Tempting-offer whose shared effect creates
// a token that is a copy of a single targeted creature (Tempt with Reflections):
// "Choose target creature you control. Create a token that's a copy of that
// creature. Each opponent may create a token that's a copy of that creature. For
// each opponent who does, create a token that's a copy of that creature." The
// creature is targeted once by the controller at cast time, so its
// TargetPermanentReference is shared by every resolution; the acting player enters
// their own copy through GroupOfferMemberReference(). It fails closed on any token
// shape that is not exactly a one-for-one copy of the single target.
func lowerTemptingOfferCopyToken(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	c := ability.Content
	if len(c.Targets) != 1 {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "a copy token creation requires exactly one target")
	}
	targetSpec, ok := permanentTargetSpec(c.Targets[0])
	if !ok {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the copy token target is unsupported")
	}
	var tokens [3]game.CreateToken
	for i := range c.Effects {
		token, ok := temptingOfferCopyTokenPrimitive(c.Effects[i])
		if !ok {
			return game.AbilityContent{}, temptingOfferDiagnostic(ability, "a copy token creation is unsupported")
		}
		tokens[i] = token
	}
	if !reflect.DeepEqual(tokens[0], tokens[1]) || !reflect.DeepEqual(tokens[0], tokens[2]) {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the three copy token creations are not identical")
	}
	primitive := tokens[0]
	primitive.Recipient = opt.Val(game.GroupOfferMemberReference())
	return temptingOfferAbilityContent(primitive, []game.TargetSpec{targetSpec}), nil
}

// temptingOfferCopyTokenPrimitive builds the CreateToken for one clause of a
// copy-of-target Tempting offer, normalizing the clause's context, optionality,
// and exactness (the each-opponent and reward clauses inherit their idiom prefix
// and are not marked exact) so all three lower to the same copy token. The clause
// must copy exactly the single ability-level target through a "that creature"
// reference, so it carries one target-bound ReferenceThatObject and no other copy
// modifier. Any richer copy shape fails closed.
func temptingOfferCopyTokenPrimitive(effect compiler.CompiledEffect) (game.CreateToken, bool) {
	normalized := effect
	normalized.Context = parser.EffectContextController
	normalized.Optional = false
	normalized.Exact = true
	if normalized.Kind != compiler.EffectCreate ||
		normalized.Negated ||
		normalized.DelayedTiming != 0 ||
		normalized.Duration != compiler.DurationNone ||
		normalized.TokenCopyEntersTapped ||
		len(normalized.TokenCopyGrantKeywords) != 0 ||
		!normalized.Amount.Known ||
		normalized.Amount.Value != 1 {
		return game.CreateToken{}, false
	}
	if len(normalized.References) != 1 ||
		normalized.References[0].Kind != compiler.ReferenceThatObject ||
		normalized.References[0].Binding != compiler.ReferenceBindingTarget ||
		normalized.References[0].Occurrence != 0 {
		return game.CreateToken{}, false
	}
	spec, ok := tokenCopyModifiers(&normalized, game.TargetPermanentReference(0))
	if !ok {
		return game.CreateToken{}, false
	}
	return game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenCopyOf(spec),
	}, true
}

// lowerTemptingOfferCounter lowers a Tempting-offer whose shared effect places
// permanent counters on each creature the acting player controls (Tempt with
// Glory: "Put a +1/+1 counter on each creature you control. Each opponent may put
// a +1/+1 counter on each creature they control. For each opponent who does, put a
// +1/+1 counter on each creature you control."). Each clause lowers to an
// AddCounter over the acting player's creatures, addressed through
// GroupOfferMemberReference(); "you control" and "they control" both name the
// acting player, so all three must lower identically. It fails closed on any
// counter shape it cannot express.
func lowerTemptingOfferCounter(
	ability compiler.CompiledAbility,
	_ *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	c := ability.Content
	if len(c.Targets) != 0 {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "a counter placement with a target is unsupported")
	}
	if !temptingOfferBenignReferences(c.References) {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "a counter placement with a non-pronoun reference is unsupported")
	}
	var placements [3]game.AddCounter
	for i := range c.Effects {
		add, ok := temptingOfferCounterPrimitive(c.Effects[i])
		if !ok {
			return game.AbilityContent{}, temptingOfferDiagnostic(ability, "a counter placement is unsupported")
		}
		placements[i] = add
	}
	if !reflect.DeepEqual(placements[0], placements[1]) || !reflect.DeepEqual(placements[0], placements[2]) {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the three counter placements are not identical")
	}
	return temptingOfferAbilityContent(placements[0], nil), nil
}

// temptingOfferCounterPrimitive builds the AddCounter for one clause of a
// counter-placement Tempting offer, mirroring the target-player-controlled counter
// lowering: it accepts a supported, permanent-only, fixed counter placement on a
// player's creatures and reconstructs the recipient group through the shared
// group-counter recipient projection. The acting player supplies the group's
// controller relationship through GroupOfferMemberReference(), so the selector's
// own "you control" / "they control" constraint is normalized away before the
// group is built. Any richer counter shape fails closed.
func temptingOfferCounterPrimitive(effect compiler.CompiledEffect) (game.AddCounter, bool) {
	if effect.Kind != compiler.EffectPut ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		!effect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.CounterKind.PlayerOnly() ||
		effect.CounterRecipientSingleChoice ||
		effect.Selector.MatchCounter ||
		effect.Selector.MatchAnyCounter ||
		!effect.Amount.Known ||
		effect.Amount.Value <= 0 {
		return game.AddCounter{}, false
	}
	selector := effect.Selector
	selector.Controller = compiler.ControllerAny
	group, ok := targetPlayerControlledCounterGroup(selector, game.GroupOfferMemberReference())
	if !ok {
		return game.AddCounter{}, false
	}
	return game.AddCounter{
		Amount:      game.Fixed(effect.Amount.Value),
		Group:       group,
		CounterKind: effect.CounterKind,
	}, true
}

// lowerTemptingOfferReturn lowers a Tempting-offer whose shared effect reanimates
// a creature from the acting player's graveyard (Tempt with Immortality: "Return a
// creature card from your graveyard to the battlefield. Each opponent may return a
// creature card from their graveyard to the battlefield. For each opponent who
// does, return a creature card from your graveyard to the battlefield."). Each
// clause lowers to a graveyard-return choice addressed through
// GroupOfferMemberReference(), so the acting player chooses from and reanimates
// under their own control; "your graveyard" and "their graveyard" both name the
// acting player, so all three must lower identically. It fails closed on any
// reanimation shape it cannot express.
func lowerTemptingOfferReturn(
	ability compiler.CompiledAbility,
	_ *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	c := ability.Content
	if len(c.Targets) != 0 {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "a reanimation with a target is unsupported")
	}
	if !temptingOfferBenignReferences(c.References) {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "a reanimation with a non-pronoun reference is unsupported")
	}
	var returns [3]game.ChooseFromZone
	for i := range c.Effects {
		env, ok := temptingOfferReturnPrimitive(c.Effects[i])
		if !ok {
			return game.AbilityContent{}, temptingOfferDiagnostic(ability, "a reanimation is unsupported")
		}
		returns[i] = env
	}
	if !reflect.DeepEqual(returns[0], returns[1]) || !reflect.DeepEqual(returns[0], returns[2]) {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the three reanimations are not identical")
	}
	return temptingOfferAbilityContent(returns[0], nil), nil
}

// temptingOfferReturnPrimitive builds the graveyard-return ChooseFromZone for one
// clause of a reanimation Tempting offer, mirroring the chosen-card graveyard
// return lowering: exactly one non-target creature card returns from the acting
// player's graveyard to the battlefield with no entry riders. The acting player
// supplies the graveyard and the entering permanent's controller through
// GroupOfferMemberReference(): the runtime scopes the candidate pool to that
// player's graveyard, so the card's own zone owner ("your" / "their" graveyard)
// is already honored by that scoping. A card in a graveyard has no controller, so
// the selector's controller relation is normalized to "any" before the selection
// is built; keeping a ControllerYou relation would compare the acting player
// against a card's absent (zero-value) controller and wrongly reject every
// opponent's own graveyard cards. Any richer reanimation shape fails closed.
func temptingOfferReturnPrimitive(effect compiler.CompiledEffect) (game.ChooseFromZone, bool) {
	if !plainGraveyardReturn(effect) ||
		effect.ToZone != zone.Battlefield ||
		effect.EntersTapped ||
		effect.CounterKindKnown ||
		!effect.Amount.Known ||
		effect.Amount.RangeKnown ||
		effect.Amount.VariableX ||
		effect.Amount.DynamicKind != 0 ||
		effect.Amount.Value != 1 {
		return game.ChooseFromZone{}, false
	}
	selector := effect.Selector
	if selector.Zone != zone.Graveyard ||
		selector.All ||
		selector.Another ||
		selector.Other ||
		selector.Attacking ||
		selector.Blocking ||
		selector.Tapped ||
		selector.Untapped {
		return game.ChooseFromZone{}, false
	}
	selector.Controller = compiler.ControllerAny
	selection, ok := cardSelectionForSelector(selector)
	if !ok {
		return game.ChooseFromZone{}, false
	}
	return game.ReturnFromGraveyardChoice(
		game.GroupOfferMemberReference(),
		selection,
		game.Fixed(1),
		zone.Battlefield,
		false,
		opt.V[int]{},
		false,
		"",
	), true
}

// lowerTemptingOfferToken lowers a single Tempting-offer effect as a plain
// controller-recipient token creation, normalizing its context, optionality, and
// exactness so the each-opponent "may" clause and the "for each opponent who
// does" reward clause lower to the same base token as the controller clause. The
// parser marks only the first clause Exact; the two idiom clauses inherit their
// idiom prefix ("Each opponent may", "For each opponent who does") and so are not
// exact, though their token specifications are identical. Normalizing Exact is
// safe because the caller compares all three lowered CreateToken primitives for
// deep equality, so any real difference in the token created still fails closed.
// It returns the CreateToken primitive or a diagnostic when the effect is not a
// plain synthesized-token creation.
func lowerTemptingOfferToken(
	parent contentCtx,
	ability compiler.CompiledAbility,
	effect compiler.CompiledEffect,
) (game.CreateToken, *shared.Diagnostic) {
	normalized := effect
	normalized.Context = parser.EffectContextController
	normalized.Optional = false
	normalized.Exact = true
	effectCtx := contextForEffect(parent, &normalized)
	content, diagnostic := lowerCreateTokenSpell(effectCtx)
	if diagnostic != nil {
		return game.CreateToken{}, temptingOfferDiagnostic(ability, "a token creation is unsupported")
	}
	if content.IsModal() ||
		len(content.SharedTargets) != 0 ||
		len(content.Modes) != 1 ||
		len(content.Modes[0].Targets) != 0 ||
		len(content.Modes[0].Sequence) != 1 {
		return game.CreateToken{}, temptingOfferDiagnostic(ability, "a token creation has an unexpected shape")
	}
	token, ok := content.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok || token.Recipient.Exists || token.RecipientGroup.Kind != game.PlayerGroupReferenceNone {
		return game.CreateToken{}, temptingOfferDiagnostic(ability, "a token creation is not a plain controller-recipient token")
	}
	return token, nil
}

// temptingOfferBodyAbilityContent wraps a multi-instruction shared body as the
// lone Tempting-offer instruction: it is optional, offered to the opponents as a
// group, and flagged TemptingOffer, and carries its body in TemptingOfferBody
// (with a nil Primitive) so the runtime runs the whole body atomically for the
// base, per-opponent, and reward resolutions through resolveTemptingOffer. Every
// body instruction addresses the acting player through GroupOfferMemberReference().
func temptingOfferBodyAbilityContent(body []game.Instruction, targets []game.TargetSpec) game.AbilityContent {
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Optional:           true,
			OptionalActorGroup: opt.Val(game.OpponentsReference()),
			TemptingOffer:      true,
			TemptingOfferBody:  body,
		}},
	}.Ability()
}

// lowerTemptingOfferSearch lowers the search-and-ramp Tempting offer (Tempt with
// Discovery): "Search your library for a land card and put it onto the
// battlefield. Each opponent may search their library for a land card and put it
// onto the battlefield. For each opponent who searches a library this way, search
// your library for a land card and put it onto the battlefield. Then each player
// who searched a library this way shuffles." Each clause is a search+put pair
// collapsed to one game.Search primitive addressed through
// GroupOfferMemberReference(): the runtime searches the controller's library for
// the base and each reward and the accepting opponent's library for that
// opponent's own search, entering the found land under the searcher's control.
//
// The count that drives the controller's reward repeat is the number of opponents
// who searched — that is, who accepted the offer — regardless of whether they
// found a land (CR search rulings: a player who searches but declines to find
// still searched). The generic resolveTemptingOffer already repeats the body once
// per accepting member, so this idiom needs no card-specific repeat primitive.
// Every searcher's library is shuffled by the search primitive itself (each search
// shuffles the searched library), which is exactly the trailing "then each player
// who searched a library this way shuffles" — a player who declined never searched
// and never shuffles. The trailing shuffle clause and the "for each opponent who
// searches a library this way" quantifier are therefore verified and consumed
// rather than lowered to their own instructions. Any richer search shape fails
// closed.
func lowerTemptingOfferSearch(
	ability compiler.CompiledAbility,
	_ *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	c := ability.Content
	if len(c.Targets) != 0 {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "a search Tempting offer with a target is unsupported")
	}
	if !temptingOfferBenignReferences(c.References) {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "a search Tempting offer with a non-pronoun reference is unsupported")
	}
	// Base clause: the controller searches and puts (a non-optional "you" pair).
	base := c.Effects[0]
	baseSearch, baseNext, ok := temptingOfferSearchPair(c.Effects, 0)
	if !ok || base.Context != parser.EffectContextController || base.Optional {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the base clause is not a plain controller library search")
	}
	// Offer clause: each opponent may search and put (an optional each-opponent pair).
	if baseNext >= len(c.Effects) ||
		c.Effects[baseNext].Context != parser.EffectContextEachOpponent ||
		!c.Effects[baseNext].Optional {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the offer clause is not an each-opponent-may library search")
	}
	offerSearch, offerNext, ok := temptingOfferSearchPair(c.Effects, baseNext)
	if !ok {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the offer clause is not a plain library search")
	}
	// "For each opponent who searches a library this way" reparses its "searches a
	// library this way" quantifier as a bare search (a search not followed by a
	// put). It carries no rules of its own — the accepter count already drives the
	// reward repeat — so it is skipped.
	rewardStart := offerNext
	if rewardStart < len(c.Effects) &&
		c.Effects[rewardStart].Kind == compiler.EffectSearch &&
		(rewardStart+1 >= len(c.Effects) || c.Effects[rewardStart+1].Kind != compiler.EffectPut) {
		rewardStart++
	}
	// Reward clause: the controller searches and puts again (a non-optional pair).
	if rewardStart >= len(c.Effects) || c.Effects[rewardStart].Optional {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the reward clause is not a plain controller library search")
	}
	rewardSearch, rewardNext, ok := temptingOfferSearchPair(c.Effects, rewardStart)
	if !ok {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the reward clause is not a plain library search")
	}
	// Trailing "then each player who searched a library this way shuffles". Each
	// search already shuffles the searched library, so this clause is verified and
	// consumed rather than lowered.
	if rewardNext >= len(c.Effects) || c.Effects[rewardNext].Kind != compiler.EffectShuffle {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "a search Tempting offer must end with a shuffle")
	}
	if rewardNext+1 != len(c.Effects) {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "a search Tempting offer has unexpected trailing effects")
	}
	if !reflect.DeepEqual(baseSearch, offerSearch) || !reflect.DeepEqual(baseSearch, rewardSearch) {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the three library searches are not identical")
	}
	return temptingOfferAbilityContent(baseSearch, nil), nil
}

// temptingOfferSearchPair collapses the search+put pair at index i of the effect
// list into a single game.Search primitive addressed through
// GroupOfferMemberReference(), returning the primitive and the index following the
// pair. It requires effects[i] to be a plain library search and effects[i+1] a
// plain put of the found card onto the battlefield or into the hand, with no
// reveal, split, top-of-library, correlation, or control riders; the search's own
// context and optionality are ignored (the caller checks them) so the base,
// offer, and reward pairs collapse to the same primitive. It fails closed on any
// richer search or put.
func temptingOfferSearchPair(effects []compiler.CompiledEffect, i int) (game.Search, int, bool) {
	if i+1 >= len(effects) {
		return game.Search{}, 0, false
	}
	search, put := effects[i], effects[i+1]
	if search.Kind != compiler.EffectSearch || put.Kind != compiler.EffectPut {
		return game.Search{}, 0, false
	}
	if search.Negated || search.DelayedTiming != 0 || search.Duration != compiler.DurationNone ||
		put.Negated || put.DelayedTiming != 0 || put.Duration != compiler.DurationNone {
		return game.Search{}, 0, false
	}
	// The parser leaves an UnsupportedDetail on each search because the idiom
	// defers its shuffle to a single trailing clause ("then each player who
	// searched a library this way shuffles") rather than printing the byte-exact
	// "search ... then shuffle" wording the generic search path reconstructs. That
	// is a wording mismatch, not a semantic gap: the compiled Selector still fully
	// describes the search, and searchSpecForSelector below fails closed on any
	// selector it cannot represent, so the detail is intentionally not gated here.
	if search.SearchSharedSubtype ||
		search.SearchDifferentNames ||
		search.SearchDestination == parser.EffectDestinationTop ||
		search.SearchControl != parser.SearchControlRiderNone {
		return game.Search{}, 0, false
	}
	quantity, ok := searchAmountQuantity(search)
	if !ok {
		return game.Search{}, 0, false
	}
	spec, ok := searchSpecForSelector(search.Selector)
	if !ok {
		return game.Search{}, 0, false
	}
	spec.SourceZone = zone.Library
	if put.SearchSplit.Present ||
		(put.ToZone != zone.Battlefield && put.ToZone != zone.Hand) {
		return game.Search{}, 0, false
	}
	spec.Destination = put.ToZone
	spec.EntersTapped = put.EntersTapped
	if !quantity.IsDynamic() && search.Amount.Known && search.Amount.Value == 1 && spec.IsUnrestricted() {
		spec.FailToFindPolicy = game.SearchMustFindIfAvailable
	}
	return game.Search{
		Player: game.GroupOfferMemberReference(),
		Spec:   spec,
		Amount: quantity,
	}, i + 2, true
}

// lowerTemptingOfferCompoundBody lowers a Tempting offer whose shared body is a
// multi-primitive sequence (Tempt with Bunnies: "Draw a card and create a 1/1
// white Rabbit creature token."). The body's effects repeat as three identical
// clauses — the controller base, the each-opponent-may offer, and the
// per-accepter reward — so the effect list is exactly three clauses of the same
// length. Each clause lowers to the same instruction sequence, addressed through
// GroupOfferMemberReference(), and the whole body runs atomically per acting
// player through TemptingOfferBody. It fails closed unless the three clauses are
// identical and carry the you / each-opponent-may / you idiom.
func lowerTemptingOfferCompoundBody(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	c := ability.Content
	if len(c.Targets) != 0 || len(c.References) != 0 {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "a compound Tempting offer with targets or references is unsupported")
	}
	n := len(c.Effects)
	if n < 6 || n%3 != 0 {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "a compound Tempting offer requires three equal-length clauses")
	}
	k := n / 3
	baseClause := c.Effects[0:k]
	offerClause := c.Effects[k : 2*k]
	rewardClause := c.Effects[2*k : 3*k]
	if !temptingOfferActingClause(baseClause, parser.EffectContextController) {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the base clause is not a plain controller sequence")
	}
	if !temptingOfferOfferClause(offerClause) {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the offer clause is not an each-opponent-may sequence")
	}
	if !temptingOfferActingClause(rewardClause, parser.EffectContextController) {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the reward clause is not a plain controller sequence")
	}
	parent := contentCtx{
		text:          syntax.Text,
		span:          c.Span,
		content:       c,
		enclosingKind: compiler.AbilitySpell,
	}
	baseBody, diagnostic := lowerTemptingOfferBodyClause(parent, ability, baseClause)
	if diagnostic != nil {
		return game.AbilityContent{}, diagnostic
	}
	offerBody, diagnostic := lowerTemptingOfferBodyClause(parent, ability, offerClause)
	if diagnostic != nil {
		return game.AbilityContent{}, diagnostic
	}
	rewardBody, diagnostic := lowerTemptingOfferBodyClause(parent, ability, rewardClause)
	if diagnostic != nil {
		return game.AbilityContent{}, diagnostic
	}
	if !reflect.DeepEqual(baseBody, offerBody) || !reflect.DeepEqual(baseBody, rewardBody) {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the three clause bodies are not identical")
	}
	return temptingOfferBodyAbilityContent(baseBody, nil), nil
}

// temptingOfferActingClause reports whether a clause is a single acting player's
// sequence: the first effect carries the given acting context and is not optional,
// and every following effect continues the same subject (the given context or a
// prior-subject continuation) without introducing a new offer.
func temptingOfferActingClause(clause []compiler.CompiledEffect, acting parser.EffectContextKind) bool {
	if len(clause) == 0 || clause[0].Context != acting || clause[0].Optional {
		return false
	}
	for i := 1; i < len(clause); i++ {
		if clause[i].Optional ||
			(clause[i].Context != acting && clause[i].Context != parser.EffectContextPriorSubject) {
			return false
		}
	}
	return true
}

// temptingOfferOfferClause reports whether a clause is the each-opponent-may
// offer: its first effect is an optional each-opponent effect and every following
// effect is a non-optional prior-subject continuation of that same opponent.
func temptingOfferOfferClause(clause []compiler.CompiledEffect) bool {
	if len(clause) == 0 ||
		clause[0].Context != parser.EffectContextEachOpponent ||
		!clause[0].Optional {
		return false
	}
	for i := 1; i < len(clause); i++ {
		if clause[i].Optional || clause[i].Context != parser.EffectContextPriorSubject {
			return false
		}
	}
	return true
}

// lowerTemptingOfferBodyClause lowers a clause's effect sequence to the shared
// Tempting-offer body: each effect becomes one instruction addressed to the
// acting player through GroupOfferMemberReference(). The modeled compound effects
// are a plain card draw and a synthesized-token creation (Tempt with Bunnies);
// any other effect fails closed.
func lowerTemptingOfferBodyClause(
	parent contentCtx,
	ability compiler.CompiledAbility,
	clause []compiler.CompiledEffect,
) ([]game.Instruction, *shared.Diagnostic) {
	body := make([]game.Instruction, 0, len(clause))
	for i := range clause {
		instr, diagnostic := lowerTemptingOfferBodyInstruction(parent, ability, clause[i])
		if diagnostic != nil {
			return nil, diagnostic
		}
		body = append(body, instr)
	}
	return body, nil
}

// lowerTemptingOfferBodyInstruction lowers a single compound-body effect to an
// instruction addressed to the acting player. A draw becomes a
// GroupOfferMemberReference()-addressed Draw; a synthesized-token creation reuses
// lowerTemptingOfferToken and enters the token under the acting player. Any other
// effect fails closed.
func lowerTemptingOfferBodyInstruction(
	parent contentCtx,
	ability compiler.CompiledAbility,
	effect compiler.CompiledEffect,
) (game.Instruction, *shared.Diagnostic) {
	switch effect.Kind {
	case compiler.EffectDraw:
		draw, ok := temptingOfferDrawPrimitive(effect)
		if !ok {
			return game.Instruction{}, temptingOfferDiagnostic(ability, "a draw in a compound Tempting offer is unsupported")
		}
		return game.Instruction{Primitive: draw}, nil
	case compiler.EffectCreate:
		token, diagnostic := lowerTemptingOfferToken(parent, ability, effect)
		if diagnostic != nil {
			return game.Instruction{}, diagnostic
		}
		token.Recipient = opt.Val(game.GroupOfferMemberReference())
		return game.Instruction{Primitive: token}, nil
	default:
		return game.Instruction{}, temptingOfferDiagnostic(ability, "an effect in a compound Tempting offer is unsupported")
	}
}

// temptingOfferDrawPrimitive builds the Draw for one clause effect of a compound
// Tempting offer, addressing the acting player through GroupOfferMemberReference().
// It accepts only a plain fixed-count card draw with no delay, duration,
// negation, or reference rider; any richer draw fails closed.
func temptingOfferDrawPrimitive(effect compiler.CompiledEffect) (game.Draw, bool) {
	if effect.Kind != compiler.EffectDraw ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		len(effect.References) != 0 ||
		!effect.Amount.Known ||
		effect.Amount.RangeKnown ||
		effect.Amount.VariableX ||
		effect.Amount.DynamicKind != 0 ||
		effect.Amount.Value <= 0 {
		return game.Draw{}, false
	}
	return game.Draw{
		Player: game.GroupOfferMemberReference(),
		Amount: game.Fixed(effect.Amount.Value),
	}, true
}

// temptingOfferDiagnostic builds the fail-closed diagnostic for a "Tempting
// offer" ability the executable backend cannot yet lower, attributed to the
// ability span so the coverage report points at the whole idiom.
func temptingOfferDiagnostic(ability compiler.CompiledAbility, detail string) *shared.Diagnostic {
	return &shared.Diagnostic{
		Severity: shared.SeverityWarning,
		Summary:  "unsupported Tempting offer",
		Detail:   "the executable source backend cannot yet lower this Tempting offer ability: " + detail,
		Span:     ability.Span,
	}
}
