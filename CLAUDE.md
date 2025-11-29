# CLAUDE.md - Contexto del Proyecto para Claude

Este documento proporciona contexto completo sobre el proyecto **Resume Backend Service** para facilitar la colaboración con Claude Code.

## 📋 Descripción del Proyecto

**Resume Backend Service** es un microservicio backend en Go que procesa currículums (CVs) de forma asíncrona mediante integración con AWS Lambda y S3. El servicio acepta archivos en múltiples formatos (.pdf, .txt, .docx), los convierte a PDF, los sube a S3 para procesamiento, y recibe los resultados estructurados mediante un webhook callback.

### Propósito Principal
- Recibir CVs en diferentes formatos
- Convertir archivos a PDF estandarizado
- Subir a S3 con metadatos personalizados
- Procesar información del CV mediante AWS Lambda
- Almacenar datos estructurados extraídos (TODO: implementar persistencia)

---

## 🏗️ Arquitectura del Proyecto

### Estructura de Directorios

```
resume-backend-service/
├── cmd/                          # Punto de entrada
│   └── main.go                   # Main minimalista (9 líneas)
│
├── internal/                     # Código privado de la aplicación
│   ├── config/
│   │   ├── bootstrap.go          # Inicialización de la app (50 líneas)
│   │   └── config.go             # Variables de entorno (57 líneas)
│   │
│   ├── dto/                      # Data Transfer Objects
│   │   ├── resume_dto.go         # Response de procesamiento
│   │   ├── aws_dto.go            # Estructuras Lambda (77 líneas)
│   │   └── presigned_url_dto.go  # DTOs para URLs firmadas
│   │
│   ├── handlers/                 # HTTP Handlers
│   │   ├── resume_handler.go     # Upload de CVs (56 líneas)
│   │   ├── aws_handler.go        # Callback Lambda (72 líneas)
│   │   └── health_handler.go     # Health check (24 líneas)
│   │
│   ├── services/
│   │   └── resume_service.go     # Lógica de negocio (109 líneas)
│   │
│   ├── router/
│   │   └── router.go             # Definición de rutas (36 líneas)
│   │
│   ├── domain/                   # VACÍO - Preparado para entidades
│   ├── middleware/               # VACÍO - Preparado para middlewares
│   └── repository/               # VACÍO - Preparado para persistencia
│
├── pkg/                          # Código reutilizable
│   ├── converter/
│   │   └── pdf_converter.go      # Conversión archivos (167 líneas)
│   │
│   ├── client/
│   │   └── presigned_url_client.go  # Cliente HTTP (80 líneas)
│   │
│   ├── utils/                    # VACÍO
│   └── validator/                # VACÍO
│
├── docs/
│   └── resume-backend-api.yaml   # Especificación OpenAPI 3.0
│
├── Dockerfile                    # Multi-stage build
├── docker-compose.yml            # PostgreSQL + Backend
├── Makefile                      # Comandos útiles
├── .env.example                  # Template de configuración
├── go.mod / go.sum              # Dependencias
└── README.md
```

**Total de código:** ~508 líneas Go distribuidas en 13 archivos

---

## 🔄 Flujos Principales

### 1. Flujo de Procesamiento de CV (Completo)

