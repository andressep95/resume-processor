package converter

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/nguyenthenguyen/docx"
)

// ConvertToPDF convierte archivos .txt, .doc, .docx a PDF.
// Si el archivo ya es PDF, lo retorna sin modificaciones.
// Retorna los bytes del PDF y el nombre del archivo resultante.
func ConvertToPDF(fileHeader *multipart.FileHeader) ([]byte, string, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	// Si ya es PDF, solo leer y retornar
	if ext == ".pdf" {
		pdfBytes, err := readFileToBytes(fileHeader)
		if err != nil {
			return nil, "", fmt.Errorf("error al leer archivo PDF: %w", err)
		}
		return pdfBytes, fileHeader.Filename, nil
	}

	// Convertir según el formato
	switch ext {
	case ".txt":
		return convertTextToPDF(fileHeader)
	case ".docx":
		return convertDocxToPDF(fileHeader)
	case ".doc":
		// Formato .doc (antiguo) no soportado sin LibreOffice
		return nil, "", fmt.Errorf("formato .doc no soportado. Por favor, use .docx, .txt o .pdf")
	default:
		return nil, "", fmt.Errorf("formato de archivo no soportado: %s", ext)
	}
}

// convertTextToPDF convierte un archivo de texto plano a PDF usando gofpdf
func convertTextToPDF(fileHeader *multipart.FileHeader) ([]byte, string, error) {
	// Leer contenido del archivo
	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", fmt.Errorf("error al abrir archivo de texto: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, "", fmt.Errorf("error al leer contenido del texto: %w", err)
	}

	// Crear PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	// Escribir contenido línea por línea
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		pdf.MultiCell(0, 10, line, "", "", false)
	}

	// Generar bytes del PDF
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, "", fmt.Errorf("error al generar PDF desde texto: %w", err)
	}

	// Generar nuevo nombre de archivo
	newFilename := strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename)) + ".pdf"

	return buf.Bytes(), newFilename, nil
}

// convertDocxToPDF convierte archivos .docx a PDF usando la librería docx para extraer texto
// y gofpdf para generar el PDF. No requiere LibreOffice.
func convertDocxToPDF(fileHeader *multipart.FileHeader) ([]byte, string, error) {
	// Crear archivo temporal para leer con la librería docx
	tempDir, err := os.MkdirTemp("", "docx-to-pdf-*")
	if err != nil {
		return nil, "", fmt.Errorf("error al crear directorio temporal: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Guardar archivo temporal
	inputPath := filepath.Join(tempDir, fileHeader.Filename)
	if err := saveMultipartFile(fileHeader, inputPath); err != nil {
		return nil, "", fmt.Errorf("error al guardar archivo temporal: %w", err)
	}

	// Leer el archivo .docx
	doc, err := docx.ReadDocxFile(inputPath)
	if err != nil {
		return nil, "", fmt.Errorf("error al leer archivo DOCX: %w", err)
	}
	defer doc.Close()

	// Extraer contenido y limpiar XML
	xmlContent := doc.Editable().GetContent()
	text := cleanXMLContent(xmlContent)

	// Crear PDF con el contenido extraído
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	// Escribir contenido línea por línea
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		// Manejar líneas vacías
		if strings.TrimSpace(line) == "" {
			pdf.Ln(5) // Salto de línea
			continue
		}
		pdf.MultiCell(0, 10, line, "", "", false)
	}

	// Generar bytes del PDF
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, "", fmt.Errorf("error al generar PDF desde DOCX: %w", err)
	}

	// Generar nuevo nombre de archivo
	newFilename := strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename)) + ".pdf"

	return buf.Bytes(), newFilename, nil
}

// --- Funciones auxiliares ---

// readFileToBytes lee un multipart.FileHeader y retorna sus bytes
func readFileToBytes(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

// saveMultipartFile guarda un multipart.FileHeader en una ruta del sistema de archivos
func saveMultipartFile(fileHeader *multipart.FileHeader, destPath string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

// cleanXMLContent extrae solo el texto de un contenido XML de DOCX
func cleanXMLContent(xmlContent string) string {
	// Dividir por párrafos <w:p>...</w:p>
	paragraphRe := regexp.MustCompile(`<w:p[^>]*>(.*?)</w:p>`)
	paragraphs := paragraphRe.FindAllStringSubmatch(xmlContent, -1)

	var textBuilder strings.Builder
	textRe := regexp.MustCompile(`<w:t[^>]*>([^<]*)</w:t>`)

	for _, para := range paragraphs {
		if len(para) < 2 {
			continue
		}

		// Extraer texto del párrafo
		textMatches := textRe.FindAllStringSubmatch(para[1], -1)
		if len(textMatches) == 0 {
			continue
		}

		// Concatenar texto del párrafo
		for _, match := range textMatches {
			if len(match) > 1 {
				textBuilder.WriteString(match[1])
			}
		}

		// Agregar salto de línea después de cada párrafo
		textBuilder.WriteString("\n")
	}

	return strings.TrimSpace(textBuilder.String())
}
