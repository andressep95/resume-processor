# Resume Backend Service

Microservicio backend para gestión de currículums (CVs) que acepta archivos en formato .txt, .doc, .docx y .pdf.

## 🏗️ Arquitectura del Proyecto

Este proyecto sigue las **mejores prácticas de Go** basadas en el [Standard Go Project Layout](https://github.com/golang-standards/project-layout) y principios de **Clean Architecture**.

```
resume-backend-service/
├── cmd/                    # Aplicaciones principales
│   └── main.go            # Punto de entrada de la aplicación
│
├── internal/              # Código privado de la aplicación
│   ├── config/           # Configuración y variables de entorno
│   ├── domain/           # Entidades de dominio y lógica de negocio
│   ├── dto/              # Data Transfer Objects (Request/Response)
│   ├── handlers/         # HTTP handlers (controladores)
│   ├── middleware/       # Middlewares personalizados
│   ├── repository/       # Capa de acceso a datos
│   ├── routes/           # Definición de rutas
│   └── services/         # Lógica de negocio
│
├── pkg/                   # Código reutilizable (puede ser usado por otras apps)
│   ├── converter/        # Conversión de archivos a PDF
│   ├── utils/            # Utilidades generales
│   └── validator/        # Validación de archivos
│
├── docs/                  # Documentación del API
│   └── resume-backend-api.yaml
│
├── .dockerignore
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 📦 Descripción de Carpetas

### `/cmd`
Contiene los puntos de entrada de la aplicación. El archivo `main.go` es minimalista (3 líneas) y delega la inicialización al bootstrap.

### `/internal`
Código privado de la aplicación que no puede ser importado por otros proyectos.

- **config/**: Manejo de configuración con variables de entorno y bootstrap de la aplicación
- **domain/**: Entidades de dominio (ej: Resume, User)
- **dto/**: Estructuras para requests y responses HTTP
- **handlers/**: Manejadores HTTP (similar a controllers)
- **middleware/**: Middlewares personalizados (auth, CORS, etc.)
- **repository/**: Interfaz y implementación de acceso a datos
- **routes/**: Registro de rutas HTTP
- **services/**: Lógica de negocio de la aplicación

### `/pkg`
Código que puede ser reutilizado por aplicaciones externas.

- **converter/**: Lógica para convertir .doc, .docx, .txt a PDF
- **validator/**: Validación de tipos y tamaños de archivos
- **utils/**: Funciones utilitarias generales

### `/docs`
Documentación de la API (OpenAPI/Swagger).

## 🚀 Comandos Disponibles

El proyecto incluye un `Makefile` con los siguientes comandos:

```bash
# Ejecutar el servidor localmente
make run

# Docker Compose
make up      # Levantar servicios
make down    # Detener servicios
make build   # Construir y levantar
make logs    # Ver logs
make ps      # Ver estado de servicios
make clean   # Limpiar volúmenes
```

## 🔧 Configuración

Las variables de entorno se pueden configurar en un archivo `.env` o directamente en el sistema:

```env
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_READ_TIMEOUT=10
SERVER_WRITE_TIMEOUT=10

# App
APP_NAME=Resume Backend Service
APP_ENV=development

# Storage
UPLOAD_PATH=./uploads
MAX_FILE_SIZE=10485760  # 10MB en bytes
```

## 🛠️ Tecnologías

- **Go 1.24+**
- **Fiber v2** - Framework HTTP
- **PostgreSQL** (futuro)
- **Docker & Docker Compose**

## 📝 Estructura del Código

### Main (cmd/main.go)
```go
package main

import "resume-backend-service/internal/config"

func main() {
    app := config.Bootstrap()
    app.Run()
}
```

### Bootstrap (internal/config/bootstrap.go)
Inicializa toda la aplicación:
- Carga de configuración desde variables de entorno
- Setup de Fiber con middlewares (logger, recover)
- Registro de rutas centralizadas

### Ejemplo: Health Check
```bash
# Ejecutar servidor
make run

# Probar health check
curl http://localhost:8080/api/v1/health

# Respuesta esperada:
{
  "status": "healthy",
  "service": "resume-backend-service"
}
```

## 🎯 Próximos Pasos

1. Implementar entidades de dominio en `/internal/domain/`
2. Crear DTOs para requests/responses en `/internal/dto/`
3. Implementar lógica de conversión de archivos en `/pkg/converter/`
4. Crear servicios de negocio en `/internal/services/`
5. Implementar repositorios en `/internal/repository/`
6. Agregar handlers para upload y procesamiento de CVs
7. Configurar base de datos y migraciones

## 📚 Recursos

- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Fiber Framework](https://docs.gofiber.io/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
