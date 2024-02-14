package dps

import (
	"testing"

	_ "github.com/wowsims/sod/sim/common"
	"github.com/wowsims/sod/sim/core"
	"github.com/wowsims/sod/sim/core/proto"
)

func init() {
	RegisterDpsWarlock()
}

func TestAffliction(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator(core.CharacterSuiteConfig{
		Class: proto.Class_ClassWarlock,
		Race:  proto.Race_RaceOrc,

		GearSet:     core.GetGearSet("../../../ui/warlock/gear_sets/p2", "shadow"),
		Talents:     AfflictionTalents,
		Consumes:    FullConsumes,
		SpecOptions: core.SpecOptionsCombo{Label: "Affliction Warlock", SpecOptions: DefaultAfflictionWarlock},
		OtherRotations: []core.RotationCombo{
			core.GetAplRotation("../../../ui/warlock/apls/p2", "affliction"),
		},

		ItemFilter: ItemFilter,
	}))
}

func TestDemonology(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator(core.CharacterSuiteConfig{
		Class: proto.Class_ClassWarlock,
		Race:  proto.Race_RaceOrc,

		GearSet:     core.GetGearSet("../../../ui/warlock/gear_sets/p2", "fire.succubus"),
		Talents:     DemonologyTalents,
		Consumes:    FullConsumes,
		SpecOptions: core.SpecOptionsCombo{Label: "Demonology Warlock", SpecOptions: DefaultDemonologyWarlock},
		OtherRotations: []core.RotationCombo{
			core.GetAplRotation("../../../ui/warlock/apls/p2", "demonology"),
		},

		ItemFilter: ItemFilter,
	}))
}

func TestDestruction(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator(core.CharacterSuiteConfig{
		Class: proto.Class_ClassWarlock,
		Race:  proto.Race_RaceOrc,

		GearSet:     core.GetGearSet("../../../ui/warlock/gear_sets/p2", "fire.imp"),
		Talents:     DestructionTalents,
		Consumes:    FullConsumes,
		SpecOptions: core.SpecOptionsCombo{Label: "Destruction Warlock", SpecOptions: DefaultDestroWarlock},
		OtherRotations: []core.RotationCombo{
			core.GetAplRotation("../../../ui/warlock/apls/p2", "fire.imp"),
		},
		ItemFilter: ItemFilter,
	}))
}

var ItemFilter = core.ItemFilter{
	WeaponTypes: []proto.WeaponType{
		proto.WeaponType_WeaponTypeSword,
		proto.WeaponType_WeaponTypeDagger,
	},
	HandTypes: []proto.HandType{
		proto.HandType_HandTypeOffHand,
	},
	ArmorType: proto.ArmorType_ArmorTypeCloth,
	RangedWeaponTypes: []proto.RangedWeaponType{
		proto.RangedWeaponType_RangedWeaponTypeWand,
	},
}

var AfflictionTalents = "3500253012201105--1"
var DemonologyTalents = "-2050033132501051"
var DestructionTalents = "-01-055020512000415"

var defaultDestroOptions = &proto.WarlockOptions{
	Armor:       proto.WarlockOptions_DemonArmor,
	Summon:      proto.WarlockOptions_Imp,
	WeaponImbue: proto.WarlockOptions_NoWeaponImbue,
}

var DefaultDestroWarlock = &proto.Player_Warlock{
	Warlock: &proto.Warlock{
		Options: defaultDestroOptions,
	},
}

// ---------------------------------------
var DefaultAfflictionWarlock = &proto.Player_Warlock{
	Warlock: &proto.Warlock{
		Options: defaultAfflictionOptions,
	},
}

var defaultAfflictionOptions = &proto.WarlockOptions{
	Armor:       proto.WarlockOptions_DemonArmor,
	Summon:      proto.WarlockOptions_Imp,
	WeaponImbue: proto.WarlockOptions_NoWeaponImbue,
}

// ---------------------------------------
var DefaultDemonologyWarlock = &proto.Player_Warlock{
	Warlock: &proto.Warlock{
		Options: defaultDemonologyOptions,
	},
}

var defaultDemonologyOptions = &proto.WarlockOptions{
	Armor:       proto.WarlockOptions_DemonArmor,
	Summon:      proto.WarlockOptions_Imp,
	WeaponImbue: proto.WarlockOptions_NoWeaponImbue,
}

// ---------------------------------------------------------

var FullConsumes = core.ConsumesCombo{
	Label: "Full Consumes",
	Consumes: &proto.Consumes{
		Flask:         proto.Flask_FlaskOfSupremePower,
		DefaultPotion: proto.Potions_ManaPotion,
		Food:          proto.Food_FoodBlessSunfruit,
	},
}
