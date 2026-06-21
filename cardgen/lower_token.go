package cardgen

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerCreateTokenSpell lowers vanilla creature-token creation: the controller,
// a referenced object's controller, or a single targeted player ("Target
// opponent creates ...") creates a fixed-power/toughness creature token with one
// or two subtypes, up to two colors (or colorless), an optional leading
// artifact/enchantment permanent type, an optional single creature keyword, an
// optional tapped entry, an optional attacking entry ("... token that's tapped
// and attacking"), and an optional explicit Oracle name ("... creature token
// named <Name>"). The token count may be a fixed number, the spell's variable X,
// or a recognized rules-derived dynamic count ("for each <X>", "number of …
// equal to <X>", "where X is <X>"). Richer token shapes (a blocking entry,
// quoted abilities, multiple keywords, modifiers) and unrepresentable dynamic
// counts fail closed pending follow-up work under the token-creation epic.
func lowerCreateTokenSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if effect.TokenCopyOfTarget {
		return lowerCreateCopyTokenSpell(ctx)
	}
	if effect.TokenCopyOfReference {
		return lowerCreateCopyTokenReferenceSpell(ctx)
	}
	if effect.TokenCopyOfAttached {
		return lowerCreateCopyTokenAttachedSpell(ctx)
	}
	if effect.TokenCopyOfForEach {
		return lowerCreateCopyTokenForEachSpell(ctx)
	}
	controllerRecipient := effect.Context == parser.EffectContextController
	referencedRecipient := effect.Context == parser.EffectContextReferencedObjectController
	targetRecipient := effect.Context == parser.EffectContextTarget
	extraKeywords, keywordsOK := tokenContentKeywords(ctx.content)
	expectedTargets := 0
	if targetRecipient {
		expectedTargets = 1
	}
	if len(ctx.content.Effects) != 1 ||
		effect.Kind != compiler.EffectCreate ||
		!effect.Exact ||
		(!controllerRecipient && !referencedRecipient && !targetRecipient) ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		!createTokenDurationOK(effect.Duration) ||
		len(ctx.content.Targets) != expectedTargets ||
		len(ctx.content.Conditions) != 0 ||
		!keywordsOK ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	var recipient opt.V[game.PlayerReference]
	var targets []game.TargetSpec
	switch {
	case controllerRecipient:
		if len(ctx.content.References) != 0 {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
	case referencedRecipient:
		if len(ctx.content.References) != 1 {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowTarget: true})
		if !ok {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		recipient = opt.Val(game.ObjectControllerReference(object))
	case targetRecipient:
		if len(ctx.content.References) != 0 {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		spec, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		targets = []game.TargetSpec{spec}
		recipient = opt.Val(game.TargetPlayerReference(0))
	default:
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	if effect.TokenChoice {
		return lowerCreateNamedTokenChoiceSpell(ctx, &effect, recipient, targets)
	}
	def, ok := synthesizeCreatureTokenDef(&effect, extraKeywords)
	if !ok && len(extraKeywords) == 0 {
		def, ok = synthesizeNamedArtifactTokenDef(&effect)
	}
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	amount, ok := createTokenAmount(ctx, &effect)
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount:         amount,
				Source:         game.TokenDef(def),
				Recipient:      recipient,
				EntryTapped:    effect.Selector.Tapped,
				EntryAttacking: effect.Selector.Attacking,
			},
		}},
	}.Ability(), nil
}

