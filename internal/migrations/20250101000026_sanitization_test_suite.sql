-- Migration: Sanitization Testing Suite
-- Description: Comprehensive tests to ensure PII sanitization works correctly
-- Created: 2025-01-01

-- ============================================
-- TEST DATA SAMPLES
-- ============================================

CREATE TABLE IF NOT EXISTS pii_test_cases (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    test_name text NOT NULL,
    test_category text NOT NULL,
    input_text text NOT NULL,
    expected_pii_types text[],
    should_sanitize boolean DEFAULT true,
    notes text,
    created_at timestamptz DEFAULT now()
);

-- Insert comprehensive test cases
INSERT INTO pii_test_cases (test_name, test_category, input_text, expected_pii_types, should_sanitize, notes) VALUES
-- Email tests
('Basic email', 'email', 'Contact us at support@company.com for help', ARRAY['email'], true, 'Simple email detection'),
('Multiple emails', 'email', 'Send to john@acme.com and jane@acme.com', ARRAY['email'], true, 'Multiple email addresses'),
('Email in sentence', 'email', 'Please reach out to customer.service@tech-solutions.co.uk if needed', ARRAY['email'], true, 'Complex domain'),

-- Phone tests
('US phone', 'phone', 'Call us at +1-555-0100', ARRAY['phone'], true, 'US format with country code'),
('Phone with parentheses', 'phone', 'Contact (555) 123-4567', ARRAY['phone'], true, 'Common US format'),
('International phone', 'phone', 'Call +44 20 7946 0958', ARRAY['phone'], true, 'UK format'),

-- SSN tests
('SSN', 'ssn', 'SSN: 123-45-6789', ARRAY['ssn'], true, 'Social Security Number'),

-- Credit card tests
('Credit card', 'credit_card', 'Card: 4111-1111-1111-1111', ARRAY['credit_card'], true, 'Credit card number'),

-- Combined PII
('Invoice with PII', 'combined',
 'Invoice for John Doe (john@acme.com), phone +1-555-0100, amount $5,432.10',
 ARRAY['email', 'phone', 'amount', 'name'], true,
 'Realistic invoice text'),

('Customer record', 'combined',
 'Customer: Mr. Robert Smith, Email: r.smith@example.com, Phone: (555) 987-6543, SSN: 987-65-4321',
 ARRAY['name', 'email', 'phone', 'ssn'], true,
 'Full customer profile'),

-- Safe content (no PII)
('Product description', 'safe',
 'Our premium server rack offers 42U of space with excellent cable management',
 ARRAY[]::text[], false,
 'Should NOT sanitize'),

('General query', 'safe',
 'How do I create a sales order in the system?',
 ARRAY[]::text[], false,
 'General help question'),

-- Edge cases
('Email-like but not email', 'edge_case',
 'Use the format name@domain when configuring',
 ARRAY[]::text[], false,
 'Instruction text, not actual email'),

('Amount without PII', 'edge_case',
 'The product costs $299.99 and includes free shipping',
 ARRAY['amount'], true,
 'Amount is not always PII, but we sanitize for consistency');


-- ============================================
-- TEST FUNCTIONS
-- ============================================

