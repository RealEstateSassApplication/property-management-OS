INSERT INTO rent_adjustments (
    id, organization_id, obligation_id, adjustment_type, amount_minor, reason, created_by_user_id
) VALUES
(
    '21212121-2121-2121-2121-212121212121',
    '11111111-1111-1111-1111-111111111111',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'late_fee', 500000, 'Seeded late fee', '22222222-2222-2222-2222-222222222222'
),
(
    '22212121-2121-2121-2121-212121212121',
    '11111111-1111-1111-1111-111111111111',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'credit', 500000, 'Seeded matching goodwill credit', '22222222-2222-2222-2222-222222222222'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO payments (
    id, organization_id, tenant_id, amount_minor, currency, received_at, method, reference_code, status
) VALUES (
    '23232323-2323-2323-2323-232323232323',
    '11111111-1111-1111-1111-111111111111',
    '66666666-6666-6666-6666-666666666666',
    1000000, 'LKR', '2026-09-06', 'bank_transfer', 'DEV-PAY-REVERSAL', 'void'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO payment_allocations (
    id, organization_id, payment_id, obligation_id, amount_minor
) VALUES (
    '24242424-2424-2424-2424-242424242424',
    '11111111-1111-1111-1111-111111111111',
    '23232323-2323-2323-2323-232323232323',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    1000000
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO payment_reversals (
    id, organization_id, payment_id, reason, reversed_by_user_id, reversed_at
) VALUES (
    '25252525-2525-2525-2525-252525252525',
    '11111111-1111-1111-1111-111111111111',
    '23232323-2323-2323-2323-232323232323',
    'Seeded duplicate transfer reversal',
    '22222222-2222-2222-2222-222222222222',
    '2026-09-07T07:00:00Z'
)
ON CONFLICT (organization_id, payment_id) DO NOTHING;

INSERT INTO rent_adjustments (
    id, organization_id, obligation_id, adjustment_type, amount_minor, reason, created_by_user_id
) VALUES (
    '26262626-2626-2626-2626-262626262626',
    '11111111-1111-1111-1111-111111111111',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'payment_reversal', 1000000, 'Payment reversal: Seeded duplicate transfer reversal',
    '22222222-2222-2222-2222-222222222222'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO security_deposit_accounts (
    id, organization_id, lease_id, required_amount_minor, currency
) VALUES (
    '27272727-2727-2727-2727-272727272727',
    '11111111-1111-1111-1111-111111111111',
    '88888888-8888-8888-8888-888888888888',
    30000000, 'LKR'
)
ON CONFLICT (organization_id, lease_id) DO NOTHING;

INSERT INTO security_deposit_transactions (
    id, organization_id, deposit_account_id, transaction_type, amount_minor, occurred_on, note, created_by_user_id
) VALUES
(
    '28282828-2828-2828-2828-282828282828',
    '11111111-1111-1111-1111-111111111111',
    '27272727-2727-2727-2727-272727272727',
    'received', 20000000, '2026-01-01', 'Initial deposit received',
    '22222222-2222-2222-2222-222222222222'
),
(
    '29292929-2929-2929-2929-292929292929',
    '11111111-1111-1111-1111-111111111111',
    '27272727-2727-2727-2727-272727272727',
    'deduction', 5000000, '2026-09-07', 'Approved damage deduction',
    '22222222-2222-2222-2222-222222222222'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO property_expenses (
    id, organization_id, property_id, vendor_id, work_order_id, category,
    amount_minor, currency, incurred_on, note, reference_code, created_by_user_id
) VALUES
(
    '30303030-3030-3030-3030-303030303030',
    '11111111-1111-1111-1111-111111111111',
    '33333333-3333-3333-3333-333333333333',
    'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
    '12121212-1212-1212-1212-121212121212',
    'maintenance', 1850000, 'LKR', '2026-09-07',
    'Completed kitchen sink repair', 'EXP-DEV-001',
    '22222222-2222-2222-2222-222222222222'
),
(
    '31313131-3131-3131-3131-313131313131',
    '11111111-1111-1111-1111-111111111111',
    '33333333-3333-3333-3333-333333333333',
    NULL, NULL, 'other', 500000, 'LKR', '2026-09-05',
    'Duplicate administrative expense', 'EXP-DEV-REV',
    '22222222-2222-2222-2222-222222222222'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO property_expense_reversals (
    id, organization_id, expense_id, reason, reversed_by_user_id, reversed_at
) VALUES (
    '32323232-3232-3232-3232-323232323232',
    '11111111-1111-1111-1111-111111111111',
    '31313131-3131-3131-3131-313131313131',
    'Duplicate expense entered during reconciliation',
    '22222222-2222-2222-2222-222222222222',
    '2026-09-07T07:15:00Z'
)
ON CONFLICT (organization_id, expense_id) DO NOTHING;
