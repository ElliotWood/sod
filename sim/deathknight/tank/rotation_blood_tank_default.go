package tank

import (
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/deathknight"
)

func (dk *TankDeathknight) setupBloodTankERWOpener() {
	dk.RotationSequence.
		NewAction(dk.RotationActionCallback_IT).
		NewAction(dk.RotationActionCallback_PS).
		NewAction(dk.RotationActionCallback_DS).
		NewAction(dk.RotationActionCallback_BT).
		NewAction(dk.RotationActionCallback_IT).
		NewAction(dk.RotationActionCallback_BS).
		NewAction(dk.RotationActionCallback_ERW).
		NewAction(dk.RotationActionCallback_Pesti).
		NewAction(dk.RotationActionCallback_IT).
		NewAction(dk.RotationActionCallback_IT).
		NewAction(dk.RotationActionCallback_IT).
		NewAction(dk.RotationActionCallback_DS).
		NewAction(dk.RotationActionCallback_TankBlood_PrioRotation)
}

func (dk *TankDeathknight) RotationActionCallback_TankBlood_PrioRotation(sim *core.Simulation, target *core.Unit, s *deathknight.Sequence) time.Duration {
	waitUntil := time.Duration(-1)

	attackGcd := 1500 * time.Millisecond
	spellGcd := dk.SpellGCD()
	ff := dk.FrostFeverDisease[target.Index].IsActive()
	bp := dk.BloodPlagueDisease[target.Index].IsActive()
	fbAt := core.MinDuration(dk.FrostFeverDisease[target.Index].ExpiresAt(), dk.BloodPlagueDisease[target.Index].ExpiresAt())

	if dk.VampiricBlood.CanCast(sim) && dk.CurrentHealthPercent() < 0.2 {
		dk.VampiricBlood.Cast(sim, target)
	} else if dk.RuneTap.CanCast(sim) && dk.CurrentHealthPercent() < 0.2 {
		dk.RuneTap.Cast(sim, target)
	} else {
		if !ff && dk.IcyTouch.CanCast(sim) {
			dk.IcyTouch.Cast(sim, target)
		} else if !bp && dk.PlagueStrike.CanCast(sim) {
			dk.PlagueStrike.Cast(sim, target)
		} else if dk.DeathStrike.CanCast(sim) && sim.CurrentTime+attackGcd < fbAt {
			dk.DeathStrike.Cast(sim, target)
		} else {
			if sim.CurrentTime < fbAt-2*spellGcd {
				waitUntil = fbAt - 2*spellGcd
			} else {
				dk.Pestilence.Cast(sim, target)
			}
		}
	}

	return waitUntil
}
