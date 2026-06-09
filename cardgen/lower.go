package cardgen

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// loweredStaticAbility holds a typed StaticAbility with optional rendering
// metadata. VarName, when set, is a package-level variable reference like
// "game.FlyingStaticBody" that the Renderer emits instead of a struct literal.
type loweredStaticAbility struct {
	Body    game.StaticAbility
	VarName string
}

// loweredFaceAbilities holds the categorized typed game ability values
// produced by strict executable lowering for one card face, in Oracle order.
type loweredFaceAbilities struct {
	StaticAbilities      []loweredStaticAbility
	ActivatedAbilities   []game.ActivatedAbility
	ManaAbilities        []game.ManaAbility
	TriggeredAbilities   []game.TriggeredAbility
	ReplacementAbilities []game.ReplacementAbility
	SpellAbility         opt.V[game.AbilityContent]
}

// empty reports whether the face produced no abilities.
func (f loweredFaceAbilities) empty() bool {
	return len(f.StaticAbilities) == 0 &&
		len(f.ActivatedAbilities) == 0 &&
		len(f.ManaAbilities) == 0 &&
		len(f.TriggeredAbilities) == 0 &&
		len(f.ReplacementAbilities) == 0 &&
		!f.SpellAbility.Exists
}

// abilityLowering holds the typed result of lowering one CompiledAbility.
// Fields are set according to which ability kind was matched.
type abilityLowering struct {
	staticAbilities    []loweredStaticAbility
	activatedAbility   opt.V[game.ActivatedAbility]
	manaAbility        opt.V[game.ManaAbility]
	triggeredAbility   opt.V[game.TriggeredAbility]
	replacementAbility opt.V[game.ReplacementAbility]
	spellAbility       opt.V[game.AbilityContent]
	consumed           semanticConsumption
	sourceSpans        []oracle.Span
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

// lowerExecutableFaces lowers every face of a card into typed ability values.
// It returns the face abilities in the same positional order as
// executableFaces and any diagnostics that prevented full lowering.
func lowerExecutableFaces(card *ScryfallCard) ([]loweredFaceAbilities, []oracle.Diagnostic) {
	faces := executableFaces(card)
	lowered := make([]loweredFaceAbilities, len(faces))
	var diagnostics []oracle.Diagnostic
	for i, face := range faces {
		faceAbilities, faceDiagnostics := lowerFaceAbilities(face)
		diagnostics = append(diagnostics, faceDiagnostics...)
		lowered[i] = faceAbilities
	}
	return lowered, diagnostics
}

func lowerFaceAbilities(
	face generatedCardFields,
) (loweredFaceAbilities, []oracle.Diagnostic) {
	parsedType := ParseTypeLine(face.TypeLine)
	if len(parsedType.Types) == 0 {
		return loweredFaceAbilities{}, []oracle.Diagnostic{{
			Severity: oracle.SeverityWarning,
			Summary:  "unsupported type line",
			Detail:   fmt.Sprintf("type line %q has no supported card type", face.TypeLine),
		}}
	}
	if face.OracleText == "" {
		return loweredFaceAbilities{}, nil
	}
	compilation, diagnostics := oracle.Compile(face.OracleText, oracle.ParseContext{
		CardName:         face.Name,
		InstantOrSorcery: slices.Contains(parsedType.Types, "Instant") || slices.Contains(parsedType.Types, "Sorcery"),
		Planeswalker:     slices.Contains(parsedType.Types, "Planeswalker"),
	})
	if len(diagnostics) > 0 {
		return loweredFaceAbilities{}, diagnostics
	}

	var result loweredFaceAbilities
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
		result.StaticAbilities = append(result.StaticAbilities, lowered.staticAbilities...)
		if lowered.activatedAbility.Exists {
			result.ActivatedAbilities = append(result.ActivatedAbilities, lowered.activatedAbility.Val)
		}
		if lowered.manaAbility.Exists {
			result.ManaAbilities = append(result.ManaAbilities, lowered.manaAbility.Val)
		}
		if lowered.triggeredAbility.Exists {
			result.TriggeredAbilities = append(result.TriggeredAbilities, lowered.triggeredAbility.Val)
		}
		if lowered.replacementAbility.Exists {
			result.ReplacementAbilities = append(result.ReplacementAbilities, lowered.replacementAbility.Val)
		}
		if lowered.spellAbility.Exists {
			if result.SpellAbility.Exists {
				unsupported = append(unsupported, *executableDiagnostic(
					ability,
					"unsupported multiple spell abilities",
					"the executable source backend supports only one spell ability per card face",
				))
				continue
			}
			result.SpellAbility = lowered.spellAbility
		}
	}
	if len(unsupported) > 0 {
		return loweredFaceAbilities{}, append(diagnostics, unsupported...)
	}
	return result, diagnostics
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
	if equipAbility, ok, diagnostic := lowerEquipAbility(ability, syntax); ok {
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		spans := []oracle.Span{ability.Keywords[0].Span}
		for _, reminder := range syntax.Reminders {
			spans = append(spans, reminder.Span)
		}
		return abilityLowering{
			activatedAbility: opt.Val(equipAbility),
			consumed: semanticConsumption{
				keywords: 1,
			},
			sourceSpans: spans,
		}, nil
	}
	if cyclingAbility, ok, diagnostic := lowerCyclingAbility(ability, syntax); ok {
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		spans := []oracle.Span{ability.Keywords[0].Span}
		for _, reminder := range syntax.Reminders {
			spans = append(spans, reminder.Span)
		}
		return abilityLowering{
			activatedAbility: opt.Val(cyclingAbility),
			consumed: semanticConsumption{
				keywords: 1,
			},
			sourceSpans: spans,
		}, nil
	}
	switch ability.Kind {
	case oracle.AbilityStatic:
		bodies, diagnostic := lowerKeywordAbility(ability, syntax)
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
			staticAbilities: bodies,
			consumed: semanticConsumption{
				keywords: len(ability.Keywords),
			},
			sourceSpans: spans,
		}, nil
	case oracle.AbilityActivated:
		manaAbility, diagnostic := lowerTapManaAbility(ability, syntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		return abilityLowering{
			manaAbility: opt.Val(manaAbility),
			consumed: semanticConsumption{
				cost:    true,
				effects: 1,
			},
			sourceSpans: []oracle.Span{ability.Cost.Span, ability.Effects[0].Span},
		}, nil
	case oracle.AbilitySpell:
		spellAbility, diagnostic := lowerSpell(cardName, ability, syntax)
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
			spellAbility: opt.Val(spellAbility),
			consumed: semanticConsumption{
				targets:    len(ability.Targets),
				effects:    1,
				references: len(ability.References),
			},
			sourceSpans: spans,
		}, nil
	case oracle.AbilityTriggered:
		triggeredAbility, diagnostic := lowerEnterTrigger(cardName, ability, syntax)
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
			triggeredAbility: opt.Val(triggeredAbility),
			consumed: semanticConsumption{
				trigger:    true,
				targets:    len(ability.Targets),
				effects:    len(ability.Effects),
				references: len(ability.References),
			},
			sourceSpans: spans,
		}, nil
	case oracle.AbilityReplacement:
		replacementAbility, diagnostic := lowerEntersTappedReplacement(ability)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		return abilityLowering{
			replacementAbility: opt.Val(replacementAbility),
			consumed: semanticConsumption{
				effects:    1,
				references: len(ability.References),
			},
			sourceSpans: []oracle.Span{ability.Effects[0].Span},
		}, nil
	default:
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported "+ability.Kind.String()+" ability",
			"the executable source backend does not yet lower "+ability.Kind.String()+" abilities",
		)
	}
}

