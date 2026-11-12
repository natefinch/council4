package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const forMirrodinText = "For Mirrodin! (When this Equipment enters, create a 2/2 red Rebel creature token, then attach this to it.)"

var forMirrodinRebel = &game.CardDef{CardFace: game.CardFace{
	Name:      "Rebel Token",
	Colors:    []color.Color{color.Red},
	Types:     []types.Card{types.Creature},
	Subtypes:  []types.Sub{types.Rebel},
	Power:     opt.Val(game.PT{Value: 2}),
	Toughness: opt.Val(game.PT{Value: 2}),
}}

type forMirrodinEquipmentSpec struct {
	oracleID       string
	name           string
	oracleText     string
	equippedText   string
	manaCost       cost.Mana
	cardColor      color.Color
	equipCost      cost.Mana
	powerBonus     int
	toughnessBonus int
	keywords       []game.Keyword
	creatureCount  bool
}

var mirranBardicheSpec = forMirrodinEquipmentSpec{
	oracleID: "1fd04d9e-cf89-49ba-a932-5d278f65bf77",
	name:     "Mirran Bardiche",
	oracleText: forMirrodinText + "\n" +
		"Equipped creature gets +2/+1 and has vigilance.\n" +
		"Equip {3}{W} ({3}{W}: Attach to target creature you control. Equip only as a sorcery.)",
	equippedText:   "Equipped creature gets +2/+1 and has vigilance.",
	manaCost:       cost.Mana{cost.O(4), cost.W},
	cardColor:      color.White,
	equipCost:      cost.Mana{cost.O(3), cost.W},
	powerBonus:     2,
	toughnessBonus: 1,
	keywords:       []game.Keyword{game.Vigilance},
}

var sylvokBattleChairSpec = forMirrodinEquipmentSpec{
	oracleID: "2f6c21a1-d618-497c-bfde-c0f2e1014d99",
	name:     "Sylvok Battle-Chair",
	oracleText: forMirrodinText + "\n" +
		"Equipped creature gets +4/+4 and has trample.\n" +
		"Equip {5}{G}{G}",
	equippedText:   "Equipped creature gets +4/+4 and has trample.",
	manaCost:       cost.Mana{cost.O(4), cost.G, cost.G},
	cardColor:      color.Green,
	equipCost:      cost.Mana{cost.O(5), cost.G, cost.G},
	powerBonus:     4,
	toughnessBonus: 4,
	keywords:       []game.Keyword{game.Trample},
}

var goldwardensHelmSpec = forMirrodinEquipmentSpec{
	oracleID: "4c35f41a-55e4-49ca-b211-f57d7b1b90f6",
	name:     "Goldwarden's Helm",
	oracleText: forMirrodinText + "\n" +
		"Equipped creature gets +0/+1.\n" +
		"Equip {1}{W} ({1}{W}: Attach to target creature you control. Equip only as a sorcery.)",
	equippedText:   "Equipped creature gets +0/+1.",
	manaCost:       cost.Mana{cost.O(2), cost.W},
	cardColor:      color.White,
	equipCost:      cost.Mana{cost.O(1), cost.W},
	toughnessBonus: 1,
}

var dragonwingGliderSpec = forMirrodinEquipmentSpec{
	oracleID: "4cd4c6da-a84e-4cba-893b-799a7cb5c411",
	name:     "Dragonwing Glider",
	oracleText: forMirrodinText + "\n" +
		"Equipped creature gets +2/+2 and has flying and haste.\n" +
		"Equip {3}{R}{R}",
	equippedText:   "Equipped creature gets +2/+2 and has flying and haste.",
	manaCost:       cost.Mana{cost.O(3), cost.R, cost.R},
	cardColor:      color.Red,
	equipCost:      cost.Mana{cost.O(3), cost.R, cost.R},
	powerBonus:     2,
	toughnessBonus: 2,
	keywords:       []game.Keyword{game.Flying, game.Haste},
}

var vulshokSplitterSpec = forMirrodinEquipmentSpec{
	oracleID: "6253f9d4-83ff-4a0b-b465-e9f9a565f664",
	name:     "Vulshok Splitter",
	oracleText: forMirrodinText + "\n" +
		"Equipped creature gets +2/+0.\n" +
		"Equip {2}{R} ({2}{R}: Attach to target creature you control. Equip only as a sorcery.)",
	equippedText: "Equipped creature gets +2/+0.",
	manaCost:     cost.Mana{cost.O(3), cost.R},
	cardColor:    color.Red,
	equipCost:    cost.Mana{cost.O(2), cost.R},
	powerBonus:   2,
}

var barbedBatterfistSpec = forMirrodinEquipmentSpec{
	oracleID: "93c2b91d-7d4e-483c-a11f-4a525f881503",
	name:     "Barbed Batterfist",
	oracleText: forMirrodinText + "\n" +
		"Equipped creature gets +1/-1.\n" +
		"Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)",
	equippedText:   "Equipped creature gets +1/-1.",
	manaCost:       cost.Mana{cost.O(1), cost.R},
	cardColor:      color.Red,
	equipCost:      cost.Mana{cost.O(1)},
	powerBonus:     1,
	toughnessBonus: -1,
}

var kembasBannerSpec = forMirrodinEquipmentSpec{
	oracleID: "bd31c248-8b3f-4fa5-a639-19b0e37b15c3",
	name:     "Kemba's Banner",
	oracleText: forMirrodinText + "\n" +
		"Equipped creature gets +1/+1 for each creature you control.\n" +
		"Equip {2}{W}",
	equippedText:  "Equipped creature gets +1/+1 for each creature you control.",
	manaCost:      cost.Mana{cost.O(3), cost.W},
	cardColor:     color.White,
	equipCost:     cost.Mana{cost.O(2), cost.W},
	creatureCount: true,
}

