package rules

import (
	"iter"
	"reflect"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/natefinch/council4/mtg/game"
)

// abilityParagraphKind classifies an Oracle paragraph by the kind of ability
// its wording introduces.
type abilityParagraphKind int

const (
	paragraphOther abilityParagraphKind = iota
	paragraphTriggered
	paragraphActivated
	paragraphMana
	paragraphLoyalty
	paragraphChapter
)

var (
	chapterParagraphPattern = regexp.MustCompile(`^[IVX]+(, [IVX]+)* — `)
	loyaltyParagraphPattern = regexp.MustCompile(`^[+−-]?(\d+|X): `)
	abilityWordPattern      = regexp.MustCompile(`^[^:.,(—"]+? — `)
)

// printedAbilityParagraphs returns the Oracle paragraph of each untexted
// triggered, activated, loyalty and chapter ability of face whose paragraph can
// be identified, keyed by canonical ability index. Each paragraph of rules text
// is a separate ability (CR 113.2c), so an untexted ability's own text is one
// paragraph, never the face's whole Oracle text. Keyword abilities strung
// together on one line are separate abilities too (CR 113.2c, 702), so a
// keyword ability's own text is its keyword, not the rest of its line.
//
// Paragraphs and keywords that are some ability's own recorded text, with or
// without their reminder text, are set aside. When a single paragraph remains,
// it contains every remaining untexted ability. Past that, untexted abilities of
// one kind are paired in order with the paragraphs whose wording introduces that
// kind only when exactly one pairing exists (pairAbilityParagraphs): a trigger
// word, an activation cost, a loyalty cost, chapter numerals, or a keyword the
// rules define as a triggered or activated ability (keywordParagraphs). A
// paragraph naming several trigger events is the text of each triggered ability
// it is modeled as. A triggered pairing must also agree on whether the trigger
// is an "At" trigger, and an ability tagged with a keyword pairs only with that
// keyword's paragraph. The return of an "until
// this leaves the battlefield" exile is part of the exiling ability's effect
// (CR 610.3), so it shares that ability's paragraph. Any kind whose paragraphs
// cannot be paired is left out: such an ability has no identifiable printed
// text, which is preferable to describing it with unrelated abilities.
func printedAbilityParagraphs(face *game.CardFace, oracleText string) map[int]string {
	if face == nil {
		return nil
	}
	texted := make(map[string]bool)
	var untexted []int
	for index := range face.AbilityCount() {
		body := face.BodyAt(index)
		if text := oracleRulesText(game.BodyText(body)); text != "" {
			texted[strings.ToLower(text)] = true
			continue
		}
		switch body.(type) {
		case *game.TriggeredAbility, *game.ActivatedAbility, *game.LoyaltyAbility, *game.ChapterAbility:
			untexted = append(untexted, index)
		}
	}
	if len(untexted) == 0 {
		return nil
	}
	units := abilityParagraphUnits(face, oracleText, texted)
	if len(units) == 0 {
		return nil
	}
	result := make(map[int]string, len(untexted))
	if len(units) == 1 {
		for _, index := range untexted {
			result[index] = units[0].text
		}
		return result
	}
	byKind := make(map[abilityParagraphKind][]abilityParagraphUnit)
	for _, unit := range units {
		byKind[unit.kind] = append(byKind[unit.kind], unit)
	}
	abilities := make(map[abilityParagraphKind][]int)
	for _, index := range untexted {
		body := face.BodyAt(index)
		if len(game.BodyKeywordAbilities(body)) > 0 && !pairsWithKeyword(body) {
			continue
		}
		kind := untextedAbilityParagraphKind(body)
		abilities[kind] = append(abilities[kind], index)
	}
	for kind, indexes := range abilities {
		if kind == paragraphOther {
			continue
		}
		for index, text := range pairAbilityParagraphs(face, indexes, byKind[kind]) {
			result[index] = text
		}
	}
	return result
}

// abilityParagraphUnit is one paragraph of Oracle text, or one keyword of a
// line of keywords, that has not been claimed as an ability's own text.
type abilityParagraphUnit struct {
	text    string
	kind    abilityParagraphKind
	at      bool
	keyword *keywordParagraph
}

