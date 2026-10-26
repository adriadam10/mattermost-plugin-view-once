import React from 'react';

import {createDraft, deleteMessage, sendDraft, uploadFile} from './api.js';
import {formatSize} from './i18n.js';
import {Modal} from './modal.jsx';

const MAX_FILES = 10;

// Writes a view-once message: text (taken from the message box) and files,
// added with the picker, by dropping them or by pasting them.
export default function Composer({request, onDone, t}) {
    const [text, setText] = React.useState(request.text || '');
    const [files, setFiles] = React.useState([]);
    const [dragging, setDragging] = React.useState(false);
    const [progress, setProgress] = React.useState(null);
    const [error, setError] = React.useState('');
    const picker = React.useRef(null);
    const busy = progress !== null;

    const addFiles = (list) => {
        const added = Array.from(list || []);
        if (!added.length) {
            return;
        }
        setError(files.length + added.length > MAX_FILES ? t('tooManyFiles') : '');
        setFiles([...files, ...added].slice(0, MAX_FILES));
    };

    const close = () => !busy && onDone();

    const send = async () => {
        if (!text.trim() && !files.length) {
            setError(t('empty'));
            return;
        }
        setError('');
        setProgress(0);
        let id = null;
        try {
            ({id} = await createDraft(request.channelId, request.rootId, text));
            for (const [i, file] of files.entries()) {
                // Sequential, so the bar moves steadily and one big file does not starve the rest.
                await uploadFile(id, file, (done) => setProgress((i + done) / files.length));
            }
            await sendDraft(id);
            if (request.text) {
                request.clearText();
            }
            onDone();
        } catch (err) {
            if (id) {
                deleteMessage(id).catch(() => null);
            }
            setProgress(null);
            setError(`${t('sendError')} ${err.message}`);
        }
    };

    const footer = (
        <>
            {busy && <span className='vo-note'>{t('sending')}</span>}
            <button
                type='button'
                className='vo-btn vo-btn--tertiary'
                onClick={close}
                disabled={busy}
            >
                {t('cancel')}
            </button>
            <button
                type='button'
                className='vo-btn vo-btn--primary'
                onClick={send}
                disabled={busy}
            >
                {t('send')}
            </button>
        </>
    );

    return (
        <Modal
            title={t('title')}
            onClose={close}
            footer={footer}
            t={t}
        >
            <p className='vo-note'>{t('composeHint')}</p>
            <textarea
                className='vo-textarea'
                value={text}
                placeholder={t('placeholder')}
                autoFocus={true}
                disabled={busy}
                onChange={(event) => setText(event.target.value)}
                onPaste={(event) => addFiles(event.clipboardData.files)}
                onKeyDown={(event) => (event.ctrlKey || event.metaKey) && event.key === 'Enter' && send()}
            />
            <div
                className={`vo-drop${dragging ? ' vo-drop--over' : ''}`}
                onDragOver={(event) => {
                    event.preventDefault();
                    setDragging(true);
                }}
                onDragLeave={() => setDragging(false)}
                onDrop={(event) => {
                    event.preventDefault();
                    setDragging(false);
                    addFiles(event.dataTransfer.files);
                }}
            >
                {t('dropHere')}{' '}
                <button
                    type='button'
                    className='vo-link'
                    onClick={() => picker.current.click()}
                    disabled={busy}
                >
                    {t('addFiles').toLowerCase()}
                </button>
                <input
                    ref={picker}
                    type='file'
                    multiple={true}
                    style={{display: 'none'}}
                    onChange={(event) => {
                        addFiles(event.target.files);
                        event.target.value = '';
                    }}
                />
            </div>
            {files.length > 0 && (
                <ul className='vo-files'>
                    {files.map((file, i) => (
                        <li key={`${file.name}-${i}`}>
                            <span className='vo-files__name'>{file.name}</span>
                            <span className='vo-files__size'>{formatSize(file.size)}</span>
                            {!busy && (
                                <button
                                    type='button'
                                    className='vo-link'
                                    onClick={() => setFiles(files.filter((_, j) => j !== i))}
                                >
                                    {t('remove')}
                                </button>
                            )}
                        </li>
                    ))}
                </ul>
            )}
            {busy && files.length > 0 && (
                <div className='vo-progress'>
                    <div style={{width: `${Math.round(progress * 100)}%`}}/>
                </div>
            )}
            {error && <p className='vo-error'>{error}</p>}
        </Modal>
    );
}
