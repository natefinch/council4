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
		abilityBodies, ok := executableKeywordAbility(ability, compilation.Syntax.Abilities[i])
		if !ok {
			unsupported = append(unsupported, oracle.Diagnostic{
				Severity: oracle.SeverityWarning,
				Summary:  "unsupported executable ability",
				Detail:   "the executable source backend currently supports only plain non-parameterized keyword abilities",
				Span:     ability.Span,
			})
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

func executableKeywordAbility(ability oracle.CompiledAbility, syntax oracle.Ability) ([]string, bool) {
	if ability.Kind != oracle.AbilityStatic ||
		ability.AbilityWord != "" ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Modes) > 0 ||
		len(ability.Targets) > 0 ||
		len(ability.Conditions) > 0 ||
		len(ability.Effects) > 0 ||
		len(ability.References) > 0 ||
		len(ability.Keywords) == 0 {
		return nil, false
	}
	for _, token := range syntax.Tokens {
		if token.Kind == oracle.Comma ||
			spanCoveredByKeyword(token.Span, ability.Keywords) ||
			spanCoveredByDelimited(token.Span, syntax.Reminders) {
			continue
		}
		return nil, false
	}
	bodies := make([]string, 0, len(ability.Keywords))
	for _, keyword := range ability.Keywords {
		if keyword.Parameter != "" {
			return nil, false
		}
		body, ok := keywordStaticBodies[keyword.Name]
		if !ok {
			return nil, false
		}
		bodies = append(bodies, body)
	}
	return bodies, true
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