-- Function to run all sanitization tests
CREATE OR REPLACE FUNCTION test_sanitization_accuracy()
RETURNS TABLE (
    test_name text,
    test_category text,
    input_text text,
    passed boolean,
    expected_sanitize boolean,
    actually_sanitized boolean,
    pii_detected jsonb,
    sanitized_output text,
    error_message text
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_test_case record;
    v_sanitization_result jsonb;
    v_pii_analysis jsonb;
    v_had_pii boolean;
BEGIN
    FOR v_test_case IN SELECT * FROM pii_test_cases ORDER BY test_category, id LOOP
        BEGIN
            -- Run sanitization
            v_sanitization_result := sanitize_for_groq(v_test_case.input_text, true);
            v_pii_analysis := detect_pii_advanced(v_test_case.input_text);
            v_had_pii := (v_sanitization_result->>'had_pii')::boolean;

            -- Check if result matches expectation
            test_name := v_test_case.test_name;
            test_category := v_test_case.test_category;
            input_text := v_test_case.input_text;
            expected_sanitize := v_test_case.should_sanitize;
            actually_sanitized := v_had_pii;
            pii_detected := v_pii_analysis;
            sanitized_output := v_sanitization_result->>'sanitized_text';
            passed := (expected_sanitize = v_had_pii);
            error_message := NULL;

            RETURN NEXT;
        EXCEPTION WHEN OTHERS THEN
            test_name := v_test_case.test_name;
            test_category := v_test_case.test_category;
            input_text := v_test_case.input_text;
            expected_sanitize := v_test_case.should_sanitize;
            actually_sanitized := NULL;
            pii_detected := NULL;
            sanitized_output := NULL;
            passed := false;
            error_message := SQLERRM;
            RETURN NEXT;
        END;
    END LOOP;
END;
$$;

COMMENT ON FUNCTION test_sanitization_accuracy IS
'Run all PII sanitization tests and return pass/fail results';


-- Function to test rehydration accuracy
CREATE OR REPLACE FUNCTION test_rehydration_accuracy()
RETURNS TABLE (
    test_name text,
    original_text text,
    sanitized_text text,
    rehydrated_text text,
    matches_original boolean,
    session_id text
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_test_case record;
    v_sanitization_result jsonb;
    v_session_id text;
    v_sanitized text;
    v_rehydrated text;
BEGIN
    FOR v_test_case IN
        SELECT * FROM pii_test_cases
        WHERE should_sanitize = true
        LIMIT 10
    LOOP
        -- Sanitize
        v_sanitization_result := sanitize_for_groq(v_test_case.input_text, true);
        v_session_id := v_sanitization_result->>'session_id';
        v_sanitized := v_sanitization_result->>'sanitized_text';

        -- Rehydrate
        v_rehydrated := rehydrate_pii(v_sanitized, v_session_id);

        -- Return results
        test_name := v_test_case.test_name;
        original_text := v_test_case.input_text;
        sanitized_text := v_sanitized;
        rehydrated_text := v_rehydrated;
        matches_original := (v_rehydrated = v_test_case.input_text);
        session_id := v_session_id;

        RETURN NEXT;
    END LOOP;
END;
$$;

COMMENT ON FUNCTION test_rehydration_accuracy IS
'Test that rehydration correctly restores original text';


-- Function to benchmark sanitization performance
CREATE OR REPLACE FUNCTION benchmark_sanitization(
    p_iterations int DEFAULT 100
)
RETURNS TABLE (
    test_type text,
    iterations int,
    total_time_ms bigint,
    avg_time_ms numeric,
    min_time_ms bigint,
    max_time_ms bigint
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_start_time timestamptz;
    v_end_time timestamptz;
    v_times bigint[];
    v_time_ms bigint;
    i int;
BEGIN
    -- Test simple sanitization
    v_times := ARRAY[]::bigint[];
    FOR i IN 1..p_iterations LOOP
        v_start_time := clock_timestamp();
        PERFORM sanitize_for_groq('Contact john@acme.com at +1-555-0100');
        v_end_time := clock_timestamp();
        v_time_ms := EXTRACT(MILLISECONDS FROM (v_end_time - v_start_time))::bigint;
        v_times := array_append(v_times, v_time_ms);
    END LOOP;

    test_type := 'Simple sanitization';
    iterations := p_iterations;
    total_time_ms := (SELECT SUM(t) FROM unnest(v_times) t);
    avg_time_ms := (SELECT AVG(t) FROM unnest(v_times) t);
    min_time_ms := (SELECT MIN(t) FROM unnest(v_times) t);
    max_time_ms := (SELECT MAX(t) FROM unnest(v_times) t);
    RETURN NEXT;

    -- Test complex sanitization
    v_times := ARRAY[]::bigint[];
    FOR i IN 1..p_iterations LOOP
        v_start_time := clock_timestamp();
        PERFORM sanitize_for_groq('Invoice for Mr. John Doe, email john@acme.com, phone +1-555-0100, SSN 123-45-6789, card 4111-1111-1111-1111, amount $5,432.10');
        v_end_time := clock_timestamp();
        v_time_ms := EXTRACT(MILLISECONDS FROM (v_end_time - v_start_time))::bigint;
        v_times := array_append(v_times, v_time_ms);
    END LOOP;

    test_type := 'Complex sanitization (multiple PII)';
    iterations := p_iterations;
    total_time_ms := (SELECT SUM(t) FROM unnest(v_times) t);
    avg_time_ms := (SELECT AVG(t) FROM unnest(v_times) t);
    min_time_ms := (SELECT MIN(t) FROM unnest(v_times) t);
    max_time_ms := (SELECT MAX(t) FROM unnest(v_times) t);
    RETURN NEXT;
END;
$$;

COMMENT ON FUNCTION benchmark_sanitization IS
'Benchmark sanitization performance with multiple iterations';


-- ============================================
-- VALIDATION REPORTS
-- ============================================

-- Create view for test results summary
CREATE OR REPLACE VIEW sanitization_test_summary AS
SELECT
    test_category,
    COUNT(*) as total_tests,
    COUNT(*) FILTER (WHERE passed) as passed_tests,
    COUNT(*) FILTER (WHERE NOT passed) as failed_tests,
    ROUND(100.0 * COUNT(*) FILTER (WHERE passed) / COUNT(*), 2) as pass_rate
FROM test_sanitization_accuracy()
GROUP BY test_category
ORDER BY pass_rate DESC;

COMMENT ON VIEW sanitization_test_summary IS
'Summary of sanitization test results by category';


-- ============================================
-- GRANT PERMISSIONS
-- ============================================

GRANT SELECT ON pii_test_cases TO authenticated;
GRANT EXECUTE ON FUNCTION test_sanitization_accuracy() TO authenticated;
GRANT EXECUTE ON FUNCTION test_rehydration_accuracy() TO authenticated;
GRANT EXECUTE ON FUNCTION benchmark_sanitization(int) TO authenticated;
GRANT SELECT ON sanitization_test_summary TO authenticated;


-- ============================================
-- RUN TESTS ON MIGRATION
-- ============================================

-- Run tests and display results
DO $$
DECLARE
    v_test_results record;
    v_failed_count int := 0;
    v_total_count int := 0;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'SANITIZATION TEST RESULTS';
    RAISE NOTICE '========================================';

    FOR v_test_results IN SELECT * FROM test_sanitization_accuracy() LOOP
        v_total_count := v_total_count + 1;

        IF NOT v_test_results.passed THEN
            v_failed_count := v_failed_count + 1;
            RAISE NOTICE 'FAILED: % (Category: %)', v_test_results.test_name, v_test_results.test_category;
            RAISE NOTICE '  Input: %', v_test_results.input_text;
            RAISE NOTICE '  Expected sanitize: %, Actually: %', v_test_results.expected_sanitize, v_test_results.actually_sanitized;
            IF v_test_results.error_message IS NOT NULL THEN
                RAISE NOTICE '  Error: %', v_test_results.error_message;
            END IF;
        END IF;
    END LOOP;

    RAISE NOTICE '========================================';
    RAISE NOTICE 'RESULTS: % passed, % failed out of % total tests',
        v_total_count - v_failed_count, v_failed_count, v_total_count;

    IF v_failed_count = 0 THEN
        RAISE NOTICE '✅ ALL TESTS PASSED!';
    ELSE
        RAISE WARNING '⚠️  SOME TESTS FAILED - Review sanitization logic';
    END IF;
    RAISE NOTICE '========================================';
END;
$$;


-- ============================================
-- EXAMPLE USAGE
-- ============================================

/*
-- 1. Run all sanitization tests
SELECT * FROM test_sanitization_accuracy();

-- 2. View test summary
SELECT * FROM sanitization_test_summary;

-- 3. Test rehydration accuracy
SELECT * FROM test_rehydration_accuracy();

-- 4. Benchmark performance
SELECT * FROM benchmark_sanitization(50);

-- 5. Check specific test case
SELECT * FROM test_sanitization_accuracy()
WHERE test_name = 'Invoice with PII';

-- 6. Add custom test case
INSERT INTO pii_test_cases (test_name, test_category, input_text, expected_pii_types, should_sanitize)
VALUES ('My custom test', 'custom', 'Text with my@email.com', ARRAY['email'], true);

-- Then re-run tests
SELECT * FROM test_sanitization_accuracy() WHERE test_category = 'custom';
*/
