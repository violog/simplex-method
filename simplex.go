package main

import (
	"fmt"
	"log"
	"math"
)

type simplex struct {
	// Описание условия задачи
	targetF      []float64   // коэффициенты при переменных целевой функции (ЦФ)
	targetFValue float64     // её значение (свободный член)
	argsMatrix   [][]float64 // коэффициенты при аргументах системы ограничений (СО)
	baseValues   []float64   // значения базисных переменных (свободных членов СО)
	min          bool        // экстремум ЦФ: true - min, false - max
	// Параметры, получаемые в ходе расчётов
	baseNumbers []int     // номера текущих базисных переменных
	theta       []float64 // временный параметр для выбора разрешающей строки
	lastRow     []float64 // оценки влияния переменных на ЦФ
}

// getLPPFromArgs Получить ЗЛП из набора соответствующих аргументов
func getLPPFromArgs(targetF []float64, targetFValue float64, argsMatrix [][]float64, baseValues []float64, min bool) *simplex {
	return &simplex{
		targetF:      targetF,
		targetFValue: targetFValue,
		argsMatrix:   argsMatrix,
		baseValues:   baseValues,
		min:          min,
	}
}

// convertToPreferredView Записать в предпочтительном виде
func (s *simplex) convertToPreferredView() {
	// Делаем свободные члены системы ограничений неотрицательными
	for i, v := range s.baseValues {
		if v < 0 {
			// Если член отрицательный, обратить его и соответствующее уравнение
			myMap(s.argsMatrix[i], func(n float64) float64 { return -n })
			s.baseValues[i] = -v
		}
	}
}

// setBase Находит потенциальные столбцы для ед. матрицы и превращает её ненулевые элементы в единицы, если требуется;
// устанавливает базис из переменных, соответствующих этим столбцам
func (s *simplex) setBase() {
	// т.к. задача в каноническом виде, ищем столбцы по признаку "все нули, кроме одного"
	for j := 0; j < len(s.argsMatrix[0]); j++ {
		var (
			chosenCol    bool
			chosenVal    float64
			chosenValRow int
			foundNonZero bool
		)
		// проверяем каждый элемент столбца
		for i := 0; i < len(s.argsMatrix); i++ {
			if s.argsMatrix[i][j] != 0 {
				// если больше одного ненулевого элемента, пропустить столбец
				if foundNonZero {
					chosenCol = false
					break
				}
				foundNonZero, chosenCol = true, true
				chosenVal = s.argsMatrix[i][j]
				chosenValRow = i
			}
		}
		if !foundNonZero {
			log.Fatalln("Найден столбец с нулями, что делать?")
		}
		if chosenCol {
			// если столбец с нулями, но ненулевой элемент != 1, разделить строку на него
			// WARN это может быть неверным!
			if chosenVal != 1 {
				myMap(s.argsMatrix[chosenValRow], func(n float64) float64 { return n / chosenVal })
				// не забыть разделить и базисные переменные
				s.baseValues[chosenValRow] /= chosenVal
			}
			if len(s.baseNumbers) == 0 {
				s.baseNumbers = make([]int, len(s.argsMatrix))
			}
			s.baseNumbers[chosenValRow] = j + 1
		}
	}
}

func myMap[T any](arr []T, f func(T) T) {
	for i, v := range arr {
		arr[i] = f(v)
	}
}

func (s *simplex) printTable() {
	// Под каждое значение необходимо выделить по 6 символов: знак, 2 до, точка, 2 после
	// Исключения - i, Base
	// Выравнивание по левому краю
	const partHeader = "i  Base C      X      "
	printFloat := func(n float64) {
		if n == 0 {
			// fix printing -0
			fmt.Printf("%-7v", 0)
		} else {
			fmt.Printf("%-6.2g ", n)
			// note: see usage of %g/%G when ugly out is produced
		}
	}
	printRow := func(row []float64) {
		for _, v := range row {
			printFloat(v)
		}
	}
	printSpaces := func(spaces int) {
		for spaces > 0 {
			fmt.Print(" ")
			spaces--
		}
	}
	// 1. Выводим заголовок таблицы
	fmt.Println()
	// пробелы для коэф. при ЦФ сверху
	printSpaces(len(partHeader))
	// сами коэф. при ЦФ
	printRow(s.targetF)
	fmt.Printf("\n%s", partHeader)
	// переменные ЦФ X(i)
	for i := 0; i < len(s.targetF); i++ {
		fmt.Printf("X%-5d ", i+1) // sync with printFloat
	}
	fmt.Println("theta")
	// 2. Строки таблицы
	for i := 0; i < len(s.argsMatrix); i++ {
		fmt.Printf("%-2d ", i+1)
		// если базис не ещё установлен, заполнить пробелами, иначе вывести его
		if len(s.baseNumbers) < len(s.argsMatrix) {
			printSpaces(len(partHeader) - 3 - 7)
		} else {
			bnum := s.baseNumbers[i]
			fmt.Printf("X%-2d  ", bnum)
			printFloat(s.targetF[bnum-1])
		}
		printFloat(s.baseValues[i])
		printRow(s.argsMatrix[i])
		if i < len(s.theta) {
			if s.theta[i] < 0 {
				fmt.Printf("%-6s", "-")
			} else {
				printFloat(s.theta[i])
			}
		}
		fmt.Println()
	}
	// 3. Последняя строка
	// необходимые пробелы до столбца Х
	fmt.Printf("%d  ", len(s.argsMatrix)+1)
	printSpaces(len(partHeader) - 3 - 7)
	// значение и инвертированные коэф. ЦФ
	printFloat(s.targetFValue)
	//printRow(s.targetF, true)
	printRow(s.lastRow)
	fmt.Println()
}

