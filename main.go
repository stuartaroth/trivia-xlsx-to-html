package main

import (
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"text/template"
)

var (
	maxCellRegex = regexp.MustCompile(":(\\w)(\\d+)")
)

func LetterToInt(letter string) (int, error) {
	switch letter {
	case "A":
		return 1, nil
	case "B":
		return 2, nil
	case "C":
		return 3, nil
	case "D":
		return 4, nil
	case "E":
		return 5, nil
	case "F":
		return 6, nil
	case "G":
		return 7, nil
	case "H":
		return 8, nil
	case "I":
		return 9, nil
	case "J":
		return 10, nil
	case "K":
		return 11, nil
	case "L":
		return 12, nil
	case "M":
		return 13, nil
	case "N":
		return 14, nil
	case "O":
		return 15, nil
	case "P":
		return 16, nil
	case "Q":
		return 17, nil
	case "R":
		return 18, nil
	case "S":
		return 19, nil
	case "T":
		return 20, nil
	case "U":
		return 21, nil
	case "V":
		return 22, nil
	case "W":
		return 23, nil
	case "X":
		return 24, nil
	case "Y":
		return 25, nil
	case "Z":
		return 26, nil
	default:
		return 0, fmt.Errorf("invalid input: %q is not a letter A-Z", letter)
	}
}

func IntToLetter(n int) (string, error) {
	switch n {
	case 1:
		return "A", nil
	case 2:
		return "B", nil
	case 3:
		return "C", nil
	case 4:
		return "D", nil
	case 5:
		return "E", nil
	case 6:
		return "F", nil
	case 7:
		return "G", nil
	case 8:
		return "H", nil
	case 9:
		return "I", nil
	case 10:
		return "J", nil
	case 11:
		return "K", nil
	case 12:
		return "L", nil
	case 13:
		return "M", nil
	case 14:
		return "N", nil
	case 15:
		return "O", nil
	case 16:
		return "P", nil
	case 17:
		return "Q", nil
	case 18:
		return "R", nil
	case 19:
		return "S", nil
	case 20:
		return "T", nil
	case 21:
		return "U", nil
	case 22:
		return "V", nil
	case 23:
		return "W", nil
	case 24:
		return "X", nil
	case 25:
		return "Y", nil
	case 26:
		return "Z", nil
	default:
		return "", fmt.Errorf("invalid input: %d is not in range 1-26", n)
	}
}

func getMaxCells(f *excelize.File, sheetName string) (int, int, error) {
	dimensions, err := f.GetSheetDimension(sheetName)
	if err != nil {
		return 0, 0, err
	}

	results := maxCellRegex.FindStringSubmatch(dimensions)
	if len(results) != 3 {
		return 0, 0, errors.New("bad regex")
	}

	stringColumn := results[1]
	stringRow := results[2]

	intColumn, err := LetterToInt(stringColumn)
	if err != nil {
		return 0, 0, err
	}

	intRow, err := strconv.Atoi(stringRow)
	if err != nil {
		return 0, 0, err
	}

	return intColumn, intRow, nil
}

func intRowIntColumnToCell(row, column int) (string, error) {
	columnString, err := IntToLetter(column)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v%v", columnString, row), nil
}

type TriviaSheet struct {
	Name      string
	Questions []QuestionAnswer
}

func (ts TriviaSheet) StartLink() string {
	return FormatLink(ts.Name, 1, 0)
}

func FormatLink(sheetName string, question int, state int) string {
	return fmt.Sprintf("/?sheet=%v&question=%v&state=%v", sheetName, question, state)
}

type QuestionAnswer struct {
	Q string
	A []string
}

type QuestionData struct {
	QuestionHeader string
	QuestionText   string
	Answers        []string
	BackLink       string
	NextLink       string
}

type ErrorData struct {
	Error string
}