func lowerEquipAbility(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.ActivatedAbility, bool, *oracle.Diagnostic) {
	if len(ability.Keywords) != 1 || ability.Keywords[0].Name != "Equip" {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Keywords[0]
	if keyword.Parameter == "" ||
		ability.Kind != oracle.AbilityStatic ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Effects) != 0 ||
		len(ability.References) != 0 ||
		ability.AbilityWord != "" {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Equip ability",
			"the executable source backend supports only exact Equip with a mana cost",
		)
	}
	manaCost, err := parseManaCostValue(keyword.Parameter)
	if err != nil || len(manaCost) == 0 {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Equip ability",
			"the executable source backend supports only exact Equip with a mana cost",
		)
	}
	for _, token := range syntax.Tokens {
		if spanCovered(token.Span, []oracle.Span{keyword.Span}) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Equip ability",
			"the executable source backend supports only exact Equip with a mana cost",
		)
	}
	return game.EquipActivatedAbility(manaCost), true, nil
}

func lowerCyclingAbility(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.ActivatedAbility, bool, *oracle.Diagnostic) {
	if len(ability.Keywords) != 1 || ability.Keywords[0].Name != "Cycling" {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Keywords[0]
	if keyword.Parameter == "" ||
		(ability.Kind != oracle.AbilityStatic && ability.Kind != oracle.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Effects) != 0 ||
		len(ability.References) != 0 ||
		ability.AbilityWord != "" {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Cycling ability",
			"the executable source backend supports only exact Cycling with a mana cost",
		)
	}
	manaCost, err := parseManaCostValue(keyword.Parameter)
	if err != nil || len(manaCost) == 0 {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Cycling ability",
			"the executable source backend supports only exact Cycling with a mana cost",
		)
	}
	for _, token := range syntax.Tokens {
		if spanCovered(token.Span, []oracle.Span{keyword.Span}) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Cycling ability",
			"the executable source backend supports only exact Cycling with a mana cost",
		)
	}
	return game.CyclingActivatedAbility(manaCost), true, nil
}

