import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import MainView from './views/events';

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);
root.render(
  <React.StrictMode>
    <MainView limit={100} offset={0} start={0} end={Math.floor(Date.now() / 1000) + (90 * 24 * 60 * 60)} />
  </React.StrictMode>
);