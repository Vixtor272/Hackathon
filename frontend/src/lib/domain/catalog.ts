export interface Zone {
  id: string;
  label: string;
}

export type PharmacyChain = 'Medicity' | 'Farmacias Económicas' | string;

export interface Pharmacy {
  id: string;
  name: string;
  chain: PharmacyChain;
  address: string;
  zone: string;
}
