-- ============================================================================
-- MIGRATION 002: Add Soft Delete to Resume Versions
-- Descripción: Agregar estado para soft delete de versiones
-- Fecha: 2025-12-02
-- ============================================================================

-- Agregar columna de estado a resume_versions
ALTER TABLE resume_versions 
ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';

-- Agregar constraint de validación
ALTER TABLE resume_versions 
ADD CONSTRAINT valid_version_status CHECK (status IN ('active', 'deleted'));

-- Crear índice para optimizar queries por status
CREATE INDEX idx_resume_versions_status ON resume_versions(status);

-- Actualizar función create_resume_version para incluir status
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
    -- Obtener el siguiente número de versión (solo versiones activas)
    SELECT COALESCE(MAX(version_number), 0) + 1 
    INTO v_version_number
    FROM resume_versions 
    WHERE request_id = p_request_id AND status = 'active';
    
    -- Crear la nueva versión
    INSERT INTO resume_versions (
        request_id, 
        user_id, 
        version_number, 
        structured_data, 
        version_name, 
        created_by,
        status
    ) VALUES (
        p_request_id, 
        p_user_id, 
        v_version_number, 
        p_structured_data, 
        p_version_name, 
        p_created_by,
        'active'
    ) RETURNING id INTO v_version_id;
    
    -- Actualizar la referencia de versión activa en processed_resumes
    UPDATE processed_resumes 
    SET active_version_id = v_version_id, updated_at = CURRENT_TIMESTAMP
    WHERE request_id = p_request_id;
    
    RETURN v_version_id;
END;
$$ LANGUAGE plpgsql;

-- Crear función para soft delete de versiones
CREATE OR REPLACE FUNCTION soft_delete_resume_version(
    p_version_id BIGINT,
    p_user_id VARCHAR(255)
) RETURNS BOOLEAN AS $$
DECLARE
    v_request_id UUID;
    v_active_version_id BIGINT;
BEGIN
    -- Verificar que la versión existe y pertenece al usuario
    SELECT request_id INTO v_request_id
    FROM resume_versions 
    WHERE id = p_version_id AND user_id = p_user_id AND status = 'active';
    
    IF NOT FOUND THEN
        RETURN FALSE;
    END IF;
    
    -- Verificar si es la versión activa
    SELECT active_version_id INTO v_active_version_id
    FROM processed_resumes 
    WHERE request_id = v_request_id;
    
    -- No permitir eliminar la versión activa si es la única
    IF v_active_version_id = p_version_id THEN
        -- Contar versiones activas
        IF (SELECT COUNT(*) FROM resume_versions 
            WHERE request_id = v_request_id AND status = 'active') <= 1 THEN
            RETURN FALSE; -- No se puede eliminar la única versión
        END IF;
        
        -- Cambiar a otra versión activa (la más reciente)
        SELECT id INTO v_active_version_id
        FROM resume_versions 
        WHERE request_id = v_request_id AND status = 'active' AND id != p_version_id
        ORDER BY created_at DESC 
        LIMIT 1;
        
        UPDATE processed_resumes 
        SET active_version_id = v_active_version_id, updated_at = CURRENT_TIMESTAMP
        WHERE request_id = v_request_id;
    END IF;
    
    -- Marcar versión como eliminada
    UPDATE resume_versions 
    SET status = 'deleted'
    WHERE id = p_version_id;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;