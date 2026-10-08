import { mount } from 'svelte';
import './app.css';
import App from './App.svelte';
import { startRouter } from './lib/router';

startRouter();

const target = document.getElementById('app');
if (!target) {
  throw new Error('No se encontró el contenedor #app');
}

export default mount(App, { target });
