import './styles.css';

import { createPinia } from 'pinia';
import { createApp } from 'vue';

import App from './App.vue';
import { apiEvents } from './api/http';
import { router } from './router';
import { useSessionStore } from './stores/session';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);

apiEvents.onUnauthorized = () => {
  useSessionStore().clearSession();
  if (router.currentRoute.value.path !== '/login') {
    router.push('/login');
  }
};

app.use(router);
app.mount('#app');