func main() {
	fmt.Println(os.Args)
	if len(os.Args) < 2 {
		fmt.Println("You must provide an *.xlsx file")
		return
	}

	xlsxFilename := os.Args[1]

	f, err := excelize.OpenFile(xlsxFilename)
	if err != nil {
		fmt.Println("error reading ", xlsxFilename, ", error was ", err)
	}

	sheetNames := f.GetSheetList()

	var triviaSheets []TriviaSheet

	for _, sheetName := range sheetNames {
		columns, rows, err := getMaxCells(f, sheetName)
		if err != nil {
			fmt.Println("getMaxCells error", err)
			return
		}

		var questionAnswers []QuestionAnswer

		for intRow := 1; intRow < rows+1; intRow++ {
			var allStrings []string

			for intColumn := 1; intColumn < columns+1; intColumn++ {
				cellName, err := intRowIntColumnToCell(intRow, intColumn)
				if err != nil {
					fmt.Println("error getting cellname", cellName, ", error was ", err)
					return
				}

				cell, err := f.GetCellValue(sheetName, cellName)
				if err != nil {
					fmt.Println("error getting cell", cellName, ", error was ", err)
					return
				}

				if cell != "" {
					allStrings = append(allStrings, cell)
				}
			}

			if len(allStrings) < 2 {
				fmt.Println("not enough text for question and answer: ", allStrings)
				return
			}

			questionAnswers = append(questionAnswers, QuestionAnswer{
				Q: allStrings[0],
				A: allStrings[1:],
			})
		}

		triviaSheets = append(triviaSheets, TriviaSheet{
			Name:      sheetName,
			Questions: questionAnswers,
		})
	}

	mappy := make(map[string]TriviaSheet)
	for _, triviaSheet := range triviaSheets {
		mappy[triviaSheet.Name] = triviaSheet
	}

	templates, err := template.ParseGlob("*.html")
	if err != nil {
		fmt.Println("Error getting templates", err)
		return
	}

	templateData := struct {
		Sheets []TriviaSheet
	}{
		Sheets: triviaSheets,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		queryParams := r.URL.Query()
		sheet := queryParams.Get("sheet")
		if sheet == "" {
			templates.ExecuteTemplate(w, "sheets", templateData)
			return
		}

		storedSheet, hasValue := mappy[sheet]
		if !hasValue {
			templates.ExecuteTemplate(w, "error", ErrorData{"no sheet given that name"})
			return
		}

		questionString := queryParams.Get("question")
		if questionString == "" {
			templates.ExecuteTemplate(w, "error", ErrorData{"no question provided"})
			return
		}

		questionNumber, err := strconv.Atoi(questionString)
		if err != nil {
			templates.ExecuteTemplate(w, "error", ErrorData{fmt.Sprintf("provided question did not convert: %v, %v", questionString, err)})
			return
		}

		if questionNumber < 1 {
			questionNumber = 1
		}

		questions := storedSheet.Questions

		if questionNumber > len(questions) {
			templates.ExecuteTemplate(w, "error", ErrorData{"bitch thats too many"})
			return
		}

		currentQuestion := questions[questionNumber-1]

		stateString := queryParams.Get("state")
		if stateString != "0" && stateString != "1" && stateString != "2" {
			templates.ExecuteTemplate(w, "error", ErrorData{fmt.Sprintf("%v aint no state Ive ever heard of", stateString)})
			return
		}

		var answers []string
		if stateString == "2" {
			answers = currentQuestion.A
		}

		var backLink string
		var nextLink string

		if stateString == "0" {
			backLink = FormatLink(sheet, questionNumber-1, 0)

			if questionNumber == len(questions) {
				nextLink = FormatLink(sheet, 1, 1)
			} else {
				nextLink = FormatLink(sheet, questionNumber+1, 0)
			}
		}

		if stateString == "1" {
			if questionNumber == 1 {
				backLink = FormatLink(sheet, len(questions), 0)
			} else {
				backLink = FormatLink(sheet, questionNumber-1, 2)
			}

			nextLink = FormatLink(sheet, questionNumber, 2)
		}

		if stateString == "2" {
			backLink = FormatLink(sheet, questionNumber, 1)

			if questionNumber == len(questions) {
				nextLink = FormatLink(sheet, len(questions), 2)
			} else {
				nextLink = FormatLink(sheet, questionNumber+1, 1)
			}
		}

		var questionHeader string
		if stateString == "0" {
			questionHeader = fmt.Sprintf("Question %v", questionNumber)
		} else {
			questionHeader = fmt.Sprintf("Question %v Answer", questionNumber)
		}

		questionData := QuestionData{
			QuestionHeader: questionHeader,
			QuestionText:   currentQuestion.Q,
			Answers:        answers,
			BackLink:       backLink,
			NextLink:       nextLink,
		}

		templates.ExecuteTemplate(w, "question", questionData)

	})

	fmt.Println("quizzing on http://localhost:8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server", err)
		return
	}
}