// lowerCreateNamedTokenChoiceSpell lowers an N-way (N >= 2) choice among
// predefined artifact tokens ("Create a X token or a Y token." and "Create your
// choice of a X token, a Y token, or a Z token.") to a choose-one modal ability:
// one mode per predefined artifact-token alternative, each creating a single
// token for the shared recipient. Every alternative must be a predefined
// artifact token the runtime already models. The target-recipient form is not
// lowered here because modal content cannot carry per-mode targets; it fails
// closed. Any non-predefined alternative, color, keyword, or count other than
// one also fails closed.
func lowerCreateNamedTokenChoiceSpell(ctx contentCtx, effect *compiler.CompiledEffect, recipient opt.V[game.PlayerReference], targets []game.TargetSpec) (game.AbilityContent, *shared.Diagnostic) {
	subtypes := effect.Selector.SubtypesAny()
	if len(targets) != 0 ||
		len(subtypes) < 2 ||
		len(effect.Selector.ColorsAny()) != 0 ||
		effect.Selector.Keyword != parser.KeywordUnknown ||
		effect.Selector.Tapped ||
		effect.TokenPTKnown {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	amount, ok := createTokenAmount(ctx, effect)
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	modes := make([]game.Mode, 0, len(subtypes))
	for _, sub := range subtypes {
		def, ok := namedArtifactTokenDef(sub)
		if !ok {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		modes = append(modes, game.Mode{
			Text: "Create a " + string(sub) + " token.",
			Sequence: []game.Instruction{{
				Primitive: game.CreateToken{
					Amount:    amount,
					Source:    game.TokenDef(def),
					Recipient: recipient,
				},
			}},
		})
	}
	return game.AbilityContent{
		Modes:    modes,
		MinModes: 1,
		MaxModes: 1,
	}, nil
}

// createTokenDurationOK reports whether a recognized exact create-token effect's
// compiled duration is acceptable. Creating a token is instantaneous and the
// token persists, so a create-token clause never carries its own duration. A
// recognized exact create reconstructs to "Create <spec>." with no duration
// words, which proves any non-None duration is spurious metadata that leaked
// from a sibling clause in the same sentence — for example the "this turn" of an
// intervening "if you attacked this turn" trigger condition. Only that
// turn-scoped leak is tolerated; an "until end of turn"/"until your next turn"
// value cannot leak from such a clause (its "until" wording would break create
// exactness) and so stays fail-closed.
func createTokenDurationOK(duration compiler.DurationKind) bool {
	return duration == compiler.DurationNone || duration == compiler.DurationThisTurn
}

// createTokenAmount resolves a recognized create-token effect's token count. A
// fixed literal lowers to that count; the spell's variable X lowers to the
// runtime X amount; and every recognized rules-derived count ("for each <X>",
// "equal to <X>", "where X is <X>") lowers through the shared dynamic-amount
// lowerer. Source-power counts and any unrepresented dynamic kind fail closed.
func createTokenAmount(ctx contentCtx, effect *compiler.CompiledEffect) (game.Quantity, bool) {
	switch {
	case effect.Amount.Known:
		if effect.Amount.Value < 1 {
			return game.Quantity{}, false
		}
		return game.Fixed(effect.Amount.Value), true
	case effect.Amount.VariableX:
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), true
	case effect.Amount.DynamicKind == compiler.DynamicAmountTriggeringCombatDamage:
		dynamic, ok := lowerEventCombatDamageAmount(ctx, effect.Amount)
		if !ok {
			return game.Quantity{}, false
		}
		return game.Dynamic(dynamic), true
	case effect.Amount.DynamicKind != compiler.DynamicAmountNone:
		if effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower {
			return game.Quantity{}, false
		}
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.ObjectReference{})
		if !ok {
			return game.Quantity{}, false
		}
		return game.Dynamic(dynamic), true
	default:
		return game.Quantity{}, false
	}
}

// lowerCreateCopyTokenSpell lowers "Create a token that's a copy of <target>[,
// except <it> isn't legendary][. That token gains <keyword>.]" to a CreateToken
// whose source copies the lone target object, applying any copy modifiers. The
// runtime already supports object-copy tokens (TokenCopySourceObject); only a
// single fixed target, controller recipient, and supported copy modifiers are
// accepted here.
func lowerCreateCopyTokenSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Effects) != 1 ||
		effect.Context != parser.EffectContextController ||
		!effect.Exact ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		len(ctx.content.Targets) != 1 ||
		!tokenCopyAuxiliaryReferencesOK(ctx.content.References) ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != len(effect.TokenCopyGrantKeywords) ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	spec, ok := tokenCopyModifiers(&effect, game.TargetPermanentReference(0))
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount: game.Fixed(1),
				Source: game.TokenCopyOf(spec),
			},
		}},
	}.Ability(), nil
}

// lowerCreateCopyTokenReferenceSpell lowers "Create a token that's a copy of
// <reference>[, except <it> isn't legendary][. That token gains <keyword>.]"
// (e.g. "... a copy of this creature") to a CreateToken whose source copies the
// object named by the effect's leading explicit reference. The leading reference
// binds the source permanent; any trailing references are the "except" / "that
// token" pronouns of the recognized copy modifiers. Only a controller recipient
// with supported copy modifiers is accepted here.
func lowerCreateCopyTokenReferenceSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Effects) != 1 ||
		effect.Context != parser.EffectContextController ||
		!effect.Exact ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) < 1 ||
		!tokenCopyAuxiliaryReferencesOK(ctx.content.References[1:]) ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != len(effect.TokenCopyGrantKeywords) ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	object, ok := lowerObjectReference(
		ctx.content.References[0],
		referenceLoweringContext{AllowSource: true, AllowTarget: true, AllowEvent: true},
	)
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	spec, ok := tokenCopyModifiers(&effect, object)
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount: game.Fixed(1),
				Source: game.TokenCopyOf(spec),
			},
		}},
	}.Ability(), nil
}

