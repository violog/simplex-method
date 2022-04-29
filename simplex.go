package main

import (
	"fmt"
	"log"
	"math"
	"os"
)

type simplex struct {
	// Описание условия задачи
	targetF []float64 // коэффициенты при переменных целевой функции (ЦФ)
	//targetFTerm float64     // свободный член ЦФ
	// if everything works out, I'll leave targetFValue
	targetFValue float64     // текущее значение ЦФ
	argsMatrix   [][]float64 // коэффициенты при аргументах системы ограничений (СО)
	baseValues   []float64   // значения базисных переменных (свободных членов СО)
	min          bool        // экстремум ЦФ: true - min, false - max
	// Параметры, получаемые в ходе расчётов
	baseNumbers []int     // номера текущих базисных переменных
	theta       []float64 // временный параметр для выбора направляющей строки
	delta       []float64 // оценки влияния свободных переменных на ЦФ
	deltaM      []float64 // оценки влияния коэф. при М на ЦФ
}

// GetLPPFromArgs Получить ЗЛП из набора соответствующих аргументов
func GetLPPFromArgs(targetF []float64, targetFValue float64, argsMatrix [][]float64, baseValues []float64, min bool) *simplex {
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
	//for j:=0;j<len(s.argsMatrix);j++{}
	//var added uint
	//getAppendingArray := func() []float64 {
	//	arr := make([]float64, len(s.argsMatrix))
	//	// Единица на месте соответствующей искусственной переменной
	//	arr[added] = 1
	//	return arr
	//}
	// Добавляем искусственный базис
	m := len(s.argsMatrix)
	// Получаем список значений для вставки в строку-ограничение (вынесено из цикла для оптимизации)
	arr := make([]float64, m)
	s.baseNumbers = make([]int, m)
	for added, row := range s.argsMatrix {
		// Единица на месте соответствующей искусственной переменной
		arr[added] = 1
		s.argsMatrix[added] = append(row, arr...)
		// Обнулить для последующей итерации
		arr[added] = 0
		//added++
		// Делаем переменную базисной
		s.baseNumbers[added] = len(row) + added + 1
		//if added < m-1 {
		// Нельзя получить константную длину, поэтому приходится брать длину след. строки
		//s.baseNumbers[added] = len(s.argsMatrix[m-1]) + added + 1
		//continue
		//}
		// В последнем ограничении номер баз. перем. равен текущей длине
		//s.baseNumbers[added] = len(row)
	}
	// todo what to do with s.targetF? Should I add synthetic vars with 0?
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
		}
	}
	printM := func() { // sync with printFloat
		if s.min {
			fmt.Printf("%-6s ", "M")
		} else {
			fmt.Printf("%-6s ", "-M")
		}
	}
	printX := func(n int) {
		fmt.Printf("X%-5d ", n) // sync with printFloat
	}
	printRow := func(row []float64) {
		for _, v := range row {
			printFloat(v)
		}
	}
	printSpaces := func(count int) {
		for count > 0 {
			fmt.Print(" ")
			count--
		}
	}
	m := len(s.argsMatrix)
	// 1. Выводим заголовок таблицы
	fmt.Println()
	// пробелы для коэф. при ЦФ сверху
	printSpaces(len(partHeader))
	// сами коэф. при ЦФ
	printRow(s.targetF)
	// коэф. при искусственных переменных
	//if len(s.baseNumbers) > 0 {
	for i := 0; i < len(s.baseNumbers); i++ {
		printM()
	}
	//}
	fmt.Printf("\n%s", partHeader)
	// переменные ЦФ X(i)
	for i := 1; i <= len(s.targetF); i++ {
		printX(i)
	}
	// искусств. пер. X(i)
	for i := len(s.targetF) + 1; i <= len(s.argsMatrix[0]); i++ {
		printX(i)
	}
	fmt.Println("theta")
	// 2. Основные строки таблицы
	for i := 0; i < m; i++ {
		fmt.Printf("%-2d ", i+1)
		// если базис не ещё установлен, заполнить пробелами, иначе вывести его
		if len(s.baseNumbers) < m {
			printSpaces(len(partHeader) - 3 - 7)
		} else {
			bnum := s.baseNumbers[i]
			fmt.Printf("X%-2d  ", bnum)
			// коэф. при иск. пер. вывести отдельно
			if bnum > len(s.targetF) {
				printM()
			} else {
				printFloat(s.targetF[bnum-1])
			}
		}
		printFloat(s.baseValues[i])
		printRow(s.argsMatrix[i])
		if i < len(s.theta) {
			if s.theta[i] <= 0 {
				fmt.Printf("%-6s", "-")
			} else {
				printFloat(s.theta[i])
			}
		}
		fmt.Println()
	}
	// 3. Вектор оценок для свободных переменных
	// необходимые пробелы до столбца Х
	if len(s.delta) > 0 {
		fmt.Printf("%d  ", m+1)
		printSpaces(len(partHeader) - 3 - 7)
		// значение и инвертированные коэф. ЦФ
		printFloat(s.targetFValue)
		printRow(s.delta)
	}
	fmt.Println()
	// 4. Вектор оценок для коэф. при М
	if len(s.deltaM) > 0 {
		fmt.Printf("%d  ", m+2)
		printSpaces(len(partHeader) - 3 - 7)
		printRow(s.deltaM)
	}
	fmt.Println()
}