func lowerEntersTappedReplacement(
	ability oracle.CompiledAbility,
) (game.ReplacementAbility, *oracle.Diagnostic) {
	if len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != oracle.EffectEnterTapped ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 1 ||
		ability.References[0].Kind != oracle.ReferenceThisObject {
		return game.ReplacementAbility{}, executableDiagnostic(
			ability,
			"unsupported enters-tapped replacement",
			"the executable source backend supports only exact unconditional self enters-tapped replacements",
		)
	}
	switch ability.Text {
	case "This land enters tapped.",
		"This artifact enters tapped.",
		"This creature enters tapped.",
		"This permanent enters tapped.":
	default:
		return game.ReplacementAbility{}, executableDiagnostic(
			ability,
			"unsupported enters-tapped replacement",
			"the executable source backend supports only exact unconditional self enters-tapped replacements",
		)
	}
	return game.EntersTappedReplacement(ability.Text), nil
}

func (lowering *abilityLowering) complete(
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

func lowerEnterTrigger(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.TriggeredAbility, *oracle.Diagnostic) {
	eventKind, supportedEvent := lowerSelfTriggerEvent(ability)
	summary := "unsupported enter trigger"
	detail := "the executable source backend supports only exact self-enter triggers with one supported effect"
	if ability.Trigger != nil && strings.HasSuffix(ability.Trigger.Event, " dies") {
		summary = "unsupported dies trigger"
		detail = "the executable source backend supports only exact self-dies triggers with one supported effect"
	}
	if ability.Trigger == nil ||
		ability.Trigger.Kind != oracle.TriggerWhen ||
		!supportedEvent ||
		ability.Trigger.Condition != nil ||
		len(ability.Effects) != 1 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		ability.AbilityWord != "" {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			detail,
		)
	}
	comma := slices.IndexFunc(syntax.Tokens, func(token oracle.Token) bool {
		return token.Kind == oracle.Comma
	})
	if comma < 0 || comma+1 >= len(syntax.Tokens) {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			detail,
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
	content, diagnostic := lowerSpell(cardName, body, bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary+" effect",
			diagnostic.Detail,
		)
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhen,
			Pattern: game.TriggerPattern{
				Event:  eventKind,
				Source: game.TriggerSourceSelf,
			},
		},
		Content: content,
	}, nil
}

func lowerSelfTriggerEvent(ability oracle.CompiledAbility) (game.EventKind, bool) {
	if ability.Trigger == nil {
		return 0, false
	}
	switch ability.Trigger.Event {
	case "this creature enters",
		"this permanent enters",
		"this aura enters",
		"this artifact enters",
		"this equipment enters",
		"this land enters",
		"this vehicle enters",
		"this enchantment enters":
		return game.EventPermanentEnteredBattlefield, true
	case "this creature dies", "this permanent dies":
		return game.EventPermanentDied, true
	default:
		return 0, false
	}
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

func lowerKeywordAbility(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) ([]loweredStaticAbility, *oracle.Diagnostic) {
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
	bodies := make([]loweredStaticAbility, 0, len(ability.Keywords))
	for _, keyword := range ability.Keywords {
		if keyword.Parameter != "" {
			if keyword.Name == "Ward" {
				manaCost, err := parseManaCostValue(keyword.Parameter)
				if err == nil && len(manaCost) > 0 {
					bodies = append(bodies, loweredStaticAbility{
						Body: game.WardStaticAbility(manaCost),
					})
					continue
				}
			}
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

func lowerTapManaAbility(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.ManaAbility, *oracle.Diagnostic) {
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
		len(ability.References) != 0 {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported activated ability",
			"the executable source backend supports only exact supported tap mana abilities",
		)
	}
	if exactAnyColorTapManaSyntax(syntax.Tokens) {
		return choiceTapManaAbility(
			[]string{"W", "U", "B", "R", "G"},
		), nil
	}
	if colors, ok := exactChoiceTapManaSyntax(syntax.Tokens); ok {
		return choiceTapManaAbility(colors), nil
	}
	if !exactTapManaSyntax(syntax.Tokens) {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported activated ability",
			"the executable source backend supports only exact supported tap mana abilities",
		)
	}
	colorName, ok := manaColorName(ability.Effects[0].Symbol)
	if !ok {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported mana symbol",
			fmt.Sprintf("the executable source backend cannot emit mana symbol %q", ability.Effects[0].Symbol),
		)
	}
	manaColor, ok := manaColorValue(colorName)
	if !ok {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported mana symbol",
			fmt.Sprintf("the executable source backend cannot emit mana symbol %q", ability.Effects[0].Symbol),
		)
	}
	return game.TapManaAbility(manaColor), nil
}

