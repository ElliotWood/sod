import { APLRotation } from '../core/proto/apl.js';
import {
	Class,
	Faction,
	ItemSlot,
	PartyBuffs,
	Race,
	Spec,
	Stat,
} from '../core/proto/common.js';
import { WarlockRune } from '../core/proto/warlock.js';
import { Stats } from '../core/proto_utils/stats.js';
import { Player } from '../core/player.js';
import { getSpecIcon } from '../core/proto_utils/utils.js';
import { IndividualSimUI, registerSpecConfig } from '../core/individual_sim_ui.js';

import * as BuffDebuffInputs from '../core/components/inputs/buffs_debuffs';
import * as ConsumablesInputs from '../core/components/inputs/consumables.js';
import * as OtherInputs from '../core/components/other_inputs.js';
import * as WarlockInputs from './inputs.js';
import * as Presets from './presets.js';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecTankWarlock, {
	cssClass: 'tank-warlock-sim-ui',
	cssScheme: 'warlock',
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [
		"Everything P2 should now be functional."
	],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatIntellect,
		Stat.StatSpirit,
		Stat.StatSpellPower,
		Stat.StatSpellDamage,
		Stat.StatSpellHit,
		Stat.StatSpellCrit,
		Stat.StatSpellHaste,
		Stat.StatFirePower,
		Stat.StatShadowPower,
		Stat.StatSpellPenetration,
		Stat.StatMP5,

		// Tank stats
		Stat.StatStrength,
		Stat.StatStamina,
		Stat.StatAttackPower,
		Stat.StatArmorPenetration,
		Stat.StatAgility,
		Stat.StatArmor,
		Stat.StatBonusArmor,
		Stat.StatMeleeCrit,
		Stat.StatMeleeHit,
		Stat.StatMeleeHaste,
	],
	// Reference stat against which to calculate EP. DPS classes use either spell power or attack power.
	epReferenceStat: Stat.StatSpellPower,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: [
		Stat.StatHealth,
		Stat.StatMana,
		Stat.StatIntellect,
		Stat.StatSpirit,
		Stat.StatSpellPower,
		Stat.StatSpellDamage,
		Stat.StatFirePower,
		Stat.StatShadowPower,
		Stat.StatSpellHit,
		Stat.StatSpellCrit,
		Stat.StatSpellHaste,
		Stat.StatSpellPenetration,
		Stat.StatMP5,
		// Tank stats
		Stat.StatStamina,
		Stat.StatAgility,
		Stat.StatArmor,
		Stat.StatBonusArmor,
	],

	defaults: {
		// Default equipped gear.
		gear: Presets.GearAfflictionTankDefault.gear,

		// Default EP weights for sorting gear in the gear picker.
		epWeights: Stats.fromMap({
			[Stat.StatIntellect]: 0.18,
			[Stat.StatSpirit]: 0.54,
			[Stat.StatSpellPower]: 1,
			[Stat.StatSpellDamage]: 1,
			[Stat.StatFirePower]: 1,
			[Stat.StatShadowPower]: 1,
			[Stat.StatSpellHit]: 0.93,
			[Stat.StatSpellCrit]: 0.53,
			[Stat.StatSpellHaste]: 0.81,
			[Stat.StatStamina]: 0.01,
		}),
		// Default consumes settings.
		consumes: Presets.DefaultConsumes,
		// Default talents.
		talents: Presets.AfflictionTankTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,

		// Default buffs and debuffs settings.
		raidBuffs: Presets.DefaultRaidBuffs,

		partyBuffs: PartyBuffs.create({}),

		individualBuffs: Presets.DefaultIndividualBuffs,

		debuffs: Presets.DefaultDebuffs,

		other: Presets.OtherDefaults,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [
		WarlockInputs.PetInput,
		WarlockInputs.ArmorInput,
		WarlockInputs.WeaponImbueInput,
	],
	rotationInputs: WarlockInputs.WarlockRotationConfig,

	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [
		BuffDebuffInputs.ResistanceBuff,
	],
	excludeBuffDebuffInputs: [
		BuffDebuffInputs.BleedDebuff,
		BuffDebuffInputs.SpellWintersChillDebuff,
		...ConsumablesInputs.FROST_POWER_CONFIG,
	],
	petConsumeInputs: [
	],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [
			OtherInputs.TankAssignment,
			OtherInputs.IncomingHps,
			OtherInputs.HealingCadence,
			OtherInputs.HealingCadenceVariation,
			OtherInputs.BurstWindow,
			OtherInputs.HpPercentForDefensives,
			OtherInputs.InspirationUptime,
			OtherInputs.ChannelClipDelay,
		],
	},
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotRanged],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		// Preset talents that the user can quickly select.
		talents: [
			Presets.AfflictionTankTalents,
			Presets.DestroTalents,
		],
		// Preset rotations that the user can quickly select.
		rotations: [
			Presets.RotationAfflictionTankDefault,
			Presets.RotationDestructionTankDefault,
		],

		// Preset gear configurations that the user can quickly select.
		gear: [
			Presets.GearAfflictionTankDefault,
			Presets.GearDestructionTankDefault,
		],
	},

	autoRotation: (player: Player<Spec.SpecTankWarlock>): APLRotation => {
		const hasMasterChanneler = player.getEquippedItem(ItemSlot.ItemSlotChest)?.rune?.id == WarlockRune.RuneChestMasterChanneler
		const hasLakeOfFire = player.getEquippedItem(ItemSlot.ItemSlotChest)?.rune?.id == WarlockRune.RuneChestLakeOfFire
		if (hasMasterChanneler) {
			return Presets.RotationAfflictionTankDefault.rotation.rotation!;
		} else if (hasLakeOfFire) {
			return Presets.RotationDestructionTankDefault.rotation.rotation!;
		} else {
			return Presets.RotationDestructionTankDefault.rotation.rotation!;
		}
	},

	raidSimPresets: [
		{
			spec: Spec.SpecTankWarlock,
			tooltip: 'Affliction Tank',
			defaultName: 'Affliction',
			iconUrl: getSpecIcon(Class.ClassWarlock, 0),

			talents: Presets.DefaultTalents.data,
			specOptions: Presets.DefaultOptions,
			consumes: Presets.DefaultConsumes,
			defaultFactionRaces: {
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Alliance]: Race.RaceHuman,
				[Faction.Horde]: Race.RaceOrc,
			},
			defaultGear: {
				[Faction.Unknown]: {},
				[Faction.Alliance]: {
					1: Presets.GearAfflictionTankDefault.gear,
				},
				[Faction.Horde]: {
					1: Presets.GearAfflictionTankDefault.gear,
				},
			},
			otherDefaults: Presets.OtherDefaults,
		},
		{
			spec: Spec.SpecTankWarlock,
			tooltip: 'Destruction Tank',
			defaultName: 'Destruction',
			iconUrl: getSpecIcon(Class.ClassWarlock, 2),

			talents: Presets.DefaultTalents.data,
			specOptions: Presets.DefaultOptions,
			consumes: Presets.DefaultConsumes,
			defaultFactionRaces: {
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Alliance]: Race.RaceHuman,
				[Faction.Horde]: Race.RaceOrc,
			},
			defaultGear: {
				[Faction.Unknown]: {},
				[Faction.Alliance]: {
					1: Presets.GearDestructionTankDefault.gear,
				},
				[Faction.Horde]: {
					1: Presets.GearDestructionTankDefault.gear,
				},
			},
			otherDefaults: Presets.OtherDefaults,
		},
	],
});

export class TankWarlockSimUI extends IndividualSimUI<Spec.SpecTankWarlock> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecTankWarlock>) {
		super(parentElem, player, SPEC_CONFIG);
	}
}
