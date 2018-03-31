package main

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

const (
	action_reverse int = iota - 1
	action_zero
	action_forward
)

const (
	position_min = -1.2
	position_max = 0.5
	velocity_min = -0.07
	velocity_max = 0.07
)

func main() {
	rand.Seed(time.Now().Unix())

	number_of_tilings := 8
	alpha := 0.4
	episodes := 500

	vf := valueFunction{}
	vf.New(number_of_tilings, alpha)

	for episode := 0; episode < episodes; episode++ {
		steps := semiGradientSarsa(&vf)
		fmt.Println(episode, steps)
	}

}

func semiGradientSarsa(vf *valueFunction) int {

	//random position in range (-0.6,-0.4)
	currentPosition := -1*rand.Float64()*0.2 + 0.4
	currentVelocity := 0.0

	currentAction := vf.getAction(currentPosition, currentVelocity)

	steps := 0

	for currentPosition < position_max {

		steps += 1
		//applies the current action to the current state
		newPosition, newVelocity, reward := vf.takeAction(currentPosition, currentVelocity, currentAction)
		//Get best action given a position and a velocity
		newAction := vf.getAction(newPosition, newVelocity)

		target := vf.value(newPosition, newVelocity, newAction) + reward

		vf.learn(currentPosition, currentVelocity, target, currentAction)

		currentPosition = newPosition
		currentVelocity = newVelocity
		currentAction = newAction
	}
	return steps
}

func (v *valueFunction) takeAction(position, velocity float64, action int) (float64, float64, float64) {

	newVelocity := velocity + 0.001*float64(action) - 0.0025*math.Cos(3*position)
	//Velocity bounds
	newVelocity = math.Min(math.Max(velocity_min, newVelocity), velocity_max)

	newPosition := position + newVelocity
	//Position bounds
	newPosition = math.Min(math.Max(position_min, newPosition), position_max)
	//reward's always -1
	reward := -1.0

	if newPosition == position_min {
		newVelocity = 0.0
	}
	return newPosition, newVelocity, reward
}

func (v *valueFunction) getAction(position, velocity float64) int {
	values := make([]float64, 0)
	//slice of values for each action
	for action := action_reverse; action <= action_forward; action++ {
		values = append(values, v.value(position, velocity, action))
	}
	//get the idx of the maximum value
	return getIdxMax(values)
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
	//[0,1,2] -> [-1,0,1] (reverse,zero,forward)
	return idx - 1

}

type valueFunction struct {
	weights    []float64
	hash_table map[string]int
	tilings    int
	//We need this to normalize
	posScale float64
	velScale float64

	max_size int
	alpha    float64
}

func (v *valueFunction) learn(position, velocity, target float64, action int) {
	//get the active tiles given a position, a velocity, and an action.
	activeTiles := v.getActiveTiles(position, velocity, action)
	estimation := 0.0

	//sum of weights in our active tiles
	for _, es := range activeTiles {
		estimation += v.weights[es]
	}

	delta := v.alpha * (target - estimation)
	//update of weights using our delta in the active tiles
	for _, es := range activeTiles {
		v.weights[es] += delta
	}
}

//constructor
func (v *valueFunction) New(t int, alpha float64) {

	tilings := float64(t)
	v.weights = make([]float64, 2048)
	v.hash_table = make(map[string]int)
	v.tilings = t
	v.posScale = tilings / (position_max - position_min)
	v.velScale = tilings / (velocity_max - velocity_min)
	v.max_size = 2048
	v.alpha = alpha / tilings

}

func (v *valueFunction) value(position, velocity float64, action int) float64 {
	//The value is 0 if we are in a goal state
	if position >= position_max {
		return 0.0
	}

	idxActiveTiles := v.getActiveTiles(position, velocity, action)
	val := 0.0
	//sum of weights in the active tiles
	for _, tile := range idxActiveTiles {
		val += v.weights[tile]
	}
	return val
}

func (v *valueFunction) getActiveTiles(position, velocity float64, action int) []int {
	//normalization of pos and vel
	_pos := math.Floor(position * v.posScale * float64(v.tilings))
	_vel := math.Floor(velocity * v.velScale * float64(v.tilings))
	tiles := make([]int, 0)

	for tile := 0; tile < v.tilings; tile++ {
		key := bytes.NewBufferString("") //this is the key that we'll use to save elements in our hash table

		//Using all the info that we have of this state a unique key is made
		key.WriteString(strconv.Itoa(tile))

		div := math.Floor(((_pos + float64(tile)) / float64(v.tilings))) //
		key.WriteString(strconv.FormatFloat(div, 'f', -1, 64))

		div2 := math.Floor(((_vel + 3*float64(tile)) / float64(v.tilings)))
		key.WriteString(strconv.FormatFloat(div2, 'f', -1, 64))

		key.WriteString(strconv.Itoa(action))

		tiles = append(tiles, v.Idx(key.String()))
	}
	return tiles
}

func (v *valueFunction) Idx(key string) int {
	idx, ok := v.hash_table[key]
	//if the element is in the hast table the idx is returned
	if ok {
		return idx
	}

	//overflow control
	if len(v.hash_table) >= v.max_size {
		return hash(key) % len(v.hash_table)
	}
	//if the elemen is not in the hast table, the element is added.
	v.hash_table[key] = len(v.hash_table)
	return len(v.hash_table) - 1
}

//hash function
func hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}
