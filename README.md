# Resume Backend Service

[![Go Version](https://img.shields.io/badge/Go-1.24.5-00ADD8?logo=go)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-v2.52.10-00ACD7?logo=go)](https://gofiber.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-316192?logo=postgresql)](https://www.postgresql.org/)

Microservicio backend en Go para procesamiento asíncrono de currículums (CVs) mediante integración con AWS Lambda y S3. Acepta archivos en múltiples formatos (.pdf, .txt, .docx), los convierte a PDF estandarizado, y procesa la información mediante inteligencia artificial, almacenando los resultados estructurados en PostgreSQL.

## Tabla de Contenidos

- [Características](#-características)
- [Arquitectura](#-arquitectura)
- [Tecnologías](#-tecnologías)
- [Inicio Rápido](#-inicio-rápido)
- [API Endpoints](#-api-endpoints)
- [Configuración](#-configuración)
- [Desarrollo](#-desarrollo)
- [Base de Datos](#-base-de-datos)
- [Despliegue](#-despliegue)
- [Documentación](#-documentación)

---

## Características

- **Autenticación JWT:** Seguridad con validación de tokens via JWKS
- **Request ID Tracking:** Sistema completo de tracking de solicitudes
- **Procesamiento Asíncrono:** Upload y procesamiento no bloqueante de CVs
- **Conversión Multi-formato:** Soporte para .pdf, .txt, .docx (conversión automática a PDF)
- **Integración AWS:** S3 para almacenamiento y Lambda para procesamiento con IA
- **Persistencia Completa:** PostgreSQL con migraciones automáticas
- **Extracción Estructurada:** Datos organizados (contacto, experiencia, educación, skills, etc.)
- **API RESTful:** Endpoints bien documentados con OpenAPI 3.0
- **Estados de Solicitud:** Tracking completo (pending → uploaded → completed)
- **Listado de CVs:** Endpoints para consultar CVs procesados del usuario
- **Docker Ready:** Containerización completa con Docker Compose
- **Clean Architecture:** Código organizado, mantenible y escalable
- **Health Checks:** Monitoreo de disponibilidad del servicio

---

## Arquitectura

### Flujo de Procesamiento

```
Cliente (JWT) → Backend → Presigned URL Service → AWS S3 → AWS Lambda → Backend (Callback) → PostgreSQL
     ↓             ↓                                  ↓           ↓             ↓               ↓
 Autenticación  Genera                             Trigger    Procesa       Vincula        Almacena
               Request ID                          Lambda       CV        con Request ID   Resultados
              Guarda en BD                                                                 Persistente
```

### Componentes Principales

1. **Auth Middleware:** Validación JWT con JWKS y cache
2. **Request ID System:** UUID para tracking de solicitudes
3. **File Converter:** Conversión de archivos a PDF
4. **Resume Service:** Lógica de negocio y orquestación
5. **Repositories:** Acceso a datos con PostgreSQL
6. **Domain Entities:** ResumeRequest y ProcessedResume con estados
7. **AWS Integration:** S3 para storage, Lambda para procesamiento

### Estructura del Proyecto

```
resume-backend-service/
├── cmd/main.go                   # Punto de entrada
├── internal/                     # Código privado de la aplicación
│   ├── config/                   # Configuración y bootstrap
│   ├── dto/                      # Data Transfer Objects
│   ├── handlers/                 # HTTP handlers (5 endpoints)
│   ├── services/                 # Lógica de negocio
│   ├── router/                   # Definición de rutas
│   ├── domain/                   # Entidades (ResumeRequest, ProcessedResume)
│   ├── middleware/               # Auth JWT con JWKS
│   └── repository/               # Capa de persistencia (PostgreSQL)
├── pkg/                          # Código reutilizable
│   ├── converter/                # Conversión de archivos a PDF
│   └── client/                   # Cliente HTTP para Presigned URLs
├── migrations/                   # Migraciones SQL (auto-aplicadas)
├── docs/                         # Documentación OpenAPI y técnica
├── Dockerfile                    # Multi-stage build + migraciones
├── docker-compose.yml            # PostgreSQL + Backend
└── Makefile                      # Comandos útiles
```

**Métricas:** 23 archivos Go | ~2,500 líneas | Clean Architecture completa

---

## Tecnologías

| Categoría | Tecnología | Versión | Propósito |
|-----------|-----------|---------|-----------|
| **Lenguaje** | Go | 1.24.5 | Backend development |
| **Framework** | Fiber | v2.52.10 | HTTP server rápido |
| **Autenticación** | JWX | v2.1.6 | Validación JWT con JWKS |
| **Base de Datos** | PostgreSQL | 16 | Persistencia con JSONB |
| **Conversión PDF** | gofpdf | v1.16.2 | Generación de PDFs |
| **Lectura DOCX** | docx | v0.0.0 | Extracción de texto |
| **UUID** | google/uuid | v1.6.0 | Generación de Request IDs |
| **Configuración** | godotenv | v1.5.1 | Variables de entorno |
| **Cloud** | AWS S3 + Lambda | - | Storage + Processing |
| **Containers** | Docker + Compose | Latest | Orquestación |

---

## Inicio Rápido

### Prerrequisitos

- Go 1.24.5+
- Docker & Docker Compose
- Cuenta AWS (S3 + Lambda configurados)
- Servicio de Presigned URLs
- Servicio de Autenticación (JWKS endpoint)

### Instalación

```bash
# Clonar el repositorio
git clone <repository-url>
cd resume-backend-service

# Configurar variables de entorno
cp .env.example .env
# Editar .env con tus credenciales

# Opción 1: Ejecutar con Docker Compose (Recomendado)
make up
# Las migraciones se ejecutan automáticamente

# Opción 2: Ejecutar con Go localmente
make run
```

### Verificar Instalación

```bash
# Health check
curl http://localhost:8080/api/v1/health/

# Respuesta esperada:
{
  "status": "healthy",
  "service": "resume-backend-service"
}
```

---

## API Endpoints

### Health Check
```http
GET /api/v1/health/
```

**Autenticación:** No requerida

**Respuesta (200 OK):**
```json
{
  "status": "healthy",
  "service": "resume-backend-service"
}
```

---

### Procesar CV
```http
POST /api/v1/resume/
Authorization: Bearer <JWT_TOKEN>
Content-Type: multipart/form-data
```

**Parámetros:**
- `file` (required): Archivo CV (.pdf, .txt, .docx)
- `instructions` (optional): Instrucciones personalizadas
- `language` (optional): Idioma (default: "esp")

**Ejemplo con cURL:**
```bash
curl -X POST http://localhost:8080/api/v1/resume/ \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@cv.pdf" \
  -F "language=esp" \
  -F "instructions=Extraer experiencia de los últimos 5 años"
```

**Respuesta (202 Accepted):**
```json
{
  "status": "accepted",
  "message": "Solicitud encolada para procesamiento.",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Errores:**
- `400 Bad Request`: Archivo no enviado o formato no permitido
- `401 Unauthorized`: Token JWT inválido o ausente
- `500 Internal Server Error`: Error en conversión o upload

---

### Listar Mis CVs
```http
GET /api/v1/resume/my-resumes
Authorization: Bearer <JWT_TOKEN>
```

**Respuesta (200 OK):**
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

**Estados:** `pending`, `uploaded`, `processing`, `completed`, `failed`

---

### Obtener Detalle de CV
```http
GET /api/v1/resume/:request_id
Authorization: Bearer <JWT_TOKEN>
```

**Respuesta (200 OK):**
```json
{
  "request_id": "550e8400-...",
  "original_filename": "mi-cv.pdf",
  "status": "completed",
  "created_at": "2025-12-01T10:00:00Z",
  "completed_at": "2025-12-01T10:00:20Z",
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

**Errores:**
- `400`: Request ID inválido
- `401`: No autenticado
- `403`: El CV no pertenece al usuario
- `404`: CV no encontrado

---

### Recibir Resultados (Webhook)
```http
POST /api/v1/resume/results
Content-Type: application/json
```

**Body:**
```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "input_file": "s3://bucket/inputs/2025-12-01/10-00/cv-clean.pdf",
  "output_file": "s3://bucket/outputs/2025-12-01/10-00/cv-clean.json",
  "processing_time_ms": 11919,
  "status": "success",
  "structured_data": { ... }
}
```

**Respuesta (200 OK):**
```json
{
  "status": "success",
  "message": "Datos procesados correctamente."
}
```

Ver especificación completa en `docs/resume-backend-api.yaml`

---

## Configuración

### Variables de Entorno

Crea un archivo `.env` en la raíz del proyecto:

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

### Configuración de AWS

#### S3 Bucket
- Bucket: `cv-processor-dev` (configurable en servicio externo)
- Rutas:
  - `/inputs/{date}/{time}/cv-clean.pdf` - CVs subidos
  - `/outputs/{date}/{time}/cv-clean.json` - Resultados procesados
- Metadatos: `request-id`, `language`, `instructions`

#### Lambda
- Trigger: S3 event (PUT en `/inputs/`)
- Callback: POST a `/api/v1/resume/results`
- **Importante:** Debe extraer `request-id` del metadata de S3 y devolverlo en el callback

---

## Desarrollo

### Comandos Disponibles

El proyecto incluye un `Makefile` con comandos útiles:

```bash
# Desarrollo Local
make run        # Ejecutar servidor con go run
make build      # Construir y levantar con Docker
make logs       # Ver logs en tiempo real
make ps         # Ver estado de servicios

# Docker Compose
make up         # Levantar servicios (backend + postgres + migraciones)
make down       # Detener servicios
make clean      # Detener y eliminar volúmenes
```

### Compilación Manual

```bash
# Ejecutar localmente
go run cmd/main.go

# Compilar binario
go build -o bin/server cmd/main.go

# Limpiar dependencias
go mod tidy

# Formatear código
go fmt ./...
```

### Docker

```bash
# Build manual
docker build -t resume-backend .

# Run manual
docker run -p 8080:8080 --env-file .env resume-backend

# Docker Compose
docker-compose up -d
docker-compose logs -f backend
```

---

## Base de Datos

### Sistema de Migraciones Automáticas

El proyecto incluye migraciones que se ejecutan automáticamente al iniciar el contenedor Docker:

**Flujo:**
1. Container inicia
2. `docker-entrypoint.sh` espera a PostgreSQL
3. Ejecuta migraciones pendientes desde `/migrations/`
4. Registra en `schema_migrations`
5. Inicia la aplicación

Ver [`docs/MIGRATIONS.md`](docs/MIGRATIONS.md) para detalles completos.

### Modelo de Datos

#### Tabla: `resume_requests`
Tracking de solicitudes de procesamiento.

- **request_id** (UUID PK): ID único de la solicitud
- **user_id** (VARCHAR): ID del usuario (del JWT)
- **original_filename**: Nombre del archivo subido
- **status**: Estado (pending, uploaded, completed, failed)
- **s3_input_url**, **s3_output_url**: URLs de S3
- **processing_time_ms**: Tiempo de procesamiento
- **Timestamps:** created_at, uploaded_at, completed_at

#### Tabla: `processed_resumes`
CVs procesados con datos estructurados.

- **id** (BIGSERIAL PK): ID autoincremental
- **request_id** (UUID FK): Vinculado a resume_requests
- **user_id** (VARCHAR): ID del usuario
- **structured_data** (JSONB): Datos completos del CV
- **cv_name**, **cv_email**, **cv_phone**: Datos extraídos
- **Contadores:** education_count, experience_count, etc.

**Relación:** 1 request = 1 processed_resume (1:1)

Ver [`docs/REQUEST_ID_FLOW.md`](docs/REQUEST_ID_FLOW.md) para el flujo completo.

### Queries Útiles

```sql
-- Ver solicitudes de un usuario
SELECT * FROM resume_requests WHERE user_id = 'user-123' ORDER BY created_at DESC;

-- Buscar CVs por habilidad (JSONB query)
SELECT * FROM processed_resumes
WHERE structured_data @> '{"technicalSkills": {"skills": ["Go"]}}'::jsonb;

-- Estadísticas de procesamiento
SELECT status, COUNT(*) as count, AVG(processing_time_ms) as avg_time_ms
FROM resume_requests
GROUP BY status;
```

---

## Despliegue

### Dockerfile Multi-stage

El proyecto usa build multi-stage optimizado:

**Stage 1 - Builder:**
- Base: `golang:1.24-alpine`
- Instala: git, ca-certificates, tzdata
- Compilación estática: `CGO_ENABLED=0`
- Optimización: `ldflags "-w -s"`

**Stage 2 - Runtime:**
- Base: `alpine:latest` (mínimo)
- Usuario no-root: `appuser` (seguridad)
- Healthcheck integrado
- Migraciones automáticas
- Tamaño optimizado

### Docker Compose

Incluye:
- **PostgreSQL 16:** Base de datos con volumen persistente
- **Backend Service:** API REST con migraciones automáticas
- **Network:** `resume-network` (comunicación interna)
- **Volume:** `postgres_data` (persistencia)

```bash
# Levantar todos los servicios
docker-compose up -d

# Ver logs
docker-compose logs -f

# Detener
docker-compose down
```

---

## Documentación

### OpenAPI Specification

La especificación completa de la API está en `docs/resume-backend-api.yaml`

**Visualizar con Swagger UI:**
```bash
# Opción 1: Swagger Editor online
https://editor.swagger.io/
# Pegar contenido de resume-backend-api.yaml

# Opción 2: Swagger UI local
docker run -p 8081:8080 -e SWAGGER_JSON=/docs/resume-backend-api.yaml \
  -v $(pwd)/docs:/docs swaggerapi/swagger-ui
```

### Documentación Técnica

- **[CLAUDE.md](CLAUDE.md)** - Contexto completo para desarrollo asistido con Claude Code
- **[docs/REQUEST_ID_FLOW.md](docs/REQUEST_ID_FLOW.md)** - Flujo completo del sistema de Request ID
- **[docs/MIGRATIONS.md](docs/MIGRATIONS.md)** - Sistema de migraciones automáticas
- **[docs/IMPLEMENTATION_SUMMARY.md](docs/IMPLEMENTATION_SUMMARY.md)** - Resumen de implementación
- **[docs/resume-backend-api.yaml](docs/resume-backend-api.yaml)** - Especificación OpenAPI 3.0

---

## Roadmap

### Implementado
- ✅ Estructura Clean Architecture
- ✅ Autenticación JWT con middleware y JWKS
- ✅ Sistema de Request ID para tracking
- ✅ Conversión multi-formato a PDF
- ✅ Integración con AWS S3 y Lambda
- ✅ Persistencia en PostgreSQL
- ✅ Sistema de migraciones automáticas
- ✅ Repositorios y entidades de dominio
- ✅ Endpoint de upload asíncrono
- ✅ Endpoints de listado y detalle de CVs
- ✅ Estados de solicitud (pending → completed)
- ✅ Webhook para recibir resultados
- ✅ Docker y Docker Compose
- ✅ Documentación OpenAPI
- ✅ Configuración de CORS

### Pendiente
- ⏳ Endpoint de búsqueda por criterios (JSONB queries)
- ⏳ Estadísticas del usuario
- ⏳ Tests unitarios e integración
- ⏳ CI/CD pipeline
- ⏳ Rate limiting
- ⏳ Métricas y observabilidad (Prometheus, Grafana)
- ⏳ Notificaciones push
- ⏳ Webhooks configurables
- ⏳ Soporte para .doc (LibreOffice)
- ⏳ Cache (Redis)
- ⏳ Message queue (RabbitMQ/SQS)

---

## Arquitectura de Clean Code

Este proyecto sigue [Standard Go Project Layout](https://github.com/golang-standards/project-layout) y principios de Clean Architecture:

- **Separación de capas:** Handlers → Services → Repositories → Domain
- **Inyección de dependencias:** Constructores explícitos
- **DTOs:** Desacoplamiento de estructuras internas
- **Domain entities:** Lógica de negocio en entidades
- **Repository pattern:** Abstracción de acceso a datos
- **Código privado:** `internal/` no exportable
- **Código reutilizable:** `pkg/` compartible

---

## Integración con Otros Servicios

### Presigned URL Service

El servicio debe incluir el `request_id` en la firma de la presigned URL:

```javascript
// Backend envía:
{
  "filename": "cv-clean.pdf",
  "content_type": "application/pdf",
  "metadata": {
    "request_id": "550e8400-...",  // Importante
    "language": "esp",
    "instructions": "..."
  }
}

// Servicio debe incluir en la firma:
const s3Metadata = {
  'request-id': metadata.request_id,
  'language': metadata.language,
  'instructions': metadata.instructions
}
```

### AWS Lambda

Lambda debe extraer y devolver el `request_id`:

```javascript
// Extraer metadata
const requestId = s3Object.Metadata['request-id']

// Callback
await axios.post('https://backend/api/v1/resume/results', {
  request_id: requestId,  // Crítico
  input_file: inputKey,
  output_file: outputKey,
  status: 'success',
  structured_data: extractedData
})
```

Ver [`docs/REQUEST_ID_FLOW.md`](docs/REQUEST_ID_FLOW.md) y [`docs/IMPLEMENTATION_SUMMARY.md`](docs/IMPLEMENTATION_SUMMARY.md) para detalles completos.

---

## Recursos

- [Fiber Framework Documentation](https://docs.gofiber.io/)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [OpenAPI Specification](https://swagger.io/specification/)
- [PostgreSQL JSONB](https://www.postgresql.org/docs/current/datatype-json.html)
- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)
- [AWS S3 Documentation](https://docs.aws.amazon.com/s3/)
- [AWS Lambda Documentation](https://docs.aws.amazon.com/lambda/)

---

## Licencia

[Especificar licencia]

---

## Contribuciones

[Especificar guías de contribución]

---

**Última actualización:** 2025-12-01
**Branch principal:** main
**Commits recientes:**
- `f1d892e` - fix: corregir nombres de columnas en query de listado
- `206b13a` - docs: actualizar OpenAPI con nuevos endpoints
- `9405584` - feat: agregar endpoints de listado y detalle de CVs
- `a5ac649` - fix: manejar campos NULL en queries de BD

---

**Desarrollado con Go, Fiber, PostgreSQL y AWS**
