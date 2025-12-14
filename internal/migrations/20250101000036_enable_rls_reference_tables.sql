-- Migration: Enable RLS on Reference Tables
-- Description: Re-enable Row Level Security on shared reference tables
-- Created: 2025-01-01
-- Issue: Tables have RLS policies but RLS is not enabled

-- =====================================================
-- ENABLE RLS ON REFERENCE TABLES
-- =====================================================

ALTER TABLE public.currencies ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.industries ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.states ENABLE ROW LEVEL SECURITY;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE public.currencies IS 'Reference table for currencies - RLS enabled for consistency';
COMMENT ON TABLE public.industries IS 'Reference table for industries - RLS enabled for consistency';
COMMENT ON TABLE public.states IS 'Reference table for states/provinces - RLS enabled for consistency';
