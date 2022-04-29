package main

func main() {
	// Образец из конспекта для решения обычным
	lpp := GetLPPFromArgs([]float64{2, -1, 0, 0, 0},
		0,
		[][]float64{
			{-1, 2, 1, 0, 0},
			{1, -1, 0, 1, 0},
			{-1, 5, 0, 0, 1},
		},
		[]float64{4, 1, 15},
		false,
	)
	//lpp.Solve()
	// Модификация образца для МИБ
	lpp = GetLPPFromArgs(
		[]float64{2, -1},
		0,
		[][]float64{
			{-1, 2},
			{1, -1},
			{-1, 5},
		},
		[]float64{4, 1, 15},
		false,
	)
	//lpp.Solve()
	// Образец для решения МИБ - OK
	lpp = GetLPPFromArgs(
		[]float64{3, 4, 1, 6},
		0,
		[][]float64{
			{2, -1, 1, 3},
			{1, 3, -1, 1},
		},
		[]float64{3, 5},
		false)
	//lpp.Solve()
	// Вариант №9 (мой) - !OK
	lpp = GetLPPFromArgs(
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