// isOptimal Проверяет план на оптимальность (SRP важнее производительности в данный момент)
func (s *simplex) isOptimal() bool {
	// при поиске минимума ЦФ план оптимален, если все значения <=0
	if s.min {
		for _, v := range s.lastRow {
			if v > 0 {
				return false
			}
		}
		// при поиске максимума - все >=0
	} else {
		for _, v := range s.lastRow {
			if v < 0 {
				return false
			}
		}
	}
	return true
}

// getResolvingColumnNumber - возвращает номер разрешающего столбца, начиная с 0
// optimization: if returns 0 as min or max, plan is optimal
func (s *simplex) getResolvingColumnNumber() (n int) {
	if s.min {
		max := -math.MaxFloat64
		for i, v := range s.lastRow {
			if v > max {
				max = v
				n = i
			}
		}
	} else {
		min := math.MaxFloat64
		for i, v := range s.lastRow {
			if v < min {
				min = v
				n = i
			}
		}
	}
	return
}

// setLastRow Переписать коэф. из ЦФ в последнюю строку, инвертировав их
func (s *simplex) setLastRow() {
	s.lastRow = make([]float64, len(s.targetF))
	for i, v := range s.targetF {
		s.lastRow[i] = -v
	}
}

// getResolvingRowNumber Номер разрешающей строки, начиная с 0
func (s *simplex) getResolvingRowNumber(resColNumber int) (n int) {
	s.theta = make([]float64, len(s.argsMatrix))
	// Найдём theta делением свободных членов на разрешающий столбец
	for i, row := range s.argsMatrix {
		s.theta[i] = s.baseValues[i] / row[resColNumber]
	}
	min := math.MaxFloat64
	// Находим неотрицательный минимум - это будет номером разр. строки
	for i, v := range s.theta {
		if v >= 0 && v < min {
			min = v
			n = i
		}
	}
	return
}

// setZerosInResolvingColumn Путём матричных преобразований получить нули в разрешающем столбце
func (s *simplex) setZerosInResolvingColumn(col, row int) {
	var multiplier float64
	// получить нули в матрице аргументов
	for i, r := range s.argsMatrix {
		// пропускаем разрешающую строку
		if i == row {
			continue
		}
		// проверяем, чтобы не складывать строку с нулями
		if r[col] != 0 {
			// на какое значение умножить разреш. строку, чтобы прибавить её к текущей строке и получить 0 в разреш. столбце?
			multiplier = -r[col] / s.argsMatrix[row][col]
			sumRows(s.argsMatrix[i], s.argsMatrix[row], multiplier)
			// сложить соответствующие строкам значения базисных переменных
			s.baseValues[i] += s.baseValues[row] * multiplier
		}
	}
	// получить ноль в оценках - аналогично
	if v := s.lastRow[col]; v != 0 {
		multiplier = -v / s.argsMatrix[row][col]
		sumRows(s.lastRow, s.argsMatrix[row], multiplier)
		// учёт значения ЦФ при сложении
		s.targetFValue += s.baseValues[row] * multiplier
	}
}

// sumRows Прибавить к строке forAdding строку added, умноженную на mult
func sumRows(forAdding, added []float64, mult float64) {
	if len(forAdding) != len(added) {
		fmt.Printf("Внимание: попытка сложить строки разной длины:\n%v\n%v", forAdding, added)
		return
	}
	for i := range forAdding {
		forAdding[i] += added[i] * mult
	}
}

// Solve Собственно решение задачи
func (s *simplex) Solve() {
	// подготовка к решению
	s.setLastRow()
	fmt.Print("Канонический вид:")
	s.printTable()
	s.convertToPreferredView()
	fmt.Print("\nПредпочтительный вид:")
	s.printTable()
	s.setBase()
	fmt.Print("\nУстановлен базис:")
	s.printTable()
	fmt.Println("\nПлан неоптимален, пытаемся улучшить")
	for i := 1; !s.isOptimal(); i++ {
		resCol := s.getResolvingColumnNumber()
		resRow := s.getResolvingRowNumber(resCol)
		s.setZerosInResolvingColumn(resCol, resRow)
		fmt.Printf("\nИтерация %v\nРазрешающий столбец: X%v\nРазрешающая строка: %v", i, resCol+1, resRow+1)
		s.printTable()
		s.setBase()
		fmt.Print("Установлен базис:")
		s.printTable()
		// выход из бесконечного цикла (отладка)
		if i >= 10 {
			fmt.Println("слишком много итераций")
			return
		}
	}
	s.printCurrentAnswer()
}

// printCurrentAnswer Вывод вектора аргументов и значения ЦФ
func (s *simplex) printCurrentAnswer() {
	fmt.Print("Ответ: X* = (")
	answer := make([]float64, len(s.targetF))
	// записать в ответ значения базисных переменных, остальные по умолчанию равны 0
	for i, v := range s.baseNumbers {
		answer[v-1] = s.baseValues[i]
	}
	// вывести вектор
	for i, v := range answer {
		fmt.Print(v)
		if i < len(answer)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Printf("); Fmin = %v.\n", s.targetFValue)
}
