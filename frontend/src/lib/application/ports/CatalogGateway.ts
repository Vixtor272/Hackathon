import type { Pharmacy, Zone } from '../../domain';

/** Driven port: read-only catalog data (zones and pharmacies). */
export interface CatalogGateway {
  listZones(): Promise<Zone[]>;
  listPharmacies(zoneId: string): Promise<Pharmacy[]>;
}
