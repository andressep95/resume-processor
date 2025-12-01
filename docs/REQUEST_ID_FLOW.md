# 🔄 Flujo de Request ID - Sistema de Tracking de CVs

## 📋 Resumen

Este documento describe el flujo completo de procesamiento de CVs utilizando **Request ID** como mecanismo de tracking y vinculación de datos entre el usuario, la solicitud, y los resultados procesados.

---

## 🎯 Objetivo

Vincular de forma confiable:
- **Usuario** que sube el CV (user_id, user_email del JWT)
- **Solicitud** de procesamiento (request_id, estado, timestamps)
- **Datos procesados** por AWS Lambda (información estructurada del CV)

---

## 🔑 Concepto Clave: Request ID

El **Request ID** es un UUID v4 generado al momento de recibir la solicitud de procesamiento. Este ID:

✅ **Es único** por solicitud
✅ **Viaja con el archivo** a través de metadatos de S3
✅ **Permite tracking** del estado del procesamiento
✅ **Vincula** usuario → solicitud → resultado
✅ **No depende** de que AWS Lambda devuelva el user_email

---

## 📊 Flujo Completo

### 1️⃣ Usuario Sube CV

**Endpoint:** `POST /api/v1/resume/`

**Headers:**
```http
Authorization: Bearer <JWT_TOKEN>
Content-Type: multipart/form-data
```

**Form Data:**
```
file: resume.pdf
language: esp
instructions: "Extraer últimos 5 años"
```

**Middleware de Autenticación:**
- Valida JWT token
- Extrae `user_id` y `user_email` del token
- Almacena en `c.Locals("user_id")` y `c.Locals("user_email")`

**Handler:** `resume_handler.go:ProcessResumeHandler()`
```go
userID := c.Locals("user_id").(string)      // ej: "user-123"
userEmail := c.Locals("user_email").(string) // ej: "usuario@example.com"
```

---

### 2️⃣ Servicio Genera Request ID

**Service:** `resume_service.go:ProcessResume()`

**Paso 2.1 - Crear Request:**
```go
resumeRequest := domain.NewResumeRequest(
    userID,           // "user-123"
    userEmail,        // "usuario@example.com"
    filename,         // "mi-cv.pdf"
    fileType,         // ".pdf"
    fileSize,         // 524288
    language,         // "esp"
    instructions,     // "Extraer últimos 5 años"
)
// resumeRequest.RequestID = UUID generado automáticamente
// Ejemplo: "550e8400-e29b-41d4-a716-446655440000"
```

**Paso 2.2 - Guardar en Base de Datos:**
```go
resumeRequestRepo.Create(resumeRequest)
```

**Tabla: `resume_requests`**
```
request_id: 550e8400-e29b-41d4-a716-446655440000
user_id: user-123
user_email: usuario@example.com
original_filename: mi-cv.pdf
status: pending
created_at: 2025-11-30T10:00:00Z
```

---

### 3️⃣ Subida a S3 con Request ID

**Paso 3.1 - Obtener Presigned URL:**
```go
presignedResp := presignedURLClient.GetUploadURL(
    filename,
    "application/pdf",
    requestID,           // ⭐ Se envía para ser incluido en la firma
    language,
    instructions,
)
```

**Request al Servicio de Presigned URLs:**
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

**Paso 3.2 - Upload a S3 con Metadata:**
```go
req.Header.Set("Content-Type", "application/pdf")
req.Header.Set("x-amz-meta-request-id", requestID)      // ⭐ CLAVE
req.Header.Set("x-amz-meta-language", language)
req.Header.Set("x-amz-meta-instructions", instructions)
```

**S3 Object Metadata:**
```
Key: inputs/2025-11-30/10-00-00/cv-clean.pdf
Metadata:
  request-id: "550e8400-e29b-41d4-a716-446655440000"
  language: "esp"
  instructions: "Extraer últimos 5 años"
```

> **Nota:** Ya NO enviamos `user-email` porque es redundante. El `request-id` ya vincula con el usuario en la BD.

**Paso 3.3 - Actualizar Estado:**
```go
resumeRequestRepo.MarkAsUploaded(requestID, s3InputURL)
```

**Tabla: `resume_requests` (actualizada)**
```
status: uploaded
s3_input_url: s3://bucket/inputs/2025-11-30/10-00-00/cv-clean.pdf
uploaded_at: 2025-11-30T10:00:05Z
```

---

### 4️⃣ Respuesta al Cliente

