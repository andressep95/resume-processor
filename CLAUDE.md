# CLAUDE.md - Contexto del Proyecto para Claude

Este documento proporciona contexto completo sobre el proyecto **Resume Backend Service** para facilitar la colaboración con Claude Code.

## 📋 Descripción del Proyecto

**Resume Backend Service** es un microservicio backend en Go que procesa currículums (CVs) de forma asíncrona mediante integración con AWS Lambda y S3. El servicio acepta archivos en múltiples formatos (.pdf, .txt, .docx), los convierte a PDF, los sube a S3 para procesamiento, y almacena los resultados estructurados en PostgreSQL.

### Propósito Principal
- Recibir CVs en diferentes formatos con autenticación JWT
- Convertir archivos a PDF estandarizado
- Subir a S3 con metadatos personalizados y Request ID para tracking
- Procesar información del CV mediante AWS Lambda
- Almacenar datos estructurados en PostgreSQL con vinculación a usuarios
- Proveer endpoints para listar y consultar CVs procesados

---

## 🏗️ Arquitectura del Proyecto

### Estructura de Directorios

```
resume-backend-service/
├── cmd/                          # Punto de entrada
│   └── main.go                   # Main minimalista (delega a bootstrap)
│
├── internal/                     # Código privado de la aplicación
│   ├── config/
│   │   ├── bootstrap.go          # Inicialización completa (BD + App)
│   │   ├── config.go             # Variables de entorno
│   │   └── database.go           # Conexión a PostgreSQL
│   │
│   ├── dto/                      # Data Transfer Objects
│   │   ├── resume_dto.go         # Response de procesamiento (con request_id)
│   │   ├── aws_dto.go            # Estructuras Lambda + CVProcessedData
│   │   ├── presigned_url_dto.go  # DTOs para URLs firmadas
│   │   ├── resume_list_dto.go    # DTOs para listado de CVs
│   │   └── resume_detail_dto.go  # DTO para detalle completo de CV
│   │
│   ├── handlers/                 # HTTP Handlers
│   │   ├── health_handler.go     # Health check
│   │   ├── resume_handler.go     # Upload de CVs (protegido con JWT)
│   │   ├── aws_handler.go        # Callback Lambda (guarda en BD)
│   │   └── resume_list_handler.go # Listado y detalle de CVs
│   │
│   ├── services/
│   │   └── resume_service.go     # Lógica de negocio (genera request_id)
│   │
│   ├── router/
│   │   └── router.go             # Definición de rutas + autenticación
│   │
│   ├── domain/                   # ✅ Entidades de dominio
│   │   ├── resume_request.go     # Solicitudes con estados (pending→completed)
│   │   └── processed_resume.go   # CVs procesados con datos estructurados
│   │
│   ├── middleware/               # ✅ Middlewares implementados
│   │   └── auth.go               # Validación JWT con JWKS
│   │
│   └── repository/               # ✅ Capa de persistencia
│       ├── resume_request_repository.go    # CRUD de solicitudes
│       ├── processed_resume_repository.go  # CRUD de CVs procesados
│       └── resume_list_repository.go       # Queries de listado
│
├── pkg/                          # Código reutilizable
│   ├── converter/
│   │   └── pdf_converter.go      # Conversión archivos (.txt, .docx → PDF)
│   │
│   └── client/
│       └── presigned_url_client.go  # Cliente HTTP para URLs firmadas
│
├── migrations/                   # Migraciones de base de datos
│   └── 001_create_resume_tables.sql  # Esquema inicial (auto-aplicado)
│
├── docs/
│   ├── resume-backend-api.yaml        # Especificación OpenAPI 3.0
│   ├── REQUEST_ID_FLOW.md             # Flujo completo de Request ID
│   ├── MIGRATIONS.md                  # Sistema de migraciones automáticas
│   └── IMPLEMENTATION_SUMMARY.md      # Resumen de implementación
│
├── Dockerfile                    # Multi-stage build + migraciones
├── docker-compose.yml            # PostgreSQL + Backend
├── docker-entrypoint.sh          # Script de inicialización con migraciones
├── Makefile                      # Comandos útiles
├── .env.example                  # Template de configuración
├── go.mod / go.sum              # Dependencias
└── README.md
```

**Total de código:** ~23 archivos Go | Clean Architecture completa

---

## 🔄 Flujos Principales

### 1. Flujo de Procesamiento de CV (Completo con Persistencia)