// abilityParagraphUnits splits oracleText into its unclaimed paragraphs, with
// each keyword of a recognized keyword line as its own unit.
func abilityParagraphUnits(face *game.CardFace, oracleText string, texted map[string]bool) []abilityParagraphUnit {
	isTexted := func(text string) bool {
		return texted[strings.ToLower(text)] || texted[strings.ToLower(withoutReminderText(text))]
	}
	var units []abilityParagraphUnit
	for paragraph := range oracleParagraphs(oracleText) {
		if isTexted(paragraph) {
			continue
		}
		if items, ok := keywordLineItems(paragraph, isTexted); ok {
			for _, item := range items {
				if !isTexted(item) {
					units = append(units, newAbilityParagraphUnit(face, capitalized(item)))
				}
			}
			continue
		}
		units = append(units, newAbilityParagraphUnit(face, paragraph))
	}
	return units
}

func newAbilityParagraphUnit(face *game.CardFace, text string) abilityParagraphUnit {
	if keyword, ok := paragraphKeyword(text); ok {
		if !keyword.introducesAbility(face) {
			return abilityParagraphUnit{text: text, kind: paragraphOther, keyword: keyword}
		}
		return abilityParagraphUnit{text: text, kind: keyword.kind, at: keyword.at, keyword: keyword}
	}
	kind := classifyAbilityParagraph(text)
	return abilityParagraphUnit{text: text, kind: kind, at: strings.HasPrefix(withoutAbilityWord(text), "At ")}
}

// pairAbilityParagraphs pairs the untexted abilities at indexes, in order,
// with units, or returns nil when there is no single pairing. Each unit takes
// one ability, except that a trigger paragraph naming several trigger events
// ("When this creature enters or is turned face up, ...") takes the run of
// consecutive triggered abilities with that one effect that the engine models
// it as. A triggered pairing must agree on whether the trigger is an "At"
// trigger, and an ability tagged with a keyword pairs only with that keyword's
// line.
func pairAbilityParagraphs(face *game.CardFace, indexes []int, units []abilityParagraphUnit) map[int]string {
	if len(units) == 0 || len(units) > len(indexes) {
		return nil
	}
	// ways[u][a] counts, up to 2, the pairings of units[u:] with indexes[a:].
	ways := make([][]int, len(units)+1)
	for u := range ways {
		ways[u] = make([]int, len(indexes)+1)
	}
	ways[len(units)][len(indexes)] = 1
	for u := len(units) - 1; u >= 0; u-- {
		for a := len(indexes) - 1; a >= 0; a-- {
			for size := range groupSizes(face, indexes[a:], units[u]) {
				ways[u][a] = min(2, ways[u][a]+ways[u+1][a+size])
			}
		}
	}
	if ways[0][0] != 1 {
		return nil
	}
	result := make(map[int]string, len(indexes))
	a := 0
	for u := range units {
		for size := range groupSizes(face, indexes[a:], units[u]) {
			if ways[u+1][a+size] == 1 {
				for _, index := range indexes[a : a+size] {
					result[index] = units[u].text
				}
				a += size
				break
			}
		}
	}
	return result
}

// groupSizes yields how many of the abilities at indexes, from the first, unit
// can be the text of.
func groupSizes(face *game.CardFace, indexes []int, unit abilityParagraphUnit) iter.Seq[int] {
	return func(yield func(int) bool) {
		if len(indexes) == 0 || !unitMatchesAbility(unit, face.BodyAt(indexes[0])) || !yield(1) {
			return
		}
		condition, ok := unit.multipleTriggerCondition()
		if !ok {
			return
		}
		atEvents := strings.Contains(condition, " and at ")
		first, ok := face.BodyAt(indexes[0]).(*game.TriggeredAbility)
		if !ok {
			return
		}
		for size := 2; size <= len(indexes); size++ {
			next, ok := face.BodyAt(indexes[size-1]).(*game.TriggeredAbility)
			if !ok || len(next.KeywordAbilities) > 0 || (next.Trigger.Type == game.TriggerAt && !atEvents) ||
				!sameTriggeredEffect(first, next) || !yield(size) {
				return
			}
		}
	}
}

func unitMatchesAbility(unit abilityParagraphUnit, body game.Ability) bool {
	if ability, ok := body.(*game.TriggeredAbility); ok && unit.at != (ability.Trigger.Type == game.TriggerAt) {
		return false
	}
	if len(game.BodyKeywordAbilities(body)) > 0 {
		keyword := unit.keyword
		return keyword != nil && keyword.tag != game.KeywordNone && game.BodyHasKeyword(body, keyword.tag)
	}
	return true
}

