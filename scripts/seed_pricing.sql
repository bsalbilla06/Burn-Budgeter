-- Initial pricing data for AWS, GCP, and AI services
-- Updated with 2026 data extracted from provider documentation
-- Gemini models updated with latest 2.5 and 3.x series data

TRUNCATE TABLE public.services RESTART IDENTITY CASCADE;

INSERT INTO public.services (provider, name, unit, price_per_unit) VALUES
-- AWS Compute
('AWS', 'EC2 t3.micro', 'hours', 0.010400),
('AWS', 'EC2 t3.small', 'hours', 0.020800),
('AWS', 'EC2 t3.medium', 'hours', 0.041600),
('AWS', 'EC2 m5.large', 'hours', 0.096000),
('AWS', 'EC2 m5.xlarge', 'hours', 0.192000),

-- AWS Storage
('AWS', 'S3 Standard Storage', 'GB-month', 0.023000),
('AWS', 'S3 PUT/COPY/POST/LIST', '1000 requests', 0.005000),
('AWS', 'S3 GET/SELECT/OTHER', '1000 requests', 0.000400),
('AWS', 'EBS gp3 Storage', 'GB-month', 0.080000),

-- AWS Database
('AWS', 'RDS PostgreSQL db.t3.micro', 'hours', 0.017000),
('AWS', 'RDS PostgreSQL db.t3.small', 'hours', 0.034000),
('AWS', 'RDS MySQL db.t3.micro', 'hours', 0.017000),
('AWS', 'RDS Storage (gp3)', 'GB-month', 0.115000),

-- GCP Compute
('GCP', 'Compute e2-micro', 'hours', 0.008400),
('GCP', 'Compute e2-small', 'hours', 0.016800),
('GCP', 'Compute e2-medium', 'hours', 0.033500),
('GCP', 'Compute n2-standard-2', 'hours', 0.106800),

-- GCP Storage
('GCP', 'GCS Standard Storage', 'GB-month', 0.020000),
('GCP', 'GCS Class A Operations', '1000 requests', 0.005000),
('GCP', 'GCS Class B Operations', '1000 requests', 0.000400),
('GCP', 'Persistent Disk SSD', 'GB-month', 0.170000),

-- GCP Database
('GCP', 'Cloud SQL PostgreSQL db-f1-micro', 'hours', 0.010500),
('GCP', 'Cloud SQL PostgreSQL db-g1-small', 'hours', 0.021000),
('GCP', 'Cloud SQL Storage (SSD)', 'GB-month', 0.170000),

-- OpenAI (Updated to GPT-5.2 and 4.1 series)
('OpenAI', 'GPT-5.2 Input', '1M tokens', 1.750000),
('OpenAI', 'GPT-5.2 Output', '1M tokens', 14.000000),
('OpenAI', 'GPT-5.2 Pro Input', '1M tokens', 21.000000),
('OpenAI', 'GPT-5.2 Pro Output', '1M tokens', 168.000000),
('OpenAI', 'GPT-5 Mini Input', '1M tokens', 0.250000),
('OpenAI', 'GPT-5 Mini Output', '1M tokens', 2.000000),
('OpenAI', 'GPT-4.1 Input', '1M tokens', 3.000000),
('OpenAI', 'GPT-4.1 Output', '1M tokens', 12.000000),
('OpenAI', 'GPT-4.1 Mini Input', '1M tokens', 0.800000),
('OpenAI', 'GPT-4.1 Mini Output', '1M tokens', 3.200000),

-- Anthropic
('Anthropic', 'Claude 4.6 Opus Input', '1M tokens', 5.000000),
('Anthropic', 'Claude 4.6 Opus Output', '1M tokens', 25.000000),
('Anthropic', 'Claude 4.6 Sonnet Input', '1M tokens', 3.000000),
('Anthropic', 'Claude 4.6 Sonnet Output', '1M tokens', 15.000000),
('Anthropic', 'Claude 3.5 Haiku Input', '1M tokens', 0.800000),
('Anthropic', 'Claude 3.5 Haiku Output', '1M tokens', 4.000000),

-- Google Gemini (Latest 2.5 and 3.x series)
('Gemini', 'Gemini 3.1 Pro Input (<=200k)', '1M tokens', 2.000000),
('Gemini', 'Gemini 3.1 Pro Output (<=200k)', '1M tokens', 12.000000),
('Gemini', 'Gemini 3.1 Flash Image Input', '1M tokens', 0.250000),
('Gemini', 'Gemini 3.1 Flash Image Output (Text)', '1M tokens', 1.500000),
('Gemini', 'Gemini 3 Flash Input', '1M tokens', 0.500000),
('Gemini', 'Gemini 3 Flash Output', '1M tokens', 3.000000),
('Gemini', 'Gemini 2.5 Pro Input (<=200k)', '1M tokens', 1.250000),
('Gemini', 'Gemini 2.5 Pro Output (<=200k)', '1M tokens', 10.000000),
('Gemini', 'Gemini 2.5 Flash Input', '1M tokens', 0.300000),
('Gemini', 'Gemini 2.5 Flash Output', '1M tokens', 2.500000),
('Gemini', 'Gemini 2.5 Flash-Lite Input', '1M tokens', 0.100000),
('Gemini', 'Gemini 2.5 Flash-Lite Output', '1M tokens', 0.400000);
