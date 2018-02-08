package main

import "math"

import "fmt"

import (
	"net/http"

	"github.com/wcharczuk/go-chart"
)

const (
	goal = 100
	ph   = 0.25
	θ    = 0.0000000001
)

func drawChart(res http.ResponseWriter, req *http.Request) {

	V := generate_Values()
	π, V := Value_iteration(V)

	fmt.Println(π)
	fmt.Println(V)

	xVal := make([]float64, len(π))

	for i := 0; i < len(π); i++ {
		xVal[i] = float64(i)
	}

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Style: chart.Style{
				Show: true,
			},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				Show: true,
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					Show:        true,
					StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
					FillColor:   chart.GetDefaultColor(0).WithAlpha(64),
				},
				XValues: xVal,
				YValues: V,
			},
		},
	}

	res.Header().Set("Content-Type", "image/png")
	graph.Render(chart.PNG, res)
}

func main() {

	http.HandleFunc("/", drawChart)
	http.ListenAndServe(":8080", nil)

}

func Value_iteration(V []float64) ([]float64, []float64) {

	for diff := 1.0; diff > θ; {
		diff = 0
		for s := 1; s < goal; s++ {
			v := V[s]
			V[s], _ = maxAction(s, V)
			diff += math.Abs(v - V[s])
		}
	}

	π := make([]float64, goal+1)

	for s := 1; s < goal; s++ {
		_, π[s] = maxAction(s, V)
	}

	return π, V
}

func maxAction(s int, V []float64) (float64, float64) {

	n := int(math.Min(float64(s), float64(goal-s)))

	max := ph*V[s] + (1-ph)*V[s]
	idx := 0

	for a := 1; a <= n; a++ {
		c := ph*V[s+a] + (1-ph)*V[s-a]
		if c > max {
			max = c
			idx = a
		}
	}

	return max, float64(idx)
}

func generate_Values() []float64 {
	V := make([]float64, goal+1)
	V[goal] = 1
	return V
}
