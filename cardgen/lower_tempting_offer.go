package cardgen

import (
	"reflect"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
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
// lowering. Only the token-creation effect (Tempt with Vengeance) is modeled
// today; every other kind and any richer shape fails closed pending follow-up
// work under the Tempt-cycle generalization.
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
	default:
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "the shared effect is not token creation")
	}
}

// lowerTemptingOfferCreate lowers a Tempting-offer whose shared effect creates
// creature tokens (Tempt with Vengeance). Each of the three effects is lowered
// independently as a plain controller-recipient token creation and the resulting
// CreateToken primitives are required to be identical, proving all three create
// the same tokens; the offer then addresses the acting player through
// GroupOfferMemberReference() so the runtime enters the tokens under the
// controller (base and reward) or the accepting opponent (their own copy). Token
// shapes that carry a target or a reference — a copy of another creature (Tempt
// with Reflections) — fail closed, because their recipient rebinding is not yet
// modeled.
func lowerTemptingOfferCreate(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	if len(ability.Content.Targets) != 0 || len(ability.Content.References) != 0 {
		return game.AbilityContent{}, temptingOfferDiagnostic(ability, "token creation with a target or reference is unsupported")
	}
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
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive:          primitive,
			Optional:           true,
			OptionalActorGroup: opt.Val(game.OpponentsReference()),
			TemptingOffer:      true,
		}},
	}.Ability(), nil
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
