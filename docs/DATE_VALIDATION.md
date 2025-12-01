# Validación de Fechas en Certificaciones

## Descripción

El sistema valida automáticamente las fechas en la sección de certificaciones del JSON recibido de AWS Lambda. Si una fecha no cumple con patrones válidos, se reemplaza por una cadena vacía (`""`).

## Patrones de Fecha Válidos

El sistema acepta los siguientes formatos de fecha:

### Formatos Completos
- `2024-01-15` (ISO 8601: YYYY-MM-DD)
- `15/01/2024` (DD/MM/YYYY)
- `2024/01/15` (YYYY/MM/DD)

### Formatos Parciales
- `2024-01` (Año-Mes)
- `01/2024` (Mes/Año)
- `2024` (Solo año)

### Formatos con Texto
- `Jan 2024` (Mes abreviado en inglés + año)
- `Feb 2024`, `Mar 2024`, etc.

### Validaciones Adicionales
- Años válidos: 1900 - 2100
- Meses válidos: 01 - 12
- Días válidos: 01 - 31
- Cadenas vacías son válidas

## Campos Validados

En la sección `certifications`, se validan los siguientes campos:

- `dateObtained`
- `issueDate`
- `expiryDate`
- `date`

## Ejemplos

### Entrada de AWS Lambda

```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "success",
  "structured_data": {
    "certifications": [
      {
        "name": "AWS Certified Solutions Architect",
        "dateObtained": "2024-01-15"
      },
      {
        "name": "Google Cloud Professional",
        "dateObtained": "No especificado"
      },
      {
        "name": "Azure Administrator",
        "issueDate": "Presente",
        "expiryDate": "2025-12-31"
      },
      {
        "name": "Kubernetes Certified",
        "date": "N/A"
      },
      {
        "name": "Docker Certified",
        "dateObtained": "2023"
      }
    ]
  }
}
```

### Salida Sanitizada (guardada en BD)

```json
{
  "certifications": [
    {
      "name": "AWS Certified Solutions Architect",
      "dateObtained": "2024-01-15"
    },
    {
      "name": "Google Cloud Professional",
      "dateObtained": ""
    },
    {
      "name": "Azure Administrator",
      "issueDate": "",
      "expiryDate": "2025-12-31"
    },
    {
      "name": "Kubernetes Certified",
      "date": ""
    },
    {
      "name": "Docker Certified",
      "dateObtained": "2023"
    }
  ]
}
```

## Implementación

### Función Principal

```go
// SanitizeStructuredData sanitiza todo el JSON recibido de AWS
func SanitizeStructuredData(data interface{}) (map[string]interface{}, error)
```

### Uso en el Handler

```go
// En internal/handlers/aws_handler.go
structuredDataMap, err := utils.SanitizeStructuredData(lambdaResponse.StructuredData)
if err != nil {
    // Manejar error
}
```

## Casos de Uso

### ✅ Fechas Válidas (se mantienen)
- `"2024-01-15"` → `"2024-01-15"`
- `"15/01/2024"` → `"15/01/2024"`
- `"2024"` → `"2024"`
- `"Jan 2024"` → `"Jan 2024"`
- `""` → `""` (vacío es válido)

### ❌ Fechas Inválidas (se limpian)
- `"No especificado"` → `""`
- `"Presente"` → `""`
- `"N/A"` → `""`
- `"Vigente"` → `""`
- `"En curso"` → `""`
- `"2024/13/45"` → `""` (mes/día fuera de rango)

## Beneficios

1. **Consistencia de Datos**: Garantiza que solo fechas válidas se almacenen en la BD
2. **Prevención de Errores**: Evita problemas al parsear fechas en el frontend
3. **Flexibilidad**: Acepta múltiples formatos de fecha comunes
4. **Limpieza Automática**: No requiere intervención manual
5. **Queries Confiables**: Facilita búsquedas y filtros por fecha en PostgreSQL

## Testing

Ejecutar tests de validación:

```bash
go test -v ./pkg/utils/
```

## Notas Técnicas

- La validación se ejecuta **antes** de guardar en la base de datos
- No afecta otros campos del JSON, solo fechas en certificaciones
- La función es idempotente (puede ejecutarse múltiples veces sin efectos secundarios)
- Los logs del backend muestran los datos ya sanitizados

## Extensión Futura

Si se necesita validar fechas en otras secciones (educación, experiencia), se puede extender fácilmente:

```go
// Ejemplo para validar fechas en educación
func SanitizeEducationDates(structuredData map[string]interface{}) {
    education, ok := structuredData["education"].([]interface{})
    if !ok {
        return
    }
    
    for _, edu := range education {
        eduMap, ok := edu.(map[string]interface{})
        if !ok {
            continue
        }
        
        if gradDate, exists := eduMap["graduationDate"]; exists {
            if dateStr, ok := gradDate.(string); ok {
                if !IsValidDate(dateStr) {
                    eduMap["graduationDate"] = ""
                }
            }
        }
    }
}
```

---

**Última actualización:** 2025-12-01  
**Versión:** 1.0.0
