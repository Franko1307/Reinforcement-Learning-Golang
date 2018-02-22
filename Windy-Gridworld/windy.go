package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	Up int = iota
	Down
	Left
	Right
)

type SarsaTD struct {
	Q [][]float64

	Storm []int

	Qn int
	Qm int

	Sn int
	Sm int

	ter_n int
	ter_m int
	ini_n int
	ini_m int

	α float64
	ε float64
	γ float64
}

func (q *SarsaTD) Initialize() {

	q.α = 0.5
	q.ε = 0.1
	q.γ = 1

	q.Sn = 7
	q.Sm = 10

	Actions := 4 //up, down, left, right

	q.ini_n = 3
	q.ini_m = 0

	q.ter_n = 3
	q.ter_m = 7

	q.Qn = Actions
	q.Qm = q.Sn * q.Sm

	q.Q = make([][]float64, Actions)

	for i := 0; i < Actions; i++ {
		q.Q[i] = make([]float64, q.Sn*q.Sm)
	}

	for i := 0; i < Actions; i++ {
		for j := 0; j < q.Sn*q.Sm; j++ {
			q.Q[i][j] = rand.Float64()
		}
	}

	q.SetQAll(q.ter_n, q.ter_m, 0)

	q.Storm = []int{0, 0, 0, 1, 1, 1, 2, 2, 1, 0}

}

func main() {

	rand.Seed(time.Now().Unix())
	S := SarsaTD{}
	S.Initialize()
	S.Start()
	S.Pi()
}

func (q *SarsaTD) GetQ(n, m, a int) float64 {

	return q.Q[a][q.GetIdx(n, m)]
}
func (q *SarsaTD) SetQ(n, m, a int, f float64) {
	q.Q[a][q.GetIdx(n, m)] = f
}

func (q *SarsaTD) TakeAction(a, n, m int) (float64, int, int) {

	_n := n
	_m := m

	switch a {
	case Up:
		if n != q.Sn-1 {
			_n = n + 1
		}
	case Down:
		if n != 0 {
			_n = n - 1
		}
	case Left:
		if m != 0 {
			_m = m - 1
		}
	case Right:
		if m != q.Sm-1 {
			_m = m + 1
		}
	}

	_n += q.Storm[m]
	if _n > q.Sn-1 {
		_n = q.Sn - 1
	}

	if _n == q.ter_n && _m == q.ter_m {
		return 0.0, q.ter_n, q.ter_m
	}

	return -1.0, _n, _m
}

func (q *SarsaTD) ε_greedy(n, m int) int {

	Action := q.GetAction(n, m)

	if rand.Float64() < 1-q.ε {
		return Action
	}

	return rand.Intn(q.Qn)

}

func (q *SarsaTD) GetIdx(n, m int) int {
	return n*q.Sm + m
}

func (q *SarsaTD) GetAction(n, m int) int {

	Idx := q.GetIdx(n, m)
	max := q.Q[0][Idx]
	Action := 0
	for i := 1; i < q.Qn; i++ {
		if max < q.Q[i][Idx] {
			max = q.Q[i][Idx]
			Action = i
		}
	}

	return Action
}

func (q *SarsaTD) Pi() {
	for i := q.Sn - 1; i >= 0; i-- {
		for j := 0; j < q.Sm; j++ {
			if i == q.ter_n && j == q.ter_m {
				fmt.Print(" G")
			} else {
				switch q.GetAction(i, j) {
				case Up:
					fmt.Print(" U")
				case Down:
					fmt.Print(" D")
				case Left:
					fmt.Print(" L")
				case Right:
					fmt.Print(" R")
				}
			}
		}
		fmt.Println("")
	}
	for i := 0; i < len(q.Storm); i++ {
		fmt.Print(" ", q.Storm[i])
	}
	fmt.Println("")
}

func (q *SarsaTD) Start() {

	episodes := 1000
	for i := 0; i < episodes; i++ {
		Sn := q.ini_n
		Sm := q.ini_m
		Action := q.ε_greedy(Sn, Sm)
		for Sn != q.ter_n || Sm != q.ter_m {
			r, _Sn, _Sm := q.TakeAction(Action, Sn, Sm)
			_Action := q.ε_greedy(_Sn, _Sm)

			QSA := q.GetQ(Sn, Sm, Action)
			_QSA := q.GetQ(_Sn, _Sm, _Action)

			Q := QSA + q.α*(r+q.γ*_QSA-QSA)
			q.SetQ(Sn, Sm, Action, Q)

			Sn = _Sn
			Sm = _Sm
			Action = _Action
		}
	}

}

func (q *SarsaTD) SetQAll(n, m int, f float64) {
	for a := 0; a < q.Qn; a++ {
		q.Q[a][n*q.Sn+m] = f
	}
}