// isOptimal Проверяет план на оптимальность (SRP важнее производительности в данный момент)
func (s *simplex) isOptimal() bool {
	if s.min {
		// при поиске минимума ЦФ план оптимален, если все значения <=0
		for _, v := range s.delta {
			if v > 0 {
				return false
			}
		}
	} else {
		// при поиске максимума - все >=0
		for _, v := range s.delta {
			if v < 0 {
				return false
			}
		}
	}
	return true
}

// getDirectiveColumnNumber - возвращает номер направляющего, начиная с 0
func (s *simplex) getDirectiveColumnNumber() (number int) {
	// к-во неискусственных переменных
	n := len(s.argsMatrix[0]) - len(s.argsMatrix)
	var syntheticVarsInBase bool
	// Проверяем наличие иск. пер. в базисе
	for _, v := range s.baseNumbers {
		if v > len(s.targetF) {
			syntheticVarsInBase = true
			break
		}
	}
	if s.min {
		// Ищем макс. положительное значение среди оценок
		max := -math.MaxFloat64
		// В обоих векторах оценок не учитываем значения при искусственных переменных,
		// т. к. они не должны возвращаться в базис
		for i, v := range s.delta[:n] {
			if v > max {
				max = v
				number = i
				// Другой вариант алгоритма состоит в возвращении первого попавшегося значения
				// Здесь ищем максимум среди всех
			}
		}
		// Если есть иск. пер., проверяем последнюю строку, иначе игнорируем
		if syntheticVarsInBase {
			// Сдвигаем массив вправо на 1 из-за свободного члена
			for i, v := range s.deltaM[1 : n+1] {
				if v > max {
					max = v
					number = i
				}
			}
		}
		// Если макс. неположителен, план оптимален
		if max <= 0 {
			return -1
		}
	} else {
		min := math.MaxFloat64
		// Ищем мин. по тем же принципам
		for i, v := range s.delta[:n] {
			if v < min {
				min = v
				number = i
			}
		}
		if syntheticVarsInBase {
			for i, v := range s.deltaM[1 : n+1] {
				if v < min {
					min = v
					number = i
				}
			}
		}
		// Если мин. неотрицателен, план оптимален
		if min >= 0 {
			return -1
		}
	}
	return
}

// setDeltas Установить векторы оценок для свободных членов и коэф. при М
func (s *simplex) setDeltas() {
	// К-во ограничений и неискусственных переменных
	m := len(s.argsMatrix)
	n := len(s.argsMatrix[0]) - m
	// Для свободных членов просто инвертируем коэф. при ЦФ
	s.delta = make([]float64, len(s.targetF)+m)
	for i, v := range s.targetF {
		s.delta[i] = -v
	}
	// Установим оценки для коэф. при М
	s.deltaM = make([]float64, len(s.targetF)+m+1)
	// Если выразить остальные переменные через искусственные, то коэф.
	// при каждой из первых будет равен инвертированной сумме коэф.
	// во всех ограничениях, а свободный член будет инвертирован
	//coefs := make([]float64, n+1)
	// Устанавливаем свободный член
	for _, v := range s.baseValues {
		s.deltaM[0] -= v
	}
	// Инвертируем сумму значений для соответствующих переменных
	for j := 1; j <= n; j++ {
		for i := 0; i < m; i++ {
			s.deltaM[j] -= s.argsMatrix[i][j-1]
		}
	}
	// Искусственные переменные остаются нулевыми
}

// getDirectiveRowNumber Номер направляющей строки, начиная с 0
func (s *simplex) getDirectiveRowNumber(resColNumber int) (number int) {
	s.theta = make([]float64, len(s.argsMatrix))
	// В первую очередь выведем искусственные переменные
	for i, v := range s.baseNumbers {
		// Их номера больше к-ва аргументов в ЦФ (мы не добавляли иск. пер. в неё)
		if v > len(s.targetF) {
			return i
		}
	}
	// Найдём theta делением свободных членов на напр. столбец
	for i, row := range s.argsMatrix {
		// Защита от деления на 0 (отрицательные значения не отображаются)
		if row[resColNumber] == 0 {
			s.theta[i] = -1
			fmt.Printf("Внимание: попытка деления на 0 при получении theta, столбец X%v\n", resColNumber+1)
			continue
		}
		s.theta[i] = s.baseValues[i] / row[resColNumber]
	}
	min := math.MaxFloat64
	minFound := false
	// Находим положительный минимум - это будет номером напр. строки
	for i, v := range s.theta {
		if v > 0 && v <= min {
			minFound = true
			min = v
			number = i
		}
	}
	if !minFound {
		dir := "сверху"
		if s.min {
			dir = "снизу"
		}
		fmt.Println("Задача не имеет решений: функция не ограничена", dir)
		os.Exit(0)
	}
	return
}