// multipleTriggerCondition returns the trigger condition of a trigger
// paragraph that names more than one trigger event ("enters or leaves the
// battlefield", "enters and at the beginning of your upkeep").
func (u abilityParagraphUnit) multipleTriggerCondition() (string, bool) {
	if u.kind != paragraphTriggered || u.keyword != nil {
		return "", false
	}
	text := withoutAbilityWord(u.text)
	comma := topLevelIndex(text, ',')
	if comma < 0 {
		return "", false
	}
	condition := text[:comma]
	multiple := strings.Contains(condition, " or ") || strings.Contains(condition, " and at ") ||
		strings.Contains(condition, " and when")
	return condition, multiple
}

// sameTriggeredEffect reports whether two triggered abilities differ only in
// their trigger events, as the abilities of one paragraph naming several
// trigger events do.
func sameTriggeredEffect(a, b *game.TriggeredAbility) bool {
	return a.ZoneOfFunction == b.ZoneOfFunction && a.Optional == b.Optional &&
		a.ResolveWithIllegalTargets == b.ResolveWithIllegalTargets &&
		a.MaxTriggersPerTurn == b.MaxTriggersPerTurn &&
		a.CountsResolutionsThisTurn == b.CountsResolutionsThisTurn &&
		reflect.DeepEqual(a.Content, b.Content)
}

// oracleParagraphs yields the paragraphs of Oracle text, each with the "•"
// mode lines that belong to it (CR 700.2).
func oracleParagraphs(oracleText string) iter.Seq[string] {
	return func(yield func(string) bool) {
		var current []string
		for line := range strings.Lines(oracleText) {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "•") && len(current) > 0 {
				current = append(current, line)
				continue
			}
			if len(current) > 0 && !yield(strings.Join(current, "\n")) {
				return
			}
			current = []string{line}
		}
		if len(current) > 0 {
			yield(strings.Join(current, "\n"))
		}
	}
}

// printedAbilityParagraph returns the Oracle paragraph of the untexted ability
// at index on face, or "" when it cannot be identified.
func printedAbilityParagraph(face *game.CardFace, oracleText string, index int) string {
	return printedAbilityParagraphs(face, oracleText)[index]
}

// untextedAbilityParagraphKind returns the paragraph kind that introduces an
// untexted ability's own wording.
func untextedAbilityParagraphKind(body game.Ability) abilityParagraphKind {
	switch body.(type) {
	case *game.TriggeredAbility:
		return paragraphTriggered
	case *game.ActivatedAbility:
		return paragraphActivated
	case *game.LoyaltyAbility:
		return paragraphLoyalty
	case *game.ChapterAbility:
		return paragraphChapter
	}
	return paragraphOther
}

func classifyAbilityParagraph(paragraph string) abilityParagraphKind {
	switch {
	case chapterParagraphPattern.MatchString(paragraph):
		return paragraphChapter
	case loyaltyParagraphPattern.MatchString(paragraph):
		return paragraphLoyalty
	}
	body := withoutAbilityWord(paragraph)
	if startsWithTriggerWord(body) {
		return paragraphTriggered
	}
	if colon := topLevelIndex(body, ':'); colon >= 0 {
		if strings.HasPrefix(strings.TrimSpace(body[colon+1:]), "Add ") {
			return paragraphMana
		}
		return paragraphActivated
	}
	return paragraphOther
}

func startsWithTriggerWord(text string) bool {
	return strings.HasPrefix(text, "When ") || strings.HasPrefix(text, "Whenever ") || strings.HasPrefix(text, "At ")
}

// withoutAbilityWord strips an italicized ability word ("Landfall — ") that
// has no rules meaning of its own (CR 207.2c).
func withoutAbilityWord(paragraph string) string {
	if startsWithTriggerWord(paragraph) {
		return paragraph
	}
	if prefix := abilityWordPattern.FindString(paragraph); prefix != "" {
		return paragraph[len(prefix):]
	}
	return paragraph
}

