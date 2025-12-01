# 📊 Sistema de Migraciones Automáticas

## 🎯 Descripción

El servicio implementa un sistema de migraciones automáticas que se ejecuta al iniciar el contenedor Docker, asegurando que la base de datos esté siempre actualizada con el esquema correcto.

---

## 🔄 Flujo de Inicio del Contenedor

```
1. Container starts
       ↓
2. docker-entrypoint.sh ejecuta
       ↓
3. ⏳ Espera a PostgreSQL (hasta que esté listo)
       ↓
4. 📊 Ejecuta migraciones pendientes
       ↓
5. 🚀 Inicia la aplicación
```

---

## 📁 Estructura de Migraciones

```
migrations/
├── 001_create_resume_tables.sql    # Primera migración
├── 002_add_indexes.sql              # Segunda migración (ejemplo futuro)
└── 003_alter_tables.sql             # Tercera migración (ejemplo futuro)
```

**Convención de nombres:**
- Formato: `NNN_description.sql`
- `NNN`: Número secuencial de 3 dígitos (001, 002, 003...)
- `description`: Descripción breve en snake_case
- Extensión: `.sql`

---

## 🔍 Tracking de Migraciones

### Tabla de Control

El sistema crea automáticamente una tabla `schema_migrations` para rastrear migraciones aplicadas:

```sql
CREATE TABLE schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT NOW()
);
```

**Ejemplo de datos:**
```sql
SELECT * FROM schema_migrations;
```
```
version                        | applied_at
-------------------------------|-------------------------
001_create_resume_tables       | 2025-11-30 10:00:00
002_add_indexes                | 2025-11-30 10:15:00
```

---

## 🚀 Uso

### Con Docker Compose (Recomendado)

```bash
# Levantar servicios (migraciones se ejecutan automáticamente)
docker-compose up -d

# Ver logs de migraciones
docker-compose logs backend

# Reconstruir y aplicar nuevas migraciones
docker-compose up -d --build
```

### Con Docker Manual

```bash
# Build
docker build -t resume-backend .

# Run (migraciones se ejecutan automáticamente)
docker run -d \
  --name resume-backend \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_USER=resume_user \
  -e DB_PASSWORD=resume_password \
  -e DB_NAME=resume_db \
  -p 8080:8080 \
  resume-backend
```

---

## ➕ Agregar Nueva Migración

### Paso 1: Crear Archivo de Migración

```bash
# Crear archivo con el siguiente número secuencial
touch migrations/002_add_user_timestamps.sql
```

### Paso 2: Escribir SQL

```sql
-- migrations/002_add_user_timestamps.sql

-- Agregar columnas de timestamp a resume_requests
ALTER TABLE resume_requests
ADD COLUMN IF NOT EXISTS last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Crear trigger para auto-actualizar last_updated
CREATE OR REPLACE FUNCTION update_last_updated_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_updated = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_resume_requests_last_updated
    BEFORE UPDATE ON resume_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_last_updated_column();
```

### Paso 3: Aplicar Migración

```bash
# Reconstruir contenedor
docker-compose up -d --build

# O si ya está corriendo, reiniciar
docker-compose restart backend
```

**El sistema:**
1. ✅ Detecta la nueva migración
2. ✅ Verifica que no esté aplicada
3. ✅ La ejecuta automáticamente
4. ✅ Registra en `schema_migrations`

---

## 🔧 Variables de Entorno Requeridas

El script `docker-entrypoint.sh` necesita estas variables:

```bash
DB_HOST=postgres          # Host de PostgreSQL
DB_PORT=5432             # Puerto (default: 5432)
DB_USER=resume_user      # Usuario de BD
DB_PASSWORD=resume_password  # Contraseña
DB_NAME=resume_db        # Nombre de la BD
```

Estas están definidas en:
- `docker-compose.yml` (para desarrollo)
- `.env` (para producción)

---

## 🛡️ Características de Seguridad

### 1. Idempotencia
✅ Las migraciones son idempotentes - pueden ejecutarse múltiples veces sin errores

```sql
-- Ejemplo de migración idempotente
CREATE TABLE IF NOT EXISTS resume_requests (...);
ALTER TABLE resume_requests ADD COLUMN IF NOT EXISTS new_field VARCHAR(255);
```

### 2. Tracking de Versiones
✅ El sistema previene re-ejecución de migraciones ya aplicadas

```bash
# Output de ejemplo
📊 Running database migrations...
   ⏭️  Migration 001_create_resume_tables.sql already applied
   Applying migration: 002_add_indexes.sql
   ✅ Migration 002_add_indexes.sql applied
```

### 3. Orden Garantizado
✅ Las migraciones se ejecutan en orden alfabético (por nombre de archivo)

