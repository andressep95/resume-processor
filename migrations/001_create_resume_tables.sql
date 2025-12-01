-- ============================================================================
-- MIGRATION 001: Create Resume Tables
-- Descripción: Tablas principales para tracking y persistencia de CVs
-- Fecha: 2025-11-30
-- ============================================================================

-- Habilitar extensión para UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ----------------------------------------------------------------------------
-- TABLA: resume_requests
-- Propósito: Tracking de solicitudes de procesamiento de CVs
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS resume_requests (
    -- Identificador único de la solicitud (UUID v4)
    request_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Usuario que hizo la solicitud (desde JWT token)
    user_id VARCHAR(255) NOT NULL,
    user_email VARCHAR(255) NOT NULL,

    -- Información del archivo original
    original_filename VARCHAR(500) NOT NULL,
    original_file_type VARCHAR(10) NOT NULL,
    file_size_bytes BIGINT NOT NULL,

    -- Parámetros de la solicitud
    language VARCHAR(10) DEFAULT 'esp',
    instructions TEXT,

    -- URLs de S3
    s3_input_url TEXT,
    s3_output_url TEXT,

    -- Estado del procesamiento
    status VARCHAR(20) NOT NULL DEFAULT 'pending',

    -- Tiempo de procesamiento (en milisegundos)
    processing_time_ms BIGINT,

    -- Mensajes de error (si falla)
    error_message TEXT,

    -- Timestamps de auditoría
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    uploaded_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,

    -- Constraint de validación
    CONSTRAINT valid_status CHECK (status IN ('pending', 'uploaded', 'processing', 'completed', 'failed'))
);

-- Índices para optimizar queries
CREATE INDEX idx_resume_requests_user_id ON resume_requests(user_id);
CREATE INDEX idx_resume_requests_status ON resume_requests(status);
CREATE INDEX idx_resume_requests_created_at ON resume_requests(created_at DESC);

-- ----------------------------------------------------------------------------
-- TABLA: processed_resumes
-- Propósito: Almacenar datos estructurados extraídos del CV
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS processed_resumes (
    id BIGSERIAL PRIMARY KEY,

    -- Relación con la solicitud original
    request_id UUID NOT NULL UNIQUE REFERENCES resume_requests(request_id) ON DELETE CASCADE,

    -- Usuario propietario
    user_id VARCHAR(255) NOT NULL,

    -- Datos extraídos del CV (JSONB para flexibilidad)
    structured_data JSONB NOT NULL,

    -- Campos denormalizados para búsquedas rápidas
    cv_name VARCHAR(500),
    cv_email VARCHAR(255),
    cv_phone VARCHAR(100),

    -- Contadores
    education_count INT DEFAULT 0,
    experience_count INT DEFAULT 0,
    certifications_count INT DEFAULT 0,
    projects_count INT DEFAULT 0,
    skills_count INT DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Índices
CREATE INDEX idx_processed_resumes_user_id ON processed_resumes(user_id);
CREATE INDEX idx_processed_resumes_request_id ON processed_resumes(request_id);
CREATE INDEX idx_processed_resumes_structured_data ON processed_resumes USING GIN (structured_data);

-- ----------------------------------------------------------------------------
-- TRIGGER: Auto-actualizar updated_at
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_processed_resumes_updated_at
    BEFORE UPDATE ON processed_resumes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