func choiceTapManaAbility(colorNames []string) game.ManaAbility {
	colors := make([]mana.Color, 0, len(colorNames))
	for _, name := range colorNames {
		if color, ok := manaColorValue(name); ok {
			colors = append(colors, color)
		}
	}
	return game.TapManaChoiceAbility(colors...)
}

func lowerSpell(
	cardName string,
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.AbilityContent, *oracle.Diagnostic) {
	if len(ability.Effects) == 1 {
		switch ability.Effects[0].Kind {
		case oracle.EffectDealDamage:
			return lowerFixedDamageSpell(cardName, ability)
		case oracle.EffectDraw:
			return lowerFixedDrawSpell(ability, syntax)
		case oracle.EffectDestroy:
			return lowerFixedDestroySpell(ability)
		case oracle.EffectGain:
			return lowerFixedLifeSpell(ability, "gain", func(amount int, player game.PlayerReference) game.Primitive {
				return game.GainLife{Amount: game.Fixed(amount), Player: player}
			})
		case oracle.EffectLose:
			return lowerFixedLifeSpell(ability, "lose", func(amount int, player game.PlayerReference) game.Primitive {
				return game.LoseLife{Amount: game.Fixed(amount), Player: player}
			})
		case oracle.EffectScry:
			return lowerFixedControllerSpell(ability, syntax, "scry", func(amount int, player game.PlayerReference) game.Primitive {
				return game.Scry{Amount: game.Fixed(amount), Player: player}
			})
		case oracle.EffectDiscard:
			return lowerFixedCardCountPlayerSpell(
				ability, syntax, "discard", "discards", func(amount int, player game.PlayerReference) game.Primitive {
					return game.Discard{Amount: game.Fixed(amount), Player: player}
				},
			)
		case oracle.EffectMill:
			return lowerFixedCardCountPlayerSpell(
				ability, syntax, "mill", "mills", func(amount int, player game.PlayerReference) game.Primitive {
					return game.Mill{Amount: game.Fixed(amount), Player: player}
				},
			)
		case oracle.EffectTap:
			return lowerFixedPermanentTargetSpell(ability, "Tap", func(object game.ObjectReference) game.Primitive {
				return game.Tap{Object: object}
			})
		case oracle.EffectUntap:
			return lowerFixedPermanentTargetSpell(ability, "Untap", func(object game.ObjectReference) game.Primitive {
				return game.Untap{Object: object}
			})
		case oracle.EffectExile:
			return lowerFixedPermanentTargetSpell(ability, "Exile", func(object game.ObjectReference) game.Primitive {
				return game.Exile{Object: object}
			})
		case oracle.EffectReturn:
			return lowerFixedBounceSpell(ability)
		case oracle.EffectModifyPT:
			return lowerFixedModifyPTSpell(ability)
		default:
		}
	}
	return game.AbilityContent{}, executableDiagnostic(
		ability,
		"unsupported spell ability",
		"the executable source backend does not yet lower this spell ability",
	)
}

func lowerFixedDamageSpell(
	cardName string,
	ability oracle.CompiledAbility,
) (game.AbilityContent, *oracle.Diagnostic) {
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
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed damage to one target",
		)
	}
	target, ok := damageTargetSpec(ability.Targets[0])
	if !ok ||
		ability.Text != fmt.Sprintf(
			"%s deals %d damage to %s.",
			cardName,
			ability.Effects[0].Amount.Value,
			ability.Targets[0].Text,
		) {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed damage to one target",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{
			{
				Primitive: game.Damage{
					Amount:    game.Fixed(ability.Effects[0].Amount.Value),
					Recipient: game.AnyTargetDamageRecipient(0),
				},
			},
		},
	}.Ability(), nil
}

