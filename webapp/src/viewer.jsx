import React from 'react';

import {closeMessage, downloadFile, openMessage} from './api.js';
import {formatSize} from './i18n.js';
import {Modal} from './modal.jsx';

const isMedia = (type) => /^(image|video|audio)\//.test(type);

// Markdown like a regular post, with the helpers Mattermost exposes to plugins.
function MessageText({text}) {
    const utils = window.PostUtils;
    if (!utils?.formatText || !utils?.messageHtmlToComponent) {
        return <div className='vo-text vo-text--plain'>{text}</div>;
    }
    const html = utils.formatText(text, {atMentions: true, markdown: true});
    return <div className='vo-text'>{utils.messageHtmlToComponent(html, {mentions: true})}</div>;
}

function saveBlob(blob, name) {
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = name;
    link.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
}

function Media({file, url}) {
    if (file.type.startsWith('image/')) {
        return (
            <img
                className='vo-media'
                src={url}
                alt={file.name}
            />
        );
    }
    if (file.type.startsWith('video/')) {
        return (
            <video
                className='vo-media'
                src={url}
                controls={true}
            />
        );
    }
    return (
        <audio
            src={url}
            controls={true}
        />
    );
}

// Shows a message once. Opening it is what spends the recipient's only view;
// closing the viewer ends it, so files are fetched only while it is open.
export default function Viewer({request, onDone, t}) {
    const {id} = request;
    const [content, setContent] = React.useState(null);
    const [error, setError] = React.useState('');
    const [urls, setUrls] = React.useState({});
    const blobs = React.useRef({});

    React.useEffect(() => {
        let current = true;
        const urlsToRevoke = [];
        openMessage(id).then(async (data) => {
            if (!current) {
                return;
            }
            setContent(data);
            for (const [i, file] of data.files.entries()) {
                if (!isMedia(file.type)) {
                    continue;
                }
                const blob = await downloadFile(id, i).catch(() => null);
                if (!current || !blob) {
                    continue;
                }
                blobs.current[i] = blob;
                const url = URL.createObjectURL(new Blob([blob], {type: file.type}));
                urlsToRevoke.push(url);
                setUrls((prev) => ({...prev, [i]: url}));
            }
        }).catch((err) => current && setError(err.status === 410 ? t('alreadyOpened') : t('openError')));

        // Closing the tab counts as closing the viewer.
        const onUnload = () => closeMessage(id);
        window.addEventListener('pagehide', onUnload);
        return () => {
            current = false;
            window.removeEventListener('pagehide', onUnload);
            urlsToRevoke.forEach((url) => URL.revokeObjectURL(url));
        };
    }, [id, t]);

    const close = React.useCallback(() => {
        closeMessage(id).catch(() => null);
        onDone();
    }, [id, onDone]);

    const save = async (file, i) => {
        const blob = blobs.current[i] || await downloadFile(id, i).catch(() => null);
        if (blob) {
            saveBlob(blob, file.name);
        } else {
            setError(t('openError'));
        }
    };

    return (
        <Modal
            title={t('title')}
            onClose={close}
            wide={true}
            footer={(
                <button
                    type='button'
                    className='vo-btn vo-btn--primary'
                    onClick={close}
                >
                    {t('close')}
                </button>
            )}
            t={t}
        >
            <p className='vo-note'>{t('viewerNote')}</p>
            {!content && !error && <p className='vo-note'>{t('loading')}</p>}
            {error && <p className='vo-error'>{error}</p>}
            {content?.text && <MessageText text={content.text}/>}
            {content?.files.map((file, i) => (
                <div key={`${file.name}-${i}`}>
                    {urls[i] && (
                        <Media
                            file={file}
                            url={urls[i]}
                        />
                    )}
                    <ul className='vo-files'>
                        <li>
                            <span className='vo-files__name'>{file.name}</span>
                            <span className='vo-files__size'>{formatSize(file.size)}</span>
                            <button
                                type='button'
                                className='vo-link'
                                onClick={() => save(file, i)}
                            >
                                {t('download')}
                            </button>
                        </li>
                    </ul>
                </div>
            ))}
        </Modal>
    );
}