**Response (202 Accepted):**
```json
{
  "status": "accepted",
  "message": "Solicitud encolada para procesamiento.",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

El cliente ahora puede:
- Guardar el `request_id` para tracking
- Hacer polling del estado (futuro endpoint)
- Recibir notificación cuando complete

---

### 5️⃣ AWS Lambda Procesa CV

**Trigger:** S3 Event (PUT en `/inputs/`)

**Lambda Function:**
1. Lee el PDF de S3
2. Extrae metadata del objeto S3:
   ```javascript
   const requestId = s3Object.metadata['request-id']      // ⭐ CRÍTICO
   const language = s3Object.metadata['language']
   const instructions = s3Object.metadata['instructions']
   ```
3. Procesa el CV con IA
4. Genera `structured_data` (JSON)
5. Sube resultado a S3: `outputs/2025-11-30/10-00-00/cv-clean.json`
6. **Callback al backend** con el `request_id`

---

### 6️⃣ Callback de AWS Lambda

**Endpoint:** `POST /api/v1/resume/results`

**Request Body:**
```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "input_file": "s3://bucket/inputs/2025-11-30/10-00-00/cv-clean.pdf",
  "output_file": "s3://bucket/outputs/2025-11-30/10-00-00/cv-clean.json",
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
    "education": [...],
    "professionalExperience": [...],
    "technicalSkills": {...},
    "certifications": [...],
    "projects": [...]
  }
}
```

---

### 7️⃣ Backend Vincula Datos

**Handler:** `aws_handler.go:ProcessResumeResultsHandler()`

**Paso 7.1 - Validar Request ID:**
```go
requestID, err := uuid.Parse(lambdaResponse.RequestID)
if err != nil {
    return c.Status(400).JSON(...)
}
```

**Paso 7.2 - Buscar Solicitud Original:**
```go
resumeRequest, err := resumeRequestRepo.FindByRequestID(requestID)
// Obtiene:
// - user_id: "user-123"
// - user_email: "usuario@example.com"
// - original_filename: "mi-cv.pdf"
// - status: "uploaded"
```

✅ **Aquí se vincula el resultado con el usuario original**

**Paso 7.3 - Crear CV Procesado:**
```go
processedResume := domain.NewProcessedResume(
    requestID,
    resumeRequest.UserID,  // user-123 (del request original)
    &lambdaResponse.StructuredData,
)
```

**Paso 7.4 - Guardar en Base de Datos:**
```go
processedResumeRepo.Create(processedResume)
```

**Tabla: `processed_resumes`**
```
id: 1
request_id: 550e8400-e29b-41d4-a716-446655440000
user_id: user-123
structured_data: {...JSON...}
cv_name: Juan Pérez
cv_email: juan.perez@example.com
cv_phone: +34 600 123 456
education_count: 2
experience_count: 3
created_at: 2025-11-30T10:00:20Z
```

**Paso 7.5 - Actualizar Estado de Solicitud:**
```go
resumeRequestRepo.MarkAsCompleted(requestID, outputFile, processingTimeMs)
```

**Tabla: `resume_requests` (actualizada)**
```
status: completed
s3_output_url: s3://bucket/outputs/2025-11-30/10-00-00/cv-clean.json
processing_time_ms: 11919
completed_at: 2025-11-30T10:00:20Z
```

---

## 🗄️ Modelo de Datos

### Tabla: `resume_requests`

Tracking de solicitudes de procesamiento.

```sql
CREATE TABLE resume_requests (
    request_id UUID PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    user_email VARCHAR(255) NOT NULL,
    original_filename VARCHAR(500),
    original_file_type VARCHAR(10),
    file_size_bytes BIGINT,
    language VARCHAR(10),
    instructions TEXT,
    s3_input_url TEXT,
    s3_output_url TEXT,
    status VARCHAR(20),  -- pending, uploaded, processing, completed, failed
    processing_time_ms BIGINT,
    error_message TEXT,
    created_at TIMESTAMP,
    uploaded_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

### Tabla: `processed_resumes`

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
    education_count INT,
    experience_count INT,
    certifications_count INT,
    projects_count INT,
    skills_count INT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

**Relación:** 1:1 (un request = un CV procesado)

---

## 🔍 Queries Útiles

### Ver solicitudes de un usuario

```sql
SELECT * FROM resume_requests
WHERE user_id = 'user-123'
ORDER BY created_at DESC;
```

### Ver CVs procesados de un usuario

```sql
SELECT pr.*, rr.original_filename, rr.created_at as requested_at
FROM processed_resumes pr
INNER JOIN resume_requests rr ON pr.request_id = rr.request_id
WHERE pr.user_id = 'user-123'
ORDER BY pr.created_at DESC;
```

### Ver estado de una solicitud específica

```sql
SELECT
    rr.request_id,
    rr.status,
    rr.created_at,
    rr.completed_at,
    rr.processing_time_ms,
    pr.cv_name,
    pr.cv_email
FROM resume_requests rr
LEFT JOIN processed_resumes pr ON rr.request_id = pr.request_id
WHERE rr.request_id = '550e8400-e29b-41d4-a716-446655440000';
```

### Buscar por habilidad (usando JSONB)

```sql
SELECT * FROM processed_resumes
WHERE structured_data @> '{"technicalSkills": {"skills": ["Go"]}}'::jsonb;
```

---

## 🚨 Manejo de Errores

### Error en Conversión de Archivo

```go
// Marcar solicitud como fallida
resumeRequestRepo.MarkAsFailed(requestID, "Error al convertir archivo a PDF")
```

### Error en AWS Lambda

```json
{
  "request_id": "550e8400-...",
  "status": "failed",
  "error_message": "PDF corrupto"
}
```

**Backend:**
```go
resumeRequestRepo.MarkAsFailed(requestID, "AWS Lambda reportó status: failed")
```

### Request ID no encontrado

Si AWS callback envía un `request_id` que no existe en la BD:
```go
return c.Status(404).JSON(fiber.Map{
    "status": "error",
    "message": "Solicitud no encontrada."
})
```

---

## 📝 Próximos Pasos para Otros Servicios

### 1. Servicio de Presigned URLs

**Debe recibir y usar el request-id en la presigned URL:**

```javascript
// Endpoint: POST /api/v1/presigned-url/upload
// Body: { filename, content_type, metadata: { request_id, language, instructions } }

app.post('/api/v1/presigned-url/upload', (req, res) => {
  const { filename, content_type, metadata } = req.body

  // ⭐ CRÍTICO: Usar el request_id del request para generar la firma
  const s3Metadata = {
    'request-id': metadata.request_id,  // Del request body
    'language': metadata.language,
    'instructions': metadata.instructions
  }

  const presignedUrl = s3.getSignedUrl('putObject', {
    Bucket: bucketName,
    Key: key,
    Metadata: s3Metadata,  // ⭐ Incluir en la firma
    Expires: 3600
  })

  res.json({ url: presignedUrl, expires_in: '1 hour' })
})
```

**Importante:** El servicio debe incluir el `request-id` recibido en la firma de la presigned URL, de lo contrario S3 rechazará el upload con error `SignatureDoesNotMatch`.

### 2. AWS Lambda

**Debe extraer y devolver el request-id:**

```javascript
// Extraer metadata del objeto S3
const s3Object = await s3.getObject({
  Bucket: event.Records[0].s3.bucket.name,
  Key: event.Records[0].s3.object.key
}).promise()

const requestId = s3Object.Metadata['request-id']

// Callback al backend
await axios.post('https://backend/api/v1/resume/results', {
  request_id: requestId,           // ⭐ Importante
  input_file: inputKey,
  output_file: outputKey,
  processing_time_ms: processingTime,
  status: 'success',
  structured_data: extractedData
})
```

---

## ✅ Ventajas de esta Arquitectura

1. **No depende de AWS Lambda** para mantener user_email
2. **Auditoría completa** de todas las solicitudes
3. **Tracking de estado** en tiempo real (pending → uploaded → completed)
4. **Resiliencia** ante fallos (se pueden reintentar solicitudes fallidas)
5. **Escalabilidad** (base de datos relacional con índices optimizados)
6. **Trazabilidad** completa del ciclo de vida del procesamiento

---

## 🎯 Resumen

```
Usuario → JWT → Backend
              ↓
         Genera REQUEST_ID
              ↓
         Guarda en BD (pending)
              ↓
         Sube a S3 con metadata (request-id)
              ↓
         Responde con request_id
              ↓
    AWS Lambda lee metadata
              ↓
    Procesa CV y callback con request-id
              ↓
    Backend busca solicitud por request-id
              ↓
    Obtiene user_id de la solicitud
              ↓
    Guarda CV procesado vinculado a user_id
              ↓
    Actualiza estado a completed
```

---

**Fecha:** 2025-11-30
**Versión:** 1.0
