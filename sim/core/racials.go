package core

import (
	"time"

	"github.com/wowsims/sod/sim/core/proto"
	"github.com/wowsims/sod/sim/core/stats"
)

func applyRaceEffects(agent Agent) {
	character := agent.GetCharacter()

	switch character.Race {
	case proto.Race_RaceDwarf:
		character.PseudoStats.ReducedFrostHitTakenChance += 0.02

		// Gun specialization (+1% ranged crit when using a gun).
		if character.Ranged().RangedWeaponType == proto.RangedWeaponType_RangedWeaponTypeGun {
			character.AddBonusRangedCritRating(1 * CritRatingPerCritChance)
		}

		actionID := ActionID{SpellID: 20594}

		statDep := character.NewDynamicMultiplyStat(stats.Armor, 1.1)
		stoneFormAura := character.NewTemporaryStatsAuraWrapped("Stoneform", actionID, stats.Stats{}, time.Second*8, func(aura *Aura) {
			aura.ApplyOnGain(func(aura *Aura, sim *Simulation) {
				aura.Unit.EnableDynamicStatDep(sim, statDep)
			})
			aura.ApplyOnExpire(func(aura *Aura, sim *Simulation) {
				aura.Unit.DisableDynamicStatDep(sim, statDep)
			})
		})

		spell := character.RegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				stoneFormAura.Activate(sim)
			},
		})

		character.AddMajorCooldown(MajorCooldown{
			Spell: spell,
			Type:  CooldownTypeDPS,
		})
	case proto.Race_RaceGnome:
		character.PseudoStats.ReducedArcaneHitTakenChance += 0.02
		character.MultiplyStat(stats.Intellect, 1.05)
	case proto.Race_RaceHuman:
		character.MultiplyStat(stats.Spirit, 1.03)
		character.ApplyWeaponSpecialization(5, proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeSword)
	case proto.Race_RaceNightElf:
		character.PseudoStats.ReducedNatureHitTakenChance += 0.02
		character.PseudoStats.ReducedPhysicalHitTakenChance += 0.02
		// TODO: Shadowmeld?
	case proto.Race_RaceOrc:
		// Command (Pet damage +5%)
		for _, pet := range character.Pets {
			pet.PseudoStats.DamageDealtMultiplier *= 1.05
		}

		// Blood Fury
		actionID := ActionID{SpellID: 20572}
		apBonus := float64(character.Level)*4 + 2
		spBonus := float64(character.Level)*2 + 3
		bloodFuryAura := character.NewTemporaryStatsAura("Blood Fury", actionID, stats.Stats{stats.AttackPower: apBonus, stats.RangedAttackPower: apBonus, stats.SpellPower: spBonus}, time.Second*15)

		spell := character.RegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				bloodFuryAura.Activate(sim)
			},
		})

		character.AddMajorCooldown(MajorCooldown{
			Spell: spell,
			Type:  CooldownTypeDPS,
		})

		// Axe specialization
		character.ApplyWeaponSpecialization(5, proto.WeaponType_WeaponTypeAxe, proto.WeaponType_WeaponTypeFist)
	case proto.Race_RaceTauren:
		character.PseudoStats.ReducedNatureHitTakenChance += 0.02
		character.AddStat(stats.Health, character.GetBaseStats()[stats.Health]*0.05)
	case proto.Race_RaceTroll:
		// Bow specialization (+1% ranged crit when using a bow).
		if character.Ranged().RangedWeaponType == proto.RangedWeaponType_RangedWeaponTypeBow {
			character.AddBonusRangedCritRating(1 * CritRatingPerCritChance)
		}

		// Beast Slaying (+5% damage to beasts)
		if character.CurrentTarget.MobType == proto.MobType_MobTypeBeast {
			character.PseudoStats.DamageDealtMultiplier *= 1.05
		}

		// Berserking
		actionID := ActionID{SpellID: 26297}

		berserkingAura := character.RegisterAura(Aura{
			Label:    "Berserking (Troll)",
			ActionID: actionID,
			Duration: time.Second * 10,
			OnGain: func(aura *Aura, sim *Simulation) {
				character.MultiplyCastSpeed(1.2)
				character.MultiplyAttackSpeed(sim, 1.2)
			},
			OnExpire: func(aura *Aura, sim *Simulation) {
				character.MultiplyCastSpeed(1 / 1.2)
				character.MultiplyAttackSpeed(sim, 1/1.2)
			},
		})

		berserkingSpell := character.RegisterSpell(SpellConfig{
			ActionID: actionID,

			Cast: CastConfig{
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 3,
				},
			},

			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				berserkingAura.Activate(sim)
			},
		})

		character.AddMajorCooldown(MajorCooldown{
			Spell: berserkingSpell,
			Type:  CooldownTypeDPS,
		})
	case proto.Race_RaceUndead:
		character.PseudoStats.ReducedShadowHitTakenChance += 0.02
	}
}