var (
	MirranBardiche = cardset.Entry{
		OracleID: mirranBardicheSpec.oracleID, Name: mirranBardicheSpec.name,
		New: newMirranBardiche,
		Verification: "mtg/rules/for_mirrodin_equipment_test.go: " +
			"TestMirranBardicheScenario",
	}
	SylvokBattleChair = cardset.Entry{
		OracleID: sylvokBattleChairSpec.oracleID, Name: sylvokBattleChairSpec.name,
		New: newSylvokBattleChair,
		Verification: "mtg/rules/for_mirrodin_equipment_test.go: " +
			"TestSylvokBattleChairScenario",
	}
	GoldwardensHelm = cardset.Entry{
		OracleID: goldwardensHelmSpec.oracleID, Name: goldwardensHelmSpec.name,
		New: newGoldwardensHelm,
		Verification: "mtg/rules/for_mirrodin_equipment_test.go: " +
			"TestGoldwardensHelmScenario",
	}
	DragonwingGlider = cardset.Entry{
		OracleID: dragonwingGliderSpec.oracleID, Name: dragonwingGliderSpec.name,
		New: newDragonwingGlider,
		Verification: "mtg/rules/for_mirrodin_equipment_test.go: " +
			"TestDragonwingGliderScenario",
	}
	VulshokSplitter = cardset.Entry{
		OracleID: vulshokSplitterSpec.oracleID, Name: vulshokSplitterSpec.name,
		New: newVulshokSplitter,
		Verification: "mtg/rules/for_mirrodin_equipment_test.go: " +
			"TestVulshokSplitterScenario",
	}
	BarbedBatterfist = cardset.Entry{
		OracleID: barbedBatterfistSpec.oracleID, Name: barbedBatterfistSpec.name,
		New: newBarbedBatterfist,
		Verification: "mtg/rules/for_mirrodin_equipment_test.go: " +
			"TestBarbedBatterfistScenario",
	}
	KembasBanner = cardset.Entry{
		OracleID: kembasBannerSpec.oracleID, Name: kembasBannerSpec.name,
		New: newKembasBanner,
		Verification: "mtg/rules/for_mirrodin_equipment_test.go: " +
			"TestKembasBannerScenario",
	}
)

func newForMirrodinEquipment(spec forMirrodinEquipmentSpec) *game.CardDef {
	attached := game.AttachedObjectGroup(game.SourcePermanentReference())
	modifier := game.ContinuousEffect{
		Layer:          game.LayerPowerToughnessModify,
		Group:          attached,
		PowerDelta:     spec.powerBonus,
		ToughnessDelta: spec.toughnessBonus,
	}
	if spec.creatureCount {
		count := game.DynamicAmount{
			Kind:       game.DynamicAmountCountSelector,
			Multiplier: 1,
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerYou,
			}),
		}
		modifier.PowerDeltaDynamic = opt.Val(count)
		modifier.ToughnessDeltaDynamic = opt.Val(count)
	}
	effects := []game.ContinuousEffect{modifier}
	if len(spec.keywords) != 0 {
		effects = append(effects, game.ContinuousEffect{
			Layer:       game.LayerAbility,
			Group:       attached,
			AddKeywords: append([]game.Keyword(nil), spec.keywords...),
		})
	}
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(spec.cardColor),
		CardFace: game.CardFace{
			Name:     spec.name,
			ManaCost: opt.Val(append(cost.Mana(nil), spec.manaCost...)),
			Colors:   []color.Color{spec.cardColor},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			TriggeredAbilities: []game.TriggeredAbility{
				forMirrodinAbility(spec.oracleID),
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(append(cost.Mana(nil), spec.equipCost...)),
			},
			StaticAbilities: []game.StaticAbility{{
				Text:              spec.equippedText,
				ContinuousEffects: effects,
			}},
			OracleText: spec.oracleText,
		},
	}
}

func forMirrodinAbility(oracleID string) game.TriggeredAbility {
	linkedRebel := game.LinkedKey(oracleID + "/for-mirrodin/rebel")
	return game.TriggeredAbility{
		Text: forMirrodinText,
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhen,
			Pattern: game.TriggerPattern{
				Event:  game.EventPermanentEnteredBattlefield,
				Source: game.TriggerSourceSelf,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{
			{
				Primitive: game.CreateToken{
					Amount:        game.Fixed(1),
					Source:        game.TokenDef(forMirrodinRebel),
					PublishLinked: linkedRebel,
				},
			},
			{
				Primitive: game.Attach{
					Attachment: game.SourcePermanentReference(),
					Target:     game.LinkedObjectReference(string(linkedRebel)),
				},
			},
		}}.Ability(),
	}
}

func newMirranBardiche() *game.CardDef {
	return newForMirrodinEquipment(mirranBardicheSpec)
}

func newSylvokBattleChair() *game.CardDef {
	return newForMirrodinEquipment(sylvokBattleChairSpec)
}

func newGoldwardensHelm() *game.CardDef {
	return newForMirrodinEquipment(goldwardensHelmSpec)
}

func newDragonwingGlider() *game.CardDef {
	return newForMirrodinEquipment(dragonwingGliderSpec)
}

func newVulshokSplitter() *game.CardDef {
	return newForMirrodinEquipment(vulshokSplitterSpec)
}

func newBarbedBatterfist() *game.CardDef {
	return newForMirrodinEquipment(barbedBatterfistSpec)
}

func newKembasBanner() *game.CardDef {
	return newForMirrodinEquipment(kembasBannerSpec)
}