---

## 🐛 Troubleshooting

### Problema: Migración falla

**Síntomas:**
```bash
ERROR: relation "resume_requests" already exists
```

**Solución:**
```sql
-- Hacer la migración idempotente
CREATE TABLE IF NOT EXISTS resume_requests (...);
```

---

### Problema: Migración no se aplica

**Verificar:**
```bash
# 1. Ver logs del contenedor
docker-compose logs backend

# 2. Verificar que el archivo está en el contenedor
docker exec resume-backend-service ls -la /app/migrations/

# 3. Verificar permisos
docker exec resume-backend-service ls -lh /app/migrations/
```

**Solución:**
```bash
# Reconstruir imagen
docker-compose build --no-cache backend
docker-compose up -d
```

---

### Problema: Conexión a BD falla

**Verificar:**
```bash
# 1. PostgreSQL está corriendo
docker-compose ps postgres

# 2. Variables de entorno correctas
docker exec resume-backend-service env | grep DB_

# 3. Logs de PostgreSQL
docker-compose logs postgres
```

---

## 🧪 Testing de Migraciones

### Test Manual

```bash
# 1. Bajar todo
docker-compose down -v  # ⚠️ Borra volúmenes

# 2. Levantar desde cero
docker-compose up -d

# 3. Verificar migraciones aplicadas
docker-compose exec postgres psql -U resume_user -d resume_db -c "SELECT * FROM schema_migrations;"

# 4. Verificar tablas creadas
docker-compose exec postgres psql -U resume_user -d resume_db -c "\dt"
```

### Test de Nueva Migración

```bash
# 1. Agregar nueva migración
echo "CREATE TABLE test_table (id SERIAL PRIMARY KEY);" > migrations/999_test.sql

# 2. Reconstruir y levantar
docker-compose up -d --build

# 3. Verificar que se aplicó
docker-compose exec postgres psql -U resume_user -d resume_db -c "\dt test_table"

# 4. Limpiar
rm migrations/999_test.sql
```

---

## 📝 Mejores Prácticas

### ✅ DO

1. **Usar IF EXISTS / IF NOT EXISTS**
   ```sql
   CREATE TABLE IF NOT EXISTS my_table (...);
   DROP TABLE IF EXISTS old_table;
   ```

2. **Una migración = una responsabilidad**
   ```
   ✅ 001_create_users_table.sql
   ✅ 002_create_posts_table.sql
   ❌ 001_create_all_tables.sql  (demasiado amplio)
   ```

3. **Describir bien las migraciones**
   ```
   ✅ 003_add_email_unique_constraint.sql
   ❌ 003_fix.sql
   ```

4. **Probar migraciones localmente primero**
   ```bash
   # Test en BD local
   psql -U resume_user -d resume_db -f migrations/003_new_migration.sql
   ```

### ❌ DON'T

1. **NO modificar migraciones ya aplicadas**
   ```
   ❌ Editar 001_create_resume_tables.sql después de aplicarla
   ✅ Crear 002_alter_resume_tables.sql
   ```

2. **NO usar transacciones complejas**
   ```sql
   ❌ BEGIN; ... múltiples operaciones ... COMMIT;
   ✅ Operaciones atómicas y simples
   ```

3. **NO hardcodear datos sensibles**
   ```sql
   ❌ INSERT INTO users VALUES ('admin', 'password123');
   ✅ Usar variables de entorno o secrets
   ```

---

## 🔄 Rollback de Migraciones

**Actualmente no implementado.**

Para rollback manual:

```bash
# 1. Conectar a la BD
docker-compose exec postgres psql -U resume_user -d resume_db

# 2. Ejecutar operaciones inversas manualmente
DROP TABLE resume_requests;

# 3. Remover de tracking
DELETE FROM schema_migrations WHERE version = '001_create_resume_tables';
```

**Futuro:** Implementar migraciones reversibles con archivos `up` y `down`.

---

## 📚 Archivos Relacionados

| Archivo | Propósito |
|---------|-----------|
| `docker-entrypoint.sh` | Script que ejecuta migraciones |
| `Dockerfile` | Copia migraciones al contenedor |
| `docker-compose.yml` | Define variables de entorno |
| `migrations/*.sql` | Archivos de migraciones |

---

## 🎯 Resumen

✅ **Automático**: Las migraciones se ejecutan al iniciar el contenedor
✅ **Seguro**: Previene re-ejecución con tracking
✅ **Simple**: Solo agregar archivo SQL numerado
✅ **Visible**: Logs claros del proceso
✅ **Idempotente**: Puede ejecutarse múltiples veces

---

**Fecha:** 2025-11-30
**Versión:** 1.0