// lowerCreateCopyTokenAttachedSpell lowers "Create a token that's a copy of
// equipped/enchanted creature[, except <it> isn't legendary][. That token gains
// <keyword>.]" to a CreateToken whose source copies the permanent the source
// Equipment or Aura is attached to. The runtime resolves the attached permanent
// at resolution; only a controller recipient with supported copy modifiers is
// accepted here.
func lowerCreateCopyTokenAttachedSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Effects) != 1 ||
		effect.Context != parser.EffectContextController ||
		!effect.Exact ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		len(ctx.content.Targets) != 0 ||
		!tokenCopyAuxiliaryReferencesOK(ctx.content.References) ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != len(effect.TokenCopyGrantKeywords) ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	spec, ok := tokenCopyModifiers(&effect, game.SourceAttachedPermanentReference())
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount: game.Fixed(1),
				Source: game.TokenCopyOf(spec),
			},
		}},
	}.Ability(), nil
}

// lowerCreateCopyTokenForEachSpell lowers a per-each copy-token create whose
// copy source is each member of a controlled battlefield group ("For each token
// you control, create a token that's a copy of that permanent." — Second
// Harvest) to a CreateToken whose source iterates the group, copying each
// matched permanent in turn. The "that permanent" reference is a benign
// per-iteration pronoun the runtime resolves member-by-member; only a controller
// recipient over a controlled group with supported copy modifiers is accepted.
func lowerCreateCopyTokenForEachSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Effects) != 1 ||
		effect.Context != parser.EffectContextController ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		len(ctx.content.Targets) != 0 ||
		!tokenCopyAuxiliaryReferencesOK(ctx.content.References) ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != len(effect.TokenCopyGrantKeywords) ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	selection, ok := copyForEachGroupSelection(effect.TokenCopyForEachGroup)
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	spec, ok := tokenCopyForEachModifiers(&effect, game.BattlefieldGroup(selection))
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount: game.Fixed(1),
				Source: game.TokenCopyOf(spec),
			},
		}},
	}.Ability(), nil
}

// copyForEachGroupSelection lowers the controlled battlefield group iterated by a
// per-each copy-token create to a runtime Selection. It requires a "you control"
// controller scope and reuses massGroupSelection for the type/color/keyword
// filters. A bare "token you control" carries no card type, so its token filter
// is the constraint: synthesize a permanent kind so massGroupSelection accepts
// it, then restore the token/nontoken filter the shared lowering does not model.
func copyForEachGroupSelection(selector compiler.CompiledSelector) (game.Selection, bool) {
	if selector.Controller != compiler.ControllerYou {
		return game.Selection{}, false
	}
	adjusted := selector
	if selector.Kind == compiler.SelectorUnknown && (selector.TokenOnly || selector.NonToken) {
		adjusted.Kind = compiler.SelectorPermanent
	}
	selection, ok := massGroupSelection(adjusted)
	if !ok {
		return game.Selection{}, false
	}
	selection.TokenOnly = selector.TokenOnly
	selection.NonToken = selector.NonToken
	return selection, true
}

// tokenCopyForEachModifiers builds the runtime copy spec for a per-each copy
// over the iterated group, applying the "except <it> isn't legendary" supertype
// drop and any folded "that token gains <keyword>" rider keywords. It fails
// closed when a granted keyword has no reusable runtime static form.
func tokenCopyForEachModifiers(effect *compiler.CompiledEffect, group game.GroupReference) (game.TokenCopySpec, bool) {
	spec := game.TokenCopySpec{
		Source:          game.TokenCopySourceEachInGroup,
		Group:           game.GroupRef(group),
		SetNotLegendary: effect.TokenCopyDropLegendary,
	}
	for _, kind := range effect.TokenCopyGrantKeywords {
		keyword, ok := runtimeKeyword(kind)
		if !ok {
			return game.TokenCopySpec{}, false
		}
		spec.AddKeywords = append(spec.AddKeywords, keyword)
	}
	return spec, true
}