// topLevelIndex returns the byte index of the first r outside quotation marks
// and parenthesized reminder text, or -1.
func topLevelIndex(text string, r rune) int {
	quoted := false
	depth := 0
	for i, c := range text {
		switch {
		case c == '"' || c == '“' || c == '”':
			quoted = !quoted
		case c == '(':
			depth++
		case c == ')' && depth > 0:
			depth--
		case c == r && !quoted && depth == 0:
			return i
		}
	}
	return -1
}

// delayedTriggerOwnText returns the sentence of a creating ability's text that
// is an event-based delayed triggered ability's own wording, through the end of
// its paragraph ("... put a +1/+1 counter on target creature. Whenever that
// creature attacks the monarch this turn, it gains double strike and trample
// until end of turn."). A delayed triggered ability has only its own text (CR
// 603.7a), not the clause of the ability that created it. It reports false
// unless exactly one later sentence of a paragraph begins with a trigger word,
// so wording that cannot be singled out is kept whole.
func delayedTriggerOwnText(text string) (string, bool) {
	var found string
	count := 0
	for line := range strings.Lines(text) {
		line = strings.TrimSpace(line)
		for start := range laterSentenceStarts(line) {
			sentence := line[start:]
			if strings.HasPrefix(sentence, "When ") || strings.HasPrefix(sentence, "Whenever ") {
				found = sentence
				count++
			}
		}
	}
	return found, count == 1
}

// laterSentenceStarts yields the byte offsets of the sentences after the first
// in paragraph, ignoring sentence ends inside quotation marks and reminder text.
func laterSentenceStarts(paragraph string) iter.Seq[int] {
	return func(yield func(int) bool) {
		quoted := false
		depth := 0
		for i, c := range paragraph {
			switch {
			case c == '"' || c == '“' || c == '”':
				quoted = !quoted
			case c == '(':
				depth++
			case c == ')' && depth > 0:
				depth--
			case c == '.' && !quoted && depth == 0 && strings.HasPrefix(paragraph[i:], ". ") && i+2 < len(paragraph):
				if !yield(i + 2) {
					return
				}
			}
		}
	}
}

// keywordParagraph is a keyword the rules define as a single triggered or
// activated ability (CR 702), whose printed keyword line is that ability's own
// text. Keywords that stand for more than one triggered or activated ability
// (champion, soulbond, suspend, haunt) are not listed, since their line cannot
// be paired with one ability. tag is the keyword's engine identity, if any.
type keywordParagraph struct {
	name string
	kind abilityParagraphKind
	at   bool
	tag  game.Keyword
}

