// Company back office: cashier and courier board plus the notifications
// outbox. It is a separate app on its own port and talks only to the
// company surface of the API (see backend/cmd/server).
import { mount } from 'svelte';
import '../src/app.css';
import EmpresaApp from '../src/EmpresaApp.svelte';

const target = document.getElementById('app');
if (!target) {
  throw new Error('No se encontró el contenedor #app');
}

export default mount(EmpresaApp, { target });
