# Resume Backend Service

[![Go Version](https://img.shields.io/badge/Go-1.24.5-00ADD8?logo=go)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-v2.52.10-00ACD7?logo=go)](https://gofiber.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-316192?logo=postgresql)](https://www.postgresql.org/)

Microservicio backend en Go para procesamiento asíncrono de currículums (CVs) mediante integración con AWS Lambda y S3. Acepta archivos en múltiples formatos (.pdf, .txt, .docx), los convierte a PDF estandarizado, y procesa la información mediante inteligencia artificial.

## Tabla de Contenidos

- [Características](#-características)
- [Arquitectura](#-arquitectura)
- [Tecnologías](#-tecnologías)
- [Inicio Rápido](#-inicio-rápido)
- [API Endpoints](#-api-endpoints)
- [Configuración](#-configuración)
- [Desarrollo](#-desarrollo)
- [Despliegue](#-despliegue)
- [Documentación](#-documentación)

---

## Características

- **Procesamiento Asíncrono:** Upload y procesamiento no bloqueante de CVs
- **Conversión Multi-formato:** Soporte para .pdf, .txt, .docx (conversión automática a PDF)
- **Integración AWS:** S3 para almacenamiento y Lambda para procesamiento con IA
- **Extracción Estructurada:** Datos organizados (contacto, experiencia, educación, skills, etc.)
- **API RESTful:** Endpoints bien documentados con OpenAPI 3.0
- **Docker Ready:** Containerización completa con Docker Compose
- **Clean Architecture:** Código organizado, mantenible y escalable
- **Health Checks:** Monitoreo de disponibilidad del servicio

---

## Arquitectura

### Flujo de Procesamiento

```
Cliente → Backend → Presigned URL Service → AWS S3 → AWS Lambda → Backend (Callback)
  ↓          ↓                                  ↓           ↓             ↓
Upload    Convierte                          Trigger   Procesa       Almacena
  CV      a PDF                              Lambda      CV          Resultados
         (si necesario)                                              (TODO: DB)
```

### Estructura del Proyecto

```
resume-backend-service/
├── cmd/                    # Punto de entrada de la aplicación
│   └── main.go            # Main minimalista (delega a bootstrap)
│
├── internal/              # Código privado de la aplicación
│   ├── config/           # Configuración y bootstrap
│   ├── dto/              # Data Transfer Objects
│   ├── handlers/         # HTTP handlers (controladores)
│   ├── services/         # Lógica de negocio
│   ├── router/           # Definición de rutas
│   ├── domain/           # Entidades de dominio (preparado)
│   ├── middleware/       # Middlewares personalizados (preparado)
│   └── repository/       # Capa de acceso a datos (preparado)
│
├── pkg/                   # Código reutilizable
│   ├── converter/        # Conversión de archivos a PDF
│   ├── client/           # Clientes HTTP externos
│   ├── utils/            # Utilidades generales (preparado)
│   └── validator/        # Validación de archivos (preparado)
│
├── docs/                  # Documentación OpenAPI
│   └── resume-backend-api.yaml
│
├── Dockerfile            # Build multi-stage optimizado
├── docker-compose.yml    # PostgreSQL + Backend
├── Makefile             # Comandos útiles
└── .env.example         # Template de configuración
```

**Código:** 13 archivos Go | ~508 líneas | Clean Architecture

---

## Tecnologías

| Categoría | Tecnología | Versión | Propósito |
|-----------|-----------|---------|-----------|
| **Lenguaje** | Go | 1.24.5 | Backend development |
| **Framework** | Fiber | v2.52.10 | HTTP server rápido |
| **Conversión PDF** | gofpdf | v1.16.2 | Generación de PDFs |
| **Lectura DOCX** | docx | v0.0.0 | Extracción de texto |
| **Configuración** | godotenv | v1.5.1 | Variables de entorno |
| **Base de Datos** | PostgreSQL | 16 | Persistencia (preparado) |
| **Cloud** | AWS S3 + Lambda | - | Storage + Processing |
| **Containers** | Docker + Compose | Latest | Orquestación |

---

## Inicio Rápido

### Prerrequisitos

- Go 1.24.5+
- Docker & Docker Compose
- Cuenta AWS (S3 + Lambda configurados)
- Servicio de Presigned URLs

### Instalación

```bash
# Clonar el repositorio
git clone <repository-url>
cd resume-backend-service

# Configurar variables de entorno
cp .env.example .env
# Editar .env con tus credenciales

# Opción 1: Ejecutar con Go
make run

# Opción 2: Ejecutar con Docker Compose
make up
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
Content-Type: multipart/form-data
```

**Parámetros:**
- `file` (required): Archivo CV (.pdf, .txt, .docx)
- `instructions` (optional): Instrucciones personalizadas
- `language` (optional): Idioma (default: "esp")

**Ejemplo con cURL:**
```bash
curl -X POST http://localhost:8080/api/v1/resume/ \
  -F "file=@cv.pdf" \
  -F "language=esp" \
  -F "instructions=Extraer experiencia de los últimos 5 años"
```

**Respuesta (202 Accepted):**
```json
{
  "status": "accepted",
  "message": "Solicitud encolada para procesamiento."
}
```

**Errores:**
- `400 Bad Request`: Archivo no enviado o formato no permitido
- `500 Internal Server Error`: Error en conversión o upload

---

### Recibir Resultados (Webhook)
```http
POST /api/v1/resume/results
Content-Type: application/json
```

**Body:**
```json
{
  "input_file": "s3://bucket/inputs/2025-11-29/14-30-00/cv-clean.pdf",
  "output_file": "s3://bucket/outputs/2025-11-29/14-30-00/cv-clean.json",
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
    "professionalExperience": [
      {
        "company": "Tech Corp",
        "position": "Senior Developer",
        "period": { "start": "2020-01", "end": "2023-12" },
        "responsibilities": [
          "Desarrollo de aplicaciones web",
          "Liderazgo de equipo técnico"
        ]
      }
    ],
    "education": [
      {
        "institution": "Universidad de Madrid",
        "degree": "Ingeniería en Informática",
        "graduationDate": "2019-06"
      }
    ],
    "technicalSkills": {
      "skills": ["JavaScript", "React", "Node.js", "Go"]
    },
    "certifications": [
      {
        "name": "AWS Certified Developer",
        "dateObtained": "2021-03"
      }
    ],
    "projects": [
      {
        "name": "E-commerce Platform",
        "description": "Plataforma escalable",
        "technologies": ["React", "Node.js", "MongoDB"]
      }
    ]
  }
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
SERVER_PORT=8081                    # Puerto del servidor (default: 8080)

# Archivos
MAX_FILE_SIZE_MB=10                 # Tamaño máximo en MB (default: 10)

# Servicios Externos
PRESIGNED_URL_SERVICE_ENDPOINT=https://api.cloudcentinel.com/signature/api/v1/presigned-url/upload
```

### Configuración de AWS

#### S3 Bucket
- Bucket: `cv-processor-dev` (configurable en servicio externo)
- Rutas:
  - `/inputs/{date}/{time}/cv-clean.pdf`
  - `/outputs/{date}/{time}/cv-clean.json`

#### Lambda
- Trigger: S3 event (PUT en `/inputs/`)
- Callback: POST a `/api/v1/resume/results`
- Metadatos disponibles: `language`, `instructions`

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
make up         # Levantar servicios (backend + postgres)
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
- Tamaño optimizado

### Docker Compose

Incluye:
- **PostgreSQL 16:** Base de datos (preparado para futuro uso)
- **Backend Service:** API REST
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

### CLAUDE.md

El archivo `CLAUDE.md` contiene contexto completo para desarrollo asistido con Claude Code:
- Arquitectura detallada
- Flujos completos de procesamiento
- Convenciones de código
- TODOs y próximos pasos

---

## Roadmap

### Implementado
- ✅ Estructura Clean Architecture
- ✅ Conversión multi-formato a PDF
- ✅ Integración con AWS S3 y Lambda
- ✅ Endpoint de upload asíncrono
- ✅ Webhook para recibir resultados
- ✅ Docker y Docker Compose
- ✅ Documentación OpenAPI

### Pendiente
- ⏳ Persistencia en PostgreSQL
- ⏳ Autenticación y autorización
- ⏳ Tests unitarios e integración
- ⏳ CI/CD pipeline
- ⏳ Rate limiting
- ⏳ Métricas y observabilidad
- ⏳ Soporte para .doc (LibreOffice)

---

## Arquitectura de Clean Code

Este proyecto sigue [Standard Go Project Layout](https://github.com/golang-standards/project-layout) y principios de Clean Architecture:

- **Separación de capas:** Handlers → Services → Repositories
- **Inyección de dependencias:** Constructores explícitos
- **DTOs:** Desacoplamiento de estructuras internas
- **Código privado:** `internal/` no exportable
- **Código reutilizable:** `pkg/` compartible

---

## Recursos

- [Fiber Framework Documentation](https://docs.gofiber.io/)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [OpenAPI Specification](https://swagger.io/specification/)
- [AWS S3 Documentation](https://docs.aws.amazon.com/s3/)
- [AWS Lambda Documentation](https://docs.aws.amazon.com/lambda/)

---

## Licencia

[Especificar licencia]

---

## Contribuciones

[Especificar guías de contribución]

---

**Última actualización:** 2025-11-29
**Commits recientes:**
- `6065bb0` - Corregir parseo de datos de AWS Lambda con estructura wrapper
- `92d677e` - Mejorar logging del endpoint de resultados procesados
- `5a49811` - Corregir puerto en Docker