```
┌─────────────────────────────────────────────────────────────────┐
│ CLIENTE                                                         │
│ POST /api/v1/resume/                                            │
│ Authorization: Bearer <JWT_TOKEN>                               │
│ - file: resume.pdf                                              │
│ - language: esp                                                 │
│ - instructions: "Extraer últimos 5 años"                       │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ AUTH MIDDLEWARE (middleware/auth.go)                            │
│ - Valida JWT contra JWKS (con cache)                            │
│ - Extrae user_id (subject del token)                            │
│ - Guarda en c.Locals("user_subject")                            │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ RESUME HANDLER (resume_handler.go:ProcessResumeHandler)        │
│ - Obtiene user_id del context                                   │
│ - Valida archivo y extrae metadata                              │
│ - Delega a ResumeService                                        │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ RESUME SERVICE (resume_service.go:ProcessResume)               │
│ 1. Genera REQUEST_ID (UUID v4)                                  │
│ 2. Crea ResumeRequest entity (status: pending)                  │
│ 3. Guarda solicitud en BD (resume_requests table)               │
│ 4. Convierte archivo a PDF (si necesario)                       │
│ 5. Obtiene presigned URL (con request_id en metadata)           │
│ 6. Upload a S3 con metadatos (request-id, language, instr.)     │
│ 7. Actualiza solicitud en BD (status: uploaded)                 │
│ 8. Retorna request_id al cliente                                │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ PRESIGNED URL CLIENT (presigned_url_client.go)                 │
│ POST https://api.cloudcentinel.com/.../presigned-url/upload    │
│ Body: {                                                          │
│   filename: "cv-clean.pdf",                                      │
│   content_type: "application/pdf",                              │
│   metadata: {                                                    │
│     request_id: "550e8400-...",                                  │
│     language: "esp",                                             │
│     instructions: "..."                                          │
│   }                                                              │
│ }                                                                │
│ Retorna: { url: "https://s3.../", expires_in: "1 hour" }       │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ UPLOAD A S3                                                     │
│ PUT presigned_url                                               │
│ Headers:                                                         │
│   - Content-Type: application/pdf                               │
│   - x-amz-meta-request-id: "550e8400-..."                       │
│   - x-amz-meta-language: esp                                    │
│   - x-amz-meta-instructions: "..."                              │
│ Path: s3://.../inputs/2025-12-01/HH:MM/cv-clean.pdf           │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ RESPUESTA AL CLIENTE                                            │
│ 202 Accepted                                                    │
│ {                                                                │
│   "status": "accepted",                                          │
│   "message": "Solicitud encolada para procesamiento.",          │
│   "request_id": "550e8400-e29b-41d4-a716-446655440000"          │
│ }                                                                │
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
│ 1. Lee metadata del objeto S3 (request-id, language)            │
│ 2. Procesa PDF y extrae:                                        │
│    - Header (nombre, email, teléfono)                           │
│    - Educación (institución, grado, fecha)                      │
│    - Experiencia laboral (empresa, cargo, período)             │
│    - Habilidades técnicas                                       │
│    - Proyectos                                                   │
│    - Certificaciones                                             │
│ 3. Sube resultado: s3://.../outputs/.../cv-clean.json          │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ CALLBACK A BACKEND                                              │
│ POST /api/v1/resume/results                                     │
│ Body: {                                                          │
│   request_id: "550e8400-...",  ⭐ CLAVE                         │
│   input_file: "s3://...",                                        │
│   output_file: "s3://...",                                       │
│   status: "success",                                             │
│   processing_time_ms: 11919,                                     │
│   structured_data: { ... }                                       │
│ }                                                                │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│ AWS HANDLER (aws_handler.go:ProcessResumeResultsHandler)       │
│ 1. Parsea request_id del callback                               │
│ 2. Busca solicitud original en BD por request_id                │
│ 3. Obtiene user_id de la solicitud (vinculación)                │
│ 4. Crea ProcessedResume entity con datos estructurados          │
│ 5. Guarda en processed_resumes table                            │
│ 6. Actualiza resume_requests (status: completed)                │
│ 7. Logging detallado de resultados                              │
│                                                                  │
│ Respuesta: 200 OK { status: "success", ... }                   │
└─────────────────────────────────────────────────────────────────┘
```

### 2. Flujo de Listado de CVs

```
GET /api/v1/resume/my-resumes
Authorization: Bearer <JWT_TOKEN>
     ↓
Auth Middleware → Extrae user_id
     ↓
ResumeListHandler.GetMyResumes()
     ↓
ResumeRequestRepository.GetUserResumes(user_id)
     ↓
Query SQL con LEFT JOIN:
  - resume_requests (solicitudes)
  - processed_resumes (datos procesados)
     ↓
Retorna: {
  total: 2,
  resumes: [
    {
      request_id: "...",
      original_filename: "cv.pdf",
      status: "completed",
      created_at: "...",
      full_name: "Juan Pérez",  // Si completado
      email: "juan@example.com" // Si completado
    }
  ]
}
```

### 3. Flujo de Detalle de CV

```
GET /api/v1/resume/:request_id
Authorization: Bearer <JWT_TOKEN>
     ↓
Auth Middleware → Extrae user_id
     ↓
ResumeListHandler.GetResumeDetail()
     ↓
1. Busca resume_request por request_id
2. Verifica que request.user_id == user_id (autorización)
3. Si status == "completed", busca processed_resume
4. Deserializa structured_data (JSONB → CVProcessedData)
     ↓
Retorna: {
  request_id, filename, status, timestamps,
  structured_data: {
    header: {...},
    education: [...],
    professionalExperience: [...],
    ...
  }
}
```

