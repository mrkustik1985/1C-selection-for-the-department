package server

import (
	"fmt"
	"strconv"
	"strings"
)

func checkWin(board [3][3]int, player int) bool {
	// Проверка строк
	for i := 0; i < 3; i++ {
		if board[i][0] == player && board[i][1] == player && board[i][2] == player {
			return true
		}
	}

	// Проверка столбцов
	for i := 0; i < 3; i++ {
		if board[0][i] == player && board[1][i] == player && board[2][i] == player {
			return true
		}
	}

	// Проверка диагоналей
	if board[0][0] == player && board[1][1] == player && board[2][2] == player {
		return true
	}
	if board[0][2] == player && board[1][1] == player && board[2][0] == player {
		return true
	}

	return false
}

func checkGame(input []string) string {
	matrix := [3][3]int{
		{-1, -1, -1},
		{-1, -1, -1},
		{-1, -1, -1},
	}

	for _, line := range input {
		// Разделяем строку на части
		parts := strings.Split(line, " ")

		if len(parts) != 4 {
			fmt.Println("Неверный формат строки:", line)
			continue
		}

		// Получаем индексы в матрице
		i, err1 := strconv.Atoi(parts[2])
		j, err2 := strconv.Atoi(parts[3])
		i -= 1
		j -= 1
		if err1 != nil || err2 != nil || i > 2 || j > 2 {
			fmt.Println("Неверные индексы:", parts[2], parts[3])
			continue
		}
		if matrix[i][j] != -1 {
			return "try_add_where_stand"
		}
		// Если в строке есть "zero", ставим 0, иначе 1
		if parts[1] == "zero" {
			matrix[i][j] = 0
		} else {
			matrix[i][j] = 1
		}
		if checkWin(matrix, 1) {
			return "Eddie"
		}
		if checkWin(matrix, 0) {
			return "Player"
		}
	}
	return "nobody"
}
