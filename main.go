package main

func main() {
	// Образец из конспекта - OK
	lpp := getLPPFromArgs(
		[]float64{2, -1, 0, 0, 0},
		0,
		[][]float64{
			{-1, 2, 1, 0, 0},
			{1, -1, 0, 1, 0},
			{-1, 5, 0, 0, 1},
		},
		[]float64{4, 1, 15},
		false,
	)
	// Вариант №9 (мой) - !OK
	//lpp := getLPPFromArgs(
	//	[]float64{1, -3, 2, 0, 0},
	//	9,
	//	[][]float64{
	//		{-1, -1, -2, -1, 0},
	//		{1, -1, -1, 0, -1},
	//	},
	//	[]float64{-2, 1},
	//	true)
	lpp.Solve()
}