// keywordParagraphs lists the keyword abilities whose line introduces exactly
// one triggered or activated ability, by CR 702 rule.
var keywordParagraphs = []keywordParagraph{
	{name: "ward", kind: paragraphTriggered, tag: game.Ward},                                    // 702.21a
	{name: "equip", kind: paragraphActivated, tag: game.Equip},                                  // 702.6a
	{name: "cumulative upkeep", kind: paragraphTriggered, at: true, tag: game.CumulativeUpkeep}, // 702.24a
	{name: "echo", kind: paragraphTriggered, at: true, tag: game.Echo},                          // 702.30a
	{name: "rampage", kind: paragraphTriggered, tag: game.Rampage},                              // 702.23a
	{name: "flanking", kind: paragraphTriggered, tag: game.Flanking},                            // 702.25a
	{name: "fortify", kind: paragraphActivated},                                                 // 702.67a
	{name: "provoke", kind: paragraphTriggered},                                                 // 702.39a
	{name: "storm", kind: paragraphTriggered, tag: game.Storm},                                  // 702.40a
	{name: "modular", kind: paragraphTriggered},                                                 // 702.43a
	{name: "bushido", kind: paragraphTriggered},                                                 // 702.45a
	{name: "soulshift", kind: paragraphTriggered, tag: game.Soulshift},                          // 702.46a
	{name: "ninjutsu", kind: paragraphActivated, tag: game.Ninjutsu},                            // 702.49a
	{name: "commander ninjutsu", kind: paragraphActivated, tag: game.Ninjutsu},                  // 702.49d
	{name: "transmute", kind: paragraphActivated, tag: game.Transmute},                          // 702.53a
	{name: "recover", kind: paragraphTriggered},                                                 // 702.59a
	{name: "ripple", kind: paragraphTriggered},                                                  // 702.60a
	{name: "aura swap", kind: paragraphActivated},                                               // 702.65a
	{name: "frenzy", kind: paragraphTriggered},                                                  // 702.68a
	{name: "gravestorm", kind: paragraphTriggered},                                              // 702.69a
	{name: "poisonous", kind: paragraphTriggered},                                               // 702.70a
	{name: "transfigure", kind: paragraphActivated},                                             // 702.71a
	{name: "graft", kind: paragraphTriggered},                                                   // 702.58a
	{name: "hideaway", kind: paragraphTriggered, tag: game.Hideaway},                            // 702.75a
	{name: "persist", kind: paragraphTriggered, tag: game.Persist},                              // 702.79a
	{name: "reinforce", kind: paragraphActivated},                                               // 702.77a
	{name: "exalted", kind: paragraphTriggered, tag: game.Exalted},                              // 702.83a
	{name: "unearth", kind: paragraphActivated, tag: game.Unearth},                              // 702.84a
	{name: "cascade", kind: paragraphTriggered, tag: game.Cascade},                              // 702.85a
	{name: "annihilator", kind: paragraphTriggered, tag: game.Annihilator},                      // 702.86a
	{name: "level up", kind: paragraphActivated},                                                // 702.87a
	{name: "battle cry", kind: paragraphTriggered},                                              // 702.91a
	{name: "living weapon", kind: paragraphTriggered, tag: game.LivingWeapon},                   // 702.92a
	{name: "undying", kind: paragraphTriggered, tag: game.Undying},                              // 702.93a
	{name: "evolve", kind: paragraphTriggered, tag: game.Evolve},                                // 702.100a
	{name: "extort", kind: paragraphTriggered},                                                  // 702.101a
	{name: "scavenge", kind: paragraphActivated, tag: game.Scavenge},                            // 702.97a
	{name: "outlast", kind: paragraphActivated, tag: game.Outlast},                              // 702.107a
	{name: "dethrone", kind: paragraphTriggered, tag: game.Dethrone},                            // 702.105a
	{name: "prowess", kind: paragraphTriggered, tag: game.Prowess},                              // 702.108a
	{name: "exploit", kind: paragraphTriggered},                                                 // 702.110a
	{name: "renown", kind: paragraphTriggered},                                                  // 702.112a
	{name: "ingest", kind: paragraphTriggered},                                                  // 702.115a
	{name: "myriad", kind: paragraphTriggered, tag: game.Myriad},                                // 702.116a
	{name: "melee", kind: paragraphTriggered},                                                   // 702.121a
	{name: "crew", kind: paragraphActivated, tag: game.Crew},                                    // 702.122a
	{name: "fabricate", kind: paragraphTriggered, tag: game.Fabricate},                          // 702.123a
	{name: "partner with", kind: paragraphTriggered, tag: game.PartnerWith},                     // 702.124j
	{name: "embalm", kind: paragraphActivated, tag: game.Embalm},                                // 702.128a
	{name: "eternalize", kind: paragraphActivated, tag: game.Eternalize},                        // 702.129a
	{name: "afflict", kind: paragraphTriggered},                                                 // 702.130a
	{name: "afterlife", kind: paragraphTriggered},                                               // 702.135a
	{name: "mentor", kind: paragraphTriggered},                                                  // 702.134a
	{name: "encore", kind: paragraphActivated},                                                  // 702.141a
	{name: "training", kind: paragraphTriggered, tag: game.Training},                            // 702.149a
	{name: "demonstrate", kind: paragraphTriggered},                                             // 702.144a
	{name: "decayed", kind: paragraphTriggered},                                                 // 702.147a
	{name: "backup", kind: paragraphTriggered},                                                  // 702.165a
	{name: "for mirrodin!", kind: paragraphTriggered},                                           // 702.163a
	{name: "craft", kind: paragraphActivated, tag: game.Craft},                                  // 702.167a
	{name: "saddle", kind: paragraphActivated, tag: game.Saddle},                                // 702.171a
	{name: "mobilize", kind: paragraphTriggered, tag: game.Mobilize},                            // 702.181a
	{name: "station", kind: paragraphActivated},                                                 // 702.184a
}

// cyclingKeywordParagraph is the cycling family: cycling and typecycling
// ("Swampcycling", "Basic landcycling") are each one activated ability
// (CR 702.29a, 702.29e).
var cyclingKeywordParagraph = keywordParagraph{name: "cycling", kind: paragraphActivated, tag: game.Cycling}