func lowerFixedModifyPTSpell(
	ability oracle.CompiledAbility,
) (game.AbilityContent, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		ability.Targets[0].Selector.Kind != oracle.SelectorCreature ||
		!effect.PowerDelta.Known ||
		!effect.ToughnessDelta.Known ||
		effect.Negated ||
		effect.Duration != oracle.DurationUntilEndOfTurn ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 ||
		ability.Text != fmt.Sprintf(
			"Target creature gets %s/%s until end of turn.",
			signedAmountText(effect.PowerDelta),
			signedAmountText(effect.ToughnessDelta),
		) {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact fixed target-creature power/toughness changes until end of turn",
		)
	}
	target := ability.Targets[0]
	target.Text = "target creature"
	targetSpec, ok := permanentTargetSpec(target)
	if !ok {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact fixed target-creature power/toughness changes until end of turn",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.ModifyPT{
					Object:         game.TargetPermanentReference(0),
					PowerDelta:     game.Fixed(effect.PowerDelta.Value),
					ToughnessDelta: game.Fixed(effect.ToughnessDelta.Value),
					Duration:       game.DurationUntilEndOfTurn,
				},
			},
		},
	}.Ability(), nil
}

func lowerFixedBounceSpell(
	ability oracle.CompiledAbility,
) (game.AbilityContent, *oracle.Diagnostic) {
	if len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		ability.Effects[0].Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 1 ||
		ability.References[0].Kind != oracle.ReferencePronoun ||
		!strings.EqualFold(ability.References[0].Text, "its") {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported return spell",
			"the executable source backend supports only exact return of one target permanent to its owner's hand",
		)
	}
	target := ability.Targets[0]
	var targetSpec game.TargetSpec
	var ok bool
	for _, noun := range []string{"artifact", "creature", "enchantment", "land", "permanent"} {
		if ability.Text != "Return target "+noun+" to its owner's hand." {
			continue
		}
		target.Text = "target " + noun
		targetSpec, ok = permanentTargetSpec(target)
		break
	}
	if !ok {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported return spell",
			"the executable source backend supports only exact return of one target permanent to its owner's hand",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.Bounce{
					Object: game.TargetPermanentReference(0),
				},
			},
		},
	}.Ability(), nil
}

func lowerFixedPermanentTargetSpell(
	ability oracle.CompiledAbility,
	verb string,
	primitiveFactory func(object game.ObjectReference) game.Primitive,
) (game.AbilityContent, *oracle.Diagnostic) {
	if len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		ability.Effects[0].Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported "+strings.ToLower(verb)+" spell",
			"the executable source backend supports only exact "+strings.ToLower(verb)+" of one target permanent",
		)
	}
	targetSpec, ok := permanentTargetSpec(ability.Targets[0])
	if !ok || ability.Text != verb+" "+ability.Targets[0].Text+"." {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported "+strings.ToLower(verb)+" spell",
			"the executable source backend supports only exact "+strings.ToLower(verb)+" of one target permanent",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(game.TargetPermanentReference(0)),
			},
		},
	}.Ability(), nil
}

func lowerFixedCardCountPlayerSpell(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
	controllerVerb string,
	targetVerb string,
	primitiveFactory func(amount int, player game.PlayerReference) game.Primitive,
) (game.AbilityContent, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if !effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
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
		targetSpec, ok := playerTargetSpec(ability.Targets[0])
		if !ok {
			return game.AbilityContent{}, executableDiagnostic(
				ability,
				"unsupported "+controllerVerb+" spell",
				"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
			)
		}
		playerRef = game.TargetPlayerReference(0)
		targets = []game.TargetSpec{targetSpec}
	default:
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(effect.Amount.Value, playerRef),
			},
		},
	}.Ability(), nil
}

func lowerFixedControllerSpell(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
	verb string,
	primitiveFactory func(amount int, player game.PlayerReference) game.Primitive,
) (game.AbilityContent, *oracle.Diagnostic) {
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
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported "+verb+" spell",
			"the executable source backend supports only exact fixed controller "+verb,
		)
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(effect.Amount.Value, game.ControllerReference()),
			},
		},
	}.Ability(), nil
}

func lowerFixedLifeSpell(
	ability oracle.CompiledAbility,
	verb string,
	primitiveFactory func(amount int, player game.PlayerReference) game.Primitive,
) (game.AbilityContent, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if !effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported life spell",
			"the executable source backend supports only exact fixed life changes",
		)
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	switch {
	case len(ability.Targets) == 0 &&
		ability.Text == fmt.Sprintf("You %s %d life.", verb, effect.Amount.Value):
	case len(ability.Targets) == 1:
		targetSpec, ok := playerTargetSpec(ability.Targets[0])
		if !ok ||
			ability.Text != fmt.Sprintf(
				"%s %ss %d life.",
				titleFirst(ability.Targets[0].Text),
				verb,
				effect.Amount.Value,
			) {
			return game.AbilityContent{}, executableDiagnostic(
				ability,
				"unsupported life spell",
				"the executable source backend supports only exact fixed life changes",
			)
		}
		targets = []game.TargetSpec{targetSpec}
		playerRef = game.TargetPlayerReference(0)
	default:
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported life spell",
			"the executable source backend supports only exact fixed life changes",
		)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(effect.Amount.Value, playerRef),
			},
		},
	}.Ability(), nil
}

