package cardgen

import (
	"reflect"

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
const temptingOfferAbilityWord = "Tempting offer"

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
	if ability.Kind != compiler.AbilitySpell || ability.AbilityWord != temptingOfferAbilityWord {
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
// or fails closed with a diagnostic. It requires exactly three effects of the
// same kind with the you / each-opponent-may / you contexts, no modes,
// conditions, or keywords, and dispatches the shared effect to a kind-specific
// lowering. The modeled shared effects are token creation (Tempt with Vengeance),
// a copy of a single targeted creature (Tempt with Reflections), a +1/+1 counter
// placement on each creature the acting player controls (Tempt with Glory), and a
// creature reanimation from the acting player's graveyard (Tempt with
// Immortality). Every other kind and any richer shape fails closed pending
// follow-up work under the Tempt-cycle generalization.
func lowerTemptingOfferContent(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	c := ability.Content
	if len(c.Effects) != 3 ||
		len(c.Modes) != 0 ||
		len(c.Conditions) != 0 ||
		len(c.Keywords) != 0 {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "expected exactly three same-kind effects")
	}
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
