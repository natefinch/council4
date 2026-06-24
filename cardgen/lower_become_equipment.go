package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// lowerBecomeEquipmentGrant recognizes the legendary permanent's optional
// become-Equipment transform that gains an Equip ability and a static buff and
// loses every other ability ("you may have ~ become a legendary Equipment
// artifact named Everflame, Heroes' Legacy. If you do, it gains equip {3} and
// \"Equipped creature gets +3/+3\" and loses all other abilities.", The
// Irencrag). The parser captures the become as an optional grant-keyword effect
// carrying the new name, the Equipment subtype, and the legendary supertype; the
// gain and lose riders as their own effects; the Equip ability as a keyword; the
// "if you do" join as a condition; and the granted static buff as a quoted
// ability on the syntax.
//
// It lowers the whole transform to one optional permanent ApplyContinuous on the
// source: a text layer renames it, a type layer makes it a legendary Equipment
// artifact, and an ability layer removes every prior ability while adding the
// granted Equip activated ability and the granted static buff. Declining the
// optional instruction performs nothing; accepting applies the whole transform,
// matching the printed "if you do". It fails closed for any deviation from this
// exact three-effect shape so no other card reaches it.
func lowerBecomeEquipmentGrant(ctx contentCtx, syntax *parser.Ability) (game.AbilityContent, bool) {
	content := ctx.content
	if ctx.optional ||
		len(content.Effects) != 3 ||
		len(content.Targets) != 0 ||
		len(content.Modes) != 0 ||
		len(content.Keywords) != 1 ||
		len(content.Conditions) != 1 ||
		syntax == nil ||
		len(syntax.Quoted) == 0 {
		return game.AbilityContent{}, false
	}
	become := content.Effects[0]
	gain := content.Effects[1]
	lose := content.Effects[2]
	subtypes := become.Selector.SubtypesAny()
	if become.Kind != compiler.EffectGrantKeyword ||
		become.Context != parser.EffectContextController ||
		!become.Optional ||
		become.Negated ||
		become.Selector.Kind != compiler.SelectorArtifact ||
		become.Selector.RequiredName == "" ||
		len(subtypes) != 1 ||
		subtypes[0] != types.Equipment ||
		len(become.References) != 1 ||
		become.References[0].Binding != compiler.ReferenceBindingSource {
		return game.AbilityContent{}, false
	}
	if gain.Kind != compiler.EffectGain || gain.Optional || gain.Negated {
		return game.AbilityContent{}, false
	}
	if lose.Kind != compiler.EffectLose ||
		lose.Optional ||
		lose.Negated ||
		!lose.Selector.All ||
		!lose.Selector.Other {
		return game.AbilityContent{}, false
	}
	if content.Conditions[0].Predicate != compiler.ConditionPredicatePriorInstructionAccepted {
		return game.AbilityContent{}, false
	}
	keyword := content.Keywords[0]
	if keyword.Kind != parser.KeywordEquip ||
		keyword.ParameterKind != parser.KeywordParameterManaCost ||
		len(keyword.ManaCost) == 0 {
		return game.AbilityContent{}, false
	}
	statics, ok := lowerBecomeGrantedStatics(syntax.Quoted)
	if !ok {
		return game.AbilityContent{}, false
	}
	object, ok := lowerObjectReference(become.References[0], referenceLoweringContext{AllowSource: true})
	if !ok {
		return game.AbilityContent{}, false
	}
	equip := lowerBecomeEquipAbility(keyword)
	abilities := make([]game.Ability, 0, 1+len(statics))
	abilities = append(abilities, &equip)
	for i := range statics {
		abilities = append(abilities, &statics[i])
	}
	lowered := game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(object),
				ContinuousEffects: []game.ContinuousEffect{
					{
						Layer:              game.LayerAbility,
						RemoveAllAbilities: true,
					},
					{
						Layer:   game.LayerText,
						SetName: become.Selector.RequiredName,
					},
					{
						Layer:         game.LayerType,
						SetTypes:      []types.Card{types.Artifact},
						SetSubtypes:   slices.Clone(subtypes),
						AddSupertypes: slices.Clone(become.Selector.Supertypes()),
					},
					{
						Layer:        game.LayerAbility,
						AddAbilities: abilities,
					},
				},
				Duration: game.DurationPermanent,
			},
		}},
	}.Ability()
	if !markSingleInstructionOptional(&lowered) {
		return game.AbilityContent{}, false
	}
	return lowered, true
}

// lowerBecomeEquipAbility builds the granted Equip activated ability from the
// single Equip keyword the become carries, mirroring lowerEquipAbility's mana
// cost and optional equip restriction handling.
func lowerBecomeEquipAbility(keyword compiler.CompiledKeyword) game.ActivatedAbility {
	if keyword.EquipRestriction != nil {
		return game.EquipRestrictedActivatedAbility(
			slices.Clone(keyword.ManaCost),
			slices.Clone(keyword.EquipRestriction.Supertypes),
			slices.Clone(keyword.EquipRestriction.Subtypes),
		)
	}
	return game.EquipActivatedAbility(slices.Clone(keyword.ManaCost))
}

// lowerBecomeGrantedStatics lowers every quoted ability the become grants to a
// runtime static ability. Each quoted clause must parse and lower to exactly one
// static ability; any other ability kind, parse failure, or lowering failure
// fails closed so the whole transform is rejected rather than partially applied.
func lowerBecomeGrantedStatics(quoted []parser.Delimited) ([]game.StaticAbility, bool) {
	statics := make([]game.StaticAbility, 0, len(quoted))
	for i := range quoted {
		granted, ok := parser.ParseGrantedStaticAbility(quoted[i])
		if !ok {
			return nil, false
		}
		static, ok := lowerBecomeGrantedStaticBody(&granted)
		if !ok {
			return nil, false
		}
		statics = append(statics, static)
	}
	if len(statics) == 0 {
		return nil, false
	}
	return statics, true
}

// lowerBecomeGrantedStaticBody compiles and lowers one quoted static ability the
// become grants to its runtime StaticAbility body, mirroring
// lowerStaticGrantedQuotedAbility's recursive compile and lower but returning the
// static body rather than a triggered, activated, or mana ability.
func lowerBecomeGrantedStaticBody(granted *parser.StaticGrantedAbilitySyntax) (game.StaticAbility, bool) {
	innerDocument, innerDiags := granted.Inner()
	if len(innerDiags) != 0 {
		return game.StaticAbility{}, false
	}
	innerComp, compilerDiags := compiler.Compile(innerDocument, compiler.Context{})
	if len(compilerDiags) != 0 ||
		len(innerComp.Abilities) != 1 ||
		len(innerComp.Syntax.Abilities) != 1 {
		return game.StaticAbility{}, false
	}
	lowered, diagnostic := lowerExecutableAbility("", false, nil, innerComp.Abilities[0], &innerComp.Syntax.Abilities[0])
	if diagnostic != nil || len(lowered.staticAbilities) != 1 {
		return game.StaticAbility{}, false
	}
	return lowered.staticAbilities[0].Body, true
}
