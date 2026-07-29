import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';
import App from './App';

// wayfinding identity accent (deep navy, on the brand axis) for the admin portal
document.documentElement.dataset.app = 'admin';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