// cyclingHeadPattern matches the lowercase head of a cycling-family line.
var cyclingHeadPattern = regexp.MustCompile(`^([a-z]+ )?[a-z]*cycling( |—|$)`)

// paragraphKeyword reports the keyword whose line text is, ignoring reminder
// text. A keyword line has no activation cost colon, does not begin with a
// trigger word, and is not a sentence unless its cost follows an em dash
// ("Equip—Sacrifice a creature.").
func paragraphKeyword(text string) (*keywordParagraph, bool) {
	head := withoutReminderText(text)
	if head == "" || startsWithTriggerWord(head) || topLevelIndex(head, ':') >= 0 {
		return nil, false
	}
	lower := strings.ToLower(head)
	keyword, rest := (*keywordParagraph)(nil), ""
	if match := cyclingHeadPattern.FindString(lower); match != "" {
		keyword, rest = &cyclingKeywordParagraph, lower[len(match):]
		if strings.HasSuffix(match, "—") {
			rest = "—" + rest
		}
	} else {
		for i := range keywordParagraphs {
			after, ok := strings.CutPrefix(lower, keywordParagraphs[i].name)
			if ok && (after == "" || strings.HasPrefix(after, " ") || strings.HasPrefix(after, "—")) {
				keyword, rest = &keywordParagraphs[i], after
				break
			}
		}
	}
	if keyword == nil || (strings.HasSuffix(head, ".") && !strings.HasPrefix(rest, "—")) {
		return nil, false
	}
	return keyword, true
}

// introducesAbility reports whether a keyword line on face stands for an
// untexted ability. A keyword with an engine identity that face carries only on
// some other ability (for example a static ward ability with its own text, or
// an alternative cost) is that ability's line rather than an untexted one.
func (k *keywordParagraph) introducesAbility(face *game.CardFace) bool {
	if k.tag == game.KeywordNone {
		return true
	}
	carried := false
	for index := range face.AbilityCount() {
		body := face.BodyAt(index)
		if !game.BodyHasKeyword(body, k.tag) {
			continue
		}
		if untextedAbilityParagraphKind(body) == k.kind && oracleRulesText(game.BodyText(body)) == "" {
			return true
		}
		carried = true
	}
	return !carried
}

// pairsWithKeyword reports whether an ability tagged with keywords can pair
// with some keyword line at all. Tags the keyword lines do not describe keep
// the ability out of pairing.
func pairsWithKeyword(body game.Ability) bool {
	for _, keyword := range game.BodyKeywordAbilities(body) {
		tag := game.KeywordAbilityKind(keyword)
		if tag == cyclingKeywordParagraph.tag {
			return true
		}
		for i := range keywordParagraphs {
			if keywordParagraphs[i].tag == tag {
				return true
			}
		}
	}
	return false
}

// keywordLineItems splits a line of keywords separated by commas ("Flying,
// annihilator 2") into its keywords, each its own ability (CR 113.2c), when
// every item is a keyword line or an ability's recorded text. Reminder text
// stays with the keyword it follows.
func keywordLineItems(paragraph string, isTexted func(string) bool) ([]string, bool) {
	var items []string
	rest := paragraph
	for {
		comma := topLevelIndex(rest, ',')
		if comma < 0 {
			items = append(items, strings.TrimSpace(rest))
			break
		}
		items = append(items, strings.TrimSpace(rest[:comma]))
		rest = rest[comma+1:]
	}
	if len(items) < 2 {
		return nil, false
	}
	for _, item := range items {
		if _, ok := paragraphKeyword(item); !ok && !isTexted(capitalized(item)) {
			return nil, false
		}
	}
	return items, true
}

// withoutReminderText returns text without its parenthesized reminder text
// (CR 207.2a).
func withoutReminderText(text string) string {
	var b strings.Builder
	depth := 0
	for _, c := range text {
		switch {
		case c == '(':
			depth++
		case c == ')' && depth > 0:
			depth--
		case depth == 0:
			b.WriteRune(c)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func capitalized(text string) string {
	if text == "" {
		return text
	}
	r, size := utf8.DecodeRuneInString(text)
	return string(unicode.ToUpper(r)) + text[size:]
}
