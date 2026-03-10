import { html, useState, useEffect, useCallback } from '../lib.js';
import { useLang } from '../i18n.js';
import { api } from '../api.js';

const IMAGE_FLAVORS = [
  {
    id: 'lightweight',
    name: 'Lightweight',
    base: 'node:22-bookworm (Debian)',
    size: '~1.4 GB',
    desktop: 'XFCE4',
    available: true,
    recommended: true,
  },
  {
    id: 'full-desktop',
    name: 'Full Desktop',
    base: 'ubuntu:24.04',
    size: '~2.5 GB',
    desktop: 'XFCE4 Full',
    available: false,
    recommended: false,
  },
];

export function ImagePage({ addToast }) {
  const { t } = useLang();
  const [imageStatus, setImageStatus] = useState(null);
  const [loading, setLoading] = useState(true);
  const [building, setBuilding] = useState(false);
  const [buildLogs, setBuildLogs] = useState([]);
  const [selectedFlavor, setSelectedFlavor] = useState('lightweight');

  const checkStatus = useCallback(async () => {
    try {
      const data = await api.imageStatus();
      setImageStatus(data);
    } catch (err) {
      addToast(err.message, 'error');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { checkStatus(); }, [checkStatus]);

  const handleBuild = async () => {
    setBuilding(true);
    setBuildLogs([]);

    try {
      const proto = location.protocol === 'https:' ? 'https:' : 'http:';
      const url = `${proto}//${location.host}/api/v1/image/build`;
      const response = await fetch(url, { method: 'POST' });
      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop();

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const msg = line.slice(6);
            setBuildLogs(prev => [...prev, msg]);
          }
          if (line.startsWith('event: done')) {
            addToast(t('image.buildSuccess'), 'success');
          }
          if (line.startsWith('event: error')) {
            addToast(t('image.buildFailed'), 'error');
          }
        }
      }
    } catch (err) {
      addToast(err.message, 'error');
    } finally {
      setBuilding(false);
      checkStatus();
    }
  };

  if (loading) {
    return html`<div class="page-content"><div class="dashboard-loading"><p>${t('dashboard.loading')}</p></div></div>`;
  }

  return html`
    <div class="page-content">
      <div class="page-header">
        <h2 class="page-title">${t('sidebar.image')}</h2>
      </div>

      <div class="image-section">
        <h3 class="section-title">${t('image.selectFlavor')}</h3>
        <div class="image-flavors">
          ${IMAGE_FLAVORS.map(flavor => html`
            <div
              key=${flavor.id}
              class="image-flavor-card ${selectedFlavor === flavor.id ? 'image-flavor-selected' : ''} ${!flavor.available ? 'image-flavor-disabled' : ''}"
              onClick=${() => flavor.available && setSelectedFlavor(flavor.id)}
            >
              <div class="image-flavor-header">
                <div class="image-flavor-radio">
                  ${selectedFlavor === flavor.id ? '●' : '○'}
                </div>
                <div class="image-flavor-name">
                  ${flavor.name}
                  ${flavor.recommended ? html` <span class="image-flavor-badge">${t('image.recommended')}</span>` : ''}
                  ${!flavor.available ? html` <span class="image-flavor-badge-soon">Coming Soon</span>` : ''}
                </div>
              </div>
              <div class="image-flavor-details">
                <div class="image-flavor-detail">${t('image.baseImage')}: ${flavor.base}</div>
                <div class="image-flavor-detail">${t('image.size')}: ${flavor.size}</div>
                <div class="image-flavor-detail">${t('image.desktop')}: ${flavor.desktop}</div>
              </div>
            </div>
          `)}
        </div>
      </div>

      <div class="image-section">
        <h3 class="section-title">${t('image.currentStatus')}</h3>
        <div class="image-status-card">
          ${imageStatus?.built
            ? html`<span class="image-status-built">✅ ${t('image.built')} ${imageStatus.image}</span>`
            : html`<span class="image-status-missing">⚠️ ${t('image.notBuilt')}</span>`
          }
        </div>
      </div>

      <div class="image-actions">
        <button class="btn btn-primary" onClick=${handleBuild} disabled=${building}>
          ${building ? t('image.building') : t('image.build')}
        </button>
      </div>

      ${buildLogs.length > 0 && html`
        <div class="image-section">
          <h3 class="section-title">${t('image.buildLog')}</h3>
          <div class="image-build-log">
            ${buildLogs.map((line, i) => html`<div key=${i} class="logs-line">${line}</div>`)}
          </div>
        </div>
      `}
    </div>
  `;
}