### 4. Health Check
```
GET /api/v1/health/
→ HealthHandler.HandleHealthCheck()
→ { "status": "healthy", "service": "resume-backend-service" }
```

---

## 🛠️ Tecnologías y Dependencias

### Dependencias Principales (go.mod)

```go
github.com/gofiber/fiber/v2 v2.52.10       // Framework HTTP rápido
github.com/google/uuid v1.6.0              // Generación de UUIDs (request_id)
github.com/joho/godotenv v1.5.1            // Carga .env
github.com/jung-kurt/gofpdf v1.16.2        // Generación de PDFs
github.com/lestrrat-go/jwx/v2 v2.1.6       // Validación JWT con JWKS
github.com/lib/pq v1.10.9                  // Driver PostgreSQL
github.com/nguyenthenguyen/docx v0.0.0     // Lectura de archivos DOCX
```

### Stack Tecnológico

| Categoría | Tecnología | Versión | Uso |
|-----------|-----------|---------|-----|
| **Lenguaje** | Go | 1.24.5 | Backend |
| **Framework HTTP** | Fiber | v2.52.10 | Servidor REST API |
| **Autenticación** | JWX | v2.1.6 | Validación JWT con JWKS |
| **Base de Datos** | PostgreSQL | 16 | Persistencia (resume_requests, processed_resumes) |
| **Conversión PDF** | gofpdf | v1.16.2 | TXT/DOCX → PDF |
| **Lectura DOCX** | docx | v0.0.0 | Extracción de texto |
| **UUID** | google/uuid | v1.6.0 | Generación de request_id |
| **Configuración** | godotenv | v1.5.1 | Variables de entorno |
| **Contenedores** | Docker | Latest | Empaquetamiento |
| **Orquestación** | Docker Compose | - | Entorno local |
| **Almacenamiento** | AWS S3 | - | Archivos input/output |
| **Procesamiento** | AWS Lambda | - | Extracción de datos |

---

## 🌐 Endpoints de la API

### GET /api/v1/health/
**Handler:** `health_handler.go:HandleHealthCheck()`
**Autenticación:** No requerida

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
**Autenticación:** JWT (Bearer token)
**Middleware:** `ValidateJWT()`

**Request:**
```http
Authorization: Bearer <JWT_TOKEN>
Content-Type: multipart/form-data

file: (binary)              # Requerido - .pdf, .txt, .docx
instructions: (string)      # Opcional - Instrucciones personalizadas
language: (string)          # Opcional - Default: "esp"
```

