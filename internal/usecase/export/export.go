package export

import (
	"context"
	"fmt"
	"github.com/xuri/excelize/v2"
	"os"
	"path/filepath"
	"strconv"
	"test_go/internal/usecase"
	"test_go/pkg/logger"
	"time"
)

type useCase struct {
	auс        usecase.Author
	buc        usecase.Book
	l          logger.Interface
	exportPath string
}

func New(
	aUc usecase.Author,
	bUc usecase.Book,
	l logger.Interface,
	exportPath string,
) *useCase {
	if err := os.MkdirAll(exportPath, 0755); err != nil {
		return nil
	}
	return &useCase{
		aUc, bUc, l, exportPath,
	}
}

func (uc *useCase) GenerateExcelFile(ctx context.Context) (*excelize.File, error) {
	authors, err := uc.auс.GetAuthors(ctx)
	if err != nil {
		return nil, fmt.Errorf("ExportUseCase - GenerateExcelFile - uc.auс.GetAuthors: %w", err)
	}

	books, err := uc.buc.GetBooks(ctx)
	if err != nil {
		return nil, fmt.Errorf("ExportUseCase - GenerateExcelFile - uc.buc.GetBooks: %w", err)
	}

	bookCount := make(map[int64]int)
	for _, book := range books {
		bookCount[book.AuthorId]++
	}
	f := excelize.NewFile()

	sheetName := "Authors"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	headers := []string{"ID", "Имя", "Пол", "Количество книг", "Статус"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	for i, author := range authors {
		row := i + 2
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), author.ID)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), author.Name)

		gender := "Женский"
		if author.Gender {
			gender = "Мужской"
		}
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), gender)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), bookCount[author.ID])

		status := "Начинающий писатель"
		if bookCount[author.ID] > 5 {
			status = "Профессионал"
		}
		f.SetCellValue(sheetName, "E"+strconv.Itoa(row), status)
	}
	lastRow := len(authors) + 2
	f.SetCellValue(sheetName, "A"+strconv.Itoa(lastRow), "Всего книг: ")
	f.SetCellFormula(sheetName, "D"+strconv.Itoa(lastRow), "SUM(D2:D"+strconv.Itoa(lastRow-1)+")")

	f.SetActiveSheet(index)
	return f, nil
}

func (uc *useCase) SaveToFile(f *excelize.File) (string, error) {
	fileName := "export_" + strconv.FormatInt(time.Now().Unix(), 10) + ".xlsx"
	filePath := filepath.Join(uc.exportPath, fileName)
	if err := f.SaveAs(filePath); err != nil {
		return "", err
	}
	return fileName, nil
}