// tokenCopyModifiers builds the runtime copy spec for a copy-token effect over
// the given source object, applying the "except <it> isn't legendary" supertype
// drop and any folded "that token gains <keyword>" rider keywords. It fails
// closed when a granted keyword has no reusable runtime static form.
func tokenCopyModifiers(effect *compiler.CompiledEffect, object game.ObjectReference) (game.TokenCopySpec, bool) {
	spec := game.TokenCopySpec{
		Source:          game.TokenCopySourceObject,
		Object:          object,
		SetNotLegendary: effect.TokenCopyDropLegendary,
	}
	for _, kind := range effect.TokenCopyGrantKeywords {
		keyword, ok := runtimeKeyword(kind)
		if !ok {
			return game.TokenCopySpec{}, false
		}
		spec.AddKeywords = append(spec.AddKeywords, keyword)
	}
	return spec, true
}

// tokenCopyAuxiliaryReferencesOK reports whether every reference is a benign
// pronoun introduced by a copy modifier ("except it isn't legendary", "that
// token gains ..."), so the copy-token lowering can tolerate them without
// treating them as additional copy sources.
func tokenCopyAuxiliaryReferencesOK(references []compiler.CompiledReference) bool {
	for i := range references {
		switch references[i].Kind {
		case compiler.ReferencePronoun, compiler.ReferenceThatObject:
		default:
			return false
		}
	}
	return true
}

// synthesizeCreatureTokenDef builds a token CardDef from a recognized create
// effect: a creature with one or two subtypes, up to two colors (or colorless),
// an optional leading artifact/enchantment permanent type, a fixed
// power/toughness, and zero or more creature keywords. The leading "with
// <keyword>" selector keyword is carried on the effect selector; any additional
// conjoined keywords ("... and reach") arrive in extraKeywords. Each keyword
// becomes one static ability in Oracle order. The token's name is its explicit
// Oracle name when one is printed ("... token named <Name>"); otherwise it is
// the joined subtypes, matching paper tokens.
func synthesizeCreatureTokenDef(effect *compiler.CompiledEffect, extraKeywords []parser.KeywordKind) (*game.CardDef, bool) {
	if !effect.TokenPTKnown {
		return nil, false
	}
	subtypes := effect.Selector.SubtypesAny()
	if len(subtypes) < 1 || len(subtypes) > 2 {
		return nil, false
	}
	colors := effect.Selector.ColorsAny()
	if len(colors) > 2 {
		return nil, false
	}
	cardTypes, ok := creatureTokenCardTypes(effect.Selector)
	if !ok {
		return nil, false
	}
	names := make([]string, 0, len(subtypes))
	for _, sub := range subtypes {
		names = append(names, string(sub))
	}
	name := strings.Join(names, " ")
	if effect.TokenName != "" {
		name = effect.TokenName
	}
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name:      name,
			Colors:    slices.Clone(colors),
			Types:     cardTypes,
			Subtypes:  slices.Clone(subtypes),
			Power:     opt.Val(game.PT{Value: effect.TokenPower}),
			Toughness: opt.Val(game.PT{Value: effect.TokenToughness}),
		},
	}
	keywords := make([]parser.KeywordKind, 0, 1+len(extraKeywords))
	if effect.Selector.Keyword != parser.KeywordUnknown {
		keywords = append(keywords, effect.Selector.Keyword)
	}
	keywords = append(keywords, extraKeywords...)
	for _, keyword := range keywords {
		static, ok := keywordStaticBodies[keyword]
		if !ok {
			return nil, false
		}
		def.StaticAbilities = append(def.StaticAbilities, static.Body)
	}
	return def, true
}

// creatureTokenCardTypes returns the card types for a synthesized creature
// token. A bare creature token compiles to [Creature]; an artifact- or
// enchantment-creature token prepends the additional permanent type, matching
// the Oracle "<type> creature" ordering. Any other required-type set fails
// closed.
func creatureTokenCardTypes(selector compiler.CompiledSelector) ([]types.Card, bool) {
	required := selector.RequiredTypesAny()
	if len(required) == 0 {
		return []types.Card{types.Creature}, true
	}
	hasCreature := false
	var extra types.Card
	for _, cardType := range required {
		switch cardType {
		case types.Creature:
			hasCreature = true
		case types.Artifact, types.Enchantment:
			if extra != "" {
				return nil, false
			}
			extra = cardType
		default:
			return nil, false
		}
	}
	if !hasCreature {
		return nil, false
	}
	if extra == "" {
		return []types.Card{types.Creature}, true
	}
	return []types.Card{extra, types.Creature}, true
}

