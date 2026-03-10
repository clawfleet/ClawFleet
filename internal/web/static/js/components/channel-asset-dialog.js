import { html, useState } from '../lib.js';
import { useLang } from '../i18n.js';
import { api } from '../api.js';

export function ChannelAssetDialog({ channel, onClose, onSave, addToast }) {
  const { t } = useLang();
  const isEdit = !!channel;

  const [name, setName] = useState(channel?.name || '');
  const [channelType, setChannelType] = useState(channel?.channel || 'telegram');
  const [token, setToken] = useState(channel?.token || '');
  const [appID, setAppID] = useState(channel?.app_id || '');
  const [appSecret, setAppSecret] = useState(channel?.app_secret || '');
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [validated, setValidated] = useState(isEdit && channel?.validated);

  const isLark = channelType === 'lark';

  const handleTest = async () => {
    setTesting(true);
    try {
      const result = await api.testChannelAsset({
        channel: channelType,
        token: isLark ? '' : token,
        app_id: isLark ? appID : '',
        app_secret: isLark ? appSecret : '',
      });
      if (result.valid) {
        setValidated(true);
        addToast(t('assets.testSuccess'), 'success');
      } else {
        setValidated(false);
        addToast(result.error || t('assets.testFailed'), 'error');
      }
    } catch (err) {
      setValidated(false);
      addToast(err.message, 'error');
    } finally {
      setTesting(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validated) {
      addToast(t('assets.mustValidate'), 'error');
      return;
    }

    setSaving(true);
    try {
      const payload = {
        name,
        channel: channelType,
        token: isLark ? '' : token,
        app_id: isLark ? appID : '',
        app_secret: isLark ? appSecret : '',
      };
      if (isEdit) {
        await api.updateChannelAsset(channel.id, payload);
      } else {
        await api.createChannelAsset(payload);
      }
      addToast(isEdit ? t('assets.updated') : t('assets.created'), 'success');
      onSave();
    } catch (err) {
      addToast(err.message, 'error');
    } finally {
      setSaving(false);
    }
  };

  const canTest = isLark ? (appID && appSecret) : !!token;

  return html`
    <div class="dialog-overlay" onClick=${onClose}>
      <div class="dialog" style="max-width:480px" onClick=${(e) => e.stopPropagation()}>
        <div class="dialog-header">
          <h2>${isEdit ? t('assets.editChannel') : t('assets.addChannel')}</h2>
          <button class="dialog-close" onClick=${onClose}>✕</button>
        </div>
        <form onSubmit=${handleSubmit}>
          <div class="dialog-body">
            <label class="form-label">
              ${t('assets.name')}
              <input type="text" class="form-input" value=${name} onInput=${(e) => setName(e.target.value)}
                placeholder=${t('assets.channelNameHint')} />
            </label>

            <label class="form-label">
              ${t('configure.channel')}
              <select class="form-input" value=${channelType}
                onChange=${(e) => { setChannelType(e.target.value); setValidated(false); }}>
                <option value="telegram">Telegram</option>
                <option value="discord">Discord</option>
                <option value="slack">Slack</option>
                <option value="lark">Lark</option>
              </select>
            </label>

            ${isLark ? html`
              <label class="form-label">
                App ID
                <input type="text" class="form-input" value=${appID}
                  onInput=${(e) => { setAppID(e.target.value); setValidated(false); }}
                  required />
              </label>
              <label class="form-label">
                App Secret
                <input type="password" class="form-input" value=${appSecret}
                  onInput=${(e) => { setAppSecret(e.target.value); setValidated(false); }}
                  required />
              </label>
            ` : html`
              <label class="form-label">
                Bot Token
                <input type="password" class="form-input" value=${token}
                  onInput=${(e) => { setToken(e.target.value); setValidated(false); }}
                  required autofocus />
              </label>
            `}

            <div style="margin-top: 12px">
              <button type="button" class="btn btn-configure" onClick=${handleTest}
                disabled=${testing || !canTest}>
                ${testing ? t('assets.testing') : t('assets.test')}
              </button>
              ${validated && html`<span style="margin-left:8px;color:var(--success)">✅ ${t('assets.validated')}</span>`}
            </div>
          </div>
          <div class="dialog-footer">
            <button type="button" class="btn btn-ghost" onClick=${onClose}>${t('configure.cancel')}</button>
            <button type="submit" class="btn btn-primary" disabled=${saving || !validated}>
              ${saving ? t('assets.saving') : t('configure.submit')}
            </button>
          </div>
        </form>
      </div>
    </div>
  `;
}
