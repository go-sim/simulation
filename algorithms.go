package godesim

import "github.com/soypat/godesim/state"

// RK4Solver Integrates simulation state for next timesteps
// using 4th order Runge-Kutta multivariable algorithm
func RK4Solver(sim *Simulation) []state.State {
	const overSix float64 = 1. / 6.
	states := make([]state.State, sim.Algorithm.Steps+1)
	h := sim.Dt() / float64(sim.Algorithm.Steps)
	states[0] = sim.State.Clone()
	for i := 0; i < len(states)-1; i++ {
		// create auxiliary states for calculation
		t := states[i].Time()
		b, c, d := states[i].CloneBlank(t+.5*h), states[i].CloneBlank(t+.5*h), states[i].CloneBlank(t+h)

		a := StateDiff(sim.change, states[i])

		state.AddScaledTo(b, states[i], 0.5*h, a)
		b = StateDiff(sim.change, b)

		state.AddScaledTo(c, states[i], 0.5*h, b)
		c = StateDiff(sim.change, c)

		state.AddScaledTo(d, states[i], h, c)
		d = StateDiff(sim.change, d)

		state.Add(a, d)
		state.Add(b, c)
		state.AddScaled(a, 2, b)
		states[i+1] = states[i].Clone()
		state.AddScaled(states[i+1], h*overSix, a)
		states[i+1].SetTime(h + states[i].Time())
	}
	return states
}

// RKF45Solver an attempt at a Runge-Kutta-Fehlberg method
// solver. Seems to not work for now
func RKF45Solver(sim *Simulation) []state.State {
	// Butcher Tableau for Fehlbergs  4(5) method (Table III https://en.wikipedia.org/wiki/Runge%E2%80%93Kutta%E2%80%93Fehlberg_method)
	const c20, c21 = 1. / 4., 1. / 4.
	const c30, c31, c32 = 3. / 8., 3. / 32., 9. / 32.
	const c40, c41, c42, c43 = 12. / 13., 1932. / 2197., -7200. / 2197., 7296. / 2197
	const c50, c51, c52, c53, c54 = 1., 439. / 216., -8., 3680. / 513., -845. / 4104.
	const c60, c61, c62, c63, c64, c65 = .5, -8. / 27., 2., -3544. / 2565., 1859. / 4104., -11. / 40.
	// Fourth order
	const a1, a3, a4, a5 = 25. / 216., 1408. / 2565., 2197. / 4104., -1. / 5.
	// Fifth order
	const b1, b3, b4, b5, b6 = 16. / 135., 6656. / 12825., 28561. / 56430., -9. / 50., 2. / 55.

	states := make([]state.State, sim.Algorithm.Steps+1)
	h := sim.Dt() / float64(sim.Algorithm.Steps)
	states[0] = sim.State.Clone()
	for i := 0; i < len(states)-1; i++ {
		// create auxiliary states for calculation
		t := states[i].Time()
		k2, k3, k4, k5, k6, s4, s5, err45 := states[i].CloneBlank(t+c20*h), states[i].CloneBlank(t+c30*h), states[i].CloneBlank(t+c40*h),
			states[i].CloneBlank(t+h), states[i].CloneBlank(t+c60*h), states[i].CloneBlank(t+h), states[i].CloneBlank(t+h), states[i].CloneBlank(t+h)

		k1 := StateDiff(sim.change, states[i])

		// k2 calc
		state.AddScaledTo(k2, states[i], c21, k1)
		k2 = StateDiff(sim.change, k2)

		// k3 calc
		state.AddScaledTo(k3, states[i], c31, k1)
		state.AddScaled(k3, c32, k2)
		k3 = StateDiff(sim.change, k3)
		// k4 calc
		state.AddScaledTo(k4, states[i], c41, k1)
		state.AddScaled(k4, c42, k2)
		state.AddScaled(k4, c43, k3)
		k4 = StateDiff(sim.change, k4)
		// k5 calc
		state.AddScaledTo(k5, states[i], c51, k1)
		state.AddScaled(k5, c52, k2)
		state.AddScaled(k5, c53, k3)
		state.AddScaled(k5, c54, k4)
		k5 = StateDiff(sim.change, k5)
		// k6 calc
		state.AddScaledTo(k6, states[i], c61, k1)
		state.AddScaled(k6, c62, k2)
		state.AddScaled(k6, c63, k3)
		state.AddScaled(k6, c64, k4)
		state.AddScaled(k6, c65, k5)
		k6 = StateDiff(sim.change, k6)

		// fourth order approximation calc
		state.AddScaledTo(s4, states[i], h*a1, k1)
		state.AddScaled(s4, h*a3, k3)
		state.AddScaled(s4, h*a4, k4)
		state.AddScaled(s4, h*a5, k5)

		// fifth order approximation calc
		state.AddScaledTo(s5, states[i], h*b1, k1)
		state.AddScaled(s5, h*b3, k3)
		state.AddScaled(s5, h*b4, k4)
		state.AddScaled(s5, h*b5, k5)
		state.AddScaled(s5, h*b6, k6)

		states[i+1] = s5.Clone()
		// calculate error. Should be absolute value
		state.Abs(state.SubTo(err45, s4, s5))

	}
	return states
}

