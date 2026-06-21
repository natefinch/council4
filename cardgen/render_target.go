package cardgen

import (
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func (r Renderer) renderTargetSpec(ctx *renderCtx, spec *game.TargetSpec) (string, error) {
	fields := []string{
		fmt.Sprintf("MinTargets: %d,", spec.MinTargets),
		fmt.Sprintf("MaxTargets: %d,", spec.MaxTargets),
	}
	if spec.Constraint != "" {
		fields = append(fields, fmt.Sprintf("Constraint: %q,", spec.Constraint))
	}
	if spec.Allow != game.TargetAllowUnspecified {
		fields = append(fields, fmt.Sprintf("Allow: %s,", renderTargetAllow(spec.Allow)))
	}
	if spec.TargetZone != zone.None {
		targetZone, err := renderZone(spec.TargetZone)
		if err != nil {
			return "", err
		}
		ctx.need(importZone)
		fields = append(fields, fmt.Sprintf("TargetZone: %s,", targetZone))
	}
	if spec.Selection.Exists {
		selection, err := r.renderSelection(ctx, spec.Selection.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Selection: opt.Val(%s),", selection))
	}
	if predicate, ok, err := r.renderTargetPredicate(ctx, spec.Predicate); err != nil {
		return "", err
	} else if ok {
		fields = append(fields, fmt.Sprintf("Predicate: %s,", predicate))
	}
	return structLit("game.TargetSpec", fields), nil
}

// appendSupertypeFields renders the shared Supertypes slice and the scalar
// ExcludedSupertype filter, which both TargetPredicate and Selection carry, onto
// the literal field list.
func appendSupertypeFields(ctx *renderCtx, fields []string, supertypes []types.Super, excluded types.Super) ([]string, error) {
	if len(supertypes) > 0 {
		ctx.need(importTypes)
		literals := make([]string, 0, len(supertypes))
		for _, st := range supertypes {
			lit, err := supertypeLiteral(st)
			if err != nil {
				return nil, err
			}
			literals = append(literals, lit)
		}
		fields = append(fields, fmt.Sprintf("Supertypes: []types.Super{%s},", strings.Join(literals, ", ")))
	}
	if excluded != "" {
		ctx.need(importTypes)
		lit, err := supertypeLiteral(excluded)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("ExcludedSupertype: %s,", lit))
	}
	return fields, nil
}

func appendCardTypePredicateField(
	ctx *renderCtx,
	fields []string,
	name string,
	values []types.Card,
) ([]string, error) {
	if len(values) == 0 {
		return fields, nil
	}
	ctx.need(importTypes)
	literals, err := renderTypesCardSlice(ctx, values)
	if err != nil {
		return nil, err
	}
	return append(fields, fmt.Sprintf("%s: %s,", name, literals)), nil
}

func (Renderer) renderTargetPredicate(ctx *renderCtx, predicate game.TargetPredicate) (lit string, ok bool, err error) {
	var fields []string
	if len(predicate.PermanentTypes) > 0 {
		ctx.need(importTypes)
		lits, err := renderTypesCardSlice(ctx, predicate.PermanentTypes)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("PermanentTypes: %s,", lits))
	}
	if predicate.PermanentTypesConjunctive {
		fields = append(fields, "PermanentTypesConjunctive: true,")
	}
	if len(predicate.ExcludedTypes) > 0 {
		lits, err := renderTypesCardSlice(ctx, predicate.ExcludedTypes)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("ExcludedTypes: %s,", lits))
	}
	if len(predicate.Supertypes) > 0 || predicate.ExcludedSupertype != "" {
		fields, err = appendSupertypeFields(ctx, fields, predicate.Supertypes, predicate.ExcludedSupertype)
		if err != nil {
			return "", false, err
		}
	}
	if len(predicate.Subtypes) > 0 {
		ctx.need(importTypes)
		literals := make([]string, 0, len(predicate.Subtypes))
		for _, sub := range predicate.Subtypes {
			literals = append(literals, SubtypeToLiteral(string(sub), nil))
		}
		fields = append(fields, fmt.Sprintf("Subtypes: []types.Sub{%s},", strings.Join(literals, ", ")))
	}
	for _, field := range []struct {
		name   string
		values []types.Card
	}{
		{name: "SpellCardTypes", values: predicate.SpellCardTypes},
		{name: "SpellCardTypesAny", values: predicate.SpellCardTypesAny},
		{name: "ExcludedSpellCardTypes", values: predicate.ExcludedSpellCardTypes},
	} {
		fields, err = appendCardTypePredicateField(ctx, fields, field.name, field.values)
		if err != nil {
			return "", false, err
		}
	}
	if len(predicate.StackObjectKinds) > 0 || len(predicate.StackObjectSourceTypes) > 0 ||
		len(predicate.SpellSupertypes) > 0 || predicate.SpellColorless ||
		len(predicate.SpellColors) > 0 || len(predicate.SpellExcludedColors) > 0 ||
		predicate.SpellMulticolored {
		stackFields, err := renderStackObjectPredicateFields(ctx, predicate)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, stackFields...)
	}
	if len(predicate.Colors) > 0 {
		colors, err := renderColorSlice(ctx, predicate.Colors)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("Colors: %s,", colors))
	}
	if len(predicate.ExcludedColors) > 0 {
		colors, err := renderColorSlice(ctx, predicate.ExcludedColors)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("ExcludedColors: %s,", colors))
	}
	if predicate.Player != game.PlayerAny {
		pr, err := renderPlayerRelation(predicate.Player)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("Player: %s,", pr))
	}
	if predicate.Controller != game.ControllerAny {
		cr, err := renderControllerRelation(predicate.Controller)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("Controller: %s,", cr))
	}
	if predicate.Tapped != game.TriAny {
		ts, err := renderTriState(predicate.Tapped)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("Tapped: %s,", ts))
	}
	if predicate.CombatState != game.CombatStateAny {
		cs, err := renderCombatStateFilter(predicate.CombatState)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("CombatState: %s,", cs))
	}
	if predicate.Keyword != game.KeywordNone {
		kw, err := renderKeyword(predicate.Keyword)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("Keyword: %s,", kw))
	}
	if predicate.ExcludedKeyword != game.KeywordNone {
		kw, err := renderKeyword(predicate.ExcludedKeyword)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("ExcludedKeyword: %s,", kw))
	}
	if predicate.ManaValue.Exists {
		ctx.need(importOpt)
		cmp, err := renderCompareInt(ctx, predicate.ManaValue.Val)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("ManaValue: opt.Val(%s),", cmp))
	}
	if predicate.Power.Exists {
		ctx.need(importOpt)
		cmp, err := renderCompareInt(ctx, predicate.Power.Val)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("Power: opt.Val(%s),", cmp))
	}
	if predicate.Toughness.Exists {
		ctx.need(importOpt)
		cmp, err := renderCompareInt(ctx, predicate.Toughness.Val)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("Toughness: opt.Val(%s),", cmp))
	}
	if predicate.Another {
		fields = append(fields, "Another: true,")
	}
	if len(fields) == 0 {
		return "", false, nil
	}
	return structLit("game.TargetPredicate", fields), true, nil
}

