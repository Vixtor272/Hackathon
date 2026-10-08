import type { Pharmacy, Zone } from '../../domain';
import type { CatalogGateway } from '../../application/ports';
import { clone } from './fixtures';

const ZONES: Zone[] = [
  { id: 'uio-norte', label: 'Quito — zona norte' },
  { id: 'uio-centro', label: 'Quito — zona centro' },
];

const PHARMACIES: Pharmacy[] = [
  { id: 'med-norte', name: 'Medicity Demo Norte', chain: 'Medicity', address: 'Av. Ficticia A 123', zone: 'uio-norte' },
  { id: 'eco-norte', name: 'Económicas Demo Norte', chain: 'Farmacias Económicas', address: 'Av. Ficticia D 456', zone: 'uio-norte' },
  { id: 'eco-centro', name: 'Económicas Demo Centro', chain: 'Farmacias Económicas', address: 'Dirección ficticia B', zone: 'uio-centro' },
];

export class InMemoryCatalogGateway implements CatalogGateway {
  async listZones(): Promise<Zone[]> {
    return clone(ZONES);
  }

  async listPharmacies(zoneId: string): Promise<Pharmacy[]> {
    return clone(PHARMACIES.filter((pharmacy) => pharmacy.zone === zoneId));
  }
}
