-- Migration: Remove user_email column from resume_requests
-- Description: Simplify data model by removing redundant user_email field
-- Only user_id (UUID from auth service) is needed for user identification

-- Remove user_email column from resume_requests table
ALTER TABLE resume_requests DROP COLUMN IF EXISTS user_email;
