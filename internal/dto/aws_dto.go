package dto

// AWSLambdaResponse es la estructura completa que envía AWS Lambda
type AWSLambdaResponse struct {
	RequestID          string          `json:"request_id"`          // UUID de tracking (viene de metadata)
	InputFile          string          `json:"input_file"`
	OutputFile         string          `json:"output_file"`
	ProcessingTimeMs   int64           `json:"processing_time_ms"`
	Status             string          `json:"status"`
	StructuredData     CVProcessedData `json:"structured_data"`
}

// CVProcessedData es la estructura principal que contiene los datos extraídos y procesados del CV.
type CVProcessedData struct {
	Certifications         []Certification `json:"certifications"`
	Education              []Education     `json:"education"`
	Header                 Header          `json:"header"`
	ProfessionalExperience []Experience    `json:"professionalExperience"`
	Projects               []Project       `json:"projects"`
	TechnicalSkills        TechnicalSkills `json:"technicalSkills"`
}

// Certification representa una certificación o curso obtenido.
type Certification struct {
	DateObtained string `json:"dateObtained"`
	Name         string `json:"name"`
}

// Education representa un grado académico o educación formal.
type Education struct {
	Degree         string   `json:"degree"`
	GraduationDate string   `json:"graduationDate"`
	Institution    string   `json:"institution"`
	Achievements   []string `json:"achievements,omitempty"` // Opcional
}

// Header contiene la información básica de contacto y nombre.
type Header struct {
	Contact Contact `json:"contact"`
	Name    string  `json:"name"`
}

// Contact contiene los detalles de contacto de la persona.
type Contact struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// Experience representa una posición o empleo previo.
type Experience struct {
	Company          string   `json:"company"`
	Period           Period   `json:"period"`
	Position         string   `json:"position"`
	Responsibilities []string `json:"responsibilities"`
}

// Period representa el rango de tiempo de un empleo.
type Period struct {
	End   string `json:"end"`
	Start string `json:"start"`
}

// Project representa un proyecto personal o profesional relevante.
type Project struct {
	Description  string   `json:"description"`
	Name         string   `json:"name"`
	Technologies []string `json:"technologies"`
}

// TechnicalSkills agrupa la lista de habilidades técnicas.
type TechnicalSkills struct {
	Skills []string `json:"skills"`
}

type AWSProcessResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