```
┌─────────────────────────────────────────────────────────────────┐
│ CLIENTE                                                         │
│ POST /api/v1/resume/                                            │
│ - file: resume.pdf                                              │
│ - language: esp                                                 │
│ - instructions: "Extraer últimos 5 años"                       │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ RESUME HANDLER (resume_handler.go:ProcessResumeHandler)        │
│ - Extrae form fields                                            │
│ - Valida presencia del archivo                                  │
│ - Delega a ResumeService                                        │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ RESUME SERVICE (resume_service.go:ProcessResume)               │
│ 1. Validación de formato (.pdf, .txt, .docx permitidos)        │
│ 2. Conversión a PDF (pkg/converter/pdf_converter.go)           │
│ 3. Obtiene presigned URL del servicio externo                   │
│ 4. Upload a S3 con metadatos (language, instructions)          │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ PRESIGNED URL CLIENT (presigned_url_client.go)                 │
│ POST https://api.cloudcentinel.com/.../presigned-url/upload    │
│ Retorna: { url: "https://s3.../", expires_in: "1 hour" }      │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ UPLOAD A S3                                                     │
│ PUT presigned_url                                               │
│ Headers:                                                         │
│   - Content-Type: application/pdf                               │
│   - x-amz-meta-language: esp                                    │
│   - x-amz-meta-instructions: "..."                              │
│ Path: s3://.../inputs/2025-11-29/HH:MM/cv-clean.pdf           │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ RESPUESTA AL CLIENTE                                            │
│ 202 Accepted                                                    │
│ { "status": "accepted",                                         │
│   "message": "Solicitud encolada para procesamiento." }        │
└─────────────────────────────────────────────────────────────────┘

════════════════════════════════════════════════════════════════════
(Procesamiento asíncrono en AWS)
════════════════════════════════════════════════════════════════════

┌─────────────────────────────────────────────────────────────────┐
│ AWS S3 EVENT TRIGGER                                            │
│ S3 detecta PUT en /inputs/ → dispara Lambda                     │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ AWS LAMBDA                                                      │
│ Procesa PDF y extrae:                                           │
│ - Header (nombre, email, teléfono)                             │
│ - Educación (institución, grado, fecha)                         │
│ - Experiencia laboral (empresa, cargo, período)                │
│ - Habilidades técnicas                                          │
│ - Proyectos                                                      │
│ - Certificaciones                                               │
│                                                                  │
│ Sube resultado: s3://.../outputs/2025-11-29/.../cv-clean.json │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ CALLBACK A BACKEND                                              │
│ POST /api/v1/resume/results                                     │
│ Body: AWSLambdaResponse con CVProcessedData                     │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ AWS HANDLER (aws_handler.go:ProcessResumeResultsHandler)       │
│ - Parsea respuesta de Lambda                                    │
│ - Extrae metadata y datos estructurados                         │
│ - Logging detallado de resultados                               │
│ - TODO: Guardar en base de datos                                │
│                                                                  │
│ Respuesta: 200 OK { status: "success", ... }                   │
└─────────────────────────────────────────────────────────────────┘
```

### 2. Health Check
```
GET /api/v1/health/
→ HealthHandler.HandleHealthCheck()
→ { "status": "healthy", "service": "resume-backend-service" }
```

---

## 🛠️ Tecnologías y Dependencias

### Dependencias Principales (go.mod)

```go
github.com/gofiber/fiber/v2 v2.52.10     // Framework HTTP rápido
github.com/joho/godotenv v1.5.1          // Carga .env
github.com/jung-kurt/gofpdf v1.16.2      // Generación de PDFs
github.com/nguyenthenguyen/docx v0.0.0   // Lectura de archivos DOCX
```

### Stack Tecnológico

| Categoría | Tecnología | Versión | Uso |
|-----------|-----------|---------|-----|
| **Lenguaje** | Go | 1.24.5 | Backend |
| **Framework HTTP** | Fiber | v2.52.10 | Servidor REST API |
| **Conversión PDF** | gofpdf | v1.16.2 | TXT/DOCX → PDF |
| **Lectura DOCX** | docx | v0.0.0 | Extracción de texto |
| **Configuración** | godotenv | v1.5.1 | Variables de entorno |
| **Base de Datos** | PostgreSQL | 16 | (Preparado, no usado aún) |
| **Contenedores** | Docker | Latest | Empaquetamiento |
| **Orquestación** | Docker Compose | - | Entorno local |
| **Almacenamiento** | AWS S3 | - | Archivos input/output |
| **Procesamiento** | AWS Lambda | - | Extracción de datos |

---

## 🌐 Endpoints de la API

### GET /api/v1/health/
**Handler:** `health_handler.go:HandleHealthCheck()`

**Respuesta:**
```json
{
  "status": "healthy",
  "service": "resume-backend-service"
}
```

---

### POST /api/v1/resume/
**Handler:** `resume_handler.go:ProcessResumeHandler()`

**Request:**
```
Content-Type: multipart/form-data

file: (binary)              # Requerido - .pdf, .txt, .docx
instructions: (string)      # Opcional - Instrucciones personalizadas
language: (string)          # Opcional - Default: "esp"
```