//  RKF45TableauSolver same as RKF45Solver but using arrays
// as tableaus. Should be slower than RKF45Solver in all respects (except for unidimensional problems).
// This function was written to test RKF45, which still needs to be implemented
func RKF45TableauSolver(sim *Simulation) []state.State {
	// Butcher Tableau for Fehlbergs  4(5) method (from Table III https://en.wikipedia.org/wiki/Runge%E2%80%93Kutta%E2%80%93Fehlberg_method)
	A := [6]float64{0, 1. / 4., 3. / 8., 12. / 13., 1., 1. / 2.}
	B := [6][5]float64{
		{0, 0, 0, 0, 0},
		{1. / 4., 0, 0, 0, 0},
		{3. / 32., 9. / 32., 0, 0, 0},
		{1932. / 2197., -7200. / 2197., 7296. / 2197., 0, 0},
		{439. / 216., -8., 3680. / 513., -845. / 4104., 0},
		{-8. / 27., 2., -3544. / 2565., 1859. / 4104., -11. / 40.},
	}
	C := [6]float64{25. / 216., 0, 1408. / 2565., 2197. / 4104., -.2, 0}
	CH := [6]float64{16. / 135., 0, 6656. / 12825., 28561. / 56430., -9. / 50., 2. / 55.}

	states := make([]state.State, sim.Algorithm.Steps+1)
	h := sim.Dt() / float64(sim.Algorithm.Steps)
	states[0] = sim.State.Clone()
	for i := 0; i < len(states)-1; i++ {
		// create auxiliary states for calculation
		t := states[i].Time()
		k2, k3, k4, k5, k6, s4, s5, err45 := states[i].CloneBlank(t+A[1]*h), states[i].CloneBlank(t+A[2]*h), states[i].CloneBlank(t+A[3]*h),
			states[i].CloneBlank(t+A[4]*h), states[i].CloneBlank(t+A[5]*h), states[i].CloneBlank(t+h), states[i].CloneBlank(t+h), states[i].CloneBlank(t+h)

		k1 := StateDiff(sim.change, states[i])
		state.Scale(h, k1)

		// k2 calc
		state.AddScaledTo(k2, states[i], B[1][0], k1)
		k2 = StateDiff(sim.change, k2)
		state.Scale(h, k2)

		// k3 calc
		state.AddScaledTo(k3, states[i], B[2][0], k1)
		state.AddScaled(k3, B[2][1], k2)
		k3 = StateDiff(sim.change, k3)
		state.Scale(h, k3)

		// k4 calc
		state.AddScaledTo(k4, states[i], B[3][0], k1)
		state.AddScaled(k4, B[3][1], k2)
		state.AddScaled(k4, B[3][2], k3)
		k4 = StateDiff(sim.change, k4)
		state.Scale(h, k4)

		// k5 calc
		state.AddScaledTo(k5, states[i], B[4][0], k1)
		state.AddScaled(k5, B[4][1], k2)
		state.AddScaled(k5, B[4][2], k3)
		state.AddScaled(k5, B[4][3], k4)
		k5 = StateDiff(sim.change, k5)
		state.Scale(h, k5)
		// k6 calc
		state.AddScaledTo(k6, states[i], B[5][0], k1)
		state.AddScaled(k6, B[5][1], k2)
		state.AddScaled(k6, B[5][2], k3)
		state.AddScaled(k6, B[5][3], k4)
		state.AddScaled(k6, B[5][4], k5)
		k6 = StateDiff(sim.change, k6)
		state.Scale(h, k6)

		// fifth order approximation calc
		state.AddScaledTo(s5, states[i], CH[0], k1)
		state.AddScaled(s5, CH[1], k2)
		state.AddScaled(s5, CH[2], k3)
		state.AddScaled(s5, CH[3], k4)
		state.AddScaled(s5, CH[4], k5)
		state.AddScaled(s5, CH[5], k6)

		// fourth order approximation calc
		state.AddScaledTo(s4, states[i], C[0], k1)
		state.AddScaled(s4, C[1], k3)
		state.AddScaled(s4, C[2], k4)
		state.AddScaled(s4, C[3], k5)
		state.AddScaled(s4, C[4], k6)

		states[i+1] = s5.Clone()
		// calculate error. Should be absolute value
		state.Abs(state.SubTo(err45, s4, s5))
	}
	return states
}