func lowerFixedDestroySpell(
	ability oracle.CompiledAbility,
) (game.AbilityContent, *oracle.Diagnostic) {
	if group, ok := exactMassDestroyGroup(ability); ok {
		return game.Mode{
			Sequence: []game.Instruction{
				{
					Primitive: game.Destroy{
						Group: group,
					},
				},
			},
		}.Ability(), nil
	}
	if len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 ||
		ability.Effects[0].Negated {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported destroy spell",
			"the executable source backend supports only exact destruction of one target permanent",
		)
	}
	targetSpec, ok := permanentTargetSpec(ability.Targets[0])
	if !ok || ability.Text != "Destroy "+ability.Targets[0].Text+"." {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported destroy spell",
			"the executable source backend supports only exact destruction of one target permanent",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.Destroy{
					Object: game.TargetPermanentReference(0),
				},
			},
		},
	}.Ability(), nil
}

func exactMassDestroyGroup(ability oracle.CompiledAbility) (game.GroupReference, bool) {
	if len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 ||
		ability.Effects[0].Negated {
		return game.GroupReference{}, false
	}
	switch ability.Text {
	case "Destroy all creatures.":
		return game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}), true
	case "Destroy all artifacts.":
		return game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}}), true
	case "Destroy all enchantments.":
		return game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}}), true
	default:
		return game.GroupReference{}, false
	}
}

func lowerFixedDrawSpell(
	ability oracle.CompiledAbility,
	syntax oracle.Ability,
) (game.AbilityContent, *oracle.Diagnostic) {
	effect := ability.Effects[0]
	if !effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Negated ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		len(ability.References) != 0 {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported draw spell",
			"the executable source backend supports only exact fixed card draw",
		)
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	switch {
	case len(ability.Targets) == 0 &&
		exactControllerDrawSyntax(syntax.Tokens, effect.Amount.Value):
	case len(ability.Targets) == 1 &&
		exactTargetPlayerDrawSyntax(syntax.Tokens, effect.Amount.Value) &&
		ability.Targets[0].Cardinality.Min == 1 &&
		ability.Targets[0].Cardinality.Max == 1 &&
		ability.Targets[0].Selector.Kind == oracle.SelectorPlayer:
		playerRef = game.TargetPlayerReference(0)
		targets = []game.TargetSpec{
			{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target player",
				Allow:      game.TargetAllowPlayer,
			},
		}
	default:
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported draw spell",
			"the executable source backend supports only exact fixed card draw",
		)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{
			{
				Primitive: game.Draw{
					Amount: game.Fixed(effect.Amount.Value),
					Player: playerRef,
				},
			},
		},
	}.Ability(), nil
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

func fixedCardCountSyntax(amountToken, cardToken oracle.Token, amount int) bool {
	if amount == 1 &&
		strings.EqualFold(amountToken.Text, "a") &&
		strings.EqualFold(cardToken.Text, "card") {
		return true
	}
	return fixedNumberToken(amountToken, amount) &&
		strings.EqualFold(cardToken.Text, "cards")
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

func damageTargetSpec(target oracle.CompiledTarget) (game.TargetSpec, bool) {
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: target.Text,
	}
	switch target.Selector.Kind {
	case oracle.SelectorAny:
		if target.Text != "any target" {
			return game.TargetSpec{}, false
		}
		spec.Allow = game.TargetAllowPermanent | game.TargetAllowPlayer
	case oracle.SelectorCreature:
		if target.Text != "target creature" {
			return game.TargetSpec{}, false
		}
		spec.Allow = game.TargetAllowPermanent
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Creature}}
	case oracle.SelectorPlaneswalker:
		if target.Text != "target planeswalker" {
			return game.TargetSpec{}, false
		}
		spec.Allow = game.TargetAllowPermanent
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Planeswalker}}
	case oracle.SelectorPlayer:
		if target.Text != "target player" {
			return game.TargetSpec{}, false
		}
		spec.Allow = game.TargetAllowPlayer
	case oracle.SelectorOpponent:
		if target.Text != "target opponent" {
			return game.TargetSpec{}, false
		}
		spec.Allow = game.TargetAllowPlayer
		spec.Predicate = game.TargetPredicate{Player: game.PlayerOpponent}
	default:
		return game.TargetSpec{}, false
	}
	return spec, true
}

