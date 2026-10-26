// Mattermost's theme variables, so the plugin follows the user's theme.
const MUTED = 'rgba(var(--center-channel-color-rgb), 0.64)';
const BORDER = 'rgba(var(--center-channel-color-rgb), 0.16)';
const HOVER = 'rgba(var(--center-channel-color-rgb), 0.08)';

export const STYLES = `
.vo-editor-btn {
    display: inline-flex;
    width: 32px;
    height: 32px;
    align-items: center;
    justify-content: center;
    padding: 0;
    border: none;
    border-radius: 4px;
    background: none;
    color: ${MUTED};
    cursor: pointer;
}
.vo-editor-btn:hover { background: ${HOVER}; color: var(--center-channel-color); }
.vo-backdrop {
    position: fixed;
    inset: 0;
    z-index: 10000;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.5);
}
.vo-modal {
    display: flex;
    width: min(560px, calc(100vw - 32px));
    max-height: calc(100vh - 64px);
    flex-direction: column;
    border-radius: 8px;
    background: var(--center-channel-bg);
    box-shadow: 0 20px 32px rgba(0, 0, 0, 0.12);
    color: var(--center-channel-color);
}
.vo-modal--wide { width: min(900px, calc(100vw - 32px)); }
.vo-modal__header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 16px 20px;
    border-bottom: 1px solid ${BORDER};
}
.vo-modal__title { flex: 1; margin: 0; font-size: 18px; font-weight: 600; }
.vo-modal__close {
    padding: 0 4px;
    border: none;
    background: none;
    color: ${MUTED};
    font-size: 24px;
    line-height: 1;
    cursor: pointer;
}
.vo-modal__body {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px 20px;
    overflow-y: auto;
}
.vo-modal__footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;
    padding: 12px 20px;
    border-top: 1px solid ${BORDER};
}
.vo-note { margin: 0; color: ${MUTED}; font-size: 13px; }
.vo-error { margin: 0; color: var(--error-text, #d24b4e); font-size: 13px; }
.vo-textarea {
    min-height: 120px;
    padding: 8px 12px;
    border: 1px solid ${BORDER};
    border-radius: 4px;
    background: var(--center-channel-bg);
    color: var(--center-channel-color);
    font: inherit;
    resize: vertical;
}
.vo-drop {
    padding: 12px;
    border: 1px dashed ${BORDER};
    border-radius: 4px;
    color: ${MUTED};
    font-size: 13px;
    text-align: center;
}
.vo-drop--over { border-color: var(--button-bg); background: ${HOVER}; }
.vo-link {
    padding: 0;
    border: none;
    background: none;
    color: var(--link-color, var(--button-bg));
    font: inherit;
    cursor: pointer;
}
.vo-files { display: flex; flex-direction: column; gap: 4px; margin: 0; padding: 0; list-style: none; }
.vo-files li {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 8px;
    border-radius: 4px;
    background: ${HOVER};
    font-size: 13px;
}
.vo-files__name { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.vo-files__size { color: ${MUTED}; }
.vo-progress { height: 4px; border-radius: 2px; background: ${HOVER}; overflow: hidden; }
.vo-progress > div { height: 100%; background: var(--button-bg); transition: width 0.2s; }
.vo-btn {
    padding: 8px 16px;
    border: none;
    border-radius: 4px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
}
.vo-btn:disabled { opacity: 0.5; cursor: default; }
.vo-btn--primary { background: var(--button-bg); color: var(--button-color); }
.vo-btn--tertiary { background: rgba(var(--button-bg-rgb), 0.08); color: var(--button-bg); }
.vo-btn--danger { background: none; color: var(--error-text, #d24b4e); }
.vo-btn--small { padding: 4px 10px; font-size: 12px; }
.vo-card {
    display: inline-flex;
    max-width: 100%;
    flex-direction: column;
    gap: 8px;
    margin: 4px 0;
    padding: 10px 12px;
    border: 1px solid ${BORDER};
    border-radius: 8px;
}
.vo-card__head { display: flex; align-items: center; gap: 8px; }
.vo-card__title { font-weight: 600; }
.vo-card__meta { color: ${MUTED}; font-size: 12px; }
.vo-card--done { color: ${MUTED}; }
.vo-card__actions { display: flex; flex-wrap: wrap; gap: 8px; }
.vo-viewers { display: flex; flex-direction: column; gap: 4px; }
.vo-viewer { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.vo-viewer img { width: 20px; height: 20px; border-radius: 50%; }
.vo-viewer__when { margin-left: auto; color: ${MUTED}; font-size: 12px; }
.vo-text { overflow-wrap: anywhere; }
.vo-text--plain { white-space: pre-wrap; }
.vo-media { display: block; max-width: 100%; max-height: 60vh; margin: 0 auto; border-radius: 4px; }
`;
