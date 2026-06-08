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

	var bodies []string
	var unsupported []oracle.Diagnostic
	for i, ability := range compilation.Abilities {
		abilityBodies, diagnostic := executableKeywordAbility(ability, compilation.Syntax.Abilities[i])
		if diagnostic != nil {
			unsupported = append(unsupported, *diagnostic)
			continue
		}
		bodies = append(bodies, abilityBodies...)
	}
	if len(unsupported) > 0 {
		return nil, append(diagnostics, unsupported...)
	}
	if len(bodies) == 0 {
		return nil, diagnostics
	}
	return []string{staticAbilityField(bodies)}, diagnostics
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
	if ability.Kind != oracle.AbilityStatic {
		return nil, executableDiagnostic(
			ability,
			"unsupported "+ability.Kind.String()+" ability",
			"the executable source backend does not yet lower "+ability.Kind.String()+" abilities",
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