func renderStackObjectPredicateFields(ctx *renderCtx, predicate game.TargetPredicate) ([]string, error) {
	var fields []string
	if len(predicate.StackObjectKinds) > 0 {
		kinds, err := renderStackObjectKinds(predicate.StackObjectKinds)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("StackObjectKinds: %s,", kinds))
	}
	if len(predicate.StackObjectSourceTypes) > 0 {
		ctx.need(importTypes)
		lits, err := renderTypesCardSlice(ctx, predicate.StackObjectSourceTypes)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("StackObjectSourceTypes: %s,", lits))
	}
	if len(predicate.SpellSupertypes) > 0 {
		ctx.need(importTypes)
		literals := make([]string, 0, len(predicate.SpellSupertypes))
		for _, st := range predicate.SpellSupertypes {
			lit, err := supertypeLiteral(st)
			if err != nil {
				return nil, err
			}
			literals = append(literals, lit)
		}
		fields = append(fields, fmt.Sprintf("SpellSupertypes: []types.Super{%s},", strings.Join(literals, ", ")))
	}
	if predicate.SpellColorless {
		fields = append(fields, "SpellColorless: true,")
	}
	if len(predicate.SpellColors) > 0 {
		colors, err := renderColorSlice(ctx, predicate.SpellColors)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("SpellColors: %s,", colors))
	}
	if len(predicate.SpellExcludedColors) > 0 {
		colors, err := renderColorSlice(ctx, predicate.SpellExcludedColors)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("SpellExcludedColors: %s,", colors))
	}
	if predicate.SpellMulticolored {
		fields = append(fields, "SpellMulticolored: true,")
	}
	return fields, nil
}

