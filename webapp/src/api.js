import manifest from '../../plugin.json';

// Mattermost rejects cookie-authenticated requests without its CSRF header.
function csrfToken() {
    return document.cookie.match(/(?:^|;\s*)MMCSRF=([^;]+)/)?.[1] ?? '';
}

const HEADERS = () => ({
    'X-Requested-With': 'XMLHttpRequest',
    'X-CSRF-Token': csrfToken(),
});

const base = () => `${window.basename || ''}/plugins/${manifest.id}/api/v1/messages`;

export class HttpError extends Error {
    constructor(status, text) {
        super(text || `HTTP ${status}`);
        this.status = status;
    }
}

async function request(method, path, body, options = {}) {
    const response = await fetch(`${base()}${path}`, {
        method,
        credentials: 'include',
        headers: {...HEADERS(), ...(body ? {'Content-Type': 'application/json'} : {})},
        body: body && JSON.stringify(body),
        ...options,
    });
    if (!response.ok) {
        throw new HttpError(response.status, (await response.text()).trim());
    }
    return response.status === 204 ? null : response.json();
}

export const createDraft = (channelId, rootId, text) => request('POST', '', {channel_id: channelId, root_id: rootId || '', text});
export const sendDraft = (id) => request('POST', `/${id}/send`);
export const deleteMessage = (id) => request('DELETE', `/${id}`);
export const fetchInfo = (id) => request('GET', `/${id}`);
export const openMessage = (id) => request('POST', `/${id}/open`);

// keepalive lets the close go out even when the tab is being closed.
export const closeMessage = (id) => request('POST', `/${id}/close`, null, {keepalive: true});

export async function downloadFile(id, index) {
    const response = await fetch(`${base()}/${id}/files/${index}`, {credentials: 'include', headers: HEADERS()});
    if (!response.ok) {
        throw new HttpError(response.status);
    }
    return response.blob();
}

// XMLHttpRequest instead of fetch, for the upload progress.
export function uploadFile(id, file, onProgress) {
    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();
        const query = new URLSearchParams({name: file.name, type: file.type});
        xhr.open('POST', `${base()}/${id}/files?${query}`);
        xhr.withCredentials = true;
        Object.entries(HEADERS()).forEach(([k, v]) => xhr.setRequestHeader(k, v));
        xhr.upload.onprogress = (event) => event.lengthComputable && onProgress(event.loaded / event.total);
        xhr.onload = () => (xhr.status < 300 ? resolve() : reject(new HttpError(xhr.status, xhr.responseText.trim())));
        xhr.onerror = () => reject(new HttpError(0, 'network error'));
        xhr.send(file);
    });
}

// Profiles of users the webapp has not loaded yet (viewers in big channels).
export async function fetchUsers(ids) {
    const response = await fetch(`${window.basename || ''}/api/v4/users/ids`, {
        method: 'POST',
        credentials: 'include',
        headers: {...HEADERS(), 'Content-Type': 'application/json'},
        body: JSON.stringify(ids),
    });
    return response.ok ? response.json() : [];
}
