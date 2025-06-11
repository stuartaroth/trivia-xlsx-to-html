package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"net/http"
	"os"
	"strconv"
	"text/template"
)

var (
	letters = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L"}
)

type TriviaSheet struct {
	Name      string
	Questions []QuestionAnswer
}

func (ts TriviaSheet) StartLink() string {
	return GetIndexedLink(ts.Name, 0)
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

func GetIndexedLink(sheetName string, index int) string {
	return fmt.Sprintf("/?sheet=%v&index=%v", sheetName, index)
}

func GetIndexedQuestionData(triviaSheets []TriviaSheet) map[string][]QuestionData {
	indexedQuestionData := make(map[string][]QuestionData)
	for _, triviaSheet := range triviaSheets {
		var allQuestionData []QuestionData

		index := 0
		for questionIndex, question := range triviaSheet.Questions {
			questionHeader := fmt.Sprintf("Question %v", questionIndex+1)

			questionData := QuestionData{
				QuestionHeader: questionHeader,
				QuestionText:   question.Q,
				Answers:        []string{},
				BackLink:       GetIndexedLink(triviaSheet.Name, index-1),
				NextLink:       GetIndexedLink(triviaSheet.Name, index+1),
			}

			allQuestionData = append(allQuestionData, questionData)

			index++
		}

		for questionIndex, question := range triviaSheet.Questions {
			questionHeader := fmt.Sprintf("Question %v Answer", questionIndex+1)

			allQuestionData = append(allQuestionData, QuestionData{
				QuestionHeader: questionHeader,
				QuestionText:   question.Q,
				Answers:        []string{},
				BackLink:       GetIndexedLink(triviaSheet.Name, index-1),
				NextLink:       GetIndexedLink(triviaSheet.Name, index+1),
			})

			index++

			allQuestionData = append(allQuestionData, QuestionData{
				QuestionHeader: questionHeader,
				QuestionText:   question.Q,
				Answers:        question.A,
				BackLink:       GetIndexedLink(triviaSheet.Name, index-1),
				NextLink:       GetIndexedLink(triviaSheet.Name, index+1),
			})

			index++
		}

		indexedQuestionData[triviaSheet.Name] = allQuestionData
	}

	return indexedQuestionData
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
		return
	}

	sheetNames := f.GetSheetList()

	var triviaSheets []TriviaSheet

	for _, sheetName := range sheetNames {

		var questionAnswers []QuestionAnswer

		for intRow := 1; intRow < 251; intRow++ {
			var allStrings []string

			for _, letter := range letters {
				cellName := fmt.Sprintf("%v%v", letter, intRow)
				cell, err := f.GetCellValue(sheetName, cellName)
				if err != nil {
					fmt.Println("error getting cell", cellName, ", error was ", err)
					return
				}

				if cell != "" {
					allStrings = append(allStrings, cell)
				}
			}

			if len(allStrings) > 1 {
				questionAnswers = append(questionAnswers, QuestionAnswer{
					Q: allStrings[0],
					A: allStrings[1:],
				})
			}
		}

		triviaSheets = append(triviaSheets, TriviaSheet{
			Name:      sheetName,
			Questions: questionAnswers,
		})
	}

	indexedQuestionData := GetIndexedQuestionData(triviaSheets)

	templates, err := template.ParseGlob("*.html")
	if err != nil {
		fmt.Println("Error getting templates", err)
		return
	}

	sheetData := struct {
		Sheets []TriviaSheet
	}{
		Sheets: triviaSheets,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		sheet := queryParams.Get("sheet")
		if sheet == "" {
			templates.ExecuteTemplate(w, "sheets", sheetData)
			return
		}

		sheetQuestionData, hasValue := indexedQuestionData[sheet]
		if !hasValue {
			templates.ExecuteTemplate(w, "error", ErrorData{"no sheet given that name"})
			return
		}

		indexString := queryParams.Get("index")
		if indexString == "" {
			templates.ExecuteTemplate(w, "error", ErrorData{"no index provided"})
			return
		}

		indexNumber, err := strconv.Atoi(indexString)
		if err != nil {
			templates.ExecuteTemplate(w, "error", ErrorData{fmt.Sprintf("provided index did not convert: %v, %v", indexString, err)})
			return
		}

		if indexNumber < 0 {
			indexNumber = 0
		} else if indexNumber >= len(sheetQuestionData) {
			indexNumber = len(sheetQuestionData) - 1
		}

		templates.ExecuteTemplate(w, "question", sheetQuestionData[indexNumber])
	})

	fmt.Println("quizzing on http://localhost:8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server", err)
		return
	}
}