func renderStackObjectKinds(kinds []game.StackObjectKind) (string, error) {
	lits := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		switch kind {
		case game.StackSpell:
			lits = append(lits, "game.StackSpell")
		case game.StackActivatedAbility:
			lits = append(lits, "game.StackActivatedAbility")
		case game.StackTriggeredAbility:
			lits = append(lits, "game.StackTriggeredAbility")
		default:
			return "", fmt.Errorf("render: unsupported stack-object kind %d", kind)
		}
	}
	return "[]game.StackObjectKind{" + strings.Join(lits, ", ") + "}", nil
}

func (r Renderer) renderGroupReference(ctx *renderCtx, group game.GroupReference) (string, error) {
	selection, err := r.renderSelection(ctx, group.Selection())
	if err != nil {
		return "", err
	}
	exclude, hasExclude := group.Exclusion()
	switch group.Domain() {
	case game.GroupDomainBattlefield:
		if hasExclude {
			rendered, err := r.renderObjectReference(exclude)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("game.BattlefieldGroupExcluding(%s, %s)", selection, rendered), nil
		}
		return fmt.Sprintf("game.BattlefieldGroup(%s)", selection), nil
	case game.GroupDomainAttachedObject:
		anchor, _ := group.Anchor()
		rendered, err := r.renderObjectReference(anchor)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.AttachedObjectGroup(%s)", rendered), nil
	case game.GroupDomainObjectControlled:
		anchor, _ := group.Anchor()
		renderedAnchor, err := r.renderObjectReference(anchor)
		if err != nil {
			return "", err
		}
		if hasExclude {
			renderedExclude, err := r.renderObjectReference(exclude)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("game.ObjectControlledGroupExcluding(%s, %s, %s)", renderedAnchor, selection, renderedExclude), nil
		}
		return fmt.Sprintf("game.ObjectControlledGroup(%s, %s)", renderedAnchor, selection), nil
	default:
		return "", fmt.Errorf("render: unsupported group reference domain %d", group.Domain())
	}
}