// setZerosInColumn Путём матричных преобразований получить нули в направляющем столбце
func (s *simplex) setZerosInColumn(col, row int) {
	var multiplier float64
	// получить нули в матрице аргументов
	for i, r := range s.argsMatrix {
		// пропускаем напр. строку
		if i == row {
			continue
		}
		// проверяем, чтобы не складывать строку с нулями
		if r[col] != 0 {
			// на какое значение умножить напр. строку, чтобы прибавить её к текущей строке и получить 0 в напр. столбце?
			multiplier = -r[col] / s.argsMatrix[row][col]
			sumRows(s.argsMatrix[i], s.argsMatrix[row], multiplier)
			// сложить соответствующие строкам значения базисных переменных
			s.baseValues[i] += s.baseValues[row] * multiplier
		}
	}
	// получить ноль в оценках - аналогично
	if v := s.delta[col]; v != 0 {
		multiplier = -v / s.argsMatrix[row][col]
		sumRows(s.delta, s.argsMatrix[row], multiplier)
		// учёт значения ЦФ при сложении
		s.targetFValue += s.baseValues[row] * multiplier
	}
	// в оценках коэф. при М аналогично, с оговорками
	if v := s.deltaM[col+1]; v != 0 {
		multiplier = -v / s.argsMatrix[row][col]
		// не учитываем первый элемент (свободный член)
		sumRows(s.deltaM[1:], s.argsMatrix[row], multiplier)
		// его складываем отдельно
		s.deltaM[0] += s.baseValues[row] * multiplier
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

// newBase Заменить базисную переменную
func (s *simplex) newBase(popped, pushed int) {
	// Делаем базис единичным
	if v := s.argsMatrix[popped][pushed]; v != 1 {
		if v == 0 {
			log.Fatalf("Попытка деления на ноль: popped=%v, pushed=%v\n", popped, pushed)
		}
		myMap(s.argsMatrix[popped], func(n float64) float64 { return n / v })
		// не забываем разделить значения баз. пер.
		s.baseValues[popped] /= v
	}
	s.baseNumbers[popped] = pushed + 1
	// Ищем выводимую переменную и вводим новую на её место
	//for i, v := range s.baseNumbers {
	//	if v == popped+1 {
	//		s.baseNumbers[i] = pushed + 1
	//		break
	//	}
	//}
}

// Solve Собственно решение задачи
func (s *simplex) Solve() {
	// подготовка к решению
	fmt.Print("Канонический вид (условие задачи):")
	s.printTable()
	s.convertToPreferredView()
	fmt.Print("\nПредпочтительный вид:")
	s.printTable()
	s.setDeltas()
	fmt.Print("\nНайдены оценки:")
	s.printTable()
	// Проверим, что план не оптимален
	if resCol := s.getDirectiveColumnNumber(); resCol != -1 {
		fmt.Println("\nПлан неоптимален, пытаемся улучшить")
		for i := 1; resCol != -1; i++ {
			resRow := s.getDirectiveRowNumber(resCol)
			s.setZerosInColumn(resCol, resRow)
			fmt.Printf("\nИтерация %v\nРазрешающий столбец: X%v\nРазрешающая строка: %v", i, resCol+1, resRow+1)
			s.printTable()
			s.newBase(resRow, resCol)
			fmt.Print("Установлен базис:")
			s.printTable()
			// Вычисляем след. напр. столбец
			resCol = s.getDirectiveColumnNumber()
			// note выход из бесконечного цикла (отладка)
			if i >= 10 {
				fmt.Println("слишком много итераций")
				return
			}
		}
	} else {
		fmt.Println("\nПлан уже оптимален")
	}
	fmt.Println()
	s.printCurrentAnswer()
	fmt.Print("\n\n")
}

// printCurrentAnswer Вывод вектора аргументов и значения ЦФ
func (s *simplex) printCurrentAnswer() {
	answer := make([]float64, len(s.targetF))
	// записать в ответ значения базисных переменных, остальные по умолчанию равны 0
	for i, v := range s.baseNumbers {
		if v > len(s.targetF) {
			fmt.Println("Задача не имеет решений: искусственные переменные остались в базисе")
			os.Exit(0)
		}
		answer[v-1] = s.baseValues[i]
	}
	fmt.Print("Ответ: X* = (")
	// вывести вектор
	for i, v := range answer {
		fmt.Printf("%.4g", v)
		if i < len(answer)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Printf("); Fmin = %.4g.\n", s.targetFValue)
}
