package handlers

import (
	"encoding/json"
	"log"
	"resume-backend-service/internal/dto"

	"github.com/gofiber/fiber/v2"
)

type AWSHandler struct{}

func NewAWSHandler() *AWSHandler {
	return &AWSHandler{}
}

func (h *AWSHandler) ProcessResumeResultsHandler(c *fiber.Ctx) error {
	// Log del body raw para debug
	bodyRaw := c.Body()
	log.Printf("📥 Body raw recibido (%d bytes): %s", len(bodyRaw), string(bodyRaw))

	// Log de headers para debug
	log.Printf("📋 Headers recibidos:")
	c.Request().Header.VisitAll(func(key, value []byte) {
		log.Printf("  %s: %s", string(key), string(value))
	})

	// Parsear la respuesta completa de AWS Lambda
	var lambdaResponse dto.AWSLambdaResponse
	if err := c.BodyParser(&lambdaResponse); err != nil {
		log.Printf("❌ Error al parsear el cuerpo de la solicitud: %v", err)

		response := dto.AWSProcessResponse{
			Status:  "error",
			Message: "Error al parsear el cuerpo de la solicitud.",
		}

		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	// Verificar el status de AWS
	if lambdaResponse.Status != "success" {
		log.Printf("⚠️  AWS reportó status: %s", lambdaResponse.Status)
	}

	// Log del JSON parseado de forma legible
	jsonPretty, _ := json.MarshalIndent(lambdaResponse.StructuredData, "", "  ")
	log.Printf("✅ CV procesado correctamente:")
	log.Printf("   📄 Input: %s", lambdaResponse.InputFile)
	log.Printf("   📄 Output: %s", lambdaResponse.OutputFile)
	log.Printf("   ⏱️  Tiempo: %dms", lambdaResponse.ProcessingTimeMs)
	log.Printf("   👤 Nombre: %s", lambdaResponse.StructuredData.Header.Name)
	log.Printf("   📧 Email: %s", lambdaResponse.StructuredData.Header.Contact.Email)
	log.Printf("   📞 Teléfono: %s", lambdaResponse.StructuredData.Header.Contact.Phone)
	log.Printf("   🎓 Educación: %d registros", len(lambdaResponse.StructuredData.Education))
	log.Printf("   💼 Experiencia: %d registros", len(lambdaResponse.StructuredData.ProfessionalExperience))
	log.Printf("   🏆 Certificaciones: %d registros", len(lambdaResponse.StructuredData.Certifications))
	log.Printf("   🚀 Proyectos: %d registros", len(lambdaResponse.StructuredData.Projects))
	log.Printf("   🛠️  Skills: %d registros", len(lambdaResponse.StructuredData.TechnicalSkills.Skills))
	log.Printf("\n📋 Datos completos:\n%s", string(jsonPretty))

	// TODO: Aquí deberías guardar los datos en la base de datos
	// Por ejemplo: h.resumeService.SaveProcessedResume(&lambdaResponse.StructuredData)

	response := dto.AWSProcessResponse{
		Status:  "success",
		Message: "Datos procesados correctamente.",
	}

	return c.Status(fiber.StatusOK).JSON(response)

}