func (Renderer) renderSelection(ctx *renderCtx, selection game.Selection) (string, error) {
	var fields []string

	if len(selection.AnyOf) > 0 {
		alternatives := make([]string, 0, len(selection.AnyOf))
		for i := range selection.AnyOf {
			rendered, err := (Renderer{}).renderSelection(ctx, selection.AnyOf[i])
			if err != nil {
				return "", err
			}
			alternatives = append(alternatives, rendered)
		}
		fields = append(fields, fmt.Sprintf("AnyOf: []game.Selection{%s},", strings.Join(alternatives, ", ")))
	}
	if len(selection.RequiredTypes) > 0 {
		ctx.need(importTypes)
		lits, err := renderTypesCardSlice(ctx, selection.RequiredTypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("RequiredTypes: %s,", lits))
	}
	if len(selection.RequiredTypesAny) > 0 {
		ctx.need(importTypes)
		lits, err := renderTypesCardSlice(ctx, selection.RequiredTypesAny)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("RequiredTypesAny: %s,", lits))
	}
	if len(selection.ExcludedTypes) > 0 {
		ctx.need(importTypes)
		lits, err := renderTypesCardSlice(ctx, selection.ExcludedTypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ExcludedTypes: %s,", lits))
	}

	if len(selection.Supertypes) > 0 || selection.ExcludedSupertype != "" {
		var stErr error
		fields, stErr = appendSupertypeFields(ctx, fields, selection.Supertypes, selection.ExcludedSupertype)
		if stErr != nil {
			return "", stErr
		}
	}
	if len(selection.SubtypesAny) > 0 {
		ctx.need(importTypes)
		literals := make([]string, 0, len(selection.SubtypesAny))
		for _, sub := range selection.SubtypesAny {
			literals = append(literals, SubtypeToLiteral(string(sub), nil))
		}
		fields = append(fields, fmt.Sprintf("SubtypesAny: []types.Sub{%s},", strings.Join(literals, ", ")))
	}
	fields = appendSubtypeFromSourceEntryChoiceField(fields, selection.SubtypeFromSourceEntryChoice)
	if len(selection.ColorsAny) > 0 {
		colorLits, err := renderColorSlice(ctx, selection.ColorsAny)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ColorsAny: %s,", colorLits))
	}
	if len(selection.ExcludedColors) > 0 {
		colorLits, err := renderColorSlice(ctx, selection.ExcludedColors)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ExcludedColors: %s,", colorLits))
	}
	if selection.Colorless {
		fields = append(fields, "Colorless: true,")
	}
	if selection.Multicolored {
		fields = append(fields, "Multicolored: true,")
	}

	if selection.Controller != game.ControllerAny {
		cr, err := renderControllerRelation(selection.Controller)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Controller: %s,", cr))
	}
	if selection.Player != game.PlayerAny {
		pr, err := renderPlayerRelation(selection.Player)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Player: %s,", pr))
	}

	if selection.Tapped != game.TriAny {
		ts, err := renderTriState(selection.Tapped)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Tapped: %s,", ts))
	}
	if selection.CombatState != game.CombatStateAny {
		cs, err := renderCombatStateFilter(selection.CombatState)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("CombatState: %s,", cs))
	}

	if selection.Keyword != game.KeywordNone {
		kw, err := renderKeyword(selection.Keyword)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Keyword: %s,", kw))
	}
	if selection.ExcludedKeyword != game.KeywordNone {
		kw, err := renderKeyword(selection.ExcludedKeyword)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ExcludedKeyword: %s,", kw))
	}

	compareFields, err := renderSelectionComparisons(ctx, selection)
	if err != nil {
		return "", err
	}
	fields = append(fields, compareFields...)

	if selection.ExcludeSource {
		fields = append(fields, "ExcludeSource: true,")
	}
	if selection.NonToken {
		fields = append(fields, "NonToken: true,")
	}
	if selection.TokenOnly {
		fields = append(fields, "TokenOnly: true,")
	}

	for i := range fields {
		fields[i] = strings.TrimSuffix(fields[i], ",")
	}
	return compactStructLit("game.Selection", fields), nil
}

// renderSelectionComparisons renders the Selection fields that compare numeric
// characteristics (mana value, power, toughness) and the counter requirement,
// returning the rendered fields in declaration order.
func renderSelectionComparisons(ctx *renderCtx, selection game.Selection) ([]string, error) {
	var fields []string
	for _, c := range []struct {
		name  string
		value opt.V[compare.Int]
	}{
		{"ManaValue", selection.ManaValue},
		{"Power", selection.Power},
		{"Toughness", selection.Toughness},
	} {
		if !c.value.Exists {
			continue
		}
		ctx.need(importOpt)
		cmp, err := renderCompareInt(ctx, c.value.Val)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("%s: opt.Val(%s),", c.name, cmp))
	}
	if selection.MatchCounter {
		ctx.need(importCounter)
		kind, err := renderCounterKind(selection.RequiredCounter)
		if err != nil {
			return nil, err
		}
		fields = append(fields, "MatchCounter: true,", fmt.Sprintf("RequiredCounter: %s,", kind))
	}
	return fields, nil
}

