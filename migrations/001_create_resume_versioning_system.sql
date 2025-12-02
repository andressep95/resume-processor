-- ============================================================================
-- MIGRATION 001: Create Resume Versioning System (Clean)
-- Descripción: Sistema completo de versionado de CVs desde el inicio
-- Fecha: 2025-12-02
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
-- TABLA: resume_versions
-- Propósito: Almacenar todas las versiones de un CV (historial completo)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS resume_versions (
    id BIGSERIAL PRIMARY KEY,
    
    -- Relación con la solicitud original
    request_id UUID NOT NULL REFERENCES resume_requests(request_id) ON DELETE CASCADE,
    
    -- Usuario propietario
    user_id VARCHAR(255) NOT NULL,
    
    -- Versionado
    version_number INT NOT NULL DEFAULT 1,
    
    -- Datos del CV en esta versión
    structured_data JSONB NOT NULL,
    
    -- Metadatos de la versión
    version_name VARCHAR(255), -- Ej: "Versión inicial", "Actualización skills", etc.
    created_by VARCHAR(50) NOT NULL DEFAULT 'system', -- 'system' o 'user'
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Índices únicos
    UNIQUE(request_id, version_number)
);

-- Índices para optimizar queries
CREATE INDEX idx_resume_versions_request_id ON resume_versions(request_id);
CREATE INDEX idx_resume_versions_user_id ON resume_versions(user_id);
CREATE INDEX idx_resume_versions_created_at ON resume_versions(created_at DESC);

-- ----------------------------------------------------------------------------
-- TABLA: processed_resumes
-- Propósito: Referencia a la versión activa de cada CV
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS processed_resumes (
    id BIGSERIAL PRIMARY KEY,

    -- Relación con la solicitud original
    request_id UUID NOT NULL UNIQUE REFERENCES resume_requests(request_id) ON DELETE CASCADE,

    -- Usuario propietario
    user_id VARCHAR(255) NOT NULL,

    -- Referencia a la versión activa
    active_version_id BIGINT REFERENCES resume_versions(id),

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Índices
CREATE INDEX idx_processed_resumes_user_id ON processed_resumes(user_id);
CREATE INDEX idx_processed_resumes_request_id ON processed_resumes(request_id);
CREATE INDEX idx_processed_resumes_active_version ON processed_resumes(active_version_id);

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

-- ----------------------------------------------------------------------------
-- FUNCIÓN: Crear nueva versión de CV
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION create_resume_version(
    p_request_id UUID,
    p_user_id VARCHAR(255),
    p_structured_data JSONB,
    p_version_name VARCHAR(255) DEFAULT NULL,
    p_created_by VARCHAR(50) DEFAULT 'user'
) RETURNS BIGINT AS $$
DECLARE
    v_version_number INT;
    v_version_id BIGINT;
BEGIN
    -- Obtener el siguiente número de versión
    SELECT COALESCE(MAX(version_number), 0) + 1 
    INTO v_version_number
    FROM resume_versions 
    WHERE request_id = p_request_id;
    
    -- Crear la nueva versión
    INSERT INTO resume_versions (
        request_id, 
        user_id, 
        version_number, 
        structured_data, 
        version_name, 
        created_by
    ) VALUES (
        p_request_id, 
        p_user_id, 
        v_version_number, 
        p_structured_data, 
        p_version_name, 
        p_created_by
    ) RETURNING id INTO v_version_id;
    
    -- Actualizar la referencia de versión activa en processed_resumes
    UPDATE processed_resumes 
    SET active_version_id = v_version_id, updated_at = CURRENT_TIMESTAMP
    WHERE request_id = p_request_id;
    
    RETURN v_version_id;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- FUNCIÓN: Activar una versión específica
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION activate_resume_version(
    p_request_id UUID,
    p_version_id BIGINT
) RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar que la versión existe y pertenece al request_id
    IF NOT EXISTS (
        SELECT 1 FROM resume_versions 
        WHERE id = p_version_id AND request_id = p_request_id
    ) THEN
        RETURN FALSE;
    END IF;
    
    -- Actualizar la versión activa
    UPDATE processed_resumes 
    SET active_version_id = p_version_id, updated_at = CURRENT_TIMESTAMP
    WHERE request_id = p_request_id;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;