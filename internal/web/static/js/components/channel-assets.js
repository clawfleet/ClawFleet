import { html, useState, useEffect, useCallback } from '../lib.js';
import { useLang } from '../i18n.js';
import { api } from '../api.js';
import { ChannelAssetDialog } from './channel-asset-dialog.js';

function maskToken(token) {
  if (!token || token.length < 8) return '••••';
  return '••••' + token.slice(-4);
}

export function ChannelAssets({ addToast }) {
  const { t } = useLang();
  const [channels, setChannels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showDialog, setShowDialog] = useState(false);
  const [editChannel, setEditChannel] = useState(null);
  const [testing, setTesting] = useState({});

  const refresh = useCallback(async () => {
    try {
      const data = await api.listChannelAssets();
      setChannels(data || []);
    } catch (err) {
      addToast(err.message, 'error');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { refresh(); }, [refresh]);

  const handleTest = async (channel) => {
    setTesting(prev => ({ ...prev, [channel.id]: true }));
    try {
      const result = await api.testChannelAsset({
        channel: channel.channel,
        token: channel.token,
        app_token: channel.app_token,
        app_id: channel.app_id,
        app_secret: channel.app_secret,
      });
      if (result.valid) {
        addToast(t('assets.testSuccess'), 'success');
      } else {
        addToast(result.error || t('assets.testFailed'), 'error');
      }
    } catch (err) {
      addToast(err.message, 'error');
    } finally {
      setTesting(prev => { const n = { ...prev }; delete n[channel.id]; return n; });
    }
  };

  const handleDelete = async (channel) => {
    if (!confirm(t('assets.confirmDelete', channel.name))) return;
    try {
      await api.deleteChannelAsset(channel.id);
      addToast(t('assets.deleted', channel.name), 'success');
      refresh();
    } catch (err) {
      addToast(err.message, 'error');
    }
  };

  const handleSave = async () => {
    setShowDialog(false);
    setEditChannel(null);
    refresh();
  };

  const handleEdit = (channel) => {
    setEditChannel(channel);
    setShowDialog(true);
  };

  if (loading) {
    return html`<div class="page-content"><div class="dashboard-loading"><p>${t('dashboard.loading')}</p></div></div>`;
  }

  const channelDisplayName = (ch) => {
    const map = { telegram: 'Telegram', discord: 'Discord', slack: 'Slack', lark: 'Lark' };
    return map[ch] || ch;
  };

  return html`
    <div class="page-content">
      <div class="page-header">
        <h2 class="page-title">${t('sidebar.channels')}</h2>
        <button class="btn btn-primary" onClick=${() => { setEditChannel(null); setShowDialog(true); }}>
          ${t('assets.addChannel')}
        </button>
      </div>

      ${channels.length === 0 ? html`
        <div class="assets-empty">
          <div class="assets-empty-icon">💬</div>
          <h3>${t('assets.noChannels')}</h3>
          <p>${t('assets.noChannelsDesc')}</p>
        </div>
      ` : html`
        <div class="assets-list">
          ${channels.map(c => html`
            <div class="asset-card" key=${c.id}>
              <div class="asset-card-header">
                <div class="asset-card-name">${c.name}</div>
                <span class="asset-provider-badge">${channelDisplayName(c.channel)}</span>
              </div>
              <div class="asset-card-details">
                ${c.channel === 'lark' ? html`
                  <div class="asset-detail">
                    <span class="asset-detail-label">${t('assets.appID')}</span>
                    <span class="asset-detail-value mono">${maskToken(c.app_id)}</span>
                  </div>
                ` : c.channel === 'slack' ? html`
                  <div class="asset-detail">
                    <span class="asset-detail-label">${t('assets.botToken')}</span>
                    <span class="asset-detail-value mono">${maskToken(c.token)}</span>
                  </div>
                  <div class="asset-detail">
                    <span class="asset-detail-label">${t('assets.appToken')}</span>
                    <span class="asset-detail-value mono">${maskToken(c.app_token)}</span>
                  </div>
                ` : html`
                  <div class="asset-detail">
                    <span class="asset-detail-label">${t('assets.botToken')}</span>
                    <span class="asset-detail-value mono">${maskToken(c.token)}</span>
                  </div>
                `}
                <div class="asset-detail">
                  <span class="asset-detail-label">${t('assets.status')}</span>
                  <span class="asset-detail-value">${c.validated ? '✅ ' + t('assets.validated') : '⏳ ' + t('assets.unvalidated')}</span>
                </div>
                <div class="asset-detail">
                  <span class="asset-detail-label">${t('assets.usedBy')}</span>
                  <span class="asset-detail-value">${c.used_by || '--'}</span>
                </div>
              </div>
              <div class="asset-card-actions">
                <button class="btn btn-sm btn-configure" onClick=${() => handleTest(c)} disabled=${!!testing[c.id]}>
                  ${testing[c.id] ? t('assets.testing') : t('assets.test')}
                </button>
                <button class="btn btn-sm btn-desktop" onClick=${() => handleEdit(c)}>${t('assets.edit')}</button>
                <button class="btn btn-sm btn-danger" onClick=${() => handleDelete(c)} disabled=${!!c.used_by}>
                  ${t('assets.delete')}
                </button>
              </div>
            </div>
          `)}
        </div>
      `}

      ${showDialog && html`
        <${ChannelAssetDialog}
          channel=${editChannel}
          onClose=${() => { setShowDialog(false); setEditChannel(null); }}
          onSave=${handleSave}
          addToast=${addToast}
        />
      `}
    </div>
  `;
}