// appendSubtypeFromSourceEntryChoiceField appends the Selection field that ties a
// group to the creature type chosen as the source permanent entered, leaving
// fields untouched when the restriction is absent.
func appendSubtypeFromSourceEntryChoiceField(fields []string, fromEntryChoice bool) []string {
	if !fromEntryChoice {
		return fields
	}
	return append(fields, "SubtypeFromSourceEntryChoice: true,")
}

func renderColorSlice(ctx *renderCtx, colors []color.Color) (string, error) {
	ctx.need(importColor)
	literals := make([]string, 0, len(colors))
	for _, c := range colors {
		lit, err := colorValueToLiteral(c)
		if err != nil {
			return "", err
		}
		literals = append(literals, lit)
	}
	return "[]color.Color{" + strings.Join(literals, ", ") + "}", nil
}

func renderColorArguments(ctx *renderCtx, colors []color.Color) (string, error) {
	ctx.need(importColor)
	literals := make([]string, 0, len(colors))
	seen := make(map[color.Color]struct{}, len(colors))
	for _, c := range colors {
		if _, ok := seen[c]; ok {
			return "", fmt.Errorf("render: duplicate color %q", c)
		}
		seen[c] = struct{}{}
		literal, err := colorValueToLiteral(c)
		if err != nil {
			return "", err
		}
		literals = append(literals, literal)
	}
	return strings.Join(literals, ", "), nil
}

func renderCardTypeArguments(ctx *renderCtx, cardTypes []types.Card) (string, error) {
	ctx.need(importTypes)
	literals := make([]string, 0, len(cardTypes))
	for _, t := range cardTypes {
		lit, err := cardTypeLiteral(t)
		if err != nil {
			return "", err
		}
		literals = append(literals, lit)
	}
	return strings.Join(literals, ", "), nil
}

func renderSubtypeArguments(ctx *renderCtx, subtypes []types.Sub) (string, error) {
	ctx.need(importTypes)
	literals := make([]string, 0, len(subtypes))
	for _, sub := range subtypes {
		lit := SubtypeToLiteral(string(sub), []string{"Creature", "Land"})
		if strings.HasPrefix(lit, "/*") {
			return "", fmt.Errorf("render: unsupported subtype %q", string(sub))
		}
		literals = append(literals, lit)
	}
	return strings.Join(literals, ", "), nil
}

func renderControllerRelation(cr game.ControllerRelation) (string, error) {
	switch cr {
	case game.ControllerAny:
		return "game.ControllerAny", nil
	case game.ControllerYou:
		return "game.ControllerYou", nil
	case game.ControllerOpponent:
		return "game.ControllerOpponent", nil
	case game.ControllerNotYou:
		return "game.ControllerNotYou", nil
	default:
		return "", fmt.Errorf("render: unsupported controller relation %d", cr)
	}
}

func renderTriState(ts game.TriState) (string, error) {
	switch ts {
	case game.TriAny:
		return "game.TriAny", nil
	case game.TriTrue:
		return "game.TriTrue", nil
	case game.TriFalse:
		return "game.TriFalse", nil
	default:
		return "", fmt.Errorf("render: unsupported tri-state %d", ts)
	}
}

func renderCombatStateFilter(cs game.CombatStateFilter) (string, error) {
	switch cs {
	case game.CombatStateAny:
		return "game.CombatStateAny", nil
	case game.CombatStateAttacking:
		return "game.CombatStateAttacking", nil
	case game.CombatStateBlocking:
		return "game.CombatStateBlocking", nil
	case game.CombatStateAttackingOrBlocking:
		return "game.CombatStateAttackingOrBlocking", nil
	default:
		return "", fmt.Errorf("render: unsupported combat state filter %d", cs)
	}
}

