package export

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"os"
	"path/filepath"
	"strconv"
	"test_go/internal/usecase"
	"test_go/pkg/logger"
	"time"
)

type useCase struct {
	aUc        usecase.Author
	bUc        usecase.Book
	cUc        usecase.Command
	opUc       usecase.Operation
	l          logger.Interface
	exportPath string
}

func New(
	aUc usecase.Author,
	bUc usecase.Book,
	cUc usecase.Command,
	opUc usecase.Operation,
	l logger.Interface,
	exportPath string,
) *useCase {
	if err := os.MkdirAll(exportPath, 0755); err != nil {
		return nil
	}
	return &useCase{
		aUc, bUc, cUc, opUc, l, exportPath,
	}
}

func (uc *useCase) GenerateExcelFile(ctx context.Context) (*excelize.File, error) {
	authors, err := uc.aUc.GetAuthors(ctx)
	if err != nil {
		return nil, fmt.Errorf("ExportUseCase - GenerateExcelFile - uc.auс.GetAuthors: %w", err)
	}

	books, err := uc.bUc.GetBooks(ctx)
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

func (uc *useCase) ExportCommandsToCSV(ctx context.Context) ([]byte, string, error) {
	commands, err := uc.cUc.GetCommands(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("ExportUseCase - ExportCommandsToCSV - uc.cUc.GetCommands: %w", err)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	headers := []string{
		"ID", "Name", "SystemName", "Reagent", "AverageTime",
		"VolumeWaste", "VolumeDriveFluid", "VolumeContainer", "DefaultAddress",
	}
	if err := writer.Write(headers); err != nil {
		return nil, "", fmt.Errorf("writer.Write(headers): %w", err)
	}

	for _, cmd := range commands {
		record := []string{
			strconv.FormatInt(cmd.ID, 10),
			cmd.Name,
			cmd.SystemName,
			string(cmd.Reagent),
			strconv.FormatInt(cmd.AverageTime, 10),
			strconv.FormatInt(cmd.VolumeWaste, 10),
			strconv.FormatInt(cmd.VolumeDriveFluid, 10),
			strconv.FormatInt(cmd.VolumeContainer, 10),
			string(cmd.DefaultAddress),
		}
		if err := writer.Write(record); err != nil {
			return nil, "", fmt.Errorf("writer.Write: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", fmt.Errorf("writer.Flush: %w", err)
	}

	fileName := fmt.Sprintf("commands_%s_%s.csv", uuid.New().String()[:8], uuid.New().String()[:4])
	return buf.Bytes(), fileName, nil
}

func (uc *useCase) ExportCommandsToPDF(ctx context.Context) ([]byte, string, error) {
	commands, err := uc.cUc.GetCommands(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("ExportUseCase - ExportCommandsToPDF - uc.cUc.GetCommands: %w", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.AddUTF8Font("DejaVu", "", "fonts/dejavu-sans-ttf-2.37/ttf/DejaVuSans.ttf")
	pdf.AddUTF8Font("DejaVuBold", "B", "fonts/dejavu-sans-ttf-2.37/ttf/DejaVuSans.ttf")

	pdf.AddPage()
	pdf.SetFont("DejaVuBold", "B", 12)

	pdf.Cell(0, 10, "Команды:")
	pdf.Ln(12)

	pdf.SetFont("DejaVuBold", "B", 10)
	pdf.Cell(15, 10, "ID")
	pdf.Cell(30, 10, "Name")
	pdf.Cell(30, 10, "SystemName")
	pdf.Cell(25, 10, "Reagent")
	pdf.Cell(20, 10, "AvgTime")
	pdf.Cell(20, 10, "Waste")
	pdf.Cell(20, 10, "DriveFluid")
	pdf.Cell(20, 10, "Container")
	pdf.Cell(30, 10, "Address")
	pdf.Ln(10)

	pdf.SetFont("DejaVu", "", 10)
	for _, cmd := range commands {
		pdf.Cell(15, 10, strconv.FormatInt(cmd.ID, 10))
		pdf.Cell(30, 10, cmd.Name)
		pdf.Cell(30, 10, cmd.SystemName)
		pdf.Cell(25, 10, string(cmd.Reagent))
		pdf.Cell(20, 10, strconv.FormatInt(cmd.AverageTime, 10))
		pdf.Cell(20, 10, strconv.FormatInt(cmd.VolumeWaste, 10))
		pdf.Cell(20, 10, strconv.FormatInt(cmd.VolumeDriveFluid, 10))
		pdf.Cell(20, 10, strconv.FormatInt(cmd.VolumeContainer, 10))
		pdf.Cell(30, 10, string(cmd.DefaultAddress))
		pdf.Ln(10)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", fmt.Errorf("pdf.Output: %w", err)
	}

	fileName := fmt.Sprintf("commands_%s_%s.pdf", uuid.New().String()[:8], uuid.New().String()[:4])
	return buf.Bytes(), fileName, nil
}

func (uc *useCase) ExportOperationsToCSV(ctx context.Context) ([]byte, string, error) {
	operations, err := uc.opUc.GetOperations(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("ExportUseCase - ExportOperationsToCSV - uc.opUc.GetOperations: %w", err)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	headers := []string{"OperationID", "OperationName", "OperationDescription", "OperationAverageTime", "CommandId",
		"CommandName", "CommandSystemName", "Reagent", "AverageTime", "VolumeWaste", "VolumeDriveFluid", "VolumeContainer",
		"DefaultAddress", "Address",
	}

	if err := writer.Write(headers); err != nil {
		return nil, "", fmt.Errorf("writer.Write(headers): %w", err)
	}

	for _, op := range operations {

		if len(op.Commands) == 0 {
			record := []string{
				strconv.FormatInt(op.ID, 10),
				op.Name,
				op.Description,
				strconv.FormatInt(op.AverageTime, 10),

				"", "", "", "", "", "", "", "", "", "",
			}
			_ = writer.Write(record)
			continue
		}

		for _, oc := range op.Commands {
			c := oc.Command

			record := []string{
				strconv.FormatInt(op.ID, 10),
				op.Name,
				op.Description,
				strconv.FormatInt(op.AverageTime, 10),

				strconv.FormatInt(oc.ID, 10),
				c.Name,
				c.SystemName,
				string(c.Reagent),
				strconv.FormatInt(c.AverageTime, 10),
				strconv.FormatInt(c.VolumeWaste, 10),
				strconv.FormatInt(c.VolumeDriveFluid, 10),
				strconv.FormatInt(c.VolumeContainer, 10),
				string(c.DefaultAddress),
				string(oc.Address),
			}

			if err := writer.Write(record); err != nil {
				return nil, "", fmt.Errorf("writer.Write: %w", err)
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", fmt.Errorf("writer.Flush: %w", err)
	}

	fileName := fmt.Sprintf("operations_%s_%s.csv", uuid.New().String()[:8], uuid.New().String()[:4])
	return buf.Bytes(), fileName, nil
}

func (uc *useCase) ExportOperationsToPDF(ctx context.Context) ([]byte, string, error) {
	operations, err := uc.opUc.GetOperations(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("ExportUseCase - ExportOperationsToPDF - uc.opUc.GetOperations: %w", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.AddUTF8Font("DejaVu", "", "fonts/dejavu-sans-ttf-2.37/ttf/DejaVuSans.ttf")
	pdf.AddUTF8Font("DejaVuBold", "B", "fonts/dejavu-sans-ttf-2.37/ttf/DejaVuSans.ttf")

	pdf.AddPage()

	pdf.SetFont("DejaVuBold", "B", 14)
	pdf.Cell(0, 10, "Операции:")
	pdf.Ln(12)

	for _, op := range operations {
		pdf.SetFont("DejaVuBold", "B", 12)
		pdf.Cell(0, 8, fmt.Sprintf("Операция №%d: %s", op.ID, op.Name))
		pdf.Ln(8)

		pdf.SetFont("DejaVuBold", "B", 10)
		pdf.Cell(15, 8, "ID")
		pdf.Cell(35, 8, "Name")
		pdf.Cell(30, 8, "System")
		pdf.Cell(20, 8, "Reagent")
		pdf.Cell(15, 8, "Time")
		pdf.Cell(15, 8, "Waste")
		pdf.Cell(15, 8, "Drive")
		pdf.Cell(15, 8, "Cont")
		pdf.Cell(30, 8, "Address")
		pdf.Ln(8)

		if len(op.Commands) == 0 {
			pdf.SetFont("DejaVu", "", 10)
			pdf.Cell(0, 8, "Нет команд")
			pdf.Ln(10)
			continue
		}

		pdf.SetFont("DejaVu", "", 10)
		for _, oc := range op.Commands {
			cmd := oc.Command

			pdf.Cell(15, 8, strconv.FormatInt(cmd.ID, 10))
			pdf.Cell(35, 8, cmd.Name)
			pdf.Cell(30, 8, cmd.SystemName)
			pdf.Cell(20, 8, string(cmd.Reagent))
			pdf.Cell(15, 8, strconv.FormatInt(cmd.AverageTime, 10))
			pdf.Cell(15, 8, strconv.FormatInt(cmd.VolumeWaste, 10))
			pdf.Cell(15, 8, strconv.FormatInt(cmd.VolumeDriveFluid, 10))
			pdf.Cell(15, 8, strconv.FormatInt(cmd.VolumeContainer, 10))
			pdf.Cell(30, 8, string(oc.Address))
			pdf.Ln(8)
		}

		pdf.Ln(5)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", fmt.Errorf("pdf.Output: %w", err)
	}

	fileName := fmt.Sprintf("operations_%s_%s.pdf", uuid.New().String()[:8], uuid.New().String()[:4])
	return buf.Bytes(), fileName, nil
}
