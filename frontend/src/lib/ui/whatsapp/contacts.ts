export interface DemoContact {
  phone: string;
  name: string;
  idNumber: string;
}

/** Clients seeded in the backend (see docs/API.md → Demo data). */
export const DEMO_CONTACTS: DemoContact[] = [
  { phone: '+593991111111', name: 'María Pérez', idNumber: '1712345678' },
  { phone: '+593992222222', name: 'Juan López', idNumber: '0912345678' },
];
