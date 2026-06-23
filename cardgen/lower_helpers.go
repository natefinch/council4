package cardgen

import (
	"fmt"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

func targetCardinalityIsOne(target compiler.CompiledTarget) bool {
	return target.Cardinality.Min == 1 && target.Cardinality.Max == 1
}

func signedAmountText(amount compiler.CompiledSignedAmount) string {
	if amount.Negative {
		return fmt.Sprintf("-%d", amount.Value)
	}
	return fmt.Sprintf("+%d", amount.Value)
}

func compiledSignedAmountValue(amount compiler.CompiledSignedAmount) int {
	if amount.Negative {
		return -amount.Value
	}
	return amount.Value
}

func titleFirst(text string) string {
	if text == "" {
		return ""
	}
	return strings.ToUpper(text[:1]) + text[1:]
}

func lowerFirst(text string) string {
	if text == "" {
		return ""
	}
	return strings.ToLower(text[:1]) + text[1:]
}

func spanCoveredByKeyword(span shared.Span, keywords []compiler.CompiledKeyword) bool {
	for _, keyword := range keywords {
		if keyword.Span.Start.Offset <= span.Start.Offset &&
			keyword.Span.End.Offset >= span.End.Offset {
			return true
		}
	}
	return false
}

func spanCoveredByAbilityWord(span shared.Span, abilityWord *parser.AbilityWordClause) bool {
	return abilityWord != nil &&
		abilityWord.Span.Start.Offset <= span.Start.Offset &&
		abilityWord.Span.End.Offset >= span.End.Offset
}

func spanCoveredByDelimited(span shared.Span, groups []parser.Delimited) bool {
	for _, group := range groups {
		if group.Span.Start.Offset <= span.Start.Offset &&
			group.Span.End.Offset >= span.End.Offset {
			return true
		}
	}
	return false
}

// spanIsKeywordListSemicolon reports whether span is a semicolon token, the
// structural separator some older keyword lines use between keywords (e.g.
// "First strike; reach"). The parser already drops commas, colons, and periods
// from CoverageSpans as punctuation it owns; the semicolon keyword-list
// separator is treated the same way in the keyword-only ability path, where the
// surrounding keyword words remain must-cover so unrecognized content still
// fails closed.
func spanIsKeywordListSemicolon(span shared.Span, tokens []shared.Token) bool {
	for _, token := range tokens {
		if token.Span == span {
			return token.Kind == shared.Semicolon
		}
	}
	return false
}

// appendKeywordListSemicolonSpans appends the spans of semicolon separator
// tokens so the completeness gate credits them, mirroring spanIsKeywordListSemicolon
// in the keyword-coverage gate.
func appendKeywordListSemicolonSpans(spans []shared.Span, tokens []shared.Token) []shared.Span {
	for _, token := range tokens {
		if token.Kind == shared.Semicolon {
			spans = append(spans, token.Span)
		}
	}
	return spans
}

// keywordOnlyCovered reports whether every must-cover token span of the ability
// is accounted for by the keyword's span or a reminder, i.e. the ability is
// exactly this one keyword with no other rules text. It consumes the parser's
// must-cover assertion (CoverageSpans) rather than walking the raw token stream.
func keywordOnlyCovered(syntax *parser.Ability, keyword compiler.CompiledKeyword) bool {
	for _, span := range syntax.CoverageSpans() {
		if spanCovered(span, []shared.Span{keyword.Span}) ||
			spanCoveredByDelimited(span, syntax.Reminders) {
			continue
		}
		return false
	}
	return true
}

func executableDiagnostic(
	ability compiler.CompiledAbility,
	summary string,
	detail string,
) *shared.Diagnostic {
	return &shared.Diagnostic{
		Severity: shared.SeverityWarning,
		Summary:  summary,
		Detail:   detail,
		Span:     ability.Span,
	}
}

func mixedKeywordDiagnostic(ctx contentCtx) *shared.Diagnostic {
	names := make([]string, 0, len(ctx.content.Keywords))
	for _, keyword := range ctx.content.Keywords {
		names = append(names, keyword.Name)
	}
	return contentDiagnostic(
		ctx,
		"unsupported mixed keyword ability",
		fmt.Sprintf(
			"the executable source backend recognized %s but does not yet lower the additional rules text",
			strings.Join(names, ", "),
		),
	)
}

// keywordStaticBodies maps a typed keyword to its reusable typed StaticAbility and
// the package-level variable reference the Renderer emits for it.
var keywordStaticBodies = map[parser.KeywordKind]loweredStaticAbility{
	parser.KeywordBanding:          {Body: game.BandingStaticBody, VarName: "game.BandingStaticBody"},
	parser.KeywordDevoid:           {Body: game.DevoidStaticBody, VarName: "game.DevoidStaticBody"},
	parser.KeywordDeathtouch:       {Body: game.DeathtouchStaticBody, VarName: "game.DeathtouchStaticBody"},
	parser.KeywordDefender:         {Body: game.DefenderStaticBody, VarName: "game.DefenderStaticBody"},
	parser.KeywordDelve:            {Body: game.DelveStaticBody, VarName: "game.DelveStaticBody"},
	parser.KeywordDoubleStrike:     {Body: game.DoubleStrikeStaticBody, VarName: "game.DoubleStrikeStaticBody"},
	parser.KeywordExalted:          {Body: game.ExaltedStaticBody, VarName: "game.ExaltedStaticBody"},
	parser.KeywordEvolve:           {Body: game.EvolveStaticBody, VarName: "game.EvolveStaticBody"},
	parser.KeywordFear:             {Body: game.FearStaticBody, VarName: "game.FearStaticBody"},
	parser.KeywordFirstStrike:      {Body: game.FirstStrikeStaticBody, VarName: "game.FirstStrikeStaticBody"},
	parser.KeywordFlash:            {Body: game.FlashStaticBody, VarName: "game.FlashStaticBody"},
	parser.KeywordFlying:           {Body: game.FlyingStaticBody, VarName: "game.FlyingStaticBody"},
	parser.KeywordHaste:            {Body: game.HasteStaticBody, VarName: "game.HasteStaticBody"},
	parser.KeywordHexproof:         {Body: game.HexproofStaticBody, VarName: "game.HexproofStaticBody"},
	parser.KeywordHorsemanship:     {Body: game.HorsemanshipStaticBody, VarName: "game.HorsemanshipStaticBody"},
	parser.KeywordImprovise:        {Body: game.ImproviseStaticBody, VarName: "game.ImproviseStaticBody"},
	parser.KeywordIndestructible:   {Body: game.IndestructibleStaticBody, VarName: "game.IndestructibleStaticBody"},
	parser.KeywordInfect:           {Body: game.InfectStaticBody, VarName: "game.InfectStaticBody"},
	parser.KeywordIntimidate:       {Body: game.IntimidateStaticBody, VarName: "game.IntimidateStaticBody"},
	parser.KeywordLifelink:         {Body: game.LifelinkStaticBody, VarName: "game.LifelinkStaticBody"},
	parser.KeywordMenace:           {Body: game.MenaceStaticBody, VarName: "game.MenaceStaticBody"},
	parser.KeywordPersist:          {Body: game.PersistStaticBody, VarName: "game.PersistStaticBody"},
	parser.KeywordProwess:          {Body: game.ProwessStaticBody, VarName: "game.ProwessStaticBody"},
	parser.KeywordReadAhead:        {Body: game.ReadAheadStaticBody, VarName: "game.ReadAheadStaticBody"},
	parser.KeywordReach:            {Body: game.ReachStaticBody, VarName: "game.ReachStaticBody"},
	parser.KeywordRetrace:          {Body: game.RetraceStaticBody, VarName: "game.RetraceStaticBody"},
	parser.KeywordShadow:           {Body: game.ShadowStaticBody, VarName: "game.ShadowStaticBody"},
	parser.KeywordShroud:           {Body: game.ShroudStaticBody, VarName: "game.ShroudStaticBody"},
	parser.KeywordSkulk:            {Body: game.SkulkStaticBody, VarName: "game.SkulkStaticBody"},
	parser.KeywordSplitSecond:      {Body: game.SplitSecondStaticBody, VarName: "game.SplitSecondStaticBody"},
	parser.KeywordStorm:            {Body: game.StormStaticBody, VarName: "game.StormStaticBody"},
	parser.KeywordTrample:          {Body: game.TrampleStaticBody, VarName: "game.TrampleStaticBody"},
	parser.KeywordUndying:          {Body: game.UndyingStaticBody, VarName: "game.UndyingStaticBody"},
	parser.KeywordVigilance:        {Body: game.VigilanceStaticBody, VarName: "game.VigilanceStaticBody"},
	parser.KeywordWither:           {Body: game.WitherStaticBody, VarName: "game.WitherStaticBody"},
	parser.KeywordRiot:             {Body: game.RiotStaticBody, VarName: "game.RiotStaticBody"},
	parser.KeywordUnleash:          {Body: game.UnleashStaticBody, VarName: "game.UnleashStaticBody"},
	parser.KeywordCascade:          {Body: game.CascadeStaticBody, VarName: "game.CascadeStaticBody"},
	parser.KeywordConvoke:          {Body: game.ConvokeStaticBody, VarName: "game.ConvokeStaticBody"},
	parser.KeywordChangeling:       {Body: game.ChangelingStaticBody, VarName: "game.ChangelingStaticBody"},
	parser.KeywordLandwalk:         {Body: game.LandwalkStaticBody, VarName: "game.LandwalkStaticBody"},
	parser.KeywordPlainswalk:       {Body: game.PlainswalkStaticBody, VarName: "game.PlainswalkStaticBody"},
	parser.KeywordIslandwalk:       {Body: game.IslandwalkStaticBody, VarName: "game.IslandwalkStaticBody"},
	parser.KeywordSwampwalk:        {Body: game.SwampwalkStaticBody, VarName: "game.SwampwalkStaticBody"},
	parser.KeywordMountainwalk:     {Body: game.MountainwalkStaticBody, VarName: "game.MountainwalkStaticBody"},
	parser.KeywordForestwalk:       {Body: game.ForestwalkStaticBody, VarName: "game.ForestwalkStaticBody"},
	parser.KeywordDesertwalk:       {Body: game.DesertwalkStaticBody, VarName: "game.DesertwalkStaticBody"},
	parser.KeywordNonbasicLandwalk: {Body: game.NonbasicLandwalkStaticBody, VarName: "game.NonbasicLandwalkStaticBody"},
	parser.KeywordRebound:          {Body: game.ReboundStaticBody, VarName: "game.ReboundStaticBody"},
}
