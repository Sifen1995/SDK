import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';
import App from './App';

// wayfinding identity accent (bright sky, on the brand axis) for the advertiser portal
document.documentElement.dataset.app = 'ad';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
