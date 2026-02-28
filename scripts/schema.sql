-- Initial schema for Burn Budgeter
-- Designed for Supabase (PostgreSQL)

-- 1. Services Table (Master list of provider pricing)
CREATE TABLE IF NOT EXISTS public.services (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    price_per_unit DECIMAL(12, 6) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Projects Table
CREATE TABLE IF NOT EXISTS public.projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    cash_on_hand DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Project Services Table (Junction table for stacks)
CREATE TABLE IF NOT EXISTS public.project_services (
    id SERIAL PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES public.projects(id) ON DELETE CASCADE,
    service_id INT NOT NULL REFERENCES public.services(id) ON DELETE RESTRICT,
    quantity DECIMAL(15, 4) NOT NULL DEFAULT 1.00,
    is_optimized BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, service_id)
);

-- Enable Row Level Security (RLS)
ALTER TABLE public.projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.project_services ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.services ENABLE ROW LEVEL SECURITY;

-- Policies for 'services' (Read-only for all authenticated users)
CREATE POLICY "Allow read access to all authenticated users" 
ON public.services FOR SELECT 
TO authenticated 
USING (true);

-- Policies for 'projects' (Owner-only access)
CREATE POLICY "Users can view their own projects" 
ON public.projects FOR SELECT 
TO authenticated 
USING (auth.uid() = user_id);

CREATE POLICY "Users can create their own projects" 
ON public.projects FOR INSERT 
TO authenticated 
WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update their own projects" 
ON public.projects FOR UPDATE 
TO authenticated 
USING (auth.uid() = user_id);

CREATE POLICY "Users can delete their own projects" 
ON public.projects FOR DELETE 
TO authenticated 
USING (auth.uid() = user_id);

-- Policies for 'project_services' (Based on project ownership)
CREATE POLICY "Users can view services in their projects" 
ON public.project_services FOR SELECT 
TO authenticated 
USING (
    EXISTS (
        SELECT 1 FROM public.projects 
        WHERE projects.id = project_services.project_id 
        AND projects.user_id = auth.uid()
    )
);

CREATE POLICY "Users can add services to their projects" 
ON public.project_services FOR INSERT 
TO authenticated 
WITH CHECK (
    EXISTS (
        SELECT 1 FROM public.projects 
        WHERE projects.id = project_services.project_id 
        AND projects.user_id = auth.uid()
    )
);

CREATE POLICY "Users can update services in their projects" 
ON public.project_services FOR UPDATE 
TO authenticated 
USING (
    EXISTS (
        SELECT 1 FROM public.projects 
        WHERE projects.id = project_services.project_id 
        AND projects.user_id = auth.uid()
    )
);

CREATE POLICY "Users can remove services from their projects" 
ON public.project_services FOR DELETE 
TO authenticated 
USING (
    EXISTS (
        SELECT 1 FROM public.projects 
        WHERE projects.id = project_services.project_id 
        AND projects.user_id = auth.uid()
    )
);
