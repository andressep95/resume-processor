-- ============================================================================
-- MIGRATION 003: Add Resume Versioning System
-- Descripción: Sistema de versionado para CVs procesados
-- Fecha: 2025-12-01
-- ============================================================================

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
-- Actualizar tabla processed_resumes
-- Propósito: Simplificar y agregar referencia a versión activa
-- ----------------------------------------------------------------------------

-- Agregar referencia a la versión activa
ALTER TABLE processed_resumes 
ADD COLUMN IF NOT EXISTS active_version_id BIGINT REFERENCES resume_versions(id);

-- Eliminar campos denormalizados innecesarios
ALTER TABLE processed_resumes DROP COLUMN IF EXISTS cv_name;
ALTER TABLE processed_resumes DROP COLUMN IF EXISTS cv_email;
ALTER TABLE processed_resumes DROP COLUMN IF EXISTS cv_phone;
ALTER TABLE processed_resumes DROP COLUMN IF EXISTS education_count;
ALTER TABLE processed_resumes DROP COLUMN IF EXISTS experience_count;
ALTER TABLE processed_resumes DROP COLUMN IF EXISTS certifications_count;
ALTER TABLE processed_resumes DROP COLUMN IF EXISTS projects_count;
ALTER TABLE processed_resumes DROP COLUMN IF EXISTS skills_count;

-- Eliminar structured_data de processed_resumes (ahora está en resume_versions)
ALTER TABLE processed_resumes DROP COLUMN IF EXISTS structured_data;

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