**Response (202 Accepted):**
```json
{
  "status": "accepted",
  "message": "Solicitud encolada para procesamiento.",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Errores:**
- 400: Archivo no enviado o formato no permitido
- 401: Token JWT inválido o ausente
- 500: Error en conversión, presigned URL, o upload a S3

---

### GET /api/v1/resume/my-resumes
**Handler:** `resume_list_handler.go:GetMyResumes()`
**Autenticación:** JWT (Bearer token)
**Middleware:** `ValidateJWT()`

**Response (200 OK):**
```json
{
  "total": 2,
  "resumes": [
    {
      "request_id": "550e8400-e29b-41d4-a716-446655440000",
      "original_filename": "mi-cv.pdf",
      "status": "completed",
      "created_at": "2025-12-01T10:00:00Z",
      "completed_at": "2025-12-01T10:00:20Z",
      "full_name": "Juan Pérez",
      "email": "juan.perez@example.com"
    },
    {
      "request_id": "660e8400-e29b-41d4-a716-446655440001",
      "original_filename": "resume.docx",
      "status": "uploaded",
      "created_at": "2025-12-01T11:00:00Z"
    }
  ]
}
```

**Estados posibles:** `pending`, `uploaded`, `processing`, `completed`, `failed`

---

### GET /api/v1/resume/:request_id
**Handler:** `resume_list_handler.go:GetResumeDetail()`
**Autenticación:** JWT (Bearer token)
**Middleware:** `ValidateJWT()`

**Response (200 OK):**
```json
{
  "request_id": "550e8400-...",
  "original_filename": "mi-cv.pdf",
  "original_file_type": ".pdf",
  "file_size_bytes": 524288,
  "language": "esp",
  "instructions": "Extraer últimos 5 años",
  "status": "completed",
  "s3_input_url": "s3://bucket/inputs/.../cv-clean.pdf",
  "s3_output_url": "s3://bucket/outputs/.../cv-clean.json",
  "processing_time_ms": 11919,
  "created_at": "2025-12-01T10:00:00Z",
  "uploaded_at": "2025-12-01T10:00:05Z",
  "completed_at": "2025-12-01T10:00:20Z",
  "structured_data": {
    "header": {
      "name": "Juan Pérez",
      "contact": {
        "email": "juan.perez@example.com",
        "phone": "+34 600 123 456"
      }
    },
    "education": [...],
    "professionalExperience": [...],
    "technicalSkills": {"skills": [...]},
    "certifications": [...],
    "projects": [...]
  }
}
```

**Errores:**
- 400: Request ID inválido
- 401: No autenticado
- 403: El CV no pertenece al usuario
- 404: CV no encontrado

---

### POST /api/v1/resume/results
**Handler:** `aws_handler.go:ProcessResumeResultsHandler()`
**Autenticación:** No requerida (callback de AWS Lambda)

**Request Body (JSON):**
```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "input_file": "s3://cv-processor-dev/inputs/2025-12-01/10-00/cv-clean.pdf",
  "output_file": "s3://cv-processor-dev/outputs/2025-12-01/10-00/cv-clean.json",
  "processing_time_ms": 11919,
  "status": "success",
  "structured_data": {
    "header": {...},
    "professionalExperience": [...],
    "education": [...],
    "technicalSkills": {...},
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
SERVER_PORT=8080                    # Puerto del servidor (default: 8080)

# Archivos
MAX_FILE_SIZE_MB=10                 # Tamaño máximo en MB (default: 10)

# Servicios Externos
PRESIGNED_URL_SERVICE_ENDPOINT=https://api.cloudcentinel.com/signature/api/v1/presigned-url/upload

# Autenticación JWT
AUTH_JWKS_URL=https://auth.cloudcentinel.com/.well-known/jwks.json

# CORS
CORS_ALLOWED_ORIGINS=*              # Orígenes permitidos (separados por coma)

# Base de Datos PostgreSQL
DB_HOST=localhost                   # Host de PostgreSQL
DB_PORT=5432                        # Puerto (default: 5432)
DB_USER=resume_user                 # Usuario de BD
DB_PASSWORD=resume_password         # Contraseña
DB_NAME=resume_db                   # Nombre de la BD
DB_SSLMODE=disable                  # SSL mode (disable, require, verify-full)
```

### Configuración de Docker

**Dockerfile:** Multi-stage build con migraciones automáticas
- **Stage 1 (Builder):** golang:1.24-alpine + compilación estática
- **Stage 2 (Runtime):** alpine:latest + usuario no-root (appuser)
- **Healthcheck:** curl http://localhost:${SERVER_PORT}/api/v1/health/
- **Migraciones:** Script docker-entrypoint.sh ejecuta migraciones al inicio

**docker-compose.yml:**
- PostgreSQL 16-alpine (puerto 5432, usuario: resume_user, db: resume_db)
- Backend Service (puerto 8080, depende de PostgreSQL)
- Network: resume-network
- Volume: postgres_data (persistencia)
- Migraciones automáticas al iniciar

---

## 🗄️ Base de Datos

### Sistema de Migraciones Automáticas

El proyecto incluye un sistema de migraciones que se ejecuta automáticamente al iniciar el contenedor Docker:

**Flujo:**
1. Container inicia
2. `docker-entrypoint.sh` espera a que PostgreSQL esté listo
3. Ejecuta migraciones pendientes desde `/migrations/`
4. Registra migraciones aplicadas en `schema_migrations`
5. Inicia la aplicación

**Ver:** `docs/MIGRATIONS.md` para detalles completos.

### Modelo de Datos

#### Tabla: `resume_requests`
Tracking de solicitudes de procesamiento.

```sql
CREATE TABLE resume_requests (
    request_id UUID PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    original_filename VARCHAR(500),
    original_file_type VARCHAR(10),
    file_size_bytes BIGINT,
    language VARCHAR(10) DEFAULT 'es',
    instructions TEXT,
    s3_input_url TEXT,
    s3_output_url TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    processing_time_ms BIGINT,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    uploaded_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_resume_requests_user_id ON resume_requests(user_id);
CREATE INDEX idx_resume_requests_status ON resume_requests(status);
```

**Estados:** `pending` → `uploaded` → `completed` / `failed`

#### Tabla: `processed_resumes`
CVs procesados con datos estructurados.

```sql
CREATE TABLE processed_resumes (
    id BIGSERIAL PRIMARY KEY,
    request_id UUID UNIQUE REFERENCES resume_requests(request_id),
    user_id VARCHAR(255) NOT NULL,
    structured_data JSONB NOT NULL,
    cv_name VARCHAR(500),
    cv_email VARCHAR(255),
    cv_phone VARCHAR(100),
    education_count INT DEFAULT 0,
    experience_count INT DEFAULT 0,
    certifications_count INT DEFAULT 0,
    projects_count INT DEFAULT 0,
    skills_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_processed_resumes_user_id ON processed_resumes(user_id);
CREATE INDEX idx_processed_resumes_cv_email ON processed_resumes(cv_email);
```

**Relación:** 1 request = 1 processed_resume (1:1 via request_id)

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
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
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

**⚠️ Importante:** El servicio debe incluir el `request_id` en la firma de la presigned URL para que S3 acepte el upload con ese metadata.

### 2. AWS S3
**Bucket:** cv-processor-dev (configurable)

**Estructura de rutas:**
- `/inputs/{date}/{time}/cv-clean.pdf` - CVs subidos
- `/outputs/{date}/{time}/cv-clean.json` - Resultados procesados

**Metadatos personalizados:**
- `x-amz-meta-request-id` ⭐ CLAVE para tracking
- `x-amz-meta-language`
- `x-amz-meta-instructions`

### 3. AWS Lambda
**Trigger:** S3 event (PUT en /inputs/)

**Operación:**
1. Lee metadata del objeto S3 (especialmente `request-id`)
2. Extrae datos estructurados del CV
3. Sube resultado JSON a `/outputs/`
4. Callback al backend con `request_id`

**Callback:** POST a `http://backend:8080/api/v1/resume/results`

**⚠️ Importante:** Lambda DEBE extraer y devolver el `request-id` del metadata de S3 para vincular el resultado con la solicitud original.

### 4. Servicio de Autenticación
**JWKS Endpoint:** `https://auth.cloudcentinel.com/.well-known/jwks.json`

- El middleware valida tokens JWT contra este endpoint
- Cache automático de claves con refresh cada 10 minutos
- Extrae `subject` del token como `user_id`
- Soporta tokens con o sin `kid` (key ID)

---

## 📝 Convenciones de Código

### Estructura de Archivos
- **cmd/**: Puntos de entrada minimalistas (delegar a internal/)
- **internal/**: Código privado de la aplicación (no importable externamente)
- **pkg/**: Código reutilizable (puede ser importado por otros proyectos)
- **migrations/**: Archivos SQL de migraciones (numerados secuencialmente)

### Patrones de Diseño
- **Clean Architecture**: Separación en capas (handlers, services, repositories, domain)
- **Dependency Injection**: Inyección de dependencias mediante constructores
- **DTO Pattern**: Data Transfer Objects para requests/responses
- **Repository Pattern**: Abstracción de acceso a datos
- **Domain Entities**: Lógica de negocio en entidades de dominio

### Ejemplo: Inyección de Dependencias

```go
// Router inicializa toda la cadena de dependencias
func SetupRoutes(app *fiber.App, db *sql.DB, presignedURLEndpoint string, authMiddleware *middleware.AuthMiddleware) {
    // Repositorios
    resumeRequestRepo := repository.NewResumeRequestRepository(db)
    processedResumeRepo := repository.NewProcessedResumeRepository(db)

    // Clientes
    presignedURLClient := client.NewPresignedURLClient(presignedURLEndpoint)

    // Servicios
    resumeService := services.NewResumeService(presignedURLClient, resumeRequestRepo)

    // Handlers
    resumeHandler := handlers.NewResumeHandler(resumeService)
    awsHandler := handlers.NewAWSHandler(resumeRequestRepo, processedResumeRepo)

    // Rutas protegidas
    resume.Post("/", authMiddleware.ValidateJWT(), resumeHandler.ProcessResumeHandler)
}
```

### Manejo de Errores
```go
// Validación temprana con returns
if err != nil {
    log.Printf("❌ Error: %v", err)
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
        "status":  "error",
        "message": err.Error(),
    })
}
```

### Naming Conventions
- **Handlers:** `ProcessResumeHandler`, `HandleHealthCheck`, `GetMyResumes`
- **Services:** `ProcessResume`, `UploadToS3`
- **Repositories:** `FindByRequestID`, `MarkAsCompleted`
- **DTOs:** `ResumeProcessorResponseDTO`, `AWSLambdaResponse`, `ResumeListItemDTO`
- **Domain:** `ResumeRequest`, `ProcessedResume`
- **Variables:** camelCase para privadas, PascalCase para públicas

---

## 🚀 Comandos Disponibles

### Makefile
```bash
make run        # Ejecutar servidor localmente (go run cmd/main.go)
make up         # Levantar servicios en Docker Compose (con migraciones)
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
    cfg := Load()

    // 3. Inicializa base de datos
    db, err := InitDatabase(cfg)

    // 4. Inicializa Fiber con CORS
    app := fiber.New()
    app.Use(cors.New(...))
    app.Use(logger.New())
    app.Use(recover.New())

    // 5. Inicializa middleware de autenticación
    authMiddleware := middleware.NewAuthMiddleware(cfg.AuthJWKSURL)

    // 6. Registra rutas
    router.SetupRoutes(app, db, cfg.PresignedURLServiceEndpoint, authMiddleware)

    return &Application{App: app, Config: cfg, DB: db}
}
```

### 2. Middleware de Autenticación (internal/middleware/auth.go)
Validación JWT con JWKS:
```go
func (a *AuthMiddleware) ValidateJWT() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 1. Extraer token del header Authorization
        // 2. Obtener JWKS del cache
        // 3. Validar y parsear token
        // 4. Extraer user_id (subject)
        // 5. Guardar en c.Locals("user_subject")
        return c.Next()
    }
}
```

**Características:**
- Cache de JWKS con refresh automático cada 10 minutos
- Soporta tokens con y sin `kid` (key ID)
- Manejo de múltiples claves en el keyset
- Logging detallado de validación

### 3. Domain Entities (internal/domain/)

**ResumeRequest:**
```go
type ResumeRequest struct {
    RequestID        uuid.UUID
    UserID           string
    OriginalFilename string
    Status           ResumeRequestStatus  // pending, uploaded, completed, failed
    // ... timestamps, URLs, metadata
}

// Métodos de cambio de estado
func (r *ResumeRequest) MarkAsUploaded(s3InputURL string)
func (r *ResumeRequest) MarkAsCompleted(s3OutputURL string, processingTimeMs int64)
func (r *ResumeRequest) MarkAsFailed(errorMsg string)
```

**ProcessedResume:**
```go
type ProcessedResume struct {
    ID             int64
    RequestID      uuid.UUID
    UserID         string
    StructuredData json.RawMessage  // JSONB con datos del CV
    CVName         string
    CVEmail        string
    // ... campos de conteo (education_count, experience_count, etc.)
}

func NewProcessedResume(requestID uuid.UUID, userID string, cvData *dto.CVProcessedData) (*ProcessedResume, error)
func (p *ProcessedResume) GetStructuredData() (*dto.CVProcessedData, error)
```

### 4. Conversión de Archivos (pkg/converter/pdf_converter.go)
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

### 5. Service Layer (internal/services/resume_service.go)
Lógica de negocio con Request ID:
```go
func (s *ResumeService) ProcessResume(userID, filename string, fileData []byte, language, instructions string) (string, error) {
    // 1. Crear ResumeRequest con UUID
    resumeRequest := domain.NewResumeRequest(userID, filename, ...)

    // 2. Guardar en BD (status: pending)
    s.resumeRequestRepo.Create(resumeRequest)

    // 3. Convertir a PDF
    pdfData := converter.ConvertToPDF(...)

    // 4. Obtener presigned URL (con request_id)
    presignedURL := s.presignedURLClient.GetUploadURL(..., resumeRequest.RequestID)

    // 5. Upload a S3 con metadatos
    s.uploadToS3(presignedURL, pdfData, resumeRequest.RequestID, language, instructions)

    // 6. Actualizar BD (status: uploaded)
    s.resumeRequestRepo.MarkAsUploaded(resumeRequest.RequestID, s3InputURL)

    // 7. Retornar request_id
    return resumeRequest.RequestID.String(), nil
}
```

### 6. Repository Pattern (internal/repository/)

Métodos principales de **ResumeRequestRepository:**
- `Create(request *domain.ResumeRequest) error`
- `FindByRequestID(requestID uuid.UUID) (*domain.ResumeRequest, error)`
- `FindByUserID(userID string) ([]*domain.ResumeRequest, error)`
- `MarkAsUploaded(requestID uuid.UUID, s3InputURL string) error`
- `MarkAsCompleted(requestID uuid.UUID, s3OutputURL string, processingTimeMs int64) error`
- `MarkAsFailed(requestID uuid.UUID, errorMessage string) error`
- `GetUserResumes(userID string) ([]ResumeListItem, error)` - Join con processed_resumes

Métodos principales de **ProcessedResumeRepository:**
- `Create(resume *domain.ProcessedResume) error`
- `FindByRequestID(requestID uuid.UUID) (*domain.ProcessedResume, error)`
- `FindByUserID(userID string) ([]*domain.ProcessedResume, error)`
- `Delete(requestID uuid.UUID) error`

---

## ✅ Estado Actual del Proyecto

### Implementado
- ✅ Estructura Clean Architecture completa
- ✅ Health check endpoint
- ✅ Autenticación JWT con middleware y JWKS
- ✅ Sistema de Request ID para tracking
- ✅ Endpoint de upload de CVs (protegido con JWT)
- ✅ Conversión de archivos a PDF (.txt, .docx → .pdf)
- ✅ Integración con Presigned URL Service (con request_id)
- ✅ Upload a S3 con metadatos personalizados (request-id, language, instructions)
- ✅ Endpoint de callback para resultados de Lambda
- ✅ Persistencia en PostgreSQL (resume_requests, processed_resumes)
- ✅ Sistema de migraciones automáticas
- ✅ Repositorios completos (CRUD de solicitudes y CVs)
- ✅ Entidades de dominio (ResumeRequest, ProcessedResume)
- ✅ Endpoint de listado de CVs del usuario (GET /api/v1/resume/my-resumes)
- ✅ Endpoint de detalle completo de CV (GET /api/v1/resume/:request_id)
- ✅ Estados de solicitud (pending, uploaded, completed, failed)
- ✅ Logging detallado de resultados
- ✅ Dockerfile multi-stage con migraciones
- ✅ Docker Compose con PostgreSQL
- ✅ Configuración de CORS
- ✅ Documentación OpenAPI 3.0
- ✅ Makefile con comandos útiles
- ✅ Documentación completa (REQUEST_ID_FLOW.md, MIGRATIONS.md, IMPLEMENTATION_SUMMARY.md)

### Pendiente (TODOs)
- ⏳ Endpoint de búsqueda de CVs (por habilidades, experiencia, etc.) usando JSONB queries
- ⏳ Endpoint de estadísticas del usuario
- ⏳ Validadores reutilizables (pkg/validator/)
- ⏳ Tests unitarios y de integración
- ⏳ CI/CD pipeline
- ⏳ Métricas y observabilidad (Prometheus, Grafana)
- ⏳ Rate limiting
- ⏳ Soporte para .doc (LibreOffice integration)
- ⏳ Notificaciones push cuando el CV está procesado
- ⏳ Webhooks configurables para eventos

---

## 🐛 Problemas Conocidos y Soluciones

### Formato .doc no soportado
**Motivo:** Requiere LibreOffice o conversión externa
**Solución temporal:** Rechazar con error 400
**Solución futura:** Integrar con LibreOffice via Docker o servicio externo

### Presigned URL Service debe incluir request_id en la firma
**Motivo:** Si el request-id no está en la firma de la presigned URL, S3 rechazará el upload con `SignatureDoesNotMatch`
**Solución:** El servicio de presigned URLs debe incluir todos los metadatos (incluyendo request-id) al generar la firma
**Ver:** `docs/IMPLEMENTATION_SUMMARY.md` para detalles

### AWS Lambda debe extraer y devolver request_id
**Motivo:** Sin el request_id, el backend no puede vincular el resultado con la solicitud original
**Solución:** Lambda debe leer `s3Object.Metadata['request-id']` y devolverlo en el callback
**Ver:** `docs/REQUEST_ID_FLOW.md` para implementación

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

# Ver migraciones aplicadas
docker-compose exec postgres psql -U resume_user -d resume_db -c "SELECT * FROM schema_migrations;"
```

### Queries útiles de base de datos
```sql
-- Ver solicitudes de un usuario
SELECT * FROM resume_requests WHERE user_id = 'user-123' ORDER BY created_at DESC;

-- Ver estado de una solicitud
SELECT rr.*, pr.cv_name, pr.cv_email
FROM resume_requests rr
LEFT JOIN processed_resumes pr ON rr.request_id = pr.request_id
WHERE rr.request_id = '550e8400-...';

-- Buscar CVs por habilidad (usando JSONB)
SELECT * FROM processed_resumes
WHERE structured_data @> '{"technicalSkills": {"skills": ["Go"]}}'::jsonb;

-- Estadísticas de procesamiento
SELECT status, COUNT(*) as count, AVG(processing_time_ms) as avg_time_ms
FROM resume_requests
GROUP BY status;
```

### Healthcheck
```bash
curl http://localhost:8080/api/v1/health/
```

### Test de endpoints protegidos
```bash
# Obtener token JWT del servicio de autenticación
TOKEN="your-jwt-token"

# Upload CV
curl -X POST http://localhost:8080/api/v1/resume/ \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@cv.pdf" \
  -F "language=esp"

# Listar CVs
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/resume/my-resumes

# Detalle de CV
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/resume/550e8400-...
```

---

## 📚 Recursos y Referencias

### Documentación del Proyecto
- `docs/resume-backend-api.yaml` - Especificación completa de la API (OpenAPI 3.0)
- `docs/REQUEST_ID_FLOW.md` - Flujo completo del sistema de Request ID
- `docs/MIGRATIONS.md` - Sistema de migraciones automáticas
- `docs/IMPLEMENTATION_SUMMARY.md` - Resumen de implementación del sistema
- `.env.example` - Template de configuración
- `Dockerfile` - Build multi-stage optimizado
- `docker-compose.yml` - Orquestación de servicios
- `docker-entrypoint.sh` - Script de inicialización con migraciones

### Recursos Externos
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Fiber Framework](https://docs.gofiber.io/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [OpenAPI Specification](https://swagger.io/specification/)
- [PostgreSQL JSONB](https://www.postgresql.org/docs/current/datatype-json.html)
- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)

---

## 🤝 Trabajando con Claude

### Contexto Clave
1. El proyecto usa **Clean Architecture** - mantener separación estricta de capas
2. **Request ID** es el mecanismo de tracking - todos los flujos giran alrededor de él
3. **internal/** es privado - código no reutilizable fuera del proyecto
4. **pkg/** es público - código compartible con otros proyectos
5. **DTOs** son críticos - mantener sincronizados con Lambda y frontend
6. **JWT** es obligatorio para endpoints de usuario - siempre validar autenticación
7. **PostgreSQL** es la fuente de verdad - toda persistencia debe pasar por repositorios
8. **AWS Lambda callback** es asíncrono - el cliente no recibe datos procesados inmediatamente

### Al Añadir Nuevas Features
1. **Leer código existente** primero (especialmente bootstrap.go, router.go, y domain/)
2. **Seguir patrones existentes:**
   - Domain entities para lógica de negocio
   - Repositories para acceso a datos
   - Services para orquestación
   - Handlers para HTTP
   - DTOs para transferencia
3. **Actualizar documentación:**
   - `docs/resume-backend-api.yaml` si cambian endpoints
   - `CLAUDE.md` si cambia arquitectura
   - `README.md` si afecta uso del proyecto
4. **Considerar impactos:**
   - Integración AWS (Lambda, S3)
   - Presigned URL Service
   - Frontend (contratos de API)
   - Migraciones de BD si cambia esquema
5. **Añadir logging apropiado** (usar emojis: ✅, ❌, ⏳, ℹ️ para claridad)
6. **Proteger con JWT** si el endpoint es para usuarios autenticados

### Archivos a Revisar Frecuentemente
- `internal/config/bootstrap.go` - Inicialización completa de la app
- `internal/router/router.go` - Registro de rutas y middlewares
- `internal/domain/` - Entidades de negocio y lógica de estados
- `internal/repository/` - Acceso a datos y queries SQL
- `internal/dto/aws_dto.go` - Estructuras de datos Lambda
- `docs/resume-backend-api.yaml` - Contrato de la API
- `docs/REQUEST_ID_FLOW.md` - Flujo completo del sistema
- `migrations/` - Esquema de base de datos

### Patrones de Implementación

**Agregar nuevo endpoint protegido:**
```go
// 1. Crear DTO en internal/dto/
type MyRequestDTO struct { ... }
type MyResponseDTO struct { ... }

// 2. Agregar método al repositorio si necesita BD
func (r *MyRepository) MyMethod() error { ... }

// 3. Agregar método al servicio
func (s *MyService) DoSomething() error { ... }

// 4. Crear handler
func (h *MyHandler) HandleMyRequest(c *fiber.Ctx) error {
    userID := c.Locals("user_subject").(string)
    // ... lógica
}

// 5. Registrar ruta en router.go
myGroup.Get("/my-endpoint", authMiddleware.ValidateJWT(), myHandler.HandleMyRequest)

// 6. Actualizar OpenAPI spec
```

**Agregar nueva migración:**
```bash
# 1. Crear archivo numerado
touch migrations/002_add_my_table.sql

# 2. Escribir SQL idempotente
CREATE TABLE IF NOT EXISTS my_table (...);

# 3. Rebuild y restart
docker-compose up -d --build

# 4. Verificar en logs
docker-compose logs backend | grep migration
```

---

## 📈 Métricas del Proyecto

**Total de archivos Go:** 23
**Total de líneas de código:** ~2,500
**Endpoints implementados:** 5
- GET /api/v1/health/ (público)
- POST /api/v1/resume/ (protegido JWT)
- GET /api/v1/resume/my-resumes (protegido JWT)
- GET /api/v1/resume/:request_id (protegido JWT)
- POST /api/v1/resume/results (callback AWS)

**Integraciones externas:** 4
- Presigned URL Service
- AWS S3
- AWS Lambda
- Servicio de Autenticación (JWKS)

**Versión de Go:** 1.24.5
**Dependencias directas:** 7
**Tablas de BD:** 2 (resume_requests, processed_resumes) + 1 tracking (schema_migrations)

---

## 🎯 Próximos Pasos Prioritarios

### 1. Tests y Calidad
- Tests unitarios de handlers (mocking de servicios)
- Tests de servicios (mocking de clientes y repositorios)
- Tests de repositorios (base de datos de pruebas)
- Tests de converters (diferentes formatos)
- Integration tests completos (flujo end-to-end)
- Code coverage > 80%

### 2. Observabilidad
- Structured logging con niveles (debug, info, warn, error)
- Métricas con Prometheus (requests, latency, errors)
- Dashboards en Grafana
- Tracing distribuido con OpenTelemetry
- Health checks avanzados (BD, S3, servicios externos)

### 3. Features de Producto
- Endpoint de búsqueda de CVs por criterios
- Estadísticas del usuario (total procesados, tiempo promedio, etc.)
- Notificaciones push cuando el CV está listo
- Webhooks configurables para eventos
- Rate limiting por usuario
- Paginación en endpoints de listado

### 4. DevOps
- CI/CD pipeline (GitHub Actions / GitLab CI)
- Automated deployments
- Environment management (dev, staging, prod)
- Secrets management (HashiCorp Vault / AWS Secrets Manager)
- Monitoring y alerting
- Disaster recovery plan

### 5. Mejoras Técnicas
- Soporte para .doc (LibreOffice integration)
- Cache de resultados (Redis)
- Message queue para procesamiento (RabbitMQ / SQS)
- Retry mechanism para fallos de Lambda
- Backup automático de base de datos
- Multi-region support

---

**Última actualización:** 2025-12-01
**Branch principal:** main
**Commits recientes:**
- f1d892e: fix: corregir nombres de columnas en query de listado
- 206b13a: docs: actualizar OpenAPI con nuevos endpoints
- 9405584: feat: agregar endpoints de listado y detalle de CVs
- a5ac649: fix: manejar campos NULL en queries de BD