func renderKeyword(kw game.Keyword) (string, error) {
	switch kw {
	case game.KeywordNone:
		return "game.KeywordNone", nil
	case game.Devoid:
		return "game.Devoid", nil
	case game.Deathtouch:
		return "game.Deathtouch", nil
	case game.Defender:
		return "game.Defender", nil
	case game.DoubleStrike:
		return "game.DoubleStrike", nil
	case game.FirstStrike:
		return "game.FirstStrike", nil
	case game.Flash:
		return "game.Flash", nil
	case game.Flying:
		return "game.Flying", nil
	case game.Haste:
		return "game.Haste", nil
	case game.Hexproof:
		return "game.Hexproof", nil
	case game.Indestructible:
		return "game.Indestructible", nil
	case game.Lifelink:
		return "game.Lifelink", nil
	case game.Menace:
		return "game.Menace", nil
	case game.Protection:
		return "game.Protection", nil
	case game.Reach:
		return "game.Reach", nil
	case game.Shroud:
		return "game.Shroud", nil
	case game.Trample:
		return "game.Trample", nil
	case game.Vigilance:
		return "game.Vigilance", nil
	case game.Riot:
		return "game.Riot", nil
	case game.Ward:
		return "game.Ward", nil
	case game.SplitSecond:
		return "game.SplitSecond", nil
	case game.Equip:
		return "game.Equip", nil
	case game.Enchant:
		return "game.Enchant", nil
	case game.Cycling:
		return "game.Cycling", nil
	case game.Flashback:
		return "game.Flashback", nil
	case game.Kicker:
		return "game.Kicker", nil
	case game.Madness:
		return "game.Madness", nil
	case game.Morph:
		return "game.Morph", nil
	case game.Disguise:
		return "game.Disguise", nil
	case game.Convoke:
		return "game.Convoke", nil
	case game.Delve:
		return "game.Delve", nil
	case game.Suspend:
		return "game.Suspend", nil
	case game.Storm:
		return "game.Storm", nil
	case game.Cascade:
		return "game.Cascade", nil
	case game.Prowess:
		return "game.Prowess", nil
	case game.Mutate:
		return "game.Mutate", nil
	case game.Companion:
		return "game.Companion", nil
	case game.Ninjutsu:
		return "game.Ninjutsu", nil
	case game.Escape:
		return "game.Escape", nil
	case game.Foretell:
		return "game.Foretell", nil
	case game.Craft:
		return "game.Craft", nil
	case game.Discover:
		return "game.Discover", nil
	case game.Eternalize:
		return "game.Eternalize", nil
	case game.Affinity:
		return "game.Affinity", nil
	case game.Improvise:
		return "game.Improvise", nil
	case game.Emerge:
		return "game.Emerge", nil
	case game.Undying:
		return "game.Undying", nil
	case game.Persist:
		return "game.Persist", nil
	case game.Wither:
		return "game.Wither", nil
	case game.Infect:
		return "game.Infect", nil
	case game.Toxic:
		return "game.Toxic", nil
	case game.Annihilator:
		return "game.Annihilator", nil
	case game.Exalted:
		return "game.Exalted", nil
	case game.Evolve:
		return "game.Evolve", nil
	case game.ReadAhead:
		return "game.ReadAhead", nil
	case game.Horsemanship:
		return "game.Horsemanship", nil
	case game.Shadow:
		return "game.Shadow", nil
	case game.CumulativeUpkeep:
		return "game.CumulativeUpkeep", nil
	case game.Fear:
		return "game.Fear", nil
	case game.Skulk:
		return "game.Skulk", nil
	case game.Intimidate:
		return "game.Intimidate", nil
	default:
		return "", fmt.Errorf("render: unsupported keyword %d", kw)
	}
}

func renderCompareInt(ctx *renderCtx, cmp compare.Int) (string, error) {
	ctx.need(importCompare)
	op, err := renderCompareOp(cmp.Op)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("compare.Int{Op: %s, Value: %d}", op, cmp.Value), nil
}

func renderCompareOp(op compare.Op) (string, error) {
	switch op {
	case compare.Any:
		return "compare.Any", nil
	case compare.Equal:
		return "compare.Equal", nil
	case compare.LessOrEqual:
		return "compare.LessOrEqual", nil
	case compare.GreaterOrEqual:
		return "compare.GreaterOrEqual", nil
	case compare.LessThan:
		return "compare.LessThan", nil
	case compare.GreaterThan:
		return "compare.GreaterThan", nil
	default:
		return "", fmt.Errorf("render: unsupported compare op %d", op)
	}
}
