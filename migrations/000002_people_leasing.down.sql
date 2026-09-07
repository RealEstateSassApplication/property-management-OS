DROP TABLE IF EXISTS leases;
DROP TABLE IF EXISTS tenancy_tenants;
DROP TABLE IF EXISTS tenancies;
DROP TABLE IF EXISTS tenants;
ALTER TABLE units DROP CONSTRAINT IF EXISTS units_id_organization_unique;