func permanentTargetSpec(target oracle.CompiledTarget) (game.TargetSpec, bool) {
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
	}
	var noun string
	switch target.Selector.Kind {
	case oracle.SelectorArtifact:
		noun = "artifact"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Artifact}}
	case oracle.SelectorCreature:
		noun = "creature"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Creature}}
	case oracle.SelectorEnchantment:
		noun = "enchantment"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Enchantment}}
	case oracle.SelectorLand:
		noun = "land"
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Land}}
	case oracle.SelectorPermanent:
		noun = "permanent"
	default:
		return game.TargetSpec{}, false
	}
	if target.Text != "target "+noun {
		return game.TargetSpec{}, false
	}
	spec.Constraint = target.Text
	return spec, true
}

func playerTargetSpec(target oracle.CompiledTarget) (game.TargetSpec, bool) {
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: target.Text,
		Allow:      game.TargetAllowPlayer,
	}
	switch target.Selector.Kind {
	case oracle.SelectorPlayer:
		if !strings.EqualFold(target.Text, "target player") {
			return game.TargetSpec{}, false
		}
	case oracle.SelectorOpponent:
		if !strings.EqualFold(target.Text, "target opponent") {
			return game.TargetSpec{}, false
		}
		spec.Predicate = game.TargetPredicate{Player: game.PlayerOpponent}
	default:
		return game.TargetSpec{}, false
	}
	return spec, true
}

func signedAmountText(amount oracle.CompiledSignedAmount) string {
	if amount.Negative {
		magnitude := amount.Value
		if magnitude < 0 {
			magnitude = -magnitude
		}
		return fmt.Sprintf("-%d", magnitude)
	}
	return fmt.Sprintf("+%d", amount.Value)
}