**Response (202 Accepted):**
```json
{
  "status": "accepted",
  "message": "Solicitud encolada para procesamiento."
}
```

**Errores:**
- 400: Archivo no enviado o formato no permitido
- 500: Error en conversión, presigned URL, o upload a S3

---

### POST /api/v1/resume/results
**Handler:** `aws_handler.go:ProcessResumeResultsHandler()`

**Request Body (JSON):**
```json
{
  "input_file": "s3://cv-processor-dev/inputs/2025-11-29/05-05-08/cv-clean.pdf",
  "output_file": "s3://cv-processor-dev/outputs/2025-11-29/05-05-08/cv-clean.json",
  "processing_time_ms": 11919,
  "status": "success",
  "structured_data": {
    "header": {
      "name": "Juan Pérez",
      "contact": {
        "email": "juan.perez@example.com",
        "phone": "+34 600 123 456"
      }
    },
    "professionalExperience": [...],
    "education": [...],
    "technicalSkills": { "skills": [...] },
    "certifications": [...],
    "projects": [...]
  }
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Datos procesados correctamente."
}
```

Ver `docs/resume-backend-api.yaml` para especificación completa OpenAPI 3.0.

---

## ⚙️ Configuración

### Variables de Entorno (.env)

```bash
# Servidor
SERVER_PORT=8081                    # Puerto del servidor (default: 8080)

# Archivos
MAX_FILE_SIZE_MB=10                 # Tamaño máximo en MB (default: 10)

# Servicios Externos
PRESIGNED_URL_SERVICE_ENDPOINT=https://api.cloudcentinel.com/signature/api/v1/presigned-url/upload
```

### Configuración de Docker

**Dockerfile:** Multi-stage build
- **Stage 1 (Builder):** golang:1.24-alpine + compilación estática
- **Stage 2 (Runtime):** alpine:latest + usuario no-root (appuser)
- **Healthcheck:** curl http://localhost:${SERVER_PORT}/api/v1/health/

**docker-compose.yml:**
- PostgreSQL 16-alpine (puerto 5432, usuario: resume_user, db: resume_db)
- Backend Service (puerto 8080, depende de PostgreSQL)
- Network: resume-network
- Volume: postgres_data

---

## 🔌 Integraciones Externas

### 1. Presigned URL Service
**Endpoint:** `https://api.cloudcentinel.com/signature/api/v1/presigned-url/upload`

**Request:**
```json
{
  "filename": "cv-clean.pdf",
  "content_type": "application/pdf",
  "metadata": {
    "language": "esp",
    "instructions": "Extraer últimos 5 años"
  }
}
```

**Response:**
```json
{
  "url": "https://cv-processor-dev.s3.amazonaws.com/...",
  "expires_in": "1 hour"
}
```

### 2. AWS S3
**Bucket:** cv-processor-dev (configurable)

**Estructura de rutas:**
- `/inputs/{date}/{time}/cv-clean.pdf` - CVs subidos
- `/outputs/{date}/{time}/cv-clean.json` - Resultados procesados

**Metadatos personalizados:**
- `x-amz-meta-language`
- `x-amz-meta-instructions`

### 3. AWS Lambda
**Trigger:** S3 event (PUT en /inputs/)

**Operación:** Extrae datos estructurados del CV

**Callback:** POST a `http://backend:8081/api/v1/resume/results`

---

## 📝 Convenciones de Código

