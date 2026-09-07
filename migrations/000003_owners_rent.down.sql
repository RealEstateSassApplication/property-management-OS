DROP TABLE IF EXISTS payment_allocations;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS rent_obligations;
DROP TABLE IF EXISTS ownership_interests;
DROP TABLE IF EXISTS owners;
ALTER TABLE leases DROP CONSTRAINT IF EXISTS leases_id_organization_unique;
ALTER TABLE properties DROP CONSTRAINT IF EXISTS properties_id_organization_unique;
