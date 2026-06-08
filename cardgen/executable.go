package cardgen

import (
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle"
)

// GenerateExecutableCardSource generates a CardDef only when every Oracle-text
// ability can be lowered completely by the current executable source backend.
// Unsupported cards return diagnostics and an empty source string.
func GenerateExecutableCardSource(
	card *ScryfallCard,
	pkgName string,
) (string, []oracle.Diagnostic, error) {
	if !supportedLayouts[card.Layout] {
		return "", []oracle.Diagnostic{{
			Severity: oracle.SeverityWarning,
			Summary:  "unsupported card layout",
			Detail:   fmt.Sprintf("the source generator does not support Scryfall layout %q", card.Layout),
		}}, nil
	}
	fields := executableFaces(card)
	abilityFields := make([][]string, len(fields))
	var diagnostics []oracle.Diagnostic
	for i, face := range fields {
		generated, faceDiagnostics := executableAbilityFields(face)
		diagnostics = append(diagnostics, faceDiagnostics...)
		abilityFields[i] = generated
	}
	if len(diagnostics) > 0 {
		return "", diagnostics, nil
	}
	generated := &generatedAbilityFields{}
	if card.Layout == "reversible_card" && len(card.CardFaces) > 0 {
		generated.faces = abilityFields
	} else {
		generated.root = abilityFields[0]
		generated.faces = abilityFields[1:]
	}
	source, err := genCardSource(card, pkgName, generated)
	if err != nil {
		return "", nil, err
	}
	return source, nil, nil
}

func executableFaces(card *ScryfallCard) []generatedCardFields {
	if card.Layout == "reversible_card" && len(card.CardFaces) > 0 {
		return facesFromAllCardFaces(card)
	}
	return append([]generatedCardFields{rootFields(card)}, generatedFaces(card)...)
}

func executableAbilityFields(
	face generatedCardFields,
) ([]string, []oracle.Diagnostic) {
	parsedType := ParseTypeLine(face.TypeLine)
	if len(parsedType.Types) == 0 {
		return nil, []oracle.Diagnostic{{
			Severity: oracle.SeverityWarning,
			Summary:  "unsupported type line",
			Detail:   fmt.Sprintf("type line %q has no supported card type", face.TypeLine),
		}}
	}
	if face.OracleText == "" {
		return nil, nil
	}
	compilation, diagnostics := oracle.Compile(face.OracleText, oracle.ParseContext{
		CardName:         face.Name,
		InstantOrSorcery: slices.Contains(parsedType.Types, "Instant") || slices.Contains(parsedType.Types, "Sorcery"),
		Planeswalker:     slices.Contains(parsedType.Types, "Planeswalker"),
	})
	if len(diagnostics) > 0 {
		return nil, diagnostics
	}

	var staticBodies []string
	var manaAbilities []string
	var triggeredAbilities []string
	var spellAbility string
	var unsupported []oracle.Diagnostic
	for i, ability := range compilation.Abilities {
		syntax := compilation.Syntax.Abilities[i]
		lowered, diagnostic := lowerExecutableAbility(face.Name, ability, syntax)
		if diagnostic != nil {
			unsupported = append(unsupported, *diagnostic)
			continue
		}
		if !lowered.complete(ability, syntax) {
			unsupported = append(unsupported, *executableDiagnostic(
				ability,
				"incomplete executable lowering",
				"the executable source backend did not consume every semantic element and source token",
			))
			continue
		}
		staticBodies = append(staticBodies, lowered.staticBodies...)
		if lowered.manaAbility != "" {
			manaAbilities = append(manaAbilities, lowered.manaAbility)
		}
		if lowered.triggeredAbility != "" {
			triggeredAbilities = append(triggeredAbilities, lowered.triggeredAbility)
		}
		if lowered.spellAbility != "" {
			if spellAbility != "" {
				unsupported = append(unsupported, *executableDiagnostic(
					ability,
					"unsupported multiple spell abilities",
					"the executable source backend supports only one spell ability per card face",
				))
				continue
			}
			spellAbility = lowered.spellAbility
		}
	}
	if len(unsupported) > 0 {
		return nil, append(diagnostics, unsupported...)
	}
	var fields []string
	if len(staticBodies) > 0 {
		fields = append(fields, staticAbilityField(staticBodies))
	}
	if len(manaAbilities) > 0 {
		fields = append(fields, manaAbilityField(manaAbilities))
	}
	if len(triggeredAbilities) > 0 {
		fields = append(fields, triggeredAbilityField(triggeredAbilities))
	}
	if spellAbility != "" {
		fields = append(fields, "SpellAbility: opt.Val("+spellAbility+"),")
	}
	if len(fields) == 0 {
		return nil, diagnostics
	}
	return fields, diagnostics
}