### Estructura de Archivos
- **cmd/**: Puntos de entrada minimalistas (delegar a internal/)
- **internal/**: Código privado de la aplicación (no importable externamente)
- **pkg/**: Código reutilizable (puede ser importado por otros proyectos)

### Patrones de Diseño
- **Clean Architecture**: Separación en capas (handlers, services, repositories)
- **Dependency Injection**: Inyección de dependencias mediante constructores
- **DTO Pattern**: Data Transfer Objects para requests/responses

### Ejemplo: Inyección de Dependencias
```go
// Handler recibe Service
resumeHandler := &handlers.ResumeHandler{
    ResumeService: resumeService,
}

// Service recibe Client
resumeService := &services.ResumeService{
    PresignedURLClient: presignedURLClient,
}
```

### Manejo de Errores
```go
// Validación temprana con returns
if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
        "status":  "error",
        "message": err.Error(),
    })
}
```

### Naming Conventions
- **Handlers:** `ProcessResumeHandler`, `HandleHealthCheck`
- **Services:** `ProcessResume`, `UploadToS3`
- **DTOs:** `ResumeProcessorResponseDTO`, `AWSLambdaResponse`
- **Variables:** camelCase para privadas, PascalCase para públicas

---

## 🚀 Comandos Disponibles

### Makefile
```bash
make run        # Ejecutar servidor localmente (go run cmd/main.go)
make up         # Levantar servicios en Docker Compose
make down       # Detener servicios
make build      # Construir y levantar servicios
make logs       # Ver logs en tiempo real
make ps         # Ver estado de contenedores
make clean      # Detener y eliminar volúmenes
```

### Comandos Go
```bash
go run cmd/main.go              # Ejecutar aplicación
go build -o bin/server cmd/main.go  # Compilar binario
go mod tidy                     # Limpiar dependencias
go fmt ./...                    # Formatear código
```

### Docker
```bash
docker build -t resume-backend .
docker run -p 8080:8080 --env-file .env resume-backend
```

---

## 📊 Componentes Clave del Código

### 1. Bootstrap (internal/config/bootstrap.go)
Inicializa toda la aplicación:
```go
func Bootstrap() *Application {
    // 1. Carga .env
    godotenv.Load()

    // 2. Carga configuración
    config := Load()

    // 3. Inicializa Fiber
    app := fiber.New()

    // 4. Middlewares
    app.Use(logger.New())
    app.Use(recover.New())

    // 5. Rutas
    router.SetupRoutes(app, config)

    return &Application{App: app, Config: config}
}
```

### 2. Conversión de Archivos (pkg/converter/pdf_converter.go)
Soporta 3 formatos:

**PDF:** Lectura directa
```go
if ext == ".pdf" {
    return file.Read() // Ya es PDF
}
```

**TXT:** Conversión línea por línea
```go
pdf := gofpdf.New("P", "mm", "A4", "")
for _, line := range lines {
    pdf.Cell(0, 10, line)
}
```

**DOCX:** Extracción de texto y conversión
```go
docFile := docx.ReadDocxFile(tempFile)
text := docFile.Editable().GetContent()
// Convertir text a PDF con gofpdf
```

**NO SOPORTADO:** .doc (requeriría LibreOffice)

### 3. Service Layer (internal/services/resume_service.go)
Lógica de negocio completa:
```go
func (s *ResumeService) ProcessResume(...) error {
    // 1. Validar formato
    // 2. Convertir a PDF
    // 3. Obtener presigned URL
    // 4. Upload a S3 con metadatos
    return nil
}
```

### 4. DTOs (internal/dto/aws_dto.go)
Estructuras completas para datos procesados:
```go
type CVProcessedData struct {
    Certifications         []Certification    `json:"certifications"`
    Education              []Education        `json:"education"`
    Header                 Header             `json:"header"`
    ProfessionalExperience []Experience       `json:"professionalExperience"`
    Projects               []Project          `json:"projects"`
    TechnicalSkills        TechnicalSkills    `json:"technicalSkills"`
}
```

---

## ✅ Estado Actual del Proyecto

### Implementado
- ✅ Estructura Clean Architecture
- ✅ Health check endpoint
- ✅ Endpoint de upload de CVs
- ✅ Conversión de archivos a PDF (.txt, .docx → .pdf)
- ✅ Integración con Presigned URL Service
- ✅ Upload a S3 con metadatos personalizados
- ✅ Endpoint de callback para resultados de Lambda
- ✅ Parseo de datos estructurados del CV
- ✅ Logging detallado de resultados
- ✅ Dockerfile multi-stage
- ✅ Docker Compose con PostgreSQL
- ✅ Documentación OpenAPI 3.0
- ✅ Makefile con comandos útiles

### Pendiente (TODOs)
- ⏳ Guardar datos procesados en PostgreSQL
- ⏳ Implementar repositorios (internal/repository/)
- ⏳ Crear entidades de dominio (internal/domain/)
- ⏳ Middlewares de autenticación y autorización
- ⏳ Validadores reutilizables (pkg/validator/)
- ⏳ Tests unitarios y de integración
- ⏳ CI/CD pipeline
- ⏳ Métricas y observabilidad
- ⏳ Rate limiting
- ⏳ Gestión de usuarios y permisos

---

## 🐛 Problemas Conocidos y Soluciones

### Formato .doc no soportado
**Motivo:** Requiere LibreOffice o conversión externa
**Solución temporal:** Rechazar con error 400
**Solución futura:** Integrar con servicio de conversión o LibreOffice

### Datos procesados no se persisten
**Motivo:** Capa de repositorio no implementada
**Estado:** TODO en aws_handler.go:72
**Próximo paso:** Implementar ResumeRepository con PostgreSQL

---

## 🔍 Debugging

### Logs importantes
```bash
# Ver logs de Docker Compose
make logs

# Logs específicos del backend
docker-compose logs -f backend

# Logs de PostgreSQL
docker-compose logs -f postgres
```

### Variables de entorno en runtime
Los valores se cargan desde:
1. Archivo `.env` (si existe)
2. Variables de entorno del sistema
3. Defaults en el código

Verificar con:
```go
fmt.Printf("Config: %+v\n", config)
```

### Healthcheck
```bash
curl http://localhost:8080/api/v1/health/
```

---

## 📚 Recursos y Referencias

### Documentación
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Fiber Framework](https://docs.gofiber.io/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [OpenAPI Specification](https://swagger.io/specification/)

### Archivos Importantes
- `docs/resume-backend-api.yaml` - Especificación completa de la API
- `.env.example` - Template de configuración
- `Dockerfile` - Build multi-stage optimizado
- `docker-compose.yml` - Orquestación de servicios

---

## 🤝 Trabajando con Claude

### Contexto Clave
1. El proyecto usa **Clean Architecture** - mantener separación de capas
2. **internal/** es privado - código no reutilizable fuera del proyecto
3. **pkg/** es público - código compartible con otros proyectos
4. **DTOs** son críticos - mantener sincronizados con Lambda
5. **AWS Lambda callback** es asíncrono - no hay respuesta inmediata al cliente

### Al Añadir Nuevas Features
1. Leer código existente primero (especialmente bootstrap.go y router.go)
2. Seguir patrones existentes (inyección de dependencias)
3. Actualizar `docs/resume-backend-api.yaml` si cambian endpoints
4. Considerar impacto en integración AWS
5. Añadir logging apropiado

### Archivos a Revisar Frecuentemente
- `internal/config/bootstrap.go` - Inicialización de la app
- `internal/router/router.go` - Registro de rutas
- `internal/dto/aws_dto.go` - Estructuras de datos Lambda
- `docs/resume-backend-api.yaml` - Contrato de la API

---

## 📈 Métricas del Proyecto

**Total de archivos Go:** 13
**Total de líneas de código:** ~508
**Endpoints implementados:** 3
**Integraciones externas:** 3 (Presigned URL Service, S3, Lambda)
**Versión de Go:** 1.24.5
**Dependencias directas:** 4

---

## 🎯 Próximos Pasos Prioritarios

1. **Implementar persistencia en PostgreSQL**
   - Crear entidades en internal/domain/
   - Implementar ResumeRepository en internal/repository/
   - Migrar datos del callback a base de datos

2. **Tests unitarios**
   - Handlers (mocking de servicios)
   - Services (mocking de clientes)
   - Converters (diferentes formatos)

3. **Middleware de autenticación**
   - JWT o API Keys
   - Proteger endpoints de upload

4. **Observabilidad**
   - Structured logging
   - Métricas (Prometheus)
   - Tracing (OpenTelemetry)

---

**Última actualización:** 2025-11-29
**Branch principal:** main
**Commits recientes:**
- 6065bb0: Corregir parseo de datos de AWS Lambda con estructura wrapper
- 92d677e: Mejorar logging del endpoint de resultados procesados
- 5a49811: Corregir puerto en Docker
