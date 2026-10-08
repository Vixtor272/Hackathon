import type { Pharmacy, Zone } from '../../domain';
import type { CatalogGateway } from '../../application/ports';
import type { HttpClient } from './httpClient';

export class HttpCatalogGateway implements CatalogGateway {
  constructor(private readonly http: HttpClient) {}

  async listZones(): Promise<Zone[]> {
    const { zones } = await this.http.get<{ zones: Zone[] }>('/catalog/zones');
    return zones;
  }

  async listPharmacies(zoneId: string): Promise<Pharmacy[]> {
    const { pharmacies } = await this.http.get<{ pharmacies: Pharmacy[] }>('/catalog/pharmacies', { zone: zoneId });
    return pharmacies;
  }
}
