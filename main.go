package main

func main() {
	// Вариант №9 (мой)
	lpp := GetLPPFromArgs(
		[]float64{1, -3, 2, 0, 0},
		9,
		[][]float64{
			{-1, -1, -2, -1, 0},
			{1, -1, -1, 0, -1},
		},
		[]float64{-2, 1},
		true)
	lpp.Solve()
}