// synthesizeNamedArtifactTokenDef builds a token CardDef for a predefined
// artifact token with no printed power/toughness (Treasure, Food, Clue, Blood,
// Gold, Lander, Mutagen). These tokens carry a fixed Oracle ability the runtime
// CreateToken/TokenDef model already represents. Any other named token fails
// closed.
func synthesizeNamedArtifactTokenDef(effect *compiler.CompiledEffect) (*game.CardDef, bool) {
	if effect.TokenPTKnown {
		return nil, false
	}
	subtypes := effect.Selector.SubtypesAny()
	if len(subtypes) != 1 ||
		len(effect.Selector.ColorsAny()) != 0 ||
		effect.Selector.Keyword != parser.KeywordUnknown {
		return nil, false
	}
	return namedArtifactTokenDef(subtypes[0])
}

// namedArtifactTokenDef returns the synthesized CardDef for a recognized
// predefined artifact token, or false for an unrepresented one.
func namedArtifactTokenDef(sub types.Sub) (*game.CardDef, bool) {
	switch sub {
	case types.Treasure:
		return treasureTokenDef(), true
	case types.Food:
		return foodTokenDef(), true
	case types.Clue:
		return clueTokenDef(), true
	case types.Blood:
		return bloodTokenDef(), true
	case types.Gold:
		return goldTokenDef(), true
	case types.Lander:
		return landerTokenDef(), true
	case types.Mutagen:
		return mutagenTokenDef(), true
	default:
		return nil, false
	}
}

// sacrificeArtifactCost is the "Sacrifice this artifact" additional cost shared
// by predefined artifact tokens.
func sacrificeArtifactCost() cost.Additional {
	return cost.Additional{
		Kind:               cost.AdditionalSacrificeSource,
		Text:               "Sacrifice this artifact",
		Amount:             1,
		MatchPermanentType: true,
		PermanentType:      types.Artifact,
	}
}

// artifactTokenDef builds a single-ability colorless artifact token CardDef
// named for its subtype, matching paper predefined tokens.
func artifactTokenDef(sub types.Sub, ability *game.ActivatedAbility) *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:               string(sub),
			Types:              []types.Card{types.Artifact},
			Subtypes:           []types.Sub{sub},
			ActivatedAbilities: []game.ActivatedAbility{*ability},
		},
	}
}

// treasureTokenDef builds the Treasure token: tap and sacrifice it to add one
// mana of any color.
func treasureTokenDef() *game.CardDef {
	ability := game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)
	ability.Text = "{T}, Sacrifice this artifact: Add one mana of any color."
	ability.AdditionalCosts = append(slices.Clone(ability.AdditionalCosts), sacrificeArtifactCost())
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:          string(types.Treasure),
			Types:         []types.Card{types.Artifact},
			Subtypes:      []types.Sub{types.Treasure},
			ManaAbilities: []game.ManaAbility{ability},
		},
	}
}

// foodTokenDef builds the Food token: pay two generic mana, tap, and sacrifice
// it to gain 3 life.
func foodTokenDef() *game.CardDef {
	return artifactTokenDef(types.Food, &game.ActivatedAbility{
		Text:            "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
		ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
		AdditionalCosts: []cost.Additional{cost.T, sacrificeArtifactCost()},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.GainLife{Amount: game.Fixed(3), Player: game.ControllerReference()}},
		}}.Ability(),
	})
}

// clueTokenDef builds the Clue token: pay two generic mana and sacrifice it to
// draw a card.
func clueTokenDef() *game.CardDef {
	return artifactTokenDef(types.Clue, &game.ActivatedAbility{
		Text:            "{2}, Sacrifice this artifact: Draw a card.",
		ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
		AdditionalCosts: []cost.Additional{sacrificeArtifactCost()},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}},
		}}.Ability(),
	})
}

// bloodTokenDef builds the Blood token: pay one generic mana, tap, discard a
// card, and sacrifice it to draw a card.
func bloodTokenDef() *game.CardDef {
	return artifactTokenDef(types.Blood, &game.ActivatedAbility{
		Text:     "{1}, {T}, Discard a card, Sacrifice this artifact: Draw a card.",
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		AdditionalCosts: []cost.Additional{
			cost.T,
			{Kind: cost.AdditionalDiscard, Text: "Discard a card", Amount: 1, Source: zone.Hand},
			sacrificeArtifactCost(),
		},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}},
		}}.Ability(),
	})
}