func titleFirst(text string) string {
	if text == "" {
		return ""
	}
	return strings.ToUpper(text[:1]) + text[1:]
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

func manaColorValue(name string) (mana.Color, bool) {
	switch name {
	case "W":
		return mana.W, true
	case "U":
		return mana.U, true
	case "B":
		return mana.B, true
	case "R":
		return mana.R, true
	case "G":
		return mana.G, true
	case "C":
		return mana.C, true
	default:
		return "", false
	}
}

func exactAnyColorTapManaSyntax(tokens []oracle.Token) bool {
	return len(tokens) == 9 &&
		tokens[0].Kind == oracle.Symbol &&
		strings.EqualFold(tokens[0].Text, "{T}") &&
		tokens[1].Kind == oracle.Colon &&
		equalTokenWord(tokens[2], "add") &&
		equalTokenWord(tokens[3], "one") &&
		equalTokenWord(tokens[4], "mana") &&
		equalTokenWord(tokens[5], "of") &&
		equalTokenWord(tokens[6], "any") &&
		equalTokenWord(tokens[7], "color") &&
		tokens[8].Kind == oracle.Period
}

func equalTokenWord(token oracle.Token, word string) bool {
	return token.Kind == oracle.Word && strings.EqualFold(token.Text, word)
}

func exactChoiceTapManaSyntax(tokens []oracle.Token) ([]string, bool) {
	if len(tokens) < 7 ||
		tokens[0].Kind != oracle.Symbol ||
		!strings.EqualFold(tokens[0].Text, "{T}") ||
		tokens[1].Kind != oracle.Colon ||
		!equalTokenWord(tokens[2], "add") ||
		tokens[len(tokens)-1].Kind != oracle.Period {
		return nil, false
	}
	var colors []string
	for i := 3; i < len(tokens)-1; {
		token := tokens[i]
		color, ok := manaColorName(token.Text)
		if token.Kind != oracle.Symbol || !ok {
			return nil, false
		}
		colors = append(colors, color)
		i++
		if i == len(tokens)-1 {
			break
		}
		if tokens[i].Kind == oracle.Comma {
			i++
			if i < len(tokens)-1 && equalTokenWord(tokens[i], "or") {
				i++
			}
			continue
		}
		if !equalTokenWord(tokens[i], "or") {
			return nil, false
		}
		i++
	}
	return colors, len(colors) >= 2
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

// parseManaCostValue parses a Scryfall mana cost string (e.g., "{2}{W}") into a
// typed cost.Mana value. Empty input yields a nil cost.
func parseManaCostValue(s string) (cost.Mana, error) {
	if s == "" {
		return nil, nil
	}
	matches := manaSymbolRe.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	out := make(cost.Mana, 0, len(matches))
	for _, match := range matches {
		symbol, err := parseManaSymbolValue(match[1])
		if err != nil {
			return nil, fmt.Errorf("unsupported mana symbol {%s} in cost %q: %w", match[1], s, err)
		}
		out = append(out, symbol)
	}
	return out, nil
}

func parseManaSymbolValue(sym string) (cost.Symbol, error) {
	switch sym {
	case "X":
		return cost.X, nil
	case "C":
		return cost.C, nil
	case "S":
		return cost.S, nil
	case "W":
		return cost.W, nil
	case "U":
		return cost.U, nil
	case "B":
		return cost.B, nil
	case "R":
		return cost.R, nil
	case "G":
		return cost.G, nil
	default:
	}
	if before, ok := strings.CutSuffix(sym, "/P"); ok {
		color, ok := manaColorValue(before)
		if !ok {
			return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
		}
		return cost.PhyrexianMana(color), nil
	}
	if strings.Contains(sym, "/") {
		parts := strings.SplitN(sym, "/", 2)
		if _, err := strconv.Atoi(parts[0]); err == nil {
			color, ok := manaColorValue(parts[1])
			if !ok {
				return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
			}
			return cost.Twobrid(color), nil
		}
		first, ok := manaColorValue(parts[0])
		second, ok2 := manaColorValue(parts[1])
		if !ok || !ok2 {
			return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
		}
		return cost.HybridMana(first, second), nil
	}
	if n, err := strconv.Atoi(sym); err == nil {
		return cost.O(n), nil
	}
	return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
}

// keywordStaticBodies maps a keyword name to its reusable typed StaticAbility and
// the package-level variable reference the Renderer emits for it.
var keywordStaticBodies = map[string]loweredStaticAbility{
	"Deathtouch":     {Body: game.DeathtouchStaticBody, VarName: "game.DeathtouchStaticBody"},
	"Defender":       {Body: game.DefenderStaticBody, VarName: "game.DefenderStaticBody"},
	"Delve":          {Body: game.DelveStaticBody, VarName: "game.DelveStaticBody"},
	"Double strike":  {Body: game.DoubleStrikeStaticBody, VarName: "game.DoubleStrikeStaticBody"},
	"Exalted":        {Body: game.ExaltedStaticBody, VarName: "game.ExaltedStaticBody"},
	"First strike":   {Body: game.FirstStrikeStaticBody, VarName: "game.FirstStrikeStaticBody"},
	"Flash":          {Body: game.FlashStaticBody, VarName: "game.FlashStaticBody"},
	"Flying":         {Body: game.FlyingStaticBody, VarName: "game.FlyingStaticBody"},
	"Haste":          {Body: game.HasteStaticBody, VarName: "game.HasteStaticBody"},
	"Hexproof":       {Body: game.HexproofStaticBody, VarName: "game.HexproofStaticBody"},
	"Improvise":      {Body: game.ImproviseStaticBody, VarName: "game.ImproviseStaticBody"},
	"Indestructible": {Body: game.IndestructibleStaticBody, VarName: "game.IndestructibleStaticBody"},
	"Infect":         {Body: game.InfectStaticBody, VarName: "game.InfectStaticBody"},
	"Lifelink":       {Body: game.LifelinkStaticBody, VarName: "game.LifelinkStaticBody"},
	"Menace":         {Body: game.MenaceStaticBody, VarName: "game.MenaceStaticBody"},
	"Persist":        {Body: game.PersistStaticBody, VarName: "game.PersistStaticBody"},
	"Prowess":        {Body: game.ProwessStaticBody, VarName: "game.ProwessStaticBody"},
	"Reach":          {Body: game.ReachStaticBody, VarName: "game.ReachStaticBody"},
	"Shroud":         {Body: game.ShroudStaticBody, VarName: "game.ShroudStaticBody"},
	"Split second":   {Body: game.SplitSecondStaticBody, VarName: "game.SplitSecondStaticBody"},
	"Storm":          {Body: game.StormStaticBody, VarName: "game.StormStaticBody"},
	"Trample":        {Body: game.TrampleStaticBody, VarName: "game.TrampleStaticBody"},
	"Undying":        {Body: game.UndyingStaticBody, VarName: "game.UndyingStaticBody"},
	"Vigilance":      {Body: game.VigilanceStaticBody, VarName: "game.VigilanceStaticBody"},
	"Wither":         {Body: game.WitherStaticBody, VarName: "game.WitherStaticBody"},
	"Cascade":        {Body: game.CascadeStaticBody, VarName: "game.CascadeStaticBody"},
	"Convoke":        {Body: game.ConvokeStaticBody, VarName: "game.ConvokeStaticBody"},
}
