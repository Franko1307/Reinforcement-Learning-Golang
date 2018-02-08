package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

const (
	max_car_location_1 int = 5
	max_car_location_2     = 5

	λ_request_location_1  = 3
	λ_request_location_2  = 4
	λ_drop_off_location_1 = 3
	λ_drop_off_location_2 = 2

	γ = 0.9
	θ = 0.001

	max_cars_overflow_parking_1 = 10
	min_cars_location_1         = 0
	min_cars_location_2         = 0
	max_cars_overflow_parking_2 = 10
	//max_transferred_cars      = 5
	reward_overflow_parking_1 = -4
	reward_overflow_parking_2 = -4
	reward_rented_car         = 10
	reward_bad_move           = -1000
	reward_transferred_car    = -2

	employee_near_location_2 = true
	money_saved_by_employee  = 2
)

type State struct {
	V float64
	π Action
}

type Action struct {
	action1 int
	action2 int
}

type Mat [max_car_location_1 + 1][max_car_location_2 + 1]State

func main() {
	rand.Seed(time.Now().Unix())
	S := get_all_states()
	S = policy_iteration(S)
	print_mat(S)
	print_cars(S)
}

func print_cars(S Mat) {
	for i := 0; i <= max_car_location_1; i++ {
		for j := 0; j <= max_car_location_2; j++ {
			fmt.Print(math.Abs(float64(S[i][j].π.action1)), " ")
		}
		fmt.Println("")
	}
}

func print_mat(S Mat) {
	for i := 0; i <= max_car_location_1; i++ {
		for j := 0; j <= max_car_location_2; j++ {
			fmt.Print("(i: ", i, " j:", j, " : ", S[i][j])
		}
		fmt.Println("")
	}
}

func update_π(S Mat) (bool, Mat) {
	stable := true
	for i := 0; i <= max_car_location_1; i++ {
		for j := 0; j <= max_car_location_2; j++ {
			_action := S[i][j].π

			_i := int(math.Min(float64(i+_action.action1), float64(max_car_location_1)))
			_j := int(math.Min(float64(j+_action.action2), float64(max_car_location_2)))
			max := S[_i][_j].V
			for _, a := range get_actions(i, j) {
				__i := int(math.Min(float64(i+a.action1), float64(max_car_location_1)))
				__j := int(math.Min(float64(j+a.action2), float64(max_car_location_2)))
				if max < S[__i][__j].V {
					max = S[__i][__j].V
					_action = Action{action1: a.action1, action2: a.action2}
				}
			}
			if _action != S[i][j].π {
				S[i][j].π = _action
				stable = false
			}
		}
	}
	return stable, S
}

func policy_iteration(S Mat) Mat {
	for policy_stable := false; !policy_stable; {
		for diff := 1.0; diff > θ; {
			diff, S = update_V(S)
		}
		policy_stable, S = update_π(S)
		policy_stable = true
	}
	return S
}

func update_V(S Mat) (float64, Mat) {
	diff := 0.0
	for i := 0; i <= max_car_location_1; i++ {
		for j := 0; j <= max_car_location_2; j++ {
			_V := S[i][j].V
			S[i][j].V = get_new_V(i, j, S)
			if diff > math.Abs(_V-S[i][j].V) {
				diff += _V - S[i][j].V
			}
		}
	}
	return diff, S
}

func get_new_V(n, m int, S Mat) float64 {
	income := 0.0
	for _, a := range get_actions(n, m) {
		income += update_SV(n, m, a, S)
	}
	return income
}

func update_SV(n, m int, a Action, S Mat) float64 {
	income := get_reward(n, m, a)

	_n := n + a.action1
	_m := m + a.action2

	if _n > max_car_location_1 {
		_n = max_car_location_1
		return 0.0
	}
	if _m > max_car_location_2 {
		_m = max_car_location_2
		return 0.0
	}

	λ_req_loc_1, req_cars_1 := generate_probabilities(λ_request_location_1, max_car_location_1)
	λ_req_loc_2, req_cars_2 := generate_probabilities(λ_request_location_2, max_car_location_2)

	λ_drf_loc_1, drf_cars_1 := generate_probabilities(λ_drop_off_location_1, 0)
	λ_drf_loc_2, drf_cars_2 := generate_probabilities(λ_drop_off_location_2, 0)

	p := λ_req_loc_1 * λ_req_loc_2 * λ_drf_loc_1 * λ_drf_loc_2

	income += p * (float64(((req_cars_1 + req_cars_2 - drf_cars_1 - drf_cars_2) * reward_rented_car)) + γ*S[_n][_m].V)

	return income

}

func generate_probabilities(λ, max int) (float64, int) {
	n := 0
	p := 0.0
	for p < θ {
		p = _poisson(λ, n)
		n++
	}

	if max != 0 && n > max {
		return p, max
	}

	return p, n - 1
}

func _poisson(λ int, n int) float64 {
	return (math.Pow(float64(λ), float64(n)) / float64(factorial(n)) * math.Exp(float64(-λ)))
}

func poisson(λ int, n int) float64 {
	r := 0.0
	for i := 0; i <= n; i++ {
		r += _poisson(λ, n)
	}
	return r
}

func get_reward(n, m int, a Action) float64 {

	_n := n + a.action1
	_m := m + a.action2

	reward := math.Abs(float64(a.action1))

	if employee_near_location_2 && reward > 0 {
		reward--
	}

	reward = reward * reward_transferred_car

	if _n > max_cars_overflow_parking_1 {
		reward += reward_overflow_parking_1
	}

	if _m > max_cars_overflow_parking_2 {
		reward += reward_overflow_parking_2
	}

	return reward
}
func get_all_states() Mat {
	S := Mat{}
	for i := 0; i <= max_car_location_1; i++ {
		for j := 0; j <= max_car_location_2; j++ {
			S[i][j] = State{V: 0, π: Action{action1: 0, action2: 0}}
		}
	}
	return S
}

func get_action(n1, n2 int) Action {
	a := get_actions(n1, n2)
	return a[rand.Intn(len(a))]
}

func get_actions(n1, n2 int) []Action {
	vec := make([]Action, 1)
	for i := 1; i <= n1; i++ {
		vec = append(vec, Action{action1: -i, action2: i})
	}
	for i := 1; i <= n2; i++ {
		vec = append(vec, Action{action1: i, action2: -i})
	}

	return vec
}

func factorial(n int) uint64 {
	var factVal uint64 = 1
	if n < 0 {
		fmt.Print("Factorial of negative number doesn't exist.")
	} else {
		for i := 1; i <= n; i++ {
			factVal *= uint64(i) // mismatched types int64 and int
		}

	}
	return factVal /* return from function*/
}
