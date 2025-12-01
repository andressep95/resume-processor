# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Instalar dependencias del sistema necesarias para compilar
RUN apk add --no-cache git ca-certificates tzdata

# Establecer el directorio de trabajo
WORKDIR /build

# Copiar archivos de dependencias primero (para aprovechar cache de Docker)
COPY go.mod ./
# Copiar go.sum si existe (opcional)
COPY go.sum* ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
# CGO_ENABLED=0 para crear un binario estático
# -ldflags="-w -s" para reducir el tamaño del binario
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o resume-backend-service \
    ./cmd/main.go

# Stage 2: Runtime
FROM alpine:latest

# Instalar dependencias: certificados, curl, postgresql-client para migraciones
RUN apk --no-cache add ca-certificates tzdata curl postgresql-client

# Crear usuario no-root para seguridad
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copiar el binario compilado desde el stage de build
COPY --from=builder /build/resume-backend-service .

# Copiar migraciones
COPY migrations ./migrations

# Copiar entrypoint script
COPY docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

# Cambiar propiedad al usuario no-root
RUN chown -R appuser:appuser /app

# Cambiar al usuario no-root
USER appuser

# Exponer el puerto (configurable via ENV)
EXPOSE 8080

# Healthcheck dinámico que usa la variable SERVER_PORT (default 8080)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${SERVER_PORT:-8080}/api/v1/health/ || exit 1

# Usar entrypoint para ejecutar migraciones antes de iniciar la app
ENTRYPOINT ["./docker-entrypoint.sh"]
CMD ["./resume-backend-service"]

