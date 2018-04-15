package main

import "github.com/Reinforcement-Learning-Golang/sarsa"

/*
   Mountain car problem by Francisco Enrique Cordova Gonzalez
*/

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"strconv"
	"time"
)

//Legal actions
const (
	action_reverse int = iota - 1
	action_zero
	action_forward
)

//position and velocity bounds
const (
	position_min = -1.2
	position_max = 0.5
	velocity_min = -0.07
	velocity_max = 0.07
)

type State struct {
	position     float64
	velocity     float64
	v            *sarsa.ValueFunction
	posScale     float64
	velScale     float64
	max_position float64
	max_velocity float64
	min_position float64
	min_velocity float64
	hash_table   map[string]int
	max_size     int
}

func NewState() State {
	s := State{position: -0.5, velocity: 0}
	s.v = &sarsa.ValueFunction{}
	s.v.New(1, 2048, 8, 0.4/8)
	s.max_position = 0.5
	s.min_position = -1.2
	s.max_velocity = 0.07
	s.min_velocity = -0.07
	s.hash_table = make(map[string]int)
	s.max_size = 2048
	s.posScale = float64(s.v.Tilings) / (s.max_position - s.min_position)
	s.velScale = float64(s.v.Tilings) / (s.max_velocity - s.min_velocity)
	return s
}

func (s State) GetRandomFirstPosition() sarsa.State {
	s.position = -0.5
	s.velocity = 0
	return s
}
func (s State) GetActions() []string {
	actions := make([]string, 0)
	actions = append(actions, "reverse")
	actions = append(actions, "none")
	actions = append(actions, "forward")
	return actions
}
func (s State) GetActiveTiles(action string) [][]int {
	_pos := math.Floor(s.position * s.posScale * float64(s.v.Tilings))

	_vel := math.Floor(s.velocity * s.velScale * float64(s.v.Tilings))
	tiles := make([][]int, 1)
	tiles[0] = make([]int, 1)
	for tile := 0; tile < s.v.Tilings; tile++ {
		key := bytes.NewBufferString("") //this is the key that we'll use to save elements in our hash table

		//Using all the info that we have of this state a unique key is made
		key.WriteString(strconv.Itoa(tile))

		div := math.Floor(((_pos + float64(tile)) / float64(s.v.Tilings))) //
		key.WriteString(strconv.FormatFloat(div, 'f', -1, 64))

		div2 := math.Floor(((_vel + 3*float64(tile)) / float64(s.v.Tilings)))
		key.WriteString(strconv.FormatFloat(div2, 'f', -1, 64))

		key.WriteString(action)

		tiles[0] = append(tiles[0], s.Idx(key.String()))
	}
	return tiles
}
func (s State) InGoalState() bool {
	return s.position >= s.max_position
}
func (s State) TakeAction(action string) (sarsa.State, float64) {
	val_action := 0
	if action == "reverse" {
		val_action = -1
	}
	if action == "forward" {
		val_action = 1
	}
	newVelocity := s.velocity + 0.001*float64(val_action) - 0.0025*math.Cos(3*s.position)
	//Velocity bounds
	newVelocity = math.Min(math.Max(s.min_velocity, newVelocity), s.max_velocity)

	newPosition := s.position + newVelocity
	//Position bounds
	newPosition = math.Min(math.Max(s.min_position, newPosition), s.max_position)
	//reward's always -1
	reward := -1.0

	if newPosition == s.min_position {
		newVelocity = 0.0
	}
	s.position = newPosition
	s.velocity = newVelocity
	return s, reward
}

//*************************************************************************************************************/
func main() {
	rand.Seed(time.Now().Unix())
	state := NewState()
	episodes := 2000
	for episode := 0; episode < episodes; episode++ {
		steps := sarsa.SemiGradientSarsa(state, getAction, state.v)
		fmt.Println(episode, steps)
	}

}

func getAction(state sarsa.State, vf *sarsa.ValueFunction) string {
	values := make([]float64, 0)
	actions := state.GetActions()
	for _, action := range actions {
		values = append(values, sarsa.ValueOf(state, action, vf))
	}
	//	fmt.Println("Actions: ", values)
	ac := actions[getIdxMax(values)]
	return ac
}

func (s *State) Idx(key string) int {
	idx, ok := s.hash_table[key]
	//if the element is in the hast table the idx is returned
	if ok {
		return idx
	}

	//overflow control
	if len(s.hash_table) >= s.max_size {
		return hash(key) % len(s.hash_table)
	}
	//if the elemen is not in the hast table, the element is added.
	s.hash_table[key] = len(s.hash_table)
	return len(s.hash_table) - 1
}

func getIdxMax(slice []float64) int {
	idx := 0
	max := slice[idx]
	//get the idx of the biggest element
	for i := 1; i < len(slice); i++ {
		if max < slice[i] {
			idx = i
			max = slice[i]
		}
		//If max and slice are equal, we randomly change so have more exploration in the algorithm
		if max == slice[i] {
			if rand.Float64() <= 0.5 {
				idx = i
				max = slice[i]
			}
		}
	}
	return idx
}

//hash function
func hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}