type abilityLowering struct {
	staticBodies     []string
	manaAbility      string
	triggeredAbility string
	spellAbility     string
	consumed         semanticConsumption
	sourceSpans      []oracle.Span
}

type semanticConsumption struct {
	cost       bool
	trigger    bool
	modes      int
	targets    int
	conditions int
	effects    int
	keywords   int
	references int
}

func lowerExecutableAbility(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (abilityLowering, *oracle.Diagnostic) {
	if len(ability.Modes) > 0 {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported modal ability",
			"the executable source backend does not yet lower modal abilities",
		)
	}
	switch ability.Kind {
	case oracle.AbilityStatic:
		bodies, diagnostic := executableKeywordAbility(ability, syntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		spans := make([]oracle.Span, 0, len(ability.Keywords)+len(syntax.Reminders))
		for _, keyword := range ability.Keywords {
			spans = append(spans, keyword.Span)
		}
		for _, reminder := range syntax.Reminders {
			spans = append(spans, reminder.Span)
		}
		return abilityLowering{
			staticBodies: bodies,
			consumed: semanticConsumption{
				keywords: len(ability.Keywords),
			},
			sourceSpans: spans,
		}, nil
	case oracle.AbilityActivated:
		manaAbility, diagnostic := executableTapManaAbility(ability, syntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		return abilityLowering{
			manaAbility: manaAbility,
			consumed: semanticConsumption{
				cost:    true,
				effects: 1,
			},
			sourceSpans: []oracle.Span{ability.Cost.Span, ability.Effects[0].Span},
		}, nil
	case oracle.AbilitySpell:
		spellAbility, diagnostic := executableSpell(cardName, ability, syntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		spans := []oracle.Span{ability.Effects[0].Span}
		for _, target := range ability.Targets {
			spans = append(spans, target.Span)
		}
		for _, reference := range ability.References {
			spans = append(spans, reference.Span)
		}
		return abilityLowering{
			spellAbility: spellAbility,
			consumed: semanticConsumption{
				targets:    len(ability.Targets),
				effects:    1,
				references: len(ability.References),
			},
			sourceSpans: spans,
		}, nil
	case oracle.AbilityTriggered:
		triggeredAbility, diagnostic := executableEnterTrigger(cardName, ability, syntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		spans := []oracle.Span{ability.Trigger.Span}
		for _, effect := range ability.Effects {
			spans = append(spans, effect.Span)
		}
		for _, target := range ability.Targets {
			spans = append(spans, target.Span)
		}
		for _, reference := range ability.References {
			spans = append(spans, reference.Span)
		}
		return abilityLowering{
			triggeredAbility: triggeredAbility,
			consumed: semanticConsumption{
				trigger:    true,
				targets:    len(ability.Targets),
				effects:    len(ability.Effects),
				references: len(ability.References),
			},
			sourceSpans: spans,
		}, nil
	default:
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported "+ability.Kind.String()+" ability",
			"the executable source backend does not yet lower "+ability.Kind.String()+" abilities",
		)
	}
}

func (lowering abilityLowering) complete(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) bool {
	if lowering.consumed.cost != (ability.Cost != nil) ||
		lowering.consumed.trigger != (ability.Trigger != nil) ||
		lowering.consumed.modes != len(ability.Modes) ||
		lowering.consumed.targets != len(ability.Targets) ||
		lowering.consumed.conditions != len(ability.Conditions) ||
		lowering.consumed.effects != len(ability.Effects) ||
		lowering.consumed.keywords != len(ability.Keywords) ||
		lowering.consumed.references != len(ability.References) {
		return false
	}
	for _, token := range syntax.Tokens {
		if token.Kind == oracle.Comma ||
			token.Kind == oracle.Colon ||
			token.Kind == oracle.Period ||
			spanCovered(token.Span, lowering.sourceSpans) {
			continue
		}
		return false
	}
	return true
}

func executableEnterTrigger(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (string, *oracle.Diagnostic) {
	if ability.Trigger == nil ||
		ability.Trigger.Kind != oracle.TriggerWhen ||
		(ability.Trigger.Event != "this creature enters" &&
			ability.Trigger.Event != "this permanent enters") ||
		ability.Trigger.Condition != nil ||
		len(ability.Effects) != 1 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		ability.AbilityWord != "" {
		return "", executableDiagnostic(
			ability,
			"unsupported enter trigger",
			"the executable source backend supports only exact self-enter triggers with one supported effect",
		)
	}
	comma := slices.IndexFunc(syntax.Tokens, func(token oracle.Token) bool {
		return token.Kind == oracle.Comma
	})
	if comma < 0 || comma+1 >= len(syntax.Tokens) {
		return "", executableDiagnostic(
			ability,
			"unsupported enter trigger",
			"the executable source backend supports only exact self-enter triggers with one supported effect",
		)
	}
	body := ability
	body.Kind = oracle.AbilitySpell
	body.Span = ability.Effects[0].Span
	body.Text = titleFirst(ability.Effects[0].Text)
	body.Trigger = nil
	body.References = bodyReferences(ability.References, ability.Trigger.Span)
	bodySyntax := syntax
	bodySyntax.Kind = oracle.AbilitySpell
	bodySyntax.Tokens = syntax.Tokens[comma+1:]
	bodySyntax.Span = body.Span
	bodySyntax.Text = body.Text
	content, diagnostic := executableSpell(cardName, body, bodySyntax)
	if diagnostic != nil {
		return "", executableDiagnostic(
			ability,
			"unsupported enter trigger effect",
			diagnostic.Detail,
		)
	}
	return fmt.Sprintf(`{
	Text: %s,
	Trigger: game.TriggerCondition{
		Type: game.TriggerWhen,
		Pattern: game.TriggerPattern{
			Event: game.EventPermanentEnteredBattlefield,
			Source: game.TriggerSourceSelf,
		},
	},
	Content: %s,
}`, rawStringLiteral(ability.Text), content), nil
}

func bodyReferences(
	references []oracle.CompiledReference,
	triggerSpan oracle.Span,
) []oracle.CompiledReference {
	var body []oracle.CompiledReference
	for _, reference := range references {
		if spanCovered(reference.Span, []oracle.Span{triggerSpan}) {
			continue
		}
		body = append(body, reference)
	}
	return body
}

func spanCovered(span oracle.Span, covering []oracle.Span) bool {
	for _, candidate := range covering {
		if candidate.Start.Offset <= span.Start.Offset &&
			candidate.End.Offset >= span.End.Offset {
			return true
		}
	}
	return false
}

func executableKeywordAbility(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) ([]string, *oracle.Diagnostic) {
	if len(ability.Modes) > 0 {
		return nil, executableDiagnostic(
			ability,
			"unsupported modal ability",
			"the executable source backend does not yet lower modal abilities",
		)
	}
	if ability.AbilityWord != "" {
		return nil, executableDiagnostic(
			ability,
			"unsupported ability word",
			fmt.Sprintf("the executable source backend does not yet lower the %q ability word", ability.AbilityWord),
		)
	}
	if len(ability.Keywords) == 0 {
		return nil, executableDiagnostic(
			ability,
			"unsupported static ability",
			"the executable source backend does not yet lower non-keyword static rules text",
		)
	}
	bodies := make([]string, 0, len(ability.Keywords))
	for _, keyword := range ability.Keywords {
		if keyword.Parameter != "" {
			return nil, executableDiagnostic(
				ability,
				"unsupported parameterized keyword",
				fmt.Sprintf(
					"the executable source backend does not yet lower %s with parameter %q",
					keyword.Name,
					keyword.Parameter,
				),
			)
		}
		body, ok := keywordStaticBodies[keyword.Name]
		if !ok {
			return nil, executableDiagnostic(
				ability,
				"unsupported keyword ability",
				fmt.Sprintf(
					"the executable source backend has no reusable game template for %s",
					keyword.Name,
				),
			)
		}
		bodies = append(bodies, body)
	}
	if len(ability.Targets) > 0 ||
		len(ability.Conditions) > 0 ||
		len(ability.Effects) > 0 ||
		len(ability.References) > 0 {
		return nil, mixedKeywordDiagnostic(ability)
	}
	for _, token := range syntax.Tokens {
		if token.Kind == oracle.Comma ||
			spanCoveredByKeyword(token.Span, ability.Keywords) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return nil, mixedKeywordDiagnostic(ability)
	}
	return bodies, nil
}

func executableTapManaAbility(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (string, *oracle.Diagnostic) {
	if ability.Cost == nil ||
		len(ability.Cost.Components) != 1 ||
		ability.Cost.Components[0].Kind != oracle.CostTap ||
		len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != oracle.EffectAddMana ||
		!ability.Effects[0].Amount.Known ||
		ability.Effects[0].Amount.Value != 1 ||
		ability.Effects[0].Negated ||
		len(ability.Modes) != 0 ||
		ability.AbilityWord != "" ||
		len(ability.Keywords) != 0 ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.References) != 0 ||
		!exactTapManaSyntax(syntax.Tokens) {
		return "", executableDiagnostic(
			ability,
			"unsupported activated ability",
			"the executable source backend supports only exact single-color tap mana abilities",
		)
	}
	color, ok := manaColorName(ability.Effects[0].Symbol)
	if !ok {
		return "", executableDiagnostic(
			ability,
			"unsupported mana symbol",
			fmt.Sprintf("the executable source backend cannot emit mana symbol %q", ability.Effects[0].Symbol),
		)
	}
	return fmt.Sprintf(`{
	Text: %s,
	AdditionalCosts: cost.Tap,
	Content: game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: game.AddMana{
					Amount: game.Fixed(1),
					ManaColor: mana.%s,
				},
			},
		},
	}.Ability(),
}`, rawStringLiteral(ability.Text), color), nil
}

func exactTapManaSyntax(tokens []oracle.Token) bool {
	return len(tokens) == 5 &&
		tokens[0].Kind == oracle.Symbol &&
		strings.EqualFold(tokens[0].Text, "{T}") &&
		tokens[1].Kind == oracle.Colon &&
		tokens[2].Kind == oracle.Word &&
		strings.EqualFold(tokens[2].Text, "Add") &&
		tokens[3].Kind == oracle.Symbol &&
		tokens[4].Kind == oracle.Period
}

func executableFixedDamageSpell(
	cardName string,
	ability oracle.CompiledAbility,
	_ oracle.Ability,
) (string, *oracle.Diagnostic) {
	if len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != oracle.EffectDealDamage ||
		!ability.Effects[0].Amount.Known ||
		ability.Effects[0].Amount.Value < 1 ||
		ability.Effects[0].Negated ||
		len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		!singleSelfReference(ability.References) {
		return "", executableDiagnostic(
			ability,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed damage to one target",
		)
	}
	targetSource, ok := damageTargetSource(ability.Targets[0])
	if !ok ||
		ability.Text != fmt.Sprintf(
			"%s deals %d damage to %s.",
			cardName,
			ability.Effects[0].Amount.Value,
			ability.Targets[0].Text,
		) {
		return "", executableDiagnostic(
			ability,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed damage to one target",
		)
	}
	return fmt.Sprintf(`game.Mode{
	Targets: []game.TargetSpec{
		%s,
	},
	Sequence: []game.Instruction{
		{
			Primitive: game.Damage{
				Amount: game.Fixed(%d),
				Recipient: game.TargetRecipient(0),
			},
		},
	},
}.Ability()`, targetSource, ability.Effects[0].Amount.Value), nil
}

func executableSpell(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (string, *oracle.Diagnostic) {
	if len(ability.Effects) == 1 {
		switch ability.Effects[0].Kind {
		case oracle.EffectDealDamage:
			return executableFixedDamageSpell(cardName, ability, syntax)
		case oracle.EffectDraw:
			return executableFixedDrawSpell(ability, syntax)
		case oracle.EffectDestroy:
			return executableFixedDestroySpell(ability)
		case oracle.EffectGain:
			return executableFixedLifeSpell(ability, "gain", "GainLife")
		case oracle.EffectLose:
			return executableFixedLifeSpell(ability, "lose", "LoseLife")
		case oracle.EffectScry:
			return executableFixedControllerSpell(ability, syntax, "scry", "Scry")
		case oracle.EffectDiscard:
			return executableFixedCardCountPlayerSpell(
				ability, syntax, "discard", "discards", "Discard",
			)
		case oracle.EffectMill:
			return executableFixedCardCountPlayerSpell(
				ability, syntax, "mill", "mills", "Mill",
			)
		case oracle.EffectTap:
			return executableFixedPermanentTargetSpell(ability, "Tap", "Tap")
		case oracle.EffectUntap:
			return executableFixedPermanentTargetSpell(ability, "Untap", "Untap")
		default:
		}
	}
	return "", executableDiagnostic(
		ability,
		"unsupported spell ability",
		"the executable source backend does not yet lower this spell ability",
	)
}

func executableFixedPermanentTargetSpell(
	ability oracle.CompiledAbility,
	verb string,
	primitive string,
) (string, *oracle.Diagnostic) {
	if len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		ability.Effects[0].Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 {
		return "", executableDiagnostic(
			ability,
			"unsupported "+strings.ToLower(verb)+" spell",
			"the executable source backend supports only exact "+strings.ToLower(verb)+" of one target permanent",
		)
	}
	targetSource, ok := permanentTargetSource(ability.Targets[0])
	if !ok || ability.Text != verb+" "+ability.Targets[0].Text+"." {
		return "", executableDiagnostic(
			ability,
			"unsupported "+strings.ToLower(verb)+" spell",
			"the executable source backend supports only exact "+strings.ToLower(verb)+" of one target permanent",
		)
	}
	return fmt.Sprintf(`game.Mode{
	Targets: []game.TargetSpec{
		%s,
	},
	Sequence: []game.Instruction{
		{
			Primitive: game.%s{
				TargetIndex: 0,
			},
		},
	},
}.Ability()`, targetSource, primitive), nil
}

func executableFixedCardCountPlayerSpell(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
	controllerVerb string,
	targetVerb string,
	primitive string,
) (string, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if !effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 {
		return "", executableDiagnostic(
			ability,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	targetIndex := "game.TargetIndexController"
	var targets string
	switch {
	case len(ability.Targets) == 0 &&
		len(syntax.Tokens) == 4 &&
		strings.EqualFold(syntax.Tokens[0].Text, controllerVerb) &&
		fixedCardCountSyntax(syntax.Tokens[1], syntax.Tokens[2], effect.Amount.Value) &&
		syntax.Tokens[3].Kind == oracle.Period:
	case len(ability.Targets) == 1 &&
		len(syntax.Tokens) == 6 &&
		strings.EqualFold(syntax.Tokens[0].Text, "target") &&
		strings.EqualFold(syntax.Tokens[1].Text, "player") &&
		strings.EqualFold(syntax.Tokens[2].Text, targetVerb) &&
		fixedCardCountSyntax(syntax.Tokens[3], syntax.Tokens[4], effect.Amount.Value) &&
		syntax.Tokens[5].Kind == oracle.Period:
		targetSource, ok := playerTargetSource(ability.Targets[0])
		if !ok {
			return "", executableDiagnostic(
				ability,
				"unsupported "+controllerVerb+" spell",
				"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
			)
		}
		targetIndex = "0"
		targets = "Targets: []game.TargetSpec{\n" + targetSource + ",\n},"
	default:
		return "", executableDiagnostic(
			ability,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	return fmt.Sprintf(`game.Mode{
	%s
	Sequence: []game.Instruction{
		{
			Primitive: game.%s{
				Amount: game.Fixed(%d),
				TargetIndex: %s,
			},
		},
	},
}.Ability()`, targets, primitive, effect.Amount.Value, targetIndex), nil
}

func fixedCardCountSyntax(amountToken, cardToken oracle.Token, amount int) bool {
	if amount == 1 &&
		strings.EqualFold(amountToken.Text, "a") &&
		strings.EqualFold(cardToken.Text, "card") {
		return true
	}
	return fixedNumberToken(amountToken, amount) &&
		strings.EqualFold(cardToken.Text, "cards")
}

func executableFixedControllerSpell(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
	verb string,
	primitive string,
) (string, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if !effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Negated ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 ||
		len(syntax.Tokens) != 3 ||
		!strings.EqualFold(syntax.Tokens[0].Text, verb) ||
		!fixedNumberToken(syntax.Tokens[1], effect.Amount.Value) ||
		syntax.Tokens[2].Kind != oracle.Period {
		return "", executableDiagnostic(
			ability,
			"unsupported "+verb+" spell",
			"the executable source backend supports only exact fixed controller "+verb,
		)
	}
	return fmt.Sprintf(`game.Mode{
	Sequence: []game.Instruction{
		{
			Primitive: game.%s{
				Amount: game.Fixed(%d),
				TargetIndex: game.TargetIndexController,
			},
		},
	},
}.Ability()`, primitive, effect.Amount.Value), nil
}

func executableFixedLifeSpell(
	ability oracle.CompiledAbility,
	verb string,
	primitive string,
) (string, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if !effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 {
		return "", executableDiagnostic(
			ability,
			"unsupported life spell",
			"the executable source backend supports only exact fixed life changes",
		)
	}
	targetIndex := "game.TargetIndexController"
	var targets string
	switch {
	case len(ability.Targets) == 0 &&
		ability.Text == fmt.Sprintf("You %s %d life.", verb, effect.Amount.Value):
	case len(ability.Targets) == 1:
		targetSource, ok := playerTargetSource(ability.Targets[0])
		if !ok ||
			ability.Text != fmt.Sprintf(
				"%s %ss %d life.",
				titleFirst(ability.Targets[0].Text),
				verb,
				effect.Amount.Value,
			) {
			return "", executableDiagnostic(
				ability,
				"unsupported life spell",
				"the executable source backend supports only exact fixed life changes",
			)
		}
		targets = "Targets: []game.TargetSpec{\n" + targetSource + ",\n},"
		targetIndex = "0"
	default:
		return "", executableDiagnostic(
			ability,
			"unsupported life spell",
			"the executable source backend supports only exact fixed life changes",
		)
	}
	return fmt.Sprintf(`game.Mode{
	%s
	Sequence: []game.Instruction{
		{
			Primitive: game.%s{
				Amount: game.Fixed(%d),
				TargetIndex: %s,
			},
		},
	},
}.Ability()`, targets, primitive, effect.Amount.Value, targetIndex), nil
}

func playerTargetSource(target oracle.CompiledTarget) (string, bool) {
	const format = `{
	MinTargets: 1,
	MaxTargets: 1,
	Constraint: %q,
	Allow: game.TargetAllowPlayer,%s
}`
	var predicate string
	switch target.Selector.Kind {
	case oracle.SelectorPlayer:
		if !strings.EqualFold(target.Text, "target player") {
			return "", false
		}
	case oracle.SelectorOpponent:
		if !strings.EqualFold(target.Text, "target opponent") {
			return "", false
		}
		predicate = `
	Predicate: game.TargetPredicate{
		Player: game.PlayerOpponent,
	},`
	default:
		return "", false
	}
	return fmt.Sprintf(format, target.Text, predicate), true
}

func titleFirst(text string) string {
	if text == "" {
		return ""
	}
	return strings.ToUpper(text[:1]) + text[1:]
}

func executableFixedDestroySpell(
	ability oracle.CompiledAbility,
) (string, *oracle.Diagnostic) {
	if len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 ||
		ability.Effects[0].Negated {
		return "", executableDiagnostic(
			ability,
			"unsupported destroy spell",
			"the executable source backend supports only exact destruction of one target permanent",
		)
	}
	targetSource, ok := permanentTargetSource(ability.Targets[0])
	if !ok || ability.Text != "Destroy "+ability.Targets[0].Text+"." {
		return "", executableDiagnostic(
			ability,
			"unsupported destroy spell",
			"the executable source backend supports only exact destruction of one target permanent",
		)
	}
	return fmt.Sprintf(`game.Mode{
	Targets: []game.TargetSpec{
		%s,
	},
	Sequence: []game.Instruction{
		{
			Primitive: game.Destroy{
				TargetIndex: 0,
			},
		},
	},
}.Ability()`, targetSource), nil
}

func permanentTargetSource(target oracle.CompiledTarget) (string, bool) {
	const format = `{
	MinTargets: 1,
	MaxTargets: 1,
	Constraint: %q,
	Allow: game.TargetAllowPermanent,%s
}`
	var predicate string
	var noun string
	switch target.Selector.Kind {
	case oracle.SelectorArtifact:
		noun = "artifact"
		predicate = permanentTypePredicate("Artifact")
	case oracle.SelectorCreature:
		noun = "creature"
		predicate = permanentTypePredicate("Creature")
	case oracle.SelectorEnchantment:
		noun = "enchantment"
		predicate = permanentTypePredicate("Enchantment")
	case oracle.SelectorLand:
		noun = "land"
		predicate = permanentTypePredicate("Land")
	case oracle.SelectorPermanent:
		noun = "permanent"
	default:
		return "", false
	}
	if target.Text != "target "+noun {
		return "", false
	}
	return fmt.Sprintf(format, target.Text, predicate), true
}

func permanentTypePredicate(cardType string) string {
	return fmt.Sprintf(`
	Predicate: game.TargetPredicate{
		PermanentTypes: []types.Card{types.%s},
	},`, cardType)
}

func executableFixedDrawSpell(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (string, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if !effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 {
		return "", executableDiagnostic(
			ability,
			"unsupported draw spell",
			"the executable source backend supports only exact fixed card draw",
		)
	}
	targetIndex := "game.TargetIndexController"
	var targets string
	switch {
	case len(ability.Targets) == 0 &&
		exactControllerDrawSyntax(syntax.Tokens, effect.Amount.Value):
	case len(ability.Targets) == 1 &&
		exactTargetPlayerDrawSyntax(syntax.Tokens, effect.Amount.Value) &&
		ability.Targets[0].Cardinality.Min == 1 &&
		ability.Targets[0].Cardinality.Max == 1 &&
		ability.Targets[0].Selector.Kind == oracle.SelectorPlayer:
		targetIndex = "0"
		targets = `Targets: []game.TargetSpec{
		{
			MinTargets: 1,
			MaxTargets: 1,
			Constraint: "target player",
			Allow: game.TargetAllowPlayer,
		},
	},`
	default:
		return "", executableDiagnostic(
			ability,
			"unsupported draw spell",
			"the executable source backend supports only exact fixed card draw",
		)
	}
	return fmt.Sprintf(`game.Mode{
	%s
	Sequence: []game.Instruction{
		{
			Primitive: game.Draw{
				Amount: game.Fixed(%d),
				TargetIndex: %s,
			},
		},
	},
}.Ability()`, targets, effect.Amount.Value, targetIndex), nil
}

func exactControllerDrawSyntax(tokens []oracle.Token, amount int) bool {
	if len(tokens) != 4 ||
		tokens[0].Kind != oracle.Word ||
		!strings.EqualFold(tokens[0].Text, "draw") ||
		tokens[2].Kind != oracle.Word ||
		tokens[3].Kind != oracle.Period {
		return false
	}
	if amount == 1 &&
		strings.EqualFold(tokens[1].Text, "a") &&
		strings.EqualFold(tokens[2].Text, "card") {
		return true
	}
	return fixedNumberToken(tokens[1], amount) &&
		strings.EqualFold(tokens[2].Text, "cards")
}

func exactTargetPlayerDrawSyntax(tokens []oracle.Token, amount int) bool {
	return len(tokens) == 6 &&
		tokens[0].Kind == oracle.Word &&
		strings.EqualFold(tokens[0].Text, "target") &&
		tokens[1].Kind == oracle.Word &&
		strings.EqualFold(tokens[1].Text, "player") &&
		tokens[2].Kind == oracle.Word &&
		strings.EqualFold(tokens[2].Text, "draws") &&
		fixedNumberToken(tokens[3], amount) &&
		tokens[4].Kind == oracle.Word &&
		strings.EqualFold(tokens[4].Text, "cards") &&
		tokens[5].Kind == oracle.Period
}

func fixedNumberToken(token oracle.Token, amount int) bool {
	switch strings.ToLower(token.Text) {
	case "one":
		return amount == 1
	case "two":
		return amount == 2
	case "three":
		return amount == 3
	case "four":
		return amount == 4
	default:
		return token.Kind == oracle.Integer && token.Text == fmt.Sprint(amount)
	}
}

func singleSelfReference(references []oracle.CompiledReference) bool {
	return len(references) == 1 && references[0].Kind == oracle.ReferenceSelfName
}

func damageTargetSource(target oracle.CompiledTarget) (string, bool) {
	const format = `{
	MinTargets: 1,
	MaxTargets: 1,
	Constraint: %q,
	Allow: %s,%s
}`
	var allow string
	var predicate string
	switch target.Selector.Kind {
	case oracle.SelectorAny:
		if target.Text != "any target" {
			return "", false
		}
		allow = "game.TargetAllowPermanent | game.TargetAllowPlayer"
	case oracle.SelectorCreature:
		if target.Text != "target creature" {
			return "", false
		}
		allow = "game.TargetAllowPermanent"
		predicate = `
	Predicate: game.TargetPredicate{
		PermanentTypes: []types.Card{types.Creature},
	},`
	case oracle.SelectorPlaneswalker:
		if target.Text != "target planeswalker" {
			return "", false
		}
		allow = "game.TargetAllowPermanent"
		predicate = `
	Predicate: game.TargetPredicate{
		PermanentTypes: []types.Card{types.Planeswalker},
	},`
	case oracle.SelectorPlayer:
		if target.Text != "target player" {
			return "", false
		}
		allow = "game.TargetAllowPlayer"
	case oracle.SelectorOpponent:
		if target.Text != "target opponent" {
			return "", false
		}
		allow = "game.TargetAllowPlayer"
		predicate = `
	Predicate: game.TargetPredicate{
		Player: game.PlayerOpponent,
	},`
	default:
		return "", false
	}
	return fmt.Sprintf(format, target.Text, allow, predicate), true
}

func manaColorName(symbol string) (string, bool) {
	switch strings.ToUpper(symbol) {
	case "{W}":
		return "W", true
	case "{U}":
		return "U", true
	case "{B}":
		return "B", true
	case "{R}":
		return "R", true
	case "{G}":
		return "G", true
	case "{C}":
		return "C", true
	default:
		return "", false
	}
}

func rawStringLiteral(text string) string {
	return "`\n" + text + "\n`"
}

func executableDiagnostic(
	ability oracle.CompiledAbility,
	summary string,
	detail string,
) *oracle.Diagnostic {
	return &oracle.Diagnostic{
		Severity: oracle.SeverityWarning,
		Summary:  summary,
		Detail:   detail,
		Span:     ability.Span,
	}
}

func mixedKeywordDiagnostic(ability oracle.CompiledAbility) *oracle.Diagnostic {
	names := make([]string, 0, len(ability.Keywords))
	for _, keyword := range ability.Keywords {
		names = append(names, keyword.Name)
	}
	return executableDiagnostic(
		ability,
		"unsupported mixed keyword ability",
		fmt.Sprintf(
			"the executable source backend recognized %s but does not yet lower the additional rules text",
			strings.Join(names, ", "),
		),
	)
}

func spanCoveredByKeyword(span oracle.Span, keywords []oracle.CompiledKeyword) bool {
	for _, keyword := range keywords {
		if keyword.Span.Start.Offset <= span.Start.Offset &&
			keyword.Span.End.Offset >= span.End.Offset {
			return true
		}
	}
	return false
}

func spanCoveredByDelimited(span oracle.Span, groups []oracle.Delimited) bool {
	for _, group := range groups {
		if group.Span.Start.Offset <= span.Start.Offset &&
			group.Span.End.Offset >= span.End.Offset {
			return true
		}
	}
	return false
}

var keywordStaticBodies = map[string]string{
	"Deathtouch":     "game.DeathtouchStaticBody",
	"Defender":       "game.DefenderStaticBody",
	"Delve":          "game.DelveStaticBody",
	"Double strike":  "game.DoubleStrikeStaticBody",
	"Exalted":        "game.ExaltedStaticBody",
	"First strike":   "game.FirstStrikeStaticBody",
	"Flash":          "game.FlashStaticBody",
	"Flying":         "game.FlyingStaticBody",
	"Haste":          "game.HasteStaticBody",
	"Hexproof":       "game.HexproofStaticBody",
	"Improvise":      "game.ImproviseStaticBody",
	"Indestructible": "game.IndestructibleStaticBody",
	"Infect":         "game.InfectStaticBody",
	"Lifelink":       "game.LifelinkStaticBody",
	"Menace":         "game.MenaceStaticBody",
	"Persist":        "game.PersistStaticBody",
	"Prowess":        "game.ProwessStaticBody",
	"Reach":          "game.ReachStaticBody",
	"Shroud":         "game.ShroudStaticBody",
	"Split second":   "game.SplitSecondStaticBody",
	"Storm":          "game.StormStaticBody",
	"Trample":        "game.TrampleStaticBody",
	"Undying":        "game.UndyingStaticBody",
	"Vigilance":      "game.VigilanceStaticBody",
	"Wither":         "game.WitherStaticBody",
	"Cascade":        "game.CascadeStaticBody",
	"Convoke":        "game.ConvokeStaticBody",
}

func staticAbilityField(bodies []string) string {
	var builder strings.Builder
	_, _ = builder.WriteString("StaticAbilities: []game.StaticAbility{\n")
	for _, body := range bodies {
		_, _ = fmt.Fprintf(&builder, "\t%s,\n", body)
	}
	_, _ = builder.WriteString("},")
	return builder.String()
}

func manaAbilityField(abilities []string) string {
	var builder strings.Builder
	_, _ = builder.WriteString("ManaAbilities: []game.ManaAbility{\n")
	for _, ability := range abilities {
		_, _ = fmt.Fprintf(&builder, "\t%s,\n", ability)
	}
	_, _ = builder.WriteString("},")
	return builder.String()
}

func triggeredAbilityField(abilities []string) string {
	var builder strings.Builder
	_, _ = builder.WriteString("TriggeredAbilities: []game.TriggeredAbility{\n")
	for _, ability := range abilities {
		_, _ = fmt.Fprintf(&builder, "\t%s,\n", ability)
	}
	_, _ = builder.WriteString("},")
	return builder.String()
}