// sacrificeTokenCost is the "Sacrifice this token" additional cost carried by
// the predefined tokens (Gold, Lander, Mutagen) whose printed Oracle text names
// the token rather than its artifact type.
func sacrificeTokenCost() cost.Additional {
	return cost.Additional{
		Kind:   cost.AdditionalSacrificeSource,
		Text:   "Sacrifice this token",
		Amount: 1,
	}
}

// goldTokenDef builds the Gold token: sacrifice it to add one mana of any color.
func goldTokenDef() *game.CardDef {
	choice := game.ChoiceKey("oracle-mana-color")
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     string(types.Gold),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Gold},
			ManaAbilities: []game.ManaAbility{{
				Text:            "Sacrifice this token: Add one mana of any color.",
				AdditionalCosts: []cost.Additional{sacrificeTokenCost()},
				Content: game.Mode{Sequence: []game.Instruction{
					{Primitive: game.Choose{
						Choice: game.ResolutionChoice{
							Kind:   game.ResolutionChoiceMana,
							Prompt: "Choose a color",
							Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
						},
						PublishChoice: choice,
					}},
					{Primitive: game.AddMana{Amount: game.Fixed(1), ChoiceFrom: choice}},
				}}.Ability(),
			}},
		},
	}
}

// landerTokenDef builds the Lander token: pay two generic mana, tap, and
// sacrifice it to search your library for a basic land and put it onto the
// battlefield tapped.
func landerTokenDef() *game.CardDef {
	return artifactTokenDef(types.Lander, &game.ActivatedAbility{
		Text:            "{2}, {T}, Sacrifice this token: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
		AdditionalCosts: []cost.Additional{cost.T, sacrificeTokenCost()},
		ZoneOfFunction:  zone.Battlefield,
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Search{
				Player: game.ControllerReference(),
				Spec: game.SearchSpec{
					SourceZone:   zone.Library,
					Destination:  zone.Battlefield,
					CardType:     opt.Val(types.Land),
					Supertype:    opt.Val(types.Basic),
					EntersTapped: true,
				},
				Amount: game.Fixed(1),
			}},
		}}.Ability(),
	})
}

// mutagenTokenDef builds the Mutagen token: pay one generic mana, tap, and
// sacrifice it to put a +1/+1 counter on target creature, at sorcery speed.
func mutagenTokenDef() *game.CardDef {
	return artifactTokenDef(types.Mutagen, &game.ActivatedAbility{
		Text:            "{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target creature. Activate only as a sorcery.",
		ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
		AdditionalCosts: []cost.Additional{cost.T, sacrificeTokenCost()},
		ZoneOfFunction:  zone.Battlefield,
		Timing:          game.SorceryOnly,
		Content: game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature",
				Allow:      game.TargetAllowPermanent,
				Predicate:  game.TargetPredicate{PermanentTypes: []types.Card{types.Creature}},
			}},
			Sequence: []game.Instruction{
				{Primitive: game.AddCounter{
					Amount:      game.Fixed(1),
					Object:      game.TargetPermanentReference(0),
					CounterKind: counter.PlusOnePlusOne,
				}},
			},
		}.Ability(),
	})
}

// tokenContentKeywords returns the kinds of the conjoined bare keywords in a
// create-token ability ("... and reach" -> [Reach]); these are the rider
// keywords beyond the leading "with <keyword>" selector keyword. It reports
// false when any compiled ability keyword carries a parameter, since a
// parameterized keyword is not a plain creature-token rider and must fail closed.
func tokenContentKeywords(content compiler.AbilityContent) ([]parser.KeywordKind, bool) {
	kinds := make([]parser.KeywordKind, 0, len(content.Keywords))
	for _, keyword := range content.Keywords {
		if keyword.ParameterKind != parser.KeywordParameterNone {
			return nil, false
		}
		kinds = append(kinds, keyword.Kind)
	}
	return kinds, true
}

func unsupportedTokenCreationDiagnostic(ctx contentCtx) *shared.Diagnostic {
	return contentDiagnostic(
		ctx,
		"unsupported token creation",
		"the executable source backend supports only a single fixed-power/toughness creature token with one subtype and at most one color",
	)
}
