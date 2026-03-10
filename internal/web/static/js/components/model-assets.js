import { html, useState, useEffect, useCallback } from '../lib.js';
import { useLang } from '../i18n.js';
import { api } from '../api.js';
import { ModelAssetDialog } from './model-asset-dialog.js';

function maskKey(key) {
  if (!key || key.length < 8) return '••••';
  return '••••' + key.slice(-4);
}

export function ModelAssets({ addToast }) {
  const { t } = useLang();
  const [models, setModels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showDialog, setShowDialog] = useState(false);
  const [editModel, setEditModel] = useState(null);
  const [testing, setTesting] = useState({});

  const refresh = useCallback(async () => {
    try {
      const data = await api.listModelAssets();
      setModels(data || []);
    } catch (err) {
      addToast(err.message, 'error');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { refresh(); }, [refresh]);

  const handleTest = async (model) => {
    setTesting(prev => ({ ...prev, [model.id]: true }));
    try {
      const result = await api.testModelAsset({
        provider: model.provider,
        api_key: model.api_key,
        model: model.model,
      });
      if (result.valid) {
        addToast(t('assets.testSuccess'), 'success');
      } else {
        addToast(result.error || t('assets.testFailed'), 'error');
      }
    } catch (err) {
      addToast(err.message, 'error');
    } finally {
      setTesting(prev => { const n = { ...prev }; delete n[model.id]; return n; });
    }
  };

  const handleDelete = async (model) => {
    if (!confirm(t('assets.confirmDelete', model.name))) return;
    try {
      await api.deleteModelAsset(model.id);
      addToast(t('assets.deleted', model.name), 'success');
      refresh();
    } catch (err) {
      addToast(err.message, 'error');
    }
  };

  const handleSave = async () => {
    setShowDialog(false);
    setEditModel(null);
    refresh();
  };

  const handleEdit = (model) => {
    setEditModel(model);
    setShowDialog(true);
  };

  if (loading) {
    return html`<div class="page-content"><div class="dashboard-loading"><p>${t('dashboard.loading')}</p></div></div>`;
  }

  return html`
    <div class="page-content">
      <div class="page-header">
        <h2 class="page-title">${t('sidebar.models')}</h2>
        <button class="btn btn-primary" onClick=${() => { setEditModel(null); setShowDialog(true); }}>
          ${t('assets.addModel')}
        </button>
      </div>

      ${models.length === 0 ? html`
        <div class="assets-empty">
          <div class="assets-empty-icon">🤖</div>
          <h3>${t('assets.noModels')}</h3>
          <p>${t('assets.noModelsDesc')}</p>
        </div>
      ` : html`
        <div class="assets-list">
          ${models.map(m => html`
            <div class="asset-card" key=${m.id}>
              <div class="asset-card-header">
                <div class="asset-card-name">${m.name}</div>
                <span class="asset-provider-badge">${providerDisplay(m.provider)}</span>
              </div>
              <div class="asset-card-details">
                <div class="asset-detail">
                  <span class="asset-detail-label">${t('configure.model')}</span>
                  <span class="asset-detail-value">${m.model}</span>
                </div>
                <div class="asset-detail">
                  <span class="asset-detail-label">${t('configure.apiKey')}</span>
                  <span class="asset-detail-value mono">${maskKey(m.api_key)}</span>
                </div>
                <div class="asset-detail">
                  <span class="asset-detail-label">${t('assets.status')}</span>
                  <span class="asset-detail-value">${m.validated ? '✅ ' + t('assets.validated') : '⏳ ' + t('assets.unvalidated')}</span>
                </div>
              </div>
              <div class="asset-card-actions">
                <button class="btn btn-sm btn-configure" onClick=${() => handleTest(m)} disabled=${!!testing[m.id]}>
                  ${testing[m.id] ? t('assets.testing') : t('assets.test')}
                </button>
                <button class="btn btn-sm btn-desktop" onClick=${() => handleEdit(m)}>${t('assets.edit')}</button>
                <button class="btn btn-sm btn-danger" onClick=${() => handleDelete(m)}>
                  ${t('assets.delete')}
                </button>
              </div>
            </div>
          `)}
        </div>
      `}

      ${showDialog && html`
        <${ModelAssetDialog}
          model=${editModel}
          onClose=${() => { setShowDialog(false); setEditModel(null); }}
          onSave=${handleSave}
          addToast=${addToast}
        />
      `}
    </div>
  `;
}

function providerDisplay(provider) {
  const map = { anthropic: 'Anthropic', openai: 'OpenAI', google: 'Google', deepseek: 'DeepSeek' };
  return map[provider] || provider;
}
