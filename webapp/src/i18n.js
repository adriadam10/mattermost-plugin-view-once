const STRINGS = {
    en: {
        button: 'Send view once',
        title: 'View once',
        composeHint: 'Each recipient can open it only once. It is deleted when everybody has seen it.',
        placeholder: 'Write a message…',
        addFiles: 'Add files',
        dropHere: 'Drop files here, paste them or',
        send: 'Send',
        sending: 'Sending…',
        cancel: 'Cancel',
        remove: 'Remove',
        empty: 'Write something or add a file.',
        tooManyFiles: 'At most 10 files.',
        sendError: 'Could not send it:',
        cardTitle: 'View once',
        open: 'Open',
        opened: 'Opened',
        gone: 'No longer available',
        openedBy: 'Opened by',
        of: 'of',
        nobodyYet: 'Nobody has opened it yet.',
        whoOpened: 'Who opened it',
        deleteForAll: 'Delete for everyone',
        confirmDelete: 'Delete this view-once message for everyone?',
        expires: 'Deleted at the latest on',
        text: 'Text',
        files: 'files',
        file: 'file',
        viewerNote: 'Once you close this, you will not be able to open it again.',
        close: 'Close',
        download: 'Download',
        alreadyOpened: 'You already opened this message.',
        openError: 'Could not open it.',
        loading: 'Loading…',
    },
    es: {
        button: 'Enviar para ver una vez',
        title: 'Ver una vez',
        composeHint: 'Cada destinatario solo puede abrirlo una vez. Se borra cuando lo han visto todos.',
        placeholder: 'Escribe un mensaje…',
        addFiles: 'Añadir ficheros',
        dropHere: 'Arrastra ficheros aquí, pégalos o',
        send: 'Enviar',
        sending: 'Enviando…',
        cancel: 'Cancelar',
        remove: 'Quitar',
        empty: 'Escribe algo o añade un fichero.',
        tooManyFiles: 'Como mucho 10 ficheros.',
        sendError: 'No se ha podido enviar:',
        cardTitle: 'Ver una vez',
        open: 'Abrir',
        opened: 'Abierto',
        gone: 'Ya no está disponible',
        openedBy: 'Abierto por',
        of: 'de',
        nobodyYet: 'Nadie lo ha abierto todavía.',
        whoOpened: 'Quién lo ha abierto',
        deleteForAll: 'Borrar para todos',
        confirmDelete: '¿Borrar este mensaje de ver una vez para todos?',
        expires: 'Se borra como tarde el',
        text: 'Texto',
        files: 'ficheros',
        file: 'fichero',
        viewerNote: 'Cuando lo cierres no podrás volver a abrirlo.',
        close: 'Cerrar',
        download: 'Descargar',
        alreadyOpened: 'Ya has abierto este mensaje.',
        openError: 'No se ha podido abrir.',
        loading: 'Cargando…',
    },
};

export function translator(locale) {
    const strings = STRINGS[(locale || '').split('-')[0]] || STRINGS.en;
    return (key) => strings[key];
}

export const LANGUAGES = Object.keys(STRINGS);
export const keysOf = (lang) => Object.keys(STRINGS[lang]);

const UNITS = ['B', 'KB', 'MB', 'GB'];

export function formatSize(bytes) {
    let size = bytes;
    let unit = 0;
    while (size >= 1024 && unit < UNITS.length - 1) {
        size /= 1024;
        unit++;
    }
    return `${unit === 0 ? size : size.toFixed(1)} ${UNITS[unit]}`;
}

// "Text · 2 files", what a closed message holds, without showing any of it.
export function summary(info, t) {
    const parts = [];
    if (info.has_text) {
        parts.push(t('text'));
    }
    if (info.file_count) {
        parts.push(`${info.file_count} ${info.file_count === 1 ? t('file') : t('files')}`);
    }
    return parts.join(' · ');
}
